package justest_test

import (
	. "github.com/arikkfir/devbot/backend/internal/util/testing/justest"
	"testing"
)

func TestBeNil(t *testing.T) {
	cases := []struct {
		name                string
		positiveExpectation bool
		actuals             []any
		wantErr             bool
	}{
		{
			name:                "NilWillBeNil_Matches",
			positiveExpectation: true,
			actuals:             []any{nil},
			wantErr:             false,
		},
		{
			name:                "NilWillNotBeNil_Mismatches",
			positiveExpectation: false,
			actuals:             []any{nil},
			wantErr:             true,
		},
		{
			name:                "NonNilWillBeNil_Mismatches",
			positiveExpectation: true,
			actuals:             []any{"abc"},
			wantErr:             true,
		},
		{
			name:                "NonNilWillNotBeNil_Matches",
			positiveExpectation: false,
			actuals:             []any{"abc"},
			wantErr:             false,
		},
		{
			name:                "NoActualsWillBeNil_Matches",
			positiveExpectation: true,
			actuals:             []any{},
			wantErr:             false,
		},
		{
			name:                "NoActualsWillNotBeNil_Matches",
			positiveExpectation: false,
			actuals:             []any{},
			wantErr:             false,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mt := &MockT{parent: t}
			defer verifyTestCaseError(t, mt, tc.wantErr)
			if tc.positiveExpectation {
				For(mt).Expect(tc.actuals...).Will(BeNil())
			} else {
				For(mt).Expect(tc.actuals...).WillNot(BeNil())
			}
		})
	}
}
