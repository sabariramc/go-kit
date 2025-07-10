package errors_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sabariramc/go-kit/errors/internal"
	"gotest.tools/v3/assert"
)

func TestWriteError(t *testing.T) {
	mux := internal.NewServer()
	req, err := http.NewRequest("GET", "/error", nil)
	assert.NilError(t, err)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	assert.Equal(t, w.Code, http.StatusBadRequest)
	assert.Equal(t, w.Body.String(), `{"code":"BAD_REQUEST","message":"Invalid input"}`)
	assert.Equal(t, w.Header().Get("Content-Type"), "application/json")
	req, err = http.NewRequest("GET", "/custom-error", nil)
	assert.NilError(t, err)
	w = httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	assert.Equal(t, w.Code, http.StatusInternalServerError)
	assert.Equal(t, w.Body.String(), `{"code":"CUSTOM_ERROR","message":"Custom error occurred","description":"description"}`)
	assert.Equal(t, w.Header().Get("Content-Type"), "application/json")
	req, err = http.NewRequest("GET", "/internal-error", nil)
	assert.NilError(t, err)
	w = httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	assert.Equal(t, w.Code, http.StatusInternalServerError)
	assert.Equal(t, w.Body.String(), `{"code":"CUSTOM_ERROR","message":"Custom error occurred","description":{"nextStep":"check input"}}`)
	assert.Equal(t, w.Header().Get("Content-Type"), "application/json")
}
