package opt

import "encoding/json"

type Optional[T any] struct {
	Value   *T
	Present bool
}

func Some[T any](value T) Optional[T] {
	return Optional[T]{Value: &value, Present: true}
}

func None[T any]() Optional[T] {
	return Optional[T]{Present: false}
}

func FromMap[T any](m map[string]T, key string) Optional[T] {
	if value, exists := m[key]; exists {
		return Some(value)
	}
	return None[T]()
}

func FromOker[T any](value T, ok bool) Optional[T] {
	if ok {
		return Some(value)
	}
	return None[T]()
}

func (o Optional[T]) IsPresent() bool {
	return o.Present
}

func (o Optional[T]) OrElse(defaultValue T) T {
	if o.Present {
		return *o.Value
	}
	return defaultValue
}

func (o Optional[T]) MarshalJSON() ([]byte, error) {
	if o.Present {
		return json.Marshal(o.Value)
	}
	return json.Marshal(nil)
}

func (o *Optional[T]) UnmarshalJSON(data []byte) error {
	var value *T
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	if value != nil {
		o.Value = value
		o.Present = true
	} else {
		o.Value = nil
		o.Present = false
	}
	return nil
}
