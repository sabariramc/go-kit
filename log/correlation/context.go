package correlation

import (
	"context"
	"net/http"
)

type ContextKey string

var (
	contextKeyCorrelation ContextKey = ContextKey("ContextKeyCorrelation")
)

// GetContextWithCorrelationParam returns a context.Context with the provided CorrelationParam.
func GetContextWithCorrelationParam(ctx context.Context, c *EventCorrelation) context.Context {
	ctx = context.WithValue(ctx, &contextKeyCorrelation, c)
	return ctx
}

// ExtractCorrelationParam retrieves the CorrelationParam stored within the context.Context.
func ExtractCorrelationParam(ctx context.Context) (*EventCorrelation, bool) {
	iVal := ctx.Value(&contextKeyCorrelation)
	if iVal == nil {
		return &EventCorrelation{}, false
	}
	val, ok := iVal.(*EventCorrelation)
	if !ok {
		return &EventCorrelation{}, false
	}
	return val, true
}

// SetCorrelationHeader adds the CorrelationParam and UserIdentifier from the context.Context into the http.Request Header.
// These values are marshalled with the header struct tag.
func SetCorrelationHeader(ctx context.Context, req *http.Request) {
	if corr, ok := ExtractCorrelationParam(ctx); ok && corr != nil {
		headers := corr.GetHeader()
		for i, v := range headers {
			req.Header.Add(i, v)
		}
	}
}
