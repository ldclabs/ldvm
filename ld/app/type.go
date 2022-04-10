// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package app

import (
	"encoding/json"

	"github.com/ipld/go-ipld-prime"
	"github.com/ipld/go-ipld-prime/codec/dagjson"
	"github.com/ipld/go-ipld-prime/datamodel"
	"github.com/ldclabs/ldvm/ld"
)

type MapStringString struct {
	Keys   []string
	Values map[string]string
}

func (m *MapStringString) MarshalJSON() ([]byte, error) {
	if m == nil {
		return ld.Null, nil
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
		m.Keys = append(m.Keys, key) // TODO: sort?
	}
}

func (m *MapStringString) Delete(key string) {
	if m.Values == nil {
		return
	}

	if _, ok := m.Values[key]; ok {
		delete(m.Values, key)
		m.Keys = m.Keys[:0]
		for k := range m.Values {
			m.Keys = append(m.Keys, k)
		}
	}
}

type MapStringAny struct {
	Keys   []string
	Values map[string]datamodel.Node
}

func (m *MapStringAny) MarshalJSON() ([]byte, error) {
	if m == nil {
		return ld.Null, nil
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
		m.Keys = append(m.Keys, key) // TODO: sort?
	}
}

func (m *MapStringAny) Delete(key string) {
	if m.Values == nil {
		return
	}
	if _, ok := m.Values[key]; ok {
		delete(m.Values, key)
		m.Keys = m.Keys[:0]
		for k := range m.Values {
			m.Keys = append(m.Keys, k)
		}
	}
}
