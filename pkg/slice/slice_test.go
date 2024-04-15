package slice

import (
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func Test_Sub(t *testing.T) {
	var items []string
	ret := Sub(items, 0, 1)
	assert.EqualValues(t, len(ret), 0)

	items = append(items, "a", "b", "c", "d", "e")
	ret = Sub(items, 0, 1)
	assert.EqualValues(t, len(ret), 1)
	assert.EqualValues(t, ret[0], "a")

	ret = Sub(items, 0, 100)
	assert.EqualValues(t, len(ret), 5)
	assert.EqualValues(t, ret[2], "c")

	ret = Sub(items, 3, 2)
	assert.EqualValues(t, len(ret), 0)

	ret = Sub(items, 3, 4)
	assert.EqualValues(t, len(ret), 1)
	assert.EqualValues(t, ret[0], "d")

	ret = Sub(items, 3, 6)
	assert.EqualValues(t, len(ret), 2)
	assert.EqualValues(t, ret[0], "d")
	assert.EqualValues(t, ret[1], "e")
}

func Test_Map(t *testing.T) {
	items := []string{"Hello", "worLd"}
	items2 := Map(items, func(item string) string {
		return strings.ToLower(item)
	})
	assert.Equal(t, items2[0], "hello")
	assert.Equal(t, items2[1], "world")
}

func Test_Filter(t *testing.T) {
	items := []string{"Hello", "worLd"}
	items2 := Filter(items, func(item string) bool {
		return strings.EqualFold(item, "world")
	})
	assert.Equal(t, len(items2), 1)
	assert.Equal(t, items2[0], "worLd")
}
