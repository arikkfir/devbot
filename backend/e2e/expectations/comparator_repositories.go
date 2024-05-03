package expectations

import (
	apiv1 "github.com/arikkfir/devbot/backend/api/v1"
	. "github.com/arikkfir/justest"
)

func RepositoriesComparator(t T, e, a any) {
	expected := e.([]RepositoryE)
	actual := a.([]apiv1.Repository)
	With(t).Verify(len(actual)).Will(EqualTo(len(expected))).OrFail()
	for _, e := range expected {
		found := false
		for _, a := range actual {
			if a.Name == e.Name {
				With(t).Verify(a).Will(EqualTo(e).Using(RepositoryComparator)).OrFail()
				found = true
				break
			}
		}
		With(t).Verify(found).Will(EqualTo(true)).OrFail()
	}
}

func RepositoryComparator(t T, e, a any) {
	expected := e.(RepositoryE)
	var actual *apiv1.Repository
	if rp, ok := a.(*apiv1.Repository); ok {
		actual = rp
	} else {
		v := a.(apiv1.Repository)
		actual = &v
	}
	With(t).Verify(actual.Status.Conditions).Will(EqualTo(expected.Status.Conditions).Using(ConditionsComparator)).OrFail()
	With(t).Verify(actual.Status.DefaultBranch).Will(EqualTo(expected.Status.DefaultBranch)).OrFail()
	With(t).Verify(actual.Status.Revisions).Will(EqualTo(expected.Status.Revisions)).OrFail()
}
