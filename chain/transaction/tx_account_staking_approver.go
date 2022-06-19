// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package transaction

import (
	"encoding/json"
	"fmt"

	"github.com/ldclabs/ldvm/ld"
	"github.com/ldclabs/ldvm/util"
)

type TxUpdateStakeApprover struct {
	TxBase
	input *ld.TxAccounter
}

func (tx *TxUpdateStakeApprover) MarshalJSON() ([]byte, error) {
	if tx == nil || tx.ld == nil {
		return []byte("null"), nil
	}
	v := tx.ld.Copy()
	if tx.input == nil {
		return nil, fmt.Errorf("TxUpdateStakeApprover.MarshalJSON error: invalid tx.input")
	}
	d, err := json.Marshal(tx.input)
	if err != nil {
		return nil, err
	}
	v.Data = d
	return json.Marshal(v)
}

func (tx *TxUpdateStakeApprover) SyntacticVerify() error {
	var err error
	errp := util.ErrPrefix("TxUpdateStakeApprover.SyntacticVerify error: ")

	if err = tx.TxBase.SyntacticVerify(); err != nil {
		return errp.ErrorIf(err)
	}

	switch {
	case tx.ld.To == nil:
		return errp.Errorf("nil to as stake account")

	case tx.ld.Amount != nil:
		return errp.Errorf("invalid amount, should be nil")

	case tx.ld.Token != nil:
		return errp.Errorf("invalid token, should be nil")

	case len(tx.ld.Data) == 0:
		return errp.Errorf("invalid data")
	}

	if stake := util.StakeSymbol(*tx.ld.To); !stake.Valid() {
		return errp.Errorf("invalid stake account %s", stake.GoString())
	}

	tx.input = &ld.TxAccounter{}
	if err = tx.input.Unmarshal(tx.ld.Data); err != nil {
		return errp.ErrorIf(err)
	}
	if err = tx.input.SyntacticVerify(); err != nil {
		return errp.ErrorIf(err)
	}

	switch {
	case tx.input.Approver == nil:
		return errp.Errorf("nil approver")

	case tx.input.ApproveList != nil:
		return errp.Errorf("invalid approveList, should be nil")
	}

	return nil
}

func (tx *TxUpdateStakeApprover) Verify(bctx BlockContext, bs BlockState) error {
	var err error
	errp := util.ErrPrefix("TxUpdateStakeApprover.Verify error: ")

	if err = tx.TxBase.Verify(bctx, bs); err != nil {
		return errp.ErrorIf(err)
	}
	if err = tx.to.CheckUpdateStakeApprover(tx.ld.From, *tx.input.Approver, tx.signers); err != nil {
		return errp.ErrorIf(err)
	}
	return nil
}

func (tx *TxUpdateStakeApprover) Accept(bctx BlockContext, bs BlockState) error {
	errp := util.ErrPrefix("TxUpdateStakeApprover.Accept error: ")

	err := tx.to.UpdateStakeApprover(tx.ld.From, *tx.input.Approver, tx.signers)
	if err != nil {
		return errp.ErrorIf(err)
	}
	return errp.ErrorIf(tx.TxBase.Accept(bctx, bs))
}
