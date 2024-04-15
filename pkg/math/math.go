package math

import "math"

func MinInt64(values ...int64) int64 {
	m := int64(math.MaxInt64)
	for _, v := range values {
		if v < m {
			m = v
		}
	}
	return m
}
