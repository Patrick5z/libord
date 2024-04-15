package conv

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_Bool(t *testing.T) {
	assert.Equal(t, false, Bool("hello"))
	assert.Equal(t, true, Bool("true"))
	assert.Equal(t, false, Bool(struct{}{}))
	assert.Equal(t, false, Bool([]string{}))
	assert.Equal(t, true, Bool([]string{"hello"}))
	assert.Equal(t, true, Bool("cantparse", true))
}
