package config

import (
	grpc_config "github.com/arrowwhi/go-utils/grpcserver/config"
)

type Config struct {
	ServerConfig grpc_config.Config

	LogLevel string `envconfig:"LOG_LEVEL" default:"debug"`
}
