package grpcserver

type EntrypointOption interface {
	apply(*options)
}
