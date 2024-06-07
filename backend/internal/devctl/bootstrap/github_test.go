package bootstrap

import (
	"context"
	"testing"
	"time"

	. "github.com/arikkfir/justest"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"

	stringsutil "github.com/arikkfir/devbot/backend/internal/util/strings"
	. "github.com/arikkfir/devbot/backend/internal/util/testing"
)

func TestBootstrapper(t *testing.T) {
	verifyDevbotDeployment := features.New("devctl bootstrap github").
		Assess("pods from kube-system", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			gb, err := NewGitHubBootstrapper(ctx, GH(ctx, t).Token, cfg.Client().RESTConfig())
			With(t).Verify(err).Will(BeNil()).OrFail()
			With(t).Verify(gb.Bootstrap(ctx, GitHubOwner, stringsutil.Name(), "public")).Will(Succeed()).OrFail()
			With(t).Verify(cfg.Client().Resources().Get(ctx, "devbot", "", &corev1.Namespace{})).Will(Succeed()).OrFail()
			With(t).
				Verify(func(t T) {
					d := &v1.Deployment{}
					With(t).Verify(cfg.Client().Resources().Get(ctx, "devbot-controller", "devbot", d)).Will(Succeed()).OrFail()
					With(t).Verify(d.Status.Replicas).Will(EqualTo(1)).OrFail()
					With(t).Verify(d.Status.AvailableReplicas).Will(EqualTo(1)).OrFail()
					With(t).Verify(d.Status.ReadyReplicas).Will(EqualTo(1)).OrFail()
					With(t).Verify(d.Status.UnavailableReplicas).Will(EqualTo(0)).OrFail()
				}).
				Will(Succeed()).
				Within(1*time.Minute, 1*time.Second)
			return ctx
		}).Feature()

	testEnv.Test(t, verifyDevbotDeployment)
}
