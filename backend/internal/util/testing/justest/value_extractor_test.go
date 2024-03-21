package justest_test

import (
	"context"
	"fmt"
	"github.com/arikkfir/devbot/backend/internal/util/lang"
	"github.com/arikkfir/devbot/backend/internal/util/testing/justest"
	"github.com/google/go-cmp/cmp"
	"reflect"
	"testing"
)

func TestValueExtractor_MustExtractSingleValue(t *testing.T) {
	testCases := map[string]struct {
		defaultExtractor justest.Extractor
		extractorsMap    map[reflect.Kind]justest.Extractor
		actuals          []any
		expected         any
		wantErr          bool
	}{
		"zero_actuals_fails_test": {
			defaultExtractor: justest.ExtractorUnsupported,
			actuals:          []any{},
			wantErr:          true,
		},
		"more_than_one_actual_fails_test": {
			defaultExtractor: justest.ExtractorUnsupported,
			actuals:          []any{"a", "b"},
			wantErr:          true,
		},
		"fail_on_value_not_found": {
			defaultExtractor: func(_ context.Context, t justest.JustT, v any) (any, bool) {
				return nil, false
			},
			actuals: []any{"a"},
			wantErr: true,
		},
		"returns_extracted_value": {
			defaultExtractor: func(_ context.Context, t justest.JustT, v any) (any, bool) {
				return "bar", true
			},
			actuals:  []any{"foo"},
			expected: "bar",
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			mt := &MockT{parent: t}
			defer verifyTestCaseError(t, mt, tc.wantErr)

			ve := justest.NewValueExtractor(tc.defaultExtractor)
			for k, v := range tc.extractorsMap {
				ve[k] = v
			}

			actual := ve.MustExtractSingleValue(context.Background(), mt, tc.actuals...)
			if !cmp.Equal(tc.expected, actual) {
				t.Fatalf("Unexpected value returned:\n%s", cmp.Diff(tc.expected, actual))
			}
		})
	}
}

func TestValueExtractor_ExtractValue(t *testing.T) {
	testCases := map[string]struct {
		defaultExtractor justest.Extractor
		extractorsMap    map[reflect.Kind]justest.Extractor
		actual           any
		expectedValue    any
		expectedFound    bool
		wantErr          bool
	}{
		"nil_actual_finds_nil_value": {
			defaultExtractor: justest.ExtractorUnsupported,
			actual:           nil,
			expectedValue:    nil,
			expectedFound:    true,
		},
		"returns_value_from_found_extractor_1": {
			defaultExtractor: justest.ExtractorUnsupported,
			extractorsMap: map[reflect.Kind]justest.Extractor{
				reflect.String: func(_ context.Context, t justest.JustT, v any) (any, bool) {
					return "foo: " + v.(string), true
				},
			},
			actual:        "bar",
			expectedValue: "foo: bar",
			expectedFound: true,
		},
		"returns_value_from_found_extractor_2": {
			defaultExtractor: justest.ExtractorUnsupported,
			extractorsMap: map[reflect.Kind]justest.Extractor{
				reflect.String: func(_ context.Context, t justest.JustT, v any) (any, bool) {
					return "not found", false
				},
			},
			actual:        "bar",
			expectedValue: "not found",
			expectedFound: false,
		},
		"returns_value_from_default_extractor": {
			defaultExtractor: func(_ context.Context, t justest.JustT, v any) (any, bool) {
				return "bar", true
			},
			actual:        "foo",
			expectedValue: "bar",
			expectedFound: true,
		},
		"fails_on_missing_default_extractor": {
			defaultExtractor: func(_ context.Context, t justest.JustT, v any) (any, bool) {
				return "bar", true
			},
			actual:        "foo",
			expectedValue: "bar",
			expectedFound: true,
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			mt := &MockT{parent: t}
			defer verifyTestCaseError(t, mt, tc.wantErr)

			ve := justest.NewValueExtractor(tc.defaultExtractor)
			for k, v := range tc.extractorsMap {
				ve[k] = v
			}

			actual, found := ve.ExtractValue(context.Background(), mt, tc.actual)
			if tc.expectedFound && !found {
				t.Fatalf("Expected value to be found, but it was not")
			} else if !tc.expectedFound && found {
				t.Fatalf("Expected value to not be found, but it was")
			} else if !cmp.Equal(tc.expectedValue, actual) {
				t.Fatalf("Unexpected value returned:\n%s", cmp.Diff(tc.expectedValue, actual))
			}
		})
	}
}

