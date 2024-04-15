package conv

import "fmt"

// 十进制转十六进制
func Hex(v int64) string {
	return fmt.Sprintf("0x%x", v)
}
