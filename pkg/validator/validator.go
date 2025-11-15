package validator

import (
	"github.com/go-playground/validator/v10"
)

// Validator is a wrapper for go-playground validator
type Validator struct {
	validate *validator.Validate
}

// New creates a new validator instance
func New() *Validator {
	return &Validator{
		validate: validator.New(),
	}
}

// Validate validates a struct
func (v *Validator) Validate(data interface{}) error {
	return v.validate.Struct(data)
}
