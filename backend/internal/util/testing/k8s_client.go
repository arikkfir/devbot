package testing

import (
	apiv1 "github.com/arikkfir/devbot/backend/api/v1"
	"github.com/arikkfir/devbot/backend/internal/util/k8s"
	"github.com/arikkfir/devbot/backend/internal/util/strings"
	. "github.com/arikkfir/devbot/backend/internal/util/testing/justest"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
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
	kClientKey                                   = "___Kubernetes"
)

var (
	k8sServiceAccountKind = reflect.TypeOf(corev1.ServiceAccount{}).Name()
	k8sClusterRoleKind    = reflect.TypeOf(rbacv1.ClusterRole{}).Name()
)

type KClient struct {
	Client client.Client
}

func K(t T) *KClient {
	if v := For(t).Value(kClientKey); v == nil {
		userHomeDir, err := os.UserHomeDir()
		For(t).Expect(err).Will(BeNil()).OrFail()

		kubeConfigPath := filepath.Join(userHomeDir, ".kube", "config")
		kubeConfig, err := clientcmd.BuildConfigFromFlags("", kubeConfigPath)
		For(t).Expect(err).Will(BeNil()).OrFail()

		scheme := k8s.CreateScheme()
		mgr, err := ctrl.NewManager(kubeConfig, ctrl.Options{
			Scheme: scheme,
			Client: client.Options{
				Scheme: scheme,
				Cache: &client.CacheOptions{
					DisableFor: nil,
				},
			},
			LeaderElection:         false,
			Metrics:                metricsserver.Options{BindAddress: "0"},
			HealthProbeBindAddress: "0",
			PprofBindAddress:       "0",
		})
		For(t).Expect(err).Will(BeNil()).OrFail()
		For(t).Expect(k8s.AddOwnershipIndex(For(t).Context(), mgr.GetFieldIndexer(), &apiv1.Environment{})).Will(Succeed()).OrFail()
		For(t).Expect(k8s.AddOwnershipIndex(For(t).Context(), mgr.GetFieldIndexer(), &apiv1.Deployment{})).Will(Succeed()).OrFail()

		go func() {
			if err := mgr.Start(For(t).Context()); err != nil {
				t.Logf("Kubernetes manager failed: %+v", err)
			}
		}()
		For(t).AddValue(kClientKey, &KClient{Client: mgr.GetClient()})

		time.Sleep(3 * time.Second)
		return K(t)
	} else {
		return v.(*KClient)
	}
}

func (k *KClient) CreateNamespace(t T) *KNamespace {
	devbotGitOpsName := "devbot-gitops"

	r := &corev1.Namespace{ObjectMeta: ctrl.ObjectMeta{Name: strings.RandomHash(7)}}
	For(t).Expect(k.Client.Create(For(t).Context(), r)).Will(Succeed()).OrFail()
	t.Cleanup(func() { For(t).Expect(k.Client.Delete(For(t).Context(), r)).Will(Succeed()).OrFail() })

	sa := &corev1.ServiceAccount{ObjectMeta: ctrl.ObjectMeta{Name: devbotGitOpsName, Namespace: r.Name}}
	For(t).Expect(k.Client.Create(For(t).Context(), sa)).Will(Succeed()).OrFail()
	t.Cleanup(func() { For(t).Expect(k.Client.Delete(For(t).Context(), sa)).Will(Succeed()).OrFail() })

	role := &rbacv1.Role{
		ObjectMeta: ctrl.ObjectMeta{Name: devbotGitOpsName, Namespace: r.Name},
		Rules:      []rbacv1.PolicyRule{{APIGroups: []string{"*"}, Resources: []string{"*"}, Verbs: []string{"*"}}},
	}
	For(t).Expect(k.Client.Create(For(t).Context(), role)).Will(Succeed()).OrFail()
	t.Cleanup(func() { For(t).Expect(k.Client.Delete(For(t).Context(), role)).Will(Succeed()).OrFail() })

	rb := &rbacv1.RoleBinding{
		ObjectMeta: ctrl.ObjectMeta{Name: devbotGitOpsName, Namespace: r.Name},
		RoleRef:    rbacv1.RoleRef{APIGroup: rbacv1.GroupName, Kind: "Role", Name: devbotGitOpsName},
		Subjects:   []rbacv1.Subject{{Kind: rbacv1.ServiceAccountKind, Name: devbotGitOpsName}},
	}
	For(t).Expect(k.Client.Create(For(t).Context(), rb)).Will(Succeed()).OrFail()
	t.Cleanup(func() { For(t).Expect(k.Client.Delete(For(t).Context(), rb)).Will(Succeed()).OrFail() })

	return &KNamespace{Name: r.Name, k: k}
}
