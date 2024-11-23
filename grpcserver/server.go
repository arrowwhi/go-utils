package grpcserver

import (
	"context"
	"fmt"
	"github.com/arrowwhi/go-utils/grpcserver/interceptors"
	"github.com/arrowwhi/go-utils/grpcserver/metrics"
	"go.uber.org/zap"
	"google.golang.org/grpc/reflection"
	"net"

	"google.golang.org/grpc"
)

type Server struct {
	options
	metricsPort string
	logger      *zap.Logger
	grpcAddress string
	httpAddress string
	serviceName string
	grpcServer  *grpc.Server
}

func NewServer(grpcAddress, httpAddress, serviceName, metricsPort string, logger *zap.Logger, opts ...EntrypointOption) (*Server, error) {
	o := options{}

	for _, opt := range opts {
		opt.apply(&o)
	}

	return &Server{
		logger:      logger,
		options:     o,
		metricsPort: metricsPort,
		grpcAddress: fmt.Sprintf("localhost:%s", grpcAddress),
		httpAddress: fmt.Sprintf("localhost:%s", httpAddress),
		serviceName: serviceName,
	}, nil
}

// Start запускает gRPC сервер и начинает прослушивание входящих запросов.
func (s *Server) Start(ctx context.Context) error {

	ints := []grpc.ServerOption{grpc.ChainUnaryInterceptor(interceptors.MetricsMiddleware(s.serviceName))}

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

	listener, err := net.Listen("tcp", fmt.Sprintf(s.grpcAddress))
	if err != nil {
		return fmt.Errorf("failed to listen on port %v", err)
	}

	for _, v := range s.adapters {
		v.RegisterServer(s.grpcServer)
	}

	reflection.Register(s.grpcServer)
	return s.grpcServer.Serve(listener)
}

// Stop корректно завершает работу gRPC сервера.
func (s *Server) Stop() {
	s.grpcServer.GracefulStop()
}
