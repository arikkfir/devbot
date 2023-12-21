package controllers

import "github.com/arikkfir/devbot/backend/internal/config"

type Config struct {
	config.CommandConfig
	MetricsAddr          string `env:"METRICS_BIND_ADDR" value-name:"ADDR" long:"metrics-bind-address" description:"The address the metric endpoint binds to" default:":8000"`
	HealthProbeAddr      string `env:"HEALTH_PROBE_BIND_ADDR" value-name:"ADDR" long:"health-probe-bind-address" description:"The address the probe endpoint binds to" default:":9000"`
	EnableLeaderElection bool   `env:"ENABLE_LEADER_ELECTION" value-name:"VALUE" long:"leader-elect" description:"Enable leader election, ensuring only one controller is active"`
}
