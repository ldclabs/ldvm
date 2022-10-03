// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package transactions

import (
	"encoding/json"

	"github.com/ldclabs/ldvm/constants"
	"github.com/ldclabs/ldvm/ld"
	"github.com/ldclabs/ldvm/util"
)

type TxBorrow struct {
	TxBase
	exSigners util.EthIDs
	input     *ld.TxTransfer
	dueTime   uint64
}

func (tx *TxBorrow) MarshalJSON() ([]byte, error) {
	if tx == nil || tx.ld == nil {
		return []byte("null"), nil
	}

	v := tx.ld.Copy()
	errp := util.ErrPrefix("TxBorrow.MarshalJSON error: ")
	if tx.input == nil {
		return nil, errp.Errorf("nil tx.input")
	}
	d, err := json.Marshal(tx.input)
	if err != nil {
		return nil, errp.ErrorIf(err)
	}
	v.Tx.Data = d
	return errp.ErrorMap(json.Marshal(v))
}

func (tx *TxBorrow) SyntacticVerify() error {
	var err error
	errp := util.ErrPrefix("TxBorrow.SyntacticVerify error: ")

	if err = tx.TxBase.SyntacticVerify(); err != nil {
		return errp.ErrorIf(err)
	}

	switch {
	case tx.ld.Tx.To == nil:
		return errp.Errorf("nil to as lender")

	case tx.ld.Tx.Amount != nil:
		return errp.Errorf("invalid amount, should be nil")

	case len(tx.ld.Tx.Data) == 0:
		return errp.Errorf("invalid data")
	}

	tx.input = &ld.TxTransfer{}
	if err = tx.input.Unmarshal(tx.ld.Tx.Data); err != nil {
		return errp.ErrorIf(err)
	}
	if err = tx.input.SyntacticVerify(); err != nil {
		return errp.ErrorIf(err)
	}

	switch {
	case tx.input.From == nil:
		return errp.Errorf("nil from as lender")

	case *tx.input.From != *tx.ld.Tx.To:
		return errp.Errorf("invalid to as borrower, expected %s, got %s", tx.input.From, tx.ld.Tx.To)

	case tx.input.To == nil:
		return errp.Errorf("nil to as borrower")

	case *tx.input.To != tx.ld.Tx.From:
		return errp.Errorf("invalid from as lender, expected %s, got %s", tx.input.To, tx.ld.Tx.From)

	case tx.input.Token == nil && tx.token != constants.NativeToken:
		return errp.Errorf("invalid token, expected %s, got %s",
			constants.NativeToken.GoString(), tx.token.GoString())

	case tx.input.Token != nil && tx.token != *tx.input.Token:
		return errp.Errorf("invalid token, expected %s, got %s",
			tx.input.Token.GoString(), tx.token.GoString())

	case tx.input.Amount == nil || tx.input.Amount.Sign() <= 0:
		return errp.Errorf("invalid amount, expected >= 1")

	case tx.input.Expire < tx.ld.Timestamp:
		return errp.Errorf("data expired")
	}

	if len(tx.input.Data) > 0 {
		u := uint64(0)
		if err = util.UnmarshalCBOR(tx.input.Data, &u); err != nil {
			return errp.Errorf("invalid dueTime, %v", err)
		}
		if u <= tx.ld.Timestamp {
			return errp.Errorf("invalid dueTime, expected > %d, got %d",
				tx.ld.Timestamp, u)
		}
		tx.dueTime = u
	}

	tx.exSigners, err = tx.ld.ExSigners()
	if err != nil {
		return errp.Errorf("invalid exSignatures, %v", err)
	}
	return nil
}

func (tx *TxBorrow) Apply(ctx ChainContext, cs ChainState) error {
	var err error
	errp := util.ErrPrefix("TxBorrow.Apply error: ")

	if err = tx.TxBase.verify(ctx, cs); err != nil {
		return errp.ErrorIf(err)
	}

	// verify lender's signatures
	if !tx.to.SatisfySigning(tx.exSigners) {
		return errp.Errorf("invalid exSignatures for lending keepers")
	}

	if err = cs.LoadLedger(tx.to); err != nil {
		return errp.ErrorIf(err)
	}

	if err = tx.to.Borrow(
		tx.token, tx.ld.Tx.From, tx.input.Amount, tx.dueTime); err != nil {
		return errp.ErrorIf(err)
	}
	if err = tx.to.SubByNonceTable(
		tx.token, tx.input.Expire, tx.input.Nonce, tx.input.Amount); err != nil {
		return errp.ErrorIf(err)
	}
	if err = tx.from.Add(tx.token, tx.input.Amount); err != nil {
		return errp.ErrorIf(err)
	}
	return errp.ErrorIf(tx.TxBase.accept(ctx, cs))
}
