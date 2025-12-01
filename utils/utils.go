package utils

func Contains[V comparable](a []V, b V) bool {
	for _, e := range a {
		if e == b {
			return true
		}
	}

	return false
}

func MapValues[K comparable, V any](m map[K]V) []V {
	res := make([]V, len(m))

	i := 0

	for _, v := range m {
		res[i] = v
		i++
	}

	return res
}
