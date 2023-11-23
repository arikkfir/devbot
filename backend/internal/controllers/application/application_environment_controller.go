package application

import (
	"context"
	apiv1 "github.com/arikkfir/devbot/backend/api/v1"
	"github.com/arikkfir/devbot/backend/internal/util"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type EnvironmentReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

func (r *EnvironmentReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	env := &apiv1.ApplicationEnvironment{}

	if result, err := util.PrepareReconciliation(ctx, r.Client, req, env, "envs.applications.finalizers."+apiv1.GroupVersion.Group); result != nil || err != nil {
		return *result, err
	}

	// TODO: implement application environment controller

	// Search for refs that await deployment
	//for ref, s := range app.Status.Refs {
	//	if s.LastAppliedCommit != s.LatestAvailableCommit || s.LastAppliedCommit == "" {
	//		// Ref is outdated or never deployed, cancel ongoing deployments and create a new one for the latest commit
	//
	//		// If last applied commit is empty, no point checking for old deployments - we'll just deploy
	//		if s.LastAppliedCommit != "" {
	//
	//			// Fetch all deployments
	//			deployments := &apiv1.DeploymentList{}
	//			fields := client.MatchingFields{"spec.application": app.Name, "spec.env": ref}
	//			if err := r.List(ctx, deployments, client.InNamespace(app.Namespace), fields); err != nil {
	//				// Failed listing deployments, report and try again
	//				return ctrl.Result{}, err
	//			}
	//
	//			// For each deployment found, delete it if it's stale (should be all of them)
	//			alreadyDeployed := false
	//			for _, deployment := range deployments.Items {
	//				if deployment.Spec.Application != app.Name || deployment.Spec.Ref != ref {
	//					// Safeguard when listing deployments by fields returns partial matches
	//					log.FromContext(ctx).WithValues("fields", fields).Info("Found deployment with partial match")
	//				} else if deployment.Spec.Commit == s.LatestAvailableCommit {
	//					// Deployment for this env already up-to-date, nothing to do
	//					alreadyDeployed = true
	//				} else if deployment.Status {
	//				} else if deployment.Status.Status == status.TerminatingStatus {
	//					// Stale deployment; already terminating, nothing to do
	//					return ctrl.Result{Requeue: true, RequeueAfter: timeToWaitForDeploymentTermination}, nil
	//				} else if err := r.Delete(ctx, &deployment); err != nil {
	//					// Stale deployment; not terminating, delete and requeue for later
	//					return ctrl.Result{}, err
	//				}
	//			}
	//
	//			// If by any chance we did find a deployment for this latest commit then we have nothing to do
	//			if alreadyDeployed {
	//				continue
	//			}
	//		}
	//
	//		// If we got here, there are no deployments for this env, create a new one
	//		deployment := &apiv1.Deployment{
	//			TypeMeta: metav1.TypeMeta{APIVersion: apiv1.GroupVersion.String(), Kind: "Deployment"},
	//			ObjectMeta: metav1.ObjectMeta{
	//				Name:      util.RandomHash(16),
	//				Namespace: app.Namespace,
	//				OwnerReferences: []metav1.OwnerReference{
	//					*metav1.NewControllerRef(app, apiv1.GroupVersion.WithKind("Application")),
	//				},
	//			},
	//			Spec: apiv1.DeploymentSpec{
	//				Application: app.Name,
	//				Ref:         ref,
	//				Commit:      s.LatestAvailableCommit,
	//			},
	//		}
	//		if err := r.Create(ctx, deployment); err != nil {
	//			// Failed creating deployment, report and try again
	//			return ctrl.Result{}, err
	//		}
	//	}
	//}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *EnvironmentReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&apiv1.ApplicationEnvironment{}).
		Complete(r)
}
