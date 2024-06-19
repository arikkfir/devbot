package util

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	ctrl "sigs.k8s.io/controller-runtime"

	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/arikkfir/devbot/internal/util/strings"
)

func CreateK8sGitOpsServiceAccount(ctx context.Context, c client.Client, ns string) string {
	GinkgoHelper()
	sa := &corev1.ServiceAccount{ObjectMeta: ctrl.ObjectMeta{Name: strings.RandomHash(7), Namespace: ns}}
	Expect(c.Create(ctx, sa)).To(Succeed())
	DeferCleanup(func(ctx context.Context) { Expect(c.Delete(ctx, sa)).To(Succeed()) })

	role := &rbacv1.Role{
		ObjectMeta: ctrl.ObjectMeta{Name: sa.Name, Namespace: sa.Namespace},
		Rules:      []rbacv1.PolicyRule{{APIGroups: []string{"*"}, Resources: []string{"*"}, Verbs: []string{"*"}}},
	}
	Expect(c.Create(ctx, role)).To(Succeed())
	DeferCleanup(func(ctx context.Context) { Expect(c.Delete(ctx, role)).To(Succeed()) })

	rb := &rbacv1.RoleBinding{
		ObjectMeta: ctrl.ObjectMeta{Name: sa.Name, Namespace: sa.Namespace},
		RoleRef:    rbacv1.RoleRef{APIGroup: rbacv1.GroupName, Kind: "Role", Name: role.Name},
		Subjects:   []rbacv1.Subject{{Kind: rbacv1.ServiceAccountKind, Name: sa.Name}},
	}
	Expect(c.Create(ctx, rb)).To(Succeed())
	DeferCleanup(func(ctx context.Context) { Expect(c.Delete(ctx, rb)).To(Succeed()) })

	return sa.Name
}

func CreateK8sSecretWithGitHubAuthToken(ctx context.Context, c client.Client, ns, token string) (string, string) {
	GinkgoHelper()
	secretName := strings.RandomHash(7)
	key := strings.RandomHash(7)

	// Create a specific secret with the GitHub token
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Namespace: ns, Name: secretName},
		Data:       map[string][]byte{key: []byte(token)},
	}
	Expect(c.Create(ctx, secret)).To(Succeed())
	DeferCleanup(func(ctx context.Context) { Expect(c.Delete(ctx, secret)).To(Succeed()) })
	return secretName, key
}

func GrantK8sAccessToSecret(ctx context.Context, c client.Client, ns, secretName string) {
	GinkgoHelper()
	clusterRole := &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{"devbot.kfirs.com/purpose": "test", "devbot.kfirs.com/target": ns},
			Name:   strings.RandomHash(7),
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups:     []string{corev1.GroupName},
				Resources:     []string{"secrets"},
				Verbs:         []string{"get", "list", "watch"},
				ResourceNames: []string{secretName},
			},
		},
	}
	Expect(c.Create(ctx, clusterRole)).To(Succeed())
	DeferCleanup(func(ctx context.Context) { Expect(c.Delete(ctx, clusterRole)).To(Succeed()) })

	roleBinding := &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{Namespace: ns, Name: strings.RandomHash(7)},
		RoleRef:    rbacv1.RoleRef{APIGroup: rbacv1.SchemeGroupVersion.Group, Kind: k8sClusterRoleKind, Name: clusterRole.Name},
		Subjects:   []rbacv1.Subject{{Kind: k8sServiceAccountKind, Name: DevbotControllerServiceAccountName, Namespace: DevbotNamespace}},
	}
	Expect(c.Create(ctx, roleBinding)).To(Succeed())
	DeferCleanup(func(ctx context.Context) { Expect(c.Delete(ctx, roleBinding)).To(Succeed()) })
}

func GrantK8sAccessToAllSecrets(ctx context.Context, c client.Client, ns string) {
	GinkgoHelper()
	clusterRole := &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{"devbot.kfirs.com/purpose": "test", "devbot.kfirs.com/target": ns},
			Name:   strings.RandomHash(7),
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{corev1.GroupName},
				Resources: []string{"secrets"},
				Verbs:     []string{"get", "list", "watch"},
			},
		},
	}
	Expect(c.Create(ctx, clusterRole)).To(Succeed())
	DeferCleanup(func(ctx context.Context) { Expect(c.Delete(ctx, clusterRole)).To(Succeed()) })

	roleBinding := &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{Namespace: ns, Name: strings.RandomHash(7)},
		RoleRef:    rbacv1.RoleRef{APIGroup: rbacv1.SchemeGroupVersion.Group, Kind: k8sClusterRoleKind, Name: clusterRole.Name},
		Subjects:   []rbacv1.Subject{{Kind: k8sServiceAccountKind, Name: DevbotControllerServiceAccountName, Namespace: DevbotNamespace}},
	}
	Expect(c.Create(ctx, roleBinding)).To(Succeed())
	DeferCleanup(func(ctx context.Context) { Expect(c.Delete(ctx, roleBinding)).To(Succeed()) })
}
