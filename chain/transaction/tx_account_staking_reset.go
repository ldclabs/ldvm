// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package transaction

import (
	"encoding/json"
	"fmt"

	"github.com/ldclabs/ldvm/ld"
	"github.com/ldclabs/ldvm/util"
)

type TxResetStakeAccount struct {
	TxBase
	input *ld.StakeConfig
}

func (tx *TxResetStakeAccount) MarshalJSON() ([]byte, error) {
	if tx == nil || tx.ld == nil {
		return []byte("null"), nil
	}
	v := tx.ld.Copy()
	if tx.input == nil {
		return nil, fmt.Errorf("TxResetStakeAccount.MarshalJSON error: invalid tx.input")
	}
	d, err := json.Marshal(tx.input)
	if err != nil {
		return nil, err
	}
	v.Data = d
	return json.Marshal(v)
}

func (tx *TxResetStakeAccount) SyntacticVerify() error {
	var err error
	errp := util.ErrPrefix("TxResetStakeAccount.SyntacticVerify error: ")

	if err = tx.TxBase.SyntacticVerify(); err != nil {
		return errp.ErrorIf(err)
	}

	switch {
	case tx.ld.To != nil:
		return errp.Errorf("invalid to, should be nil")

	case tx.ld.Token != nil:
		return errp.Errorf("invalid token, should be nil")

	case tx.ld.Amount != nil:
		return errp.Errorf("invalid amount, should be nil")

	case len(tx.ld.Data) == 0:
		return errp.Errorf("invalid data")
	}

	if stake := util.StakeSymbol(tx.ld.From); !stake.Valid() {
		return errp.Errorf("invalid stake account %s", stake.GoString())
	}

	tx.input = &ld.StakeConfig{}
	if err = tx.input.Unmarshal(tx.ld.Data); err != nil {
		return errp.ErrorIf(err)
	}
	if err = tx.input.SyntacticVerify(); err != nil {
		return errp.ErrorIf(err)
	}

	if tx.input.LockTime < tx.ld.Timestamp {
		return errp.Errorf("invalid lockTime, expected >= %d", tx.ld.Timestamp)
	}
	return nil
}

func (tx *TxResetStakeAccount) Verify(bctx BlockContext, bs BlockState) error {
	var err error
	errp := util.ErrPrefix("TxResetStakeAccount.Verify error: ")

	if err = tx.TxBase.Verify(bctx, bs); err != nil {
		return errp.ErrorIf(err)
	}
	if !tx.from.SatisfySigningPlus(tx.signers) {
		return errp.Errorf("invalid signatures for stake keepers")
	}
	if err = tx.from.CheckResetStake(tx.input); err != nil {
		return errp.ErrorIf(err)
	}
	return nil
}

func (tx *TxResetStakeAccount) Accept(bctx BlockContext, bs BlockState) error {
	errp := util.ErrPrefix("TxResetStakeAccount.Accept error: ")

	if err := tx.TxBase.Accept(bctx, bs); err != nil {
		return errp.ErrorIf(err)
	}
	// do it after TxBase.Accept
	return errp.ErrorIf(tx.from.ResetStake(tx.input))
}
