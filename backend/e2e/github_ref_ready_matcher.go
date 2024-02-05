package e2e

import (
	"fmt"
	apiv1 "github.com/arikkfir/devbot/backend/api/v1"
	"github.com/onsi/gomega/types"
)

type GitHubRepositoryRefReadyMatcher struct {
	repo   apiv1.GitHubRepository
	branch string
	sha    string
}

func (m *GitHubRepositoryRefReadyMatcher) Match(actual interface{}) (success bool, err error) {
	if ref, ok := actual.(apiv1.GitHubRepositoryRef); ok {
		return ref.Spec.Ref == m.branch &&
			ref.Status.RepositoryOwner == m.repo.Spec.Owner &&
			ref.Status.RepositoryName == m.repo.Spec.Name &&
			ref.Status.CommitSHA == m.sha &&
			ref.Status.GetFailedToInitializeCondition() == nil &&
			ref.Status.GetFinalizingCondition() == nil &&
			ref.Status.GetInvalidCondition() == nil &&
			ref.Status.GetStaleCondition() == nil &&
			ref.Status.GetUnauthenticatedCondition() == nil, nil
	}
	return false, fmt.Errorf("value is not a GitHubRepositoryRef")
}

func (m *GitHubRepositoryRefReadyMatcher) FailureMessage(_ interface{}) (message string) {
	return "Ref is not ready"
}

func (m *GitHubRepositoryRefReadyMatcher) NegatedFailureMessage(_ interface{}) (message string) {
	return "Ref is ready"
}

func BeReady(repo apiv1.GitHubRepository, branch, sha string) types.GomegaMatcher {
	return &GitHubRepositoryRefReadyMatcher{
		repo:   repo,
		branch: branch,
		sha:    sha,
	}
}
