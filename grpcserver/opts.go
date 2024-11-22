package grpcserver

import (
	"go-utils/grpcserver/adapter"
	"google.golang.org/grpc"
)

type options struct {
	adapters                    []adapter.ImplementationAdapter
	grpcUnaryServerInterceptors []grpc.UnaryServerInterceptor
	grpcAddress                 string
	httpAddress                 string
}

type option func(o *options)

func (o option) apply(os *options) { o(os) }

func WithImplementationAdapters(adapters ...adapter.ImplementationAdapter) EntrypointOption {
	return option(func(o *options) { o.adapters = append(o.adapters, adapters...) })
}

func WithGrpcUnaryServerInterceptors(grpcUnaryServerInterceptors ...grpc.UnaryServerInterceptor) EntrypointOption {
	return option(func(o *options) {
		o.grpcUnaryServerInterceptors = append(o.grpcUnaryServerInterceptors, grpcUnaryServerInterceptors...)
	})
}
