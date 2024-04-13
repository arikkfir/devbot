package justest_test

import (
	"github.com/arikkfir/devbot/backend/internal/util/testing/justest"
	"testing"
)

func TestGetHelper(t *testing.T) {
	tt := justest.NewTT(t)
	if h := justest.GetHelper(tt); h != t {
		t.Fatalf("Expected GetHelper(tt) to return original t, but it did not: %+v", h)
	}

	ttt := justest.NewTT(tt)
	if h := justest.GetHelper(ttt); h != t {
		t.Fatalf("Expected GetHelper(tt) to return original t, but it did not: %+v", h)
	}
}
