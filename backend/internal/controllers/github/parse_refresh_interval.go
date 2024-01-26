package github

import (
	"context"
	apiv1 "github.com/arikkfir/devbot/backend/api/v1"
	"github.com/arikkfir/devbot/backend/pkg/k8s/reconcile"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"time"
)

func NewParseRefreshInterval(refreshInterval *time.Duration) reconcile.Action {
	return func(_ context.Context, _ client.Client, o client.Object) (*ctrl.Result, error) {
		r := o.(*apiv1.GitHubRepository)
		if duration, err := time.ParseDuration(r.Spec.RefreshInterval); err != nil {
			r.Status.SetInvalidDueToInvalidRefreshInterval("Refresh interval is invalid: %+v", err)
			return reconcile.DoNotRequeue()
		} else if duration.Seconds() < 5 {
			r.Status.SetInvalidDueToInvalidRefreshInterval("Refresh interval '%s' is too low (must not be less than 5 seconds)", r.Spec.RefreshInterval)
			return reconcile.DoNotRequeue()
		} else {
			*refreshInterval = duration
			return reconcile.Continue()
		}
	}
}
