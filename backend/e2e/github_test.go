package e2e_test

import (
	"context"
	apiv1 "github.com/arikkfir/devbot/backend/api/v1"
	"github.com/arikkfir/devbot/backend/e2e"
	"github.com/arikkfir/devbot/backend/internal/util/k8s"
	"github.com/arikkfir/devbot/backend/internal/util/strings"
	"github.com/google/go-github/v56/github"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"os"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"slices"
	strings2 "strings"
	"time"
)

var _ = Describe("GitHub", func() {

	When("github repository object is created", func() {
		var gh *github.Client
		BeforeEach(func(ctx context.Context) { e2e.CreateGitHubClient(ctx, &gh) })

		var ghOwner, ghRepoName string
		BeforeEach(func(ctx context.Context) { e2e.CreateGitHubRepository(ctx, gh, &ghOwner, &ghRepoName) })

		var mainSHA string
		BeforeEach(func(ctx context.Context) {
			e2e.GetGitHubBranchCommitSHA(ctx, gh, ghOwner, ghRepoName, "main", &mainSHA)
		})

		var k client.Client
		BeforeEach(func(ctx context.Context) { e2e.CreateKubernetesClient(scheme, &k) })

		var ghAuthSecretNamespace, ghAuthSecretName, ghAuthSecretKeyName string
		BeforeEach(func(ctx context.Context) {
			ghAuthSecretKeyName = "token"
			gitHubAuthToken := os.Getenv("GITHUB_TOKEN")
			Expect(gitHubAuthToken).ToNot(BeEmpty())

			ghAuthSecretNamespace, ghAuthSecretName = "default", strings.RandomHash(7)
			secret := &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{Namespace: ghAuthSecretNamespace, Name: ghAuthSecretName},
				Data:       map[string][]byte{ghAuthSecretKeyName: []byte(gitHubAuthToken)},
			}
			Expect(k.Create(ctx, secret)).To(Succeed())
			DeferCleanup(func() { Expect(k.Delete(context.WithoutCancel(ctx), secret)).To(Succeed()) })

			clusterRole := &rbacv1.ClusterRole{
				ObjectMeta: metav1.ObjectMeta{Name: "e2e-github-secret-reader"},
				Rules: []rbacv1.PolicyRule{
					{
						APIGroups:     []string{corev1.GroupName},
						Resources:     []string{"secrets"},
						Verbs:         []string{"get", "list", "watch"},
						ResourceNames: []string{ghAuthSecretName},
					},
				},
			}
			Expect(k.Create(ctx, clusterRole)).To(Succeed())
			DeferCleanup(func() { Expect(k.Delete(context.WithoutCancel(ctx), clusterRole)).To(Succeed()) })

			saKind := reflect.TypeOf(corev1.ServiceAccount{}).Name()
			clusterRoleKind := reflect.TypeOf(*clusterRole).Name()
			roleBinding := &rbacv1.RoleBinding{
				ObjectMeta: metav1.ObjectMeta{Namespace: secret.Namespace, Name: "e2e-github-secret-reader"},
				RoleRef:    rbacv1.RoleRef{APIGroup: rbacv1.SchemeGroupVersion.Group, Kind: clusterRoleKind, Name: clusterRole.Name},
				Subjects: []rbacv1.Subject{
					{Kind: saKind, Name: "devbot-github-repository-controller", Namespace: "devbot"},
					{Kind: saKind, Name: "devbot-github-ref-controller", Namespace: "devbot"},
				},
			}
			Expect(k.Create(ctx, roleBinding)).To(Succeed())
			DeferCleanup(func() { Expect(k.Delete(context.WithoutCancel(ctx), roleBinding)).To(Succeed()) })
		})

		var namespace, repoObjName string
		BeforeEach(func(ctx context.Context) {
			namespace, repoObjName = "default", strings.RandomHash(7)
			r := &apiv1.GitHubRepository{
				ObjectMeta: metav1.ObjectMeta{Namespace: namespace, Name: repoObjName},
				Spec: apiv1.GitHubRepositorySpec{
					Owner: ghOwner,
					Name:  ghRepoName,
					Auth: apiv1.GitHubRepositoryAuth{
						PersonalAccessToken: &apiv1.GitHubRepositoryAuthPersonalAccessToken{
							Secret: apiv1.SecretReferenceWithOptionalNamespace{
								Name:      ghAuthSecretName,
								Namespace: ghAuthSecretNamespace,
							},
							Key: ghAuthSecretKeyName,
						},
					},
					RefreshInterval: "60s",
				},
			}
			Expect(k.Create(ctx, r)).To(Succeed())
			DeferCleanup(func() { Expect(k.Delete(context.WithoutCancel(ctx), r)).To(Succeed()) })
		})

		It("should sync github repository object and default branch", func(ctx context.Context) {
			Eventually(func(o Gomega) {
				r := &apiv1.GitHubRepository{}
				o.Expect(k.Get(ctx, client.ObjectKey{Namespace: namespace, Name: repoObjName}, r)).To(Succeed())
				o.Expect(r.Status.GetFailedToInitializeCondition()).To(BeNil())
				o.Expect(r.Status.GetFinalizingCondition()).To(BeNil())
				o.Expect(r.Status.GetInvalidCondition()).To(BeNil())
				o.Expect(r.Status.GetStaleCondition()).To(BeNil())
				o.Expect(r.Status.GetUnauthenticatedCondition()).To(BeNil())

				refs := &apiv1.GitHubRepositoryRefList{}
				o.Expect(k.List(ctx, refs, client.InNamespace(namespace), k8s.OwnedBy(scheme, r))).To(Succeed())
				o.Expect(refs.Items).To(HaveLen(1))
				o.Expect(refs.Items[0].Spec.Ref).To(Equal("main"))
				o.Expect(refs.Items[0].Status.RepositoryOwner).To(Equal(r.Spec.Owner))
				o.Expect(refs.Items[0].Status.RepositoryName).To(Equal(r.Spec.Name))
				o.Expect(refs.Items[0].Status.CommitSHA).To(Equal(mainSHA))
				o.Expect(refs.Items[0].Status.GetFailedToInitializeCondition()).To(BeNil())
				o.Expect(refs.Items[0].Status.GetFinalizingCondition()).To(BeNil())
				o.Expect(refs.Items[0].Status.GetInvalidCondition()).To(BeNil())
				o.Expect(refs.Items[0].Status.GetStaleCondition()).To(BeNil())
				o.Expect(refs.Items[0].Status.GetUnauthenticatedCondition()).To(BeNil())
			}).Within(2 * time.Minute).WithPolling(5 * time.Second).Should(Succeed())
		})

		When("a new branch is created", func() {

			var newBranchSHA, newBranchName string
			BeforeEach(func(ctx context.Context) {
				newBranchName = "z" + strings.RandomHash(7) // adding "z" prefix to make sure new branch is last on sort
				e2e.CreateGitHubBranch(ctx, gh, ghOwner, ghRepoName, newBranchName)
				e2e.GetGitHubBranchCommitSHA(ctx, gh, ghOwner, ghRepoName, newBranchName, &newBranchSHA)
			})

			It("should create github repository ref object", func(ctx context.Context) {
				Eventually(func(o Gomega) {
					r := &apiv1.GitHubRepository{}
					o.Expect(k.Get(ctx, client.ObjectKey{Namespace: namespace, Name: repoObjName}, r)).To(Succeed())

					refs := &apiv1.GitHubRepositoryRefList{}
					o.Expect(k.List(ctx, refs, client.InNamespace(r.Namespace), k8s.OwnedBy(scheme, r))).To(Succeed())
					o.Expect(refs.Items).To(HaveLen(2))
					slices.SortFunc(refs.Items, func(i, j apiv1.GitHubRepositoryRef) int { return strings2.Compare(i.Spec.Ref, j.Spec.Ref) })
					o.Expect(refs.Items[0].Spec.Ref).To(Equal("main"))
					o.Expect(refs.Items[0].Status.RepositoryOwner).To(Equal(r.Spec.Owner))
					o.Expect(refs.Items[0].Status.RepositoryName).To(Equal(r.Spec.Name))
					o.Expect(refs.Items[0].Status.CommitSHA).To(Equal(mainSHA))
					o.Expect(refs.Items[0].Status.GetFailedToInitializeCondition()).To(BeNil())
					o.Expect(refs.Items[0].Status.GetFinalizingCondition()).To(BeNil())
					o.Expect(refs.Items[0].Status.GetInvalidCondition()).To(BeNil())
					o.Expect(refs.Items[0].Status.GetStaleCondition()).To(BeNil())
					o.Expect(refs.Items[0].Status.GetUnauthenticatedCondition()).To(BeNil())
					o.Expect(refs.Items[1].Spec.Ref).To(Equal(newBranchName))
					o.Expect(refs.Items[1].Status.RepositoryOwner).To(Equal(r.Spec.Owner))
					o.Expect(refs.Items[1].Status.RepositoryName).To(Equal(r.Spec.Name))
					o.Expect(refs.Items[1].Status.CommitSHA).To(Equal(newBranchSHA))
					o.Expect(refs.Items[1].Status.GetFailedToInitializeCondition()).To(BeNil())
					o.Expect(refs.Items[1].Status.GetFinalizingCondition()).To(BeNil())
					o.Expect(refs.Items[1].Status.GetInvalidCondition()).To(BeNil())
					o.Expect(refs.Items[1].Status.GetStaleCondition()).To(BeNil())
					o.Expect(refs.Items[1].Status.GetUnauthenticatedCondition()).To(BeNil())
				}).Within(2 * time.Minute).WithPolling(5 * time.Second).Should(Succeed())
			})

			When("the new branch is updated", func() {
				var updatedBranchSHA string
				BeforeEach(func(ctx context.Context) {
					e2e.CreateGitHubFile(ctx, gh, ghOwner, ghRepoName, newBranchName, &updatedBranchSHA)
				})

				It("should sync github repository ref object", func(ctx context.Context) {
					Eventually(func(o Gomega) {
						r := &apiv1.GitHubRepository{}
						o.Expect(k.Get(ctx, client.ObjectKey{Namespace: namespace, Name: repoObjName}, r)).To(Succeed())

						refs := &apiv1.GitHubRepositoryRefList{}
						o.Expect(k.List(ctx, refs, client.InNamespace(r.Namespace), k8s.OwnedBy(scheme, r))).To(Succeed())

						var ref *apiv1.GitHubRepositoryRef
						for _, r := range refs.Items {
							if r.Spec.Ref == newBranchName {
								ref = &r
								break
							}
						}
						o.Expect(ref).ToNot(BeNil())
						o.Expect(ref.Status.CommitSHA).To(Equal(updatedBranchSHA))
						o.Expect(ref.Status.GetFailedToInitializeCondition()).To(BeNil())
						o.Expect(ref.Status.GetFinalizingCondition()).To(BeNil())
						o.Expect(ref.Status.GetInvalidCondition()).To(BeNil())
						o.Expect(ref.Status.GetStaleCondition()).To(BeNil())
						o.Expect(ref.Status.GetUnauthenticatedCondition()).To(BeNil())
					}).Within(2 * time.Minute).WithPolling(5 * time.Second).Should(Succeed())
				})
			})

			When("branch is subsequently removed", func() {

				BeforeEach(func(ctx context.Context) {
					e2e.DeleteGitHubBranch(ctx, gh, ghOwner, ghRepoName, newBranchName)
				})

				It("should delete github repository ref object", func(ctx context.Context) {
					Eventually(func(o Gomega) {
						r := &apiv1.GitHubRepository{}
						o.Expect(k.Get(ctx, client.ObjectKey{Namespace: namespace, Name: repoObjName}, r)).To(Succeed())

						refs := &apiv1.GitHubRepositoryRefList{}
						o.Expect(k.List(ctx, refs, client.InNamespace(r.Namespace), k8s.OwnedBy(scheme, r))).To(Succeed())
						o.Expect(refs.Items).To(HaveLen(1))
						o.Expect(refs.Items[0].Spec.Ref).To(Equal("main"))
					}).Within(2 * time.Minute).WithPolling(5 * time.Second).Should(Succeed())
				})
			})
		})
	})
})
