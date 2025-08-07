package http

import (
	"net/http"
)

// HealthCheck handles the HTTP request for the health check endpoint. It runs the health check and returns a 500 status code if there is an error, otherwise it returns a 204 status code.
func (h *Server) HealthCheck(w http.ResponseWriter, r *http.Request) {
	err := h.base.RunHealthCheck(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
