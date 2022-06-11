// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package ld

import (
	"fmt"
)

// TxTester TODO
type TxTester struct {
	Data string

	// external assignment fields
	raw []byte `cbor:"-" json:"-"`
}

// SyntacticVerify verifies that a *TxTester is well-formed.
func (t *TxTester) SyntacticVerify() error {
	errPrefix := "TxTester.SyntacticVerify failed:"
	if t == nil {
		return fmt.Errorf("%s invalid TxTester", errPrefix)
	}

	var err error
	if t.raw, err = t.Marshal(); err != nil {
		return fmt.Errorf("%s %v", errPrefix, err)
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
	return DecMode.Unmarshal(data, t)
}

func (t *TxTester) Marshal() ([]byte, error) {
	return EncMode.Marshal(t)
}
