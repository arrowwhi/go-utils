package grpcserver

import (
	"context"
	"errors"
	"fmt"
	"github.com/arrowwhi/go-utils/grpcserver/gateway"
	"github.com/arrowwhi/go-utils/grpcserver/interceptors"
	"github.com/arrowwhi/go-utils/grpcserver/metrics"
	"go.uber.org/zap"
	"google.golang.org/grpc/reflection"
	"net"
	"sync"

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
	// Interceptors
	ints := []grpc.ServerOption{grpc.ChainUnaryInterceptor(interceptors.MetricsMiddleware(s.config.ServiceName))}
	for _, v := range s.grpcUnaryServerInterceptors {
		ints = append(ints, grpc.ChainUnaryInterceptor(v))
	}

	// Create gRPC server
	s.grpcServer = grpc.NewServer(ints...)

	// Initialize and start metrics
	if err := metrics.InitMetrics(s.logger); err != nil {
		return fmt.Errorf("init metrics: %w", err)
	}
	s.logger.Info("Metrics initialized successfully")

	if err := metrics.StartPrometheusServer(s.logger, s.metricsPort); err != nil {
		return fmt.Errorf("start Prometheus server: %w", err)
	}
	s.logger.Info("Prometheus server started", zap.String("prometheus-port", s.metricsPort))

	// Listen on the configured port
	listener, err := net.Listen("tcp", fmt.Sprintf(":%s", s.config.GRPCPort))
	if err != nil {
		return fmt.Errorf("failed to listen on port %v", err)
	}

	// Register services
	for _, v := range s.adapters {
		v.RegisterServer(s.grpcServer)
	}

	// Register reflection service on gRPC server.
	reflection.Register(s.grpcServer)

	// Prepare HTTP gateway options
	var gatewayOptions []gateway.Option
	for _, v := range s.adapters {
		gatewayOptions = append(gatewayOptions, gateway.WithHandler(v.RegisterHandler))
	}

	// Create the gateway
	gw := gateway.NewGateway(
		s.config,
		s.logger,
		gatewayOptions...,
	)

	// Use a WaitGroup to wait for the servers to shut down gracefully
	var wg sync.WaitGroup
	wg.Add(2) // We'll have two goroutines: one for gRPC server and one for HTTP gateway

	// Channel to capture errors
	errChan := make(chan error, 2)

	// Start the gRPC server in a goroutine
	go func() {
		defer wg.Done()
		s.logger.Info("Starting gRPC server", zap.String("address", s.config.GRPCPort))
		if err := s.grpcServer.Serve(listener); err != nil && err != grpc.ErrServerStopped {
			s.logger.Error("Failed to serve gRPC", zap.Error(err))
			errChan <- err
		}
		s.logger.Info("gRPC server stopped")
	}()

	// Start the HTTP gateway in a goroutine
	go func() {
		defer wg.Done()
		s.logger.Info("Starting HTTP gateway",
			zap.String("address", s.config.GatewayPort),
			zap.String("grpc_address", s.config.GRPCPort),
			zap.String("service_name", s.config.ServiceName),
		)
		if err := gw.Start(ctx); err != nil && !errors.Is(err, context.Canceled) {
			s.logger.Error("Failed to start HTTP gateway", zap.Error(err))
			errChan <- err
		}
		s.logger.Info("HTTP gateway stopped")
	}()

	// Listen for context cancellation or errors
	select {
	case <-ctx.Done():
		s.logger.Info("Context canceled, initiating graceful shutdown")
		s.Stop()
	case err := <-errChan:
		s.logger.Error("Server encountered an error", zap.Error(err))
		s.Stop()
		return err
	}

	// Wait for all goroutines to finish
	wg.Wait()
	s.logger.Info("Server shut down gracefully")
	return nil
}

// Stop корректно завершает работу gRPC сервера.
func (s *Server) Stop() {
	s.logger.Info("Stopping gRPC server...")
	s.grpcServer.GracefulStop()
	// If you have a mechanism to stop the HTTP gateway, you should call it here
	// e.g., gw.Stop()
	s.logger.Info("gRPC server stopped")
}
