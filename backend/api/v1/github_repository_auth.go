package v1

import (
	"context"
	"fmt"
	"github.com/google/go-github/v56/github"
	"github.com/secureworks/errors"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (o *GitHubRepository) GetGitHubClient(ctx context.Context, r client.Client) (gh *github.Client, status metav1.ConditionStatus, reason, message string, e error) {
	status = metav1.ConditionFalse

	if o.Spec.Auth.PersonalAccessToken != nil {
		// Authentication should be done using a PersonalAccessToken

		// Validate secret name is provided
		authSecretName := o.Spec.Auth.PersonalAccessToken.Secret.Name
		if authSecretName == "" {
			reason, message = ReasonGitHubAuthSecretNameMissing, "Auth secret name may not be empty"
			gh, e = nil, nil
			return
		}

		// Default the namespace to the object's namespace, if not specified inline
		authSecretNamespace := o.Spec.Auth.PersonalAccessToken.Secret.Namespace
		if authSecretNamespace == "" {
			authSecretNamespace = o.Namespace
		}

		// Get the secret
		authSecret := &corev1.Secret{}
		if err := r.Get(ctx, client.ObjectKey{Namespace: authSecretNamespace, Name: authSecretName}, authSecret); err != nil {
			if apierrors.IsNotFound(err) {
				reason, message = ReasonGitHubAuthSecretNotFound, fmt.Sprintf("Secret '%s/%s' not found", authSecretNamespace, authSecretName)
				gh, e = nil, errors.New("secret '%s/%s' could not be found", authSecretNamespace, authSecretName, err)
				return
			} else if apierrors.IsForbidden(err) {
				reason, message = ReasonGitHubAuthSecretForbidden, fmt.Sprintf("Forbidden from reading secret '%s/%s': %s", authSecretNamespace, authSecretName, err.Error())
				gh, e = nil, errors.New("secret '%s/%s' is forbidden", authSecretNamespace, authSecretName, err)
				return
			} else {
				reason, message = ReasonInternalError, fmt.Sprintf("Failed getting secret '%s/%s': %s", authSecretNamespace, authSecretName, err.Error())
				gh, e = nil, errors.New("failed to get secret '%s/%s'", authSecretNamespace, authSecretName, err)
				return
			}
		}

		// Get the token from the secret
		secretKey := o.Spec.Auth.PersonalAccessToken.Key
		pat, ok := authSecret.Data[secretKey]
		if string(pat) == "" || !ok {
			reason, message = ReasonGitHubAuthSecretEmptyToken, fmt.Sprintf("Key '%s' in secret '%s/%s' is missing or empty", secretKey, authSecretNamespace, authSecretName)
			gh, e = nil, errors.New("key '%s' in secret '%s/%s' is missing or empty", secretKey, authSecretNamespace, authSecretName)
			return
		}

		// Construct the GitHub client
		gh = github.NewClient(nil).WithAuthToken(string(pat))

	} else {
		reason, message = ReasonAuthConfigError, "Auth config is missing"
		gh, e = nil, nil
		return
	}

	// Validate the GitHub client is working
	if req, err := gh.NewRequest("GET", "user", nil); err != nil {
		reason, message = ReasonGitHubAPIFailed, fmt.Sprintf("GitHub connection failed: %s", err.Error())
		gh, e = nil, errors.New("failed to validate GitHub client", err)
		return
	} else if _, err := gh.Do(ctx, req, nil); err != nil {
		reason, message = ReasonGitHubAPIFailed, fmt.Sprintf("GitHub connection failed: %s", err.Error())
		gh, e = nil, errors.New("failed to validate GitHub client", err)
		return
	} else {
		status = metav1.ConditionTrue
		reason, message = ReasonAuthenticated, "Authenticated to GitHub"
		return
	}
}
