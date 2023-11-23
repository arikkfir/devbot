package github

import (
	"context"
	"encoding/base64"
	apiv1 "github.com/arikkfir/devbot/backend/api/v1"
	"github.com/secureworks/errors"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"
)

func getGitHubRepositoryRefLabels(owner, name, ref string) client.MatchingLabels {
	// TODO: use other means than labels to lookup GitHubRepositoryRef objects for a given repo & ref
	repoOwnerAsBase64 := strings.Trim(base64.StdEncoding.EncodeToString([]byte(owner)), "-=")
	repoNameAsBase64 := strings.Trim(base64.StdEncoding.EncodeToString([]byte(name)), "-=")
	refNameAsBase64 := strings.Trim(base64.StdEncoding.EncodeToString([]byte(ref)), "-=")
	return client.MatchingLabels{
		"github.devbot.kfirs.com/repository-owner": repoOwnerAsBase64,
		"github.devbot.kfirs.com/repository-name":  repoNameAsBase64,
		"github.devbot.kfirs.com/ref":              refNameAsBase64,
	}
}

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
