// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package transaction

import (
	"encoding/json"
	"fmt"

	"github.com/ldclabs/ldvm/constants"
	"github.com/ldclabs/ldvm/ld"
	"github.com/ldclabs/ldvm/util"
)

type TxWithdrawStake struct {
	TxBase
	input *ld.TxTransfer
}

func (tx *TxWithdrawStake) MarshalJSON() ([]byte, error) {
	if tx == nil || tx.ld == nil {
		return []byte("null"), nil
	}
	v := tx.ld.Copy()
	if tx.input == nil {
		return nil, fmt.Errorf("TxWithdrawStake.MarshalJSON error: invalid tx.input")
	}
	d, err := json.Marshal(tx.input)
	if err != nil {
		return nil, err
	}
	v.Data = d
	return json.Marshal(v)
}

func (tx *TxWithdrawStake) SyntacticVerify() error {
	var err error
	errp := util.ErrPrefix("TxWithdrawStake.SyntacticVerify error: ")

	if err = tx.TxBase.SyntacticVerify(); err != nil {
		return errp.ErrorIf(err)
	}

	switch {
	case tx.ld.To == nil:
		return errp.Errorf("nil to as stake account")

	case tx.ld.Amount != nil:
		return errp.Errorf("invalid amount, should be nil")

	case len(tx.ld.Data) == 0:
		return errp.Errorf("invalid data")
	}

	if stake := util.StakeSymbol(*tx.ld.To); !stake.Valid() {
		return errp.Errorf("invalid stake account %s", stake.GoString())
	}

	tx.input = &ld.TxTransfer{}
	if err = tx.input.Unmarshal(tx.ld.Data); err != nil {
		return errp.ErrorIf(err)
	}
	if err = tx.input.SyntacticVerify(); err != nil {
		return errp.ErrorIf(err)
	}

	switch {
	case tx.input.Token == nil && tx.token != constants.NativeToken:
		return errp.Errorf("invalid token, expected %s, got %s",
			constants.NativeToken.GoString(), tx.token.GoString())

	case tx.input.Token != nil && tx.token != *tx.input.Token:
		return errp.Errorf("invalid token, expected %s, got %s",
			tx.input.Token.GoString(), tx.token.GoString())

	case tx.input.Amount == nil || tx.input.Amount.Sign() <= 0:
		return errp.Errorf("invalid amount, expected >= 1")
	}
	return nil
}

func (tx *TxWithdrawStake) Verify(bctx BlockContext, bs BlockState) error {
	var err error
	errp := util.ErrPrefix("TxWithdrawStake.Verify error: ")

	if err = tx.TxBase.Verify(bctx, bs); err != nil {
		return errp.ErrorIf(err)
	}
	if err = tx.to.CheckWithdrawStake(
		tx.token, tx.ld.From, tx.signers, tx.input.Amount); err != nil {
		return errp.ErrorIf(err)
	}
	return nil
}

func (tx *TxWithdrawStake) Accept(bctx BlockContext, bs BlockState) error {
	errp := util.ErrPrefix("TxWithdrawStake.Accept error: ")

	// must WithdrawStake and then Accept
	withdraw, err := tx.to.WithdrawStake(tx.token, tx.ld.From, tx.signers, tx.input.Amount)
	if err != nil {
		return errp.ErrorIf(err)
	}
	if err = tx.to.Sub(tx.token, withdraw); err != nil {
		return errp.ErrorIf(err)
	}
	if err = tx.from.Add(tx.token, withdraw); err != nil {
		return errp.ErrorIf(err)
	}
	return errp.ErrorIf(tx.TxBase.Accept(bctx, bs))
}
