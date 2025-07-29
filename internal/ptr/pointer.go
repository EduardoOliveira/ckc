package ptr

import "github.com/EduardoOliveira/ckc/internal/opt"

func To[T any](v T) *T {
	return &v
}

func ToOptional[T any](v *T) opt.Optional[T] {
	if v == nil {
		return opt.None[T]()
	}
	return opt.Some(*v)
}
