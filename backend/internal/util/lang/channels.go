package lang

func ChanOf[T any](values ...T) chan T {
	c := make(chan T, len(values))
	for _, v := range values {
		c <- v
	}
	return c
}
