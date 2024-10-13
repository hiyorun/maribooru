package validation

import (
	"regexp"

	"github.com/go-playground/validator"
)

func ValidateSlug(fl validator.FieldLevel) bool {
	slug := fl.Field().String()
	match, _ := regexp.MatchString("^[a-zA-Z0-9_()]+$", slug)
	return match
}
