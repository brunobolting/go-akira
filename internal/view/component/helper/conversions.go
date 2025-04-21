package helper

import "strconv"

func String(v any) string {
	if v == nil {
		return ""
	}
	switch v := v.(type) {
	case string:
		return v
	case []byte:
		return string(v)
	case int:
		return strconv.Itoa(v)
	case float32:
		return strconv.FormatFloat(float64(v), 'f', -1, 32)
	default:
		return ""
	}
}

func Conditional(v bool, truthy, falsy any) string {
	if v {
		return String(truthy)
	}
	return String(falsy)
}
