package utils

import "strings"

func StringLinedToSlice(s string) []string {
	return strings.Split(s, "\n")
}
