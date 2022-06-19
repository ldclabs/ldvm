// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package transaction

import (
	"encoding/json"
	"fmt"

	"github.com/ldclabs/ldvm/ld"
	"github.com/ldclabs/ldvm/util"
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
		return nil, fmt.Errorf("TxOpenLending.MarshalJSON error: invalid tx.input")
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
	errp := util.ErrPrefix("TxOpenLending.SyntacticVerify error: ")

	if err = tx.TxBase.SyntacticVerify(); err != nil {
		return errp.ErrorIf(err)
	}

	switch {
	case tx.ld.To != nil:
		return errp.Errorf("invalid to, should be nil")

	case tx.ld.Amount != nil:
		return errp.Errorf("invalid amount, should be nil")

	case tx.ld.Token != nil:
		return errp.Errorf("invalid token, should be nil")

	case len(tx.ld.Data) == 0:
		return errp.Errorf("invalid data")
	}

	tx.input = &ld.LendingConfig{}
	if err = tx.input.Unmarshal(tx.ld.Data); err != nil {
		return errp.ErrorIf(err)
	}
	if err = tx.input.SyntacticVerify(); err != nil {
		return errp.ErrorIf(err)
	}
	return nil
}

func (tx *TxOpenLending) Verify(bctx BlockContext, bs BlockState) error {
	var err error
	errp := util.ErrPrefix("TxOpenLending.Verify error: ")

	if err = tx.TxBase.Verify(bctx, bs); err != nil {
		return errp.ErrorIf(err)
	}
	if err = tx.from.CheckOpenLending(tx.input); err != nil {
		return errp.ErrorIf(err)
	}
	return nil
}

func (tx *TxOpenLending) Accept(bctx BlockContext, bs BlockState) error {
	var err error
	errp := util.ErrPrefix("TxOpenLending.Accept error: ")

	if err = tx.from.OpenLending(tx.input); err != nil {
		return errp.ErrorIf(err)
	}
	return errp.ErrorIf(tx.TxBase.Accept(bctx, bs))
}
