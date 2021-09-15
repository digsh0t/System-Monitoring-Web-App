package utils

import "strings"

func FindInStringArray(slice []string, val string) bool {
	for _, item := range slice {
		if item == val {
			return true
		}
	}
	return false
}

func ExtractSubString(str string, start string, end string) (result string) {
	s := strings.Index(str, start)
	if s == -1 {
		return
	}
	s += len(start)

	e := strings.Index(str, end)
	if s == -1 {
		return
	}
	return str[s:e]
}
