package justest_test

import (
	. "github.com/arikkfir/devbot/backend/internal/util/testing/justest"
	"testing"
)

func TestBeEmpty(t *testing.T) {
	cases := []struct {
		name                string
		positiveExpectation bool
		actuals             []any
		wantErr             bool
	}{
		{
			name:                "EmptyArrayWillBeEmpty_Matches",
			positiveExpectation: true,
			actuals:             []any{[0]int{}},
			wantErr:             false,
		},
		{
			name:                "EmptyArrayWillNotBeEmpty_Mismatches",
			positiveExpectation: false,
			actuals:             []any{[0]int{}},
			wantErr:             true,
		},
		{
			name:                "NonEmptyArrayWillBeEmpty_Mismatches",
			positiveExpectation: true,
			actuals:             []any{[2]int{100, 200}},
			wantErr:             true,
		},
		{
			name:                "NonEmptyArrayWillNotBeEmpty_Matches",
			positiveExpectation: false,
			actuals:             []any{[2]int{100, 200}},
			wantErr:             false,
		},
		{
			name:                "EmptyMapWillBeEmpty_Matches",
			positiveExpectation: true,
			actuals:             []any{map[string]int{}},
			wantErr:             false,
		},
		{
			name:                "EmptyMapWillNotBeEmpty_Mismatches",
			positiveExpectation: false,
			actuals:             []any{map[string]int{}},
			wantErr:             true,
		},
		{
			name:                "NonEmptyMapWillBeEmpty_Mismatches",
			positiveExpectation: true,
			actuals:             []any{map[string]int{"k": 100}},
			wantErr:             true,
		},
		{
			name:                "NonEmptyMapWillNotBeEmpty_Matches",
			positiveExpectation: false,
			actuals:             []any{map[string]int{"k": 100}},
			wantErr:             false,
		},
		{
			name:                "EmptyArrayPointerWillBeEmpty_Matches",
			positiveExpectation: true,
			actuals:             []any{&[0]int{}},
			wantErr:             false,
		},
		{
			name:                "EmptyArrayPointerWillNotBeEmpty_Mismatches",
			positiveExpectation: false,
			actuals:             []any{&[0]int{}},
			wantErr:             true,
		},
		{
			name:                "NonEmptyArrayPointerWillBeEmpty_Mismatches",
			positiveExpectation: true,
			actuals:             []any{&[2]int{100, 200}},
			wantErr:             true,
		},
		{
			name:                "NonEmptyArrayPointerWillNotBeEmpty_Matches",
			positiveExpectation: false,
			actuals:             []any{&[2]int{100, 200}},
			wantErr:             false,
		},
		{
			name:                "EmptySliceWillBeEmpty_Matches",
			positiveExpectation: true,
			actuals:             []any{[]int{}},
			wantErr:             false,
		},
		{
			name:                "EmptySliceWillNotBeEmpty_Mismatches",
			positiveExpectation: false,
			actuals:             []any{[]int{}},
			wantErr:             true,
		},
		{
			name:                "NonEmptySliceWillBeEmpty_Mismatches",
			positiveExpectation: true,
			actuals:             []any{[]int{100, 200}},
			wantErr:             true,
		},
		{
			name:                "NonEmptySliceWillNotBeEmpty_Matches",
			positiveExpectation: false,
			actuals:             []any{[]int{100, 200}},
			wantErr:             false,
		},
		{
			name:                "EmptyStringWillBeEmpty_Matches",
			positiveExpectation: true,
			actuals:             []any{""},
			wantErr:             false,
		},
		{
			name:                "EmptyStringWillNotBeEmpty_Mismatches",
			positiveExpectation: false,
			actuals:             []any{""},
			wantErr:             true,
		},
		{
			name:                "NonEmptyStringWillBeEmpty_Mismatches",
			positiveExpectation: true,
			actuals:             []any{"abc"},
			wantErr:             true,
		},
		{
			name:                "NonEmptyStringWillNotBeEmpty_Matches",
			positiveExpectation: false,
			actuals:             []any{"abc"},
			wantErr:             false,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mt := &MockT{parent: t}
			defer verifyTestCaseError(t, mt, tc.wantErr)
			if tc.positiveExpectation {
				For(mt).Expect(tc.actuals...).Will(BeEmpty())
			} else {
				For(mt).Expect(tc.actuals...).WillNot(BeEmpty())
			}
		})
	}
}
