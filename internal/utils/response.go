package utils

import (
	"encoding/json"
	"net/http"

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
	errors := make(map[string]string)
	for _, e := range err.(validator.ValidationErrors) {
		errors[e.Field()] = "Invalid value (" + e.Tag() + ")"
	}
	JSON(w, http.StatusBadRequest, map[string]interface{}{
		"error":   "validation failed",
		"details": errors,
	})
}
