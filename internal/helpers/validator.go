package helpers

import (
	"github.com/go-playground/validator"
)

type Validator struct {
	validator *validator.Validate
}

func NewValidator(validator *validator.Validate) *Validator {
	return &Validator{
		validator: validator,
	}
}

func (cv *Validator) Validate(i interface{}) error {
	if err := cv.validator.Struct(i); err != nil {
		return err
	}
	return nil
}
