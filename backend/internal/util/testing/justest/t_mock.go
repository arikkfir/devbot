package justest

type MockT struct {
	Parent      T
	Cleanups    []func()
	Failures    []FormatAndArgs
	LogMessages []FormatAndArgs
}

//go:noinline
func NewMockT(parent T) *MockT {
	return &MockT{Parent: parent}
}

//go:noinline
func (t *MockT) GetParent() T { return t.Parent }

//go:noinline
func (t *MockT) Name() string {
	return t.Parent.Name()
}

//go:noinline
func (t *MockT) Cleanup(f func()) { GetHelper(t).Helper(); t.Cleanups = append(t.Cleanups, f) }

//go:noinline
func (t *MockT) Fatalf(format string, args ...any) {
	GetHelper(t).Helper()
	t.Failures = append(t.Failures, FormatAndArgs{Format: &format, Args: args})
	panic(t)
}

//go:noinline
func (t *MockT) Failed() bool { GetHelper(t).Helper(); return len(t.Failures) > 0 }

//go:noinline
func (t *MockT) Log(args ...any) {
	GetHelper(t).Helper()
	t.LogMessages = append(t.LogMessages, FormatAndArgs{Args: args})
}

//go:noinline
func (t *MockT) Logf(format string, args ...any) {
	GetHelper(t).Helper()
	t.LogMessages = append(t.LogMessages, FormatAndArgs{Format: &format, Args: args})
}
