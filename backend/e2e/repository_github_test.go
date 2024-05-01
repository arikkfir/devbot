package e2e_test

import (
	apiv1 "github.com/arikkfir/devbot/backend/api/v1"
	. "github.com/arikkfir/devbot/backend/e2e/expectations"
	"github.com/arikkfir/devbot/backend/internal/util/lang"
	. "github.com/arikkfir/devbot/backend/internal/util/testing"
	. "github.com/arikkfir/devbot/backend/internal/util/testing/justest"
	. "k8s.io/apimachinery/pkg/apis/meta/v1"
	"regexp"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"testing"
	"time"
)

func TestRepositoryGitHubConnection(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		owner, name        string
		repoContents       string
		patProvider        func(namespace, secretName, secretKey string) *apiv1.GitHubRepositoryPersonalAccessToken
		restrictSecretRole bool
		invalid            *ConditionE
		unauthenticated    *ConditionE
		stale              *ConditionE
	}{
		"OwnerMissing": {
			name: "myRepo",
			invalid: &ConditionE{
				Type:    apiv1.Invalid,
				Status:  lang.Ptr(string(ConditionTrue)),
				Reason:  regexp.MustCompile(regexp.QuoteMeta(apiv1.RepositoryOwnerMissing)),
				Message: regexp.MustCompile(regexp.QuoteMeta("Repository owner is empty")),
			},
			unauthenticated: &ConditionE{
				Type:    apiv1.Unauthenticated,
				Status:  lang.Ptr(string(ConditionTrue)),
				Reason:  regexp.MustCompile(regexp.QuoteMeta(apiv1.Invalid)),
				Message: regexp.MustCompile(regexp.QuoteMeta("Repository owner is empty")),
			},
			stale: &ConditionE{
				Type:    apiv1.Stale,
				Status:  lang.Ptr(string(ConditionUnknown)),
				Reason:  regexp.MustCompile(regexp.QuoteMeta(apiv1.Unauthenticated)),
				Message: regexp.MustCompile(regexp.QuoteMeta("Repository owner is empty")),
			},
		},
		"NameMissing": {
			owner: "myOwner",
			invalid: &ConditionE{
				Type:    apiv1.Invalid,
				Status:  lang.Ptr(string(ConditionTrue)),
				Reason:  regexp.MustCompile(regexp.QuoteMeta(apiv1.RepositoryNameMissing)),
				Message: regexp.MustCompile(regexp.QuoteMeta("Repository name is empty")),
			},
			unauthenticated: &ConditionE{
				Type:    apiv1.Unauthenticated,
				Status:  lang.Ptr(string(ConditionTrue)),
				Reason:  regexp.MustCompile(regexp.QuoteMeta(apiv1.Invalid)),
				Message: regexp.MustCompile(regexp.QuoteMeta("Repository name is empty")),
			},
			stale: &ConditionE{
				Type:    apiv1.Stale,
				Status:  lang.Ptr(string(ConditionUnknown)),
				Reason:  regexp.MustCompile(regexp.QuoteMeta(apiv1.Unauthenticated)),
				Message: regexp.MustCompile(regexp.QuoteMeta("Repository name is empty")),
			},
		},
		"NoAuthProvided": {
			owner: "someRepoOwner",
			name:  "someRepoName",
			patProvider: func(namespace, secretName, secretKey string) *apiv1.GitHubRepositoryPersonalAccessToken {
				return nil
			},
			invalid: &ConditionE{
				Type:    apiv1.Invalid,
				Status:  lang.Ptr(string(ConditionTrue)),
				Reason:  regexp.MustCompile(regexp.QuoteMeta(apiv1.AuthConfigMissing)),
				Message: regexp.MustCompile(regexp.QuoteMeta("Auth config is missing")),
			},
			unauthenticated: &ConditionE{
				Type:    apiv1.Unauthenticated,
				Status:  lang.Ptr(string(ConditionTrue)),
				Reason:  regexp.MustCompile(regexp.QuoteMeta(apiv1.Invalid)),
				Message: regexp.MustCompile(regexp.QuoteMeta("Auth config is missing")),
			},
			stale: &ConditionE{
				Type:    apiv1.Stale,
				Status:  lang.Ptr(string(ConditionUnknown)),
				Reason:  regexp.MustCompile(regexp.QuoteMeta(apiv1.Unauthenticated)),
				Message: regexp.MustCompile(regexp.QuoteMeta("Auth config is missing")),
			},
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
			invalid: &ConditionE{
				Type:    apiv1.Invalid,
				Status:  lang.Ptr(string(ConditionTrue)),
				Reason:  regexp.MustCompile(regexp.QuoteMeta(apiv1.AuthSecretNameMissing)),
				Message: regexp.MustCompile(regexp.QuoteMeta("Auth secret name is empty")),
			},
			unauthenticated: &ConditionE{
				Type:    apiv1.Unauthenticated,
				Status:  lang.Ptr(string(ConditionTrue)),
				Reason:  regexp.MustCompile(regexp.QuoteMeta(apiv1.Invalid)),
				Message: regexp.MustCompile(regexp.QuoteMeta("Auth secret name is empty")),
			},
			stale: &ConditionE{
				Type:    apiv1.Stale,
				Status:  lang.Ptr(string(ConditionUnknown)),
				Reason:  regexp.MustCompile(regexp.QuoteMeta(apiv1.Unauthenticated)),
				Message: regexp.MustCompile(regexp.QuoteMeta("Auth secret name is empty")),
			},
		},
		// "AuthSecretKeyMissing": {
		// 	repoContents: "repositories/bare",
		// 	patProvider: func(namespace, secretName, secretKey string) *apiv1.GitHubRepositoryPersonalAccessToken {
		// 		return &apiv1.GitHubRepositoryPersonalAccessToken{
		// 			Secret: apiv1.SecretReferenceWithOptionalNamespace{
		// 				Name:      secretName,
		// 				Namespace: namespace,
		// 			},
		// 			Key: "",
		// 		}
		// 	},
		// 	invalid: &ConditionE{
		// 		Type:    apiv1.Invalid,
		// 		Status:  lang.Ptr(string(ConditionTrue)),
		// 		Reason:  regexp.MustCompile(regexp.QuoteMeta(apiv1.AuthSecretKeyMissing)),
		// 		Message: regexp.MustCompile(regexp.QuoteMeta("Auth secret key is missing")),
		// 	},
		// 	unauthenticated: &ConditionE{
		// 		Type:    apiv1.Unauthenticated,
		// 		Status:  lang.Ptr(string(ConditionTrue)),
		// 		Reason:  regexp.MustCompile(regexp.QuoteMeta(apiv1.Invalid)),
		// 		Message: regexp.MustCompile(regexp.QuoteMeta("Auth secret key is missing")),
		// 	},
		// 	stale: &ConditionE{
		// 		Type:    apiv1.Stale,
		// 		Status:  lang.Ptr(string(ConditionUnknown)),
		// 		Reason:  regexp.MustCompile(regexp.QuoteMeta(apiv1.Unauthenticated)),
		// 		Message: regexp.MustCompile(regexp.QuoteMeta("Auth secret key is missing")),
		// 	},
		// },
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
			unauthenticated: &ConditionE{
				Type:    apiv1.Unauthenticated,
				Status:  lang.Ptr(string(ConditionTrue)),
				Reason:  regexp.MustCompile(regexp.QuoteMeta(apiv1.AuthSecretNotFound)),
				Message: regexp.MustCompile(regexp.QuoteMeta("Secret '[a-z0-9]+/non-existent-secret-implicitly-same-namespace' not found")),
			},
			stale: &ConditionE{
				Type:    apiv1.Stale,
				Status:  lang.Ptr(string(ConditionUnknown)),
				Reason:  regexp.MustCompile(regexp.QuoteMeta(apiv1.Unauthenticated)),
				Message: regexp.MustCompile(regexp.QuoteMeta("Secret '[a-z0-9]+/non-existent-secret-implicitly-same-namespace' not found")),
			},
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
			unauthenticated: &ConditionE{
				Type:    apiv1.Unauthenticated,
				Status:  lang.Ptr(string(ConditionTrue)),
				Reason:  regexp.MustCompile(regexp.QuoteMeta(apiv1.AuthSecretNotFound)),
				Message: regexp.MustCompile(regexp.QuoteMeta("Secret '[a-z0-9]+/non-existent-secret-implicitly-same-namespace' not found")),
			},
			stale: &ConditionE{
				Type:    apiv1.Stale,
				Status:  lang.Ptr(string(ConditionUnknown)),
				Reason:  regexp.MustCompile(regexp.QuoteMeta(apiv1.Unauthenticated)),
				Message: regexp.MustCompile(regexp.QuoteMeta("Secret '[a-z0-9]+/non-existent-secret-implicitly-same-namespace' not found")),
			},
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
			unauthenticated: &ConditionE{
				Type:    apiv1.Unauthenticated,
				Status:  lang.Ptr(string(ConditionTrue)),
				Reason:  regexp.MustCompile(regexp.QuoteMeta(apiv1.AuthSecretForbidden)),
				Message: regexp.MustCompile(regexp.QuoteMeta("Secret '[a-z0-9]+/non-existent-secret-implicitly-same-namespace' is not accessible.*")),
			},
			stale: &ConditionE{
				Type:    apiv1.Stale,
				Status:  lang.Ptr(string(ConditionUnknown)),
				Reason:  regexp.MustCompile(regexp.QuoteMeta(apiv1.Unauthenticated)),
				Message: regexp.MustCompile(regexp.QuoteMeta("Secret '[a-z0-9]+/non-existent-secret-implicitly-same-namespace' is not accessible.*")),
			},
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
			unauthenticated: &ConditionE{
				Type:    apiv1.Unauthenticated,
				Status:  lang.Ptr(string(ConditionTrue)),
				Reason:  regexp.MustCompile(regexp.QuoteMeta(apiv1.AuthSecretForbidden)),
				Message: regexp.MustCompile(regexp.QuoteMeta("Secret '[a-z0-9]+/non-existent-secret-implicitly-same-namespace' is not accessible.*")),
			},
			stale: &ConditionE{
				Type:    apiv1.Stale,
				Status:  lang.Ptr(string(ConditionUnknown)),
				Reason:  regexp.MustCompile(regexp.QuoteMeta(apiv1.Unauthenticated)),
				Message: regexp.MustCompile(regexp.QuoteMeta("Secret '[a-z0-9]+/non-existent-secret-implicitly-same-namespace' is not accessible.*")),
			},
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			e2e := NewE2E(t)
			ns := e2e.K.CreateNamespace(e2e.Ctx, t)

			var ghRepo *GitHubRepositoryInfo
			var ghAuthSecretName, ghAuthSecretKeyName string
			var pat *apiv1.GitHubRepositoryPersonalAccessToken

			if tc.repoContents != "" {
				if tc.owner != "" || tc.name != "" {
					t.Fatalf("owner and name must be empty when repoContents is not empty")
				}
				ghRepo = e2e.GH.CreateRepository(e2e.Ctx, t, repositoriesFS, tc.repoContents)
				ghAuthSecretName, ghAuthSecretKeyName = ns.CreateGitHubAuthSecret(e2e.Ctx, t, e2e.GH.Token, tc.restrictSecretRole)
				tc.owner = ghRepo.Owner
				tc.name = ghRepo.Name
				pat = tc.patProvider(ns.Name, ghAuthSecretName, ghAuthSecretKeyName)
			}

			kRepoName := ns.CreateRepository(e2e.Ctx, t, apiv1.RepositorySpec{
				GitHub: &apiv1.GitHubRepositorySpec{
					Owner:               tc.owner,
					Name:                tc.name,
					PersonalAccessToken: pat,
				},
				RefreshInterval: "10s",
			})

			With(t).Verify(func(t T) {
				repositoryExpectation := RepositoryE{
					Name: kRepoName,
					Status: RepositoryStatusE{
						Conditions: map[string]*ConditionE{
							apiv1.FailedToInitialize: nil,
							apiv1.Finalizing:         nil,
							apiv1.Invalid:            tc.invalid,
							apiv1.Stale:              tc.unauthenticated,
							apiv1.Unauthenticated:    tc.stale,
						},
						DefaultBranch: "main",
					},
				}
				repo := &apiv1.Repository{}
				With(t).Verify(e2e.K.Client.Get(e2e.Ctx, client.ObjectKey{Namespace: ns.Name, Name: kRepoName}, repo)).Will(Succeed())
				With(t).Verify(repo).Will(EqualTo(repositoryExpectation).Using(RepositoryComparator))
			}).Will(Succeed()).Within(10*time.Second, 100*time.Millisecond)
		})
	}
}
