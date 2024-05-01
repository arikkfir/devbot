package e2e_test

import (
	apiv1 "github.com/arikkfir/devbot/backend/api/v1"
	. "github.com/arikkfir/devbot/backend/e2e/expectations"
	"github.com/arikkfir/devbot/backend/internal/util/lang"
	. "github.com/arikkfir/devbot/backend/internal/util/testing/justest"
	. "k8s.io/apimachinery/pkg/apis/meta/v1"
	"regexp"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"testing"
	"time"
)

func TestRepositoryRefreshIntervalParsing(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		defaultBranch   string
		refreshInterval string
		invalid         *ConditionE
		unauthenticated *ConditionE
		stale           *ConditionE
	}{
		"Empty": {
			defaultBranch:   "",
			refreshInterval: "",
			invalid: &ConditionE{
				Type:    apiv1.Invalid,
				Status:  lang.Ptr(string(ConditionTrue)),
				Reason:  regexp.MustCompile(regexp.QuoteMeta(apiv1.UnknownRepositoryType)),
				Message: regexp.MustCompile(regexp.QuoteMeta("Unknown repository type")),
			},
			unauthenticated: &ConditionE{
				Type:    apiv1.Unauthenticated,
				Status:  lang.Ptr(string(ConditionTrue)),
				Reason:  regexp.MustCompile(regexp.QuoteMeta(apiv1.Invalid)),
				Message: regexp.MustCompile(regexp.QuoteMeta("Unknown repository type")),
			},
			stale: &ConditionE{
				Type:    apiv1.Stale,
				Status:  lang.Ptr(string(ConditionUnknown)),
				Reason:  regexp.MustCompile(regexp.QuoteMeta(apiv1.Invalid)),
				Message: regexp.MustCompile(regexp.QuoteMeta("Unknown repository type")),
			},
		},
		"Invalid": {
			refreshInterval: "abc",
			invalid: &ConditionE{
				Type:    apiv1.Invalid,
				Status:  lang.Ptr(string(ConditionTrue)),
				Reason:  regexp.MustCompile(regexp.QuoteMeta(apiv1.InvalidRefreshInterval)),
				Message: regexp.MustCompile(regexp.QuoteMeta(`time: invalid duration "abc"`)),
			},
			unauthenticated: nil,
			stale: &ConditionE{
				Type:    apiv1.Stale,
				Status:  lang.Ptr(string(ConditionUnknown)),
				Reason:  regexp.MustCompile(regexp.QuoteMeta(apiv1.Invalid)),
				Message: regexp.MustCompile(regexp.QuoteMeta(`time: invalid duration "abc"`)),
			},
		},
		"TooLow": {
			refreshInterval: "2s",
			invalid: &ConditionE{
				Type:    apiv1.Invalid,
				Status:  lang.Ptr(string(ConditionTrue)),
				Reason:  regexp.MustCompile(regexp.QuoteMeta(apiv1.InvalidRefreshInterval)),
				Message: regexp.MustCompile(regexp.QuoteMeta(`refresh interval '2s' is too low \(must not be less than 5s\)`)),
			},
			unauthenticated: nil,
			stale: &ConditionE{
				Type:    apiv1.Stale,
				Status:  lang.Ptr(string(ConditionUnknown)),
				Reason:  regexp.MustCompile(regexp.QuoteMeta(apiv1.Invalid)),
				Message: regexp.MustCompile(regexp.QuoteMeta(`refresh interval '2s' is too low \(must not be less than 5s\)`)),
			},
		},
		"Valid": {
			refreshInterval: "6s",
			invalid: &ConditionE{
				Type:    apiv1.Invalid,
				Status:  lang.Ptr(string(ConditionTrue)),
				Reason:  regexp.MustCompile(regexp.QuoteMeta(apiv1.UnknownRepositoryType)),
				Message: regexp.MustCompile(regexp.QuoteMeta(`Unknown repository type`)),
			},
			unauthenticated: &ConditionE{
				Type:    apiv1.Unauthenticated,
				Status:  lang.Ptr(string(ConditionTrue)),
				Reason:  regexp.MustCompile(regexp.QuoteMeta(apiv1.Invalid)),
				Message: regexp.MustCompile(regexp.QuoteMeta(`Unknown repository type`)),
			},
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			e2e := NewE2E(t)
			ns := e2e.K.CreateNamespace(e2e.Ctx, t)

			kRepoName := ns.CreateRepository(e2e.Ctx, t, apiv1.RepositorySpec{RefreshInterval: tc.refreshInterval})
			defaultBranch := tc.defaultBranch
			if defaultBranch == "" {
				defaultBranch = "main"
			}
			With(t).Verify(func(t T) {
				repositoryExpectation := RepositoryE{
					Name: kRepoName,
					Status: RepositoryStatusE{
						Conditions: map[string]*ConditionE{
							apiv1.FailedToInitialize: nil,
							apiv1.Finalizing:         nil,
							apiv1.Invalid:            tc.invalid,
							apiv1.Stale:              tc.stale,
							apiv1.Unauthenticated:    tc.unauthenticated,
						},
						DefaultBranch: tc.defaultBranch,
					},
				}
				repo := &apiv1.Repository{}
				With(t).Verify(e2e.K.Client.Get(e2e.Ctx, client.ObjectKey{Namespace: ns.Name, Name: kRepoName}, repo)).Will(Succeed()).OrFail()
				With(t).Verify(*repo).Will(EqualTo(repositoryExpectation).Using(RepositoryComparator)).OrFail()
			}).Will(Succeed()).Within(30*time.Second, 1*time.Second)
		})
	}
}
