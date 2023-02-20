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

func PtrIntToPtrFloat(i *int) *float32 {
	if i == nil {
		return nil
	}
	f := float32(*i)
	return &f
}

func PtrInt64ToPtrFloat(i *int64) *float32 {
	if i == nil {
		return nil
	}
	f := float32(*i)
	return &f
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

func StringSliceToStringSpacePtr(l []string) *string {
	if len(l) == 0 {
		return nil
	}
	s := strings.Join(l, " ")
	return &s
}

func StringSliceToLinedString(l []string) string {
	str := ""
	for _, s := range l {
		str = str + s + "\n"
	}
	return str
}

func StringSliceToLinedStringPtr(l []string) *string {
	if len(l) == 0 {
		return nil
	}
	s := StringSliceToLinedString(l)
	return &s
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

func SliceToStringCommaListPtr(s []string) *string {
	if len(s) == 0 {
		return nil
	}
	r := SliceToStringCommaList(s)
	return &r
}

func StringSemiColonListToSlice(s string) []string {
	if s == "" {
		return []string{}
	}
	return strings.Split(s, ";")
}

func SliceToStringSemiColonList(s []string) string {
	return strings.Join(s, ";")
}
func SliceToStringSemiColonListPtr(s []string) *string {
	if len(s) == 0 {
		return nil
	}
	r := SliceToStringSemiColonList(s)
	return &r
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
