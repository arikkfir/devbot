package github_test

import (
	"context"
	. "github.com/arikkfir/devbot/backend/api/v1"
	act "github.com/arikkfir/devbot/backend/internal/controllers/github"
	"github.com/arikkfir/devbot/backend/internal/util/strings"
	. "github.com/arikkfir/devbot/backend/internal/util/testing"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"time"
)

var _ = Describe("NewParseRefreshInterval", func() {
	var k client.Client

	var namespace, repoObjName string
	BeforeEach(func(ctx context.Context) { namespace, repoObjName = "default", strings.RandomHash(7) })

	When("refresh interval is invalid", func() {
		DescribeTable(
			"should set conditions and abort",
			func(ctx context.Context, refreshIntervalValue string) {
				r := &GitHubRepository{
					ObjectMeta: metav1.ObjectMeta{Name: repoObjName, Namespace: namespace},
					Spec:       GitHubRepositorySpec{RefreshInterval: refreshIntervalValue},
				}
				k = fake.NewClientBuilder().WithScheme(scheme).WithObjects(r).WithStatusSubresource(r).Build()

				rr := &GitHubRepository{}
				Expect(k.Get(ctx, client.ObjectKeyFromObject(r), rr)).To(Succeed())
				var refreshInterval time.Duration
				result, err := act.NewParseRefreshInterval(&refreshInterval).Execute(ctx, k, rr)
				Expect(err).ToNot(HaveOccurred())
				Expect(result).To(Equal(&ctrl.Result{}))
				Expect(refreshInterval).To(Equal(time.Duration(0)))

				rrr := &GitHubRepository{}
				Expect(k.Get(ctx, client.ObjectKeyFromObject(r), rrr)).To(Succeed())
				Expect(rrr.Status.GetInvalidCondition()).To(BeTrueDueTo(InvalidRefreshInterval))
				Expect(rrr.Status.GetStaleCondition()).To(BeUnknownDueTo(Invalid))
				Expect(rrr.Status.GetUnauthenticatedCondition()).To(BeUnknownDueTo(Invalid))
			},
			Entry("when refresh interval is empty", ""),
			Entry("when refresh interval is just a number", "5"),
			Entry("when refresh interval is too low", "3s"),
		)
	})

	When("refresh interval is valid and not too low", func() {
		It("should store the refresh internal and continue", func(ctx context.Context) {
			r := &GitHubRepository{
				ObjectMeta: metav1.ObjectMeta{Name: repoObjName, Namespace: namespace},
				Spec:       GitHubRepositorySpec{RefreshInterval: "10s"},
			}
			k = fake.NewClientBuilder().WithScheme(scheme).WithObjects(r).WithStatusSubresource(r).Build()

			rr := &GitHubRepository{}
			Expect(k.Get(ctx, client.ObjectKeyFromObject(r), rr)).To(Succeed())
			var refreshInterval time.Duration
			result, err := act.NewParseRefreshInterval(&refreshInterval).Execute(ctx, k, rr)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(BeNil())
			Expect(refreshInterval).To(Equal(10 * time.Second))

			rrr := &GitHubRepository{}
			Expect(k.Get(ctx, client.ObjectKeyFromObject(r), rrr)).To(Succeed())
			Expect(rrr.Status.GetInvalidCondition()).To(BeNil())
			Expect(rrr.Status.GetStaleCondition()).To(BeNil())
			Expect(rrr.Status.GetUnauthenticatedCondition()).To(BeNil())
		})
	})
})
