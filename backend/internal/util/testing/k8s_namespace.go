package testing

import (
	. "github.com/arikkfir/justest"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	apiv1 "github.com/arikkfir/devbot/backend/api/v1"
	"github.com/arikkfir/devbot/backend/internal/util/strings"
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
		ObjectMeta: metav1.ObjectMeta{
			Namespace:   n.Name,
			Name:        secretName,
			Annotations: map[string]string{"test": t.Name()},
		},
		Data: map[string][]byte{key: []byte(token)},
	}
	With(t).Verify(n.k.Client.Create(n.k.ctx, secret)).Will(Succeed()).OrFail()
	t.Cleanup(func() { With(t).Verify(n.k.Client.Delete(n.k.ctx, secret)).Will(Succeed()).OrFail() })

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
	With(t).Verify(n.k.Client.Create(n.k.ctx, clusterRole)).Will(Succeed()).OrFail()
	t.Cleanup(func() { With(t).Verify(n.k.Client.Delete(n.k.ctx, clusterRole)).Will(Succeed()).OrFail() })

	// Bind the cluster role to the devbot controllers, thus allowing them access to the specific secret
	roleBinding := &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{Namespace: secret.Namespace, Name: strings.RandomHash(7)},
		RoleRef:    rbacv1.RoleRef{APIGroup: rbacv1.SchemeGroupVersion.Group, Kind: k8sClusterRoleKind, Name: clusterRole.Name},
		Subjects: []rbacv1.Subject{
			{Kind: k8sServiceAccountKind, Name: DevbotRepositoryControllerServiceAccountName, Namespace: DevbotNamespace},
		},
	}
	With(t).Verify(n.k.Client.Create(n.k.ctx, roleBinding)).Will(Succeed()).OrFail()
	t.Cleanup(func() { With(t).Verify(n.k.Client.Delete(n.k.ctx, roleBinding)).Will(Succeed()).OrFail() })
	return
}

func (n *KNamespace) CreateRepository(t T, spec apiv1.RepositorySpec) string {
	repo := &apiv1.Repository{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:   n.Name,
			Name:        strings.RandomHash(7),
			Annotations: map[string]string{"test": t.Name()},
		},
		Spec: spec,
	}
	With(t).Verify(n.k.Client.Create(n.k.ctx, repo)).Will(Succeed()).OrFail()
	t.Cleanup(func() { With(t).Verify(n.k.Client.Delete(n.k.ctx, repo)).Will(Succeed()).OrFail() })

	return repo.Name
}

func (n *KNamespace) CreateApplication(t T, spec apiv1.ApplicationSpec) string {
	app := &apiv1.Application{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:   n.Name,
			Name:        strings.RandomHash(7),
			Annotations: map[string]string{"test": t.Name()},
		},
		Spec: spec,
	}
	With(t).Verify(n.k.Client.Create(n.k.ctx, app)).Will(Succeed()).OrFail()
	t.Cleanup(func() { With(t).Verify(n.k.Client.Delete(n.k.ctx, app)).Will(Succeed()).OrFail() })
	return app.Name
}
