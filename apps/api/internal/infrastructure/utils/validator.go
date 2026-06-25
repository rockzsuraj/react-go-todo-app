package utils

import (
	"log"

	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

func InitValidator() {
	validate = validator.New()
	log.Println("Validator initialized")
}

func GetValidator() *validator.Validate {
	if validate == nil {
		InitValidator()
	}
	return validate
}

func ValidateStruct(s interface{}) error {
	v := GetValidator()
	return v.Struct(s)
}
