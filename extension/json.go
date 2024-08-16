package extension

import "encoding/json"

func Jsonify(t any) string {
	bytes, _ := json.Marshal(t)
	return string(bytes)
}
