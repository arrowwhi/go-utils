package grpcserver

import (
	"context"
	"fmt"
	"github.com/arrowwhi/go-utils/grpcserver/gateway"
	"github.com/arrowwhi/go-utils/grpcserver/interceptors"
	"github.com/arrowwhi/go-utils/grpcserver/metrics"
	"go.uber.org/zap"
	"google.golang.org/grpc/reflection"
	"net"

	"google.golang.org/grpc"

	"github.com/arrowwhi/go-utils/grpcserver/grpc_config"
)

type Server struct {
	options
	metricsPort string
	logger      *zap.Logger
	grpcServer  *grpc.Server
	config      grpc_config.Config
}

func NewServer(serverConfig grpc_config.Config, logger *zap.Logger, opts ...EntrypointOption) (*Server, error) {
	o := options{}

	for _, opt := range opts {
		opt.apply(&o)
	}

	return &Server{
		logger:  logger,
		options: o,
		config:  serverConfig,
	}, nil
}

// Start запускает gRPC сервер и начинает прослушивание входящих запросов.
func (s *Server) Start(ctx context.Context) error {

	ints := []grpc.ServerOption{grpc.ChainUnaryInterceptor(interceptors.MetricsMiddleware(s.config.ServiceName))}

	for _, v := range s.grpcUnaryServerInterceptors {
		ints = append(ints, grpc.ChainUnaryInterceptor(v))
	}

	s.grpcServer = grpc.NewServer(ints...)

	// Инициализация и запуск метрик

	if err := metrics.InitMetrics(s.logger); err != nil {
		return fmt.Errorf("init metrics: %w", err)
	}
	s.logger.Info("Metrics initialized successfully")

	if err := metrics.StartPrometheusServer(s.logger, s.metricsPort); err != nil {
		return fmt.Errorf("start Prometheus server: %w", err)
	}
	s.logger.Info("Prometheus server started", zap.String("prometheus-port", s.metricsPort))

	//---

	listener, err := net.Listen("tcp", fmt.Sprintf("localhost:%s", s.config.GRPCPort))
	if err != nil {
		return fmt.Errorf("failed to listen on port %v", err)
	}

	for _, v := range s.adapters {
		v.RegisterServer(s.grpcServer)
	}

	reflection.Register(s.grpcServer)

	var gatewayOptions []gateway.Option
	for _, v := range s.adapters {
		gatewayOptions = append(gatewayOptions, gateway.WithHandler(v.RegisterHandler))
	}

	gw := gateway.NewGateway(
		s.config,
		s.logger,
		gatewayOptions...,
	)

	go func() {
		s.logger.Info("Starting HTTP gateway",
			zap.String("address", s.config.GatewayPort),
			zap.String("grpc_address", s.config.GRPCPort),
			zap.String("service_name", s.config.ServiceName),
		)
		if err := gw.Start(ctx); err != nil {
			s.logger.Error("Failed to start HTTP gateway", zap.Error(err))
			// Обработка ошибки при необходимости
		}
	}()

	return s.grpcServer.Serve(listener)
}

// Stop корректно завершает работу gRPC сервера.
func (s *Server) Stop() {
	s.grpcServer.GracefulStop()
}
