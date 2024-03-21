package controllers_test

import (
	"context"
	apiv1 "github.com/arikkfir/devbot/backend/api/v1"
	"github.com/arikkfir/devbot/backend/internal/util/k8s"
	. "github.com/arikkfir/devbot/backend/internal/util/testing"
	. "github.com/arikkfir/devbot/backend/internal/util/testing/justest"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"testing"
	"time"
)

type (
	conditionExpectations map[string]*metav1.Condition
	repositoryExpectation struct {
		conditions    conditionExpectations
		defaultBranch string
		revisions     map[string]string
	}
	repositoryExpectations map[string]repositoryExpectation
	deploymentExpectation  struct {
		branch                string
		clonePath             string
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
				For(t).Expect(deployment.Spec.Repository.APIVersion).Will(BeEqualTo(apiv1.RepositoryGVK.GroupVersion().String()))
				For(t).Expect(deployment.Spec.Repository.Kind).Will(BeEqualTo(apiv1.RepositoryGVK.Kind))
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
				For(t).Expect(deployment.Status.ClonePath).Will(BeEqualTo(depExpectation.clonePath))
				For(t).Expect(deployment.Status.LastAppliedRevision).Will(BeEqualTo(depExpectation.lastAppliedRevision))
				For(t).Expect(deployment.Status.LastAttemptedRevision).Will(BeEqualTo(depExpectation.lastAttemptedRevision))
				for conditionType, condition := range depExpectation.conditions {
					For(t).Expect(deployment.Status.GetCondition(conditionType)).Will(EqualCondition(condition))
					delete(depExpectation.conditions, conditionType)
				}
				if len(depExpectation.conditions) > 0 {
					t.Fatalf("Missing deployment conditions: %+v", depExpectation.conditions)
				}
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

func TestApplicationDeploymentReconciliation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	k := NewKubernetes(t)
	ns := k.CreateNamespace(t, ctx)
	gh := NewGitHub(t, ctx)

	ghCommonRepo := gh.CreateRepository(t, ctx, "app1/common")
	kCommonRepoName := ns.CreateRepository(t, ctx, apiv1.RepositorySpec{
		GitHub: &apiv1.GitHubRepositorySpec{
			Owner:               ghCommonRepo.Owner,
			Name:                ghCommonRepo.Name,
			PersonalAccessToken: ns.CreateGitHubAuthSecretSpec(t, ctx, gh.Token, true),
		},
		RefreshInterval: "10s",
	})

	ghServerRepo := gh.CreateRepository(t, ctx, "app1/server")
	kServerRepoName := ns.CreateRepository(t, ctx, apiv1.RepositorySpec{
		GitHub: &apiv1.GitHubRepositorySpec{
			Owner:               ghServerRepo.Owner,
			Name:                ghServerRepo.Name,
			PersonalAccessToken: ns.CreateGitHubAuthSecretSpec(t, ctx, gh.Token, true),
		},
		RefreshInterval: "10s",
	})

	appName := ns.CreateApplication(t, ctx, apiv1.ApplicationSpec{
		Repositories: []apiv1.ApplicationSpecRepository{
			{
				Name:                  kCommonRepoName,
				Namespace:             ns.Name,
				MissingBranchStrategy: "UseDefaultBranch",
			},
			{
				Name:                  kServerRepoName,
				Namespace:             ns.Name,
				MissingBranchStrategy: "UseDefaultBranch",
			},
		},
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
								clonePath:             "/a",
								lastAppliedRevision:   ghCommonRepo.GetBranchSHA(t, ctx, "main"),
								lastAttemptedRevision: ghCommonRepo.GetBranchSHA(t, ctx, "main"),
								conditions: map[string]*metav1.Condition{
									apiv1.FailedToInitialize: nil,
									apiv1.Finalizing:         nil,
									apiv1.Stale:              nil,
								},
							},
							{
								repositoryName:        kServerRepoName,
								branch:                "main",
								clonePath:             "/b",
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
	}).Will(Eventually(Succeed()).Within(5 * time.Minute).ProbingEvery(100 * time.Millisecond))

	// Create new branches
	_ = ghCommonRepo.CreateBranch(t, ctx, "feature1")
	ghCommonRepo.CreateFile(t, ctx, "feature1")
	_ = ghServerRepo.CreateBranch(t, ctx, "feature2")
	ghServerRepo.CreateFile(t, ctx, "feature2")

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
								clonePath:             "/a",
								lastAppliedRevision:   ghCommonRepo.GetBranchSHA(t, ctx, "main"),
								lastAttemptedRevision: ghCommonRepo.GetBranchSHA(t, ctx, "main"),
								conditions: map[string]*metav1.Condition{
									apiv1.FailedToInitialize: nil,
									apiv1.Finalizing:         nil,
									apiv1.Stale:              nil,
								},
							},
							{
								repositoryName:        kServerRepoName,
								branch:                "main",
								clonePath:             "/b",
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
								clonePath:             "/a",
								lastAppliedRevision:   ghCommonRepo.GetBranchSHA(t, ctx, "feature1"),
								lastAttemptedRevision: ghCommonRepo.GetBranchSHA(t, ctx, "feature1"),
								conditions: map[string]*metav1.Condition{
									apiv1.FailedToInitialize: nil,
									apiv1.Finalizing:         nil,
									apiv1.Stale:              nil,
								},
							},
							{
								repositoryName:        kServerRepoName,
								branch:                "main",
								clonePath:             "/b",
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
								clonePath:             "/a",
								lastAppliedRevision:   ghCommonRepo.GetBranchSHA(t, ctx, "main"),
								lastAttemptedRevision: ghCommonRepo.GetBranchSHA(t, ctx, "main"),
								conditions: map[string]*metav1.Condition{
									apiv1.FailedToInitialize: nil,
									apiv1.Finalizing:         nil,
									apiv1.Stale:              nil,
								},
							},
							{
								repositoryName:        kServerRepoName,
								branch:                "feature2",
								clonePath:             "/b",
								lastAppliedRevision:   ghServerRepo.GetBranchSHA(t, ctx, "feature2"),
								lastAttemptedRevision: ghServerRepo.GetBranchSHA(t, ctx, "feature2"),
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
