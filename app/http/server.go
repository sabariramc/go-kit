package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/sabariramc/go-kit/app/base"
	"github.com/sabariramc/go-kit/log"
)

type Server struct {
	*http.Server
	handler http.Handler
	base    *base.Base
	log     *log.Logger
}

func New(opt ...Option) (*Server, error) {
	cfg, err := NewConfig(opt...)
	if err != nil {
		return nil, err
	}
	h := &Server{
		base:    cfg.Base,
		handler: cfg.Server.Handler,
		log:     cfg.Log,
		Server:  cfg.Server,
	}
	return h, nil
}

func (h *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.handler.ServeHTTP(w, r)
}

func (h *Server) Close(ctx context.Context) error {
	return h.Server.Shutdown(ctx)
}

func (h *Server) ListenAndServe() {
	h.log.Info(context.Background()).Msgf("Server starting at %v", h.Server.Addr)
	go h.base.StartSignalMonitor(context.Background())
	err := h.Server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		h.log.Error(context.Background()).Err(err).Msg("Server crashed")
	}
	h.base.AwaitShutdownCompletion()
}

func (h *Server) CopyRequestBody(r *http.Request) ([]byte, error) {
	blobBody, err := h.GetRequestBody(r)
	if err != nil {
		return blobBody, fmt.Errorf("HttpServer.CopyRequestBody: %w", err)
	}
	r.Body = io.NopCloser(bytes.NewReader(blobBody))
	return blobBody, nil
}

func (h *Server) ExtractRequestMetadata(r *http.Request) map[string]any {
	res := map[string]any{
		"Method":        r.Method,
		"Header":        r.Header,
		"URL":           r.URL,
		"Proto":         r.Proto,
		"ContentLength": r.ContentLength,
		"Host":          r.Host,
		"RemoteAddr":    r.RemoteAddr,
		"RequestURI":    r.RequestURI,
	}
	return res
}

func (h *Server) GetRequestBody(r *http.Request) ([]byte, error) {
	if r.ContentLength <= 0 {
		return nil, nil
	}
	body := r.Body
	defer body.Close()
	blobBody, err := io.ReadAll(body)
	if err != nil {
		err = fmt.Errorf("Server.GetRequestBody: error reading request body: %w", err)
	}
	return blobBody, err
}

func (h *Server) LoadRequestJSONBody(r *http.Request, body any) error {
	blobBody, err := h.GetRequestBody(r)
	if err != nil {
		return fmt.Errorf("Server.LoadJSONBody: %w", err)
	}
	err = json.Unmarshal(blobBody, body)
	if err != nil {
		err = fmt.Errorf("Server.LoadJSONBody: error loading request body to object: %w", err)
	}
	return err
}
