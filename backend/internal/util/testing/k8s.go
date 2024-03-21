package testing

import (
	"context"
	apiv1 "github.com/arikkfir/devbot/backend/api/v1"
	"github.com/arikkfir/devbot/backend/internal/util/k8s"
	"github.com/arikkfir/devbot/backend/internal/util/strings"
	. "github.com/arikkfir/devbot/backend/internal/util/testing/justest"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd"
	"os"
	"path/filepath"
	"reflect"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
	"time"
)

const (
	DevbotNamespace                              = "devbot"
	DevbotRepositoryControllerServiceAccountName = "devbot-repository-controller"
)

var (
	k8sServiceAccountKind = reflect.TypeOf(corev1.ServiceAccount{}).Name()
	k8sClusterRoleKind    = reflect.TypeOf(rbacv1.ClusterRole{}).Name()
)

type KClient struct {
	Client client.Client
}

type KNamespace struct {
	Name string
	k    *KClient
}

func NewKubernetes(t JustT) *KClient {
	ctx, cancel := context.WithCancel(context.Background())

	userHomeDir, err := os.UserHomeDir()
	For(t).Expect(err).Will(BeNil())

	kubeConfigPath := filepath.Join(userHomeDir, ".kube", "config")
	kubeConfig, err := clientcmd.BuildConfigFromFlags("", kubeConfigPath)
	For(t).Expect(err).Will(BeNil())

	scheme := k8s.CreateScheme()
	mgr, err := ctrl.NewManager(kubeConfig, ctrl.Options{
		Scheme: scheme,
		Client: client.Options{
			Cache: &client.CacheOptions{
				DisableFor: []client.Object{
					//&apiv1.Repository{},
					//&apiv1.Application{},
					//&apiv1.Environment{},
					//&apiv1.Deployment{},
					//&corev1.ConfigMap{},
					//&corev1.Namespace{},
					//&corev1.Secret{},
				},
			},
		},
		LeaderElection:         false,
		Metrics:                metricsserver.Options{BindAddress: "0"},
		HealthProbeBindAddress: "0",
		PprofBindAddress:       "0",
	})
	For(t).Expect(err).Will(BeNil())
	For(t).Expect(k8s.AddOwnershipIndex(ctx, mgr.GetFieldIndexer(), &apiv1.Deployment{})).Will(Succeed())
	For(t).Expect(k8s.AddOwnershipIndex(ctx, mgr.GetFieldIndexer(), &apiv1.Environment{})).Will(Succeed())

	go func() {
		if err := mgr.Start(ctx); err != nil {
			t.Logf("Kubernetes manager failed: %+v", err)
		}
	}()
	t.Cleanup(func() {
		cancel()
		time.Sleep(2 * time.Second)
	})

	return &KClient{Client: mgr.GetClient()}
}

func (k *KClient) CreateNamespace(t JustT, ctx context.Context) *KNamespace {
	r := &corev1.Namespace{ObjectMeta: ctrl.ObjectMeta{Name: strings.RandomHash(7)}}
	For(t).Expect(k.Client.Create(ctx, r)).Will(Succeed())
	t.Cleanup(func() { For(t).Expect(k.Client.Delete(ctx, r)).Will(Succeed()) })
	return &KNamespace{Name: r.Name, k: k}
}

func (n *KNamespace) CreateGitHubAuthSecretSpec(t JustT, ctx context.Context, token string, restrictRole bool) *apiv1.GitHubRepositoryPersonalAccessToken {
	ghAuthSecretName, ghAuthSecretKeyName := n.CreateGitHubAuthSecret(t, ctx, token, restrictRole)
	return &apiv1.GitHubRepositoryPersonalAccessToken{
		Secret: apiv1.SecretReferenceWithOptionalNamespace{
			Name:      ghAuthSecretName,
			Namespace: n.Name,
		},
		Key: ghAuthSecretKeyName,
	}
}

func (n *KNamespace) CreateGitHubAuthSecret(t JustT, ctx context.Context, token string, restrictRole bool) (secretName, key string) {
	key = strings.RandomHash(7)
	secretName = strings.RandomHash(7)

	// Create a specific secret with the GitHub token
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Namespace: n.Name, Name: secretName},
		Data:       map[string][]byte{key: []byte(token)},
	}
	For(t).Expect(n.k.Client.Create(ctx, secret)).Will(Succeed())
	t.Cleanup(func() { For(t).Expect(n.k.Client.Delete(ctx, secret)).Will(Succeed()) })

	// List of resource names to restrict the role to (if any)
	var resourceNames []string
	if restrictRole {
		resourceNames = []string{secretName}
	}

	// Create the cluster role that grants access to our specific secret
	clusterRole := &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{Name: strings.RandomHash(7)},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups:     []string{corev1.GroupName},
				Resources:     []string{"secrets"},
				Verbs:         []string{"get", "list", "watch"},
				ResourceNames: resourceNames,
			},
		},
	}
	For(t).Expect(n.k.Client.Create(ctx, clusterRole)).Will(Succeed())
	t.Cleanup(func() { For(t).Expect(n.k.Client.Delete(ctx, clusterRole)).Will(Succeed()) })

	// Bind the cluster role to the devbot controllers, thus allowing them access to the specific secret
	roleBinding := &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{Namespace: secret.Namespace, Name: strings.RandomHash(7)},
		RoleRef:    rbacv1.RoleRef{APIGroup: rbacv1.SchemeGroupVersion.Group, Kind: k8sClusterRoleKind, Name: clusterRole.Name},
		Subjects: []rbacv1.Subject{
			{Kind: k8sServiceAccountKind, Name: DevbotRepositoryControllerServiceAccountName, Namespace: DevbotNamespace},
		},
	}
	For(t).Expect(n.k.Client.Create(ctx, roleBinding)).Will(Succeed())
	t.Cleanup(func() { For(t).Expect(n.k.Client.Delete(ctx, roleBinding)).Will(Succeed()) })
	return
}

func (n *KNamespace) CreateRepository(t JustT, ctx context.Context, spec apiv1.RepositorySpec) string {
	repo := &apiv1.Repository{ObjectMeta: metav1.ObjectMeta{Namespace: n.Name, Name: strings.RandomHash(7)}, Spec: spec}
	For(t).Expect(n.k.Client.Create(ctx, repo)).Will(Succeed())
	t.Cleanup(func() { For(t).Expect(n.k.Client.Delete(ctx, repo)).Will(Succeed()) })

	return repo.Name
}

func (n *KNamespace) CreateApplication(t JustT, ctx context.Context, spec apiv1.ApplicationSpec) string {
	app := &apiv1.Application{ObjectMeta: metav1.ObjectMeta{Namespace: n.Name, Name: strings.RandomHash(7)}, Spec: spec}
	For(t).Expect(n.k.Client.Create(ctx, app)).Will(Succeed())
	t.Cleanup(func() { For(t).Expect(n.k.Client.Delete(ctx, app)).Will(Succeed()) })
	return app.Name
}
