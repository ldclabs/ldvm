// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package ld

import (
	"encoding/json"
	"math/big"
	"sort"

	ipld "github.com/ipld/go-ipld-prime"
	"github.com/ipld/go-ipld-prime/codec/dagjson"
	"github.com/ipld/go-ipld-prime/datamodel"

	"github.com/ldclabs/ldvm/util"
)

type LDObject interface {
	SyntacticVerify() error
	Unmarshal(data []byte) error
	Marshal() ([]byte, error)
	MarshalJSON() ([]byte, error)
}

type BigUint []byte

func FromUint(u *big.Int) BigUint {
	if u == nil {
		return []byte{}
	}
	return u.Bytes()
}

func PtrFromUint(u *big.Int) *BigUint {
	if u == nil {
		return nil
	}
	b := BigUint(u.Bytes())
	return &b
}

func (b *BigUint) Value() *big.Int {
	u := new(big.Int)
	if b != nil {
		u.SetBytes(*b)
	}
	return u
}

func (b *BigUint) PtrValue() *big.Int {
	if b == nil {
		return nil
	}
	return b.Value()
}

type Uint8 []byte

func FromUint8(u uint8) Uint8 {
	if u == 0 {
		return []byte{}
	}
	return []byte{u}
}

func PtrFromUint8(u uint8) *Uint8 {
	if u == 0 {
		return nil
	}
	b := Uint8([]byte{u})
	return &b
}

func (b *Uint8) Valid() bool {
	return b == nil || len(*b) <= 1
}

func (b *Uint8) Value() uint8 {
	if b == nil || len(*b) == 0 {
		return 0
	}
	return (*b)[0]
}

type Uint64 []byte

func FromUint64(u uint64) Uint64 {
	return new(big.Int).SetUint64(u).Bytes()
}

func PtrFromUint64(u uint64) *Uint64 {
	if u == 0 {
		return nil
	}
	b := Uint64(new(big.Int).SetUint64(u).Bytes())
	return &b
}

func (b *Uint64) Valid() bool {
	return b == nil || len(*b) <= 8
}

func (b *Uint64) Value() uint64 {
	if b == nil {
		return 0
	}
	return new(big.Int).SetBytes(*b).Uint64()
}

type MapStringString struct {
	Keys   []string
	Values map[string]string
}

func NewMapStringString(size int) *MapStringString {
	return &MapStringString{
		Keys:   make([]string, 0, size),
		Values: make(map[string]string, size),
	}
}

func (m *MapStringString) MarshalJSON() ([]byte, error) {
	if m == nil {
		return util.Null, nil
	}
	return json.Marshal(m.Values)
}

func (m *MapStringString) Equal(o *MapStringString) bool {
	if o == nil {
		return m == nil
	}
	if len(m.Keys) != len(o.Keys) {
		return false
	}
	for _, k := range m.Keys {
		if m.Values[k] != o.Values[k] {
			return false
		}
	}
	return true
}

func (m *MapStringString) Has(key string) bool {
	if m.Values == nil {
		return false
	}
	_, ok := m.Values[key]
	return ok
}

func (m *MapStringString) Get(key string) string {
	if m.Values == nil {
		return ""
	}
	return m.Values[key]
}

func (m *MapStringString) Set(key, value string) {
	ok := false
	if m.Values == nil {
		m.Values = make(map[string]string)
	} else {
		_, ok = m.Values[key]
	}
	m.Values[key] = value
	if !ok {
		m.Keys = append(m.Keys, key)
		sort.Stable(sort.StringSlice(m.Keys))
	}
}

func (m *MapStringString) Delete(key string) {
	if m.Values == nil {
		return
	}

	if _, ok := m.Values[key]; ok {
		delete(m.Values, key)
		for i, k := range m.Keys {
			if k == key {
				n := copy(m.Keys[i:], m.Keys[i+1:])
				m.Keys = m.Keys[:i+n]
				break
			}
		}
	}
}

type MapStringAny struct {
	Keys   []string
	Values map[string]datamodel.Node
}

func NewMapStringAny(size int) *MapStringAny {
	return &MapStringAny{
		Keys:   make([]string, 0, size),
		Values: make(map[string]datamodel.Node, size),
	}
}

func (m *MapStringAny) MarshalJSON() ([]byte, error) {
	if m == nil {
		return util.Null, nil
	}

	v := make(map[string]json.RawMessage, len(m.Keys))
	for _, k := range m.Keys {
		raw, err := ipld.Encode(m.Values[k], dagjson.Encode)
		if err != nil {
			return nil, err
		}
		v[k] = raw
	}
	return json.Marshal(v)
}

func (m *MapStringAny) Equal(o *MapStringAny) bool {
	if o == nil {
		return m == nil
	}
	if len(m.Keys) != len(o.Keys) {
		return false
	}
	for _, k := range m.Keys {
		if !datamodel.DeepEqual(m.Values[k], o.Values[k]) {
			return false
		}
	}
	return true
}

func (m *MapStringAny) Has(key string) bool {
	if m.Values == nil {
		return false
	}
	_, ok := m.Values[key]
	return ok
}

func (m *MapStringAny) Get(key string) datamodel.Node {
	if m.Values == nil {
		return nil
	}
	return m.Values[key]
}

func (m *MapStringAny) Set(key string, value datamodel.Node) {
	ok := false
	if m.Values == nil {
		m.Values = make(map[string]datamodel.Node)
	} else {
		_, ok = m.Values[key]
	}
	m.Values[key] = value
	if !ok {
		m.Keys = append(m.Keys, key)
		sort.Stable(sort.StringSlice(m.Keys))
	}
}

func (m *MapStringAny) Delete(key string) {
	if m.Values == nil {
		return
	}
	if _, ok := m.Values[key]; ok {
		delete(m.Values, key)
		for i, k := range m.Keys {
			if k == key {
				n := copy(m.Keys[i:], m.Keys[i+1:])
				m.Keys = m.Keys[:i+n]
				break
			}
		}
	}
}
