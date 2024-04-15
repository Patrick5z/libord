package conv

import "strconv"

// Float32 converts `any` to float32.
func Float32(args ...any) float32 {
	return float32(Float64(args...))
}

// Float64 converts `any` to float64.
func Float64(args ...any) float64 {
	if len(args) == 0 {
		return 0
	}
	arg := args[0]
	defaultValue := float64(0)
	if len(args) > 1 {
		defaultValue = Float64(args[1])
	}
	switch arg.(type) {
	case float32:
		return float64(arg.(float32))
	case float64:
		return arg.(float64)
	default:
		if r, err := strconv.ParseFloat(String(arg), 64); err == nil {
			return r
		}
		return defaultValue
	}
}
