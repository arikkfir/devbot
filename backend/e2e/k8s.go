package e2e

import (
	"context"
	apiv1 "github.com/arikkfir/devbot/backend/api/v1"
	"github.com/arikkfir/devbot/backend/internal/util/k8s"
	"github.com/arikkfir/devbot/backend/internal/util/strings"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/rs/zerolog/log"
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
)

var (
	k8sServiceAccountKind = reflect.TypeOf(corev1.ServiceAccount{}).Name()
	k8sClusterRoleKind    = reflect.TypeOf(rbacv1.ClusterRole{}).Name()
)

type Kubernetes struct {
	Client client.Client
}

func NewKubernetes(_ context.Context) *Kubernetes {
	userHomeDir, err := os.UserHomeDir()
	Expect(err).NotTo(HaveOccurred())

	kubeConfigPath := filepath.Join(userHomeDir, ".kube", "config")
	kubeConfig, err := clientcmd.BuildConfigFromFlags("", kubeConfigPath)
	Expect(err).NotTo(HaveOccurred())

	mgrContext, cancel := context.WithCancel(context.Background())
	DeferCleanup(cancel)

	mgr, err := ctrl.NewManager(kubeConfig, ctrl.Options{
		Scheme:                 scheme,
		LeaderElection:         false,
		Metrics:                metricsserver.Options{BindAddress: "0"},
		HealthProbeBindAddress: "0",
		PprofBindAddress:       "0",
	})
	Expect(err).NotTo(HaveOccurred())
	Expect(k8s.AddOwnershipIndex(mgrContext, mgr.GetFieldIndexer(), &apiv1.GitHubRepositoryRef{})).Error().NotTo(HaveOccurred())
	Expect(k8s.AddOwnershipIndex(mgrContext, mgr.GetFieldIndexer(), &apiv1.Deployment{})).Error().NotTo(HaveOccurred())
	Expect(k8s.AddOwnershipIndex(mgrContext, mgr.GetFieldIndexer(), &apiv1.Environment{})).Error().NotTo(HaveOccurred())

	go func(ctx context.Context) {
		if err := mgr.Start(ctx); err != nil {
			log.Fatal().Err(err).Msg("Kubernetes manager failed")
		} else {
			log.Info().Msg("Kubernetes manager stopped")
		}
	}(mgrContext)

	return &Kubernetes{Client: mgr.GetClient()}
}

func (k *Kubernetes) CreateNamespace(ctx context.Context) *Namespace {
	r := &corev1.Namespace{ObjectMeta: ctrl.ObjectMeta{Name: strings.RandomHash(7)}}
	Expect(k.Client.Create(ctx, r)).Error().NotTo(HaveOccurred())
	DeferCleanup(func(ctx context.Context) { Expect(k.Client.Delete(ctx, r)).Error().NotTo(HaveOccurred()) })
	return &Namespace{Name: r.Name, k8s: k}
}

type Namespace struct {
	Name string
	k8s  *Kubernetes
}

func (n *Namespace) CreateGitHubAuthSecret(ctx context.Context, token string) (secretName, key string) {
	key = strings.RandomHash(7)
	secretName = strings.RandomHash(7)

	// Create a specific secret with the GitHub token
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Namespace: n.Name, Name: secretName},
		Data:       map[string][]byte{key: []byte(token)},
	}
	Expect(n.k8s.Client.Create(ctx, secret)).Error().NotTo(HaveOccurred())
	DeferCleanup(func(ctx context.Context) { Expect(n.k8s.Client.Delete(ctx, secret)).Error().NotTo(HaveOccurred()) })

	// Create the cluster role that grants access to our specific secret
	clusterRole := &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{Name: strings.RandomHash(7)},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups:     []string{corev1.GroupName},
				Resources:     []string{"secrets"},
				Verbs:         []string{"get", "list", "watch"},
				ResourceNames: []string{secret.Name},
			},
		},
	}
	Expect(n.k8s.Client.Create(ctx, clusterRole)).Error().NotTo(HaveOccurred())
	DeferCleanup(func(ctx context.Context) { Expect(n.k8s.Client.Delete(ctx, clusterRole)).Error().NotTo(HaveOccurred()) })

	// Bind the cluster role to the devbot controllers, thus allowing them access to the specific secret
	roleBinding := &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{Namespace: secret.Namespace, Name: strings.RandomHash(7)},
		RoleRef:    rbacv1.RoleRef{APIGroup: rbacv1.SchemeGroupVersion.Group, Kind: k8sClusterRoleKind, Name: clusterRole.Name},
		Subjects: []rbacv1.Subject{
			{Kind: k8sServiceAccountKind, Name: DevbotGitHubRepositoryControllerServiceAccountName, Namespace: DevbotNamespace},
			{Kind: k8sServiceAccountKind, Name: DevbotGitHubRefControllerServiceAccountName, Namespace: DevbotNamespace},
		},
	}
	Expect(n.k8s.Client.Create(ctx, roleBinding)).Error().NotTo(HaveOccurred())
	DeferCleanup(func(ctx context.Context) { Expect(n.k8s.Client.Delete(ctx, roleBinding)).Error().NotTo(HaveOccurred()) })
	return
}
