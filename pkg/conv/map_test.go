package conv

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_Map(t *testing.T) {
	arr := []string{"a", "b", "c"}
	ret := Map(arr)
	for _, item := range arr {
		assert.NotNil(t, ret[item])
	}

	str := "this is no map str"
	ret = Map(str)
	assert.Equal(t, len(ret), 0)

	b := []byte("{\"name\":\"a\"}")
	ret = Map(b)
	assert.Equal(t, ret["name"], "a")

	type user struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}
	ret = Map(&user{
		Name: "alex",
		Age:  30,
	})
	assert.Equal(t, ret["name"], "alex")
	assert.Equal(t, ret["age"], float64(30))
}
