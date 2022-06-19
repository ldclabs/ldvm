// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package transaction

import (
	"encoding/json"
	"fmt"

	"github.com/ldclabs/ldvm/ld"
	"github.com/ldclabs/ldvm/util"
)

type TxDeleteData struct {
	TxBase
	input *ld.TxUpdater
	di    *ld.DataInfo
}

func (tx *TxDeleteData) MarshalJSON() ([]byte, error) {
	if tx == nil || tx.ld == nil {
		return []byte("null"), nil
	}
	v := tx.ld.Copy()
	if tx.input == nil {
		return nil, fmt.Errorf("TxDeleteData.MarshalJSON error: invalid tx.input")
	}
	d, err := json.Marshal(tx.input)
	if err != nil {
		return nil, err
	}
	v.Data = d
	return json.Marshal(v)
}

func (tx *TxDeleteData) SyntacticVerify() error {
	var err error
	errp := util.ErrPrefix("TxDeleteData.SyntacticVerify error: ")

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

	tx.input = &ld.TxUpdater{}
	if err = tx.input.Unmarshal(tx.ld.Data); err != nil {
		return errp.ErrorIf(err)
	}
	if err = tx.input.SyntacticVerify(); err != nil {
		return errp.ErrorIf(err)
	}

	switch {
	case tx.input.ID == nil || *tx.input.ID == util.DataIDEmpty:
		return errp.Errorf("invalid data id")

	case tx.input.Version == 0:
		return errp.Errorf("invalid data version")
	}
	return nil
}

func (tx *TxDeleteData) Verify(bctx BlockContext, bs BlockState) error {
	var err error
	errp := util.ErrPrefix("TxDeleteData.Verify error: ")

	if err = tx.TxBase.Verify(bctx, bs); err != nil {
		return errp.ErrorIf(err)
	}

	tx.di, err = bs.LoadData(*tx.input.ID)
	switch {
	case err != nil:
		return errp.ErrorIf(err)

	case tx.di.Version != tx.input.Version:
		return errp.Errorf("invalid version, expected %d, got %d",
			tx.di.Version, tx.input.Version)

	case !util.SatisfySigning(tx.di.Threshold, tx.di.Keepers, tx.signers, false):
		return errp.Errorf("invalid signatures for data keepers")

	case tx.ld.NeedApprove(tx.di.Approver, tx.di.ApproveList) && !tx.signers.Has(*tx.di.Approver):
		return errp.Errorf("invalid signature for data approver")
	}
	return nil
}

func (tx *TxDeleteData) Accept(bctx BlockContext, bs BlockState) error {
	var err error
	errp := util.ErrPrefix("TxDeleteData.Accept error: ")

	if err = bs.DeleteData(*tx.input.ID, tx.di, tx.input.Data); err != nil {
		return errp.ErrorIf(err)
	}
	return errp.ErrorIf(tx.TxBase.Accept(bctx, bs))
}
