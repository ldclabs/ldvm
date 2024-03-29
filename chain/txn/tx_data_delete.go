// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package txn

import (
	"encoding/json"

	"github.com/ldclabs/ldvm/ids"
	"github.com/ldclabs/ldvm/ld"
	"github.com/ldclabs/ldvm/ld/service"
	"github.com/ldclabs/ldvm/util/erring"
	"github.com/ldclabs/ldvm/util/validating"
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
	errp := erring.ErrPrefix("txn.TxDeleteData.MarshalJSON: ")
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

func (tx *TxDeleteData) SyntacticVerify() error {
	var err error
	errp := erring.ErrPrefix("txn.TxDeleteData.SyntacticVerify: ")

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

	tx.input = &ld.TxUpdater{}
	if err = tx.input.Unmarshal(tx.ld.Tx.Data); err != nil {
		return errp.ErrorIf(err)
	}
	if err = tx.input.SyntacticVerify(); err != nil {
		return errp.ErrorIf(err)
	}

	switch {
	case tx.input.ID == nil || *tx.input.ID == ids.EmptyDataID:
		return errp.Errorf("invalid data id")

	case tx.input.Version == 0:
		return errp.Errorf("invalid data version")

	case !validating.ValidMessage(string(tx.input.Data)):
		return errp.Errorf("invalid deleting message")
	}
	return nil
}

func (tx *TxDeleteData) Apply(ctx ChainContext, cs ChainState) error {
	var err error
	errp := erring.ErrPrefix("txn.TxDeleteData.Apply: ")

	if err = tx.TxBase.verify(ctx, cs); err != nil {
		return errp.ErrorIf(err)
	}

	tx.di, err = cs.LoadData(*tx.input.ID)
	switch {
	case err != nil:
		return errp.ErrorIf(err)

	case tx.di.Version != tx.input.Version:
		return errp.Errorf("invalid version, expected %d, got %d",
			tx.di.Version, tx.input.Version)

	case !tx.di.VerifyPlus(tx.ld.TxHash(), tx.ld.Signatures):
		return errp.Errorf("invalid signatures for data keepers")

	case !tx.ld.IsApproved(tx.di.Approver, tx.di.ApproveList, false):
		return errp.Errorf("invalid signature for data approver")
	}

	if ctx.ChainConfig().IsNameService(tx.di.ModelID) {
		ns := &service.Name{}
		if err = ns.Unmarshal(tx.di.Payload); err != nil {
			return errp.ErrorIf(err)
		}
		if err = ns.SyntacticVerify(); err != nil {
			return errp.ErrorIf(err)
		}

		ns.DataID = tx.di.ID
		if err = cs.DeleteName(ns); err != nil {
			return errp.ErrorIf(err)
		}
	}

	if err = cs.DeleteData(tx.di, tx.input.Data); err != nil {
		return errp.ErrorIf(err)
	}
	return errp.ErrorIf(tx.TxBase.accept(ctx, cs))
}
