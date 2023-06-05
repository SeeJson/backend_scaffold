package utils

import "encoding/json"

// UnmarshalJSON from string
func UnmarshalJSON(s string, dst interface{}) error {
	return json.Unmarshal([]byte(s), dst)
}

func MarshalJSON(v interface{}) string {
	d, _ := json.Marshal(v)
	return string(d)
}
