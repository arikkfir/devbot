package internal

type CommandConfig struct {
	LogLevel   string `env:"LOG_LEVEL" value-name:"LEVEL" long:"log-level" description:"Log level" default:"info" enum:"trace,debug,info,warn,error,fatal,panic"`
	DevMode    bool   `env:"DEV_MODE" long:"dev-mode" description:"Development mode"`
	HealthPort int    `env:"HEALTH_PORT" value-name:"PORT" long:"health-port" description:"Port to listen on for health checks" default:"9000"`
}

type WebhookConfig struct {
	Port   int    `env:"PORT" value-name:"PORT" long:"port" description:"Port to listen on" default:"8000"`
	Secret string `env:"SECRET" value-name:"SECRET" long:"secret" description:"Webhook secret" required:"yes"`
}

type WebhookCommandConfig struct {
	CommandConfig
	Webhook WebhookConfig `group:"webhook" namespace:"webhook" env-namespace:"WEBHOOK"`
}
