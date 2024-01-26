package environment

import (
	"context"
	apiv1 "github.com/arikkfir/devbot/backend/api/v1"
	"github.com/arikkfir/devbot/backend/pkg/k8s/reconcile"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"time"
)

var (
	Finalizer = "environments.finalizers." + apiv1.GroupVersion.Group
)

type Reconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	var env *apiv1.Environment
	var app *apiv1.Application
	reconciliation := reconcile.Reconciliation{
		Actions: []reconcile.Action{
			reconcile.NewSaveObjectReferenceAction(&env),
			reconcile.NewFinalizeAction(Finalizer, nil),
			reconcile.NewAddFinalizerAction(Finalizer),
			reconcile.NewGetControllerAction(false, env, app),
			NewSyncSourcesAction(app),
			reconcile.NewRequeueAfterAction(time.Minute),
		},
	}
	return reconciliation.Execute(ctx, r.Client, req, &apiv1.Environment{})
}

// SetupWithManager sets up the controller with the Manager.
// TODO: watch & reconcile on changes to controlling application, repositories and branches
func (r *Reconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&apiv1.Environment{}).
		Complete(r)
}
