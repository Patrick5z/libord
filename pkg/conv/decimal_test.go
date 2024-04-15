package conv

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_Decimal(t *testing.T) {
	d := Decimal("32")
	assert.Equal(t, d.StringFixed(0), "32")

	d = Decimal(int64(2222222))
	assert.Equal(t, d.StringFixed(0), "2222222")

	d = Decimal(0.23232)
	assert.Equal(t, d.StringFixed(5), "0.23232")

	d = Decimal("0x1234")
	assert.Equal(t, d.StringFixed(0), "4660")
}
