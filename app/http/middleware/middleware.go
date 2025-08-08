package middleware

import (
	"fmt"
	"net/http"
	"runtime/debug"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sabariramc/go-kit/app/base"
	"github.com/sabariramc/go-kit/app/http/errorhandler"
	span "github.com/sabariramc/go-kit/instrumentation"
	"github.com/sabariramc/go-kit/log"
	"github.com/sabariramc/go-kit/log/correlation"
)

type Middleware func(http.Handler) http.Handler

func GetCorrelationParam(r *http.Request) *correlation.EventCorrelation {
	headers := r.Header
	corr := &correlation.EventCorrelation{}
	for key, values := range headers {
		switch strings.ToLower(key) {
		case strings.ToLower(correlation.CorrelationIDHeader):
			corr.CorrelationID = values[0]
		case strings.ToLower(correlation.ScenarioIDHeader):
			corr.ScenarioID = values[0]
		case strings.ToLower(correlation.SessionIDHeader):
			corr.SessionID = values[0]
		case strings.ToLower(correlation.ScenarioNameHeader):
			corr.ScenarioName = values[0]
		}
	}
	if corr.CorrelationID == "" {
		serviceName := base.GetServiceName()
		corr.CorrelationID = serviceName + "_" + uuid.New().String()
	}
	return corr
}

func SetCorrelationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		corr := GetCorrelationParam(r)
		ctx := correlation.GetContextWithCorrelationParam(r.Context(), corr)
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}

func RequestTimerMiddleware(log *log.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			st := time.Now()
			next.ServeHTTP(w, r)
			log.Info(r.Context()).Str("method", r.Method).Str("url", r.URL.Path).Dur("latencyInMS", time.Duration(time.Since(st).Milliseconds())).Msg("Request completed")
		})
	}
}

func PanicHandleMiddleware(log *log.Logger, tr span.SpanOp) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					stackTrace := string(debug.Stack())
					log.Error(r.Context()).Any("panic", rec).Str("stacktrace", stackTrace).Msg("Panic recovered")
					var err error
					err, ok := rec.(error)
					if !ok {
						err = fmt.Errorf("error occurred during request processing")
					}
					statusCode, body := errorhandler.Handle(r.Context(), err)
					if tr != nil {
						sp, ok := tr.GetSpanFromContext(r.Context())
						if ok {
							sp.SetError(err, stackTrace)
							sp.SetStatus(statusCode, http.StatusText(statusCode))
						}
					}
					w.WriteHeader(statusCode)
					w.Write(body)
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}
