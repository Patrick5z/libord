package conv

import (
	"encoding/json"
)

// Array json str to array
func Array(strOrBytes any) (ret []any) {
	if strOrBytes == nil {
		return
	}
	switch strOrBytes.(type) {
	case []byte:
		json.Unmarshal(strOrBytes.([]byte), &ret)
	case string:
		json.Unmarshal([]byte(strOrBytes.(string)), &ret)
	}

	return
}
