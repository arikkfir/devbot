package testing

import (
	apiv1 "github.com/arikkfir/devbot/backend/api/v1"
	"github.com/arikkfir/devbot/backend/internal/util/strings"
	. "github.com/arikkfir/devbot/backend/internal/util/testing/justest"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type KNamespace struct {
	Name string
	k    *KClient
}

func (n *KNamespace) CreateGitHubAuthSecretSpec(t T, token string, restrictRole bool) *apiv1.GitHubRepositoryPersonalAccessToken {
	ghAuthSecretName, ghAuthSecretKeyName := n.CreateGitHubAuthSecret(t, token, restrictRole)
	return &apiv1.GitHubRepositoryPersonalAccessToken{
		Secret: apiv1.SecretReferenceWithOptionalNamespace{
			Name:      ghAuthSecretName,
			Namespace: n.Name,
		},
		Key: ghAuthSecretKeyName,
	}
}

func (n *KNamespace) CreateGitHubAuthSecret(t T, token string, restrictRole bool) (secretName, key string) {
	key = strings.RandomHash(7)
	secretName = strings.RandomHash(7)

	// Create a specific secret with the GitHub token
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Namespace: n.Name, Name: secretName},
		Data:       map[string][]byte{key: []byte(token)},
	}
	For(t).Expect(n.k.Client.Create(For(t).Context(), secret)).Will(Succeed())
	t.Cleanup(func() { For(t).Expect(n.k.Client.Delete(For(t).Context(), secret)).Will(Succeed()) })

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
	For(t).Expect(n.k.Client.Create(For(t).Context(), clusterRole)).Will(Succeed())
	t.Cleanup(func() { For(t).Expect(n.k.Client.Delete(For(t).Context(), clusterRole)).Will(Succeed()) })

	// Bind the cluster role to the devbot controllers, thus allowing them access to the specific secret
	roleBinding := &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{Namespace: secret.Namespace, Name: strings.RandomHash(7)},
		RoleRef:    rbacv1.RoleRef{APIGroup: rbacv1.SchemeGroupVersion.Group, Kind: k8sClusterRoleKind, Name: clusterRole.Name},
		Subjects: []rbacv1.Subject{
			{Kind: k8sServiceAccountKind, Name: DevbotRepositoryControllerServiceAccountName, Namespace: DevbotNamespace},
		},
	}
	For(t).Expect(n.k.Client.Create(For(t).Context(), roleBinding)).Will(Succeed())
	t.Cleanup(func() { For(t).Expect(n.k.Client.Delete(For(t).Context(), roleBinding)).Will(Succeed()) })
	return
}

func (n *KNamespace) CreateRepository(t T, spec apiv1.RepositorySpec) string {
	repo := &apiv1.Repository{ObjectMeta: metav1.ObjectMeta{Namespace: n.Name, Name: strings.RandomHash(7)}, Spec: spec}
	For(t).Expect(n.k.Client.Create(For(t).Context(), repo)).Will(Succeed())
	t.Cleanup(func() { For(t).Expect(n.k.Client.Delete(For(t).Context(), repo)).Will(Succeed()) })

	return repo.Name
}

func (n *KNamespace) CreateApplication(t T, spec apiv1.ApplicationSpec) string {
	app := &apiv1.Application{ObjectMeta: metav1.ObjectMeta{Namespace: n.Name, Name: strings.RandomHash(7)}, Spec: spec}
	For(t).Expect(n.k.Client.Create(For(t).Context(), app)).Will(Succeed())
	t.Cleanup(func() { For(t).Expect(n.k.Client.Delete(For(t).Context(), app)).Will(Succeed()) })
	return app.Name
}