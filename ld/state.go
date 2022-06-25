// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package ld

import (
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ldclabs/ldvm/util"
)

type State struct {
	Parent   ids.ID                  `cbor:"p" json:"parent"` // The genesis State's parent ID is ids.Empty.
	Accounts map[util.EthID]ids.ID   `cbor:"a" json:"accounts"`
	Datas    map[util.DataID]ids.ID  `cbor:"d" json:"datas"`
	Models   map[util.ModelID]ids.ID `cbor:"m" json:"models"`

	// external assignment fields
	ID  ids.ID `cbor:"-" json:"id"`
	raw []byte `cbor:"-" json:"-"` // the block's raw bytes
}

func NewState(parent ids.ID) *State {
	return &State{
		Parent:   parent,
		Accounts: make(map[util.EthID]ids.ID),
		Datas:    make(map[util.DataID]ids.ID),
		Models:   make(map[util.ModelID]ids.ID),
	}
}

// SyntacticVerify verifies that a *State is well-formed.
func (s *State) SyntacticVerify() error {
	errp := util.ErrPrefix("State.SyntacticVerify error: ")

	switch {
	case s == nil:
		return errp.Errorf("nil pointer")

	case s.Accounts == nil:
		return errp.Errorf("nil accounts")

	case s.Datas == nil:
		return errp.Errorf("nil datas")

	case s.Models == nil:
		return errp.Errorf("nil models")
	}

	var err error
	if s.raw, err = s.Marshal(); err != nil {
		return errp.ErrorIf(err)
	}

	s.ID = util.IDFromData(s.raw)
	return nil
}

func (s *State) UpdateAccount(id util.EthID, data []byte) {
	s.Accounts[id] = util.IDFromData(data)
}

func (s *State) UpdateData(id util.DataID, data []byte) {
	s.Datas[id] = util.IDFromData(data)
}

func (s *State) UpdateModel(id util.ModelID, data []byte) {
	s.Models[id] = util.IDFromData(data)
}

func (s *State) Clone() *State {
	ns := &State{
		Parent:   s.Parent,
		Accounts: make(map[util.EthID]ids.ID, len(s.Accounts)),
		Datas:    make(map[util.DataID]ids.ID, len(s.Datas)),
		Models:   make(map[util.ModelID]ids.ID, len(s.Models)),
	}
	for k := range s.Accounts {
		ns.Accounts[k] = s.Accounts[k]
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
	return util.ErrPrefix("State.Unmarshal error: ").
		ErrorIf(util.UnmarshalCBOR(data, s))
}

func (s *State) Marshal() ([]byte, error) {
	return util.ErrPrefix("State.Marshal error: ").
		ErrorMap(util.MarshalCBOR(s))
}
