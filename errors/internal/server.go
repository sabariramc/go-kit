package internal

import (
	"net/http"

	"github.com/sabariramc/go-kit/errors"
)

func NewServer() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/error", func(w http.ResponseWriter, r *http.Request) {
		errors.WriteError(r.Context(), w, errors.ErrBadRequest)
	})
	mux.HandleFunc("/custom-error", func(w http.ResponseWriter, r *http.Request) {
		customErr := &errors.Error{"CUSTOM_ERROR", "Custom error occurred", "description"}
		errors.WriteError(r.Context(), w, customErr)
	})
	mux.HandleFunc("/internal-error", func(w http.ResponseWriter, r *http.Request) {
		customErr := &errors.Error{"CUSTOM_ERROR", "Custom error occurred", map[string]string{"nextStep": "check input"}}
		errors.WriteError(r.Context(), w, customErr)
	})
	return mux
}
