package e2e

import (
	"fmt"
	apiv1 "github.com/arikkfir/devbot/backend/api/v1"
	"github.com/onsi/gomega/types"
)

type matchState struct {
	repo   apiv1.GitHubRepository
	branch string
	sha    string
}

func (m *matchState) Match(actual interface{}) (success bool, err error) {
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

func (m *matchState) FailureMessage(actual interface{}) (message string) {
	return "Ref is ready"
}

func (m *matchState) NegatedFailureMessage(actual interface{}) (message string) {
	return "Ref is ready"
}

func BeReady(repo apiv1.GitHubRepository, branch, sha string) types.GomegaMatcher {
	return &matchState{
		repo:   repo,
		branch: branch,
		sha:    sha,
	}
}
