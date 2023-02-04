package service

import "strings"

func IntToBool(i int) bool {
	return i != 0
}

func BooleanIntegerConversion(i *float32) bool {
	if i == nil {
		return false
	}
	return *i != 0
}

func StringSpaceListToSlice(s string) []string {
	if s == "" {
		return []string{}
	}
	return strings.Split(s, " ")
}

func StringSpacePtrListToSlice(s *string) []string {
	if s == nil {
		return []string{}
	}
	return StringSpaceListToSlice(*s)
}