package e2e_test

import (
	apiv1 "github.com/arikkfir/devbot/backend/api/v1"
	. "github.com/arikkfir/devbot/backend/e2e/expectations"
	. "github.com/arikkfir/devbot/backend/internal/util/testing"
	. "github.com/arikkfir/devbot/backend/internal/util/testing/justest"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"testing"
	"time"
)

// TODO: add test case for a repository with the "IgnoreStrategy" of missing branches
// TODO: add test case with multiple applications

func TestMultiRepoApplicationDeployment(t *testing.T) {
	t.Parallel()

	ns := K(t).CreateNamespace(t)
	ghCommonRepo, kCommonRepoName := createGitHubAndK8sRepository(t, ns, "common")
	ghPortalRepo, kPortalRepoName := createGitHubAndK8sRepository(t, ns, "portal")
	ghServerRepo, kServerRepoName := createGitHubAndK8sRepository(t, ns, "server")

	appName := ns.CreateApplication(t, apiv1.ApplicationSpec{
		Repositories: []apiv1.ApplicationSpecRepository{
			{Name: kCommonRepoName, Namespace: ns.Name, MissingBranchStrategy: apiv1.UseDefaultBranchStrategy},
			{Name: kPortalRepoName, Namespace: ns.Name, MissingBranchStrategy: apiv1.UseDefaultBranchStrategy},
			{Name: kServerRepoName, Namespace: ns.Name, MissingBranchStrategy: apiv1.UseDefaultBranchStrategy},
		},
		ServiceAccountName: "devbot-gitops",
	})

	// Validate initial deployment
	For(t).Expect(func(t TT) {

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
									Repository: apiv1.DeploymentRepositoryReference{Name: kCommonRepoName, Namespace: ns.Name},
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
									ResolvedRepository:    ghCommonRepo.Owner + "/" + ghCommonRepo.Name,
								},
								Resources: []ResourceE{
									{
										Object:    &v1.ConfigMap{},
										Name:      "main-configuration",
										Namespace: ns.Name,
										Validator: func(t TT, r ResourceE) {
											cm := r.Object.(*v1.ConfigMap)
											For(t).Expect(cm.Data["env"]).Will(BeEqualTo("main"))
										},
									},
								},
							},
							{
								Spec: DeploymentSpecE{
									Repository: apiv1.DeploymentRepositoryReference{Name: kServerRepoName, Namespace: ns.Name},
								},
								Status: DeploymentStatusE{
									Branch: "main",
									Conditions: map[string]*ConditionE{
										apiv1.Invalid:            nil,
										apiv1.Finalizing:         nil,
										apiv1.FailedToInitialize: nil,
										apiv1.Stale:              nil,
									},
									LastAttemptedRevision: ghServerRepo.GetBranchSHA(t, "main"),
									LastAppliedRevision:   ghServerRepo.GetBranchSHA(t, "main"),
									ResolvedRepository:    ghServerRepo.Owner + "/" + ghServerRepo.Name,
								},
								Resources: []ResourceE{
									{
										Object:    &v1.ServiceAccount{},
										Name:      "main-server",
										Namespace: ns.Name,
									},
									{
										Object:    &v1.Service{},
										Name:      "main-server",
										Namespace: ns.Name,
										Validator: func(t TT, r ResourceE) {
											svc := r.Object.(*v1.Service)
											For(t).Expect(len(svc.Spec.Ports)).Will(BeEqualTo(1))
											For(t).Expect(svc.Spec.Ports[0].Port).Will(BeEqualTo(80))
											For(t).Expect(svc.Spec.Ports[0].TargetPort.Type).Will(BeEqualTo(intstr.String))
											For(t).Expect(svc.Spec.Ports[0].TargetPort.StrVal).Will(BeEqualTo("http"))
										},
									},
									{
										Object:    &appsv1.Deployment{},
										Name:      "main-server",
										Namespace: ns.Name,
										Validator: func(t TT, r ResourceE) {
											d := r.Object.(*appsv1.Deployment)
											For(t).Expect(len(d.Spec.Template.Spec.Containers)).Will(BeEqualTo(1))
											For(t).Expect(len(d.Spec.Template.Spec.Containers[0].Image)).Will(BeEqualTo("ealen/echo-server:latest"))
										},
									},
								},
							},
							{
								Spec: DeploymentSpecE{
									Repository: apiv1.DeploymentRepositoryReference{Name: kPortalRepoName, Namespace: ns.Name},
								},
								Status: DeploymentStatusE{
									Branch: "main",
									Conditions: map[string]*ConditionE{
										apiv1.Invalid:            nil,
										apiv1.Finalizing:         nil,
										apiv1.FailedToInitialize: nil,
										apiv1.Stale:              nil,
									},
									LastAttemptedRevision: ghPortalRepo.GetBranchSHA(t, "main"),
									LastAppliedRevision:   ghPortalRepo.GetBranchSHA(t, "main"),
									ResolvedRepository:    ghPortalRepo.Owner + "/" + ghPortalRepo.Name,
								},
								Resources: []ResourceE{
									{
										Object:    &v1.ServiceAccount{},
										Name:      "main-portal",
										Namespace: ns.Name,
									},
									{
										Object:    &v1.Service{},
										Name:      "main-portal",
										Namespace: ns.Name,
										Validator: func(t TT, r ResourceE) {
											svc := r.Object.(*v1.Service)
											For(t).Expect(len(svc.Spec.Ports)).Will(BeEqualTo(1))
											For(t).Expect(svc.Spec.Ports[0].Port).Will(BeEqualTo(80))
											For(t).Expect(svc.Spec.Ports[0].TargetPort.Type).Will(BeEqualTo(intstr.String))
											For(t).Expect(svc.Spec.Ports[0].TargetPort.StrVal).Will(BeEqualTo("http"))
										},
									},
									{
										Object:    &appsv1.Deployment{},
										Name:      "main-portal",
										Namespace: ns.Name,
										Validator: func(t TT, r ResourceE) {
											d := r.Object.(*appsv1.Deployment)
											For(t).Expect(len(d.Spec.Template.Spec.Containers)).Will(BeEqualTo(1))
											For(t).Expect(len(d.Spec.Template.Spec.Containers[0].Image)).Will(BeEqualTo("nginx"))
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
		For(t).Expect(K(t).Client.List(t, appList, client.InNamespace(ns.Name))).Will(Succeed())
		For(t).Expect(appList.Items).Will(CompareTo(applicationExpectations).Using(ApplicationsComparator))

	}).Will(Eventually(Succeed()).Within(10 * time.Second).ProbingEvery(100 * time.Millisecond))

	// Create new branches
	//_ = ghCommonRepo.CreateBranch(t, ctx, "feature1")
	//ghCommonRepo.CreateFile(t, ctx, "feature1")
	//_ = ghPortalRepo.CreateBranch(t, ctx, "feature2")
	//ghPortalRepo.CreateFile(t, ctx, "feature2")
	//_ = ghServerRepo.CreateBranch(t, ctx, "feature3")
	//ghServerRepo.CreateFile(t, ctx, "feature3")
	//
	//// Creates expected resources of given environment(s)
	//resourcesForEnv := func(environments ...string) []resourceExpectation {
	//	var expectations []resourceExpectation
	//	for _, env := range environments {
	//		expectations = append(expectations, []resourceExpectation{
	//			{
	//				name:      env + "-configuration",
	//				namespace: ns.Name,
	//				object:    &v1.ConfigMap{},
	//				validator: func(ctx context.Context, t JustT, obj client.Object) {
	//					cm := obj.(*v1.ConfigMap)
	//					For(t).Expect(cm.Data).Will(BeEqualTo(map[string]string{"env": "main"}))
	//				},
	//			},
	//			{
	//				name:      env + "-portal",
	//				namespace: ns.Name,
	//				object:    &appsv1.Deployment{},
	//				validator: func(ctx context.Context, t JustT, obj client.Object) {
	//					d := obj.(*appsv1.Deployment)
	//					For(t).Expect(d.Spec.Template.Spec.Containers[0].Name).Will(BeEqualTo("portal"))
	//				},
	//			},
	//			{
	//				name:      env + "-server",
	//				namespace: ns.Name,
	//				object:    &appsv1.Deployment{},
	//				validator: func(ctx context.Context, t JustT, obj client.Object) {
	//					d := obj.(*appsv1.Deployment)
	//					For(t).Expect(d.Spec.Template.Spec.Containers[0].Name).Will(BeEqualTo("server"))
	//				},
	//			},
	//		}...)
	//	}
	//	return expectations
	//}
	//
	//// Validate new environments & deployments created & deployed
	//For(t).Expect(func(t JustT) {
	//	validateApplications(t, ctx, k, ns, map[string]applicationExpectation{
	//		appName: {
	//			conditions: map[string]*metav1.Condition{
	//				apiv1.FailedToInitialize: nil,
	//				apiv1.Finalizing:         nil,
	//				apiv1.Stale:              nil,
	//			},
	//			environments: []environmentExpectation{
	//				{
	//					preferredBranch: "main",
	//					conditions: map[string]*metav1.Condition{
	//						apiv1.FailedToInitialize: nil,
	//						apiv1.Finalizing:         nil,
	//						apiv1.Stale:              nil,
	//					},
	//					deployments: []deploymentExpectation{
	//						{
	//							repositoryName:        kCommonRepoName,
	//							branch:                "main",
	//							lastAppliedRevision:   ghCommonRepo.GetBranchSHA(t, ctx, "main"),
	//							lastAttemptedRevision: ghCommonRepo.GetBranchSHA(t, ctx, "main"),
	//							conditions: map[string]*metav1.Condition{
	//								apiv1.FailedToInitialize: nil,
	//								apiv1.Finalizing:         nil,
	//								apiv1.Stale:              nil,
	//							},
	//						},
	//						{
	//							repositoryName:        kPortalRepoName,
	//							branch:                "main",
	//							lastAppliedRevision:   ghPortalRepo.GetBranchSHA(t, ctx, "main"),
	//							lastAttemptedRevision: ghPortalRepo.GetBranchSHA(t, ctx, "main"),
	//							conditions: map[string]*metav1.Condition{
	//								apiv1.FailedToInitialize: nil,
	//								apiv1.Finalizing:         nil,
	//								apiv1.Stale:              nil,
	//							},
	//						},
	//						{
	//							repositoryName:        kServerRepoName,
	//							branch:                "main",
	//							lastAppliedRevision:   ghServerRepo.GetBranchSHA(t, ctx, "main"),
	//							lastAttemptedRevision: ghServerRepo.GetBranchSHA(t, ctx, "main"),
	//							conditions: map[string]*metav1.Condition{
	//								apiv1.FailedToInitialize: nil,
	//								apiv1.Finalizing:         nil,
	//								apiv1.Stale:              nil,
	//							},
	//						},
	//					},
	//				},
	//				{
	//					preferredBranch: "feature1",
	//					conditions: map[string]*metav1.Condition{
	//						apiv1.FailedToInitialize: nil,
	//						apiv1.Finalizing:         nil,
	//						apiv1.Stale:              nil,
	//					},
	//					deployments: []deploymentExpectation{
	//						{
	//							repositoryName:        kCommonRepoName,
	//							branch:                "feature1",
	//							lastAppliedRevision:   ghCommonRepo.GetBranchSHA(t, ctx, "feature1"),
	//							lastAttemptedRevision: ghCommonRepo.GetBranchSHA(t, ctx, "feature1"),
	//							conditions: map[string]*metav1.Condition{
	//								apiv1.FailedToInitialize: nil,
	//								apiv1.Finalizing:         nil,
	//								apiv1.Stale:              nil,
	//							},
	//						},
	//						{
	//							repositoryName:        kPortalRepoName,
	//							branch:                "main",
	//							lastAppliedRevision:   ghPortalRepo.GetBranchSHA(t, ctx, "main"),
	//							lastAttemptedRevision: ghPortalRepo.GetBranchSHA(t, ctx, "main"),
	//							conditions: map[string]*metav1.Condition{
	//								apiv1.FailedToInitialize: nil,
	//								apiv1.Finalizing:         nil,
	//								apiv1.Stale:              nil,
	//							},
	//						},
	//						{
	//							repositoryName:        kServerRepoName,
	//							branch:                "main",
	//							lastAppliedRevision:   ghServerRepo.GetBranchSHA(t, ctx, "main"),
	//							lastAttemptedRevision: ghServerRepo.GetBranchSHA(t, ctx, "main"),
	//							conditions: map[string]*metav1.Condition{
	//								apiv1.FailedToInitialize: nil,
	//								apiv1.Finalizing:         nil,
	//								apiv1.Stale:              nil,
	//							},
	//						},
	//					},
	//				},
	//				{
	//					preferredBranch: "feature2",
	//					conditions: map[string]*metav1.Condition{
	//						apiv1.FailedToInitialize: nil,
	//						apiv1.Finalizing:         nil,
	//						apiv1.Stale:              nil,
	//					},
	//					deployments: []deploymentExpectation{
	//						{
	//							repositoryName:        kCommonRepoName,
	//							branch:                "main",
	//							lastAppliedRevision:   ghCommonRepo.GetBranchSHA(t, ctx, "main"),
	//							lastAttemptedRevision: ghCommonRepo.GetBranchSHA(t, ctx, "main"),
	//							conditions: map[string]*metav1.Condition{
	//								apiv1.FailedToInitialize: nil,
	//								apiv1.Finalizing:         nil,
	//								apiv1.Stale:              nil,
	//							},
	//						},
	//						{
	//							repositoryName:        kPortalRepoName,
	//							branch:                "feature2",
	//							lastAppliedRevision:   ghPortalRepo.GetBranchSHA(t, ctx, "feature2"),
	//							lastAttemptedRevision: ghPortalRepo.GetBranchSHA(t, ctx, "feature2"),
	//							conditions: map[string]*metav1.Condition{
	//								apiv1.FailedToInitialize: nil,
	//								apiv1.Finalizing:         nil,
	//								apiv1.Stale:              nil,
	//							},
	//						},
	//						{
	//							repositoryName:        kServerRepoName,
	//							branch:                "main",
	//							lastAppliedRevision:   ghServerRepo.GetBranchSHA(t, ctx, "main"),
	//							lastAttemptedRevision: ghServerRepo.GetBranchSHA(t, ctx, "main"),
	//							conditions: map[string]*metav1.Condition{
	//								apiv1.FailedToInitialize: nil,
	//								apiv1.Finalizing:         nil,
	//								apiv1.Stale:              nil,
	//							},
	//						},
	//					},
	//				},
	//				{
	//					preferredBranch: "feature3",
	//					conditions: map[string]*metav1.Condition{
	//						apiv1.FailedToInitialize: nil,
	//						apiv1.Finalizing:         nil,
	//						apiv1.Stale:              nil,
	//					},
	//					deployments: []deploymentExpectation{
	//						{
	//							repositoryName:        kCommonRepoName,
	//							branch:                "main",
	//							lastAppliedRevision:   ghCommonRepo.GetBranchSHA(t, ctx, "main"),
	//							lastAttemptedRevision: ghCommonRepo.GetBranchSHA(t, ctx, "main"),
	//							conditions: map[string]*metav1.Condition{
	//								apiv1.FailedToInitialize: nil,
	//								apiv1.Finalizing:         nil,
	//								apiv1.Stale:              nil,
	//							},
	//						},
	//						{
	//							repositoryName:        kPortalRepoName,
	//							branch:                "main",
	//							lastAppliedRevision:   ghPortalRepo.GetBranchSHA(t, ctx, "main"),
	//							lastAttemptedRevision: ghPortalRepo.GetBranchSHA(t, ctx, "main"),
	//							conditions: map[string]*metav1.Condition{
	//								apiv1.FailedToInitialize: nil,
	//								apiv1.Finalizing:         nil,
	//								apiv1.Stale:              nil,
	//							},
	//						},
	//						{
	//							repositoryName:        kServerRepoName,
	//							branch:                "feature3",
	//							lastAppliedRevision:   ghServerRepo.GetBranchSHA(t, ctx, "feature3"),
	//							lastAttemptedRevision: ghServerRepo.GetBranchSHA(t, ctx, "feature3"),
	//							conditions: map[string]*metav1.Condition{
	//								apiv1.FailedToInitialize: nil,
	//								apiv1.Finalizing:         nil,
	//								apiv1.Stale:              nil,
	//							},
	//						},
	//					},
	//				},
	//			},
	//			resources: resourcesForEnv("main", "feature1", "feature2", "feature3"),
	//		},
	//	})
	//}).Will(Eventually(Succeed()).Within(5 * time.Minute).ProbingEvery(100 * time.Millisecond))
}