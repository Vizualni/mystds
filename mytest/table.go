package mytest

import "testing"

type testcase[T any] struct {
	name     string
	parallel bool
	data     T
}

type Tests[T any] struct {
	got   *testing.T
	cases []testcase[T]
}

func NewTests[T any](t *testing.T) Tests[T] {
	return Tests[T]{
		got: t,
	}
}

func (t *Tests[T]) Add(name string, data T) *Tests[T] {
	t.cases = append(t.cases, testcase[T]{name: name, data: data})
	return t
}

func (t *Tests[T]) AddParallel(name string, data T) *Tests[T] {
	t.cases = append(t.cases, testcase[T]{
		name:     name,
		data:     data,
		parallel: true,
	})
	return t
}

func (t *Tests[T]) Test(f func(t *testing.T, tc T)) {
	for _, tc := range t.cases {
		t.got.Run(tc.name, func(got *testing.T) {
			if tc.parallel {
				got.Parallel()
			}
			f(got, tc.data)
		})
	}
}
