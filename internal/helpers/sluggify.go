package helpers

import (
	"regexp"
	"strings"
)

func Sluggify(input string) string {
	slug := strings.ToLower(input)

	slug = RemoveSpaces(slug)

	reg := regexp.MustCompile("[^a-z0-9_()]")
	slug = reg.ReplaceAllString(slug, "")

	slug = RemoveDoubleUnderscores(slug)

	return slug
}

func RemoveSpaces(input string) string {
	return strings.ReplaceAll(input, " ", "_")
}

func RemoveDoubleUnderscores(input string) string {
	reg := regexp.MustCompile("__+")
	return reg.ReplaceAllString(input, "_")
}
