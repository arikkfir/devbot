package webhooks

import "github.com/arikkfir/devbot/backend/internal/config"

type WebhookConfig struct {
	config.CommandConfig
	HealthPort int    `env:"HEALTH_PORT" value-name:"PORT" long:"health-port" description:"Port to listen on for health checks" default:"9000"`
	Port       int    `env:"PORT" value-name:"PORT" long:"port" description:"Port to listen on" default:"8000"`
	Secret     string `env:"SECRET" value-name:"SECRET" long:"secret" description:"Webhook secret" required:"yes"`
}
