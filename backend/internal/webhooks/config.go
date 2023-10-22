package webhooks

import "github.com/arikkfir/devbot/backend/internal/config"

type WebhookConfig struct {
	Port   int    `env:"PORT" value-name:"PORT" long:"port" description:"Port to listen on" default:"8000"`
	Secret string `env:"SECRET" value-name:"SECRET" long:"secret" description:"Webhook secret" required:"yes"`
}

type WebhookCommandConfig struct {
	config.CommandConfig
	Webhook WebhookConfig `group:"webhook" namespace:"webhook" env-namespace:"WEBHOOK"`
}
