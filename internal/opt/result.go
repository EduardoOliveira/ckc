package opt

type Result[T any] struct {
	Value T     `json:"return"`
	Error error `json:"error"`
}

func Ok[T any](value T) Result[T] {
	return Result[T]{Value: value, Error: nil}
}

func Err[T any](err error) Result[T] {
	return Result[T]{Value: *new(T), Error: err}
}
