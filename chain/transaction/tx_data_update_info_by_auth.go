// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package transaction

import (
	"encoding/json"

	"github.com/ldclabs/ldvm/constants"
	"github.com/ldclabs/ldvm/ld"
	"github.com/ldclabs/ldvm/util"
)

type TxUpdateDataInfoByAuth struct {
	TxBase
	exSigners util.EthIDs
	input     *ld.TxUpdater
	di        *ld.DataInfo
}

func (tx *TxUpdateDataInfoByAuth) MarshalJSON() ([]byte, error) {
	if tx == nil || tx.ld == nil {
		return []byte("null"), nil
	}

	v := tx.ld.Copy()
	errp := util.ErrPrefix("TxUpdateDataInfoByAuth.MarshalJSON error: ")
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

func (tx *TxUpdateDataInfoByAuth) SyntacticVerify() error {
	var err error
	errp := util.ErrPrefix("TxUpdateDataInfoByAuth.SyntacticVerify error: ")

	if err = tx.TxBase.SyntacticVerify(); err != nil {
		return errp.ErrorIf(err)
	}

	switch {
	case tx.ld.To == nil:
		return errp.Errorf("nil to")
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

	case tx.input.Keepers != nil:
		return errp.Errorf("invalid keepers, should be nil")

	case tx.input.SigClaims != nil:
		return errp.Errorf("invalid sigClaims, should be nil")

	case tx.input.Approver != nil:
		return errp.Errorf("invalid approver, should be nil")

	case tx.input.ApproveList != nil:
		return errp.Errorf("invalid approveList, should be nil")

	case tx.input.To == nil:
		return errp.Errorf("nil to")

	case *tx.input.To != *tx.ld.To:
		return errp.Errorf("invalid to, expected %s, got %s",
			tx.input.To, tx.ld.To)

	case tx.input.Amount == nil:
		return errp.Errorf("nil amount")

	case tx.input.Amount.Cmp(tx.amount) != 0:
		return errp.Errorf("invalid amount, expected %v, got %v",
			tx.input.Amount, tx.amount)

	case tx.input.Token == nil && tx.token != constants.NativeToken:
		return errp.Errorf("invalid token, expected NativeToken, got %s",
			tx.token.GoString())

	case tx.input.Token != nil && tx.token != *tx.input.Token:
		return errp.Errorf("invalid token, expected %s, got %s",
			tx.input.Token.GoString(), tx.token.GoString())
	}

	if tx.exSigners, err = tx.ld.ExSigners(); err != nil {
		return errp.Errorf("invalid exSignatures, %v", err)
	}
	return nil
}

func (tx *TxUpdateDataInfoByAuth) Apply(bctx BlockContext, bs BlockState) error {
	var err error
	errp := util.ErrPrefix("TxUpdateDataInfoByAuth.Apply error: ")

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

	case !util.SatisfySigningPlus(tx.di.Threshold, tx.di.Keepers, tx.exSigners):
		return errp.Errorf("invalid exSignatures for data keepers")

	case tx.ld.NeedApprove(tx.di.Approver, tx.di.ApproveList) && !tx.exSigners.Has(*tx.di.Approver):
		return errp.Errorf("invalid signature for data approver")
	}

	tx.di.Version++
	tx.di.Threshold = tx.from.Threshold()
	tx.di.Keepers = tx.from.Keepers()
	if len(tx.di.Keepers) == 0 {
		tx.di.Threshold = 1
		tx.di.Keepers = util.EthIDs{tx.from.id}
	}
	tx.di.Approver = nil
	tx.di.ApproveList = nil

	if err = bs.SaveData(tx.di); err != nil {
		return errp.ErrorIf(err)
	}
	return errp.ErrorIf(tx.TxBase.accept(bctx, bs))
}
