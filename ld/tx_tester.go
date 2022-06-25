// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package ld

import (
	"github.com/ldclabs/ldvm/util"
)

// TxTester TODO
type TxTester struct {
	Data string

	// external assignment fields
	raw []byte `cbor:"-" json:"-"`
}

// SyntacticVerify verifies that a *TxTester is well-formed.
func (t *TxTester) SyntacticVerify() error {
	errp := util.ErrPrefix("TxTester.SyntacticVerify error: ")
	if t == nil {
		return errp.Errorf("invalid TxTester")
	}

	var err error
	if t.raw, err = t.Marshal(); err != nil {
		return errp.ErrorIf(err)
	}
	return nil
}

func (t *TxTester) Bytes() []byte {
	if len(t.raw) == 0 {
		t.raw = MustMarshal(t)
	}
	return t.raw
}

func (t *TxTester) Unmarshal(data []byte) error {
	return util.ErrPrefix("TxTester.Unmarshal error: ").
		ErrorIf(util.UnmarshalCBOR(data, t))
}

func (t *TxTester) Marshal() ([]byte, error) {
	return util.ErrPrefix("TxTester.Marshal error: ").
		ErrorMap(util.MarshalCBOR(t))
}
