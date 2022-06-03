// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package ld

import (
	"fmt"
	"runtime/debug"
	"strconv"
)

func Recover(errName string, fn func() error) (err error) {
	defer func() {
		if re := recover(); re != nil {
			buf := debug.Stack()
			err = fmt.Errorf("%s panic: %v, stack: %s", errName, re, strconv.Quote(string(buf)))
		}
	}()

	if err = fn(); err != nil {
		return fmt.Errorf("%s error: %v", errName, err)
	}
	return nil
}

type Marshaler interface {
	Marshal() ([]byte, error)
}

func MustMarshal(v Marshaler) []byte {
	data, err := v.Marshal()
	if err != nil {
		panic(err)
	}
	return data
}

func Uint8Ptr(u uint8) *uint8 {
	return &u
}
