package orm

import (
	"database/sql"
	"libord/pkg/conv"
	"reflect"

	_ "github.com/go-sql-driver/mysql"
)

type Orm struct {
	Db *sql.DB
}

func (o *Orm) Find(m *Model) (ret []any, errRet error) {
	defer m.clean()
	rows, err := o.Db.Query((&sqlModel{Model: m}).buildSelectSQL(), m.getArgs()...)
	if err != nil {
		errRet = err
		return
	}
	defer rows.Close()

	for rows.Next() {
		columns, err3 := rows.Columns()
		if err3 != nil {
			errRet = err3
			return
		}
		columnsMp := make(map[string]any, len(columns))
		refs := make([]any, 0, len(columns))
		for _, col := range columns {
			var ref any
			columnsMp[col] = &ref
			refs = append(refs, &ref)
		}
		if err2 := rows.Scan(refs...); err2 != nil && err2 != sql.ErrNoRows {
			errRet = err2
			return
		}
		item := make(map[string]any)
		for k, v := range columnsMp {
			if v != nil {
				value := reflect.ValueOf(v).Elem().Interface()
				item[k] = conv.String(value)
			}
		}
		ret = append(ret, m.convert(item))
	}
	return
}

func (o *Orm) Save(m *Model) (affected, lastInsertId int64, errRet error) {
	defer m.clean()
	result, err := o.Db.Exec((&sqlModel{Model: m}).buildInsertSQL(), m.getArgs()...)
	if err != nil {
		errRet = err
	} else {
		affected, errRet = result.RowsAffected()
		lastInsertId, errRet = result.LastInsertId()
	}
	return
}

func (o *Orm) Update(m *Model) (affected int64, errRet error) {
	defer m.clean()
	result, err := o.Db.Exec((&sqlModel{Model: m}).buildUpdateSQL(), m.getArgs()...)
	if err != nil {
		errRet = err
	} else {
		affected, errRet = result.RowsAffected()
	}
	return
}

func (o *Orm) Delete(m *Model) (affected int64, errRet error) {
	defer m.clean()
	result, err := o.Db.Exec((&sqlModel{Model: m}).buildDeleteSQL(), m.getArgs()...)
	if err != nil {
		errRet = err
	} else {
		affected, errRet = result.RowsAffected()
	}
	return
}

// One query one rows with field, e.g: select count(*) as c, field='c' or select tx from test limit 1, field='tx'
func (o *Orm) One(m *Model, field string) (ret any, errRet error) {
	defer m.clean()
	items, err := o.Find(m)
	if err != nil {
		errRet = err
	} else if len(items) > 0 {
		if field == "" {
			ret = items[0]
		} else {
			item := conv.Map(items[0])
			ret = item[field]
		}
	}
	return
}
