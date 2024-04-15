package orm

import (
	"strings"
)

type sqlModel struct {
	Model *Model
}

func (m *sqlModel) buildSelectSQL() string {
	var columns []string
	for _, col := range m.Model.getColumns() {
		if !strings.Contains(col, "(") { // don't add "`" for count(*)
			col = "`" + col + "`"
		}
		columns = append(columns, col)
	}
	if len(columns) > 0 {
		str := "select " + strings.Join(columns, ",") + " from " + m.Model.TablePrefix + m.Model.table
		if len(m.Model.whereConditions) > 0 {
			str += " where " + strings.Join(m.Model.whereConditions, " and ")
		}
		return str + " " + m.Model.extra
	}
	return m.Model.extra
}

func (m *sqlModel) buildInsertSQL() string {
	var columns []string
	for _, column := range m.Model.getColumns() {
		if !strings.EqualFold(column, "id") {
			columns = append(columns, "`"+column+"`")
		}
	}
	if len(columns) > 0 {
		var valuesClauses []string
		for range m.Model.data {
			var clause []string
			for range columns {
				clause = append(clause, "?")
			}
			valuesClauses = append(valuesClauses, "("+strings.Join(clause, ",")+")")
		}
		return "insert ignore into " + m.Model.TablePrefix + m.Model.table + "(" + strings.Join(columns, ",") + ") values " + strings.Join(valuesClauses, ",")
	}
	return m.Model.extra
}

func (m *sqlModel) buildUpdateSQL() string {
	str := "update " + m.Model.TablePrefix + m.Model.table + " set " + strings.Join(m.Model.updateClauses, ",")
	if len(m.Model.whereConditions) > 0 {
		str += " where " + strings.Join(m.Model.whereConditions, " and ")
	} else { // don't use update without any condition, it's so danger
		return ""
	}
	return str + " " + m.Model.extra
}

func (m *sqlModel) buildDeleteSQL() string {
	str := "delete from " + m.Model.TablePrefix + m.Model.table
	if len(m.Model.whereConditions) > 0 {
		str += " where " + strings.Join(m.Model.whereConditions, " and ")
	} else { // don't use delete without any condition, it's so danger
		return ""
	}
	return str + " " + m.Model.extra
}
