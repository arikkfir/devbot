package e2e

import (
	"context"
	apiv1 "github.com/arikkfir/devbot/backend/api/v1"
	"github.com/secureworks/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"os"
	"path/filepath"
	"testing"
)

type K8sTestClient struct {
	t                *testing.T
	kubeConfig       *rest.Config
	k8sClient        *kubernetes.Clientset
	k8sDynamicClient *dynamic.DynamicClient
	namespace        string
	cleanup          []func() error
}

func NewK8sTestClient(t *testing.T, namespace string) (*K8sTestClient, error) {
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, errors.New("failed to get user home dir", err)
	}
	kubeConfigPath := filepath.Join(userHomeDir, ".kube", "config")
	kubeConfig, err := clientcmd.BuildConfigFromFlags("", kubeConfigPath)
	if err != nil {
		return nil, errors.New("failed to read Kubernetes config", errors.Meta("path", kubeConfigPath), err)
	}
	k8sClient, err := kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		return nil, errors.New("failed to create Kubernetes static client", err)
	}
	k8sDynamicClient, err := dynamic.NewForConfig(kubeConfig)
	if err != nil {
		return nil, errors.New("failed to create Kubernetes dynamic client", err)
	}

	if err := apiv1.AddToScheme(scheme.Scheme); err != nil {
		return nil, errors.New("failed to register CRDs", err)
	}

	return &K8sTestClient{
		t:                t,
		kubeConfig:       kubeConfig,
		k8sClient:        k8sClient,
		k8sDynamicClient: k8sDynamicClient,
		namespace:        namespace,
	}, nil
}

func (k *K8sTestClient) Close() {
	for _, f := range k.cleanup {
		if err := f(); err != nil {
			k.t.Errorf("Kubernetes cleanup function failed: %+v", err)
		}
	}
}

func (k *K8sTestClient) Cleanup(f func() error) {
	k.cleanup = append([]func() error{f}, k.cleanup...)
}

func (k *K8sTestClient) CreateApplication(ctx context.Context, owner, repo string) (*apiv1.Application, error) {
	gvk := schema.GroupVersionKind{
		Group:   apiv1.GroupVersion.Group,
		Version: apiv1.GroupVersion.Version,
		Kind:    "Application",
	}
	name := "devbot-test-" + owner + "-" + repo
	app := apiv1.Application{}
	err := k.k8sClient.RESTClient().
		Post().
		Namespace(k.namespace).
		Resource("Application").
		Body(&apiv1.Application{
			TypeMeta: metav1.TypeMeta{
				APIVersion: gvk.GroupVersion().String(),
				Kind:       gvk.Kind,
			},
			ObjectMeta: metav1.ObjectMeta{Name: name},
			Spec: apiv1.ApplicationSpec{
				Repository: apiv1.ApplicationRepository{
					GitHub: &apiv1.ApplicationGitHubRepository{
						Owner: owner,
						Name:  repo,
					},
				},
			},
		}).
		Do(ctx).
		Into(&app)
	if err != nil {
		return nil, errors.New("failed to create Application", err)
	}

	k.Cleanup(func() error {
		k.t.Helper()
		k.t.Logf("Deleting application '%s/%s'...", k.namespace, name)
		err := k.k8sClient.RESTClient().
			Delete().
			Namespace("default").
			Resource("Application").
			Name(name).
			Body(&metav1.DeleteOptions{}).
			Do(ctx).
			Error()
		if err != nil {
			return errors.New("failed to delete application", errors.Meta("app", name), err)
		}
		return nil
	})
	return &app, nil
}

func (k *K8sTestClient) GetApplication(ctx context.Context, name string) (*apiv1.Application, error) {
	app := apiv1.Application{}
	err := k.k8sClient.RESTClient().
		Get().
		Namespace(k.namespace).
		Resource("Application").
		Name(name).
		Body(&metav1.GetOptions{
			TypeMeta: metav1.TypeMeta{
				APIVersion: apiv1.GroupVersion.String(),
				Kind:       "Application",
			},
		}).
		Do(ctx).
		Into(&app)
	if err != nil {
		return nil, errors.New("failed to get Application", err)
	}
	return &app, nil
}
