package conv

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_HexInt(t *testing.T) {
	assert.EqualValues(t, 10, Int64("0xA"))
}

func Test_DefaultValue(t *testing.T) {
	assert.EqualValues(t, 20, Int64("hello", 20))
	assert.EqualValues(t, 10, Int8("hello", 10))
	assert.EqualValues(t, 10, Int8("", 10))
}
