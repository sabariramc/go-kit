package http_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"runtime"
	"testing"

	"github.com/rs/zerolog"
	"github.com/sabariramc/go-kit/log"
	"github.com/sabariramc/go-kit/log/correlation"
	srv "github.com/sabariramc/go-kit/server/http"
	"github.com/sabariramc/go-kit/server/http/handler"
	"gotest.tools/v3/assert"
)

func TestRouter(t *testing.T) {
	srv := New(t)
	req := httptest.NewRequest(http.MethodGet, "/service/echo", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	blob, _ := io.ReadAll(w.Body)
	assert.Equal(t, string(blob), `{"body":"","headers":{},"pathParams":null,"queryParams":{}}`)
	assert.Equal(t, w.Result().StatusCode, http.StatusOK)
	req = httptest.NewRequest(http.MethodGet, "/service/abc/search", nil)
	w = httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	blob, _ = io.ReadAll(w.Body)
	assert.Equal(t, w.Result().StatusCode, http.StatusNotFound)
	res := make(map[string]any)
	json.Unmarshal(blob, &res)
	expectedResponse := map[string]any{"error": map[string]any{"description": map[string]any{"path": "/service/abc/search"}, "message": "URL Not Found", "code": "URL_NOT_FOUND"}}
	assert.DeepEqual(t, res, expectedResponse)
	req = httptest.NewRequest(http.MethodGet, "/service/echo/tenant_ABC4567890abc", nil)
	w = httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	blob, _ = io.ReadAll(w.Body)
	assert.Equal(t, w.Result().StatusCode, http.StatusOK)
	assert.Equal(t, string(blob), `{"body":"","headers":{},"pathParams":[{"Key":"params","Value":"/tenant_ABC4567890abc"}],"queryParams":{}}`)
	req = httptest.NewRequest(http.MethodPost, "/meta/bench", nil)
	w = httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	blob, _ = io.ReadAll(w.Body)
	res = make(map[string]any)
	json.Unmarshal(blob, &res)
	expectedResponse = map[string]any{"error": map[string]any{"description": map[string]any{"method": "POST", "path": "/meta/bench"}, "message": "Method Not Allowed", "code": "METHOD_NOT_ALLOWED"}}
	assert.Equal(t, w.Result().StatusCode, http.StatusMethodNotAllowed)
	assert.DeepEqual(t, res, expectedResponse)
}

func TestRouterCustomError(t *testing.T) {
	srv := New(t)
	req := httptest.NewRequest(http.MethodGet, "/error/error500", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	blob, _ := io.ReadAll(w.Body)
	res := make(map[string]any)
	json.Unmarshal(blob, &res)
	fmt.Println(string(blob))
	expectedResponse := map[string]any{"error": map[string]any{"message": "display this", "code": "hello.new.custom.error"}}
	assert.Equal(t, w.Result().StatusCode, http.StatusInternalServerError)
	assert.DeepEqual(t, res, expectedResponse)
}

func TestRouterPanic(t *testing.T) {
	srv := New(t)
	req := httptest.NewRequest(http.MethodGet, "/error/errorWithPanic", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	blob, _ := io.ReadAll(w.Body)
	res := make(map[string]any)
	json.Unmarshal(blob, &res)
	expectedResponse := map[string]any{"error": map[string]any{"description": map[string]any{"one": string("two")}, "message": "display this", "code": "hello.new.custom.error"}}
	assert.Equal(t, w.Result().StatusCode, http.StatusServiceUnavailable)
	assert.DeepEqual(t, res, expectedResponse)
}

func TestRouterPanic2(t *testing.T) {
	srv := New(t)
	req := httptest.NewRequest(http.MethodGet, "/error/panic", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	blob, _ := io.ReadAll(w.Body)
	res := make(map[string]any)
	json.Unmarshal(blob, &res)
	expectedResponse := map[string]any{"error": map[string]any{"message": "Unknown error", "code": "INTERNAL_SERVER_ERROR"}}
	assert.Equal(t, w.Result().StatusCode, http.StatusInternalServerError)
	assert.DeepEqual(t, res, expectedResponse)
}

func TestRouterClientError(t *testing.T) {
	srv := New(t)
	req := httptest.NewRequest(http.MethodGet, "/error/errorUnauthorized", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	blob, _ := io.ReadAll(w.Body)
	res := make(map[string]any)
	json.Unmarshal(blob, &res)
	fmt.Println(string(blob))
	expectedResponse := map[string]any{"error": map[string]any{"message": "display this", "description": map[string]any{"one": string("two")}, "code": "hello.new.custom.error"}}
	assert.Equal(t, w.Result().StatusCode, 403)
	assert.DeepEqual(t, res, expectedResponse)
}

func TestRouterHealthCheck(t *testing.T) {
	srv := New(t)
	req := httptest.NewRequest(http.MethodGet, "/meta/health", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	assert.Equal(t, w.Result().StatusCode, 204)
}

func TestPost(t *testing.T) {
	srv := New(t)
	payload, _ := json.Marshal(map[string]string{"fasdfas": "FASDFASf"})
	buff := bytes.NewBuffer(payload)
	req := httptest.NewRequest(http.MethodPost, "/service/echo", buff)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	blob, _ := io.ReadAll(w.Body)
	expectedResponse := map[string]any{
		"body":        `{"fasdfas":"FASDFASf"}`,
		"headers":     map[string]any{},
		"pathParams":  nil,
		"queryParams": map[string]any{},
	}
	res := make(map[string]any)
	json.Unmarshal(blob, &res)
	assert.Equal(t, w.Result().StatusCode, http.StatusOK)
	assert.DeepEqual(t, res, expectedResponse)
}

const (
	start = 1 // actual = start  * goprocs
	end   = 8 // actual = end    * goprocs
	step  = 1
)

var goprocs = runtime.GOMAXPROCS(0) // 8

func TestBencRoute(t *testing.T) {
	srv := New(t)
	payload, _ := json.Marshal(map[string]string{"fasdfas": "FASDFASf"})
	buff := bytes.NewBuffer(payload)
	req := httptest.NewRequest(http.MethodGet, "/meta/bench", buff)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	assert.Equal(t, w.Result().StatusCode, http.StatusNoContent)
}

func BenchmarkRoutes(b *testing.B) {
	handler, _ := handler.New()
	base, _ := srv.New(func(c *srv.Config) error {
		c.Handler = handler
		return nil
	})
	srv := &TestServer{
		Server: base,
		log: log.New("TestServer", func(c *log.Config) {
			c.Level = zerolog.DebugLevel
			c.Target = io.Discard
		}),
	}
	srv.registerRoutes(handler)
	payload, _ := json.Marshal(map[string]string{"fasdfas": "FASDFASf"})
	buff := bytes.NewBuffer(payload)
	req := httptest.NewRequest(http.MethodGet, "/meta/bench", buff)
	req.Header.Set(correlation.CorrelationIDHeader, "test-correlation-id")
	for i := start; i < end; i += step {
		b.Run(fmt.Sprintf("goroutines-%d", i*goprocs), func(b *testing.B) {
			b.SetParallelism(i)
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					w := httptest.NewRecorder()
					srv.ServeHTTP(w, req)
					if w.Result().StatusCode != http.StatusNoContent {
						b.Fatalf("Expected status code %d, got %d", http.StatusNoContent, w.Result().StatusCode)
					}
				}
			})
		})
	}
}
