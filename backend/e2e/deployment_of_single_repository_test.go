package e2e_test

import (
	apiv1 "github.com/arikkfir/devbot/backend/api/v1"
	. "github.com/arikkfir/devbot/backend/e2e/expectations"
	. "github.com/arikkfir/justest"
	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"testing"
	"time"
)

// TODO: add test case for a repository with the "IgnoreStrategy" of missing branches

func TestSingleRepoApplicationDeployment(t *testing.T) {
	t.Parallel()
	e2e := NewE2E(t)
	ns := e2e.K.CreateNamespace(t)

	ghCommonRepo, kRepoName := e2e.CreateGitHubAndK8sRepository(t, ns, "common", "5s")
	appName := ns.CreateApplication(t, apiv1.ApplicationSpec{
		Repositories: []apiv1.ApplicationSpecRepository{
			{Name: kRepoName, Namespace: ns.Name, MissingBranchStrategy: apiv1.UseDefaultBranchStrategy},
		},
		ServiceAccountName: "devbot-gitops",
	})

	With(t).Verify(func(t T) {

		// Prepare fresh expectations
		applicationExpectations := []AppE{
			{
				Name: appName,
				Status: AppStatusE{
					Conditions: map[string]*ConditionE{
						apiv1.Invalid:            nil,
						apiv1.Finalizing:         nil,
						apiv1.FailedToInitialize: nil,
						apiv1.Stale:              nil,
					},
				},
				Environments: []EnvE{
					{
						Spec: EnvSpecE{PreferredBranch: "main"},
						Status: EnvStatusE{
							Conditions: map[string]*ConditionE{
								apiv1.Invalid:            nil,
								apiv1.Finalizing:         nil,
								apiv1.FailedToInitialize: nil,
								apiv1.Stale:              nil,
							},
						},
						Deployments: []DeploymentE{
							{
								Spec: DeploymentSpecE{
									Repository: apiv1.DeploymentRepositoryReference{Name: kRepoName, Namespace: ns.Name},
								},
								Status: DeploymentStatusE{
									Branch: "main",
									Conditions: map[string]*ConditionE{
										apiv1.Invalid:            nil,
										apiv1.Finalizing:         nil,
										apiv1.FailedToInitialize: nil,
										apiv1.Stale:              nil,
									},
									LastAttemptedRevision: ghCommonRepo.GetBranchSHA(t, "main"),
									LastAppliedRevision:   ghCommonRepo.GetBranchSHA(t, "main"),
									ResolvedRepository:    ns.Name + "/" + kRepoName,
								},
								Resources: []ResourceE{
									{
										Object:    &v1.ConfigMap{},
										Name:      "main-configuration",
										Namespace: ns.Name,
										Validator: func(t T, r ResourceE) {
											cm := r.Object.(*v1.ConfigMap)
											With(t).Verify(cm.Data["env"]).Will(EqualTo("main")).OrFail()
										},
									},
								},
							},
						},
					},
				},
			},
		}

		// Fetch applications & verify them
		appList := &apiv1.ApplicationList{}
		With(t).Verify(e2e.K.Client.List(e2e.Ctx, appList, client.InNamespace(ns.Name))).Will(Succeed()).OrFail()
		With(t).Verify(appList.Items).Will(EqualTo(applicationExpectations).Using(CreateApplicationsComparator(e2e.K.Client, e2e.Ctx))).OrFail()

	}).Will(Succeed()).Within(10*time.Minute, 100*time.Millisecond)

	//	// Now create a new branch, expecting a new environment & deployment will occur
	//	_ = ghRepo.CreateBranch(t, ctx, "feature1")
	//	ghRepo.CreateFile(t, ctx, "feature1")
	//
	//	With(t).Expect(func(t JustT) {
	//		validateApplications(t, ctx, k, ns, map[string]applicationExpectation{
	//			appName: {
	//				conditions: map[string]*metav1.Condition{
	//					apiv1.FailedToInitialize: nil,
	//					apiv1.Finalizing:         nil,
	//					apiv1.Stale:              nil,
	//				},
	//				environments: createEnvironmentExpectations("main", "feature1"),
	//				resources: []resourceExpectation{
	//					{
	//						name:      "main-configuration",
	//						namespace: ns.Name,
	//						object:    &v1.ConfigMap{},
	//						validator: func(ctx context.Context, t JustT, obj client.Object) {
	//							cm := obj.(*v1.ConfigMap)
	//							With(t).Expect(cm.Data).Will(BeEqualTo(map[string]string{"env": "main"}))
	//						},
	//					},
	//					{
	//						name:      "feature1-configuration",
	//						namespace: ns.Name,
	//						object:    &v1.ConfigMap{},
	//						validator: func(ctx context.Context, t JustT, obj client.Object) {
	//							cm := obj.(*v1.ConfigMap)
	//							With(t).Expect(cm.Data).Will(BeEqualTo(map[string]string{"env": "feature1"}))
	//						},
	//					},
	//				},
	//			},
	//		})
	//	}).Will(Eventually(Succeed()).Within(5 * time.Minute).ProbingEvery(100 * time.Millisecond))
	// }
}