func TestValueExtractor_ExtractValues(t *testing.T) {
	testCases := map[string]struct {
		defaultExtractor justest.Extractor
		extractorsMap    map[reflect.Kind]justest.Extractor
		actuals          []any
		expected         []any
		wantErr          bool
	}{
		"empty_actuals_returns_empty_results": {
			defaultExtractor: justest.ExtractorUnsupported,
			actuals:          []any{},
			expected:         nil,
		},
		"multiple_actuals_extracted_in_order": {
			defaultExtractor: justest.ExtractorUnsupported,
			extractorsMap: map[reflect.Kind]justest.Extractor{
				reflect.Int: func(_ context.Context, t justest.JustT, v any) (any, bool) {
					return v.(int) + 1, true
				},
				reflect.String: func(_ context.Context, t justest.JustT, v any) (any, bool) {
					return "foo: " + v.(string), true
				},
			},
			actuals:  []any{1, "bar"},
			expected: []any{2, "foo: bar"},
		},
		"unresolved_values_omitted_from_results": {
			defaultExtractor: justest.ExtractorUnsupported,
			extractorsMap: map[reflect.Kind]justest.Extractor{
				reflect.Int: func(_ context.Context, t justest.JustT, v any) (any, bool) {
					return v.(int) + 1, true
				},
				reflect.String: func(_ context.Context, t justest.JustT, v any) (any, bool) {
					return nil, false
				},
			},
			actuals:  []any{1, "bar", 2},
			expected: []any{2, 3},
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			mt := &MockT{parent: t}
			defer verifyTestCaseError(t, mt, tc.wantErr)

			ve := justest.NewValueExtractor(tc.defaultExtractor)
			for k, v := range tc.extractorsMap {
				ve[k] = v
			}

			actuals := ve.ExtractValues(context.Background(), mt, tc.actuals...)
			if !cmp.Equal(tc.expected, actuals) {
				t.Fatalf("Unexpected values returned:\n%s", cmp.Diff(tc.expected, actuals))
			}
		})
	}
}

