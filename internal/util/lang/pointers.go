package lang

func Ptr[T any](x T) *T {
	return &[]T{x}[0]
}
