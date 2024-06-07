package main

import (
	"context"
	"os"
	"path/filepath"

	"github.com/arikkfir/command"

	"github.com/arikkfir/devbot/backend/internal/devctl"
)

func main() {

	// Create command structure
	cmd := command.MustNew(
		filepath.Base(os.Args[0]),
		"Opinionated development control flow tool.",
		`Devctl allows you to bootstrap & interact with Devbot installations in Kubernetes clusters.
It provides commands for bootstrapping Devbot onto a new or existing GitOps repository, as well as inspecting &
manipulating running Devbot installations.`,
		&devctl.RootExecutor{},
		command.MustNew(
			"bootstrap",
			"BootstrapExecutor devbot",
			`This command will create Devbot manifests.`,
			&devctl.BootstrapExecutor{},
			command.MustNew(
				"github",
				"BootstrapExecutor devbot in a GitHub repository",
				`This command will create Devbot manifests in your GitHub repository.`,
				&devctl.BootstrapGitHubExecutor{Visibility: "public"},
			),
		),
	)

	// Prepare a context that gets canceled if OS termination signals are sent
	ctx, cancel := context.WithCancel(command.SetupSignalHandler())
	defer cancel()

	// Execute the correct command
	command.Execute(ctx, os.Stderr, cmd, os.Args, command.EnvVarsArrayToMap(os.Environ()))
}
