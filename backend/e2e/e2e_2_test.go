package e2e_test

import (
	"context"
	apiv1 "github.com/arikkfir/devbot/backend/api/v1"
	. "github.com/arikkfir/devbot/backend/e2e"
	"github.com/arikkfir/devbot/backend/internal/util/k8s"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"time"
)

var _ = Describe("Application Deployment", func() {

	var gh *GitHub
	var commonRepo, serverRepo *GitHubRepositoryInfo
	BeforeEach(func(ctx context.Context) {
		gh = NewGitHub(ctx)
		commonRepo = gh.CreateRepository(ctx, "app1/common")
		serverRepo = gh.CreateRepository(ctx, "app1/server")
	})

	var k *Kubernetes
	var ns *Namespace
	var commonRepoObjName, serverRepoObjName string
	BeforeEach(func(ctx context.Context) {
		k = NewKubernetes(ctx)
		ns = k.CreateNamespace(ctx)
		ghAuthSecretName, ghAuthSecretKeyName := ns.CreateGitHubAuthSecret(ctx, gh.Token)
		auth := apiv1.GitHubRepositoryAuth{
			PersonalAccessToken: &apiv1.GitHubRepositoryAuthPersonalAccessToken{
				Secret: apiv1.SecretReferenceWithOptionalNamespace{
					Name:      ghAuthSecretName,
					Namespace: ns.Name,
				},
				Key: ghAuthSecretKeyName,
			},
		}
		ns.CreateGitHubRepository(ctx, &commonRepoObjName, apiv1.GitHubRepositorySpec{
			Owner: commonRepo.Owner,
			Name:  commonRepo.Name,
			Auth:  auth,
		})
		ns.CreateGitHubRepository(ctx, &serverRepoObjName, apiv1.GitHubRepositorySpec{
			Owner: serverRepo.Owner,
			Name:  serverRepo.Name,
			Auth:  auth,
		})
	})

	When("application is created", func() {

		var appObjName string
		BeforeEach(func(ctx context.Context) {
			ns.CreateApplication(ctx, &appObjName, apiv1.ApplicationSpec{
				Repositories: []apiv1.ApplicationSpecRepository{
					{
						RepositoryReferenceWithOptionalNamespace: apiv1.RepositoryReferenceWithOptionalNamespace{
							APIVersion: apiv1.GitHubRepositoryGVK.GroupVersion().String(),
							Kind:       apiv1.GitHubRepositoryGVK.Kind,
							Name:       commonRepoObjName,
							Namespace:  ns.Name,
						},
						MissingBranchStrategy: apiv1.MissingBranchStrategyUseDefaultBranch,
					},
					{
						RepositoryReferenceWithOptionalNamespace: apiv1.RepositoryReferenceWithOptionalNamespace{
							APIVersion: apiv1.GitHubRepositoryGVK.GroupVersion().String(),
							Kind:       apiv1.GitHubRepositoryGVK.Kind,
							Name:       serverRepoObjName,
							Namespace:  ns.Name,
						},
						MissingBranchStrategy: apiv1.MissingBranchStrategyUseDefaultBranch,
					},
				},
			})
		})

		It("should create the 'main' environment", func(ctx context.Context) {
			Eventually(func(o Gomega) {
				a := &apiv1.Application{}
				o.Expect(k.Client.Get(ctx, client.ObjectKey{Namespace: ns.Name, Name: appObjName}, a)).Error().NotTo(HaveOccurred())

				envsList := &apiv1.EnvironmentList{}
				o.Expect(k.Client.List(ctx, envsList, client.InNamespace(a.Namespace), k8s.OwnedBy(k.Client.Scheme(), a))).Error().NotTo(HaveOccurred())
				o.Expect(envsList.Items).To(HaveLen(1))
				o.Expect(envsList.Items[0].Spec.PreferredBranch).To(Equal("main"))
				o.Expect(envsList.Items[0].Status.GetFailedToInitializeCondition()).To(BeNil())
				o.Expect(envsList.Items[0].Status.GetFinalizingCondition()).To(BeNil())
				o.Expect(envsList.Items[0].Status.GetInvalidCondition()).To(BeNil())
				o.Expect(envsList.Items[0].Status.GetStaleCondition()).To(BeNil())
			}).Within(2 * time.Minute).WithPolling(5 * time.Second).Should(Succeed())

			time.Sleep(5 * time.Minute)
		})
	})
})
