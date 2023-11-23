package application

import (
	"context"
	apiv1 "github.com/arikkfir/devbot/backend/api/v1"
	"github.com/arikkfir/devbot/backend/internal/util"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"time"
)

// Reconciler reconciles an Application object
type Reconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	app := &apiv1.Application{}

	if result, err := util.PrepareReconciliation(ctx, r.Client, req, app, "applications.finalizers."+apiv1.GroupVersion.Group); result != nil || err != nil {
		return *result, nil
	}

	// TODO: implement application controller

	return ctrl.Result{Requeue: true, RequeueAfter: 30 * time.Second}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *Reconciler) SetupWithManager(mgr ctrl.Manager) error {
	// TODO: remove comment
	//deploymentAppIndexer := func(obj client.Object) []string {
	//	return []string{obj.(*apiv1.Deployment).Spec.Application}
	//}
	//deploymentAppField := "spec.application"
	//if err := mgr.GetFieldIndexer().IndexField(ctx, &apiv1.Deployment{}, deploymentAppField, deploymentAppIndexer); err != nil {
	//	return err
	//}
	return ctrl.NewControllerManagedBy(mgr).
		For(&apiv1.Application{}).
		Complete(r)
}
