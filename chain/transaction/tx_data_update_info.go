// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package transaction

import (
	"encoding/json"

	"github.com/ldclabs/ldvm/ld"
	"github.com/ldclabs/ldvm/util"
)

type TxUpdateDataInfo struct {
	TxBase
	input *ld.TxUpdater
	di    *ld.DataInfo
}

func (tx *TxUpdateDataInfo) MarshalJSON() ([]byte, error) {
	if tx == nil || tx.ld == nil {
		return []byte("null"), nil
	}

	v := tx.ld.Copy()
	errp := util.ErrPrefix("TxUpdateDataInfo.MarshalJSON error: ")
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

func (tx *TxUpdateDataInfo) SyntacticVerify() error {
	var err error
	errp := util.ErrPrefix("TxUpdateDataInfo.SyntacticVerify error: ")

	if err = tx.TxBase.SyntacticVerify(); err != nil {
		return errp.ErrorIf(err)
	}

	switch {
	case tx.ld.To != nil:
		return errp.Errorf("invalid to, should be nil")

	case tx.ld.Token != nil:
		return errp.Errorf("invalid token, should be nil")

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

	case tx.input.Threshold == nil && tx.input.Approver == nil &&
		tx.input.ApproveList == nil && tx.input.SigClaims == nil:
		return errp.Errorf("no thing to update")
	}

	return nil
}

func (tx *TxUpdateDataInfo) Apply(bctx BlockContext, bs BlockState) error {
	var err error
	errp := util.ErrPrefix("TxUpdateDataInfo.Apply error: ")

	if err = tx.TxBase.verify(bctx, bs); err != nil {
		return errp.ErrorIf(err)
	}

	tx.di, err = bs.LoadData(*tx.input.ID)
	switch {
	case err != nil:
		return errp.ErrorIf(err)

	case tx.di.Version != tx.input.Version:
		return errp.Errorf("invalid version, expected %d, got %d",
			tx.di.Version, tx.input.Version)

	case !util.SatisfySigningPlus(tx.di.Threshold, tx.di.Keepers, tx.signers):
		return errp.Errorf("invalid signatures for data keepers")

	case tx.ld.NeedApprove(tx.di.Approver, tx.di.ApproveList) &&
		!tx.signers.Has(*tx.di.Approver):
		return errp.Errorf("invalid signature for data approver")
	}

	tx.di.Version++
	if tx.input.Approver != nil {
		if *tx.input.Approver == util.EthIDEmpty {
			tx.di.Approver = nil
			tx.di.ApproveList = nil
		} else {
			tx.di.Approver = tx.input.Approver
		}
	}

	if tx.input.ApproveList != nil {
		tx.di.ApproveList = tx.input.ApproveList
	}

	if tx.input.Threshold != nil {
		tx.di.Threshold = *tx.input.Threshold
		tx.di.Keepers = *tx.input.Keepers
	}

	if tx.input.SigClaims != nil {
		tx.di.SigClaims = tx.input.SigClaims
		tx.di.Sig = tx.input.Sig

		if _, err = tx.di.Signer(); err != nil {
			return errp.ErrorIf(err)
		}
	}

	if err = bs.SaveData(tx.di); err != nil {
		return errp.ErrorIf(err)
	}
	return errp.ErrorIf(tx.TxBase.accept(bctx, bs))
}
