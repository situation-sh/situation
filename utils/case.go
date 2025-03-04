package utils

import (
	"regexp"
	"strings"
)

func ConvertCamelToSnake(s string) string {
	// Match uppercase letters that are preceded by a lowercase letter or a number
	re := regexp.MustCompile(`([a-z0-9])([A-Z])`)

	// Insert an underscore before the uppercase letter and convert to lowercase
	snake := re.ReplaceAllString(s, "${1}_${2}")

	// Convert the entire string to lowercase
	return strings.ToLower(snake)
}
