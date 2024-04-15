package conv

import (
	"reflect"
	"strconv"
)

// Bool converts `any` to bool.
// It returns false if `any` is: false, "", 0, "false", "off", "no", empty slice/map.
func Bool(args ...any) bool {
	if len(args) == 0 {
		return false
	}
	arg := args[0]
	defaultValue := false
	if len(args) > 1 {
		defaultValue = Bool(args[1])
	}
	switch value := arg.(type) {
	case bool:
		return value
	default:
		rv := reflect.ValueOf(arg)
		switch rv.Kind() {
		case reflect.Ptr:
			return !rv.IsNil()
		case reflect.Map:
			fallthrough
		case reflect.Array:
			fallthrough
		case reflect.Slice:
			return rv.Len() != 0
		case reflect.Struct:
			return false
		default:
			if r, err := strconv.ParseBool(String(arg)); err == nil {
				return r
			}
			return defaultValue
		}
	}
}
