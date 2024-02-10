package github

import (
	"context"
	v1 "github.com/arikkfir/devbot/backend/api/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"time"
)

const polledAnnotationName = "devbot.kfirs.com/last-polled"

type RepositoryPoller struct {
	mgr manager.Manager
}

func NewRepositoryPoller(mgr manager.Manager) *RepositoryPoller {
	return &RepositoryPoller{mgr: mgr}
}

func (r *RepositoryPoller) fetchGitHubRepositories(ctx context.Context) ([]v1.GitHubRepository, error) {
	reposList := &v1.GitHubRepositoryList{}
	if err := r.mgr.GetClient().List(ctx, &v1.GitHubRepositoryList{}); err != nil {
		return nil, err
	}
	return reposList.Items, nil
}

func (r *RepositoryPoller) shouldRepositoryBePolled(ctx context.Context, repo *v1.GitHubRepository) bool {
	if repo.Status.IsFailedToInitialize() {
		log.FromContext(ctx).Info("Skipping initializing GitHub repository", "namespace", repo.Namespace, "name", repo.Name)
		return false
	} else if repo.Status.IsFinalizing() {
		log.FromContext(ctx).Info("Skipping finalizing GitHub repository", "namespace", repo.Namespace, "name", repo.Name)
		return false
	} else if repo.Status.IsInvalid() {
		log.FromContext(ctx).Info("Skipping invalid GitHub repository", "namespace", repo.Namespace, "name", repo.Name)
		return false
	} else if repo.Status.IsStale() {
		log.FromContext(ctx).Info("Skipping stale GitHub repository", "namespace", repo.Namespace, "name", repo.Name)
		return false
	} else if repo.Status.IsUnauthenticated() {
		log.FromContext(ctx).Info("Skipping unauthenticated GitHub repository", "namespace", repo.Namespace, "name", repo.Name)
		return false
	} else {
		return true
	}
}

func (r *RepositoryPoller) pollRepository(ctx context.Context, now time.Time, repo *v1.GitHubRepository) {
	var lastPolled time.Time

	annotations := repo.ObjectMeta.GetAnnotations()
	if annotations == nil {
		lastPolled = now
	} else if annotationValue, ok := annotations[polledAnnotationName]; !ok {
		lastPolled = now
	} else if timeValue, err := time.Parse(time.RFC3339, annotationValue); err != nil {
		log.FromContext(ctx).Error(err, "Failed to parse last-polled annotation", "namespace", repo.Namespace, "name", repo.Name, "annotationName", polledAnnotationName, "annotationValue", annotationValue)
		return
	} else {
		lastPolled = timeValue
	}

	refreshInterval, err := time.ParseDuration(repo.Spec.RefreshInterval)
	if err != nil {
		log.FromContext(ctx).Error(err, "Failed to parse refresh interval (this should not happen, as repository is not marked as Invalid!)", "namespace", repo.Namespace, "name", repo.Name, "refreshInterval", repo.Spec.RefreshInterval)
		return
	}

	if now.Sub(lastPolled) > refreshInterval {
		if annotations == nil {
			annotations = make(map[string]string)
		}
		annotations[polledAnnotationName] = now.Format(time.RFC3339)
		repo.ObjectMeta.SetAnnotations(annotations)
		if err := r.mgr.GetClient().Update(ctx, repo); err != nil {
			if apierrors.IsNotFound(err) {
				return
			} else if apierrors.IsConflict(err) {
				return
			} else {
				log.FromContext(ctx).Error(err, "Failed to update last-polled annotation", "namespace", repo.Namespace, "name", repo.Name)
			}
		}
	}
}

func (r *RepositoryPoller) Start(ctx context.Context) error {
	ticker := time.NewTicker(v1.MinGitHubRepositoryRefreshInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			now := time.Now()
			repos, err := r.fetchGitHubRepositories(ctx)
			if err != nil {
				log.FromContext(ctx).Error(err, "Failed to fetch GitHub repositories")
				continue
			}
			for _, repo := range repos {
				if r.shouldRepositoryBePolled(ctx, &repo) {
					r.pollRepository(ctx, now, &repo)
				}
			}
		}
	}
}
