package util

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

func NewValidator() *validator.Validate {
	validate := validator.New()

	validate.RegisterValidation("no_consecutive_spaces", func(fl validator.FieldLevel) bool {
		s := fl.Field().String()

		return !strings.Contains(s, "  ")
	})

	return validate
}

func FormatValidationErrors(errs validator.ValidationErrors) map[string]string {
	errorMessages := make(map[string]string)
	
	for _, err := range errs {
		fieldName := strings.ToLower(err.Field())
		
		// Kita buat ini lebih cerdas untuk menangani banyak aturan
		switch err.Tag() {
		case "required":
			errorMessages[fieldName] = "Field ini wajib diisi."
		case "email":
			errorMessages[fieldName] = "Format email tidak valid."
		case "min":
			errorMessages[fieldName] = fmt.Sprintf("Minimal %s karakter.", err.Param())
		case "max":
			errorMessages[fieldName] = fmt.Sprintf("Maksimal %s karakter.", err.Param())
		case "no_consecutive_spaces":
			errorMessages[fieldName] = "Tidak boleh mengandung spasi berturut-turut."
		default:
			errorMessages[fieldName] = fmt.Sprintf("Input tidak valid (%s).", err.Tag())
		}
	}
	return errorMessages
}