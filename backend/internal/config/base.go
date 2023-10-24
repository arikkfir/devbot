package config

type RedisConfig struct {
	Host string `env:"HOST" value-name:"HOST" long:"host" description:"Redis host name" required:"yes"`
	Port int    `env:"PORT" value-name:"PORT" long:"port" description:"Redis port" default:"6379"`
	TLS  bool   `env:"TLS" long:"tls" description:"Whether to use TLS to connect to Redis"`
}

type CommandConfig struct {
	LogLevel   string      `env:"LOG_LEVEL" value-name:"LEVEL" long:"log-level" description:"Log level" default:"info" enum:"trace,debug,info,warn,error,fatal,panic"`
	DevMode    bool        `env:"DEV_MODE" long:"dev-mode" description:"Development mode"`
	HealthPort int         `env:"HEALTH_PORT" value-name:"PORT" long:"health-port" description:"Port to listen on for health checks" default:"9000"`
	Redis      RedisConfig `group:"redis" namespace:"redis" env-namespace:"REDIS"`
}
