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
	"k8s.io/apimachinery/pkg/runtime"
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

func CreateKubernetesClient(_ context.Context, scheme *runtime.Scheme, Client *client.Client) {
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
	Expect(k8s.AddOwnershipIndex(mgrContext, mgr.GetFieldIndexer(), &apiv1.GitHubRepositoryRef{})).To(Succeed())
	Expect(k8s.AddOwnershipIndex(mgrContext, mgr.GetFieldIndexer(), &apiv1.Deployment{})).To(Succeed())
	Expect(k8s.AddOwnershipIndex(mgrContext, mgr.GetFieldIndexer(), &apiv1.Environment{})).To(Succeed())

	go func(ctx context.Context) {
		if err := mgr.Start(ctx); err != nil {
			log.Fatal().Err(err).Msg("Kubernetes manager failed")
		} else {
			log.Info().Msg("Kubernetes manager stopped")
		}
	}(mgrContext)

	*Client = mgr.GetClient()
	DeferCleanup(func() { *Client = nil })
}

func CreateKubernetesNamespace(ctx context.Context, k client.Client, namespaceName *string) {
	r := &corev1.Namespace{ObjectMeta: ctrl.ObjectMeta{Name: strings.RandomHash(7)}}
	Expect(k.Create(ctx, r)).To(Succeed())
	DeferCleanup(func(ctx context.Context) { Expect(k.Delete(ctx, r)).To(Succeed()) })
	*namespaceName = r.Name
}

func CreateGitHubSecretForDevbot(ctx context.Context, k client.Client, namespace, token string, secretName, key *string) {
	// Create a specific secret with the GitHub token
	*key = strings.RandomHash(7)
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Namespace: namespace, Name: strings.RandomHash(7)},
		Data:       map[string][]byte{*key: []byte(token)},
	}
	Expect(k.Create(ctx, secret)).To(Succeed())
	DeferCleanup(func(ctx context.Context) { Expect(k.Delete(ctx, secret)).To(Succeed()) })

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
	Expect(k.Create(ctx, clusterRole)).To(Succeed())
	DeferCleanup(func(ctx context.Context) { Expect(k.Delete(ctx, clusterRole)).To(Succeed()) })

	// Bind the cluster role to the devbot controllers, thus allowing them access to the specific secret
	roleBinding := &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{Namespace: secret.Namespace, Name: strings.RandomHash(7)},
		RoleRef:    rbacv1.RoleRef{APIGroup: rbacv1.SchemeGroupVersion.Group, Kind: k8sClusterRoleKind, Name: clusterRole.Name},
		Subjects: []rbacv1.Subject{
			{Kind: k8sServiceAccountKind, Name: DevbotGitHubRepositoryControllerServiceAccountName, Namespace: DevbotNamespace},
			{Kind: k8sServiceAccountKind, Name: DevbotGitHubRefControllerServiceAccountName, Namespace: DevbotNamespace},
		},
	}
	Expect(k.Create(ctx, roleBinding)).To(Succeed())
	DeferCleanup(func(ctx context.Context) { Expect(k.Delete(ctx, roleBinding)).To(Succeed()) })

	*secretName = secret.Name
}
