package justest_test

import (
	. "github.com/arikkfir/devbot/backend/internal/util/testing/justest"
	"testing"
)

func TestBeEqualTo(t *testing.T) {
	type T struct {
		I int
	}

	cases := []struct {
		name                string
		positiveExpectation bool
		actuals             []any
		expected            []any
		wantErr             bool
	}{
		{
			name:                "1WillBeEqualTo1_Matches",
			positiveExpectation: true,
			actuals:             []any{1},
			expected:            []any{1},
			wantErr:             false,
		},
		{
			name:                "1WillNotBeEqualTo1_Mismatches",
			positiveExpectation: false,
			actuals:             []any{1},
			expected:            []any{1},
			wantErr:             true,
		},
		{
			name:                "1WillBeEqualTo2_Mismatches",
			positiveExpectation: true,
			actuals:             []any{1},
			expected:            []any{2},
			wantErr:             true,
		},
		{
			name:                "1WillNotBeEqualTo2_Matches",
			positiveExpectation: false,
			actuals:             []any{1},
			expected:            []any{2},
			wantErr:             false,
		},
		{
			name:                "123ValuesWillBeEqualTo123_Matches",
			positiveExpectation: true,
			actuals:             []any{1, 2, 3},
			expected:            []any{1, 2, 3},
			wantErr:             false,
		},
		{
			name:                "123WillNotBeEqualTo123_Mismatches",
			positiveExpectation: false,
			actuals:             []any{1, 2, 3},
			expected:            []any{1, 2, 3},
			wantErr:             true,
		},
		{
			name:                "123WillBeEqualTo124_Mismatches",
			positiveExpectation: true,
			actuals:             []any{1, 2, 3},
			expected:            []any{1, 2, 4},
			wantErr:             true,
		},
		{
			name:                "123WillNotBeEqualTo124_Matches",
			positiveExpectation: false,
			actuals:             []any{1, 2, 3},
			expected:            []any{1, 2, 4},
			wantErr:             false,
		},
		{
			name:                "Struct1WillBeEqualToStruct1_Matches",
			positiveExpectation: true,
			actuals:             []any{T{I: 1}},
			expected:            []any{T{I: 1}},
			wantErr:             false,
		},
		{
			name:                "Struct1WillNotBeEqualToStruct1_Mismatches",
			positiveExpectation: false,
			actuals:             []any{T{I: 1}},
			expected:            []any{T{I: 1}},
			wantErr:             true,
		},
		{
			name:                "Struct1WillBeEqualToStruct2_Mismatches",
			positiveExpectation: true,
			actuals:             []any{T{I: 1}},
			expected:            []any{T{I: 2}},
			wantErr:             true,
		},
		{
			name:                "Struct1WillNotBeEqualToStruct2_Matches",
			positiveExpectation: false,
			actuals:             []any{T{I: 1}},
			expected:            []any{T{I: 2}},
			wantErr:             false,
		},
		{
			name:                "StructPtrWillBeEqualToStructPtr_Matches",
			positiveExpectation: true,
			actuals:             []any{&T{I: 1}},
			expected:            []any{&T{I: 1}},
			wantErr:             false,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mt := &MockT{parent: t}
			defer verifyTestCaseError(t, mt, tc.wantErr)
			if tc.positiveExpectation {
				For(mt).Expect(tc.actuals...).Will(BeEqualTo(tc.expected...))
			} else {
				For(mt).Expect(tc.actuals...).WillNot(BeEqualTo(tc.expected...))
			}
		})
	}
}
