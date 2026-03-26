package util

import "strconv"

func ParseInt(str string) int {
	limitI64, _ := strconv.ParseInt(str, 10, 64)
	return int(limitI64)
}

func ParseBool(str string) bool {
	if str == "" {
		return false
	}
	v, _ := strconv.ParseBool(str)
	return v
}
