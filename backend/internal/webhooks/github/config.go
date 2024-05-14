package github

type WebhookConfig struct {
	DisableJSONLogging bool   `desc:"Disable JSON logging."`
	LogLevel           string `config:"required" desc:"Log level, must be one of: trace,debug,info,warn,error,fatal,panic"`
	HealthPort         int    `config:"required" desc:"Port to listen on for health checks."`
	ServerPort         int    `config:"required" desc:"Port to listen on."`
	WebhookSecret      string `config:"required" desc:"Webhook secret."`
}
