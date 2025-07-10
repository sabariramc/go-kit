package errors

import (
	"encoding/json"
	"net/http"
	"strings"
)

var _ error = &Error{} // interface conformation
var _ error = &HTTPError{}

type Error struct {
	Code        string
	Message     string
	Description any
}

func (e *Error) Error() string {
	sb := &strings.Builder{}
	sb.Grow(len(e.Code) + len(e.Message) + 20)
	if e.Code != "" {
		sb.WriteString("code: ")
		sb.WriteString(e.Code)
		sb.WriteString(" ")
	}
	if e.Message != "" {
		sb.WriteString("message: ")
		sb.WriteString(e.Message)
	}
	return sb.String()
}

func (e *Error) MarshalJSON() ([]byte, error) {
	sb := &strings.Builder{}
	sb.Grow(len(e.Code) + len(e.Message) + 200)
	sb.Write([]byte("{"))
	addComma := false
	valueList := make([][2]string, 0, 3)
	if e.Code != "" {
		valueList = append(valueList, [2]string{"code", e.Code})
	}
	if e.Message != "" {
		valueList = append(valueList, [2]string{"message", e.Message})
	}
	for _, v := range valueList {
		if addComma {
			sb.WriteString(",")
		}
		sb.WriteString("\"" + v[0] + "\":\"" + v[1] + "\"")
		addComma = true
	}
	if e.Description != nil {
		if addComma {
			sb.WriteString(",")
		}
		sb.WriteString("\"description\":")
		blob, err := json.Marshal(e.Description)
		if err != nil {
			sb.WriteString("\"**marshal error**" + err.Error() + "\"")
		} else {
			sb.Write(blob)
		}
	}
	sb.WriteString("}")
	return []byte(sb.String()), nil
}

func (e *Error) String() string {
	sb := &strings.Builder{}
	str := e.Error()
	sb.Grow(len(str) + 200)
	sb.WriteString(str)
	if e.Description != nil {
		sb.WriteString("description: ")
		blob, err := json.Marshal(e.Description)
		if err != nil {
			sb.WriteString("**error marshalling description**")
		} else {
			sb.Write(blob)
		}
	}
	return sb.String()
}

type HTTPError struct {
	StatusCode int    `json:"-"`
	Err        *Error `json:",inline"`
}

func (e *HTTPError) Error() string {
	sb := &strings.Builder{}
	sb.Grow(len(e.Err.Error()) + 30)
	if e.StatusCode != 0 {
		sb.WriteString("status code: ")
		sb.WriteString(http.StatusText(e.StatusCode))
		sb.WriteString(" ")
	}
	sb.WriteString(e.Err.Error())
	return sb.String()
}

func (e *HTTPError) String() string {
	sb := &strings.Builder{}
	str := e.Err.String()
	sb.Grow(len(str) + 30)
	if e.StatusCode != 0 {
		sb.WriteString("status code: ")
		sb.WriteString(http.StatusText(e.StatusCode))
		sb.WriteString(" ")
	}
	sb.WriteString(str)
	return str
}

func (e *HTTPError) MarshalJSON() ([]byte, error) {
	return e.Err.MarshalJSON()
}

var ErrBadRequest = &HTTPError{
	StatusCode: http.StatusBadRequest,
	Err: &Error{
		Code:    "BAD_REQUEST",
		Message: "Invalid input",
	},
}

var ErrNotFound = &HTTPError{
	StatusCode: http.StatusNotFound,
	Err: &Error{
		Code:    "NOT_FOUND",
		Message: "URL Not Found",
	},
}

var ErrMethodNotAllowed = &HTTPError{
	StatusCode: http.StatusMethodNotAllowed,
	Err: &Error{
		Code:    "METHOD_NOT_ALLOWED",
		Message: "Method Not Allowed",
	},
}

var ErrInternalServerError = &HTTPError{
	StatusCode: http.StatusInternalServerError,
	Err: &Error{
		Code:        "INTERNAL_SERVER_ERROR",
		Message:     "Internal Server Error",
		Description: "Retry after some time, if persist contact technical team",
	},
}
