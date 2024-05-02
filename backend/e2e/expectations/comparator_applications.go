package expectations

import (
	"context"
	apiv1 "github.com/arikkfir/devbot/backend/api/v1"
	"github.com/arikkfir/devbot/backend/internal/util/k8s"
	. "github.com/arikkfir/devbot/backend/internal/util/testing/justest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func CreateApplicationsComparator(c client.Client, ctx context.Context) Comparator {
	return func(t T, e, a any) {
		expected := e.([]AppE)
		actual := a.([]apiv1.Application)
		With(t).Verify(len(actual)).Will(EqualTo(len(expected))).OrFail()
		for _, e := range expected {
			found := false
			for _, a := range actual {
				if a.Name == e.Name {
					With(t).Verify(a).Will(EqualTo(e).Using(CreateApplicationComparator(c, ctx))).OrFail()
					found = true
					break
				}
			}
			With(t).Verify(found).Will(EqualTo(true)).OrFail()
		}
	}
}

func CreateApplicationComparator(c client.Client, ctx context.Context) Comparator {
	return func(t T, e, a any) {
		appE := e.(AppE)
		actualApp := a.(apiv1.Application)
		With(t).Verify(actualApp.Status.Conditions).Will(EqualTo(appE.Status.Conditions).Using(ConditionsComparator)).OrFail()

		envList := &apiv1.EnvironmentList{}
		With(t).Verify(c.List(ctx, envList, client.InNamespace(actualApp.Namespace))).Will(Succeed()).OrFail()
		With(t).Verify(envList.Items).Will(EqualTo(appE.Environments).Using(CreateEnvironmentsComparator(c, ctx))).OrFail()
	}
}

func CreateEnvironmentsComparator(c client.Client, ctx context.Context) Comparator {
	return func(t T, e any, a any) {
		expected := e.([]EnvE)
		actual := a.([]apiv1.Environment)
		With(t).Verify(len(actual)).Will(EqualTo(len(expected))).OrFail()
		for _, e := range expected {
			found := false
			for _, a := range actual {
				if a.Spec.PreferredBranch == e.Spec.PreferredBranch {
					With(t).Verify(a).Will(EqualTo(e).Using(CreateEnvironmentComparator(c, ctx))).OrFail()
					found = true
					break
				}
			}
			With(t).Verify(found).Will(EqualTo(true)).OrFail()
		}
	}
}

func CreateEnvironmentComparator(c client.Client, ctx context.Context) Comparator {
	return func(t T, e, a any) {
		expected := e.(EnvE)
		actual := a.(apiv1.Environment)
		With(t).Verify(actual.Spec.PreferredBranch).Will(EqualTo(expected.Spec.PreferredBranch)).OrFail()
		With(t).Verify(actual.Status.Conditions).Will(EqualTo(expected.Status.Conditions).Using(ConditionsComparator)).OrFail()

		nsFilter := client.InNamespace(actual.Namespace)
		ownershipFilter := k8s.OwnedBy(c.Scheme(), &actual)

		deploymentsList := &apiv1.DeploymentList{}
		With(t).Verify(c.List(ctx, deploymentsList, nsFilter, ownershipFilter)).Will(Succeed()).OrFail()
		With(t).Verify(deploymentsList.Items).Will(EqualTo(expected.Deployments).Using(CreateDeploymentsComparator(c, ctx))).OrFail()
	}
}

func CreateDeploymentsComparator(c client.Client, ctx context.Context) Comparator {
	return func(t T, e, a any) {
		expected := e.([]DeploymentE)
		actual := a.([]apiv1.Deployment)
		With(t).Verify(len(actual)).Will(EqualTo(len(expected))).OrFail()
		for _, e := range expected {
			found := false
			for _, a := range actual {
				if a.Spec.Repository == e.Spec.Repository {
					With(t).Verify(a).Will(EqualTo(e).Using(CreateDeploymentComparator(c, ctx))).OrFail()
					found = true
					break
				}
			}
			With(t).Verify(found).Will(EqualTo(true)).OrFail()
		}
	}
}

func CreateDeploymentComparator(c client.Client, ctx context.Context) Comparator {
	return func(t T, e, a any) {
		expected := e.(DeploymentE)
		actual := a.(apiv1.Deployment)
		With(t).Verify(actual.Spec.Repository).Will(EqualTo(expected.Spec.Repository)).OrFail()
		With(t).Verify(actual.Status.Branch).Will(EqualTo(expected.Status.Branch)).OrFail()
		With(t).Verify(actual.Status.Conditions).Will(EqualTo(expected.Status.Conditions).Using(ConditionsComparator)).OrFail()
		With(t).Verify(actual.Status.LastAttemptedRevision).Will(EqualTo(expected.Status.LastAttemptedRevision)).OrFail()
		With(t).Verify(actual.Status.LastAppliedRevision).Will(EqualTo(expected.Status.LastAppliedRevision)).OrFail()
		With(t).Verify(actual.Status.ResolvedRepository).Will(EqualTo(expected.Status.ResolvedRepository)).OrFail()

		for _, resourceExp := range expected.Resources {
			With(t).Verify(c.Get(ctx, client.ObjectKey{Namespace: resourceExp.Namespace, Name: resourceExp.Name}, resourceExp.Object)).Will(Succeed()).OrFail()
			if resourceExp.Validator != nil {
				resourceExp.Validator(t, resourceExp)
			}
		}
	}
}
