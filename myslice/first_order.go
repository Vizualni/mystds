package myslice

func Map[T any, R any](slice []T, fnc func(T) R) []R {
	ret := make([]R, 0, len(slice))
	for _, el := range slice {
		ret = append(ret, fnc(el))
	}
	return ret
}
