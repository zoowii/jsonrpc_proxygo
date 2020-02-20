package utils

// 返回 val1 or val2
func StringOrElse(val1 string, val2 string) string {
	if len(val1) > 0 {
		return val1
	}
	return val2
}
