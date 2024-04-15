package maps

import "reflect"

// Keys2 use non-generics
func Keys2(m interface{}) []interface{} {
	var result []interface{}
	v := reflect.ValueOf(m)
	kind := v.Kind()
	switch kind {
	case reflect.Map:
		for _, val := range v.MapKeys() {
			result = append(result, val.Interface())
		}
	}
	return result
}

// Values2 use non-generics
func Values2(m interface{}) []interface{} {
	var result []interface{}
	v := reflect.ValueOf(m)
	kind := v.Kind()
	switch kind {
	case reflect.Map:
		for _, val := range v.MapKeys() {
			result = append(result, v.MapIndex(val).Interface())
		}
	}
	return result
}

// Keys use generics
func Keys[K comparable, V any](m map[K]V) []K {
	keys := make([]K, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func Values[K comparable, V comparable](m map[K]V) []V {
	values := make([]V, 0, len(m))
	for _, v := range m {
		values = append(values, v)
	}
	return values
}
