package maps

import "time"

func GetValueAsString[K comparable](m map[K]any, key K) (string, bool) {
	if value, ok := m[key]; ok {
		if str, ok := value.(string); ok {
			return str, true
		}
		return "", false
	}
	return "", false
}

func GetValueAsInt64[K comparable](m map[K]any, key K) (int64, bool) {
	if value, ok := m[key]; ok {
		if i, ok := value.(int64); ok {
			return i, true
		} else if f, ok := value.(float64); ok {
			return int64(f), true
		}
		return 0, false
	}
	return 0, false
}

func GetValueAsBool[K comparable](m map[K]any, key K) (bool, bool) {
	if value, ok := m[key]; ok {
		if b, ok := value.(bool); ok {
			return b, true
		} else if str, ok := value.(string); ok {
			return str == "true", true
		}
		return false, false
	}
	return false, false
}

func GetValueAsTime[K comparable](m map[K]any, key K) (time.Time, bool) {
	if value, ok := m[key]; ok {
		if t, ok := value.(time.Time); ok {
			return t, true
		} else if str, ok := value.(string); ok {
			t, err := time.Parse(time.RFC3339, str)
			if err == nil {
				return t, true
			}
		}
		return time.Time{}, false
	}
	return time.Time{}, false
}
