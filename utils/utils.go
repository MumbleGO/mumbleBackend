package utils

import (
	"encoding/json"
	"net/http"
)

func WriteJson(w http.ResponseWriter, status int, v any) error {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(v)
}

func MakeHTTPHandleFunc(f apiFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := f(w, r); err != nil {
			WriteJson(w, http.StatusBadRequest, ApiError{ErrorMessage: err.Error()})
		}
	}
}

type (
	apiFunc  func(w http.ResponseWriter, r *http.Request) error
	ApiError struct {
		ErrorMessage string `json:"error"`
	}
)
