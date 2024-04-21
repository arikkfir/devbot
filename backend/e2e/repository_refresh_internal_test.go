package e2e_test

import (
	apiv1 "github.com/arikkfir/devbot/backend/api/v1"
	. "github.com/arikkfir/devbot/backend/e2e/expectations"
	"github.com/arikkfir/devbot/backend/internal/util/lang"
	. "github.com/arikkfir/devbot/backend/internal/util/testing"
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
		refreshInterval string
		invalid         *ConditionE
		unauthenticated *ConditionE
		stale           *ConditionE
	}{
		"Empty": {
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

			ns := K(t).CreateNamespace(t)
			kRepoName := ns.CreateRepository(t, apiv1.RepositorySpec{RefreshInterval: tc.refreshInterval})
			For(t).Expect(func(t TT) {
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
						DefaultBranch: "main",
					},
				}
				repo := &apiv1.Repository{}
				For(t).Expect(K(t).Client.Get(t, client.ObjectKey{Namespace: ns.Name, Name: kRepoName}, repo)).Will(Succeed()).OrFail()
				For(t).Expect(*repo).Will(CompareTo(repositoryExpectation).Using(RepositoryComparator)).OrFail()
			}).Will(Eventually(Succeed()).Within(10 * time.Second).ProbingEvery(100 * time.Millisecond)).OrFail()
		})
	}
}
