package service

import (
	"strings"
)

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

func StringCommaListToSlice(s string) []string {
	if s == "" {
		return []string{}
	}
	return strings.Split(s, ",")
}

func SliceToStringCommaList(s []string) string {
	return strings.Join(s, ",")
}

func StringSemiColonListToSlice(s string) []string {
	if s == "" {
		return []string{}
	}
	return strings.Split(s, ";")
}

func StringSemiColonPtrListToSlice(s *string) []string {
	if s == nil {
		return []string{}
	}
	return StringSemiColonListToSlice(*s)
}

func StringCommaPtrListToSlice(s *string) []string {
	if s == nil {
		return []string{}
	}
	return StringCommaListToSlice(*s)
}

func PtrStringToString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func PtrIntToInt(i *int) int {
	if i == nil {
		return 0
	}
	return *i
}

func PtrFloatToInt64(i *float32) int64 {
	if i == nil {
		return 0
	}
	return int64(*i)
}
