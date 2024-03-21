package repository_test

import (
	"context"
	apiv1 "github.com/arikkfir/devbot/backend/api/v1"
	. "github.com/arikkfir/devbot/backend/internal/util/testing"
	. "github.com/arikkfir/devbot/backend/internal/util/testing/justest"
	. "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"testing"
	"time"
)

func TestRepositoryRefreshIntervalParsing(t *testing.T) {
	testCases := map[string]struct {
		refreshInterval string
		invalid         *Condition
		unauthenticated *Condition
		stale           *Condition
	}{
		"Empty": {
			refreshInterval: "",
			invalid:         &Condition{Status: ConditionTrue, Reason: apiv1.UnknownRepositoryType, Message: "Unknown repository type"},
			unauthenticated: &Condition{Status: ConditionTrue, Reason: apiv1.Invalid, Message: "Unknown repository type"},
			stale:           &Condition{Status: ConditionUnknown, Reason: apiv1.Invalid, Message: "Unknown repository type"},
		},
		"Invalid": {
			refreshInterval: "abc",
			invalid:         &Condition{Status: ConditionTrue, Reason: apiv1.InvalidRefreshInterval, Message: `time: invalid duration "abc"`},
			unauthenticated: nil,
			stale:           &Condition{Status: ConditionUnknown, Reason: apiv1.Invalid, Message: `time: invalid duration "abc"`},
		},
		"TooLow": {
			refreshInterval: "2s",
			invalid:         &Condition{Status: ConditionTrue, Reason: apiv1.InvalidRefreshInterval, Message: `refresh interval '2s' is too low \(must not be less than 5s\)`},
			unauthenticated: nil,
			stale:           &Condition{Status: ConditionUnknown, Reason: apiv1.Invalid, Message: `refresh interval '2s' is too low \(must not be less than 5s\)`},
		},
		"Valid": {
			refreshInterval: "6s",
			invalid:         &Condition{Status: ConditionTrue, Reason: apiv1.UnknownRepositoryType, Message: `Unknown repository type`},
			unauthenticated: &Condition{Status: ConditionTrue, Reason: apiv1.Invalid, Message: `Unknown repository type`},
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			t.Cleanup(cancel)

			k := NewKubernetes(t)
			ns := k.CreateNamespace(t, ctx)
			repoObjName := ns.CreateRepository(t, ctx, apiv1.RepositorySpec{
				RefreshInterval: tc.refreshInterval,
			})

			For(t).Expect(func(t JustT) {
				repo := &apiv1.Repository{}
				For(t).Expect(k.Client.Get(ctx, client.ObjectKey{Namespace: ns.Name, Name: repoObjName}, repo)).Will(Succeed())
				For(t).Expect(repo.Status.GetFailedToInitializeCondition()).Will(BeNil())
				For(t).Expect(repo.Status.GetFinalizingCondition()).Will(BeNil())
				For(t).Expect(repo.Status.GetInvalidCondition()).Will(EqualCondition(tc.invalid))
				For(t).Expect(repo.Status.GetUnauthenticatedCondition()).Will(EqualCondition(tc.unauthenticated))
				For(t).Expect(repo.Status.GetStaleCondition()).Will(EqualCondition(tc.stale))
			}).Will(Eventually(Succeed()).Within(10 * time.Second).ProbingEvery(100 * time.Millisecond))
		})
	}
}
