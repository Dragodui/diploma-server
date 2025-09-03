package utils

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"

	"github.com/go-playground/validator/v10"
)

func JSON(w http.ResponseWriter, code int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(data)
}

func JSONError(w http.ResponseWriter, message string, code int) {
	JSON(w, code, map[string]string{"error": message})
}

func JSONValidationErrors(w http.ResponseWriter, err error) {
	// validator.ValidationErrors is necessary
	ve := err.(validator.ValidationErrors)

	details := make(map[string]string, len(ve))
	for _, e := range ve {
		field := e.Field()
		kind := e.Kind()
		param := e.Param()

		var msg string
		switch e.Tag() {
		case "required":
			msg = fmt.Sprintf("%s is required", field)

		case "email":
			msg = fmt.Sprintf("%s must be a valid email address", field)

		case "min":
			if kind == reflect.String {
				msg = fmt.Sprintf("%s must be at least %s characters long", field, param)
			} else {
				msg = fmt.Sprintf("%s must be at least %s", field, param)
			}

		case "max":
			if kind == reflect.String {
				msg = fmt.Sprintf("%s cannot be longer than %s characters", field, param)
			} else {
				msg = fmt.Sprintf("%s cannot be greater than %s", field, param)
			}

		default:
			msg = fmt.Sprintf("%s is not valid (%s)", field, e.Tag())
		}

		details[field] = msg
	}

	JSON(w, http.StatusBadRequest, map[string]interface{}{
		"error":   "validation failed",
		"details": details,
	})
}
