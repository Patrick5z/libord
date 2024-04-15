package conv

import (
	"encoding/json"
	"reflect"
)

// Map ["a","b"] => {"a":true, "b": true}, "{"a":"1"}" => {"a":"1"}, &User{Name:"a"} => {"name":"a"}
func Map(strOrArrayOrStruct any) (ret map[string]any) {
	if strOrArrayOrStruct == nil {
		return
	}
	ret = make(map[string]interface{})
	switch strOrArrayOrStruct.(type) {
	case []byte:
		json.Unmarshal(strOrArrayOrStruct.([]byte), &ret)
	case string:
		json.Unmarshal([]byte(strOrArrayOrStruct.(string)), &ret)
	default:
		v := reflect.ValueOf(strOrArrayOrStruct)
		switch v.Kind() {
		case reflect.Slice: // 类似把 ["a","b"] => {"a":true, "b": true}
			for i := 0; i < v.Len(); i++ {
				ret[String(v.Index(i).Interface())] = true
			}
		default:
			if _bytes, _ := json.Marshal(strOrArrayOrStruct); len(_bytes) > 0 {
				_ = json.Unmarshal(_bytes, &ret)
			}
		}
	}

	return
}
