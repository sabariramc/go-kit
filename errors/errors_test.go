package errors_test

import (
	"fmt"
	"testing"

	"github.com/sabariramc/go-kit/errors"
	"gotest.tools/v3/assert"
)

func TestCustomError(t *testing.T) {
	var err error
	err = &errors.Error{
		Code:    "KEY_NOT_FOUND",
		Message: "Key not found",
	}
	assert.Error(t, err, "code: KEY_NOT_FOUND message: Key not found")
	err = &errors.Error{
		Code: "KEY_NOT_FOUND",
	}
	assert.Error(t, err, "code: KEY_NOT_FOUND ")
}

var err error
var errorString string

func BenchmarkErrors(b *testing.B) {
	b.Run("CustomError", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			err = &errors.Error{
				Code:    "KEY_NOT_FOUND",
				Message: "Key not found",
			}
			errorString = err.Error()
		}
	})
	b.Run("Fmt Errors", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			err = fmt.Errorf("code: %s message: %s", "KEY_NOT_FOUND", "Key not found")
			errorString = err.Error()
		}
	})
}
