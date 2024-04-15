package orm

import (
	"database/sql"
	"fmt"
	"libord/pkg/conv"
	"testing"

	_ "github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/assert"
)

type User struct {
	meta    string   `table:"user"`
	ID      int64    `json:"bid"`
	Name    string   `json:"name"`
	Age     int      `json:"age"`
	Address string   `json:"address"`
	Friends []*User  `json:"friends"`
	Tasks   []string `json:"tasks"`
}

type Address struct {
	City     string `json:"city"`
	Province string `json:"province"`
}

type UserOne2One struct {
	meta    string   `table:"user"`
	ID      int64    `json:"bid"`
	Name    string   `json:"name"`
	Age     int      `json:"age"`
	Address *Address `json:"address"`
	Friends []*User  `json:"friends"`
	Tasks   []string `json:"tasks"`
}

func Test_Orm_Struct(t *testing.T) {
	_db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s)/%s", "root", "password", "localhost", "test"))
	assert.Nil(t, err)
	defer _db.Close()

	o := &Orm{Db: _db}

	var users []any
	m := &Model{}

	defer func() {
		_, err = o.Delete(m.Table("user").WhereGT("bid", 0))
		assert.Nil(t, err)
	}()

	for i := 0; i < 10; i++ {
		users = append(users, &User{
			ID:      int64(i + 1),
			Name:    "test_" + conv.String(i),
			Age:     30 + i,
			Address: "beijing_" + conv.String(i),
		})
		_, _ = o.Delete(m.Bind(User{}).Where("ID", i+1))
	}
	affected, _, err3 := o.Save(m.Bind(User{}).BatchData(users...))
	assert.Nil(t, err3)
	assert.Equal(t, affected, int64(10))

	items, err4 := o.Find(m.Bind(User{}).Where("ID", 2))
	assert.Nil(t, err4)
	assert.Equal(t, items[0].(*User).Name, "test_1")

	items, err4 = o.Find(m.Bind(&User{}).WhereGT("ID", 6).Extra("order by bid asc"))
	assert.Nil(t, err4)
	assert.Equal(t, items[0].(*User).Name, "test_6")
	assert.Equal(t, items[1].(*User).Name, "test_7")
	assert.Equal(t, items[2].(*User).Name, "test_8")

	affected, err = o.Update(m.Bind(User{}).Update("Name", "change_name_2").Where("ID", 3))
	assert.Nil(t, err)
	assert.Equal(t, affected, int64(1))

	items, err4 = o.Find(m.Bind(User{}).Where("ID", 3))
	assert.Nil(t, err4)
	assert.Equal(t, items[0].(*User).Name, "change_name_2")

	affected, err = o.Update(m.Bind(User{}).Update("Name", "change_name_all").WhereGT("ID", 7))
	assert.Nil(t, err)
	assert.Equal(t, affected, int64(3))

	items, err4 = o.Find(m.Bind(User{}).Where("ID", 8))
	assert.Nil(t, err4)
	assert.Equal(t, items[0].(*User).Name, "change_name_all")

	c, err := o.One(m.Table("user").Fields("count(*) as c"), "c")
	assert.Nil(t, err)
	assert.Equal(t, c, "10")

	affected, err = o.Delete(m.Bind(User{}).Where("ID", 3))
	assert.Nil(t, err)
	assert.Equal(t, affected, int64(1))

	items, err4 = o.Find(m.Bind(User{}).Where("ID", 3))
	assert.Nil(t, err4)
	assert.Equal(t, len(items), 0)
}

