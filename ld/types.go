// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package ld

import (
	"encoding/binary"
	"fmt"
	"math/big"
	"strconv"
	"sync"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/utils/crypto"
	"github.com/ava-labs/avalanchego/utils/formatting"
)

var pool10Bytes = sync.Pool{
	New: func() any {
		b := make([]byte, 10)
		return &b
	},
}

type Uint8 []byte

func (u Uint8) Value() uint8 {
	if len(u) == 1 {
		return uint8(u[0])
	}
	return 0
}

func (u Uint8) String() string {
	return strconv.FormatUint(uint64(u.Value()), 10)
}

func (u Uint8) GoString() string {
	return u.String()
}

func FromUint8(x uint8) Uint8 {
	return []byte{x}
}

func PtrFromUint8(x uint8) *Uint8 {
	if x == 0 {
		return nil
	}
	v := FromUint8(x)
	return &v
}

type Uint64 []byte

func (u *Uint64) Value() uint64 {
	if u == nil {
		return 0
	}
	x, _ := binary.Uvarint(*u)
	return x
}

func (u *Uint64) String() string {
	return strconv.FormatUint(u.Value(), 10)
}

func (u *Uint64) GoString() string {
	return u.String()
}

func FromUint64(x uint64) Uint64 {
	buf := pool10Bytes.Get().(*[]byte)
	n := binary.PutUvarint(*buf, x)
	b := make([]byte, n)
	copy(b, *buf)
	pool10Bytes.Put(buf)
	return b
}

func PtrFromUint64(x uint64) *Uint64 {
	if x == 0 {
		return nil
	}
	v := FromUint64(x)
	return &v
}

type Signature [crypto.SECP256K1RSigLen]byte

func (s Signature) MarshalJSON() ([]byte, error) {
	str, err := formatting.EncodeWithChecksum(formatting.CB58, s[:])
	if err != nil {
		return nil, err
	}
	buf := make([]byte, len(str)+2)
	buf[0] = '"'
	copy(buf[1:], []byte(str))
	buf[len(buf)-1] = '"'
	return buf, nil
}

func SignaturesFromStrings(ss []string) ([]Signature, error) {
	sigs := make([]Signature, len(ss))
	for i, s := range ss {
		d, err := formatting.Decode(formatting.CB58, s)
		if err != nil {
			return nil, fmt.Errorf("invalid signature %s, decode failed: %v", strconv.Quote(s), err)
		}
		if len(d) != crypto.SECP256K1RSigLen {
			return nil, fmt.Errorf("invalid signature %s", strconv.Quote(s))
		}
		sigs[i] = Signature{}
		copy(sigs[i][:], d)
	}
	return sigs, nil
}

func PtrToSignatures(d *[][]byte) ([]Signature, error) {
	var data [][]byte
	if d != nil {
		data = *d
	}
	ss := make([]Signature, len(data))
	for i := range data {
		switch len(data[i]) {
		case crypto.SECP256K1RSigLen:
			ss[i] = Signature{}
			copy(ss[i][:], data[i])
		default:
			return ss, fmt.Errorf("expected 65 bytes but got %d", len(data[i]))
		}
	}
	return ss, nil
}

func PtrFromSignatures(ss []Signature) *[][]byte {
	if len(ss) == 0 {
		return nil
	}
	v := make([][]byte, len(ss))
	for i := range ss {
		v[i] = ss[i][:]
	}
	return &v
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
