// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package txn

import (
	"encoding/json"

	"github.com/ldclabs/ldvm/ids"
	"github.com/ldclabs/ldvm/ld"
	"github.com/ldclabs/ldvm/util/erring"
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
	errp := erring.ErrPrefix("txn.TxWithdrawStake.MarshalJSON: ")
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

func (tx *TxWithdrawStake) SyntacticVerify() error {
	var err error
	errp := erring.ErrPrefix("txn.TxWithdrawStake.SyntacticVerify: ")

	if err = tx.TxBase.SyntacticVerify(); err != nil {
		return errp.ErrorIf(err)
	}

	switch {
	case tx.ld.Tx.To == nil:
		return errp.Errorf("nil to as stake account")

	case tx.ld.Tx.Token != nil:
		return errp.Errorf("invalid token, should be nil")

	case tx.ld.Tx.Amount != nil:
		return errp.Errorf("invalid amount, should be nil")

	case len(tx.ld.Tx.Data) == 0:
		return errp.Errorf("invalid data")
	}

	if stake := ids.StakeSymbol(*tx.ld.Tx.To); !stake.Valid() {
		return errp.Errorf("invalid stake account %s", stake.GoString())
	}

	tx.input = &ld.TxTransfer{}
	if err = tx.input.Unmarshal(tx.ld.Tx.Data); err != nil {
		return errp.ErrorIf(err)
	}
	if err = tx.input.SyntacticVerify(); err != nil {
		return errp.ErrorIf(err)
	}

	switch {
	case tx.input.Nonce != 0:
		return errp.Errorf("invalid nonce, expected 0, got %d", tx.input.Nonce)

	case tx.input.From != nil:
		return errp.Errorf("invalid from, expected nil, got %s", tx.input.From)

	case tx.input.To != nil:
		return errp.Errorf("invalid to, expected nil, got %s", tx.input.To)

	case tx.input.Amount == nil:
		return errp.Errorf("nil amount, expected >= 0")
	}

	if tx.input.Token != nil {
		tx.token = *tx.input.Token
	}
	return nil
}

func (tx *TxWithdrawStake) Apply(ctx ChainContext, cs ChainState) error {
	var err error
	errp := erring.ErrPrefix("txn.TxWithdrawStake.Apply: ")

	if err = tx.TxBase.verify(ctx, cs); err != nil {
		return errp.ErrorIf(err)
	}

	if err = cs.LoadLedger(tx.to); err != nil {
		return errp.ErrorIf(err)
	}

	// must WithdrawStake and then accept
	withdraw, err := tx.to.WithdrawStake(tx.token, tx.ld.Tx.From, tx.input.Amount, tx.ld.IsApproved)
	if err != nil {
		return errp.ErrorIf(err)
	}
	if err = tx.to.Sub(tx.token, withdraw); err != nil {
		return errp.ErrorIf(err)
	}
	if err = tx.from.Add(tx.token, withdraw); err != nil {
		return errp.ErrorIf(err)
	}
	return errp.ErrorIf(tx.TxBase.accept(ctx, cs))
}
