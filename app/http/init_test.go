package http_test

import (
	"net/http"
	"testing"

	"github.com/julienschmidt/httprouter"
	"github.com/rs/zerolog"
	srv "github.com/sabariramc/go-kit/app/http"
	"github.com/sabariramc/go-kit/app/http/handler"
	"github.com/sabariramc/go-kit/app/http/middleware"
	"github.com/sabariramc/go-kit/errors"
	"github.com/sabariramc/go-kit/log"
	"gotest.tools/v3/assert"
)

type TestServer struct {
	*srv.Server
	log *log.Logger
}

func new() (*TestServer, error) {
	handler, err := handler.New()
	if err != nil {
		return nil, err
	}
	srv, err := srv.New(func(c *srv.Config) error {
		c.Handler = handler
		return nil
	})
	if err != nil {
		return nil, err
	}
	ts := &TestServer{
		Server: srv,
		log: log.New("TestServer", func(c *log.Config) {
			c.Level = zerolog.DebugLevel
		}),
	}
	ts.registerRoutes(handler)
	return ts, nil
}

func New(t *testing.T) *TestServer {
	ts, err := new()
	assert.NilError(t, err, "Failed to create new TestServer")
	return ts
}

func (s *TestServer) registerRoutes(router *handler.Router) {
	router.Use(middleware.SetCorrelationMiddleware, middleware.RequestTimerMiddleware(s.log), middleware.PanicHandleMiddleware(s.log, nil))
	router.HandlerFunc(http.MethodGet, "/meta/bench", s.benc)
	router.HandlerFunc(http.MethodGet, "/meta/health", s.HealthCheck)
	router.HandlePath("/service/echo", http.HandlerFunc(s.echo))
	router.HandlePath("/service/echo/*params", http.HandlerFunc(s.echo))
	router.HandlerFunc(http.MethodGet, "/error/error500", s.error500)
	router.HandlerFunc(http.MethodGet, "/error/errorWithPanic", s.errorWithPanic)
	router.HandlerFunc(http.MethodGet, "/error/errorUnauthorized", s.errorUnauthorized)
	router.HandlerFunc(http.MethodGet, "/error/panic", s.panic)
}

func (s *TestServer) echo(w http.ResponseWriter, r *http.Request) {
	data, err := s.GetRequestBody(r)
	if err != nil {
		s.WriteErrorResponse(r.Context(), w, &errors.HTTPError{StatusCode: 400, Err: &errors.Error{Code: "invalidJsonBody", Message: err.Error()}})
		return
	}
	s.log.Info(r.Context()).Msg("echo")
	param := httprouter.ParamsFromContext(r.Context())
	s.WriteJSON(r.Context(), w, map[string]any{
		"body":        string(data),
		"headers":     r.Header,
		"queryParams": r.URL.Query(),
		"pathParams":  param,
	})
}

func (s *TestServer) benc(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}

func (s *TestServer) error500(w http.ResponseWriter, r *http.Request) {
	s.WriteErrorResponse(r.Context(), w, &errors.Error{Code: "hello.new.custom.error", Message: "display this"})

}

func (s *TestServer) errorWithPanic(w http.ResponseWriter, r *http.Request) {
	panic(&errors.HTTPError{StatusCode: 503, Err: &errors.Error{Code: "hello.new.custom.error", Message: "display this", Description: map[string]any{"one": "two"}}})
}

func (s *TestServer) panic(w http.ResponseWriter, r *http.Request) {
	panic("fasdfasfsadf")
}

func (s *TestServer) errorUnauthorized(w http.ResponseWriter, r *http.Request) {
	s.WriteErrorResponse(r.Context(), w, &errors.HTTPError{StatusCode: 403, Err: &errors.Error{Code: "hello.new.custom.error", Message: "display this", Description: map[string]any{"one": "two"}}})
}
