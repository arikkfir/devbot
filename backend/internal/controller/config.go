package controller

// Config contains the configuration for the controller. This struct is parsed by Cobra and Viper from CLI and
// environment variables.
type Config struct {
	DisableJSONLogging   bool   `desc:"Disable JSON logging."`
	LogLevel             string `config:"required" desc:"Log level, must be one of: trace,debug,info,warn,error,fatal,panic"`
	MetricsAddr          string `desc:"Address the metrics endpoint should bind to."`
	HealthProbeAddr      string `desc:"Address the health endpoint should bind to"`
	EnableLeaderElection bool   `desc:"Enable leader election, ensuring only one controller is active"`
	PodNamespace         string `config:"required" desc:"Namespace of the controller pod (usually provided via downward API)"`
	PodName              string `config:"required" desc:"Name of the controller pod (usually provided via downward API)"`
	ContainerName        string `desc:"Name of the controller container"`
}
