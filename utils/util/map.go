package util

func Map[S any, T any](slice []T, f func(T) S) []S {
	result := make([]S, len(slice))
	for i, v := range slice {
		result[i] = f(v)
	}
	return result
}
