// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package ld

import (
	"fmt"
)

// TxTester TODO
type TxTester struct {
	Data string
}

// SyntacticVerify verifies that a *TxTester is well-formed.
func (t *TxTester) SyntacticVerify() error {
	if t == nil {
		return fmt.Errorf("invalid TxTester")
	}

	if _, err := t.Marshal(); err != nil {
		return fmt.Errorf("TxTester marshal error: %v", err)
	}
	return nil
}

func (t *TxTester) Unmarshal(data []byte) error {
	return DecMode.Unmarshal(data, t)
}

func (t *TxTester) Marshal() ([]byte, error) {
	data, err := EncMode.Marshal(t)
	if err != nil {
		return nil, err
	}
	return data, nil
}
