// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package ld

import (
	"fmt"
	"runtime"
)

func Recover(errfmt string, fn func() error) (err error) {
	defer func() {
		if re := recover(); re != nil {
			buf := make([]byte, 2048)
			buf = buf[:runtime.Stack(buf, false)]
			err = fmt.Errorf("%s panic: %v, stack: %s", errfmt, re, string(buf))
		}
	}()

	if err = fn(); err != nil {
		return fmt.Errorf("%s error: %v", errfmt, err)
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
