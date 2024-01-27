package config

type CommandConfig struct {
	LogLevel string `env:"LOG_LEVEL" value-name:"LEVEL" long:"log-level" description:"Log level" default:"info" enum:"trace,debug,info,warn,error,fatal,panic"`
	DevMode  bool   `env:"DEV_MODE" long:"dev-mode" description:"Development mode"`
}
