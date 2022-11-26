// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package erring

import (
	"errors"
	"fmt"
)

type ErrPrefix string

func (p ErrPrefix) ErrorIf(err error) error {
	if err != nil {
		return errors.New(string(p) + err.Error())
	}
	return nil
}

func (p ErrPrefix) Errorf(format string, a ...any) error {
	err := fmt.Sprintf(format, a...)
	return errors.New(string(p) + err)
}

func (p ErrPrefix) ErrorMap(data []byte, err error) ([]byte, error) {
	if err != nil {
		return nil, errors.New(string(p) + err.Error())
	}
	return data, nil
}

func (p ErrPrefix) Sprintf(format string, a ...any) string {
	if len(a) == 0 {
		return string(p) + format
	}

	return string(p) + fmt.Sprintf(format, a...)
}
