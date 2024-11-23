package config

import (
	"github.com/arrowwhi/go-utils/grpcserver/grpc_config"
)

type Config struct {
	ServerConfig grpc_config.Config

	LogLevel string `envconfig:"LOG_LEVEL" default:"debug"`
}
