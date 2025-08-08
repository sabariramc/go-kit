package route

import (
	"net/http"

	"github.com/sabariramc/go-kit/app/http/constant"
	"github.com/sabariramc/go-kit/app/http/errorhandler"
	"github.com/sabariramc/go-kit/errors"
)

// NotFound returns a handler function for responding with a 404 Not Found status.
func NotFound() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add(constant.HTTPHeaderContentType, constant.HTTPContentTypeJSON)
		w.WriteHeader(http.StatusNotFound)
		err := &errors.Error{
			Code:    "URL_NOT_FOUND",
			Message: "URL Not Found",
			Description: map[string]any{
				"path": r.URL.Path,
			},
		}
		_, body := errorhandler.Handle(r.Context(), err)
		w.Write(body)
	}
}

// MethodNotAllowed returns a handler function for responding with a 405 Method Not Allowed status.
func MethodNotAllowed() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add(constant.HTTPHeaderContentType, constant.HTTPContentTypeJSON)
		w.WriteHeader(http.StatusMethodNotAllowed)
		err := &errors.Error{
			Code:    "METHOD_NOT_ALLOWED",
			Message: "Method Not Allowed",
			Description: map[string]any{
				"path":   r.URL.Path,
				"method": r.Method,
			},
		}
		_, body := errorhandler.Handle(r.Context(), err)
		w.Write(body)
	}
}
