package internal

import "time"

type Config struct {
	LogLevel      string     `env:"LOG_LEVEL" value-name:"LEVEL" long:"log-level" description:"Log level" default:"info" enum:"trace,debug,info,warn,error,fatal,panic"`
	DevMode       bool       `env:"DEV_MODE" long:"dev-mode" description:"Development mode"`
	HTTP          HTTPConfig `group:"http" namespace:"http" env-namespace:"HTTP"`
	WebhookSecret string     `env:"WEBHOOK_SECRET" value-name:"SECRET" long:"webhook-secret" description:"Webhook secret" required:"yes"`
}

type HTTPConfig struct {
	Port                       int        `env:"PORT" value-name:"PORT" long:"port" description:"Port to listen on" default:"8000"`
	DisableAccessLog           bool       `env:"DISABLE_ACCESS_LOG" long:"disable-access-log" description:"Disable access log"`
	HealthPort                 int        `env:"HEALTH_PORT" value-name:"PORT" long:"health-port" description:"Port to listen on for health checks" default:"9000"`
	CORS                       CORSConfig `group:"cors" namespace:"cors" env-namespace:"CORS"`
	AccessLogExcludedHeaders   []string   `env:"ACCESS_LOG_EXCLUDED_HEADERS" value-name:"PATTERN" long:"access-log-excluded-headers" description:"List of header patterns to exclude from the access log"`
	AccessLogExcludeRemoteAddr bool       `env:"ACCESS_LOG_EXCLUDE_REMOTE_ADDR" long:"access-log-exclude-remote-addr" description:"Exclude the client remote address from the access log"`
}

type CORSConfig struct {
	AllowedOrigins     []string      `env:"ALLOWED_ORIGINS" value-name:"ORIGIN" long:"allowed-origins" description:"List of origins a cross-domain request can be executed from (https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Access-Control-Allow-Origin)" required:"yes"`
	AllowMethods       []string      `env:"ALLOWED_METHODS" value-name:"METHOD" long:"allowed-methods" description:"List of HTTP methods a cross-domain request can use (https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Access-Control-Allow-Methods)"`
	AllowHeaders       []string      `env:"ALLOWED_HEADERS" value-name:"NAME" long:"allowed-headers" description:"List HTTP headers a cross-domain request can use (https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Access-Control-Allow-Headers)" default:"accept" default:"x-greenstar-tenant-id" default:"authorization" default:"content-type"`
	DisableCredentials bool          `env:"DISABLE_CREDENTIALS" long:"disable-credentials" description:"Disable access to credentials for JavaScript client code (https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Access-Control-Allow-Credentials)"`
	ExposeHeaders      []string      `env:"EXPOSE_HEADERS" long:"expose-headers" description:"List of HTTP headers to be made available to JavaScript browser code (https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Access-Control-Expose-Headers)"`
	MaxAge             time.Duration `env:"MAX_AGE" value-name:"DURATION" long:"max-age" description:"How long results of preflights response can be cached (https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Access-Control-Max-Age)" default:"60s"`
}
