package config

type LogConfig struct {
	CallerInfo  bool `env:"CALLER_INFO" short:"c" long:"caller-info" required:"false" description:"Show caller information"`
	JSONLogging bool `env:"JSON_LOGGING" short:"j" long:"json" description:"Print every log entry as a JSON object"`
}

type HTTPConfig struct {
	Address         string `env:"ADDR" long:"address" default:":8080" required:"false" description:"HTTP address to listen to"`
	LogResponseBody bool   `env:"LOG_RESPONSE_BODY" long:"log-response-body" description:"Log HTTP response body"`
}
