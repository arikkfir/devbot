package e2e

import (
	"context"
	apiv1 "github.com/arikkfir/devbot/backend/api/v1"
	"github.com/arikkfir/devbot/backend/internal/util"
	"github.com/secureworks/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/clientcmd"
	"os"
	"path/filepath"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"testing"
)

var (
	appGVK = schema.GroupVersionKind{
		Group:   apiv1.GroupVersion.Group,
		Version: apiv1.GroupVersion.Version,
		Kind:    "Application",
	}
	repoGVK = schema.GroupVersionKind{
		Group:   apiv1.GroupVersion.Group,
		Version: apiv1.GroupVersion.Version,
		Kind:    "GitHubRepository",
	}
)

type K8sTestClient struct {
	c                 client.Client
	t                 *testing.T
	gitHubAuthSecrets map[string]string
	cleanup           []func() error
}

func NewK8sTestClient(t *testing.T) *K8sTestClient {
	t.Helper()
	if err := apiv1.AddToScheme(scheme.Scheme); err != nil {
		t.Fatalf("Failed to register CRDs: %+v", err)
	}

	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("Failed to get user home dir: %+v", err)
	}
	kubeConfigPath := filepath.Join(userHomeDir, ".kube", "config")
	kubeConfig, err := clientcmd.BuildConfigFromFlags("", kubeConfigPath)
	if err != nil {
		t.Fatalf("Failed to read Kubernetes config: %+v", err)
	}

	k8sClient, err := client.New(kubeConfig, client.Options{Scheme: scheme.Scheme})
	if err != nil {
		t.Fatalf("Failed to create Kubernetes client: %+v", err)
	}

	secret := &corev1.Secret{}
	if err := k8sClient.Get(context.Background(), client.ObjectKey{Namespace: "devbot", Name: "devbot-github-auth"}, secret); err != nil {
		t.Fatalf("Failed to get devbot-github-auth secret: %+v", err)
	}

	gitHubAuthSecrets := make(map[string]string)
	for k, v := range secret.Data {
		gitHubAuthSecrets[k] = string(v)
	}
	kc := &K8sTestClient{c: k8sClient, t: t, gitHubAuthSecrets: gitHubAuthSecrets}
	t.Cleanup(kc.Close)
	return kc
}

func (k *K8sTestClient) Close() {
	k.t.Helper()
	for _, f := range k.cleanup {
		if err := f(); err != nil {
			k.t.Errorf("Kubernetes cleanup function failed: %+v", err)
		}
	}
}

func (k *K8sTestClient) Cleanup(f func() error) {
	k.t.Helper()
	k.cleanup = append([]func() error{f}, k.cleanup...)
}

func (k *K8sTestClient) CreateGitHubRepository(ctx context.Context, namespace, repoOwner, repoName string) (*apiv1.GitHubRepository, error) {
	k.t.Helper()

	crName := util.RandomHash(7)
	repoObj := apiv1.GitHubRepository{
		TypeMeta:   metav1.TypeMeta{APIVersion: appGVK.GroupVersion().String(), Kind: repoGVK.Kind},
		ObjectMeta: metav1.ObjectMeta{Namespace: namespace, Name: crName},
		Spec: apiv1.GitHubRepositorySpec{
			Owner: repoOwner,
			Name:  repoName,
			Auth: apiv1.GitHubRepositoryAuth{
				PersonalAccessToken: &apiv1.GitHubRepositoryAuthPersonalAccessToken{
					Secret: corev1.ObjectReference{
						Name:      "devbot-github-auth",
						Namespace: "devbot",
					},
					Key: "TOKEN",
				},
			},
		},
	}
	if err := k.c.Create(ctx, &repoObj); err != nil {
		return nil, errors.New("failed to create Application", err)
	}

	k.Cleanup(func() error {
		k.t.Helper()
		k.t.Logf("Deleting GitHubRepository object '%s/%s'...", namespace, crName)
		if err := k.c.Delete(ctx, &apiv1.GitHubRepository{
			TypeMeta:   metav1.TypeMeta{APIVersion: appGVK.GroupVersion().String(), Kind: repoGVK.Kind},
			ObjectMeta: metav1.ObjectMeta{Namespace: namespace, Name: crName},
		}); err != nil {
			return errors.New("failed to delete GitHubRepository object", errors.Meta("name", crName), err)
		}
		return nil
	})
	return &repoObj, nil
}

