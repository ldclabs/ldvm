// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package transaction

import (
	"encoding/json"
	"math/big"

	"github.com/ldclabs/ldvm/constants"
	"github.com/ldclabs/ldvm/ld"
	"github.com/ldclabs/ldvm/util"
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
	errp := util.ErrPrefix("TxUpdateAccountInfo.MarshalJSON error: ")
	if tx.input == nil {
		return nil, errp.Errorf("nil tx.input")
	}
	d, err := json.Marshal(tx.input)
	if err != nil {
		return nil, errp.ErrorIf(err)
	}
	v.Data = d
	return errp.ErrorMap(json.Marshal(v))
}

func (tx *TxUpdateAccountInfo) SyntacticVerify() error {
	var err error
	errp := util.ErrPrefix("TxUpdateAccountInfo.SyntacticVerify error: ")

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

	tx.input = &ld.TxAccounter{}
	if err = tx.input.Unmarshal(tx.ld.Data); err != nil {
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
	return nil
}

// ApplyGenesis skipping signature verification
func (tx *TxUpdateAccountInfo) ApplyGenesis(bctx BlockContext, bs BlockState) error {
	var err error
	errp := util.ErrPrefix("TxUpdateAccountInfo.ApplyGenesis error: ")

	tx.input = &ld.TxAccounter{}
	if err = tx.input.Unmarshal(tx.ld.Data); err != nil {
		return errp.ErrorIf(err)
	}
	if err = tx.input.SyntacticVerify(); err != nil {
		return errp.ErrorIf(err)
	}

	tx.amount = new(big.Int)
	tx.tip = new(big.Int)
	tx.fee = new(big.Int)
	tx.cost = new(big.Int)

	if tx.ldc, err = bs.LoadAccount(constants.LDCAccount); err != nil {
		return errp.ErrorIf(err)
	}
	if tx.miner, err = bs.LoadMiner(bctx.Miner()); err != nil {
		return errp.ErrorIf(err)
	}
	if tx.from, err = bs.LoadAccount(tx.ld.From); err != nil {
		return errp.ErrorIf(err)
	}

	if err = tx.from.UpdateKeepers(
		tx.input.Threshold, tx.input.Keepers, tx.input.Approver, tx.input.ApproveList); err != nil {
		return errp.ErrorIf(err)
	}

	return errp.ErrorIf(tx.TxBase.accept(bctx, bs))
}

func (tx *TxUpdateAccountInfo) Apply(bctx BlockContext, bs BlockState) error {
	var err error
	errp := util.ErrPrefix("TxUpdateAccountInfo.Apply error: ")

	if err = tx.TxBase.verify(bctx, bs); err != nil {
		return errp.ErrorIf(err)
	}
	if !tx.from.SatisfySigningPlus(tx.signers) {
		return errp.Errorf("invalid signatures for keepers")
	}

	if err = tx.from.UpdateKeepers(
		tx.input.Threshold, tx.input.Keepers, tx.input.Approver, tx.input.ApproveList); err != nil {
		return errp.ErrorIf(err)
	}

	return errp.ErrorIf(tx.TxBase.accept(bctx, bs))
}