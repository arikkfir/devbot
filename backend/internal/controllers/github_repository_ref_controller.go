package controllers

import (
	"context"
	apiv1 "github.com/arikkfir/devbot/backend/api/v1"
	"github.com/arikkfir/devbot/backend/internal/util"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type GitHubRepositoryRefReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

func (r *GitHubRepositoryRefReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	ref := &apiv1.GitHubRepositoryRef{}

	if result, err := util.PrepareReconciliation(ctx, r.Client, req, ref, "refs.github.repositories.finalizers."+apiv1.GroupVersion.Group); result != nil || err != nil {
		return *result, err
	}

	// TODO: test rate limiter
	// TODO: set stale/current condition based on whether ref's commit SHA equals that ref's latest commit SHA in repo

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *GitHubRepositoryRefReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&apiv1.GitHubRepositoryRef{}).
		Complete(r)
}
