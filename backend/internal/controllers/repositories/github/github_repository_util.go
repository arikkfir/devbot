package github

import (
	"context"
	apiv1 "github.com/arikkfir/devbot/backend/api/v1"
	"github.com/secureworks/errors"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func getPersonalAccessToken(ctx context.Context, c client.Client, repo *apiv1.GitHubRepository) (string, error) {
	if repo.Spec.Auth.PersonalAccessToken != nil {
		secretName := repo.Spec.Auth.PersonalAccessToken.Secret.Name
		if secretName == "" {
			return "", errors.New("auth secret name not configured")
		}

		secretNamespace := repo.Spec.Auth.PersonalAccessToken.Secret.Namespace
		if secretNamespace == "" {
			secretNamespace = repo.Namespace
		}

		key := repo.Spec.Auth.PersonalAccessToken.Key
		if key == "" {
			return "", errors.New("auth secret key not configured")
		}

		secret := &corev1.Secret{}
		if err := c.Get(ctx, client.ObjectKey{Namespace: secretNamespace, Name: secretName}, secret); err != nil {
			return "", errors.New("failed to get secret '%s/%s': %w", secretNamespace, secretName, err)
		}

		token := string(secret.Data[key])
		if token == "" {
			return "", errors.New("key '%s' in secret '%s/%s' is empty", key, secretNamespace, secretName)
		}

		return token, nil

	} else {
		return "", errors.New("auth not configured")
	}
}
