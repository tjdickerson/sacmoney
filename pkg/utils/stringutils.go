package utils

import (
	"strings"
)

func IndexOfString(s string, value string) int {
	return strings.IndexAny(s, value)
}
