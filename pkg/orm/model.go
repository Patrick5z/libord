package orm

import (
	"encoding/json"
	"fmt"
	"libord/pkg/cmap"
	"libord/pkg/conv"
	"libord/pkg/maps"
	"reflect"
	"sort"
	"strings"
)

var (
	structTableNameCache cmap.ConcurrentMap
	structFieldsCache    cmap.ConcurrentMap
)

func init() {
	structTableNameCache = cmap.New()
	structFieldsCache = cmap.New()
}

type Model struct {
	TablePrefix     string
	obj             any
	table           string
	whereConditions []string
	updateClauses   []string
	extra           string
	args            []any
	columns         []string
	data            []any // data for save, may be object or map
}

func (m *Model) Table(table string) *Model {
	m.table = table
	return m
}

// Bind 1.The query fields are struct{} fields, avoiding access db columns, as db columns are changed frequently.
// 2. The query results are returned as struct{}.
// 3. The query fields are automatically detected.
func (m *Model) Bind(obj any) *Model {
	m.obj = obj
	switch reflect.ValueOf(obj).Kind() {
	case reflect.Map:
		m.columns = conv.SliceStr(maps.Keys2(obj))
	case reflect.Struct:
		m.obj = reflect.New(reflect.TypeOf(m.obj)).Interface()
		fallthrough
	case reflect.Ptr:
		if m.table == "" {
			structName := reflect.TypeOf(m.obj).Elem().Name()
			tableName := structTableNameCache.GetValue(structName, nil)
			if tableName == nil {
				t := reflect.TypeOf(m.obj).Elem()
				for i := 0; i < t.NumField(); i++ {
					f := t.Field(i)
					if strings.EqualFold(f.Name, "meta") {
						if table := f.Tag.Get("table"); table != "" {
							tableName = table
							break
						}
					}
				}
				structTableNameCache.Set(structName, tableName)
			}
			m.table = tableName.(string)
		}
		if len(m.columns) == 0 {
			field2ColumnMap := structFieldsCache.GetValue(m.table, nil)
			if field2ColumnMap == nil {
				field2ColumnMap = m.getField2ColumnMap(m.obj)
				structFieldsCache.Set(m.table, m.getField2ColumnMap(m.obj))
			}
			for _, v := range field2ColumnMap.(map[string]string) {
				m.columns = append(m.columns, v)
			}
		}
	}
	sort.Strings(m.columns)
	return m
}

// Fields custom columns
func (m *Model) Fields(fields ...string) *Model {
	m.columns = []string{}
	for _, field := range fields {
		m.columns = append(m.columns, m.getColumn(field))
	}
	return m
}

// BatchData Data for save
func (m *Model) BatchData(data ...any) *Model {
	m.data = data
	if len(data) == 0 {
		return m
	}
	rt := reflect.TypeOf(data[0]).Elem()
	for _, item := range data {
		itemMap := make(map[string]any)
		v := reflect.ValueOf(item)
		if v.Kind() == reflect.Ptr || v.Kind() == reflect.Struct {
			for i := 0; i < rt.NumField(); i++ {
				field := rt.Field(i)
				tag := field.Tag.Get("json")
				if tag != "" && tag != "-" {
					itemMap[strings.Split(tag, ",")[0]] = v.Elem().FieldByName(field.Name).Addr().Elem().Interface()
				}
			}
		} else {
			itemMap = conv.Map(item)
		}
		columns := m.getColumns()
		for _, column := range columns {
			if !strings.EqualFold(column, "id") { // auto increment, don't set
				m.args = append(m.args, m.getFieldValue(itemMap[column]))
			}
		}
	}
	return m
}

func (m *Model) Where(field string, value any) *Model {
	m.whereConditions = append(m.whereConditions, fmt.Sprintf("`%s`=?", m.getColumn(field)))
	m.args = append(m.args, m.getFieldValue(value))
	return m
}

func (m *Model) WhereLT(field string, value any) *Model {
	m.whereConditions = append(m.whereConditions, fmt.Sprintf("`%s`<?", m.getColumn(field)))
	m.args = append(m.args, m.getFieldValue(value))
	return m
}

func (m *Model) WhereLTE(field string, value any) *Model {
	m.whereConditions = append(m.whereConditions, fmt.Sprintf("`%s`<=?", m.getColumn(field)))
	m.args = append(m.args, m.getFieldValue(value))
	return m
}

func (m *Model) WhereGT(field string, value any) *Model {
	m.whereConditions = append(m.whereConditions, fmt.Sprintf("`%s`>?", m.getColumn(field)))
	m.args = append(m.args, m.getFieldValue(value))
	return m
}

func (m *Model) WhereGTE(field string, value any) *Model {
	m.whereConditions = append(m.whereConditions, fmt.Sprintf("`%s`>=?", m.getColumn(field)))
	m.args = append(m.args, m.getFieldValue(value))
	return m
}

func (m *Model) WhereIn(field string, values ...any) *Model {
	if len(values) == 0 {
		return m
	}
	str := strings.Repeat("?,", len(values))
	m.whereConditions = append(m.whereConditions, fmt.Sprintf("`%s` in ("+str[0:len(str)-1]+")", m.getColumn(field)))
	for _, v := range values {
		m.args = append(m.args, m.getFieldValue(v))
	}
	return m
}

func (m *Model) Update(field string, value any) *Model {
	m.updateClauses = append(m.updateClauses, fmt.Sprintf("`%s`=?", m.getColumn(field)))
	m.args = append(m.args, m.getFieldValue(value))
	return m
}

func (m *Model) Extra(extra string, args ...any) *Model {
	m.extra += extra
	m.args = append(m.args, args...)
	return m
}

func (m *Model) getArgs() []any {
	return m.args
}

func (m *Model) getColumns() []string {
	return m.columns
}

func (m *Model) getColumn(field string) string {
	if m.obj == nil || reflect.ValueOf(m.obj).Kind() == reflect.Map {
		return field
	}
	field2ColumnMap, _ := structFieldsCache.Get(m.table)
	if v, ok := field2ColumnMap.(map[string]string)[field]; ok {
		return v
	}
	return field
}

func (m *Model) convert(item map[string]any) any {
	if m.obj == nil || reflect.ValueOf(m.obj).Kind() == reflect.Map {
		return item
	}
	ret := reflect.New(reflect.TypeOf(m.obj).Elem()).Interface()
	conv.StructLoose(item, ret)
	return ret
}

func (m *Model) getField2ColumnMap(obj any) map[string]string {
	field2ColumnMap := make(map[string]string)
	t := reflect.TypeOf(obj).Elem()
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		tag := f.Tag.Get("json")
		if tag == "" || tag == "-" {
			continue
		}
		split := strings.Split(tag, ",")
		field2ColumnMap[f.Name] = split[0]
	}
	return field2ColumnMap
}

func (m *Model) getFieldValue(v any) any {
	switch (reflect.ValueOf(v)).Kind() {
	case reflect.Ptr:
		fallthrough
	case reflect.Slice:
		fallthrough
	case reflect.Map:
		_bytes, _ := json.Marshal(v)
		return string(_bytes)
	}
	return v
}

func (m *Model) clean() {
	m.table = ""
	m.obj = nil
	m.whereConditions = []string{}
	m.updateClauses = []string{}
	m.extra = ""
	m.args = []any{}
	m.columns = []string{}
	m.data = []any{}
}
