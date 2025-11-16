package validator

import (
	"gojwt-rest-api/internal/domain"
	"strings"

	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	en_translations "github.com/go-playground/validator/v10/translations/en"
)

// Validator is a wrapper for go-playground validator
type Validator struct {
	validate *validator.Validate
	trans    ut.Translator
}

// New creates a new validator instance
func New() (*Validator, error) {
	validate := validator.New()
	english := en.New()
	uni := ut.New(english, english)
	trans, _ := uni.GetTranslator("en")
		if err := en_translations.RegisterDefaultTranslations(validate, trans); err != nil {
		return nil, err
	}

	return &Validator{
		validate: validate,
		trans:    trans,
	}, nil
}

// Validate validates a struct and returns a slice of validation errors
func (v *Validator) Validate(data interface{}) []domain.ValidationError {
	var validationErrors []domain.ValidationError

	err := v.validate.Struct(data)
	if err != nil {
		for _, err := range err.(validator.ValidationErrors) {
			validationErrors = append(validationErrors, domain.ValidationError{
				Field: strings.ToLower(err.Field()),
				Error: err.Translate(v.trans),
			})
		}
	}

	return validationErrors
}
