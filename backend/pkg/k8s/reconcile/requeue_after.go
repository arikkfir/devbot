package reconcile

import (
	"context"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"time"
)

func NewRequeueAfterAction(refreshInterval time.Duration) Action {
	return func(_ context.Context, _ client.Client, _ client.Object) (*ctrl.Result, error) {
		return RequeueAfter(refreshInterval)
	}
}
