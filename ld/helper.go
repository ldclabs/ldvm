// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package ld

import (
	"runtime/debug"

	"github.com/ldclabs/ldvm/util"
)

func Recover(errp util.ErrPrefix, fn func() error) (err error) {
	defer func() {
		if re := recover(); re != nil {
			buf := debug.Stack()
			err = errp.Errorf("panic: %v, stack: %q", re, string(buf))
		}
	}()

	if err = fn(); err != nil {
		return errp.ErrorIf(err)
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

func Uint16Ptr(u uint16) *uint16 {
	return &u
}

type Copier interface {
	Marshal() ([]byte, error)
	Unmarshal([]byte) error
	SyntacticVerify() error
}

func Copy(dst, src Copier) error {
	data, err := src.Marshal()
	if err != nil {
		return err
	}
	if err = dst.Unmarshal(data); err != nil {
		return err
	}
	return dst.SyntacticVerify()
}
