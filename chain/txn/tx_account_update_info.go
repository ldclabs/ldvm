// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package txn

import (
	"encoding/json"
	"math/big"

	"github.com/ldclabs/ldvm/ids"
	"github.com/ldclabs/ldvm/ld"
	"github.com/ldclabs/ldvm/util/erring"
)

type TxUpdateAccountInfo struct {
	TxBase
	input *ld.TxAccounter
}

func (tx *TxUpdateAccountInfo) MarshalJSON() ([]byte, error) {
	if tx == nil || tx.ld == nil {
		return []byte("null"), nil
	}

	v := tx.ld.Copy()
	errp := erring.ErrPrefix("txn.TxUpdateAccountInfo.MarshalJSON: ")
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

func (tx *TxUpdateAccountInfo) SyntacticVerify() error {
	var err error
	errp := erring.ErrPrefix("txn.TxUpdateAccountInfo.SyntacticVerify: ")

	if err = tx.TxBase.SyntacticVerify(); err != nil {
		return errp.ErrorIf(err)
	}

	switch {
	case tx.ld.Tx.To != nil:
		return errp.Errorf("invalid to, should be nil")

	case tx.ld.Tx.Token != nil:
		return errp.Errorf("invalid token, should be nil")

	case tx.ld.Tx.Amount != nil:
		return errp.Errorf("invalid amount, should be nil")

	case len(tx.ld.Tx.Data) == 0:
		return errp.Errorf("invalid data")
	}

	tx.input = &ld.TxAccounter{}
	if err = tx.input.Unmarshal(tx.ld.Tx.Data); err != nil {
		return errp.ErrorIf(err)
	}
	if err = tx.input.SyntacticVerify(); err != nil {
		return errp.ErrorIf(err)
	}

	if tx.input.Threshold == nil && tx.input.Approver == nil && tx.input.ApproveList == nil {
		return errp.Errorf("no keepers nor approver")
	}
	if tx.input.Threshold != nil && *tx.input.Threshold == 0 {
		return errp.Errorf("invalid threshold, expected >= 1")
	}

	if tx.input.Keepers != nil {
		tx.senderKey = tx.input.Keepers.FindKeyOrAddr(tx.ld.Tx.From)
	}
	return nil
}

// ApplyGenesis skipping signature verification
func (tx *TxUpdateAccountInfo) ApplyGenesis(ctx ChainContext, cs ChainState) error {
	var err error
	errp := erring.ErrPrefix("txn.TxUpdateAccountInfo.ApplyGenesis: ")

	tx.input = &ld.TxAccounter{}
	if err = tx.input.Unmarshal(tx.ld.Tx.Data); err != nil {
		return errp.ErrorIf(err)
	}
	if err = tx.input.SyntacticVerify(); err != nil {
		return errp.ErrorIf(err)
	}

	tx.amount = new(big.Int)
	tx.tip = new(big.Int)
	tx.fee = new(big.Int)
	tx.cost = new(big.Int)

	if tx.ldc, err = cs.LoadAccount(ids.LDCAccount); err != nil {
		return errp.ErrorIf(err)
	}
	if tx.miner, err = cs.LoadBuilder(ctx.Builder()); err != nil {
		return errp.ErrorIf(err)
	}
	if tx.from, err = cs.LoadAccount(tx.ld.Tx.From); err != nil {
		return errp.ErrorIf(err)
	}

	if err = tx.from.UpdateKeepers(
		tx.input.Threshold, tx.input.Keepers, tx.input.Approver, tx.input.ApproveList); err != nil {
		return errp.ErrorIf(err)
	}

	return errp.ErrorIf(tx.TxBase.accept(ctx, cs))
}

func (tx *TxUpdateAccountInfo) Apply(ctx ChainContext, cs ChainState) error {
	var err error
	errp := erring.ErrPrefix("txn.TxUpdateAccountInfo.Apply: ")

	if err = tx.TxBase.verify(ctx, cs); err != nil {
		return errp.ErrorIf(err)
	}

	if !tx.from.VerifyPlus(tx.ld.TxHash(), tx.ld.Signatures, tx.senderKey) {
		return errp.Errorf("invalid signatures for keepers")
	}

	if err = tx.from.UpdateKeepers(
		tx.input.Threshold, tx.input.Keepers, tx.input.Approver, tx.input.ApproveList); err != nil {
		return errp.ErrorIf(err)
	}

	return errp.ErrorIf(tx.TxBase.accept(ctx, cs))
}
