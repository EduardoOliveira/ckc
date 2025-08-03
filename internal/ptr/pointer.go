package ptr

import "github.com/EduardoOliveira/ckc/internal/opt"

func Val[T any](v *T) T {
	if v == nil {
		var zero T
		return zero
	}
	return *v
}

func To[T any](v T) *T {
	return &v
}

func ToOptional[T any](v *T) opt.Optional[T] {
	if v == nil {
		return opt.None[T]()
	}
	return opt.Some(*v)
}
