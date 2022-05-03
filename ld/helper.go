// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package ld

import (
	"bytes"
	"fmt"
	"io"
	"runtime"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ldclabs/ldvm/util"
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

func ToSignature(data []byte) (util.Signature, error) {
	ss := util.Signature{}
	switch len(data) {
	case crypto.SignatureLength:
		copy(ss[:], data)
		return ss, nil
	default:
		return ss, fmt.Errorf("expected 65 bytes but got %d", len(data))
	}
}

func PtrToSignature(d *[]byte) (*util.Signature, error) {
	var data []byte
	if d == nil {
		return nil, nil
	}

	data = *d
	switch len(data) {
	case crypto.SignatureLength:
		ss := util.Signature{}
		copy(ss[:], data)
		return &ss, nil
	default:
		return nil, fmt.Errorf("expected 65 bytes but got %d", len(data))
	}
}

func PtrFromSignature(ss *util.Signature) *[]byte {
	if ss == nil || *ss == util.SignatureEmpty {
		return nil
	}
	v := (*ss)[:]
	return &v
}

func PtrToSignatures(d *[][]byte) ([]util.Signature, error) {
	var data [][]byte
	if d != nil {
		data = *d
	}
	ss := make([]util.Signature, len(data))
	for i := range data {
		switch len(data[i]) {
		case crypto.SignatureLength:
			ss[i] = util.Signature{}
			copy(ss[i][:], data[i])
		default:
			return ss, fmt.Errorf("expected 65 bytes but got %d", len(data[i]))
		}
	}
	return ss, nil
}

func PtrFromSignatures(ss []util.Signature) *[][]byte {
	if len(ss) == 0 {
		return nil
	}
	v := make([][]byte, len(ss))
	for i := range ss {
		v[i] = ss[i][:]
	}
	return &v
}

func WriteUint64s(w io.Writer, u uint64, uu ...uint64) error {
	b := FromUint64(u)
	bb := make([][]byte, len(uu))
	for i, u := range uu {
		bb[i] = FromUint64(u)
	}
	return WriteBytesList(w, b, bb...)
}

func ReadUint64s(data []byte) ([]uint64, error) {
	bb, err := ReadBytesList(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	list := make([]uint64, len(bb))
	for i, b := range bb {
		u := Uint64(b)
		if !u.Valid() {
			return nil, fmt.Errorf("ReadUint64s error: invalid uint64 bytes")
		}
		list[i] = u.Value()
	}

	return list, nil
}
