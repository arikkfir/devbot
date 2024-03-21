package justest_test

import (
	. "github.com/arikkfir/devbot/backend/internal/util/testing/justest"
	"testing"
)

func TestBeGreaterThan(t *testing.T) {
	cases := []struct {
		name                string
		positiveExpectation bool
		actuals             []any
		min                 any
		wantErr             bool
	}{
		{
			name:                "Int_1WillBeGreaterThan0_Matches",
			positiveExpectation: true,
			actuals:             []any{1},
			min:                 0,
			wantErr:             false,
		},
		{
			name:                "Int_1WillNotBeGreaterThan0_Mismatches",
			positiveExpectation: false,
			actuals:             []any{1},
			min:                 0,
			wantErr:             true,
		},
		{
			name:                "Int_1WillBeGreaterThan1_Mismatches",
			positiveExpectation: true,
			actuals:             []any{1},
			min:                 1,
			wantErr:             true,
		},
		{
			name:                "Int_1WillNotBeGreaterThan1_Matches",
			positiveExpectation: false,
			actuals:             []any{1},
			min:                 1,
			wantErr:             false,
		},
		{
			name:                "Int_1WillBeGreaterThan2_Mismatches",
			positiveExpectation: true,
			actuals:             []any{1},
			min:                 2,
			wantErr:             true,
		},
		{
			name:                "Int_1WillNotBeGreaterThan2_Matches",
			positiveExpectation: false,
			actuals:             []any{1},
			min:                 2,
			wantErr:             false,
		},
		{
			name:                "RequiresOneActual",
			positiveExpectation: true,
			actuals:             []any{1, 2},
			min:                 2,
			wantErr:             true,
		},
		{
			name:                "TypeMismatchFails",
			positiveExpectation: true,
			actuals:             []any{int8(1)},
			min:                 uint(0),
			wantErr:             true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mt := &MockT{parent: t}
			defer verifyTestCaseError(t, mt, tc.wantErr)
			if tc.positiveExpectation {
				For(mt).Expect(tc.actuals...).Will(BeGreaterThan(tc.min))
			} else {
				For(mt).Expect(tc.actuals...).WillNot(BeGreaterThan(tc.min))
			}
		})
	}
}
