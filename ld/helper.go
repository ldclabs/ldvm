// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package ld

import (
	"fmt"
	"runtime"

	"github.com/ava-labs/avalanchego/ids"
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

func ToShortID(data []byte) (ids.ShortID, error) {
	switch {
	case len(data) == 0:
		return ids.ShortEmpty, nil
	default:
		return ids.ToShortID(data)
	}
}

func PtrToShortID(data *[]byte) (ids.ShortID, error) {
	switch {
	case data == nil || len(*data) == 0:
		return ids.ShortEmpty, nil
	default:
		return ids.ToShortID(*data)
	}
}

func PtrToShortIDs(d *[][]byte) ([]ids.ShortID, error) {
	var data [][]byte
	if d != nil {
		data = *d
	}
	ss := make([]ids.ShortID, len(data))
	for i := range data {
		switch len(data[i]) {
		case 20:
			ss[i] = ids.ShortID{}
			copy(ss[i][:], data[i])
		default:
			return ss, fmt.Errorf("expected 20 bytes but got %d", len(data[i]))
		}
	}
	return ss, nil
}

func FromShortID(id ids.ShortID) []byte {
	if id != ids.ShortEmpty {
		return id[:]
	}
	return nil
}

func PtrFromShortID(id ids.ShortID) *[]byte {
	if id != ids.ShortEmpty {
		b := id[:]
		return &b
	}
	return nil
}

func ToShortIDs(data [][]byte) ([]ids.ShortID, error) {
	list := make([]ids.ShortID, len(data))
	for i := range data {
		id, err := ToShortID(data[i])
		if err != nil {
			return nil, err
		}
		list[i] = id
	}
	return list, nil
}

func FromShortIDs(list []ids.ShortID) [][]byte {
	data := make([][]byte, len(list))
	for i := range list {
		data[i] = list[i][:]
	}
	return data
}

func PtrFromShortIDs(list []ids.ShortID) *[][]byte {
	if len(list) == 0 {
		return nil
	}
	v := make([][]byte, len(list))
	for i := range list {
		v[i] = list[i][:]
	}
	return &v
}
