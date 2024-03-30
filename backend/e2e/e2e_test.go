package e2e_test

import (
	"context"
	"embed"
	apiv1 "github.com/arikkfir/devbot/backend/api/v1"
	"github.com/arikkfir/devbot/backend/internal/util/k8s"
	. "github.com/arikkfir/devbot/backend/internal/util/testing"
	. "github.com/arikkfir/devbot/backend/internal/util/testing/justest"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"testing"
	"time"
)

var (
	//go:embed all:repositories/*
	repositoriesFS embed.FS
)

type (
	conditionExpectations map[string]*metav1.Condition
	deploymentExpectation struct {
		branch                string
		conditions            conditionExpectations
		lastAttemptedRevision string
		lastAppliedRevision   string
		repositoryName        string
	}
	environmentExpectation struct {
		conditions      conditionExpectations
		preferredBranch string
		deployments     []deploymentExpectation
	}
	applicationExpectation struct {
		conditions   conditionExpectations
		environments []environmentExpectation
	}
	applicationExpectations map[string]applicationExpectation
)

func createRepository(t JustT, ctx context.Context, gh *GitHub, ns *KNamespace, name string) (*GitHubRepositoryInfo, string) {
	ghRepo := gh.CreateRepository(t, ctx, repositoriesFS, "repositories/"+name)
	kRepoName := ns.CreateRepository(t, ctx, apiv1.RepositorySpec{
		GitHub: &apiv1.GitHubRepositorySpec{
			Owner:               ghRepo.Owner,
			Name:                ghRepo.Name,
			PersonalAccessToken: ns.CreateGitHubAuthSecretSpec(t, ctx, gh.Token, true),
		},
		RefreshInterval: "30s",
	})
	return ghRepo, kRepoName
}

func TestSingleRepoApplicationDeployment(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	k := NewKubernetes(t)
	ns := k.CreateNamespace(t, ctx)
	gh := NewGitHub(t, ctx)

	ghRepo, kRepoName := createRepository(t, ctx, gh, ns, "common")

	// TODO: add "prod" repository that uses "IgnoreStrategy" for missing branches - so it is only applied on "main"
	// TODO: add expiry feature
	// TODO: slug branch/repo names to k8s names for better observability
	// TODO: stale should only be false/unknown if lastApplied<>lastAttempted

	appName := ns.CreateApplication(t, ctx, apiv1.ApplicationSpec{
		Repositories: []apiv1.ApplicationSpecRepository{
			{Name: kRepoName, Namespace: ns.Name, MissingBranchStrategy: apiv1.UseDefaultBranchStrategy},
		},
		ServiceAccountName: "devbot-gitops",
	})

	// Validate initial deployment
	For(t).Expect(func(t JustT) {
		validateApplicationExpectations(t, ctx, k, ns, applicationExpectations{
			appName: {
				conditions: map[string]*metav1.Condition{
					apiv1.FailedToInitialize: nil,
					apiv1.Finalizing:         nil,
					apiv1.Stale:              nil,
				},
				environments: []environmentExpectation{
					{
						preferredBranch: "main",
						conditions: map[string]*metav1.Condition{
							apiv1.FailedToInitialize: nil,
							apiv1.Finalizing:         nil,
							apiv1.Stale:              nil,
						},
						deployments: []deploymentExpectation{
							{
								repositoryName:        kRepoName,
								branch:                "main",
								lastAppliedRevision:   ghRepo.GetBranchSHA(t, ctx, "main"),
								lastAttemptedRevision: ghRepo.GetBranchSHA(t, ctx, "main"),
								conditions: map[string]*metav1.Condition{
									apiv1.FailedToInitialize: nil,
									apiv1.Finalizing:         nil,
									apiv1.Stale:              nil,
								},
							},
						},
					},
				},
			},
		})
	}).Will(Eventually(Succeed()).Within(5 * time.Minute).ProbingEvery(100 * time.Millisecond))

	// Now create a new branch, expecting a new environment & deployment will occur
	_ = ghRepo.CreateBranch(t, ctx, "feature1")
	ghRepo.CreateFile(t, ctx, "feature1")

	For(t).Expect(func(t JustT) {
		validateApplicationExpectations(t, ctx, k, ns, applicationExpectations{
			appName: {
				conditions: map[string]*metav1.Condition{
					apiv1.FailedToInitialize: nil,
					apiv1.Finalizing:         nil,
					apiv1.Stale:              nil,
				},
				environments: []environmentExpectation{
					{
						preferredBranch: "main",
						conditions: map[string]*metav1.Condition{
							apiv1.FailedToInitialize: nil,
							apiv1.Finalizing:         nil,
							apiv1.Stale:              nil,
						},
						deployments: []deploymentExpectation{
							{
								repositoryName:        kRepoName,
								branch:                "main",
								lastAppliedRevision:   ghRepo.GetBranchSHA(t, ctx, "main"),
								lastAttemptedRevision: ghRepo.GetBranchSHA(t, ctx, "main"),
								conditions: map[string]*metav1.Condition{
									apiv1.FailedToInitialize: nil,
									apiv1.Finalizing:         nil,
									apiv1.Stale:              nil,
								},
							},
						},
					},
					{
						preferredBranch: "feature1",
						conditions: map[string]*metav1.Condition{
							apiv1.FailedToInitialize: nil,
							apiv1.Finalizing:         nil,
							apiv1.Stale:              nil,
						},
						deployments: []deploymentExpectation{
							{
								repositoryName:        kRepoName,
								branch:                "feature1",
								lastAppliedRevision:   ghRepo.GetBranchSHA(t, ctx, "feature1"),
								lastAttemptedRevision: ghRepo.GetBranchSHA(t, ctx, "feature1"),
								conditions: map[string]*metav1.Condition{
									apiv1.FailedToInitialize: nil,
									apiv1.Finalizing:         nil,
									apiv1.Stale:              nil,
								},
							},
						},
					},
				},
			},
		})
	}).Will(Eventually(Succeed()).Within(5 * time.Minute).ProbingEvery(100 * time.Millisecond))
}

