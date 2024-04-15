package conv

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_Struct(t *testing.T) {
	type user struct {
		Name string `json:"name"`
		Age  string `json:"age"`
	}

	v := map[string]any{
		"name": "alex",
		"age":  "30",
	}

	u := &user{}
	Struct(v, u)

	assert.Equal(t, u.Name, "alex")
	assert.Equal(t, u.Age, "30")
}

func Test_StructLoose(t *testing.T) {
	type user struct {
		Name string `json:"name"`
		Age  string `json:"age"`
	}

	v := map[string]any{
		"name": "alex",
		"age":  30,
	}

	u := &user{}
	StructLoose(v, u)

	assert.Equal(t, u.Name, "alex")
	assert.Equal(t, u.Age, "30")

	type user2 struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	v = map[string]any{
		"name": "alex",
		"age":  "30",
	}

	u2 := &user2{}
	StructLoose(v, u2)

	assert.Equal(t, u2.Name, "alex")
	assert.Equal(t, u2.Age, 30)
}
