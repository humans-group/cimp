package cimp

import (
	"regexp"
	"strings"
)

func makeFullKey(prefix string, key string) string {
	key = ToSnakeCase(key)
	if len(prefix) > 0 {
		key = prefix + sep + key
	}

	return key
}

var matchFirstCap = regexp.MustCompile("(.)([A-Z][a-z]+)")
var matchAllCap = regexp.MustCompile("([a-z0-9])([A-Z])")
var matchAllSpecSymbols = regexp.MustCompile("[^A-z0-9]")
var matchAllMultipleUnderscore = regexp.MustCompile("[_]{2,}")

func ToSnakeCase(str string) string {
	str = matchAllSpecSymbols.ReplaceAllString(str, "_")
	str = matchFirstCap.ReplaceAllString(str, "${1}_${2}")
	str = matchAllCap.ReplaceAllString(str, "${1}_${2}")
	str = matchAllMultipleUnderscore.ReplaceAllString(str, "_")
	str = strings.Trim(str, "_")

	return strings.ToLower(str)
}
