package github

type WebhookConfig struct {
	DisableJSONLogging bool
	LogLevel           string
	HealthPort         int
	ServerPort         int
	WebhookSecret      string
}
