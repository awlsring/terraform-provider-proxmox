package utils

import (
	"net/url"
	"strings"
)

func StringLinedToSlice(s string) []string {
	return strings.Split(s, "\n")
}

func StringToStringList(stringList string) []string {
	var l []string
	for _, str := range strings.Split(stringList, "\n") {
		l = append(l, str)
	}
	return l
}

func DecodeStringList(s *string) []string {
	if s == nil {
		return nil
	}
	un := strings.Replace(*s, "%20", "+", -1)
	decodedKey, _ := url.QueryUnescape(un)
	return StringToStringList(decodedKey)
}
