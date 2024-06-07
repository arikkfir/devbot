package devctl

import "context"

// BootstrapExecutor is the command executor for the "devbot bootstrap" command.
type BootstrapExecutor struct{}

func (c *BootstrapExecutor) PreRun(ctx context.Context) error { return nil }

func (c *BootstrapExecutor) Run(ctx context.Context) error { return nil }