func (k *K8sTestClient) GetGitHubRepository(ctx context.Context, namespace, name string) (*apiv1.GitHubRepository, error) {
	k.t.Helper()
	r := apiv1.GitHubRepository{}
	if err := k.c.Get(ctx, client.ObjectKey{Namespace: namespace, Name: name}, &r); err != nil {
		return nil, errors.New("failed to get GitHubRepository", err)
	}
	return &r, nil
}

func (k *K8sTestClient) DeleteGitHubRepository(ctx context.Context, namespace, name string) error {
	k.t.Helper()
	if err := k.c.Delete(ctx, &apiv1.GitHubRepository{
		TypeMeta:   metav1.TypeMeta{APIVersion: repoGVK.GroupVersion().String(), Kind: repoGVK.Kind},
		ObjectMeta: metav1.ObjectMeta{Namespace: namespace, Name: name},
	}); err != nil {
		return errors.New("failed to delete GitHubRepository", errors.Meta("name", name), err)
	}
	return nil
}

func (k *K8sTestClient) CreateApplication(ctx context.Context, namespace string, repositories ...types.NamespacedName) (*apiv1.Application, error) {
	k.t.Helper()

	var repositoryReferences []corev1.ObjectReference
	for _, r := range repositories {
		repositoryReferences = append(repositoryReferences, corev1.ObjectReference{
			APIVersion: repoGVK.GroupVersion().String(),
			Kind:       repoGVK.Kind,
			Namespace:  r.Namespace,
			Name:       r.Name,
		})
	}

	crName := util.RandomHash(7)
	app := apiv1.Application{
		TypeMeta:   metav1.TypeMeta{APIVersion: appGVK.GroupVersion().String(), Kind: appGVK.Kind},
		ObjectMeta: metav1.ObjectMeta{Namespace: namespace, Name: crName},
		Spec: apiv1.ApplicationSpec{
			Repositories: repositoryReferences,
		},
	}
	if err := k.c.Create(ctx, &app); err != nil {
		return nil, errors.New("failed to create Application", err)
	}

	k.Cleanup(func() error {
		k.t.Helper()
		k.t.Logf("Deleting application '%s/%s'...", namespace, crName)
		if err := k.c.Delete(ctx, &apiv1.Application{
			TypeMeta:   metav1.TypeMeta{APIVersion: appGVK.GroupVersion().String(), Kind: appGVK.Kind},
			ObjectMeta: metav1.ObjectMeta{Namespace: namespace, Name: crName},
		}); err != nil {
			return errors.New("failed to delete application", errors.Meta("name", crName), err)
		}
		return nil
	})
	return &app, nil
}

func (k *K8sTestClient) GetApplication(ctx context.Context, namespace, name string) (*apiv1.Application, error) {
	k.t.Helper()
	a := apiv1.Application{}
	if err := k.c.Get(ctx, client.ObjectKey{Namespace: namespace, Name: name}, &a); err != nil {
		return nil, errors.New("failed to get Application", err)
	}
	return &a, nil
}

func (k *K8sTestClient) DeleteApplication(ctx context.Context, namespace, name string) error {
	k.t.Helper()
	if err := k.c.Delete(ctx, &apiv1.Application{
		TypeMeta:   metav1.TypeMeta{APIVersion: appGVK.GroupVersion().String(), Kind: appGVK.Kind},
		ObjectMeta: metav1.ObjectMeta{Namespace: namespace, Name: name},
	}); err != nil {
		return errors.New("failed to delete Application", errors.Meta("name", name), err)
	}
	return nil
}
