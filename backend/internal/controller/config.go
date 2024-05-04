package controller

// Config contains the configuration for the controller. This struct is parsed by Cobra and Viper from CLI and
// environment variables.
type Config struct {
	DisableJSONLogging   bool
	LogLevel             string
	MetricsAddr          string
	HealthProbeAddr      string
	EnableLeaderElection bool
	PodNamespace         string
	PodName              string
	ContainerName        string
}
