package conv

import (
	"strconv"
)

// Int converts `any` to int.
func Int(args ...any) int {
	return int(Int64(args...))
}

// Int8 converts `any` to int8.
func Int8(args ...any) int8 {
	return int8(Int64(args...))
}

// Int16 converts `any` to int16.
func Int16(args ...any) int16 {
	return int16(Int64(args...))
}

// Int32 converts `any` to int32.
func Int32(args ...any) int32 {
	return int32(Int64(args...))
}

// Int64 converts `any` to int64.
func Int64(args ...any) int64 {
	if len(args) == 0 {
		return 0
	}
	arg := args[0]
	defaultValue := int64(0)
	if len(args) > 1 {
		defaultValue = Int64(args[1])
	}
	switch value := arg.(type) {
	case int:
		return int64(value)
	case int8:
		return int64(value)
	case int16:
		return int64(value)
	case int32:
		return int64(value)
	case int64:
		return value
	case uint:
		return int64(value)
	case uint8:
		return int64(value)
	case uint16:
		return int64(value)
	case uint32:
		return int64(value)
	case uint64:
		return int64(value)
	case float32:
		return int64(value)
	case float64:
		return int64(value)
	case bool:
		if value {
			return 1
		}
		return 0
	default:
		var (
			s       = String(value)
			isMinus = false
		)
		if len(s) > 0 {
			if s[0] == '-' {
				isMinus = true
				s = s[1:]
			} else if s[0] == '+' {
				s = s[1:]
			}
		}
		// Hexadecimal
		if len(s) > 2 && s[0] == '0' && (s[1] == 'x' || s[1] == 'X') {
			if v, e := strconv.ParseInt(s[2:], 16, 64); e == nil {
				if isMinus {
					return -v
				}
				return v
			}
		}
		// Decimal
		if v, e := strconv.ParseInt(s, 10, 64); e == nil {
			if isMinus {
				return -v
			}
			return v
		}
		return defaultValue
	}
}
