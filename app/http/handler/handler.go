package handler

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/sabariramc/go-kit/app/http/middleware"
)

type Router struct {
	*httprouter.Router
	middleware      []middleware.Middleware
	builtMiddleware []middleware.Middleware
}

func New(opt ...Option) (*Router, error) {
	cfg, err := NewConfig(opt...)
	if err != nil {
		return nil, err
	}
	return &Router{Router: cfg.Router}, nil
}

func (r *Router) Use(middleware ...middleware.Middleware) {
	r.middleware = append(r.middleware, middleware...)
	r.buildMiddleware()
}

func (r *Router) buildMiddleware() {
	r.builtMiddleware = make([]middleware.Middleware, len(r.middleware))
	for i, j := 0, len(r.middleware)-1; i <= j; i, j = i+1, j-1 {
		r.builtMiddleware[j] = r.middleware[i]
		r.builtMiddleware[i] = r.middleware[j]
	}
}

func (r *Router) getHandler(handler http.Handler) http.Handler {
	for _, mw := range r.builtMiddleware {
		handler = mw(handler)
	}
	return handler
}

func (r *Router) Handler(method, path string, handler http.Handler) {
	r.Router.Handler(method, path, r.getHandler(handler))
}

func (r *Router) HandlerFunc(method, path string, handler http.HandlerFunc) {
	r.Router.Handler(method, path, r.getHandler(handler))
}

func (r *Router) HandlePath(path string, handler http.Handler) {
	r.Handler(http.MethodGet, path, handler)
	r.Handler(http.MethodPost, path, handler)
	r.Handler(http.MethodPut, path, handler)
	r.Handler(http.MethodDelete, path, handler)
	r.Handler(http.MethodPatch, path, handler)
}