func TestMultiRepoApplicationDeployment(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	k := NewKubernetes(t)
	ns := k.CreateNamespace(t, ctx)
	gh := NewGitHub(t, ctx)

	ghCommonRepo, kCommonRepoName := createRepository(t, ctx, gh, ns, "common")
	ghPortalRepo, kPortalRepoName := createRepository(t, ctx, gh, ns, "portal")
	ghServerRepo, kServerRepoName := createRepository(t, ctx, gh, ns, "server")

	// TODO: add "prod" repository that uses "IgnoreStrategy" for missing branches - so it is only applied on "main"
	// TODO: add expiry feature

	appName := ns.CreateApplication(t, ctx, apiv1.ApplicationSpec{
		Repositories: []apiv1.ApplicationSpecRepository{
			{Name: kCommonRepoName, Namespace: ns.Name, MissingBranchStrategy: apiv1.UseDefaultBranchStrategy},
			{Name: kPortalRepoName, Namespace: ns.Name, MissingBranchStrategy: apiv1.UseDefaultBranchStrategy},
			{Name: kServerRepoName, Namespace: ns.Name, MissingBranchStrategy: apiv1.UseDefaultBranchStrategy},
		},
		ServiceAccountName: "devbot-gitops",
	})

	// Validate initial deployment
	For(t).Expect(func(t JustT) {
		validateApplicationExpectations(t, ctx, k, ns, applicationExpectations{
			appName: {
				conditions: map[string]*metav1.Condition{
					apiv1.FailedToInitialize: nil,
					apiv1.Finalizing:         nil,
					apiv1.Stale:              nil,
				},
				environments: []environmentExpectation{
					{
						preferredBranch: "main",
						conditions: map[string]*metav1.Condition{
							apiv1.FailedToInitialize: nil,
							apiv1.Finalizing:         nil,
							apiv1.Stale:              nil,
						},
						deployments: []deploymentExpectation{
							{
								repositoryName:        kCommonRepoName,
								branch:                "main",
								lastAppliedRevision:   ghCommonRepo.GetBranchSHA(t, ctx, "main"),
								lastAttemptedRevision: ghCommonRepo.GetBranchSHA(t, ctx, "main"),
								conditions: map[string]*metav1.Condition{
									apiv1.FailedToInitialize: nil,
									apiv1.Finalizing:         nil,
									apiv1.Stale:              nil,
								},
							},
							{
								repositoryName:        kPortalRepoName,
								branch:                "main",
								lastAppliedRevision:   ghPortalRepo.GetBranchSHA(t, ctx, "main"),
								lastAttemptedRevision: ghPortalRepo.GetBranchSHA(t, ctx, "main"),
								conditions: map[string]*metav1.Condition{
									apiv1.FailedToInitialize: nil,
									apiv1.Finalizing:         nil,
									apiv1.Stale:              nil,
								},
							},
							{
								repositoryName:        kServerRepoName,
								branch:                "main",
								lastAppliedRevision:   ghServerRepo.GetBranchSHA(t, ctx, "main"),
								lastAttemptedRevision: ghServerRepo.GetBranchSHA(t, ctx, "main"),
								conditions: map[string]*metav1.Condition{
									apiv1.FailedToInitialize: nil,
									apiv1.Finalizing:         nil,
									apiv1.Stale:              nil,
								},
							},
						},
					},
				},
			},
		})
	}).Will(Eventually(Succeed()).Within(35 * time.Minute).ProbingEvery(100 * time.Millisecond))

	// Create new branches
	_ = ghCommonRepo.CreateBranch(t, ctx, "feature1")
	ghCommonRepo.CreateFile(t, ctx, "feature1")
	_ = ghPortalRepo.CreateBranch(t, ctx, "feature2")
	ghPortalRepo.CreateFile(t, ctx, "feature2")
	_ = ghServerRepo.CreateBranch(t, ctx, "feature3")
	ghServerRepo.CreateFile(t, ctx, "feature3")

	// Validate new environments & deployments created & deployed
	For(t).Expect(func(t JustT) {
		validateApplicationExpectations(t, ctx, k, ns, applicationExpectations{
			appName: {
				conditions: map[string]*metav1.Condition{
					apiv1.FailedToInitialize: nil,
					apiv1.Finalizing:         nil,
					apiv1.Stale:              nil,
				},
				environments: []environmentExpectation{
					{
						preferredBranch: "main",
						conditions: map[string]*metav1.Condition{
							apiv1.FailedToInitialize: nil,
							apiv1.Finalizing:         nil,
							apiv1.Stale:              nil,
						},
						deployments: []deploymentExpectation{
							{
								repositoryName:        kCommonRepoName,
								branch:                "main",
								lastAppliedRevision:   ghCommonRepo.GetBranchSHA(t, ctx, "main"),
								lastAttemptedRevision: ghCommonRepo.GetBranchSHA(t, ctx, "main"),
								conditions: map[string]*metav1.Condition{
									apiv1.FailedToInitialize: nil,
									apiv1.Finalizing:         nil,
									apiv1.Stale:              nil,
								},
							},
							{
								repositoryName:        kPortalRepoName,
								branch:                "main",
								lastAppliedRevision:   ghPortalRepo.GetBranchSHA(t, ctx, "main"),
								lastAttemptedRevision: ghPortalRepo.GetBranchSHA(t, ctx, "main"),
								conditions: map[string]*metav1.Condition{
									apiv1.FailedToInitialize: nil,
									apiv1.Finalizing:         nil,
									apiv1.Stale:              nil,
								},
							},
							{
								repositoryName:        kServerRepoName,
								branch:                "main",
								lastAppliedRevision:   ghServerRepo.GetBranchSHA(t, ctx, "main"),
								lastAttemptedRevision: ghServerRepo.GetBranchSHA(t, ctx, "main"),
								conditions: map[string]*metav1.Condition{
									apiv1.FailedToInitialize: nil,
									apiv1.Finalizing:         nil,
									apiv1.Stale:              nil,
								},
							},
						},
					},
					{
						preferredBranch: "feature1",
						conditions: map[string]*metav1.Condition{
							apiv1.FailedToInitialize: nil,
							apiv1.Finalizing:         nil,
							apiv1.Stale:              nil,
						},
						deployments: []deploymentExpectation{
							{
								repositoryName:        kCommonRepoName,
								branch:                "feature1",
								lastAppliedRevision:   ghCommonRepo.GetBranchSHA(t, ctx, "feature1"),
								lastAttemptedRevision: ghCommonRepo.GetBranchSHA(t, ctx, "feature1"),
								conditions: map[string]*metav1.Condition{
									apiv1.FailedToInitialize: nil,
									apiv1.Finalizing:         nil,
									apiv1.Stale:              nil,
								},
							},
							{
								repositoryName:        kPortalRepoName,
								branch:                "main",
								lastAppliedRevision:   ghPortalRepo.GetBranchSHA(t, ctx, "main"),
								lastAttemptedRevision: ghPortalRepo.GetBranchSHA(t, ctx, "main"),
								conditions: map[string]*metav1.Condition{
									apiv1.FailedToInitialize: nil,
									apiv1.Finalizing:         nil,
									apiv1.Stale:              nil,
								},
							},
							{
								repositoryName:        kServerRepoName,
								branch:                "main",
								lastAppliedRevision:   ghServerRepo.GetBranchSHA(t, ctx, "main"),
								lastAttemptedRevision: ghServerRepo.GetBranchSHA(t, ctx, "main"),
								conditions: map[string]*metav1.Condition{
									apiv1.FailedToInitialize: nil,
									apiv1.Finalizing:         nil,
									apiv1.Stale:              nil,
								},
							},
						},
					},
					{
						preferredBranch: "feature2",
						conditions: map[string]*metav1.Condition{
							apiv1.FailedToInitialize: nil,
							apiv1.Finalizing:         nil,
							apiv1.Stale:              nil,
						},
						deployments: []deploymentExpectation{
							{
								repositoryName:        kCommonRepoName,
								branch:                "main",
								lastAppliedRevision:   ghCommonRepo.GetBranchSHA(t, ctx, "main"),
								lastAttemptedRevision: ghCommonRepo.GetBranchSHA(t, ctx, "main"),
								conditions: map[string]*metav1.Condition{
									apiv1.FailedToInitialize: nil,
									apiv1.Finalizing:         nil,
									apiv1.Stale:              nil,
								},
							},
							{
								repositoryName:        kPortalRepoName,
								branch:                "feature2",
								lastAppliedRevision:   ghPortalRepo.GetBranchSHA(t, ctx, "feature2"),
								lastAttemptedRevision: ghPortalRepo.GetBranchSHA(t, ctx, "feature2"),
								conditions: map[string]*metav1.Condition{
									apiv1.FailedToInitialize: nil,
									apiv1.Finalizing:         nil,
									apiv1.Stale:              nil,
								},
							},
							{
								repositoryName:        kServerRepoName,
								branch:                "main",
								lastAppliedRevision:   ghServerRepo.GetBranchSHA(t, ctx, "main"),
								lastAttemptedRevision: ghServerRepo.GetBranchSHA(t, ctx, "main"),
								conditions: map[string]*metav1.Condition{
									apiv1.FailedToInitialize: nil,
									apiv1.Finalizing:         nil,
									apiv1.Stale:              nil,
								},
							},
						},
					},
					{
						preferredBranch: "feature3",
						conditions: map[string]*metav1.Condition{
							apiv1.FailedToInitialize: nil,
							apiv1.Finalizing:         nil,
							apiv1.Stale:              nil,
						},
						deployments: []deploymentExpectation{
							{
								repositoryName:        kCommonRepoName,
								branch:                "main",
								lastAppliedRevision:   ghCommonRepo.GetBranchSHA(t, ctx, "main"),
								lastAttemptedRevision: ghCommonRepo.GetBranchSHA(t, ctx, "main"),
								conditions: map[string]*metav1.Condition{
									apiv1.FailedToInitialize: nil,
									apiv1.Finalizing:         nil,
									apiv1.Stale:              nil,
								},
							},
							{
								repositoryName:        kPortalRepoName,
								branch:                "main",
								lastAppliedRevision:   ghPortalRepo.GetBranchSHA(t, ctx, "main"),
								lastAttemptedRevision: ghPortalRepo.GetBranchSHA(t, ctx, "main"),
								conditions: map[string]*metav1.Condition{
									apiv1.FailedToInitialize: nil,
									apiv1.Finalizing:         nil,
									apiv1.Stale:              nil,
								},
							},
							{
								repositoryName:        kServerRepoName,
								branch:                "feature3",
								lastAppliedRevision:   ghServerRepo.GetBranchSHA(t, ctx, "feature3"),
								lastAttemptedRevision: ghServerRepo.GetBranchSHA(t, ctx, "feature3"),
								conditions: map[string]*metav1.Condition{
									apiv1.FailedToInitialize: nil,
									apiv1.Finalizing:         nil,
									apiv1.Stale:              nil,
								},
							},
						},
					},
				},
			},
		})
	}).Will(Eventually(Succeed()).Within(5 * time.Minute).ProbingEvery(100 * time.Millisecond))
}

