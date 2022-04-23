// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package ld

import (
	"fmt"
	"math/big"
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

func ToID(data []byte) (ids.ID, error) {
	switch {
	case len(data) == 0:
		return ids.Empty, nil
	default:
		return ids.ToID(data)
	}
}

func PtrToID(data *[]byte) (ids.ID, error) {
	switch {
	case data == nil || len(*data) == 0:
		return ids.Empty, nil
	default:
		return ids.ToID(*data)
	}
}

func FromID(id ids.ID) []byte {
	if id != ids.Empty {
		return id[:]
	}
	return nil
}

func PtrFromID(id ids.ID) *[]byte {
	if id != ids.Empty {
		b := id[:]
		return &b
	}
	return nil
}

func ToBigInt(data []byte) *big.Int {
	i := &big.Int{}
	return i.SetBytes(data)
}

func PtrToBigInt(data *[]byte) *big.Int {
	if data == nil {
		return nil
	}
	i := &big.Int{}
	return i.SetBytes(*data)
}

func FromBigInt(i *big.Int) []byte {
	if i == nil {
		return nil
	}
	return i.Bytes()
}

func PtrFromBigInt(i *big.Int) *[]byte {
	if i == nil {
		return nil
	}
	b := i.Bytes()
	return &b
}

func PtrToBytes(data *[]byte) []byte {
	if data == nil {
		return nil
	}
	return *data
}

func PtrFromBytes(data []byte) *[]byte {
	if len(data) == 0 {
		return nil
	}
	return &data
}
