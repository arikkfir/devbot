package e2e

import (
	"context"
	apiv1 "github.com/arikkfir/devbot/backend/api/v1"
	"github.com/arikkfir/devbot/backend/internal/util"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"testing"
	"time"
)

func TestAppCreatesDistinctEnvsForBranches(t *testing.T) {
	ctx := context.Background()
	k := NewK8sTestClient(t)
	gh := NewGitHubTestClient(t, k.gitHubAuthSecrets["TOKEN"])

	ghRepo1, err := gh.CreateRepository(ctx)
	if err != nil {
		t.Fatalf("Failed to create GitHub repository: %+v", err)
	}
	ghRepoCR1, err := k.CreateGitHubRepository(ctx, "default", *ghRepo1.Owner.Login, *ghRepo1.Name)
	if err != nil {
		t.Fatalf("Failed to create GitHubRepository object: %+v", err)
	}
	ghRepoCR1NamespacedName := types.NamespacedName{Namespace: ghRepoCR1.Namespace, Name: ghRepoCR1.Name}

	ghRepo2, err := gh.CreateRepository(ctx)
	if err != nil {
		t.Fatalf("Failed to create GitHub repository: %+v", err)
	}
	ghRepoCR2, err := k.CreateGitHubRepository(ctx, "default", *ghRepo2.Owner.Login, *ghRepo2.Name)
	if err != nil {
		t.Fatalf("Failed to create GitHubRepository object: %+v", err)
	}
	ghRepoCR2NamespacedName := types.NamespacedName{Namespace: ghRepoCR2.Namespace, Name: ghRepoCR2.Name}

	if err := gh.CreateRepositoryWebhook(ctx, *ghRepo1.Name, k.gitHubAuthSecrets["WEBHOOK_SECRET"], "push"); err != nil {
		t.Fatalf("Failed to create GitHub repository webhook: %+v", err)
	} else if err := gh.CreateRepositoryWebhook(ctx, *ghRepo2.Name, k.gitHubAuthSecrets["WEBHOOK_SECRET"], "push"); err != nil {
		t.Fatalf("Failed to create GitHub repository webhook: %+v", err)
	}

	if _, err := gh.CreateFile(ctx, *ghRepo1.Name, "main", "README.md", t.Name()); err != nil {
		t.Fatalf("Failed to create GitHub file: %+v", err)
	} else if _, err := gh.CreateFile(ctx, *ghRepo1.Name, "feat1", "main.go", t.Name()); err != nil {
		t.Fatalf("Failed to create GitHub file: %+v", err)
	}

	if _, err := gh.CreateFile(ctx, *ghRepo2.Name, "main", "README.md", t.Name()); err != nil {
		t.Fatalf("Failed to create GitHub file: %+v", err)
	} else if _, err := gh.CreateFile(ctx, *ghRepo2.Name, "feat1", "main.go", t.Name()); err != nil {
		t.Fatalf("Failed to create GitHub file: %+v", err)
	} else if _, err := gh.CreateFile(ctx, *ghRepo2.Name, "feat2", "main.go", t.Name()); err != nil {
		t.Fatalf("Failed to create GitHub file: %+v", err)
	}

	if _, err := k.CreateApplication(ctx, "default", ghRepoCR1NamespacedName, ghRepoCR2NamespacedName); err != nil {
		t.Fatalf("Failed to create Application object: %+v", err)
	}

	util.Eventually(t, 5*time.Minute, 2*time.Second, func(t util.TestingT) {
		app1 := &apiv1.Application{}
		if err := k.c.Get(ctx, types.NamespacedName{Namespace: ghRepoCR1.Namespace, Name: ghRepoCR1.Name}, app1); err != nil {
			t.Fatalf("Failed getting app '%s/%s' object: %+v", ghRepoCR1.Namespace, ghRepoCR1.Name, err)
		} else if app1.GetStatusConditionValid() == nil {
			t.Fatalf("Application object '%s/%s' is not valid", app1.Namespace, app1.Name)
		} else if app1.GetStatusConditionValid().Status != v1.ConditionTrue {
			t.Fatalf("Application object '%s/%s' is not valid: %+v", app1.Namespace, app1.Name, app1.GetStatusConditionValid())
		}

		app2 := &apiv1.Application{}
		if err := k.c.Get(ctx, types.NamespacedName{Namespace: ghRepoCR2.Namespace, Name: ghRepoCR2.Name}, app2); err != nil {
			t.Fatalf("Failed getting app '%s/%s' object: %+v", ghRepoCR2.Namespace, ghRepoCR2.Name, err)
		} else if app2.GetStatusConditionValid() == nil {
			t.Fatalf("Application object '%s/%s' is not valid", app2.Namespace, app2.Name)
		} else if app2.GetStatusConditionValid().Status != v1.ConditionTrue {
			t.Fatalf("Application object '%s/%s' is not valid: %+v", app2.Namespace, app2.Name, app2.GetStatusConditionValid())
		}
	})

	util.Eventually(t, 5*time.Minute, 2*time.Second, func(t util.TestingT) {
		envs := &apiv1.ApplicationEnvironmentList{}
		if err := k.c.List(ctx, envs, client.InNamespace("default")); err != nil {
			t.Fatalf("Failed getting list of ApplicationEnvironment objects: %+v", err)
		}

		if len(envs.Items) != 3 {
			t.Fatalf("Incorrect number of ApplicationEnvironment objects, expected %d, got %d", 3, len(envs.Items))
		}
	})
}
