package application

import (
	"testing"
)

func TestAppCreatesDistinctEnvsForBranches(t *testing.T) {
	//ghRepo1, err := gh.CreateRepository(ctx)
	//if err != nil {
	//	t.Fatalf("Failed to create GitHub repository: %+v", err)
	//}
	//ghRepoCR1, err := k.CreateGitHubRepository(ctx, "default", *ghRepo1.Owner.Login, *ghRepo1.Name)
	//if err != nil {
	//	t.Fatalf("Failed to create GitHubRepository object: %+v", err)
	//}
	//ghRepoCR1NamespacedName := types.NamespacedName{Namespace: ghRepoCR1.Namespace, Name: ghRepoCR1.Name}
	//
	//ghRepo2, err := gh.CreateRepository(ctx)
	//if err != nil {
	//	t.Fatalf("Failed to create GitHub repository: %+v", err)
	//}
	//ghRepoCR2, err := k.CreateGitHubRepository(ctx, "default", *ghRepo2.Owner.Login, *ghRepo2.Name)
	//if err != nil {
	//	t.Fatalf("Failed to create GitHubRepository object: %+v", err)
	//}
	//ghRepoCR2NamespacedName := types.NamespacedName{Namespace: ghRepoCR2.Namespace, Name: ghRepoCR2.Name}
	//
	//if err := gh.CreateRepositoryWebhook(ctx, *ghRepo1.Name, k.gitHubAuthSecrets["WEBHOOK_SECRET"], "push"); err != nil {
	//	t.Fatalf("Failed to create GitHub repository webhook: %+v", err)
	//} else if err := gh.CreateRepositoryWebhook(ctx, *ghRepo2.Name, k.gitHubAuthSecrets["WEBHOOK_SECRET"], "push"); err != nil {
	//	t.Fatalf("Failed to create GitHub repository webhook: %+v", err)
	//}
	//
	//if _, err := gh.CreateFile(ctx, *ghRepo1.Name, "main", "README.md", t.Name()); err != nil {
	//	t.Fatalf("Failed to create GitHub file: %+v", err)
	//} else if _, err := gh.CreateFile(ctx, *ghRepo1.Name, "feat1", "main.go", t.Name()); err != nil {
	//	t.Fatalf("Failed to create GitHub file: %+v", err)
	//}
	//
	//if _, err := gh.CreateFile(ctx, *ghRepo2.Name, "main", "README.md", t.Name()); err != nil {
	//	t.Fatalf("Failed to create GitHub file: %+v", err)
	//} else if _, err := gh.CreateFile(ctx, *ghRepo2.Name, "feat1", "main.go", t.Name()); err != nil {
	//	t.Fatalf("Failed to create GitHub file: %+v", err)
	//} else if _, err := gh.CreateFile(ctx, *ghRepo2.Name, "feat2", "main.go", t.Name()); err != nil {
	//	t.Fatalf("Failed to create GitHub file: %+v", err)
	//}
	//
	//app, err := k.CreateApplication(ctx, "default", ghRepoCR1NamespacedName, ghRepoCR2NamespacedName)
	//if err != nil {
	//	t.Fatalf("Failed to create Application object: %+v", err)
	//}
	//
	//a := &apiv1.Application{}
	//if err := k.c.Get(ctx, types.NamespacedName{Namespace: app.Namespace, Name: app.Name}, a); err != nil {
	//	t.Errorf("Failed getting app '%s/%s' object: %+v", ghRepoCR1.Namespace, ghRepoCR1.Name, err)
	//} else if a.GetStatusConditionValid() == nil {
	//	t.Errorf("Application object '%s/%s' is not valid", a.Namespace, a.Name)
	//} else if a.GetStatusConditionValid().Status != v1.ConditionTrue {
	//	t.Errorf("Application object '%s/%s' is not valid: %+v", a.Namespace, a.Name, a.GetStatusConditionValid())
	//}

	//timeout := 1 * time.Minute
	//interval := 2 * time.Second
	//started := time.Now()
	//for {
	//	if started.Add(timeout).Before(time.Now()) {
	//		t.Errorf("Timed out waiting for Application object to become valid")
	//		break
	//	}
	//
	//	a := &apiv1.Application{}
	//	if err := k.c.Get(ctx, types.NamespacedName{Namespace: app.Namespace, Name: app.Name}, a); err != nil {
	//		t.Errorf("Failed getting app '%s/%s' object: %+v", ghRepoCR1.Namespace, ghRepoCR1.Name, err)
	//		time.Sleep(interval)
	//		continue
	//	} else if a.GetStatusConditionValid() == nil {
	//		t.Errorf("Application object '%s/%s' is not valid", a.Namespace, a.Name)
	//		time.Sleep(interval)
	//		continue
	//	} else if a.GetStatusConditionValid().Status != v1.ConditionTrue {
	//		t.Errorf("Application object '%s/%s' is not valid: %+v", a.Namespace, a.Name, a.GetStatusConditionValid())
	//		time.Sleep(interval)
	//		continue
	//	}
	//
	//	var envs []apiv1.ApplicationEnvironment
	//	envsList := &apiv1.ApplicationEnvironmentList{}
	//	if err := k.c.List(ctx, envsList); err != nil {
	//		t.Errorf("Failed getting list of ApplicationEnvironment objects: %+v", err)
	//		time.Sleep(interval)
	//		continue
	//	}
	//	ownerKey := util.GetOwnerRefKey(scheme, app)
	//	for _, env := range envsList.Items {
	//		ownerKeys := util.GetOwnerReferenceKeys(&env)
	//		if slices.Contains(ownerKeys, ownerKey) {
	//			envs = append(envs, env)
	//		}
	//	}
	//
	//	if len(envs) != 3 {
	//		t.Errorf("Incorrect number of ApplicationEnvironment objects: %d", len(envsList.Items))
	//		time.Sleep(interval)
	//		continue
	//	}
	//
	//	slices.SortFunc(envs, func(a, b apiv1.ApplicationEnvironment) int {
	//		ba := a.Spec.Branch
	//		bb := b.Spec.Branch
	//		if ba < bb {
	//			return -1
	//		} else if ba > bb {
	//			return 1
	//		} else {
	//			return 0
	//		}
	//	})
	//	expectedBranches := []string{"feat1", "feat2", "main"}
	//
	//	for i, env := range envs {
	//		if env.GetStatusConditionValid() == nil {
	//			t.Errorf("ApplicationEnvironment object '%s/%s' has nil Valid condition", env.Namespace, env.Name)
	//			time.Sleep(interval)
	//			continue
	//		} else if env.GetStatusConditionValid().Status != v1.ConditionTrue {
	//			t.Errorf("ApplicationEnvironment object '%s/%s' is not valid: %+v", env.Namespace, env.Name, env.GetStatusConditionValid())
	//			time.Sleep(interval)
	//			continue
	//		} else if env.Spec.Branch != expectedBranches[i] {
	//			t.Errorf("Incorrect branch for ApplicationEnvironment[%d]: expected '%s', was '%s'", i, expectedBranches[i], env.Spec.Branch)
	//			time.Sleep(interval)
	//			continue
	//		}
	//
	//		deploymentsList := &apiv1.DeploymentList{}
	//		if err := k.c.List(ctx, deploymentsList); err != nil {
	//			t.Errorf("Failed getting list of Deployment objects: %+v", err)
	//			time.Sleep(interval)
	//			continue
	//		}
	//		ownerKey := util.GetOwnerRefKey(scheme, &env)
	//		var deployments []*apiv1.Deployment
	//		for _, d := range deploymentsList.Items {
	//			ownerKeys := util.GetOwnerReferenceKeys(&d)
	//			if slices.Contains(ownerKeys, ownerKey) {
	//				deploymentCopy := d
	//				deployments = append(deployments, &deploymentCopy)
	//				break // TODO: remove this...
	//			}
	//		}
	//		if len(deployments) != 2 {
	//			// TODO: t.Fatalf("Incorrect number of Deployment objects found for environment '%s/%s': expected %d, was %d", env.Namespace, env.Name, 2, len(deployments))
	//			t.Errorf("Incorrect number of Deployment objects found for environment '%s/%s': expected %d, was %d", env.Namespace, env.Name, 2, len(deployments))
	//			time.Sleep(interval)
	//			continue
	//		}
	//	}
	//}

	//util.Eventually(t, 1*time.Minute, 2*time.Second, func(t util.TestingT) {
	//	a := &apiv1.Application{}
	//	if err := k.c.Get(ctx, types.NamespacedName{Namespace: app.Namespace, Name: app.Name}, a); err != nil {
	//		t.Fatalf("Failed getting app '%s/%s' object: %+v", ghRepoCR1.Namespace, ghRepoCR1.Name, err)
	//	} else if a.GetStatusConditionValid() == nil {
	//		t.Fatalf("Application object '%s/%s' is not valid", a.Namespace, a.Name)
	//	} else if a.GetStatusConditionValid().Status != v1.ConditionTrue {
	//		t.Fatalf("Application object '%s/%s' is not valid: %+v", a.Namespace, a.Name, a.GetStatusConditionValid())
	//	}
	//
	//	var envs []apiv1.ApplicationEnvironment
	//	envsList := &apiv1.ApplicationEnvironmentList{}
	//	if err := k.c.List(ctx, envsList); err != nil {
	//		t.Fatalf("Failed getting list of ApplicationEnvironment objects: %+v", err)
	//	}
	//	ownerKey := util.GetOwnerRefKey(scheme, app)
	//	for _, env := range envsList.Items {
	//		ownerKeys := util.GetOwnerReferenceKeys(&env)
	//		if slices.Contains(ownerKeys, ownerKey) {
	//			envs = append(envs, env)
	//		}
	//	}
	//
	//	if len(envs) != 3 {
	//		t.Fatalf("Incorrect number of ApplicationEnvironment objects: %d", len(envsList.Items))
	//	}
	//
	//	slices.SortFunc(envs, func(a, b apiv1.ApplicationEnvironment) int {
	//		ba := a.Spec.Branch
	//		bb := b.Spec.Branch
	//		if ba < bb {
	//			return -1
	//		} else if ba > bb {
	//			return 1
	//		} else {
	//			return 0
	//		}
	//	})
	//	expectedBranches := []string{"feat1", "feat2", "main"}
	//
	//	for i, env := range envs {
	//		if env.GetStatusConditionValid() == nil {
	//			t.Fatalf("ApplicationEnvironment object '%s/%s' has nil Valid condition", env.Namespace, env.Name)
	//		} else if env.GetStatusConditionValid().Status != v1.ConditionTrue {
	//			t.Fatalf("ApplicationEnvironment object '%s/%s' is not valid: %+v", env.Namespace, env.Name, env.GetStatusConditionValid())
	//		} else if env.Spec.Branch != expectedBranches[i] {
	//			t.Fatalf("Incorrect branch for ApplicationEnvironment[%d]: expected '%s', was '%s'", i, expectedBranches[i], env.Spec.Branch)
	//		}
	//
	//		deploymentsList := &apiv1.DeploymentList{}
	//		if err := k.c.List(ctx, deploymentsList); err != nil {
	//			t.Fatalf("Failed getting list of Deployment objects: %+v", err)
	//		}
	//		ownerKey := util.GetOwnerRefKey(scheme, &env)
	//		var deployments []*apiv1.Deployment
	//		for _, d := range deploymentsList.Items {
	//			ownerKeys := util.GetOwnerReferenceKeys(&d)
	//			if slices.Contains(ownerKeys, ownerKey) {
	//				deploymentCopy := d
	//				deployments = append(deployments, &deploymentCopy)
	//				break // TODO: remove this...
	//			}
	//		}
	//		//if len(deployments) != 2 {
	//		//	// TODO: t.Fatalf("Incorrect number of Deployment objects found for environment '%s/%s': expected %d, was %d", env.Namespace, env.Name, 2, len(deployments))
	//		//	t.Errorf("Incorrect number of Deployment objects found for environment '%s/%s': expected %d, was %d", env.Namespace, env.Name, 2, len(deployments))
	//		//}
	//		//if deployment == nil {
	//		//	t.Fatalf("Failed to find Deployment object for ApplicationEnvironment '%s/%s'", env.Namespace, env.Name)
	//		//}
	//		//if deployment.GetStatusConditionDeploying() == nil {
	//		//	t.Fatalf("Deployment object '%s/%s' is missing condition Deploying", deployment.Namespace, deployment.Name)
	//		//} else if deployment.GetStatusConditionDeploying().Status != v1.ConditionFalse {
	//		//	t.Fatalf("Deployment object '%s/%s' is still deploying: %+v", deployment.Namespace, deployment.Name, deployment.GetStatusConditionDeploying())
	//		//} else if deployment.Spec.Repository.APIVersion != apiv1.GroupVersion.String() {
	//		//	t.Fatalf("Incorrect API version for deployment repository: %+v", deployment.Spec.Repository)
	//		//} else if deployment.Spec.Repository.Kind != apiv1.GitHubRepositoryGVK.Kind {
	//		//	t.Fatalf("Incorrect Kind for deployment repository: %+v", deployment.Spec.Repository)
	//		//} else if deployment.Spec.Branch != expectedBranches[i] {
	//		//	t.Fatalf("Incorrect branch for deployment: expected '%s', was '%s'", expectedBranches[i], deployment.Spec.Branch)
	//		//}
	//	}
	//})
}