func TestNewChannelExtractor(t *testing.T) {
	testCases := map[string]struct {
		defaultExtractor justest.Extractor
		extractorsMap    map[reflect.Kind]justest.Extractor
		chanProvider     func() chan any
		recurse          bool
		expectedValue    any
		expectedFound    bool
		wantErr          bool
	}{
		"closed_channel_returns_nil_not_found": {
			defaultExtractor: justest.ExtractorUnsupported,
			chanProvider: func() chan any {
				ch := make(chan any, 1)
				close(ch)
				return ch
			},
		},
		"empty_channel_returns_nil_not_found": {
			defaultExtractor: justest.ExtractorUnsupported,
			chanProvider:     func() chan any { return make(chan any, 1) },
			expectedValue:    nil,
		},
		"recurse_extracts_value_from_channel_value_1": {
			defaultExtractor: justest.ExtractorUnsupported,
			extractorsMap: map[reflect.Kind]justest.Extractor{
				reflect.String: func(_ context.Context, t justest.JustT, v any) (any, bool) {
					return "foo: " + v.(string), true
				},
			},
			chanProvider: func() chan any {
				ch := make(chan any, 1)
				ch <- "bar"
				return ch
			},
			recurse:       true,
			expectedValue: "foo: bar",
			expectedFound: true,
		},
		"recurse_extracts_value_from_channel_value_2": {
			defaultExtractor: justest.ExtractorUnsupported,
			extractorsMap: map[reflect.Kind]justest.Extractor{
				reflect.String: func(_ context.Context, t justest.JustT, v any) (any, bool) {
					return nil, false
				},
			},
			chanProvider: func() chan any {
				ch := make(chan any, 1)
				ch <- "bar"
				return ch
			},
			recurse:       true,
			expectedValue: nil,
			expectedFound: false,
		},
		"recurse_extracts_value_from_channel_value_3": {
			defaultExtractor: justest.ExtractorUnsupported,
			extractorsMap: map[reflect.Kind]justest.Extractor{
				reflect.String: func(_ context.Context, t justest.JustT, v any) (any, bool) {
					t.Fatalf("Extractor fails")
					panic("unreachable")
				},
			},
			chanProvider: func() chan any {
				ch := make(chan any, 1)
				ch <- "bar"
				return ch
			},
			recurse: true,
			wantErr: true,
		},
		"no_recurse_returns_channel_value": {
			defaultExtractor: justest.ExtractorUnsupported,
			extractorsMap: map[reflect.Kind]justest.Extractor{
				reflect.String: func(_ context.Context, t justest.JustT, v any) (any, bool) {
					return "foo: " + v.(string), true
				},
			},
			chanProvider: func() chan any {
				ch := make(chan any, 1)
				ch <- "bar"
				return ch
			},
			expectedValue: "bar",
			expectedFound: true,
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			mt := &MockT{parent: t}
			defer verifyTestCaseError(t, mt, tc.wantErr)

			ve := justest.NewValueExtractor(tc.defaultExtractor)
			for k, v := range tc.extractorsMap {
				ve[k] = v
			}

			extractor := justest.NewChannelExtractor(ve, tc.recurse)
			actual, found := extractor(context.Background(), mt, tc.chanProvider())
			if tc.expectedFound && !found {
				t.Fatalf("Expected value to be found, but it was not")
			} else if !tc.expectedFound && found {
				t.Fatalf("Expected value to not be found, but it was")
			} else if !cmp.Equal(tc.expectedValue, actual) {
				t.Fatalf("Unexpected value returned:\n%s", cmp.Diff(tc.expectedValue, actual))
			}
		})
	}
}

func TestNewPointerExtractor(t *testing.T) {
	testCases := map[string]struct {
		defaultExtractor justest.Extractor
		extractorsMap    map[reflect.Kind]justest.Extractor
		actual           any
		recurse          bool
		expectedValue    any
		expectedFound    bool
		wantErr          bool
	}{
		"recurse_extracts_value_from_pointer_1": {
			defaultExtractor: justest.ExtractorUnsupported,
			extractorsMap: map[reflect.Kind]justest.Extractor{
				reflect.String: func(_ context.Context, t justest.JustT, v any) (any, bool) {
					return "foo: " + v.(string), true
				},
			},
			actual:        lang.Ptr[string]("bar"),
			recurse:       true,
			expectedValue: "foo: bar",
			expectedFound: true,
		},
		"recurse_extracts_value_from_pointer_2": {
			defaultExtractor: justest.ExtractorUnsupported,
			extractorsMap: map[reflect.Kind]justest.Extractor{
				reflect.String: func(_ context.Context, t justest.JustT, v any) (any, bool) {
					return nil, false
				},
			},
			actual:        lang.Ptr[string]("bar"),
			recurse:       true,
			expectedValue: nil,
			expectedFound: false,
		},
		"recurse_extracts_value_from_pointer_3": {
			defaultExtractor: justest.ExtractorUnsupported,
			extractorsMap: map[reflect.Kind]justest.Extractor{
				reflect.String: func(_ context.Context, t justest.JustT, v any) (any, bool) {
					t.Fatalf("Extractor fails")
					panic("unreachable")
				},
			},
			actual:  lang.Ptr[string]("bar"),
			recurse: true,
			wantErr: true,
		},
		"no_recurse_returns_elem_value": {
			defaultExtractor: justest.ExtractorUnsupported,
			actual:           lang.Ptr[string]("bar"),
			expectedValue:    "bar",
			expectedFound:    true,
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			mt := &MockT{parent: t}
			defer verifyTestCaseError(t, mt, tc.wantErr)

			ve := justest.NewValueExtractor(tc.defaultExtractor)
			for k, v := range tc.extractorsMap {
				ve[k] = v
			}

			extractor := justest.NewPointerExtractor(ve, tc.recurse)
			actual, found := extractor(context.Background(), mt, tc.actual)
			if tc.expectedFound && !found {
				t.Fatalf("Expected value to be found, but it was not")
			} else if !tc.expectedFound && found {
				t.Fatalf("Expected value to not be found, but it was")
			} else if !cmp.Equal(tc.expectedValue, actual) {
				t.Fatalf("Unexpected value returned:\n%s", cmp.Diff(tc.expectedValue, actual))
			}
		})
	}
}

