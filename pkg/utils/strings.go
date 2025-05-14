package utils

import (
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func Capitalize(s string) string {
	return cases.Title(language.English).String(strings.ToLower(s))
}
