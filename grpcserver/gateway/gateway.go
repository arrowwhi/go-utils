package gateway

import (
	"context"
	"errors"
	"fmt"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"go-utils/grpcserver/adapter"
	"go.uber.org/zap"
	"google.golang.org/protobuf/encoding/protojson"
	"net/http"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Gateway struct {
	logger     zap.Logger
	adapters   []adapter.ImplementationAdapter
	gatewayMux *http.ServeMux
}

func NewGateway(logger zap.Logger, adapters []adapter.ImplementationAdapter) *Gateway {
	return &Gateway{
		logger:     logger,
		adapters:   adapters,
		gatewayMux: http.NewServeMux(),
	}
}

func (g *Gateway) Start(ctx context.Context, httpPort string, grpcPort string) error {
	httpUrl := httpPort
	grpcUrl := grpcPort

	conn, err := grpc.NewClient(grpcUrl, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return err
	}
	//closer.Add(conn.Close) todo

	defaultMarshallerOption := runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{
		MarshalOptions: protojson.MarshalOptions{
			UseProtoNames:   false,
			EmitUnpopulated: true,
		},
		UnmarshalOptions: protojson.UnmarshalOptions{
			DiscardUnknown: true,
		},
	})

	opts := []runtime.ServeMuxOption{
		defaultMarshallerOption,
		//runtime.WithErrorHandler(httpError.HTTPErrorHandler),
	}

	//opts = append(opts, g.serverMuxOptions...)

	//opts = append(opts,
	//	runtime.WithIncomingHeaderMatcher(func(s string) (string, bool) {
	//		if g.incomingHeaderMatcherFunc != nil {
	//			if header, ok := g.incomingHeaderMatcherFunc(s); ok {
	//				return header, ok
	//			}
	//		}
	//
	//		return defaultIncomingHeaderMatcher(s)
	//	}),
	//	runtime.WithOutgoingHeaderMatcher(func(s string) (string, bool) {
	//		if g.outgoingHeaderMatcherFunc != nil {
	//			if header, ok := g.outgoingHeaderMatcherFunc(s); ok {
	//				return header, ok
	//			}
	//		}
	//
	//		return defaultOutgoingHeaderMatcher(s)
	//	}),
	//	runtime.WithHealthzEndpoint(health.NewHealthClient(conn)))

	gwmux := runtime.NewServeMux(
		opts...,
	)

	for _, impl := range g.adapters {
		if err := impl.RegisterHandler(ctx, gwmux, conn); err != nil {
			return err
		}
	}

	mux := http.NewServeMux()

	//g.registerSwaggerHandler(mux)

	mux.Handle("/", gwmux)

	var httpHandler http.Handler
	httpHandler = mux

	//if g.corsEnabled {
	//	httpHandler = cors.New(g.corsOptions).Handler(mux)
	//}

	httpServer := &http.Server{
		Addr:              httpUrl,
		Handler:           httpHandler,
		ReadHeaderTimeout: time.Minute,
	}

	g.logger.Info(fmt.Sprintf("gRPC GW starting on address - %s", httpUrl))

	if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		g.logger.Error(fmt.Sprintf("failed to start http gateway server: %v", err))
		return err
	}

	return nil

}
