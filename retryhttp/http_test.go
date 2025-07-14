package retryhttp_test

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"testing"

	"github.com/sabariramc/go-kit/retryhttp"
	"gotest.tools/v3/assert"
)

func getServer() *http.Server {
	handler := http.NewServeMux()
	handler.HandleFunc("/echo", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("write"))
	})
	handler.HandleFunc("/error", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})
	handler.HandleFunc("/panic", func(w http.ResponseWriter, r *http.Request) {
		panic("panic")
	})
	return &http.Server{Addr: fmt.Sprintf("%v:%v", "localhost", 63001), Handler: handler}
}

func init() {
	if os.Getenv("LOG_LEVEL") == "" {
		os.Setenv("LOG_LEVEL", "debug")
	}
}

func TestHttpRetry(t *testing.T) {
	srv := getServer()
	go srv.ListenAndServe()
	ctx := context.Background()
	defer srv.Shutdown(ctx)
	client := retryhttp.New()
	res, err := client.Get(ctx, "http://localhost:63001/echo")
	assert.Equal(t, res.StatusCode, http.StatusOK)
	assert.NilError(t, err)
	res, err = client.Get(ctx, "http://localhost:63001/error")
	assert.Equal(t, res.StatusCode, http.StatusInternalServerError)
	assert.NilError(t, err)
	res, err = client.Get(ctx, "http://localhost:63001/fsafsd")
	assert.Equal(t, res.StatusCode, http.StatusNotFound)
	assert.NilError(t, err)
	_, err = client.Get(ctx, "http://localhost:63001/panic")
	assert.Error(t, err, "Get \"http://localhost:63001/panic\": EOF")
}
