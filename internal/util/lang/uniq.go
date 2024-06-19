package lang

func Uniq[T comparable](elements []T) []T {
	encountered := map[T]bool{}
	var result []T
	for v := range elements {
		if !encountered[elements[v]] {
			encountered[elements[v]] = true
			result = append(result, elements[v])
		}
	}
	return result
}