func Test_Orm_Struct_One2One(t *testing.T) {
	_db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s)/%s", "root", "password", "localhost", "test"))
	assert.Nil(t, err)
	defer _db.Close()

	o := &Orm{Db: _db}
	m := &Model{}

	defer func() {
		_, err = o.Delete(m.Table("user").WhereGT("bid", 0))
		assert.Nil(t, err)
	}()

	var users []any
	for i := 0; i < 10; i++ {
		users = append(users, &UserOne2One{
			ID:   int64(i + 1),
			Name: "test_" + conv.String(i),
			Age:  30 + i,
			Address: &Address{
				City:     "beijing_" + conv.String(i),
				Province: "beijing",
			},
			Friends: []*User{{ID: 20, Name: "fA" + conv.String(i)}, {ID: 30, Name: "fB" + conv.String(i)}},
			Tasks:   []string{"task_" + conv.String(i)},
		})
	}
	user := UserOne2One{}
	affected, _, err3 := o.Save(m.Bind(user).BatchData(users...))
	assert.Nil(t, err3)
	assert.Equal(t, affected, int64(10))

	items, err4 := o.Find(m.Bind(user).Where("ID", 2))
	assert.Nil(t, err4)
	assert.Equal(t, items[0].(*UserOne2One).Name, "test_1")
	assert.Equal(t, items[0].(*UserOne2One).Address.City, "beijing_1")
	assert.Equal(t, items[0].(*UserOne2One).Tasks[0], "task_1")
	assert.Equal(t, items[0].(*UserOne2One).Friends[1].Name, "fB1")

	items, err4 = o.Find(m.Bind(user).WhereGT("ID", 6).Extra("order by bid asc"))
	assert.Nil(t, err4)
	assert.Equal(t, items[0].(*UserOne2One).Address.City, "beijing_6")
	assert.Equal(t, items[1].(*UserOne2One).Address.City, "beijing_7")
	assert.Equal(t, items[2].(*UserOne2One).Address.City, "beijing_8")

	affected, err = o.Update(m.Bind(user).Update("Address", &Address{City: "change_city_2"}).Where("ID", 3))
	assert.Nil(t, err)
	assert.Equal(t, affected, int64(1))

	items, err4 = o.Find(m.Bind(user).Where("ID", 3))
	assert.Nil(t, err4)
	assert.Equal(t, items[0].(*UserOne2One).Address.City, "change_city_2")

}

func Test_Orm_Map(t *testing.T) {
	_db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s)/%s", "root", "password", "localhost", "test"))
	assert.Nil(t, err)
	defer _db.Close()

	o := &Orm{Db: _db}

	var users []any
	m := &Model{}

	defer func() {
		_, err = o.Delete(m.Table("user").WhereGT("bid", 0))
		assert.Nil(t, err)
	}()

	for i := 0; i < 10; i++ {
		users = append(users, map[string]any{
			"bid":     int64(i + 1),
			"name":    "test_" + conv.String(i),
			"age":     30 + i,
			"address": &Address{City: "beijing_" + conv.String(i)},
		})
	}
	if users == nil {
		return
	}
	affected, _, err3 := o.Save(m.Table("user").Bind(users[0]).BatchData(users...))
	assert.Nil(t, err3)
	assert.Equal(t, affected, int64(10))

	items, err4 := o.Find(m.Table("user").Bind(users[0]).Where("bid", 2))
	assert.Nil(t, err4)
	assert.Equal(t, items[0].(map[string]any)["name"], "test_1")

	items, err4 = o.Find(m.Table("user").Bind(users[0]).WhereGT("bid", 6).Extra("order by bid asc"))
	assert.Nil(t, err4)
	assert.Equal(t, items[0].(map[string]any)["name"], "test_6")
	assert.Equal(t, items[1].(map[string]any)["name"], "test_7")
	assert.Equal(t, items[2].(map[string]any)["name"], "test_8")

	affected, err = o.Update(m.Table("user").Update("Name", "change_name_2").Where("bid", 3))
	assert.Nil(t, err)
	assert.Equal(t, affected, int64(1))

	items, err4 = o.Find(m.Table("user").Bind(users[0]).Where("bid", 3))
	assert.Nil(t, err4)
	assert.Equal(t, items[0].(map[string]any)["name"], "change_name_2")

	affected, err = o.Update(m.Table("user").Update("Name", "change_name_all").WhereGT("bid", 7))
	assert.Nil(t, err)
	assert.Equal(t, affected, int64(3))

	items, err4 = o.Find(m.Table("user").Bind(users[0]).Where("bid", 8))
	assert.Nil(t, err4)
	assert.Equal(t, items[0].(map[string]any)["name"], "change_name_all")

	c, err := o.One(m.Table("user").Fields("count(*) as c"), "c")
	assert.Nil(t, err)
	assert.Equal(t, c, "10")

	affected, err = o.Delete(m.Table("user").Where("bid", 3))
	assert.Nil(t, err)
	assert.Equal(t, affected, int64(1))

	items, err4 = o.Find(m.Table("user").Bind(users[0]).Where("bid", 3))
	assert.Nil(t, err4)
	assert.Equal(t, len(items), 0)
}
