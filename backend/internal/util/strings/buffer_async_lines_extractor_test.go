package strings

import (
	"bytes"
	"context"
	"github.com/google/go-cmp/cmp"
	"testing"
	"time"
)

func TestAsyncReadLinesFromBuffer(t *testing.T) {
	testCases := []struct {
		name      string
		input     *bytes.Buffer
		expOutput []string
		expError  error
		asyncFunc func(b *bytes.Buffer)
	}{
		{
			name:      "single_line",
			input:     bytes.NewBufferString("Hello, world"),
			expOutput: []string{"Hello, world"},
		},
		{
			name:      "multiple_lines",
			input:     bytes.NewBufferString("Hello\nworld"),
			expOutput: []string{"Hello", "world"},
		},
		{
			name:      "multiple_lines_ending_with_newline",
			input:     bytes.NewBufferString("Hello\nworld\n\nabc"),
			expOutput: []string{"Hello", "world", "", "abc"},
		},
		{
			name:      "empty_input",
			input:     bytes.NewBufferString(""),
			expOutput: []string{},
		},
		{
			name:      "just_new_line",
			input:     bytes.NewBufferString("\n"),
			expOutput: []string{""},
		},
		{
			name:      "lines_added_dynamically",
			input:     bytes.NewBufferString("Hello"),
			expOutput: []string{"Hello world", "abc"},
			asyncFunc: func(b *bytes.Buffer) {
				time.Sleep(50 * time.Millisecond)
				b.WriteString(" world")
				time.Sleep(200 * time.Millisecond)
				b.WriteString("\nabc")
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()

			strChan := make(chan string)
			errChan := make(chan error, 10)

			if tc.asyncFunc != nil {
				go tc.asyncFunc(tc.input)
			}

			done := false
			go func() {
				defer func() {
					done = true
					if r := recover(); r != nil {
						t.Errorf("unexpected panic: %v", r)
					}
				}()
				errChan <- AsyncReadLinesFromBuffer(ctx, tc.input, strChan)
			}()

			output := make([]string, 0)
			for {
				select {
				case <-ctx.Done():
					if done {
						if !cmp.Equal(tc.expOutput, output) {
							t.Errorf("unexpected output: %s", cmp.Diff(tc.expOutput, output))
						}
						err := <-errChan
						if !cmp.Equal(tc.expError, err) {
							t.Errorf("unexpected errors: %s", cmp.Diff(tc.expError, err))
						}
						return
					} else {
						time.Sleep(100 * time.Millisecond)
					}
				case str := <-strChan:
					output = append(output, str)
				}
			}
		})
	}
}
