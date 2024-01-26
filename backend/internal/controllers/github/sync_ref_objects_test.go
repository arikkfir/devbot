package github_test

import (
	"context"
	. "github.com/arikkfir/devbot/backend/api/v1"
	act "github.com/arikkfir/devbot/backend/internal/controllers/github"
	"github.com/arikkfir/devbot/backend/internal/util/strings"
	. "github.com/arikkfir/devbot/backend/internal/util/testing"
	"github.com/google/go-github/v56/github"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
	"github.com/onsi/gomega/types"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var _ = Describe("NewSyncGitHubRepositoryRefObjectsAction", func() {
	var k client.Client

	It("should sync stale branches", func(ctx context.Context) {
		const namespace = "default"
		branches := []*github.Branch{
			{Name: github.String("main"), Commit: &github.RepositoryCommit{SHA: github.String(strings.RandomHash(7))}},
			{Name: github.String("b1"), Commit: &github.RepositoryCommit{SHA: github.String(strings.RandomHash(7))}},
			{Name: github.String("b3"), Commit: &github.RepositoryCommit{SHA: github.String(strings.RandomHash(7))}},
			{Name: github.String("b4"), Commit: &github.RepositoryCommit{SHA: github.String(strings.RandomHash(7))}},
			{Name: github.String("b5"), Commit: &github.RepositoryCommit{SHA: github.String(strings.RandomHash(7))}},
		}
		r := &GitHubRepository{
			ObjectMeta: metav1.ObjectMeta{Name: strings.RandomHash(7), Namespace: namespace},
			Spec:       GitHubRepositorySpec{Owner: GitHubOwner, Name: strings.RandomHash(7)},
		}
		refs1 := &GitHubRepositoryRefList{
			Items: []GitHubRepositoryRef{

				// Update owner, name, sha
				{ObjectMeta: metav1.ObjectMeta{Name: "main", Namespace: namespace}, Spec: GitHubRepositoryRefSpec{Ref: "main"}},

				// Fully up-to-date
				{
					ObjectMeta: metav1.ObjectMeta{Name: "b1", Namespace: namespace},
					Spec:       GitHubRepositoryRefSpec{Ref: "b1"},
					Status:     GitHubRepositoryRefStatus{RepositoryOwner: r.Spec.Owner, RepositoryName: r.Spec.Name, CommitSHA: branches[0].Commit.GetSHA()},
				},

				// Stale, to be deleted
				{
					ObjectMeta: metav1.ObjectMeta{Name: "b2", Namespace: namespace},
					Spec:       GitHubRepositoryRefSpec{Ref: "b2"},
					Status:     GitHubRepositoryRefStatus{RepositoryOwner: r.Spec.Owner, RepositoryName: r.Spec.Name, CommitSHA: strings.RandomHash(7)},
				},

				// Update sha
				{
					ObjectMeta: metav1.ObjectMeta{Name: "b3", Namespace: namespace},
					Spec:       GitHubRepositoryRefSpec{Ref: "b3"},
					Status:     GitHubRepositoryRefStatus{RepositoryOwner: r.Spec.Owner, RepositoryName: r.Spec.Name, CommitSHA: strings.RandomHash(7)},
				},

				// Update owner
				{
					ObjectMeta: metav1.ObjectMeta{Name: "b4", Namespace: namespace},
					Spec:       GitHubRepositoryRefSpec{Ref: "b4"},
					Status:     GitHubRepositoryRefStatus{RepositoryOwner: strings.RandomHash(7), RepositoryName: r.Spec.Name, CommitSHA: branches[3].Commit.GetSHA()},
				},

				// Update name
				{
					ObjectMeta: metav1.ObjectMeta{Name: "b6", Namespace: namespace},
					Spec:       GitHubRepositoryRefSpec{Ref: "b6"},
					Status:     GitHubRepositoryRefStatus{RepositoryOwner: r.Spec.Owner, RepositoryName: strings.RandomHash(7), CommitSHA: branches[4].Commit.GetSHA()},
				},
			},
		}
		refsMatches := []types.GomegaMatcher{
			MatchFields(IgnoreExtras, Fields{
				"Status": MatchFields(IgnoreExtras, Fields{
					"RepositoryOwner": Equal(r.Spec.Owner),
					"RepositoryName":  Equal(r.Spec.Name),
					"CommitSHA":       Equal(branches[0].Commit.GetSHA()),
				}),
			}),
			MatchFields(IgnoreExtras, Fields{
				"Status": MatchFields(IgnoreExtras, Fields{
					"RepositoryOwner": Equal(r.Spec.Owner),
					"RepositoryName":  Equal(r.Spec.Name),
					"CommitSHA":       Equal(branches[0].Commit.GetSHA()),
				}),
			}),
			MatchError(apierrors.StatusError{ErrStatus: metav1.Status{Reason: metav1.StatusReasonNotFound}}),
			MatchFields(IgnoreExtras, Fields{
				"Status": MatchFields(IgnoreExtras, Fields{
					"RepositoryOwner": Equal(r.Spec.Owner),
					"RepositoryName":  Equal(r.Spec.Name),
					"CommitSHA":       Equal(branches[0].Commit.GetSHA()),
				}),
			}),
			MatchFields(IgnoreExtras, Fields{
				"Status": MatchFields(IgnoreExtras, Fields{
					"RepositoryOwner": Equal(r.Spec.Owner),
					"RepositoryName":  Equal(r.Spec.Name),
					"CommitSHA":       Equal(branches[0].Commit.GetSHA()),
				}),
			}),
			MatchFields(IgnoreExtras, Fields{
				"Status": MatchFields(IgnoreExtras, Fields{
					"RepositoryOwner": Equal(r.Spec.Owner),
					"RepositoryName":  Equal(r.Spec.Name),
					"CommitSHA":       Equal(branches[0].Commit.GetSHA()),
				}),
			}),
		}

		k = fake.NewClientBuilder().
			WithScheme(scheme).
			WithObjects(r).
			WithLists(refs1).
			WithStatusSubresource(r, &refs1.Items[0], &refs1.Items[1], &refs1.Items[2], &refs1.Items[3], &refs1.Items[4], &refs1.Items[5]).
			Build()

		rr := &GitHubRepository{}
		Expect(k.Get(ctx, client.ObjectKeyFromObject(r), rr)).To(Succeed())
		refs2 := &GitHubRepositoryRefList{}
		Expect(k.List(ctx, refs2)).To(Succeed())
		result, err := act.NewSyncGitHubRepositoryRefObjectsAction(branches, refs2).Execute(ctx, k, rr)
		Expect(err).ToNot(HaveOccurred())
		Expect(result).To(BeNil())

		for i := 0; i < 6; i++ {
			Expect(k.Get(ctx, client.ObjectKeyFromObject(&refs1.Items[i]), &GitHubRepositoryRef{})).To(refsMatches[i])
		}
	})
})
