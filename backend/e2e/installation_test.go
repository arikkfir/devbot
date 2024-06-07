package e2e_test

import (
	"testing"
	"time"

	. "github.com/arikkfir/justest"
	v1 "k8s.io/api/apps/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func TestInstallation(t *testing.T) {
	t.Parallel()
	e2e := NewE2E(t)

	deploymentControllerKey := client.ObjectKey{Namespace: "devbot", Name: "devbot-controller"}
	With(t).Verify(func(t T) {

		d := &v1.Deployment{}
		With(t).Verify(e2e.K.Client.Get(e2e.Ctx, deploymentControllerKey, d)).Will(Succeed()).OrFail()
		With(t).Verify(d.Status.Replicas).Will(EqualTo(1)).OrFail()
		With(t).Verify(d.Status.AvailableReplicas).Will(EqualTo(1)).OrFail()
		With(t).Verify(d.Status.ReadyReplicas).Will(EqualTo(1)).OrFail()
		With(t).Verify(d.Status.UnavailableReplicas).Will(EqualTo(0)).OrFail()

	}).Will(Succeed()).Within(1*time.Minute, 1*time.Second)
}
