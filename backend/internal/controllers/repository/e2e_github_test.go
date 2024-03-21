package repository_test

import (
	"context"
	apiv1 "github.com/arikkfir/devbot/backend/api/v1"
	. "github.com/arikkfir/devbot/backend/internal/util/testing"
	. "github.com/arikkfir/devbot/backend/internal/util/testing/justest"
	. "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"testing"
	"time"
)

func TestRepositoryGitHubConnection(t *testing.T) {
	testCases := map[string]struct {
		owner, name        string
		repoContents       string
		patProvider        func(namespace, secretName, secretKey string) *apiv1.GitHubRepositoryPersonalAccessToken
		restrictSecretRole bool
		invalid            *Condition
		unauthenticated    *Condition
		stale              *Condition
	}{
		"OwnerMissing": {
			name:            "myRepo",
			invalid:         &Condition{Status: ConditionTrue, Reason: apiv1.RepositoryOwnerMissing, Message: "Repository owner is empty"},
			unauthenticated: &Condition{Status: ConditionTrue, Reason: apiv1.Invalid, Message: "Repository owner is empty"},
			stale:           &Condition{Status: ConditionUnknown, Reason: apiv1.Unauthenticated, Message: "Repository owner is empty"},
		},
		"NameMissing": {
			owner:           "myOwner",
			invalid:         &Condition{Status: ConditionTrue, Reason: apiv1.RepositoryNameMissing, Message: "Repository name is empty"},
			unauthenticated: &Condition{Status: ConditionTrue, Reason: apiv1.Invalid, Message: "Repository name is empty"},
			stale:           &Condition{Status: ConditionUnknown, Reason: apiv1.Unauthenticated, Message: "Repository name is empty"},
		},
		"NoAuthProvided": {
			owner: "someRepoOwner",
			name:  "someRepoName",
			patProvider: func(namespace, secretName, secretKey string) *apiv1.GitHubRepositoryPersonalAccessToken {
				return nil
			},
			invalid:         &Condition{Status: ConditionTrue, Reason: apiv1.AuthConfigMissing, Message: "Auth config is missing"},
			unauthenticated: &Condition{Status: ConditionTrue, Reason: apiv1.Invalid, Message: "Auth config is missing"},
			stale:           &Condition{Status: ConditionUnknown, Reason: apiv1.Unauthenticated, Message: "Auth config is missing"},
		},
		"AuthSecretNameMissing": {
			repoContents: "bare",
			patProvider: func(namespace, secretName, secretKey string) *apiv1.GitHubRepositoryPersonalAccessToken {
				return &apiv1.GitHubRepositoryPersonalAccessToken{
					Secret: apiv1.SecretReferenceWithOptionalNamespace{
						Name:      "",
						Namespace: namespace,
					},
					Key: secretKey,
				}
			},
			invalid:         &Condition{Status: ConditionTrue, Reason: apiv1.AuthSecretNameMissing, Message: "Auth secret name is empty"},
			unauthenticated: &Condition{Status: ConditionTrue, Reason: apiv1.Invalid, Message: "Auth secret name is empty"},
			stale:           &Condition{Status: ConditionUnknown, Reason: apiv1.Unauthenticated, Message: "Auth secret name is empty"},
		},
		"AuthSecretKeyMissing": {
			repoContents: "bare",
			patProvider: func(namespace, secretName, secretKey string) *apiv1.GitHubRepositoryPersonalAccessToken {
				return &apiv1.GitHubRepositoryPersonalAccessToken{
					Secret: apiv1.SecretReferenceWithOptionalNamespace{
						Name:      secretName,
						Namespace: namespace,
					},
					Key: "",
				}
			},
			invalid:         &Condition{Status: ConditionTrue, Reason: apiv1.AuthSecretKeyMissing, Message: "Auth secret key is missing"},
			unauthenticated: &Condition{Status: ConditionTrue, Reason: apiv1.Invalid, Message: "Auth secret key is missing"},
			stale:           &Condition{Status: ConditionUnknown, Reason: apiv1.Unauthenticated, Message: "Auth secret key is missing"}},
		"AuthSecretWithImplicitNamespaceNotFound": {
			repoContents: "bare",
			patProvider: func(namespace, secretName, secretKey string) *apiv1.GitHubRepositoryPersonalAccessToken {
				return &apiv1.GitHubRepositoryPersonalAccessToken{
					Secret: apiv1.SecretReferenceWithOptionalNamespace{
						Name: "non-existent-secret-implicitly-same-namespace",
					},
					Key: secretKey,
				}
			},
			restrictSecretRole: false,
			unauthenticated:    &Condition{Status: ConditionTrue, Reason: apiv1.AuthSecretNotFound, Message: "Secret '[a-z0-9]+/non-existent-secret-implicitly-same-namespace' not found"},
			stale:              &Condition{Status: ConditionUnknown, Reason: apiv1.Unauthenticated, Message: "Secret '[a-z0-9]+/non-existent-secret-implicitly-same-namespace' not found"},
		},
		"AuthSecretWithSpecificNamespaceNotFound": {
			repoContents: "bare",
			patProvider: func(namespace, secretName, secretKey string) *apiv1.GitHubRepositoryPersonalAccessToken {
				return &apiv1.GitHubRepositoryPersonalAccessToken{
					Secret: apiv1.SecretReferenceWithOptionalNamespace{
						Name:      "non-existent-secret-implicitly-same-namespace",
						Namespace: namespace,
					},
					Key: secretKey,
				}
			},
			restrictSecretRole: false,
			unauthenticated:    &Condition{Status: ConditionTrue, Reason: apiv1.AuthSecretNotFound, Message: "Secret '[a-z0-9]+/non-existent-secret-implicitly-same-namespace' not found"},
			stale:              &Condition{Status: ConditionUnknown, Reason: apiv1.Unauthenticated, Message: "Secret '[a-z0-9]+/non-existent-secret-implicitly-same-namespace' not found"},
		},
		"AuthSecretWithImplicitNamespaceNotAccessible": {
			repoContents: "bare",
			patProvider: func(namespace, secretName, secretKey string) *apiv1.GitHubRepositoryPersonalAccessToken {
				return &apiv1.GitHubRepositoryPersonalAccessToken{
					Secret: apiv1.SecretReferenceWithOptionalNamespace{
						Name: "non-existent-secret-implicitly-same-namespace",
					},
					Key: secretKey,
				}
			},
			restrictSecretRole: true,
			unauthenticated:    &Condition{Status: ConditionTrue, Reason: apiv1.AuthSecretForbidden, Message: "Secret '[a-z0-9]+/non-existent-secret-implicitly-same-namespace' is not accessible.*"},
			stale:              &Condition{Status: ConditionUnknown, Reason: apiv1.Unauthenticated, Message: "Secret '[a-z0-9]+/non-existent-secret-implicitly-same-namespace' is not accessible.*"},
		},
		"AuthSecretWithSpecificNamespaceNotAccessible": {
			repoContents: "bare",
			patProvider: func(namespace, secretName, secretKey string) *apiv1.GitHubRepositoryPersonalAccessToken {
				return &apiv1.GitHubRepositoryPersonalAccessToken{
					Secret: apiv1.SecretReferenceWithOptionalNamespace{
						Name:      "non-existent-secret-implicitly-same-namespace",
						Namespace: namespace,
					},
					Key: secretKey,
				}
			},
			restrictSecretRole: true,
			unauthenticated:    &Condition{Status: ConditionTrue, Reason: apiv1.AuthSecretForbidden, Message: "Secret '[a-z0-9]+/non-existent-secret-implicitly-same-namespace' is not accessible.*"},
			stale:              &Condition{Status: ConditionUnknown, Reason: apiv1.Unauthenticated, Message: "Secret '[a-z0-9]+/non-existent-secret-implicitly-same-namespace' is not accessible.*"},
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			t.Cleanup(cancel)

			k := NewKubernetes(t)
			ns := k.CreateNamespace(t, ctx)

			var ghRepo *GitHubRepositoryInfo
			var ghAuthSecretName, ghAuthSecretKeyName string
			var pat *apiv1.GitHubRepositoryPersonalAccessToken

			gh := NewGitHub(t, ctx)
			if tc.repoContents != "" {
				if tc.owner != "" || tc.name != "" {
					t.Fatalf("owner and name must be empty when repoContents is not empty")
				}
				ghRepo = gh.CreateRepository(t, ctx, tc.repoContents)
				ghAuthSecretName, ghAuthSecretKeyName = ns.CreateGitHubAuthSecret(t, ctx, gh.Token, tc.restrictSecretRole)
				tc.owner = ghRepo.Owner
				tc.name = ghRepo.Name
				pat = tc.patProvider(ns.Name, ghAuthSecretName, ghAuthSecretKeyName)
			}

			repoObjName := ns.CreateRepository(t, ctx, apiv1.RepositorySpec{
				GitHub: &apiv1.GitHubRepositorySpec{
					Owner:               tc.owner,
					Name:                tc.name,
					PersonalAccessToken: pat,
				},
				RefreshInterval: "10s",
			})

			For(t).Expect(func(t JustT) {
				repo := &apiv1.Repository{}
				For(t).Expect(k.Client.Get(ctx, client.ObjectKey{Namespace: ns.Name, Name: repoObjName}, repo)).Will(Succeed())
				For(t).Expect(repo.Status.GetFailedToInitializeCondition()).Will(BeNil())
				For(t).Expect(repo.Status.GetFinalizingCondition()).Will(BeNil())
				For(t).Expect(repo.Status.GetInvalidCondition()).Will(EqualCondition(tc.invalid))
				For(t).Expect(repo.Status.GetUnauthenticatedCondition()).Will(EqualCondition(tc.unauthenticated))
				For(t).Expect(repo.Status.GetStaleCondition()).Will(EqualCondition(tc.stale))
			}).Will(Eventually(Succeed()).Within(10 * time.Second).ProbingEvery(100 * time.Millisecond))
		})
	}
}
