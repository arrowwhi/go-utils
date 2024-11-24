package db_config

type DBConfig struct {
	Host     string `envconfig:"HOST" required:"true"`
	Port     int    `envconfig:"PORT" default:"5432"`
	User     string `envconfig:"USER" required:"true"`
	Password string `envconfig:"PASSWORD" required:"true"`
	DBName   string `envconfig:"NAME" required:"true"`
	SSLMode  string `envconfig:"SSL_MODE" default:"disable"`
}
