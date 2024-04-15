package conv

import "reflect"

func SliceStr(v []any) (ret []string) {
	for _, p := range v {
		ret = append(ret, String(p))
	}
	return
}

// SliceAny arg is only slice element
func SliceAny(arg any) (ret []any) {
	v := reflect.ValueOf(arg)
	if v.Kind() == reflect.Slice {
		for i := 0; i < v.Len(); i++ {
			ret = append(ret, v.Index(i).Interface())
		}
	}
	return
}
