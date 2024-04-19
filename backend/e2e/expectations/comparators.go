package expectations

import (
	apiv1 "github.com/arikkfir/devbot/backend/api/v1"
	"github.com/arikkfir/devbot/backend/internal/util/k8s"
	. "github.com/arikkfir/devbot/backend/internal/util/testing"
	. "github.com/arikkfir/devbot/backend/internal/util/testing/justest"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func ApplicationsComparator(t TT, e, a any) any {
	expected := e.(map[string]AppE)
	actual := a.([]apiv1.Application)
	For(t).Expect(len(actual)).Will(BeEqualTo(len(expected)))
	for _, e := range expected {
		found := false
		for _, a := range actual {
			if a.Name == e.Name {
				For(t).Expect(a).Will(CompareTo(e).Using(ApplicationComparator))
				found = true
				break
			}
		}
		For(t).Expect(found).Will(BeEqualTo(true))
	}
	return a
}

func ApplicationComparator(t TT, e, a any) any {
	appE := e.(AppE)
	actualApp := a.(apiv1.Application)
	For(t).Expect(actualApp.Status.Conditions).Will(CompareTo(appE.Status.Conditions).Using(ConditionsComparator))

	envList := &apiv1.EnvironmentList{}
	For(t).Expect(K(t).Client.List(t, envList, client.InNamespace(actualApp.Namespace))).Will(Succeed())
	For(t).Expect(envList.Items).Will(CompareTo(appE.Environments).Using(EnvironmentsComparator))
	return a
}

func EnvironmentsComparator(t TT, e, a any) any {
	expected := e.([]EnvE)
	actual := a.([]apiv1.Environment)
	For(t).Expect(len(actual)).Will(BeEqualTo(len(expected)))
	for _, e := range expected {
		found := false
		for _, a := range actual {
			if a.Spec.PreferredBranch == e.Spec.PreferredBranch {
				For(t).Expect(a).Will(CompareTo(e).Using(EnvironmentComparator))
				found = true
				break
			}
		}
		For(t).Expect(found).Will(BeEqualTo(true))
	}
	return a
}

func EnvironmentComparator(t TT, e, a any) any {
	expected := e.(EnvE)
	actual := a.(apiv1.Environment)
	For(t).Expect(actual.Spec.PreferredBranch).Will(BeEqualTo(expected.Spec.PreferredBranch))
	For(t).Expect(actual.Status.Conditions).Will(CompareTo(expected.Status.Conditions).Using(ConditionsComparator))

	nsFilter := client.InNamespace(actual.Namespace)
	ownershipFilter := k8s.OwnedBy(K(t).Client.Scheme(), &actual)

	deploymentsList := &apiv1.DeploymentList{}
	For(t).Expect(K(t).Client.List(t, deploymentsList, nsFilter, ownershipFilter)).Will(Succeed())
	For(t).Expect(deploymentsList.Items).Will(CompareTo(expected.Deployments).Using(DeploymentsComparator))
	return a
}

func DeploymentsComparator(t TT, e, a any) any {
	expected := e.([]DeploymentE)
	actual := a.([]apiv1.Deployment)
	For(t).Expect(len(actual)).Will(BeEqualTo(len(expected)))
	for _, e := range expected {
		found := false
		for _, a := range actual {
			if a.Spec.Repository == e.Spec.Repository {
				For(t).Expect(a).Will(CompareTo(e).Using(DeploymentComparator))
				found = true
				break
			}
		}
		For(t).Expect(found).Will(BeEqualTo(true))
	}
	return a
}

func DeploymentComparator(t TT, e, a any) any {
	expected := e.(DeploymentE)
	actual := a.(apiv1.Deployment)
	For(t).Expect(actual.Spec.Repository).Will(BeEqualTo(expected.Spec.Repository))
	For(t).Expect(actual.Status.Branch).Will(BeEqualTo(expected.Status.Branch))
	For(t).Expect(actual.Status.Conditions).Will(CompareTo(expected.Status.Conditions).Using(ConditionsComparator))
	For(t).Expect(actual.Status.LastAttemptedRevision).Will(BeEqualTo(expected.Status.LastAttemptedRevision))
	For(t).Expect(actual.Status.LastAppliedRevision).Will(BeEqualTo(expected.Status.LastAppliedRevision))
	For(t).Expect(actual.Status.ResolvedRepository).Will(BeEqualTo(expected.Status.ResolvedRepository))

	for _, resourceExp := range expected.Resources {
		For(t).Expect(K(t).Client.Get(t, client.ObjectKey{Namespace: resourceExp.Namespace, Name: resourceExp.Name}, resourceExp.Object)).Will(Succeed())
		if resourceExp.Validator != nil {
			resourceExp.Validator(t, resourceExp)
		}
	}
	return a
}

func RepositoriesComparator(t TT, e, a any) any {
	expected := e.([]RepositoryE)
	actual := a.([]apiv1.Repository)
	For(t).Expect(len(actual)).Will(BeEqualTo(len(expected))).OrFail()
	for _, e := range expected {
		found := false
		for _, a := range actual {
			if a.Name == e.Name {
				For(t).Expect(a).Will(CompareTo(e).Using(RepositoryComparator)).OrFail()
				found = true
				break
			}
		}
		For(t).Expect(found).Will(BeEqualTo(true)).OrFail()
	}
	return a
}

func RepositoryComparator(t TT, e, a any) any {
	expected := e.(RepositoryE)
	actual := a.(apiv1.Repository)
	For(t).Expect(actual.Status.Conditions).Will(CompareTo(expected.Status.Conditions).Using(ConditionsComparator)).OrFail()
	For(t).Expect(actual.Status.DefaultBranch).Will(BeEqualTo(expected.Status.DefaultBranch)).OrFail()
	For(t).Expect(actual.Status.Revisions).Will(BeEqualTo(expected.Status.Revisions)).OrFail()
	return a
}

func ConditionsComparator(t TT, e, a any) any {
	expectedConditions := e.(map[string]*ConditionE)
	actualConditions := append([]metav1.Condition{}, a.([]metav1.Condition)...) // cloning so we can chip away at it

	for expectedConditionType, expectedConditionProperties := range expectedConditions {
		found := false
		for i, actualCondition := range actualConditions {
			if actualCondition.Type == expectedConditionType {
				found = true
				For(t).Expect(&actualCondition).Will(CompareTo(expectedConditionProperties).Using(ConditionComparator)).OrFail()
				actualConditions = append(actualConditions[:i], actualConditions[i+1:]...)
				break
			}
		}
		if expectedConditionProperties != nil {
			For(t).Expect(found).Will(BeEqualTo(true)).Because("Condition '%s' was not found", expectedConditionType).OrFail()
		} else {
			For(t).Expect(found).Will(BeEqualTo(false)).Because("Condition '%s' was found", expectedConditionType).OrFail()
		}
	}
	For(t).Expect(len(actualConditions)).Will(BeEqualTo(0)).OrFail()
	return a
}

func ConditionComparator(t TT, e, a any) any {
	expected := e.(*ConditionE)
	actual := a.(*metav1.Condition)

	if actual == nil && expected == nil {
		return actual
	} else if actual == nil && expected != nil {
		t.Fatalf("Expected condition '%s' to exist, but it does not", expected.Type)
	} else if actual != nil && expected == nil {
		t.Fatalf("Expected condition '%s' not to exist, but it does", actual.Type)
	}

	if expected.Status != nil {
		For(t).Expect(string(actual.Status)).Will(BeEqualTo(*expected.Status)).OrFail()
	}
	if expected.Reason != nil {
		For(t).Expect(actual.Reason).Will(Say(expected.Reason)).OrFail()
	}
	if expected.Message != nil {
		For(t).Expect(actual.Message).Will(Say(expected.Message)).Because("Condition message does not match '%s': %s", expected.Message.String(), actual.Message).OrFail()
	}
	return actual
}
