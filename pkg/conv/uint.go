// Copyright GoFrame Author(https://goframe.org). All Rights Reserved.
//
// This Source Code Form is subject to the terms of the MIT License.
// If a copy of the MIT was not distributed with this file,
// You can obtain one at https://github.com/gogf/gf.

package conv

import (
	"strconv"
)

// Uint converts `any` to uint.
func Uint(args ...any) uint {
	return uint(Uint64(args...))
}

// Uint8 converts `any` to uint8.
func Uint8(args ...any) uint8 {
	return uint8(Uint64(args...))
}

// Uint16 converts `any` to uint16.
func Uint16(args ...any) uint16 {
	return uint16(Uint64(args...))
}

// Uint32 converts `any` to uint32.
func Uint32(args ...any) uint32 {
	return uint32(Uint64(args...))
}

// Uint64 converts `any` to uint64.
func Uint64(args ...any) uint64 {
	if len(args) == 0 {
		return 0
	}
	arg := args[0]
	defaultValue := uint64(0)
	if len(args) > 1 {
		defaultValue = Uint64(args[1])
	}
	switch value := arg.(type) {
	case int:
		return uint64(value)
	case int8:
		return uint64(value)
	case int16:
		return uint64(value)
	case int32:
		return uint64(value)
	case int64:
		return uint64(value)
	case uint:
		return uint64(value)
	case uint8:
		return uint64(value)
	case uint16:
		return uint64(value)
	case uint32:
		return uint64(value)
	case uint64:
		return value
	case float32:
		return uint64(value)
	case float64:
		return uint64(value)
	case bool:
		if value {
			return 1
		}
		return 0
	default:
		s := String(value)
		// Hexadecimal
		if len(s) > 2 && s[0] == '0' && (s[1] == 'x' || s[1] == 'X') {
			if v, e := strconv.ParseUint(s[2:], 16, 64); e == nil {
				return v
			}
		}
		// Decimal
		if v, e := strconv.ParseUint(s, 10, 64); e == nil {
			return v
		}
		return defaultValue
	}
}
