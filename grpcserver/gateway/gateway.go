package gateway

import (
	"context"
	"errors"
	"fmt"
	"github.com/arrowwhi/go-utils/grpcserver/grpc_config"
	"net/http"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Option defines a function type for configuring the Gateway.
type Option func(*Gateway)

// Gateway represents the HTTP gateway for the gRPC server.
type Gateway struct {
	ServerConfig grpc_config.Config
	logger       *zap.Logger
	handlers     []func(ctx context.Context, mux *runtime.ServeMux, conn *grpc.ClientConn) error
	dialOptions  []grpc.DialOption
}

// NewGateway creates a new Gateway instance with the provided options.
func NewGateway(ServerConfig grpc_config.Config, logger *zap.Logger, opts ...Option) *Gateway {
	g := &Gateway{
		ServerConfig: ServerConfig,
		logger:       logger,
	}

	for _, opt := range opts {
		opt(g)
	}

	return g
}

// WithHandler adds a gRPC handler to the Gateway.
func WithHandler(registerHandler func(ctx context.Context, mux *runtime.ServeMux, conn *grpc.ClientConn) error) Option {
	return func(g *Gateway) {
		g.handlers = append(g.handlers, registerHandler)
	}
}

// WithDialOptions sets custom gRPC dial options for the Gateway.
func WithDialOptions(dialOpts ...grpc.DialOption) Option {
	return func(g *Gateway) {
		g.dialOptions = append(g.dialOptions, dialOpts...)
	}
}

// Start launches the HTTP gateway server.
func (g *Gateway) Start(ctx context.Context) error {
	conn, err := grpc.NewClient(fmt.Sprintf("0.0.0.0:%s", g.ServerConfig.GRPCPort), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return err
	}
	defer func(conn *grpc.ClientConn) {
		err := conn.Close()
		if err != nil {
			g.logger.Error("failed to close gRPC connection", zap.Error(err))
		}
	}(conn)

	gwmux := runtime.NewServeMux(
	//opts...,
	)

	for _, impl := range g.handlers {
		if err := impl(ctx, gwmux, conn); err != nil {
			return err
		}
	}

	mux := http.NewServeMux()

	//g.registerSwaggerHandler(mux)

	mux.Handle("/", gwmux)

	var httpHandler http.Handler
	httpHandler = mux

	// Start the HTTP server
	g.logger.Info("Starting HTTP gateway",
		zap.String("address", g.ServerConfig.GatewayPort),
		zap.String("service_name", g.ServerConfig.ServiceName),
	)

	httpServer := &http.Server{
		Addr:              fmt.Sprintf("0.0.0.0:%s", g.ServerConfig.GatewayPort),
		Handler:           httpHandler,
		ReadHeaderTimeout: time.Minute,
	}

	g.logger.Info(fmt.Sprintf("gRPC GW starting on address - %s", g.ServerConfig.GatewayPort))

	if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		g.logger.Error(fmt.Sprintf("failed to start http gateway server: %v", err))
		return err
	}

	return nil
}
