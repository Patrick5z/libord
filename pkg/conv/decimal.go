package conv

import (
	"github.com/shopspring/decimal"
	"math/big"
	"strings"
)

func Decimal(v interface{}) (ret decimal.Decimal) {
	if v == nil {
		return decimal.Zero
	}
	switch v.(type) {
	case string:
		str := v.(string)
		if len(str) > 2 && strings.ToLower(str)[0:2] == "0x" {
			if s, r := new(big.Int).SetString(str[2:], 16); !r || s == nil {
				return decimal.Zero
			} else {
				return decimal.NewFromBigInt(s, 0)
			}
		}
		ret, _ = decimal.NewFromString(v.(string))
	case uint, uint8, uint16, uint32, uint64, int, int8, int16, int32, int64:
		ret = decimal.NewFromInt(Int64(v))
	case float32, float64:
		ret = decimal.NewFromFloat(Float64(v))
	}
	return
}
