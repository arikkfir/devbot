//go:generate controller-gen rbac:roleName=applications-controller paths="." output:rbac:artifacts:config=/tmp/applications-controller-rbac
//go:generate mv /tmp/applications-controller-rbac/role.yaml ../../../../deploy/applications-controller-rbac.yaml
package controller

import (
	"context"
	api "github.com/arikkfir/devbot/backend/applications-controller/api/v1"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// ApplicationReconciler reconciles an Application object
type ApplicationReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=devbot.kfirs.com,resources=applications,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=devbot.kfirs.com,resources=applications/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=devbot.kfirs.com,resources=applications/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *ApplicationReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	// TODO(user): your logic here

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ApplicationReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&api.Application{}).
		Complete(r)
}
