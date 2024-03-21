package justest_test

import (
	. "github.com/arikkfir/devbot/backend/internal/util/testing/justest"
	"testing"
)

func TestBeBetween(t *testing.T) {
	cases := []struct {
		name                string
		positiveExpectation bool
		actuals             []any
		min, max            any
		wantErr             bool
	}{
		{
			name:                "Int_1WillBeBetween0And2_Matches",
			positiveExpectation: true,
			actuals:             []any{1},
			min:                 0,
			max:                 2,
			wantErr:             false,
		},
		{
			name:                "Int_1WillBeBetween1And2_Matches",
			positiveExpectation: true,
			actuals:             []any{1},
			min:                 1,
			max:                 2,
			wantErr:             false,
		},
		{
			name:                "Int_1WillBeBetween0And1_Matches",
			positiveExpectation: true,
			actuals:             []any{1},
			min:                 0,
			max:                 1,
			wantErr:             false,
		},
		{
			name:                "Int_1WillBeBetween2And3_Matches",
			positiveExpectation: true,
			actuals:             []any{1},
			min:                 2,
			max:                 3,
			wantErr:             true,
		},
		{
			name:                "Int_5WillBeBetween2And3_Matches",
			positiveExpectation: true,
			actuals:             []any{5},
			min:                 2,
			max:                 3,
			wantErr:             true,
		},
		{
			name:                "Int_1WillNotBeBetween0And2_Mismatches",
			positiveExpectation: false,
			actuals:             []any{1},
			min:                 0,
			max:                 2,
			wantErr:             true,
		},
		{
			name:                "RequiresOneActual",
			positiveExpectation: true,
			actuals:             []any{1, 2},
			min:                 2,
			max:                 3,
			wantErr:             true,
		},
		{
			name:                "ActualTypeStringMismatchFails",
			positiveExpectation: true,
			actuals:             []any{"abc"},
			min:                 uint(0),
			max:                 int8(2),
			wantErr:             true,
		},
		{
			name:                "ActualTypeMismatchFails",
			positiveExpectation: true,
			actuals:             []any{"abc"},
			min:                 uint(0),
			max:                 int8(2),
			wantErr:             true,
		},
		{
			name:                "MinTypeMismatchFails",
			positiveExpectation: true,
			actuals:             []any{int8(1)},
			min:                 uint(0),
			max:                 int8(2),
			wantErr:             true,
		},
		{
			name:                "MaxTypeMismatchFails",
			positiveExpectation: true,
			actuals:             []any{uint(1)},
			min:                 uint(0),
			max:                 int8(2),
			wantErr:             true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mt := &MockT{parent: t}
			defer verifyTestCaseError(t, mt, tc.wantErr)
			if tc.positiveExpectation {
				For(mt).Expect(tc.actuals...).Will(BeBetween(tc.min, tc.max))
			} else {
				For(mt).Expect(tc.actuals...).WillNot(BeBetween(tc.min, tc.max))
			}
		})
	}
}
