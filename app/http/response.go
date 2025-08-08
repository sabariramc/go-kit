// package http provides utilities for managing an HTTP server, including logging and response handling.
package http

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/sabariramc/go-kit/app/http/constant"
	"github.com/sabariramc/go-kit/app/http/errorhandler"
)

// ResponseWriter is a custom response writer that logs responses and request bodies.
type ResponseWriter struct {
	Status              int    // HTTP status code
	Body                []byte // Pointer to the request body
	http.ResponseWriter        // Embedded Gin response writer
}

// WriteHeader logs the status code and writes the header to the response.
func (w *ResponseWriter) WriteHeader(code int) {
	w.Status = code
	w.ResponseWriter.WriteHeader(code)
}

// Write logs the response body and writes it to the response.
func (w *ResponseWriter) Write(body []byte) (int, error) {
	if w.Status == 0 {
		w.Status = http.StatusOK
	}
	w.Body = body
	return w.ResponseWriter.Write(body)
}

// WriteJSON writes a JSON response with a status code of 200 OK.
func (h *Server) WriteJSON(ctx context.Context, w http.ResponseWriter, responseBody any) {
	h.WriteJSONWithStatusCode(ctx, w, http.StatusOK, responseBody)
}

// WriteJSONWithStatusCode writes a JSON response with the specified status code.
func (h *Server) WriteJSONWithStatusCode(ctx context.Context, w http.ResponseWriter, statusCode int, responseBody any) {
	var err error
	blob, ok := responseBody.([]byte)
	if !ok {
		blob, err = json.Marshal(responseBody)
		if err != nil {
			h.log.Error(ctx).Err(err).Msg("Error in response json marshall")
			statusCode = http.StatusInternalServerError
			blob = []byte("{\"error\": \"Internal server error\"}")
		}
	}
	h.WriteResponseWithStatusCode(ctx, w, statusCode, constant.HTTPContentTypeJSON, blob)
}

// WriteResponse writes a response with a status code of 200 OK and the specified content type.
func (h *Server) WriteResponse(ctx context.Context, w http.ResponseWriter, contentType string, responseBody []byte) {
	h.WriteResponseWithStatusCode(ctx, w, http.StatusOK, contentType, responseBody)
}

// WriteErrorResponse writes an error response, logging the error and stack trace.
func (h *Server) WriteErrorResponse(ctx context.Context, w http.ResponseWriter, err error) {
	statusCode, body := errorhandler.Handle(ctx, err)
	h.WriteJSONWithStatusCode(ctx, w, statusCode, body)
}

// WriteResponseWithStatusCode writes a response with the specified status code and content type.
func (h *Server) WriteResponseWithStatusCode(ctx context.Context, w http.ResponseWriter, statusCode int, contentType string, responseBody []byte) {
	w.Header().Set(constant.HTTPHeaderContentType, contentType)
	w.WriteHeader(statusCode)
	w.Write(responseBody)
}
