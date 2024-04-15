package conv

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_String(t *testing.T) {
	i := 10
	assert.Equal(t, String(i), "10")

	type user struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}
	u := &user{
		Name: "alex",
		Age:  30,
	}
	assert.Equal(t, String(u), "{\"name\":\"alex\",\"age\":30}")
}

func Test_Byte_Array(t *testing.T) {
	var b any
	b = [11]byte{'h', 'e'}

	assert.EqualValues(t, "he", String(b))
}
