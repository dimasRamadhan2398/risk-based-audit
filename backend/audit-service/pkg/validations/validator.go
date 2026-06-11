package validations

import (
	"github.com/go-playground/validator/v10"
)

// Validator wraps the go-playground validator
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
func (v *Validator) Validate(s interface{}) error {
	return v.validate.Struct(s)
}

// Engine returns the underlying validator engine
func (v *Validator) Engine() interface{} {
	return v.validate
}
