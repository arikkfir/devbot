package justest_test

import (
	. "github.com/arikkfir/devbot/backend/internal/util/testing/justest"
	"testing"
)

func TestBeNil(t *testing.T) {
	t.Run("Nil", func(t *testing.T) {
		t.Parallel()
		mt := NewMockT(NewTT(t))
		defer expectNoFailure(t, mt)
		For(mt).Expect(nil).Will(BeNil()).OrFail()
	})
	t.Run("Not nil", func(t *testing.T) {
		t.Parallel()
		mt := NewMockT(NewTT(t))
		defer expectFailure(t, mt, `Expected actual to be nil, but it is not: abc`)
		For(mt).Expect("abc").Will(BeNil()).OrFail()
	})
}
