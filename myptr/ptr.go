package myptr

// Sometimes implementations want to receive a pointer, and then you have to do this
// silly dance to get a pointer to a value.
// value := 42
// ptr := &value
// callTheImplementation(ptr)
// This package provides a function to do this silly dance for you.

func Ref[T any](v T) *T {
	return &v
}
