package github

import (
	"context"
	"errors"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

type RepositoryPoller struct {
	mgr manager.Manager
}

func NewRepositoryPoller(mgr manager.Manager) *RepositoryPoller {
	return &RepositoryPoller{mgr: mgr}
}

func (r *RepositoryPoller) Start(ctx context.Context) error {
	// TODO: fetch all GitHubRepository objects; for each:
	//		 - If repository has "Invalid" condition -> skip
	//		 - Extract repo's refresh interval from its spec
	//		 - If repo has annotation "devbot.kfirs.com/last-polled":
	//		   - If it's value is later than "Now() - refreshInterval" -> skip
	//		   - Otherwise, update the annotation to "Now()"
	//		   - This will trigger a reconciliation :)
	<-ctx.Done()
	err := ctx.Err()
	if errors.Is(err, context.Canceled) {
		return nil
	}
	return err
}
