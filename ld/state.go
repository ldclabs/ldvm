// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package ld

import (
	"github.com/ldclabs/ldvm/ids"
	"github.com/ldclabs/ldvm/util/encoding"
	"github.com/ldclabs/ldvm/util/erring"
)

type State struct {
	Parent   ids.ID32            `cbor:"p" json:"parent"` // The genesis State's parent ID is ids.Empty.
	Accounts map[string]ids.ID32 `cbor:"a" json:"accounts"`
	Ledgers  map[string]ids.ID32 `cbor:"l" json:"ledgers"`
	Datas    map[string]ids.ID32 `cbor:"d" json:"datas"`
	Models   map[string]ids.ID32 `cbor:"m" json:"models"`

	// external assignment fields
	ID  ids.ID32 `cbor:"-" json:"id"`
	raw []byte   `cbor:"-" json:"-"` // the block's raw bytes
}

func NewState(parent ids.ID32) *State {
	return &State{
		Parent:   parent,
		Accounts: make(map[string]ids.ID32),
		Ledgers:  make(map[string]ids.ID32),
		Datas:    make(map[string]ids.ID32),
		Models:   make(map[string]ids.ID32),
	}
}

// SyntacticVerify verifies that a *State is well-formed.
func (s *State) SyntacticVerify() error {
	errp := erring.ErrPrefix("ld.State.SyntacticVerify: ")

	switch {
	case s == nil:
		return errp.Errorf("nil pointer")

	case s.Accounts == nil:
		return errp.Errorf("nil accounts")

	case s.Ledgers == nil:
		return errp.Errorf("nil ledgers")

	case s.Datas == nil:
		return errp.Errorf("nil datas")

	case s.Models == nil:
		return errp.Errorf("nil models")
	}

	var err error
	if s.raw, err = s.Marshal(); err != nil {
		return errp.ErrorIf(err)
	}

	s.ID = ids.ID32FromData(s.raw)
	return nil
}

func (s *State) UpdateAccount(id ids.Address, data []byte) {
	s.Accounts[string(id[:])] = ids.ID32FromData(data)
}

func (s *State) UpdateLedger(id ids.Address, data []byte) {
	s.Ledgers[string(id[:])] = ids.ID32FromData(data)
}

func (s *State) UpdateData(id ids.DataID, data []byte) {
	s.Datas[string(id[:])] = ids.ID32FromData(data)
}

func (s *State) UpdateModel(id ids.ModelID, data []byte) {
	s.Models[string(id[:])] = ids.ID32FromData(data)
}

func (s *State) Clone() *State {
	ns := &State{
		Parent:   s.Parent,
		Accounts: make(map[string]ids.ID32, len(s.Accounts)),
		Ledgers:  make(map[string]ids.ID32, len(s.Ledgers)),
		Datas:    make(map[string]ids.ID32, len(s.Datas)),
		Models:   make(map[string]ids.ID32, len(s.Models)),
	}

	for k := range s.Accounts {
		ns.Accounts[k] = s.Accounts[k]
	}

	for k := range s.Ledgers {
		ns.Ledgers[k] = s.Ledgers[k]
	}

	for k := range s.Datas {
		ns.Datas[k] = s.Datas[k]
	}

	for k := range s.Models {
		ns.Models[k] = s.Models[k]
	}
	return ns
}

func (s *State) Bytes() []byte {
	if len(s.raw) == 0 {
		s.raw = MustMarshal(s)
	}
	return s.raw
}

func (s *State) Unmarshal(data []byte) error {
	return erring.ErrPrefix("ld.State.Unmarshal: ").
		ErrorIf(encoding.UnmarshalCBOR(data, s))
}

func (s *State) Marshal() ([]byte, error) {
	return erring.ErrPrefix("ld.State.Marshal: ").
		ErrorMap(encoding.MarshalCBOR(s))
}
