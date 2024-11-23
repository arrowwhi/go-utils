package gateway

import (
	"context"
	"fmt"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Option defines a function type for configuring the Gateway.
type Option func(*Gateway)

// Gateway represents the HTTP gateway for the gRPC server.
type Gateway struct {
	grpcAddress string
	httpAddress string
	serviceName string
	logger      *zap.Logger
	handlers    []func(context.Context, *runtime.ServeMux, string, []grpc.DialOption) error
	dialOptions []grpc.DialOption
}

// NewGateway creates a new Gateway instance with the provided options.
func NewGateway(grpcAddress, httpAddress, serviceName string, logger *zap.Logger, opts ...Option) *Gateway {
	g := &Gateway{
		grpcAddress: grpcAddress,
		httpAddress: httpAddress,
		serviceName: serviceName,
		logger:      logger,
	}

	for _, opt := range opts {
		opt(g)
	}

	return g
}

// WithHandler adds a gRPC handler to the Gateway.
func WithHandler(registerHandler func(context.Context, *runtime.ServeMux, string, []grpc.DialOption) error) Option {
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
	mux := runtime.NewServeMux()

	// Use provided dial options or default to insecure credentials
	var dialOpts []grpc.DialOption
	if len(g.dialOptions) > 0 {
		dialOpts = g.dialOptions
	} else {
		dialOpts = []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	}

	// Register all handlers
	for _, handler := range g.handlers {
		if err := handler(ctx, mux, g.grpcAddress, dialOpts); err != nil {
			return fmt.Errorf("failed to register handler: %w", err)
		}
	}

	// Start the HTTP server
	g.logger.Info("Starting HTTP gateway",
		zap.String("address", g.httpAddress),
		zap.String("grpc_address", g.grpcAddress),
		zap.String("service_name", g.serviceName),
	)

	return http.ListenAndServe(g.httpAddress, mux)
}
