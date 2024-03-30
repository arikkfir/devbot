package config

var (
	PodName      = "unknown"
	PodNamespace = "unknown"
	Version      = "unknown"
	Image        = "unknown"
)

type CommandConfig struct {
	DevMode      bool   `env:"DEV_MODE" long:"dev-mode" description:"Development mode"`
	LogLevel     string `env:"LOG_LEVEL" value-name:"LEVEL" long:"log-level" description:"Log level" default:"info" enum:"trace,debug,info,warn,error,fatal,panic"`
	PodName      string `env:"POD_NAME" long:"pod-name" description:"Name of the pod"`
	PodNamespace string `env:"POD_NAMESPACE" long:"pod-namespace" description:"Namespace of the pod"`
}
