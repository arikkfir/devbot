package justest_test

import (
	. "github.com/arikkfir/devbot/backend/internal/util/testing/justest"
	"testing"
)

func TestWillAssertion(t *testing.T) {
	t.Run("FailsOnMismatch", func(t *testing.T) {
		mt := &MockT{parent: t}
		defer verifyTestCaseError(t, mt, true)
		For(mt).Expect(1).Will(BeEqualTo(2))
	})
	t.Run("DoesNotFailOnMatch", func(t *testing.T) {
		mt := &MockT{parent: t}
		defer verifyTestCaseError(t, mt, false)
		For(mt).Expect(1).Will(BeEqualTo(1))
	})
}

func TestWillNotAssertion(t *testing.T) {
	t.Run("FailsOnMatch", func(t *testing.T) {
		mt := &MockT{parent: t}
		defer verifyTestCaseError(t, mt, true)
		For(mt).Expect(1).WillNot(BeEqualTo(1))
	})
	t.Run("DoesNotFailOnMismatch", func(t *testing.T) {
		mt := &MockT{parent: t}
		defer verifyTestCaseError(t, mt, false)
		For(mt).Expect(1).WillNot(BeEqualTo(2))
	})
}
