package config

type Config struct {
	ServiceName string `envconfig:"NAME" required:"true"`    // Имя приложения
	Version     string `envconfig:"VERSION" required:"true"` // Версия приложения

	GRPCPort       string `envconfig:"GRPC_PORT" default:"50051"`      // GRPC-Порт
	GatewayPort    string `envconfig:"GW_PORT" required:"true"`        // GW-Порт
	PrometheusPort string `envconfig:"PROMETHEUS_PORT" default:"9090"` // Порт Prometheus

	LogLevel string  `envconfig:"LOG_LEVEL" default:"debug"` // Уровень логирования
	Env      EnvMode `envconfig:"ENV" default:"dev"`         // Среда окружения

}
