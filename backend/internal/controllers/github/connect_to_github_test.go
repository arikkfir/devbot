package github_test

import (
	"context"
	. "github.com/arikkfir/devbot/backend/api/v1"
	act "github.com/arikkfir/devbot/backend/internal/controllers/github"
	"github.com/arikkfir/devbot/backend/internal/util/strings"
	. "github.com/arikkfir/devbot/backend/internal/util/testing"
	"github.com/google/go-github/v56/github"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"io"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"os"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"
	"time"
)

var _ = Describe("NewConnectToGitHubAction", func() {
	var namespace, repoName string
	BeforeEach(func(ctx context.Context) {
		namespace = "default"
		repoName = strings.RandomHash(7)
	})

	When("auth config is not provided", func() {
		var k client.WithWatch
		BeforeEach(func(ctx context.Context) {
			o := &GitHubRepository{ObjectMeta: metav1.ObjectMeta{Name: repoName, Namespace: namespace}}
			k = fake.NewClientBuilder().WithScheme(scheme).WithObjects(o).WithStatusSubresource(o).Build()
		})
		It("should set conditions and abort", func(ctx context.Context) {
			o := &GitHubRepository{}
			Expect(k.Get(ctx, client.ObjectKeyFromObject(o), o)).To(Succeed())
			result, err := act.NewConnectToGitHubAction(0, nil).Execute(ctx, k, o)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal(&ctrl.Result{}))

			oo := &GitHubRepository{}
			Expect(k.Get(ctx, client.ObjectKeyFromObject(o), oo)).To(Succeed())
			Expect(o.Status.GetInvalidCondition()).To(BeTrueDueTo(AuthConfigMissing))
			Expect(o.Status.GetUnauthenticatedCondition()).To(BeTrueDueTo(Invalid))
			Expect(o.Status.GetStaleCondition()).To(BeTrueDueTo(Unauthenticated))
		})
	})

	When("pat secret name is not provided", func() {
		var k client.WithWatch
		BeforeEach(func(ctx context.Context) {
			o := &GitHubRepository{
				ObjectMeta: metav1.ObjectMeta{Name: repoName, Namespace: namespace},
				Spec: GitHubRepositorySpec{
					Auth: GitHubRepositoryAuth{PersonalAccessToken: &GitHubRepositoryAuthPersonalAccessToken{}},
				},
			}
			k = fake.NewClientBuilder().WithScheme(scheme).WithObjects(o).WithStatusSubresource(o).Build()
		})
		It("should set conditions and abort", func(ctx context.Context) {
			o := &GitHubRepository{}
			Expect(k.Get(ctx, client.ObjectKeyFromObject(o), o)).To(Succeed())
			result, err := act.NewConnectToGitHubAction(0, nil).Execute(ctx, k, o)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal(&ctrl.Result{}))

			oo := &GitHubRepository{}
			Expect(k.Get(ctx, client.ObjectKeyFromObject(o), oo)).To(Succeed())
			Expect(o.Status.GetInvalidCondition()).To(BeTrueDueTo(AuthSecretNameMissing))
			Expect(o.Status.GetUnauthenticatedCondition()).To(BeTrueDueTo(Invalid))
			Expect(o.Status.GetStaleCondition()).To(BeUnknownDueTo(Unauthenticated))
		})
	})

	When("pat secret is not found", func() {
		const refreshInterval = 5 * time.Minute
		var k client.WithWatch
		BeforeEach(func(ctx context.Context) {
			o := &GitHubRepository{
				ObjectMeta: metav1.ObjectMeta{Name: repoName, Namespace: namespace},
				Spec: GitHubRepositorySpec{
					Auth: GitHubRepositoryAuth{PersonalAccessToken: &GitHubRepositoryAuthPersonalAccessToken{
						Secret: SecretReferenceWithOptionalNamespace{Name: strings.RandomHash(7)},
					}},
				},
			}
			k = fake.NewClientBuilder().WithScheme(scheme).WithObjects(o).WithStatusSubresource(o).Build()
		})
		It("should set conditions and abort", func(ctx context.Context) {
			o := &GitHubRepository{}
			Expect(k.Get(ctx, client.ObjectKeyFromObject(o), o)).To(Succeed())
			result, err := act.NewConnectToGitHubAction(refreshInterval, nil).Execute(ctx, k, o)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal(&ctrl.Result{RequeueAfter: refreshInterval}))

			oo := &GitHubRepository{}
			Expect(k.Get(ctx, client.ObjectKeyFromObject(o), oo)).To(Succeed())
			Expect(o.Status.GetInvalidCondition()).To(BeNil())
			Expect(o.Status.GetUnauthenticatedCondition()).To(BeTrueDueTo(AuthSecretNotFound))
			Expect(o.Status.GetStaleCondition()).To(BeUnknownDueTo(Unauthenticated))
		})
	})

	When("pat secret is inaccessible", func() {
		const refreshInterval = 5 * time.Minute
		var k client.WithWatch
		BeforeEach(func(ctx context.Context) {
			o := &GitHubRepository{
				ObjectMeta: metav1.ObjectMeta{Name: repoName, Namespace: namespace},
				Spec: GitHubRepositorySpec{
					Auth: GitHubRepositoryAuth{PersonalAccessToken: &GitHubRepositoryAuthPersonalAccessToken{
						Secret: SecretReferenceWithOptionalNamespace{Name: strings.RandomHash(7)},
					}},
				},
			}
			k = fake.NewClientBuilder().
				WithScheme(scheme).
				WithObjects(o).
				WithStatusSubresource(o).
				WithInterceptorFuncs(interceptor.Funcs{
					Get: func(ctx context.Context, c client.WithWatch, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
						if key.Name == o.Spec.Auth.PersonalAccessToken.Secret.Name && key.Namespace == o.Namespace {
							return apierrors.NewForbidden(schema.GroupResource{}, key.Name, io.EOF)
						}
						return c.Get(ctx, key, obj, opts...)
					},
				}).
				Build()
		})
		It("should set conditions and abort", func(ctx context.Context) {
			o := &GitHubRepository{}
			Expect(k.Get(ctx, client.ObjectKeyFromObject(o), o)).To(Succeed())
			result, err := act.NewConnectToGitHubAction(refreshInterval, nil).Execute(ctx, k, o)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal(&ctrl.Result{RequeueAfter: refreshInterval}))

			oo := &GitHubRepository{}
			Expect(k.Get(ctx, client.ObjectKeyFromObject(o), oo)).To(Succeed())
			Expect(o.Status.GetInvalidCondition()).To(BeNil())
			Expect(o.Status.GetUnauthenticatedCondition()).To(BeTrueDueTo(AuthSecretForbidden))
			Expect(o.Status.GetStaleCondition()).To(BeUnknownDueTo(Unauthenticated))
		})
	})

	When("pat secret fails to be fetched due to an internal error", func() {
		const refreshInterval = 5 * time.Minute
		var k client.WithWatch
		BeforeEach(func(ctx context.Context) {
			o := &GitHubRepository{
				ObjectMeta: metav1.ObjectMeta{Name: repoName, Namespace: namespace},
				Spec: GitHubRepositorySpec{
					Auth: GitHubRepositoryAuth{PersonalAccessToken: &GitHubRepositoryAuthPersonalAccessToken{
						Secret: SecretReferenceWithOptionalNamespace{Name: strings.RandomHash(7)},
					}},
				},
			}
			k = fake.NewClientBuilder().
				WithScheme(scheme).
				WithObjects(o).
				WithStatusSubresource(o).
				WithInterceptorFuncs(interceptor.Funcs{
					Get: func(ctx context.Context, c client.WithWatch, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
						if key.Name == o.Spec.Auth.PersonalAccessToken.Secret.Name && key.Namespace == o.Namespace {
							return apierrors.NewInternalError(io.EOF)
						}
						return c.Get(ctx, key, obj, opts...)
					},
				}).
				Build()
		})
		It("should set conditions and abort", func(ctx context.Context) {
			o := &GitHubRepository{}
			Expect(k.Get(ctx, client.ObjectKeyFromObject(o), o)).To(Succeed())
			result, err := act.NewConnectToGitHubAction(refreshInterval, nil).Execute(ctx, k, o)
			Expect(err).ToNot(BeNil())
			Expect(result).To(Equal(&ctrl.Result{}))

			oo := &GitHubRepository{}
			Expect(k.Get(ctx, client.ObjectKeyFromObject(o), oo)).To(Succeed())
			Expect(o.Status.GetInvalidCondition()).To(BeNil())
			Expect(o.Status.GetUnauthenticatedCondition()).To(BeTrueDueTo(AuthSecretGetFailed))
			Expect(o.Status.GetStaleCondition()).To(BeUnknownDueTo(Unauthenticated))
		})
	})

	When("pat key is not provided", func() {
		var k client.WithWatch
		BeforeEach(func(ctx context.Context) {
			s := &corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: strings.RandomHash(7), Namespace: namespace}}
			o := &GitHubRepository{
				ObjectMeta: metav1.ObjectMeta{Name: repoName, Namespace: namespace},
				Spec: GitHubRepositorySpec{
					Auth: GitHubRepositoryAuth{PersonalAccessToken: &GitHubRepositoryAuthPersonalAccessToken{
						Secret: SecretReferenceWithOptionalNamespace{Name: s.Name},
					}},
				},
			}
			k = fake.NewClientBuilder().WithScheme(scheme).WithObjects(s, o).WithStatusSubresource(o).Build()
		})
		It("should set conditions and abort", func(ctx context.Context) {
			o := &GitHubRepository{}
			Expect(k.Get(ctx, client.ObjectKeyFromObject(o), o)).To(Succeed())
			result, err := act.NewConnectToGitHubAction(0, nil).Execute(ctx, k, o)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal(&ctrl.Result{}))

			oo := &GitHubRepository{}
			Expect(k.Get(ctx, client.ObjectKeyFromObject(o), oo)).To(Succeed())
			Expect(o.Status.GetInvalidCondition()).To(BeTrueDueTo(AuthSecretKeyMissing))
			Expect(o.Status.GetUnauthenticatedCondition()).To(BeTrueDueTo(Invalid))
			Expect(o.Status.GetStaleCondition()).To(BeUnknownDueTo(Unauthenticated))
		})
	})

	When("pat key is not found in secret", func() {
		const refreshInterval = 5 * time.Minute
		var k client.WithWatch
		BeforeEach(func(ctx context.Context) {
			s := &corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: strings.RandomHash(7), Namespace: namespace}}
			o := &GitHubRepository{
				ObjectMeta: metav1.ObjectMeta{Name: repoName, Namespace: namespace},
				Spec: GitHubRepositorySpec{
					Auth: GitHubRepositoryAuth{PersonalAccessToken: &GitHubRepositoryAuthPersonalAccessToken{
						Secret: SecretReferenceWithOptionalNamespace{Name: s.Name},
						Key:    strings.RandomHash(7),
					}},
				},
			}
			k = fake.NewClientBuilder().WithScheme(scheme).WithObjects(s, o).WithStatusSubresource(o).Build()
		})
		It("should set conditions and abort", func(ctx context.Context) {
			o := &GitHubRepository{}
			Expect(k.Get(ctx, client.ObjectKeyFromObject(o), o)).To(Succeed())
			result, err := act.NewConnectToGitHubAction(refreshInterval, nil).Execute(ctx, k, o)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal(&ctrl.Result{RequeueAfter: refreshInterval}))

			oo := &GitHubRepository{}
			Expect(k.Get(ctx, client.ObjectKeyFromObject(o), oo)).To(Succeed())
			Expect(o.Status.GetInvalidCondition()).To(BeNil())
			Expect(o.Status.GetUnauthenticatedCondition()).To(BeTrueDueTo(AuthSecretKeyNotFound))
			Expect(o.Status.GetStaleCondition()).To(BeUnknownDueTo(Unauthenticated))
		})
	})

	When("pat from secret is empty", func() {
		const refreshInterval = 5 * time.Minute
		var k client.WithWatch
		BeforeEach(func(ctx context.Context) {
			s := &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{Name: strings.RandomHash(7), Namespace: namespace},
				Data:       map[string][]byte{"k": []byte("")},
			}
			o := &GitHubRepository{
				ObjectMeta: metav1.ObjectMeta{Name: repoName, Namespace: namespace},
				Spec: GitHubRepositorySpec{
					Auth: GitHubRepositoryAuth{PersonalAccessToken: &GitHubRepositoryAuthPersonalAccessToken{
						Secret: SecretReferenceWithOptionalNamespace{Name: s.Name},
						Key:    "k",
					}},
				},
			}
			k = fake.NewClientBuilder().WithScheme(scheme).WithObjects(s, o).WithStatusSubresource(o).Build()
		})
		It("should set conditions and abort", func(ctx context.Context) {
			o := &GitHubRepository{}
			Expect(k.Get(ctx, client.ObjectKeyFromObject(o), o)).To(Succeed())
			result, err := act.NewConnectToGitHubAction(refreshInterval, nil).Execute(ctx, k, o)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal(&ctrl.Result{RequeueAfter: refreshInterval}))

			oo := &GitHubRepository{}
			Expect(k.Get(ctx, client.ObjectKeyFromObject(o), oo)).To(Succeed())
			Expect(o.Status.GetInvalidCondition()).To(BeNil())
			Expect(o.Status.GetUnauthenticatedCondition()).To(BeTrueDueTo(AuthTokenEmpty))
			Expect(o.Status.GetStaleCondition()).To(BeUnknownDueTo(Unauthenticated))
		})
	})

	When("pat from secret is invalid", func() {
		const refreshInterval = 5 * time.Minute
		var k client.WithWatch
		BeforeEach(func(ctx context.Context) {
			s := &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{Name: strings.RandomHash(7), Namespace: namespace},
				Data:       map[string][]byte{"k": []byte(strings.RandomHash(7))},
			}
			o := &GitHubRepository{
				ObjectMeta: metav1.ObjectMeta{Name: repoName, Namespace: namespace},
				Spec: GitHubRepositorySpec{
					Auth: GitHubRepositoryAuth{PersonalAccessToken: &GitHubRepositoryAuthPersonalAccessToken{
						Secret: SecretReferenceWithOptionalNamespace{Name: s.Name},
						Key:    "k",
					}},
				},
			}
			k = fake.NewClientBuilder().WithScheme(scheme).WithObjects(s, o).WithStatusSubresource(o).Build()
		})
		It("should set conditions and abort", func(ctx context.Context) {
			o := &GitHubRepository{}
			Expect(k.Get(ctx, client.ObjectKeyFromObject(o), o)).To(Succeed())
			result, err := act.NewConnectToGitHubAction(refreshInterval, nil).Execute(ctx, k, o)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal(&ctrl.Result{RequeueAfter: refreshInterval}))

			oo := &GitHubRepository{}
			Expect(k.Get(ctx, client.ObjectKeyFromObject(o), oo)).To(Succeed())
			Expect(o.Status.GetInvalidCondition()).To(BeNil())
			Expect(o.Status.GetUnauthenticatedCondition()).To(BeTrueDueTo(TokenValidationFailed))
			Expect(o.Status.GetStaleCondition()).To(BeUnknownDueTo(Unauthenticated))
		})
	})

	When("pat from secret is valid", func() {
		var k client.WithWatch
		var gh *github.Client
		BeforeEach(func(ctx context.Context) {
			s := &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{Name: strings.RandomHash(7), Namespace: namespace},
				Data:       map[string][]byte{"k": []byte(os.Getenv("GITHUB_TOKEN"))},
			}
			o := &GitHubRepository{
				ObjectMeta: metav1.ObjectMeta{Name: repoName, Namespace: namespace},
				Spec: GitHubRepositorySpec{
					Auth: GitHubRepositoryAuth{PersonalAccessToken: &GitHubRepositoryAuthPersonalAccessToken{
						Secret: SecretReferenceWithOptionalNamespace{Name: s.Name},
						Key:    "k",
					}},
				},
			}
			k = fake.NewClientBuilder().WithScheme(scheme).WithObjects(s, o).WithStatusSubresource(o).Build()
		})
		It("should set conditions and abort", func(ctx context.Context) {
			o := &GitHubRepository{}
			Expect(k.Get(ctx, client.ObjectKeyFromObject(o), o)).To(Succeed())
			result, err := act.NewConnectToGitHubAction(0, &gh).Execute(ctx, k, o)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(BeNil())
			Expect(gh).ToNot(BeNil())

			oo := &GitHubRepository{}
			Expect(k.Get(ctx, client.ObjectKeyFromObject(o), oo)).To(Succeed())
			Expect(o.Status.GetInvalidCondition()).To(BeNil())
			Expect(o.Status.GetUnauthenticatedCondition()).To(BeNil())
			Expect(o.Status.GetStaleCondition()).To(BeNil())
		})
	})
})
