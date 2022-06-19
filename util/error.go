// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package util

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
