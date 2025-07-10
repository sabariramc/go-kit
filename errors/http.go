package errors

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
)

func WriteError(ctx context.Context, w http.ResponseWriter, err error) {
	var httpErr *HTTPError
	var customErr *Error
	if errors.As(err, &httpErr) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(httpErr.StatusCode)
		blob, _ := json.Marshal(httpErr)
		w.Write(blob)
		return
	} else if errors.As(err, &customErr) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		blob, _ := json.Marshal(customErr)
		w.Write(blob)
		return
	} else {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
}
