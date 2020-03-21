package utils

import "encoding/json"

func JsonDumpsToStringSilently(value interface{}, defaultValue string) string {
	b, err := json.Marshal(value)
	if err != nil {
		return defaultValue
	}
	return string(b)
}
