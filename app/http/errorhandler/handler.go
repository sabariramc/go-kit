package errorhandler

import (
	"context"
	e "errors"
	"net/http"

	"github.com/sabariramc/go-kit/errors"
)

func Handle(ctx context.Context, err error) (int, []byte) {
	var statusCode int
	var body []byte
	statusCode = http.StatusInternalServerError
	var custErr *errors.Error
	var httpErr *errors.HTTPError
	if e.As(err, &httpErr) {
		statusCode = httpErr.StatusCode
		body, _ = httpErr.MarshalJSON()
	} else if e.As(err, &custErr) {
		statusCode = http.StatusInternalServerError
		body, _ = custErr.MarshalJSON()
	} else {
		statusCode = http.StatusInternalServerError
		body, _ = (&errors.Error{
			Code:    "INTERNAL_SERVER_ERROR",
			Message: "Unknown error",
		}).MarshalJSON()
	}
	res := make([]byte, 0, len(body)+100)
	res = append(res, []byte("{\"error\": ")...)
	res = append(res, body...)
	res = append(res, []byte("}")...)
	return statusCode, res
}
