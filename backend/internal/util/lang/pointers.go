package lang

func BoolPtr(b bool) *bool {
	lb := b
	return &lb
}
