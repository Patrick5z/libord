package conv

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_SliceAny(t *testing.T) {
	items := []string{"a", "b", "c"}
	ret := SliceAny(items)
	assert.EqualValues(t, len(ret), 3)
	assert.EqualValues(t, ret[0], "a")
	assert.EqualValues(t, ret[1], "b")
	assert.EqualValues(t, ret[2], "c")

	type user struct {
		name string
	}
	users := []*user{{name: "a"}, {name: "b"}, {name: "c"}}
	ret = SliceAny(users)
	assert.EqualValues(t, len(ret), 3)
	assert.EqualValues(t, ret[0].(*user).name, "a")
	assert.EqualValues(t, ret[1].(*user).name, "b")
	assert.EqualValues(t, ret[2].(*user).name, "c")

	str := "xx"
	ret = SliceAny(str)
	assert.EqualValues(t, len(ret), 0)
}
