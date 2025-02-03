package validation

import "github.com/go-playground/validator/v10"

type Validator interface {
	Validate(interface{}) error
}

type RequestValidator struct {
	validator *validator.Validate
}

func NewRequestValidator() *RequestValidator {
	return &RequestValidator{
		validator: validator.New(),
	}
}

func (v *RequestValidator) Validate(obj interface{}) error {
	return v.validator.Struct(obj)
}
