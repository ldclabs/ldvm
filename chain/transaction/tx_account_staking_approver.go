// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package transaction

import (
	"encoding/json"

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
	errp := util.ErrPrefix("TxUpdateStakeApprover.MarshalJSON error: ")
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

func (tx *TxUpdateStakeApprover) Apply(bctx BlockContext, bs BlockState) error {
	var err error
	errp := util.ErrPrefix("TxUpdateStakeApprover.Apply error: ")

	if err = tx.TxBase.verify(bctx, bs); err != nil {
		return errp.ErrorIf(err)
	}

	if err = tx.to.UpdateStakeApprover(tx.ld.From, *tx.input.Approver, tx.signers); err != nil {
		return errp.ErrorIf(err)
	}
	return errp.ErrorIf(tx.TxBase.accept(bctx, bs))
}