func validateApplicationExpectations(t JustT, ctx context.Context, k *KClient, ns *KNamespace, expectations applicationExpectations) {

	// Fetch applications
	applications := &apiv1.ApplicationList{}
	For(t).Expect(k.Client.List(ctx, applications, client.InNamespace(ns.Name))).Will(Succeed())

	// Validate applications
	appList := &apiv1.ApplicationList{}
	For(t).Expect(k.Client.List(ctx, appList, client.InNamespace(ns.Name))).Will(Succeed())
	for _, app := range appList.Items {

		// Verify application
		appExpectation := expectations[app.Name]
		delete(expectations, app.Name)
		for conditionType, condition := range appExpectation.conditions {
			For(t).Expect(app.Status.GetCondition(conditionType)).Will(EqualCondition(condition))
			delete(appExpectation.conditions, conditionType)
		}
		if len(appExpectation.conditions) > 0 {
			t.Fatalf("Missing application conditions: %+v", appExpectation.conditions)
		}

		// Verify environments
		envs := &apiv1.EnvironmentList{}
		For(t).Expect(k.Client.List(ctx, envs, client.InNamespace(app.Namespace), k8s.OwnedBy(k.Client.Scheme(), &app))).Will(Succeed())
		for _, env := range envs.Items {
			var envExpectation *environmentExpectation
			for i, e := range appExpectation.environments {
				if env.Spec.PreferredBranch == e.preferredBranch {
					envExpectation = &e
					appExpectation.environments = append(appExpectation.environments[:i], appExpectation.environments[i+1:]...)
					break
				}
			}
			For(t).Expect(envExpectation).WillNot(BeNil())
			for conditionType, condition := range envExpectation.conditions {
				For(t).Expect(env.Status.GetCondition(conditionType)).Will(EqualCondition(condition))
				delete(envExpectation.conditions, conditionType)
			}
			if len(envExpectation.conditions) > 0 {
				t.Fatalf("Missing environment conditions: %+v", envExpectation.conditions)
			}

			// Verify deployments
			deployments := &apiv1.DeploymentList{}
			For(t).Expect(k.Client.List(ctx, deployments, client.InNamespace(app.Namespace), k8s.OwnedBy(k.Client.Scheme(), &env))).Will(Succeed())
			for _, deployment := range deployments.Items {
				For(t).Expect(deployment.Spec.Repository.Namespace).Will(BeEqualTo(ns.Name))

				var depExpectation *deploymentExpectation
				for i, e := range envExpectation.deployments {
					if deployment.Spec.Repository.Name == e.repositoryName {
						depExpectation = &e
						envExpectation.deployments = append(envExpectation.deployments[:i], envExpectation.deployments[i+1:]...)
						break
					}
				}
				For(t).Expect(depExpectation).WillNot(BeNil())
				For(t).Expect(deployment.Spec.Repository.Name).Will(BeEqualTo(depExpectation.repositoryName))
				For(t).Expect(deployment.Status.Branch).Will(BeEqualTo(depExpectation.branch))
				For(t).Expect(deployment.Status.LastAppliedRevision).Will(BeEqualTo(depExpectation.lastAppliedRevision))
				For(t).Expect(deployment.Status.LastAttemptedRevision).Will(BeEqualTo(depExpectation.lastAttemptedRevision))
				For(t).Expect(deployment.Status.PersistentVolumeClaimName).WillNot(BeEmpty())
				for conditionType, condition := range depExpectation.conditions {
					For(t).Expect(deployment.Status.GetCondition(conditionType)).Will(EqualCondition(condition))
					delete(depExpectation.conditions, conditionType)
				}
				if len(depExpectation.conditions) > 0 {
					t.Fatalf("Missing deployment conditions: %+v", depExpectation.conditions)
				}

				// TODO: create a pod that actually inspects the persistent volume, or alternatively - inspects the target cluster!
			}
			if len(envExpectation.deployments) > 0 {
				t.Fatalf("Missing deployments: %+v", envExpectation.deployments)
			}
		}
		if len(appExpectation.environments) > 0 {
			t.Fatalf("Missing environments: %+v", appExpectation.environments)
		}
	}
	if len(expectations) > 0 {
		t.Fatalf("Missing applications: %+v", expectations)
	}
}