func TestNewFuncExtractor(t *testing.T) {
	type testCase struct {
		defaultExtractor justest.Extractor
		extractorsMap    map[reflect.Kind]justest.Extractor
		actualProvider   func(*testCase) any
		recurse          bool
		expectedValue    any
		expectedFound    bool
		called           bool
		wantCalled       bool
		wantErr          bool
	}
	testCases := map[string]testCase{
		"no_in_no_out_returns_nil_not_found": {
			defaultExtractor: justest.ExtractorUnsupported,
			extractorsMap: map[reflect.Kind]justest.Extractor{
				reflect.String: func(_ context.Context, t justest.JustT, v any) (any, bool) {
					return "foo: " + v.(string), true
				},
			},
			actualProvider: func(tc *testCase) any {
				return func() {
					tc.called = true
				}
			},
			expectedValue: nil,
			expectedFound: false,
			wantCalled:    true,
		},
		"no_in_one_non_error_out_returns_out_and_found": {
			defaultExtractor: justest.ExtractorUnsupported,
			extractorsMap: map[reflect.Kind]justest.Extractor{
				reflect.String: func(_ context.Context, t justest.JustT, v any) (any, bool) {
					return "foo: " + v.(string), true
				},
			},
			actualProvider: func(tc *testCase) any {
				return func() string {
					tc.called = true
					return "bar"
				}
			},
			expectedValue: "bar",
			expectedFound: true,
			wantCalled:    true,
		},
		"recurse_no_in_one_non_error_out_returns_out_and_found": {
			defaultExtractor: justest.ExtractorUnsupported,
			extractorsMap: map[reflect.Kind]justest.Extractor{
				reflect.String: func(_ context.Context, t justest.JustT, v any) (any, bool) {
					return "foo: " + v.(string), true
				},
			},
			actualProvider: func(tc *testCase) any {
				return func() string {
					tc.called = true
					return "bar"
				}
			},
			expectedValue: "foo: bar",
			expectedFound: true,
			recurse:       true,
			wantCalled:    true,
		},
		"no_in_one_non-nil_error_out_fails": {
			defaultExtractor: justest.ExtractorUnsupported,
			extractorsMap: map[reflect.Kind]justest.Extractor{
				reflect.String: func(_ context.Context, t justest.JustT, v any) (any, bool) {
					return "foo: " + v.(string), true
				},
			},
			actualProvider: func(tc *testCase) any {
				return func() error {
					tc.called = true
					return fmt.Errorf("foobar")
				}
			},
			wantCalled: true,
			wantErr:    true,
		},
		"no_in_one_nil_error_out_returns_nil_and_found": {
			defaultExtractor: justest.ExtractorUnsupported,
			extractorsMap: map[reflect.Kind]justest.Extractor{
				reflect.String: func(_ context.Context, t justest.JustT, v any) (any, bool) {
					return "foo: " + v.(string), true
				},
			},
			actualProvider: func(tc *testCase) any {
				return func() error {
					tc.called = true
					return nil
				}
			},
			expectedValue: nil,
			expectedFound: true,
			wantCalled:    true,
		},
		"no_in_val_and_nil_error_out_returns_val_and_found": {
			defaultExtractor: justest.ExtractorUnsupported,
			extractorsMap: map[reflect.Kind]justest.Extractor{
				reflect.String: func(_ context.Context, t justest.JustT, v any) (any, bool) {
					return "foo: " + v.(string), true
				},
			},
			actualProvider: func(tc *testCase) any {
				return func() (string, error) {
					tc.called = true
					return "bar", nil
				}
			},
			expectedValue: "bar",
			expectedFound: true,
			wantCalled:    true,
		},
		"recurse_no_in_val_and_nil_error_out_returns_val_and_found": {
			defaultExtractor: justest.ExtractorUnsupported,
			extractorsMap: map[reflect.Kind]justest.Extractor{
				reflect.String: func(_ context.Context, t justest.JustT, v any) (any, bool) {
					return "foo: " + v.(string), true
				},
			},
			actualProvider: func(tc *testCase) any {
				return func() (string, error) {
					tc.called = true
					return "bar", nil
				}
			},
			recurse:       true,
			expectedValue: "foo: bar",
			expectedFound: true,
			wantCalled:    true,
		},
		"no_in_val_and_non-nil_error_out_fails": {
			defaultExtractor: justest.ExtractorUnsupported,
			extractorsMap: map[reflect.Kind]justest.Extractor{
				reflect.String: func(_ context.Context, t justest.JustT, v any) (any, bool) {
					return "foo: " + v.(string), true
				},
			},
			actualProvider: func(tc *testCase) any {
				return func() (string, error) {
					tc.called = true
					return "bar", nil
				}
			},
			expectedValue: "bar",
			expectedFound: true,
			wantCalled:    true,
		},
		"no_in_val_and_non-error_out_fails": {
			defaultExtractor: justest.ExtractorUnsupported,
			extractorsMap: map[reflect.Kind]justest.Extractor{
				reflect.String: func(_ context.Context, t justest.JustT, v any) (any, bool) {
					return "foo: " + v.(string), true
				},
			},
			actualProvider: func(tc *testCase) any {
				return func() (string, int) {
					tc.called = true
					return "bar", 2
				}
			},
			wantCalled: false,
			wantErr:    true,
		},
	}
	for name, tc := range testCases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			var actual any
			var found bool

			mt := &MockT{parent: t}
			defer func() {
				justest.GetHelper(t).Helper()
				if tc.wantCalled && !tc.called {
					t.Fatalf("Expected function to be called, but it was not")
				} else if !tc.wantCalled && tc.called {
					t.Fatalf("Expected function to not be called, but it was")
				}
			}()
			defer verifyTestCaseError(t, mt, tc.wantErr)

			ve := justest.NewValueExtractor(tc.defaultExtractor)
			for k, v := range tc.extractorsMap {
				ve[k] = v
			}

			extractor := justest.NewFuncExtractor(ve, tc.recurse)
			actual, found = extractor(context.Background(), mt, tc.actualProvider(&tc))
			if tc.expectedFound && !found {
				t.Fatalf("Expected value to be found, but it was not")
			} else if !tc.expectedFound && found {
				t.Fatalf("Expected value to not be found, but it was")
			} else if !cmp.Equal(tc.expectedValue, actual) {
				t.Fatalf("Unexpected value returned:\n%s", cmp.Diff(tc.expectedValue, actual))
			}
		})
	}
}
