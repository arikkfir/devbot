package main

import (
	"context"
	"github.com/arikkfir/command"
	"github.com/arikkfir/devbot/backend/internal/devctl"
	"github.com/arikkfir/devbot/backend/internal/util/logging"
	"os"
	"path/filepath"
)

func main() {

	// Create command structure
	cmd := command.MustNew(
		filepath.Base(os.Args[0]),
		"Opinionated development control flow tool.",
		`Devctl allows you to bootstrap & interact with Devbot installations in Kubernetes clusters.
It provides commands for bootstrapping Devbot onto a new or existing GitOps repository, as well as inspecting &
manipulating running Devbot installations.`,
		nil,
		[]command.PreRunHook{&logging.InitHook{DisableJSONLogging: true, LogLevel: "info"}, &logging.SentryInitHook{}},
		[]command.PostRunHook{&logging.SentryFlushHook{}},
		command.MustNew(
			"bootstrap",
			"BootstrapExecutor devbot",
			`This command will create Devbot manifests.`,
			nil,
			nil,
			nil,
			command.MustNew(
				"github",
				"BootstrapExecutor devbot in a GitHub repository",
				`This command will create Devbot manifests in your GitHub repository.`,
				&devctl.GitHubBootstrapAction{Visibility: "public"},
				nil,
				nil,
			),
		),
	)

	// Prepare a context that gets canceled if OS termination signals are sent
	ctx, cancel := context.WithCancel(command.SetupSignalHandler())
	defer cancel()

	// Execute the correct command
	os.Exit(int(command.Execute(ctx, os.Stderr, cmd, os.Args, command.EnvVarsArrayToMap(os.Environ()))))
}
