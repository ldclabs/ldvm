// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package transaction

import (
	"encoding/json"
	"fmt"

	"github.com/ldclabs/ldvm/ld"
)

type TxOpenLending struct {
	TxBase
	input *ld.LendingConfig
}

func (tx *TxOpenLending) MarshalJSON() ([]byte, error) {
	if tx == nil || tx.ld == nil {
		return []byte("null"), nil
	}
	v := tx.ld.Copy()
	if tx.input == nil {
		return nil, fmt.Errorf("TxOpenLending.MarshalJSON failed: invalid tx.input")
	}
	d, err := json.Marshal(tx.input)
	if err != nil {
		return nil, err
	}
	v.Data = d
	return json.Marshal(v)
}

func (tx *TxOpenLending) SyntacticVerify() error {
	var err error
	errPrefix := "TxOpenLending.SyntacticVerify failed:"
	if err = tx.TxBase.SyntacticVerify(); err != nil {
		return fmt.Errorf("%s %v", errPrefix, err)
	}

	switch {
	case tx.ld.To != nil:
		return fmt.Errorf("%s invalid to, should be nil", errPrefix)

	case tx.ld.Amount != nil:
		return fmt.Errorf("%s invalid amount, should be nil", errPrefix)

	case tx.ld.Token != nil:
		return fmt.Errorf("%s invalid token, should be nil", errPrefix)

	case len(tx.ld.Data) == 0:
		return fmt.Errorf("%s invalid data", errPrefix)
	}

	tx.input = &ld.LendingConfig{}
	if err = tx.input.Unmarshal(tx.ld.Data); err != nil {
		return fmt.Errorf("%s %v", errPrefix, err)
	}
	if err = tx.input.SyntacticVerify(); err != nil {
		return fmt.Errorf("%s %v", errPrefix, err)
	}
	return nil
}

func (tx *TxOpenLending) Verify(bctx BlockContext, bs BlockState) error {
	var err error
	if err = tx.TxBase.Verify(bctx, bs); err != nil {
		return fmt.Errorf("TxOpenLending.Verify failed: %v", err)
	}
	if err = tx.from.CheckOpenLending(tx.input); err != nil {
		return fmt.Errorf("TxOpenLending.Verify failed: %v", err)
	}
	return nil
}

func (tx *TxOpenLending) Accept(bctx BlockContext, bs BlockState) error {
	var err error
	if err = tx.from.OpenLending(tx.input); err != nil {
		return err
	}
	return tx.TxBase.Accept(bctx, bs)
}
