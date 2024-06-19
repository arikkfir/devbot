package e2e_test

import (
	"context"
	"fmt"
	"slices"

	"github.com/google/go-github/v56/github"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	apiv1 "github.com/arikkfir/devbot/api/v1"
	"github.com/arikkfir/devbot/e2e/util"
)

// TODO: add test case for a repository with the "IgnoreStrategy" of missing branches

var _ = Describe("Application Deployment", func() {

	var nsName string
	BeforeEach(func(ctx context.Context) { nsName = util.CreateK8sNamespace(ctx, c) })
	JustAfterEach(func(ctx SpecContext) { util.PrintK8sDebugInfo(ctx, c, rc, nsName) })

	var tokenSecretName, tokenSecretKey, token string
	var secretRef apiv1.SecretReferenceWithOptionalNamespace
	BeforeEach(func(ctx context.Context) {
		token = util.GetGitHubToken()
		tokenSecretName, tokenSecretKey = util.CreateK8sSecretWithGitHubAuthToken(ctx, c, nsName, token)
		util.GrantK8sAccessToSecret(ctx, c, nsName, tokenSecretName)
		secretRef = apiv1.SecretReferenceWithOptionalNamespace{Name: tokenSecretName, Namespace: nsName}
	})

	var ghCommonRepo, ghServerRepo, ghPortalRepo *github.Repository
	BeforeEach(func(ctx context.Context) {
		ghCommonRepo = util.CreateGitHubRepository(ctx, gh, repositoriesFS, "repositories/common")
		ghServerRepo = util.CreateGitHubRepository(ctx, gh, repositoriesFS, "repositories/server")
		ghPortalRepo = util.CreateGitHubRepository(ctx, gh, repositoriesFS, "repositories/portal")
	})

	var kCommonRepoName, kServerRepoName, kPortalRepoName string
	BeforeEach(func(ctx context.Context) {
		pat := apiv1.GitHubRepositoryPersonalAccessToken{Secret: secretRef, Key: tokenSecretKey}
		createGitHubSpec := func(r *github.Repository) *apiv1.GitHubRepositorySpec {
			return &apiv1.GitHubRepositorySpec{Owner: r.Owner.GetLogin(), Name: r.GetName(), PersonalAccessToken: pat}
		}
		createRepoSpec := func(r *github.Repository) apiv1.RepositorySpec {
			return apiv1.RepositorySpec{GitHub: createGitHubSpec(r), RefreshInterval: "5s"}
		}
		kCommonRepoName = util.CreateK8sRepository(ctx, c, nsName, createRepoSpec(ghCommonRepo))
		kServerRepoName = util.CreateK8sRepository(ctx, c, nsName, createRepoSpec(ghServerRepo))
		kPortalRepoName = util.CreateK8sRepository(ctx, c, nsName, createRepoSpec(ghPortalRepo))
	})

	var devopsServiceAccountName, appName string
	BeforeEach(func(ctx context.Context) {
		devopsServiceAccountName = util.CreateK8sGitOpsServiceAccount(ctx, c, nsName)
		appName = util.CreateK8sApplication(ctx, c, nsName, apiv1.ApplicationSpec{
			Repositories: []apiv1.ApplicationSpecRepository{
				{Namespace: nsName, Name: kCommonRepoName, MissingBranchStrategy: apiv1.UseDefaultBranchStrategy},
				{Namespace: nsName, Name: kServerRepoName, MissingBranchStrategy: apiv1.UseDefaultBranchStrategy},
				{Namespace: nsName, Name: kPortalRepoName, MissingBranchStrategy: apiv1.UseDefaultBranchStrategy},
			},
			ServiceAccountName: devopsServiceAccountName,
		})
	})

	type repoInfo struct {
		g          *github.Repository
		k          *apiv1.Repository
		branchSHAs map[string]string
	}

	It("should reconcile and deploy the application", func(ctx context.Context) {
		verifyApp := func(g Gomega, allowedGitHubRepos ...*github.Repository) {
			GinkgoHelper()
			app := &apiv1.Application{}
			g.Expect(c.Get(ctx, client.ObjectKey{Namespace: nsName, Name: appName}, app)).To(Succeed())
			g.Expect(app.Status.Conditions).To(BeEmpty())

			var detectedBranchesInRepos []string
			reposByOwnerAndName := make(map[string]repoInfo)
			reposByKRepoNames := make(map[string]repoInfo)
			for _, rr := range app.Spec.Repositories {
				r := &apiv1.Repository{}
				g.Expect(c.Get(ctx, client.ObjectKey{Namespace: rr.Namespace, Name: rr.Name}, r)).To(Succeed())
				g.Expect(r.Status.DefaultBranch).To(Equal("main"))
				g.Expect(r.Status.Conditions).To(BeEmpty())
				var ghr *github.Repository
				for _, ghRepo := range []*github.Repository{ghCommonRepo, ghServerRepo, ghPortalRepo} {
					if r.Spec.GitHub.Owner == ghRepo.Owner.GetLogin() && r.Spec.GitHub.Name == ghRepo.GetName() {
						ghr = ghRepo
						break
					}
				}
				g.Expect(ghr).ToNot(BeNil())
				info := repoInfo{g: ghr, k: r, branchSHAs: util.GetGitHubRepositoryBranchNamesAndSHA(ctx, gh, ghr)}
				reposByOwnerAndName[ghr.Owner.GetLogin()+"/"+ghr.GetName()] = info
				reposByKRepoNames[rr.GetObjectKey(app.Namespace).String()] = info
				for branch := range info.branchSHAs {
					if !slices.Contains(detectedBranchesInRepos, branch) {
						detectedBranchesInRepos = append(detectedBranchesInRepos, branch)
					}
				}
			}
			g.Expect(reposByOwnerAndName).To(HaveLen(len(app.Spec.Repositories)))

			envList := &apiv1.EnvironmentList{}
			g.Expect(c.List(ctx, envList, client.InNamespace(nsName))).To(Succeed())
			g.Expect(envList.Items).To(HaveLen(len(detectedBranchesInRepos)))

			deploymentsList := &apiv1.DeploymentList{}
			g.Expect(c.List(ctx, deploymentsList, client.InNamespace(nsName))).To(Succeed())
			g.Expect(deploymentsList.Items).To(HaveLen(len(envList.Items) * len(app.Spec.Repositories)))

			for _, env := range envList.Items {
				g.Expect(env.Spec.PreferredBranch).To(BeElementOf(detectedBranchesInRepos))
				detectedBranchesInRepos = slices.DeleteFunc(detectedBranchesInRepos, func(envName string) bool { return env.Spec.PreferredBranch == envName })
				g.Expect(env.Status.Conditions).To(BeEmpty())

				for _, rr := range app.Spec.Repositories {
					depIndex := slices.IndexFunc(deploymentsList.Items, func(d apiv1.Deployment) bool {
						return d.Spec.Repository.GetObjectKey() == rr.GetObjectKey(app.Namespace) && metav1.IsControlledBy(&d, &env)
					})
					g.Expect(depIndex).To(BeNumerically(">=", 0))
					d := deploymentsList.Items[depIndex].DeepCopy()
					deploymentsList.Items = slices.Delete(deploymentsList.Items, depIndex, depIndex+1)
					g.Expect(d.Status.Conditions).To(BeEmpty())

					info := reposByKRepoNames[rr.GetObjectKey(app.Namespace).String()]
					var expectedDeploymentBranch string
					if _, branchExistsForRepo := info.branchSHAs[env.Spec.PreferredBranch]; branchExistsForRepo {
						expectedDeploymentBranch = env.Spec.PreferredBranch
					} else {
						expectedDeploymentBranch = info.k.Status.DefaultBranch
					}
					g.Expect(d.Status.Branch).To(Equal(expectedDeploymentBranch))
					g.Expect(d.Status.LastAttemptedRevision).To(Equal(info.branchSHAs[expectedDeploymentBranch]))
					g.Expect(d.Status.LastAppliedRevision).To(Equal(info.branchSHAs[expectedDeploymentBranch]))
				}

				cm := &corev1.ConfigMap{}
				g.Expect(c.Get(ctx, client.ObjectKey{Namespace: nsName, Name: fmt.Sprintf("%s-configuration", env.Spec.PreferredBranch)}, cm)).To(Succeed())
				g.Expect(cm.Data["env"]).To(Equal(env.Spec.PreferredBranch))

				for _, name := range []string{"server", "portal"} {
					sa := &corev1.ServiceAccount{}
					g.Expect(c.Get(ctx, client.ObjectKey{Namespace: nsName, Name: fmt.Sprintf("%s-%s", env.Spec.PreferredBranch, name)}, sa)).To(Succeed())

					svc := &corev1.Service{}
					g.Expect(c.Get(ctx, client.ObjectKey{Namespace: nsName, Name: fmt.Sprintf("%s-%s", env.Spec.PreferredBranch, name)}, svc)).To(Succeed())

					d := &appsv1.Deployment{}
					g.Expect(c.Get(ctx, client.ObjectKey{Namespace: nsName, Name: fmt.Sprintf("%s-%s", env.Spec.PreferredBranch, name)}, d)).To(Succeed())
					g.Expect(d.Status.AvailableReplicas).To(BeNumerically("==", 1))
					g.Expect(d.Status.UpdatedReplicas).To(BeNumerically("==", 1))
					g.Expect(d.Status.Replicas).To(BeNumerically("==", 1))
					g.Expect(d.Status.ReadyReplicas).To(BeNumerically("==", 1))
				}
			}
		}

		// Ensure application, environments & deployments are reconciled to the desired state:
		// - Repositories, Applications, Environments, and Deployments are reconciled and ready
		// - Resources were successfully created & reconciled according to the desired state
		Eventually(func(g Gomega) { verifyApp(g, ghCommonRepo, ghServerRepo, ghPortalRepo) }, "3m", "5s").Should(Succeed())

		// Get current SHA and verify that's indeed what's in the repository status
		oldServerMainSHA := util.GetGitHubRepositoryBranchSHA(ctx, gh, ghServerRepo, "main")
		r := &apiv1.Repository{}
		Expect(c.Get(ctx, client.ObjectKey{Namespace: nsName, Name: kServerRepoName}, r)).To(Succeed())
		Expect(r.Status.Revisions["main"]).To(Equal(oldServerMainSHA))

		// Create a new commit in the server repository and verify that indeed the repository status was updated
		util.CreateFileInGitHubRepositoryBranch(ctx, gh, ghServerRepo, "main")
		Eventually(func(g Gomega) {
			r := &apiv1.Repository{}
			g.Expect(c.Get(ctx, client.ObjectKey{Namespace: nsName, Name: kServerRepoName}, r)).To(Succeed())
			g.Expect(r.Status.Revisions["main"]).ToNot(Equal(oldServerMainSHA))
		}, "3m", "5s").Should(Succeed())

		// Verify that all objects have been reconciled according to the new commit
		Eventually(func(g Gomega) { verifyApp(g, ghCommonRepo, ghServerRepo, ghPortalRepo) }, "3m", "5s").Should(Succeed())

		// Create more environments by creating new branches in the repositories
		util.CreateGitHubRepositoryBranch(ctx, gh, ghCommonRepo, "feature1")
		util.CreateFileInGitHubRepositoryBranch(ctx, gh, ghCommonRepo, "feature1")
		util.CreateGitHubRepositoryBranch(ctx, gh, ghServerRepo, "feature2")
		util.CreateFileInGitHubRepositoryBranch(ctx, gh, ghServerRepo, "feature2")
		util.CreateGitHubRepositoryBranch(ctx, gh, ghPortalRepo, "feature2")
		util.CreateFileInGitHubRepositoryBranch(ctx, gh, ghPortalRepo, "feature2")
		util.CreateGitHubRepositoryBranch(ctx, gh, ghPortalRepo, "feature3")
		util.CreateFileInGitHubRepositoryBranch(ctx, gh, ghPortalRepo, "feature3")
		Eventually(func(g Gomega) { verifyApp(g, ghCommonRepo, ghServerRepo, ghPortalRepo) }, "3m", "5s").Should(Succeed())
	})
})
