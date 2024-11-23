package config

type Config struct {
	ServiceName string `envconfig:"NAME" required:"true"`
	Version     string `envconfig:"VERSION" required:"true"`

	GRPCPort       string `envconfig:"GRPC_PORT" default:"50051"`
	GatewayPort    string `envconfig:"GW_PORT" default:"8080"`
	PrometheusPort string `envconfig:"PROMETHEUS_PORT" default:"9090"`
}
