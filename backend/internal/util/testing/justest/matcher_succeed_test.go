package justest_test

import (
	"fmt"
	. "github.com/arikkfir/devbot/backend/internal/util/testing/justest"
	"testing"
)

func TestSucceed(t *testing.T) {
	cases := []struct {
		name                string
		positiveExpectation bool
		actuals             []any
		wantErr             bool
	}{
		{
			name:                "NoActualsWillSucceed_Matches",
			positiveExpectation: true,
			actuals:             []any{},
			wantErr:             false,
		},
		{
			name:                "NoActualsWillNotSucceed_Mismatches",
			positiveExpectation: false,
			actuals:             []any{},
			wantErr:             true,
		},
		{
			name:                "SingleNonErrorActualWillSucceed_Matches",
			positiveExpectation: true,
			actuals:             []any{"abc"},
			wantErr:             false,
		},
		{
			name:                "SingleNonErrorActualWillNotSucceed_Mismatches",
			positiveExpectation: false,
			actuals:             []any{"abc"},
			wantErr:             true,
		},
		{
			name:                "SingleNilErrorActualWillSucceed_Matches",
			positiveExpectation: true,
			actuals:             []any{error(nil)},
			wantErr:             false,
		},
		{
			name:                "SingleNilErrorActualWillNotSucceed_Mismatches",
			positiveExpectation: false,
			actuals:             []any{"abc"},
			wantErr:             true,
		},
		{
			name:                "SingleNonNilErrorActualWillSucceed_Mismatches",
			positiveExpectation: true,
			actuals:             []any{fmt.Errorf("failed")},
			wantErr:             true,
		},
		{
			name:                "SingleNonNilErrorActualWillNotSucceed_Matches",
			positiveExpectation: false,
			actuals:             []any{fmt.Errorf("failed")},
			wantErr:             false,
		},
		{
			name:                "MultipleActualsWithNilErrorWillSucceed_Matches",
			positiveExpectation: true,
			actuals:             []any{1, 2, error(nil)},
			wantErr:             false,
		},
		{
			name:                "MultipleActualsWithNilErrorWillNotSucceed_Mismatches",
			positiveExpectation: false,
			actuals:             []any{1, 2, error(nil)},
			wantErr:             true,
		},
		{
			name:                "MultipleActualsWithErrorWillSucceed_Mismatches",
			positiveExpectation: true,
			actuals:             []any{1, 2, fmt.Errorf("failed")},
			wantErr:             true,
		},
		{
			name:                "MultipleActualsWithErrorWillNotSucceed_Matches",
			positiveExpectation: false,
			actuals:             []any{1, 2, fmt.Errorf("failed")},
			wantErr:             false,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mt := &MockT{parent: t}
			defer verifyTestCaseError(t, mt, tc.wantErr)
			if tc.positiveExpectation {
				For(mt).Expect(tc.actuals...).Will(Succeed())
			} else {
				For(mt).Expect(tc.actuals...).WillNot(Succeed())
			}
		})
	}
}
