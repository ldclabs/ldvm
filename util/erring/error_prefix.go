// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package erring

import (
	"errors"
	"fmt"
)

// ErrPrefix is a string that can be used to prefix an error message.
type ErrPrefix string

// ErrorIf returns a error with prefix if the error is not nil, otherwise it returns nil.
func (p ErrPrefix) ErrorIf(err error) error {
	if err != nil {
		return errors.New(string(p) + err.Error())
	}
	return nil
}

// Errorf returns a error with prefix and given format and arguments.
func (p ErrPrefix) Errorf(format string, a ...any) error {
	err := fmt.Sprintf(format, a...)
	return errors.New(string(p) + err)
}

// ErrorMap is used to unwrap result from a function that returns a byte slice and an error.
// If the error is not nil, it returns the error with prefix, otherwise it returns the byte slice.
func (p ErrPrefix) ErrorMap(data []byte, err error) ([]byte, error) {
	if err != nil {
		return nil, errors.New(string(p) + err.Error())
	}
	return data, nil
}

// Sprintf returns a string with prefix and given format and arguments.
func (p ErrPrefix) Sprintf(format string, a ...any) string {
	if len(a) == 0 {
		return string(p) + format
	}

	return string(p) + fmt.Sprintf(format, a...)
}
