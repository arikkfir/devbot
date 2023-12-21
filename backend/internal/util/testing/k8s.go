package testing

import (
	"context"
	apiv1 "github.com/arikkfir/devbot/backend/api/v1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/clientcmd"
	"os"
	"path/filepath"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func CreateKubernetesClient(Client *client.Client) {
	scheme := runtime.NewScheme()
	utilruntime.Must(apiv1.AddToScheme(scheme))
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	userHomeDir, err := os.UserHomeDir()
	Expect(err).ToNot(HaveOccurred())

	kubeConfigPath := filepath.Join(userHomeDir, ".kube", "config")
	kubeConfig, err := clientcmd.BuildConfigFromFlags("", kubeConfigPath)
	Expect(err).ToNot(HaveOccurred())

	c, err := client.New(kubeConfig, client.Options{Scheme: scheme})
	Expect(err).ToNot(HaveOccurred())
	*Client = c

	DeferCleanup(func() { *Client = nil })
}

func CreateGitHubRepositoryCR(ctx context.Context, k8s client.Client, gitHubRepositoryName string, crName, crNamespace *string, opts ...ApplyOption[*apiv1.GitHubRepository]) {
	*crNamespace = "default"
	*crName = GitHubOwner + "-" + gitHubRepositoryName
	cr := &apiv1.GitHubRepository{
		ObjectMeta: metav1.ObjectMeta{Namespace: *crNamespace, Name: *crName},
		Spec: apiv1.GitHubRepositorySpec{
			Owner: GitHubOwner,
			Name:  gitHubRepositoryName,
			Auth: apiv1.GitHubRepositoryAuth{
				PersonalAccessToken: &apiv1.GitHubRepositoryAuthPersonalAccessToken{
					Secret: corev1.SecretReference{
						Name:      GitHubSecretName,
						Namespace: GitHubSecretNamespace,
					},
					Key: GitHubSecretPATKey,
				},
			},
		},
	}
	for _, opt := range opts {
		opt(cr)
	}
	Expect(k8s.Create(ctx, cr)).To(Succeed())
	DeferCleanup(func() {
		var namespace, name = *crNamespace, *crName
		*crName = ""
		*crNamespace = ""
		Expect(k8s.Delete(context.Background(), &apiv1.GitHubRepository{
			ObjectMeta: metav1.ObjectMeta{Namespace: namespace, Name: name},
		})).To(Succeed())
	})
}

func ObtainGitHubWebhookSecret(ctx context.Context, c client.Client) string {
	gitHubSecretKey := client.ObjectKey{Namespace: GitHubSecretNamespace, Name: GitHubSecretName}
	secret := &corev1.Secret{}
	Expect(c.Get(ctx, gitHubSecretKey, secret)).To(Succeed())
	return string(secret.Data["WEBHOOK_SECRET"])
}

func ObtainGitHubPAT(ctx context.Context, c client.Client) string {
	gitHubSecretKey := client.ObjectKey{Namespace: GitHubSecretNamespace, Name: GitHubSecretName}
	secret := &corev1.Secret{}
	Expect(c.Get(ctx, gitHubSecretKey, secret)).To(Succeed())
	return string(secret.Data["TOKEN"])
}
