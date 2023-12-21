package testing

import (
	"context"
	"encoding/json"
	apiv1 "github.com/arikkfir/devbot/backend/api/v1"
	"github.com/arikkfir/devbot/backend/internal/util/k8s"
	stringsutil "github.com/arikkfir/devbot/backend/internal/util/strings"
	"github.com/google/go-github/v56/github"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"testing"
	"time"
)

type ApplyOption[T any] func(T)

func SetBreakAnnotation(o client.Object, breakAnnName string, requeue bool, requeueAfter time.Duration) {
	bl := k8s.BreakAnnotation{
		Break: true,
		Result: &ctrl.Result{
			Requeue:      requeue,
			RequeueAfter: requeueAfter,
		},
	}

	b, err := json.Marshal(bl)
	Expect(err).ToNot(HaveOccurred())

	annotations := o.GetAnnotations()
	if annotations == nil {
		annotations = map[string]string{}
	}
	annotations[k8s.GetBreakAnnotationFullName(breakAnnName)] = string(b)
	o.SetAnnotations(annotations)
}

type ControllerTestHarness struct {
	Ctx context.Context
	*testing.T
	*runtime.Scheme
	KClient       client.Client
	GitHubOwner   string
	GHClient      *github.Client
	GitHubSecrets map[string]string
}

func (h *ControllerTestHarness) CreateK8sApplication(namespace string, repositories ...types.NamespacedName) *apiv1.Application {
	var repositoryReferences []corev1.ObjectReference
	for _, r := range repositories {
		repositoryReferences = append(repositoryReferences, corev1.ObjectReference{
			APIVersion: apiv1.GitHubRepositoryGVK.GroupVersion().String(),
			Kind:       apiv1.GitHubRepositoryGVK.Kind,
			Namespace:  r.Namespace,
			Name:       r.Name,
		})
	}

	objName := stringsutil.RandomHash(7)
	app := apiv1.Application{
		TypeMeta:   metav1.TypeMeta{APIVersion: apiv1.GitHubRepositoryGVK.GroupVersion().String(), Kind: apiv1.GitHubRepositoryGVK.Kind},
		ObjectMeta: metav1.ObjectMeta{Namespace: namespace, Name: objName},
		Spec: apiv1.ApplicationSpec{
			Strategy: apiv1.ApplicationSpecStrategy{
				Missing:       "Default",
				DefaultBranch: "main",
			},
			Repositories: repositoryReferences,
		},
	}
	if err := h.KClient.Create(h.Ctx, &app); err != nil {
		h.T.Fatalf("Failed to create Application: %+v", err)
	}

	h.T.Cleanup(func() {
		h.T.Logf("Deleting application '%s/%s'...", namespace, objName)
		if err := h.KClient.Delete(h.Ctx, &apiv1.Application{ObjectMeta: metav1.ObjectMeta{Namespace: namespace, Name: objName}}); err != nil {
			h.T.Errorf("Failed to delete application: %+v", err)
		}
	})
	return &app
}

func (h *ControllerTestHarness) GetK8sApplication(namespace, name string) *apiv1.Application {
	a := apiv1.Application{}
	if err := h.KClient.Get(h.Ctx, client.ObjectKey{Namespace: namespace, Name: name}, &a); err != nil {
		h.T.Fatalf("Failed to get Application: %+v", err)
	}
	return &a
}

func (h *ControllerTestHarness) DeleteK8sApplication(namespace, name string) {
	if err := h.KClient.Delete(h.Ctx, &apiv1.Application{ObjectMeta: metav1.ObjectMeta{Namespace: namespace, Name: name}}); err != nil {
		h.T.Fatalf("Failed to delete Application: %+v", err)
	}
}
