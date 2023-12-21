package github_test

import (
	"context"
	"fmt"
	apiv1 "github.com/arikkfir/devbot/backend/api/v1"
	controller "github.com/arikkfir/devbot/backend/internal/controllers/repositories/github"
	k8sutil "github.com/arikkfir/devbot/backend/internal/util/k8s"
	stringsutil "github.com/arikkfir/devbot/backend/internal/util/strings"
	. "github.com/arikkfir/devbot/backend/internal/util/testing"
	"github.com/google/go-github/v56/github"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"time"
)

var _ = Describe(apiv1.GitHubRepositoryGVK.Kind, func() {
	var k8s client.Client
	var gh *github.Client
	var repoName string

	BeforeEach(func(ctx context.Context) { CreateKubernetesClient(&k8s) })
	BeforeEach(func(ctx context.Context) { CreateGitHubClient(ctx, k8s, &gh) })
	BeforeEach(func(ctx context.Context) { CreateGitHubRepository(ctx, k8s, gh, &repoName) })

	Describe("initialization", func() {
		var crNamespace, crName string
		BeforeEach(func(ctx context.Context) {
			CreateGitHubRepositoryCR(ctx, k8s, repoName, &crName, &crNamespace,
				func(r *apiv1.GitHubRepository) { SetBreakAnnotation(r, "init", false, 0) },
			)
		})
		It("should initialize conditions", func(ctx context.Context) {
			Eventually(func(o Gomega) {
				r := &apiv1.GitHubRepository{}
				o.Expect(k8s.Get(ctx, client.ObjectKey{Namespace: crNamespace, Name: crName}, r)).To(Succeed())
				o.Expect(r.Status.Conditions).To(ConsistOf(MatchFields(IgnoreExtras, Fields{
					"Type":    Equal(apiv1.ConditionTypeCurrent),
					"Status":  Equal(metav1.ConditionUnknown),
					"Reason":  Equal(apiv1.ReasonInitializing),
					"Message": Equal("Initializing"),
				}), MatchFields(IgnoreExtras, Fields{
					"Type":    Equal(apiv1.ConditionTypeAuthenticatedToGitHub),
					"Status":  Equal(metav1.ConditionUnknown),
					"Reason":  Equal(apiv1.ReasonInitializing),
					"Message": Equal("Initializing"),
				})))
			}, 30*time.Second, 1*time.Second).Should(Succeed())
		})
		It("should initialize finalizer", func(ctx context.Context) {
			Eventually(func(o Gomega) {
				r := &apiv1.GitHubRepository{}
				o.Expect(k8s.Get(ctx, client.ObjectKey{Namespace: crNamespace, Name: crName}, r)).To(Succeed())
				o.Expect(r.Finalizers).To(ContainElement(controller.RepositoryFinalizer))
			}, 30*time.Second, 1*time.Second).Should(Succeed())
		})
	})
	Describe("auth configuration parsing", func() {
		When("setting a nil PAT config", func() {
			var crNamespace, crName string
			BeforeEach(func(ctx context.Context) {
				CreateGitHubRepositoryCR(ctx, k8s, repoName, &crName, &crNamespace, func(r *apiv1.GitHubRepository) {
					r.Spec.Auth.PersonalAccessToken = nil
				})
			})
			It("should correctly signal auth condition", func(ctx context.Context) {
				Eventually(func(o Gomega) {
					r := &apiv1.GitHubRepository{}
					o.Expect(k8s.Get(ctx, client.ObjectKey{Namespace: crNamespace, Name: crName}, r)).To(Succeed())
					o.Expect(r.Status.Conditions).To(ContainElement(MatchFields(IgnoreExtras, Fields{
						"Type":    Equal(apiv1.ConditionTypeAuthenticatedToGitHub),
						"Status":  Equal(metav1.ConditionFalse),
						"Reason":  Equal(apiv1.ReasonAuthConfigError),
						"Message": Equal("Auth config is missing"),
					})))
				}, 30*time.Second, 1*time.Second).Should(Succeed())
			})
		})
		When("setting an empty secret name", func() {
			var crNamespace, crName string
			BeforeEach(func(ctx context.Context) {
				CreateGitHubRepositoryCR(ctx, k8s, repoName, &crName, &crNamespace, func(r *apiv1.GitHubRepository) {
					r.Spec.Auth.PersonalAccessToken = &apiv1.GitHubRepositoryAuthPersonalAccessToken{
						Secret: corev1.SecretReference{Name: "", Namespace: "foo"},
					}
				})
			})
			It("should correctly signal auth condition", func(ctx context.Context) {
				Eventually(func(o Gomega) {
					r := &apiv1.GitHubRepository{}
					o.Expect(k8s.Get(ctx, client.ObjectKey{Namespace: crNamespace, Name: crName}, r)).To(Succeed())
					o.Expect(r.Status.Conditions).To(ContainElement(MatchFields(IgnoreExtras, Fields{
						"Type":    Equal(apiv1.ConditionTypeAuthenticatedToGitHub),
						"Status":  Equal(metav1.ConditionFalse),
						"Reason":  Equal(apiv1.ReasonGitHubAuthSecretNameMissing),
						"Message": Equal("Auth secret name may not be empty"),
					})))
				}, 2*time.Minute, 1*time.Second).Should(Succeed())
			})
		})
		When("setting a non-existing secret name", func() {
			var crNamespace, crName string
			var hash = "github-auth-" + stringsutil.RandomHash(4)
			BeforeEach(func(ctx context.Context) {
				// Grant the role so that the controller can read secrets in general, even if our secret doesn't exist
				role := &v1.ClusterRole{
					ObjectMeta: metav1.ObjectMeta{Name: hash},
					Rules: []v1.PolicyRule{
						{
							APIGroups: []string{""},
							Resources: []string{"secrets"},
							Verbs:     []string{"get", "list"},
						},
					},
				}
				Expect(k8s.Create(ctx, role)).To(Succeed())
				DeferCleanup(func() { Expect(k8s.Delete(context.Background(), role)).To(Succeed()) })
				roleBinding := &v1.RoleBinding{
					ObjectMeta: metav1.ObjectMeta{Namespace: "default", Name: hash},
					RoleRef:    v1.RoleRef{APIGroup: "rbac.authorization.k8s.io", Kind: "ClusterRole", Name: hash},
					Subjects:   []v1.Subject{{Kind: "ServiceAccount", Name: "devbot-github-repository-controller", Namespace: "devbot"}},
				}
				Expect(k8s.Create(ctx, roleBinding)).To(Succeed())
				DeferCleanup(func() { Expect(k8s.Delete(context.Background(), roleBinding)).To(Succeed()) })
				CreateGitHubRepositoryCR(ctx, k8s, repoName, &crName, &crNamespace, func(r *apiv1.GitHubRepository) {
					r.Spec.Auth.PersonalAccessToken = &apiv1.GitHubRepositoryAuthPersonalAccessToken{
						Secret: corev1.SecretReference{Name: "non-existing-secret", Namespace: "default"},
					}
				})
			})
			It("should correctly signal auth condition", func(ctx context.Context) {
				Eventually(func(o Gomega) {
					r := &apiv1.GitHubRepository{}
					o.Expect(k8s.Get(ctx, client.ObjectKey{Namespace: crNamespace, Name: crName}, r)).To(Succeed())
					o.Expect(r.Status.Conditions).To(ContainElement(MatchFields(IgnoreExtras, Fields{
						"Type":    Equal(apiv1.ConditionTypeAuthenticatedToGitHub),
						"Status":  Equal(metav1.ConditionFalse),
						"Reason":  Equal(apiv1.ReasonGitHubAuthSecretNotFound),
						"Message": Equal("Secret 'default/non-existing-secret' not found"),
					})))
				}, 2*time.Minute, 1*time.Second).Should(Succeed())
			})
		})
		When("setting an existing but inaccessible secret name", func() {
			var crNamespace, crName string
			var secretName = "github-auth-" + stringsutil.RandomHash(4)
			BeforeEach(func(ctx context.Context) {
				secret := &corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{Namespace: "default", Name: secretName},
					StringData: map[string]string{"pat": ObtainGitHubPAT(ctx, k8s)},
					Type:       corev1.SecretTypeOpaque,
				}
				Expect(k8s.Create(ctx, secret)).To(Succeed())
				DeferCleanup(func() { Expect(k8s.Delete(context.Background(), secret)).To(Succeed()) })
				CreateGitHubRepositoryCR(ctx, k8s, repoName, &crName, &crNamespace, func(r *apiv1.GitHubRepository) {
					r.Spec.Auth.PersonalAccessToken = &apiv1.GitHubRepositoryAuthPersonalAccessToken{
						Secret: corev1.SecretReference{Name: secretName, Namespace: "default"},
						Key:    "pat",
					}
				})
			})
			It("should correctly signal auth condition", func(ctx context.Context) {
				Eventually(func(o Gomega) {
					r := &apiv1.GitHubRepository{}
					o.Expect(k8s.Get(ctx, client.ObjectKey{Namespace: crNamespace, Name: crName}, r)).To(Succeed())
					o.Expect(r.Status.Conditions).To(ContainElement(MatchFields(IgnoreExtras, Fields{
						"Type":    Equal(apiv1.ConditionTypeAuthenticatedToGitHub),
						"Status":  Equal(metav1.ConditionFalse),
						"Reason":  Equal(apiv1.ReasonGitHubAuthSecretForbidden),
						"Message": Equal(fmt.Sprintf("Forbidden from reading secret 'default/%s': secrets \"%s\" is forbidden: User \"system:serviceaccount:devbot:devbot-github-repository-controller\" cannot get resource \"secrets\" in API group \"\" in the namespace \"default\"", secretName, secretName)),
					})))
				}, 30*time.Second, 1*time.Second).Should(Succeed())
			})
		})
		When("creating an accessible secret", func() {
			var crNamespace, crName string
			var secretName = "github-auth-" + stringsutil.RandomHash(4)
			BeforeEach(func(ctx context.Context) {
				secret := &corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{Namespace: "default", Name: secretName},
					StringData: map[string]string{"invalid-pat": "invalid-pat", "empty-pat": "", "valid-pat": ObtainGitHubPAT(ctx, k8s)},
					Type:       corev1.SecretTypeOpaque,
				}
				Expect(k8s.Create(ctx, secret)).To(Succeed())
				DeferCleanup(func() { Expect(k8s.Delete(context.Background(), secret)).To(Succeed()) })
				role := &v1.ClusterRole{
					ObjectMeta: metav1.ObjectMeta{Name: secretName},
					Rules: []v1.PolicyRule{
						{
							APIGroups:     []string{""},
							Resources:     []string{"secrets"},
							ResourceNames: []string{secretName},
							Verbs:         []string{"get"},
						},
					},
				}
				Expect(k8s.Create(ctx, role)).To(Succeed())
				DeferCleanup(func() { Expect(k8s.Delete(context.Background(), role)).To(Succeed()) })
				roleBinding := &v1.RoleBinding{
					ObjectMeta: metav1.ObjectMeta{Namespace: "default", Name: secretName},
					RoleRef:    v1.RoleRef{APIGroup: "rbac.authorization.k8s.io", Kind: "ClusterRole", Name: secretName},
					Subjects:   []v1.Subject{{Kind: "ServiceAccount", Name: "devbot-github-repository-controller", Namespace: "devbot"}},
				}
				Expect(k8s.Create(ctx, roleBinding)).To(Succeed())
				DeferCleanup(func() { Expect(k8s.Delete(context.Background(), roleBinding)).To(Succeed()) })
			})
			When("not specifying the secret namespace", func() {
				BeforeEach(func(ctx context.Context) {
					CreateGitHubRepositoryCR(ctx, k8s, repoName, &crName, &crNamespace, func(r *apiv1.GitHubRepository) {
						r.Spec.Auth.PersonalAccessToken = &apiv1.GitHubRepositoryAuthPersonalAccessToken{
							Secret: corev1.SecretReference{Name: secretName},
							Key:    "valid-pat",
						}
					})
				})
				It("should default secret namespace to repository namespace", func(ctx context.Context) {
					Eventually(func(o Gomega) {
						r := &apiv1.GitHubRepository{}
						o.Expect(k8s.Get(ctx, client.ObjectKey{Namespace: crNamespace, Name: crName}, r)).To(Succeed())
						o.Expect(r.Status.Conditions).To(ContainElement(MatchFields(IgnoreExtras, Fields{
							"Type":    Equal(apiv1.ConditionTypeAuthenticatedToGitHub),
							"Status":  Equal(metav1.ConditionTrue),
							"Reason":  Equal(apiv1.ReasonAuthenticated),
							"Message": Equal("Authenticated to GitHub"),
						})))
					}, 30*time.Second, 1*time.Second).Should(Succeed())
				})
			})
			When("specifying the secret namespace", func() {
				When("using a key with an invalid token", func() {
					BeforeEach(func(ctx context.Context) {
						CreateGitHubRepositoryCR(ctx, k8s, repoName, &crName, &crNamespace, func(r *apiv1.GitHubRepository) {
							r.Spec.Auth.PersonalAccessToken = &apiv1.GitHubRepositoryAuthPersonalAccessToken{
								Secret: corev1.SecretReference{Name: secretName, Namespace: "default"},
								Key:    "invalid-pat",
							}
						})
					})
					It("should correctly signal auth condition", func(ctx context.Context) {
						Eventually(func(o Gomega) {
							r := &apiv1.GitHubRepository{}
							o.Expect(k8s.Get(ctx, client.ObjectKey{Namespace: crNamespace, Name: crName}, r)).To(Succeed())
							o.Expect(r.Status.Conditions).To(ContainElement(MatchFields(IgnoreExtras, Fields{
								"Type":    Equal(apiv1.ConditionTypeAuthenticatedToGitHub),
								"Status":  Equal(metav1.ConditionFalse),
								"Reason":  Equal(apiv1.ReasonGitHubAPIFailed),
								"Message": MatchRegexp(`GitHub connection failed: GET https://api.github.com/user: 401 Bad credentials.*`),
							})))
						}, 30*time.Second, 1*time.Second).Should(Succeed())
					})
				})
				When("using a key with an empty value", func() {
					BeforeEach(func(ctx context.Context) {
						CreateGitHubRepositoryCR(ctx, k8s, repoName, &crName, &crNamespace, func(r *apiv1.GitHubRepository) {
							r.Spec.Auth.PersonalAccessToken = &apiv1.GitHubRepositoryAuthPersonalAccessToken{
								Secret: corev1.SecretReference{Name: secretName, Namespace: "default"},
								Key:    "empty-pat",
							}
						})
					})
					It("should correctly signal auth condition", func(ctx context.Context) {
						Eventually(func(o Gomega) {
							r := &apiv1.GitHubRepository{}
							o.Expect(k8s.Get(ctx, client.ObjectKey{Namespace: crNamespace, Name: crName}, r)).To(Succeed())
							o.Expect(r.Status.Conditions).To(ContainElement(MatchFields(IgnoreExtras, Fields{
								"Type":    Equal(apiv1.ConditionTypeAuthenticatedToGitHub),
								"Status":  Equal(metav1.ConditionFalse),
								"Reason":  Equal(apiv1.ReasonGitHubAuthSecretEmptyToken),
								"Message": Equal(fmt.Sprintf("Key 'empty-pat' in secret 'default/%s' is missing or empty", secretName)),
							})))
						}, 30*time.Second, 1*time.Second).Should(Succeed())
					})
				})
				When("using a missing secret key", func() {
					BeforeEach(func(ctx context.Context) {
						CreateGitHubRepositoryCR(ctx, k8s, repoName, &crName, &crNamespace, func(r *apiv1.GitHubRepository) {
							r.Spec.Auth.PersonalAccessToken = &apiv1.GitHubRepositoryAuthPersonalAccessToken{
								Secret: corev1.SecretReference{Name: secretName, Namespace: "default"},
								Key:    "missing-pat",
							}
						})
					})
					It("should signal unknown condition", func(ctx context.Context) {
						Eventually(func(o Gomega) {
							r := &apiv1.GitHubRepository{}
							o.Expect(k8s.Get(ctx, client.ObjectKey{Namespace: crNamespace, Name: crName}, r)).To(Succeed())
							o.Expect(r.Status.Conditions).To(ContainElement(MatchFields(IgnoreExtras, Fields{
								"Type":    Equal(apiv1.ConditionTypeAuthenticatedToGitHub),
								"Status":  Equal(metav1.ConditionFalse),
								"Reason":  Equal(apiv1.ReasonGitHubAuthSecretEmptyToken),
								"Message": Equal(fmt.Sprintf("Key 'missing-pat' in secret 'default/%s' is missing or empty", secretName)),
							})))
						}, 30*time.Second, 1*time.Second).Should(Succeed())
					})
				})
			})
		})
	})
	Describe("branch detection", func() {
		var crNamespace, crName string
		BeforeEach(func(ctx context.Context) { CreateGitHubRepositoryCR(ctx, k8s, repoName, &crName, &crNamespace) })
		When("main branch exists", func() {
			It("should detect main branch", func(ctx context.Context) {
				Eventually(func(o Gomega) {
					r, refs := &apiv1.GitHubRepository{}, &apiv1.GitHubRepositoryRefList{}
					o.Expect(k8s.Get(ctx, client.ObjectKey{Namespace: crNamespace, Name: crName}, r)).To(Succeed())
					o.Expect(r.Status.Conditions).To(ContainElement(MatchFields(IgnoreExtras, Fields{
						"Type":    Equal(apiv1.ConditionTypeCurrent),
						"Status":  Equal(metav1.ConditionTrue),
						"Reason":  Equal(apiv1.ReasonSynced),
						"Message": Equal("Synchronized refs"),
					})))
					o.Expect(k8sutil.GetOwnedChildrenManually(ctx, k8s, r, refs)).To(Succeed())
					o.Expect(refs.Items).To(ConsistOf(
						MatchFields(IgnoreExtras, Fields{
							"Spec": MatchFields(IgnoreExtras, Fields{
								"Ref": Equal("refs/heads/main"),
							}),
						}),
					))
				}, 30*time.Second, 1*time.Second).Should(Succeed())
			})
		})
		When("creating a new branch", func() {
			const myBranchName = "my-branch"
			var myBranchSHA string
			BeforeEach(func(ctx context.Context) { CreateGitHubBranch(ctx, gh, repoName, myBranchName, true) })
			BeforeEach(func(ctx context.Context) { GetGitHubBranchCommitSHA(ctx, gh, repoName, myBranchName, &myBranchSHA) })
			It("should detect the new branch", func(ctx context.Context) {
				Eventually(func(o Gomega) {
					r, refs := &apiv1.GitHubRepository{}, &apiv1.GitHubRepositoryRefList{}
					o.Expect(k8s.Get(ctx, client.ObjectKey{Namespace: crNamespace, Name: crName}, r)).To(Succeed())
					o.Expect(r.Status.Conditions).To(ContainElement(MatchFields(IgnoreExtras, Fields{
						"Type":    Equal(apiv1.ConditionTypeCurrent),
						"Status":  Equal(metav1.ConditionTrue),
						"Reason":  Equal(apiv1.ReasonSynced),
						"Message": Equal("Synchronized refs"),
					})))
					o.Expect(k8sutil.GetOwnedChildrenManually(ctx, k8s, r, refs)).To(Succeed())
					o.Expect(refs.Items).To(ConsistOf(
						MatchFields(IgnoreExtras, Fields{
							"Spec": MatchFields(IgnoreExtras, Fields{
								"Ref": Equal("refs/heads/main"),
							}),
						}),
						MatchFields(IgnoreExtras, Fields{
							"Spec": MatchFields(IgnoreExtras, Fields{
								"Ref": Equal("refs/heads/" + myBranchName),
							}),
						}),
					))
				}, 30*time.Second, 1*time.Second).Should(Succeed())
			})
		})
		When("deleting a branch", func() {
			const myBranchName = "my-branch"
			var myBranchSHA string
			BeforeEach(func(ctx context.Context) { CreateGitHubBranch(ctx, gh, repoName, myBranchName, false) })
			BeforeEach(func(ctx context.Context) { GetGitHubBranchCommitSHA(ctx, gh, repoName, myBranchName, &myBranchSHA) })
			It("should delete the branch ref object", func(ctx context.Context) {
				Eventually(func(o Gomega) {
					r, refs := &apiv1.GitHubRepository{}, &apiv1.GitHubRepositoryRefList{}
					o.Expect(k8s.Get(ctx, client.ObjectKey{Namespace: crNamespace, Name: crName}, r)).To(Succeed())
					o.Expect(r.Status.Conditions).To(ContainElement(MatchFields(IgnoreExtras, Fields{
						"Type":    Equal(apiv1.ConditionTypeCurrent),
						"Status":  Equal(metav1.ConditionTrue),
						"Reason":  Equal(apiv1.ReasonSynced),
						"Message": Equal("Synchronized refs"),
					})))
					o.Expect(k8sutil.GetOwnedChildrenManually(ctx, k8s, r, refs)).To(Succeed())
					o.Expect(refs.Items).To(ConsistOf(
						MatchFields(IgnoreExtras, Fields{
							"Spec": MatchFields(IgnoreExtras, Fields{
								"Ref": Equal("refs/heads/main"),
							}),
						}),
						MatchFields(IgnoreExtras, Fields{
							"Spec": MatchFields(IgnoreExtras, Fields{
								"Ref": Equal("refs/heads/" + myBranchName),
							}),
						}),
					))
				}, 30*time.Second, 1*time.Second).Should(Succeed())

				DeleteGitHubBranch(ctx, gh, repoName, myBranchName)

				Eventually(func(o Gomega) {
					r, refs := &apiv1.GitHubRepository{}, &apiv1.GitHubRepositoryRefList{}
					o.Expect(k8s.Get(ctx, client.ObjectKey{Namespace: crNamespace, Name: crName}, r)).To(Succeed())
					o.Expect(r.Status.Conditions).To(ContainElement(MatchFields(IgnoreExtras, Fields{
						"Type":    Equal(apiv1.ConditionTypeCurrent),
						"Status":  Equal(metav1.ConditionTrue),
						"Reason":  Equal(apiv1.ReasonSynced),
						"Message": Equal("Synchronized refs"),
					})))
					o.Expect(k8sutil.GetOwnedChildrenManually(ctx, k8s, r, refs)).To(Succeed())
					o.Expect(refs.Items).To(ConsistOf(
						MatchFields(IgnoreExtras, Fields{
							"Spec": MatchFields(IgnoreExtras, Fields{
								"Ref": Equal("refs/heads/main"),
							}),
						}),
					))
				}, 30*time.Second, 1*time.Second).Should(Succeed())
			})
		})
	})
})
