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

type TxTakeStake struct {
	TxBase
	exSigners util.EthIDs
	input     *ld.TxTransfer
	lockTime  uint64
}

func (tx *TxTakeStake) MarshalJSON() ([]byte, error) {
	if tx == nil || tx.ld == nil {
		return []byte("null"), nil
	}
	v := tx.ld.Copy()
	if tx.input == nil {
		return nil, fmt.Errorf("TxTakeStake.MarshalJSON error: invalid tx.input")
	}
	d, err := json.Marshal(tx.input)
	if err != nil {
		return nil, err
	}
	v.Data = d
	return json.Marshal(v)
}

func (tx *TxTakeStake) SyntacticVerify() error {
	var err error
	errp := util.ErrPrefix("TxTakeStake.SyntacticVerify error: ")

	if err = tx.TxBase.SyntacticVerify(); err != nil {
		return errp.ErrorIf(err)
	}

	switch {
	case tx.ld.To == nil:
		return errp.Errorf("nil to as stake account")

	case tx.ld.Amount == nil:
		return errp.Errorf("nil amount")

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
	case tx.input.From == nil:
		return errp.Errorf("nil from")

	case *tx.input.From != tx.ld.From:
		return errp.Errorf("invalid from, expected %s, got %s",
			tx.input.From, tx.ld.From)

	case tx.input.Token == nil && tx.token != constants.NativeToken:
		return errp.Errorf("invalid token, expected %s, got %s",
			constants.NativeToken.GoString(), tx.token.GoString())

	case tx.input.Token != nil && tx.token != *tx.input.Token:
		return errp.Errorf("invalid token, expected %s, got %s",
			tx.input.Token.GoString(), tx.token.GoString())

	case tx.input.Amount == nil:
		return errp.Errorf("nil amount")

	case tx.input.Amount.Cmp(tx.ld.Amount) != 0:
		return errp.Errorf("invalid amount, expected %v, got %v",
			tx.input.Amount, tx.ld.Amount)

	case tx.input.Expire < tx.ld.Timestamp:
		return errp.Errorf("data expired")
	}

	if len(tx.input.Data) > 0 {
		u := uint64(0)
		if err = util.UnmarshalCBOR(tx.input.Data, &u); err != nil {
			return errp.Errorf("invalid lockTime, %v", err)
		}
		tx.lockTime = u
	}
	tx.exSigners, err = tx.ld.ExSigners()
	if err != nil {
		return errp.Errorf("invalid exSignatures, %v", err)
	}
	return nil
}

func (tx *TxTakeStake) Verify(bctx BlockContext, bs BlockState) error {
	var err error
	errp := util.ErrPrefix("TxTakeStake.Verify error: ")

	if err = tx.TxBase.Verify(bctx, bs); err != nil {
		return errp.ErrorIf(err)
	}
	if !tx.to.SatisfySigning(tx.exSigners) {
		return errp.Errorf("invalid exSignatures for stake keepers")
	}
	if err = tx.to.CheckTakeStake(tx.token, tx.ld.From, tx.ld.Amount, tx.lockTime); err != nil {
		return errp.ErrorIf(err)
	}
	return nil
}

func (tx *TxTakeStake) Accept(bctx BlockContext, bs BlockState) error {
	var err error
	errp := util.ErrPrefix("TxTakeStake.Accept error: ")
	// must TakeStake and then Accept
	if err = tx.to.TakeStake(tx.token, tx.ld.From, tx.ld.Amount, tx.lockTime); err != nil {
		return errp.ErrorIf(err)
	}
	return errp.ErrorIf(tx.TxBase.Accept(bctx, bs))
}
