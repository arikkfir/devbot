package justest_test

import (
	. "github.com/arikkfir/devbot/backend/internal/util/testing/justest"
	"testing"
)

func TestBeLessThan(t *testing.T) {
	cases := []struct {
		name                string
		positiveExpectation bool
		actuals             []any
		max                 any
		wantErr             bool
	}{
		{
			name:                "Int_1WillBeLessThan2_Matches",
			positiveExpectation: true,
			actuals:             []any{1},
			max:                 2,
			wantErr:             false,
		},
		{
			name:                "Int_1WillNotBeLessThan2_Mismatches",
			positiveExpectation: false,
			actuals:             []any{1},
			max:                 2,
			wantErr:             true,
		},
		{
			name:                "Int_1WillBeLessThan1_Mismatches",
			positiveExpectation: true,
			actuals:             []any{1},
			max:                 1,
			wantErr:             true,
		},
		{
			name:                "Int_1WillNotBeLessThan1_Matches",
			positiveExpectation: false,
			actuals:             []any{1},
			max:                 1,
			wantErr:             false,
		},
		{
			name:                "Int_2WillBeLessThan1_Mismatches",
			positiveExpectation: true,
			actuals:             []any{2},
			max:                 1,
			wantErr:             true,
		},
		{
			name:                "Int_2WillNotBeLessThan1_Matches",
			positiveExpectation: false,
			actuals:             []any{2},
			max:                 1,
			wantErr:             false,
		},
		{
			name:                "RequiresOneActual",
			positiveExpectation: true,
			actuals:             []any{1, 2},
			max:                 2,
			wantErr:             true,
		},
		{
			name:                "TypeMismatchFails",
			positiveExpectation: true,
			actuals:             []any{int8(1)},
			max:                 uint(0),
			wantErr:             true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mt := &MockT{parent: t}
			defer verifyTestCaseError(t, mt, tc.wantErr)
			if tc.positiveExpectation {
				For(mt).Expect(tc.actuals...).Will(BeLessThan(tc.max))
			} else {
				For(mt).Expect(tc.actuals...).WillNot(BeLessThan(tc.max))
			}
		})
	}
}
