package application_test

import (
	"context"
	. "github.com/arikkfir/devbot/backend/api/v1"
	"github.com/arikkfir/devbot/backend/internal/controllers/application"
	strings2 "github.com/arikkfir/devbot/backend/internal/util/strings"
	. "github.com/arikkfir/devbot/backend/internal/util/testing"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"slices"
	"strings"
)

var _ = Describe("NewCreateMissingEnvironmentsAction", func() {
	var namespace, appName string

	newGitHubRepository := func(name string) GitHubRepository {
		return GitHubRepository{
			ObjectMeta: metav1.ObjectMeta{Name: strings2.RandomHash(7), Namespace: namespace},
			Spec:       GitHubRepositorySpec{Owner: GitHubOwner, Name: name},
		}
	}
	newGitHubRepositoryRef := func(controller *GitHubRepository, ref string) GitHubRepositoryRef {
		return GitHubRepositoryRef{
			ObjectMeta: metav1.ObjectMeta{
				Name:            strings2.RandomHash(7),
				Namespace:       namespace,
				OwnerReferences: []metav1.OwnerReference{*metav1.NewControllerRef(controller, GitHubRepositoryGVK)},
			},
			Spec: GitHubRepositoryRefSpec{Ref: ref},
		}
	}
	newAppSpecRepository := func(repoName, missingBranchStrategy string) ApplicationSpecRepository {
		return ApplicationSpecRepository{
			RepositoryReferenceWithOptionalNamespace: RepositoryReferenceWithOptionalNamespace{
				APIVersion: GitHubRepositoryGVK.GroupVersion().String(),
				Kind:       GitHubRepositoryGVK.Kind,
				Name:       repoName,
				Namespace:  namespace,
			},
			MissingBranchStrategy: missingBranchStrategy,
		}
	}

	var k client.WithWatch
	BeforeEach(func(ctx context.Context) {
		namespace = "default"
		appName = strings2.RandomHash(7)
		ghRepoList := &GitHubRepositoryList{Items: []GitHubRepository{
			newGitHubRepository("repo1"),
			newGitHubRepository("repo2"),
			newGitHubRepository("repo3"),
		}}
		ghRepoRefList := &GitHubRepositoryRefList{Items: []GitHubRepositoryRef{
			newGitHubRepositoryRef(&ghRepoList.Items[0], "b1"),
			newGitHubRepositoryRef(&ghRepoList.Items[0], "b2"),
			newGitHubRepositoryRef(&ghRepoList.Items[1], "b1"),
			newGitHubRepositoryRef(&ghRepoList.Items[2], "b2"),
		}}
		envList := &EnvironmentList{}
		app := &Application{
			ObjectMeta: metav1.ObjectMeta{Name: appName, Namespace: namespace},
			Spec: ApplicationSpec{
				Repositories: []ApplicationSpecRepository{
					newAppSpecRepository(ghRepoList.Items[0].Name, MissingBranchStrategyUseDefaultBranch),
					newAppSpecRepository(ghRepoList.Items[1].Name, MissingBranchStrategyUseDefaultBranch),
					newAppSpecRepository(ghRepoList.Items[2].Name, MissingBranchStrategyIgnore),
				},
			},
		}
		k = fake.NewClientBuilder().
			WithScheme(scheme).
			WithObjects(app).
			WithLists(ghRepoList, ghRepoRefList, envList).
			Build()
	})
	It("should create missing environments properly", func(ctx context.Context) {
		r := &Application{}
		Expect(k.Get(ctx, client.ObjectKey{Namespace: namespace, Name: appName}, r)).To(Succeed())

		result, err := application.NewCreateMissingEnvironmentsAction().Execute(ctx, k, r)
		Expect(err).ToNot(HaveOccurred())
		Expect(result).To(BeNil())

		envList := &EnvironmentList{}
		Expect(k.List(ctx, envList)).To(Succeed())
		slices.SortFunc(envList.Items, func(i, j Environment) int { return strings.Compare(i.Spec.PreferredBranch, j.Spec.PreferredBranch) })
		Expect(envList.Items).To(HaveLen(2))

		Expect(metav1.GetControllerOf(&envList.Items[0])).To(Equal(metav1.NewControllerRef(r, ApplicationGVK)))
		Expect(envList.Items[0].Namespace).To(Equal(r.Namespace))
		Expect(envList.Items[0].Spec.PreferredBranch).To(Equal("b1"))

		Expect(metav1.GetControllerOf(&envList.Items[1])).To(Equal(metav1.NewControllerRef(r, ApplicationGVK)))
		Expect(envList.Items[1].Spec.PreferredBranch).To(Equal("b2"))
		Expect(envList.Items[1].Namespace).To(Equal(r.Namespace))
	})

	// TODO: test conditions are set in case of API errors
})
