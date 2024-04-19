package justest_test

import (
	"fmt"
	"github.com/arikkfir/devbot/backend/internal/util/lang"
	. "github.com/arikkfir/devbot/backend/internal/util/testing/justest"
	"github.com/google/go-cmp/cmp"
	"reflect"
	"testing"
)

func TestValueExtractor(t *testing.T) {
	t.Run("Default extractor is used when no extractors have been defined", func(t *testing.T) {
		t.Parallel()
		mt := NewMockT(NewTT(t))
		ve := NewValueExtractor(func(t TT, v any) (any, bool) { return "bar", true })
		v := ve.MustExtractValue(mt, "foo")
		if "bar" != v {
			t.Fatalf("Expected '%s', got '%s'", "bar", v)
		}
	})
	t.Run("Nil actual finds nil result", func(t *testing.T) {
		t.Parallel()
		mt := NewMockT(NewTT(t))
		ve := NewValueExtractor(func(t TT, v any) (any, bool) { return "bar", true })
		v, found := ve.ExtractValue(mt, nil)
		if v != nil {
			t.Fatalf("Expected 'nil', got '%s'", v)
		}
		if !found {
			t.Fatalf("Expected found to be true, but it is false")
		}
	})
	t.Run("Invokes correct extractor when kind found", func(t *testing.T) {
		mt := NewMockT(NewTT(t))
		ve := NewValueExtractor(ExtractorUnsupported)
		ve[reflect.String] = func(t TT, v any) (any, bool) { return "foo: " + v.(string), true }
		v, found := ve.ExtractValue(mt, "bar")
		if v == nil {
			t.Fatalf("Expected 'foo: bar', got 'nil'")
		} else if v != "foo: bar" {
			t.Fatalf("Expected 'foo: bar', got '%s'", v)
		}
		if !found {
			t.Fatalf("Expected found to be true, but it is false")
		}
	})
	t.Run("Default extractor when kind not found", func(t *testing.T) {
		mt := NewMockT(NewTT(t))
		ve := NewValueExtractor(func(t TT, v any) (any, bool) { return "bar", true })
		ve[reflect.String] = func(t TT, v any) (any, bool) { return "foo: " + v.(string), true }
		v, found := ve.ExtractValue(mt, 1)
		if v == nil {
			t.Fatalf("Expected 'foo: bar', got 'nil'")
		} else if v != "bar" {
			t.Fatalf("Expected 'bar', got '%s'", v)
		}
		if !found {
			t.Fatalf("Expected found to be true, but it is false")
		}
	})
	t.Run("Failure occurs when value is required and not found", func(t *testing.T) {
		mt := NewMockT(NewTT(t))
		defer expectFailure(t, mt, `Value could not be extracted from an actual of type 'int': 1`)

		ve := NewValueExtractor(func(t TT, v any) (any, bool) { return nil, false })
		_ = ve.MustExtractValue(mt, 1)
	})
}

