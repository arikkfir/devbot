package deployment

import (
	"context"
	apiv1 "github.com/arikkfir/devbot/backend/api/v1"
	"github.com/arikkfir/devbot/backend/pkg/k8s/reconcile"
	"github.com/go-git/go-git/v5"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"time"
)

const (
	kustomizeBinaryFilePath = "/usr/local/bin/kustomize"
	yqBinaryFilePath        = "/usr/local/bin/yq"
	kubectlBinaryFilePath   = "/usr/local/bin/kubectl"
)

var (
	Finalizer = "deployments.finalizers." + apiv1.GroupVersion.Group
)

type Reconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	var deployment *apiv1.Deployment
	var env *apiv1.Environment
	var app *apiv1.Application
	var cloneDir, gitURL string
	var ghRepo *git.Repository
	var resourcesFile string
	reconciliation := reconcile.Reconciliation{
		Actions: []reconcile.Action{
			reconcile.NewSaveObjectReferenceAction(&deployment),
			reconcile.NewFinalizeAction(Finalizer, nil),
			reconcile.NewAddFinalizerAction(Finalizer),
			reconcile.NewGetControllerAction(false, deployment, env),
			reconcile.NewGetControllerAction(false, env, app),
			NewPrepareCloneDirectoryAction(&cloneDir, &gitURL),
			NewCloneRepositoryAction(gitURL, &ghRepo),
			NewBakeAction(app, env, ghRepo, &resourcesFile),
			NewApplyAction(resourcesFile),
			reconcile.NewRequeueAfterAction(time.Minute), // TODO: make this configurable
		},
	}
	return reconciliation.Execute(ctx, r.Client, req, &apiv1.Deployment{})
}

// SetupWithManager sets up the controller with the Manager.
// TODO: watch & reconcile on changes to controlling application, repositories and branches
func (r *Reconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&apiv1.Deployment{}).
		Complete(r)
}
