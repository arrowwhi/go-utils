package config

type EnvMode string

const (
	DevMode   EnvMode = "dev"
	StageMode EnvMode = "stage"
	ProdMode  EnvMode = "production"
)
