// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package txn

import (
	"encoding/json"

	"github.com/ldclabs/ldvm/ids"
	"github.com/ldclabs/ldvm/ld"
	"github.com/ldclabs/ldvm/signer"
	"github.com/ldclabs/ldvm/util/erring"
)

type TxUpdateModelInfo struct {
	TxBase
	input *ld.TxUpdater
	mi    *ld.ModelInfo
}

func (tx *TxUpdateModelInfo) MarshalJSON() ([]byte, error) {
	if tx == nil || tx.ld == nil {
		return []byte("null"), nil
	}

	v := tx.ld.Copy()
	errp := erring.ErrPrefix("txn.TxUpdateModelInfo.MarshalJSON: ")
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

func (tx *TxUpdateModelInfo) SyntacticVerify() error {
	var err error
	errp := erring.ErrPrefix("txn.TxUpdateModelInfo.SyntacticVerify: ")

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
	case tx.input.ModelID == nil || *tx.input.ModelID == ids.EmptyModelID:
		return errp.Errorf("invalid mid")
	case tx.input.Threshold == nil && tx.input.Approver == nil:
		return errp.Errorf("nothing to update")
	}
	return nil
}

func (tx *TxUpdateModelInfo) Apply(ctx ChainContext, cs ChainState) error {
	var err error
	errp := erring.ErrPrefix("txn.TxUpdateModelInfo.Apply: ")

	if err = tx.TxBase.verify(ctx, cs); err != nil {
		return errp.ErrorIf(err)
	}

	tx.mi, err = cs.LoadModel(*tx.input.ModelID)
	switch {
	case err != nil:
		return errp.ErrorIf(err)

	case !tx.mi.VerifyPlus(tx.ld.TxHash(), tx.ld.Signatures):
		return errp.Errorf("invalid signatures for keepers")

	case !tx.ld.IsApproved(tx.mi.Approver, nil, false):
		return errp.Errorf("invalid signature for approver")
	}

	if tx.input.Approver != nil {
		if tx.input.Approver.Kind() == signer.Unknown {
			tx.mi.Approver = nil
		} else {
			tx.mi.Approver = *tx.input.Approver
		}
	}

	if tx.input.Threshold != nil {
		tx.mi.Threshold = *tx.input.Threshold
		tx.mi.Keepers = *tx.input.Keepers
	}
	if err = cs.SaveModel(tx.mi); err != nil {
		return errp.ErrorIf(err)
	}
	return errp.ErrorIf(tx.TxBase.accept(ctx, cs))
}
