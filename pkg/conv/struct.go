package conv

import (
	"encoding/json"
	"reflect"
	"strings"
	"time"
)

func Struct(_map, _struct any) {
	if _bytes, _ := json.Marshal(_map); len(_bytes) > 0 {
		json.Unmarshal(_bytes, &_struct)
	}
}

// StructLoose map type is not restrict.
func StructLoose[V any](_map map[string]V, _struct any) {
	v := reflect.ValueOf(_struct)
	t := reflect.TypeOf(_struct)
	elem := t.Elem()
	for i := 0; i < elem.NumField(); i++ {
		field := elem.Field(i)
		if field.Name[0] >= 'A' && field.Name[0] <= 'Z' { // only process UpperCase
			tag := field.Tag.Get("json")
			if tag == "" || tag == "-" {
				continue
			}
			split := strings.Split(tag, ",")
			if value, _found := _map[split[0]]; _found {
				SetFieldValue(value, v.Elem().FieldByName(field.Name))
			}
		}
	}
}

func SetFieldValue(value interface{}, f reflect.Value) {
	if value == nil || !f.IsValid() {
		return
	}
	switch value.(type) {
	case []byte, string:
		var v string
		switch value.(type) {
		case string:
			v = value.(string)
		case []byte:
			v = string(value.([]byte))
		}
		switch f.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			f.SetInt(Int64(v))
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
			f.SetUint(uint64(Int64(v)))
		case reflect.Float32, reflect.Float64:
			f.SetFloat(Float64(v))
		case reflect.Bool:
			f.SetBool(Bool(v))
		case reflect.String:
			f.SetString(v)
		case reflect.Slice:
			if v != "" {
				_slice := reflect.SliceOf(f.Type().Elem())   // construct []*Alarm
				_slicePtr := reflect.New(_slice).Interface() // got *[]*Alarm
				if err := json.Unmarshal([]byte(v), _slicePtr); err == nil {
					f.Set(reflect.ValueOf(_slicePtr).Elem())
				}
			}
		case reflect.Map:
			m := Map(v)
			f.Set(reflect.ValueOf(m))
		case reflect.Ptr:
			_typePtr := reflect.New(f.Type().Elem()).Interface() // got *Alarm
			if err3 := json.Unmarshal([]byte(v), _typePtr); err3 == nil {
				f.Set(reflect.ValueOf(_typePtr))
			}
		}
	case float64:
		v := value.(float64)
		switch f.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			f.SetInt(int64(v))
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
			f.SetUint(uint64(v))
		case reflect.Float32, reflect.Float64:
			f.SetFloat(v)
		case reflect.String:
			f.SetString(String(f))
		}
	case time.Time:
		v := value.(time.Time)
		switch f.Kind() {
		case reflect.Int64:
			f.SetInt(v.Unix())
		default:
			f.Set(reflect.ValueOf(value))
		}
	default:
		switch f.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			f.SetInt(Int64(value))
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
			f.SetUint(uint64(Int64(value)))
		case reflect.Float32, reflect.Float64:
			f.SetFloat(Float64(value))
		case reflect.Bool:
			f.SetBool(Bool(value))
		case reflect.String:
			f.SetString(String(value))
		case reflect.Ptr:
			f.Set(reflect.ValueOf(value))
		}
	}
}
