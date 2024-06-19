package util

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/controller-runtime/pkg/client"

	apiv1 "github.com/arikkfir/devbot/api/v1"
	"github.com/arikkfir/devbot/internal/util/strings"
)

func CreateK8sRepository(ctx context.Context, c client.Client, ns string, spec apiv1.RepositorySpec) string {
	GinkgoHelper()
	ghOwner := spec.GitHub.Owner
	ghName := spec.GitHub.Name
	kName := fmt.Sprintf("%s-%s-%s", ghOwner, ghName, strings.RandomHash(7))
	repo := &apiv1.Repository{ObjectMeta: metav1.ObjectMeta{Namespace: ns, Name: kName}, Spec: spec}
	Expect(c.Create(ctx, repo)).To(Succeed())
	DeferCleanup(func(ctx context.Context) { Expect(c.Delete(ctx, repo)).To(Succeed()) })
	return repo.Name
}
