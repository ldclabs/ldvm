// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package erring

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
)

// RespondError represents a response with error that can be sent to the client.
type RespondError struct {
	Err interface{} `json:"error" cbor:"error"`
}

// Error represents a error with more information, can be used as response error.
type Error struct {
	Code    int         `json:"code" cbor:"code"`
	Message string      `json:"message" cbor:"message"`
	Data    interface{} `json:"data,omitempty" cbor:"data,omitempty"`

	// underlying errors that should not reponsed to client.
	errs []error
}

// Error returns the full error information as a JSON string, used as error log.
func (e *Error) Error() string {
	errs := &bytes.Buffer{}
	errs.WriteRune('[')
	for i, err := range e.errs {
		if i > 0 {
			errs.WriteRune(',')
		}
		errs.WriteString(strconv.Quote(err.Error()))
	}

	data, err := json.Marshal(e.Data)
	if err != nil {
		if errs.Len() > 1 {
			errs.WriteRune(',')
		}
		errs.WriteString(strconv.Quote(err.Error()))
	}
	errs.WriteRune(']')
	if data == nil {
		data = []byte("null")
	}

	return fmt.Sprintf(`{"code":%d,"message":%q,"errors":%s,"data":%s}`,
		e.Code, e.Message, errs.String(), data)
}

// HasErrs return true if the Error has underlying errors.
func (e *Error) HasErrs() bool {
	if e == nil {
		return false
	}
	return len(e.errs) > 0
}

// CatchIf catches the error and return true if the error is not nil.
func (e *Error) CatchIf(err error) bool {
	if err == nil {
		return false
	}

	e.errs = append(e.errs, err)
	return true
}

// Unwrap implements the errors.Unwrap interface.
func (e *Error) Unwrap() []error {
	return e.errs
}

// Is implements the errors.As interface.
func (e *Error) Is(target error) bool {
	if e == nil {
		return target == nil
	}

	if error(e) == target {
		return true
	}
	for _, err := range e.errs {
		if errors.Is(err, target) {
			return true
		}
	}
	return false
}

// As implements the errors.As interface.
func (e *Error) As(target any) bool {
	if e == nil {
		return false
	}

	if er, ok := target.(**Error); ok {
		*er = e
		return true
	}

	for _, err := range e.errs {
		if errors.As(err, target) {
			return true
		}
	}
	return false
}