func TestNewChannelExtractor(t *testing.T) {
	testCases := map[string]struct {
		defaultExtractor Extractor
		extractorsMap    map[reflect.Kind]Extractor
		chanProvider     func() chan any
		recurse          bool
		expectedValue    any
		expectedFound    bool
		wantErr          bool
	}{
		"closed_channel_returns_nil_not_found": {
			defaultExtractor: ExtractorUnsupported,
			chanProvider: func() chan any {
				ch := make(chan any, 1)
				close(ch)
				return ch
			},
		},
		"empty_channel_returns_nil_not_found": {
			defaultExtractor: ExtractorUnsupported,
			chanProvider:     func() chan any { return make(chan any, 1) },
			expectedValue:    nil,
		},
		"recurse_extracts_value_from_channel_value_1": {
			defaultExtractor: ExtractorUnsupported,
			extractorsMap: map[reflect.Kind]Extractor{
				reflect.String: func(t TT, v any) (any, bool) {
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
			defaultExtractor: ExtractorUnsupported,
			extractorsMap: map[reflect.Kind]Extractor{
				reflect.String: func(t TT, v any) (any, bool) {
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
			defaultExtractor: ExtractorUnsupported,
			extractorsMap: map[reflect.Kind]Extractor{
				reflect.String: func(t TT, v any) (any, bool) {
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
			defaultExtractor: ExtractorUnsupported,
			extractorsMap: map[reflect.Kind]Extractor{
				reflect.String: func(t TT, v any) (any, bool) {
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
			mt := NewMockT(NewTT(t))
			defer verifyTestCaseError(t, mt, tc.wantErr)

			ve := NewValueExtractor(tc.defaultExtractor)
			for k, v := range tc.extractorsMap {
				ve[k] = v
			}

			extractor := NewChannelExtractor(ve, tc.recurse)
			actual, found := extractor(mt, tc.chanProvider())
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
		defaultExtractor Extractor
		extractorsMap    map[reflect.Kind]Extractor
		actual           any
		recurse          bool
		expectedValue    any
		expectedFound    bool
		wantErr          bool
	}{
		"recurse_extracts_value_from_pointer_1": {
			defaultExtractor: ExtractorUnsupported,
			extractorsMap: map[reflect.Kind]Extractor{
				reflect.String: func(t TT, v any) (any, bool) {
					return "foo: " + v.(string), true
				},
			},
			actual:        lang.Ptr[string]("bar"),
			recurse:       true,
			expectedValue: "foo: bar",
			expectedFound: true,
		},
		"recurse_extracts_value_from_pointer_2": {
			defaultExtractor: ExtractorUnsupported,
			extractorsMap: map[reflect.Kind]Extractor{
				reflect.String: func(t TT, v any) (any, bool) {
					return nil, false
				},
			},
			actual:        lang.Ptr[string]("bar"),
			recurse:       true,
			expectedValue: nil,
			expectedFound: false,
		},
		"recurse_extracts_value_from_pointer_3": {
			defaultExtractor: ExtractorUnsupported,
			extractorsMap: map[reflect.Kind]Extractor{
				reflect.String: func(t TT, v any) (any, bool) {
					t.Fatalf("Extractor fails")
					panic("unreachable")
				},
			},
			actual:  lang.Ptr[string]("bar"),
			recurse: true,
			wantErr: true,
		},
		"no_recurse_returns_elem_value": {
			defaultExtractor: ExtractorUnsupported,
			actual:           lang.Ptr[string]("bar"),
			expectedValue:    "bar",
			expectedFound:    true,
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			mt := NewMockT(NewTT(t))
			defer verifyTestCaseError(t, mt, tc.wantErr)

			ve := NewValueExtractor(tc.defaultExtractor)
			for k, v := range tc.extractorsMap {
				ve[k] = v
			}

			extractor := NewPointerExtractor(ve, tc.recurse)
			actual, found := extractor(mt, tc.actual)
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
		defaultExtractor Extractor
		extractorsMap    map[reflect.Kind]Extractor
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
			defaultExtractor: ExtractorUnsupported,
			extractorsMap: map[reflect.Kind]Extractor{
				reflect.String: func(t TT, v any) (any, bool) {
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
			defaultExtractor: ExtractorUnsupported,
			extractorsMap: map[reflect.Kind]Extractor{
				reflect.String: func(t TT, v any) (any, bool) {
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
			defaultExtractor: ExtractorUnsupported,
			extractorsMap: map[reflect.Kind]Extractor{
				reflect.String: func(t TT, v any) (any, bool) {
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
			defaultExtractor: ExtractorUnsupported,
			extractorsMap: map[reflect.Kind]Extractor{
				reflect.String: func(t TT, v any) (any, bool) {
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
			defaultExtractor: ExtractorUnsupported,
			extractorsMap: map[reflect.Kind]Extractor{
				reflect.String: func(t TT, v any) (any, bool) {
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
			defaultExtractor: ExtractorUnsupported,
			extractorsMap: map[reflect.Kind]Extractor{
				reflect.String: func(t TT, v any) (any, bool) {
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
			defaultExtractor: ExtractorUnsupported,
			extractorsMap: map[reflect.Kind]Extractor{
				reflect.String: func(t TT, v any) (any, bool) {
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
			defaultExtractor: ExtractorUnsupported,
			extractorsMap: map[reflect.Kind]Extractor{
				reflect.String: func(t TT, v any) (any, bool) {
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
			defaultExtractor: ExtractorUnsupported,
			extractorsMap: map[reflect.Kind]Extractor{
				reflect.String: func(t TT, v any) (any, bool) {
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

			mt := NewMockT(NewTT(t))
			defer func() {
				GetHelper(t).Helper()
				if tc.wantCalled && !tc.called {
					t.Fatalf("Expected function to be called, but it was not")
				} else if !tc.wantCalled && tc.called {
					t.Fatalf("Expected function to not be called, but it was")
				}
			}()
			defer verifyTestCaseError(t, mt, tc.wantErr)

			ve := NewValueExtractor(tc.defaultExtractor)
			for k, v := range tc.extractorsMap {
				ve[k] = v
			}

			extractor := NewFuncExtractor(ve, tc.recurse)
			actual, found = extractor(mt, tc.actualProvider(&tc))
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
