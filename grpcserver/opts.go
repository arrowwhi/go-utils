package grpcserver

import (
	"github.com/arrowwhi/go-utils/grpcserver/handler_adapter"
	"google.golang.org/grpc"
)

type options struct {
	adapters                    []handler_adapter.ImplementationAdapter
	grpcUnaryServerInterceptors []grpc.UnaryServerInterceptor
}

type option func(o *options)

func (o option) apply(os *options) { o(os) }

func WithImplementationAdapters(adapters ...handler_adapter.ImplementationAdapter) EntrypointOption {
	return option(func(o *options) { o.adapters = append(o.adapters, adapters...) })
}

func WithGrpcUnaryServerInterceptors(grpcUnaryServerInterceptors ...grpc.UnaryServerInterceptor) EntrypointOption {
	return option(func(o *options) {
		o.grpcUnaryServerInterceptors = append(o.grpcUnaryServerInterceptors, grpcUnaryServerInterceptors...)
	})
}

type EntrypointOption interface {
	apply(*options)
}
