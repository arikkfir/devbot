package justest

import (
	"bytes"
	"fmt"
	"github.com/alecthomas/chroma/v2/quick"
	"github.com/secureworks/errors"
	"os"
	"path/filepath"
	"strings"
)

type Results interface {
	Result() []any
	OrFail()
}

type Evaluator interface {
	Results
	Because(string, ...any) Results
}

func newEvaluator(t TT, matcher Matcher, actuals ...any) Evaluator {
	GetHelper(t).Helper()
	e := &evaluator{t: t, actuals: actuals, matcher: matcher}

	var location, source string
	for _, frame := range errors.CallStackAt(0) {
		function, file, line := frame.Location()

		startsWithAnIgnoredPrefix := false
		for _, prefix := range ignoredStackTracePrefixes {
			if strings.HasPrefix(function, prefix) {
				startsWithAnIgnoredPrefix = true
				break
			}
		}

		if !startsWithAnIgnoredPrefix {
			location = fmt.Sprintf("%s:%d", filepath.Base(file), line)

			// Try to read the actual statement that fails
			if b, err := os.ReadFile(file); err == nil {
				fileContents := string(b)
				lines := strings.Split(fileContents, "\n")
				if len(lines) > line {
					source = strings.TrimSpace(lines[line-1])

					output := bytes.Buffer{}
					err := quick.Highlight(&output, source, "go", goSourceFormatter, goSourceStyle)
					if err == nil {
						source = output.String()
					}
				}
			}
			break
		}
	}

	t.Cleanup(func() {
		GetHelper(t).Helper()
		if !e.evaluated {
			_, _ = fmt.Fprintf(os.Stderr, "\t%s: %s <-- Unevaluated expectation!\n", location, source)
			//t.Logf("%s: %s <-- Unevaluated expectation!", location, source)
		}
		source = location
	})

	return e
}

type evaluator struct {
	t         TT
	actuals   []any
	matcher   Matcher
	evaluated bool
}

func (e *evaluator) Because(format string, args ...any) Results {
	GetHelper(e.t).Helper()
	et := &explainedTT{parent: e.t, format: format, args: args}
	e.t.Cleanup(et.PerformCleanups)
	e.t = et
	return e
}

func (e *evaluator) OrFail() {
	GetHelper(e.t).Helper()
	e.Result()
}

func (e *evaluator) Result() []any {
	GetHelper(e.t).Helper()
	e.evaluated = true
	result := e.matcher(e.t, e.actuals...)
	return result
}
