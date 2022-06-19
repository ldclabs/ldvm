// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package transaction

import (
	"encoding/json"
	"fmt"

	"github.com/ldclabs/ldvm/ld"
	"github.com/ldclabs/ldvm/util"
)

type TxUpdateModelKeepers struct {
	TxBase
	input *ld.TxUpdater
	mi    *ld.ModelInfo
}

func (tx *TxUpdateModelKeepers) MarshalJSON() ([]byte, error) {
	if tx == nil || tx.ld == nil {
		return []byte("null"), nil
	}
	v := tx.ld.Copy()
	if tx.input == nil {
		return nil, fmt.Errorf("TxUpdateModelKeepers.MarshalJSON error: invalid tx.input")
	}
	d, err := json.Marshal(tx.input)
	if err != nil {
		return nil, err
	}
	v.Data = d
	return json.Marshal(v)
}

func (tx *TxUpdateModelKeepers) SyntacticVerify() error {
	var err error
	errp := util.ErrPrefix("TxUpdateModelKeepers.SyntacticVerify error: ")

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
	case tx.input.ModelID == nil || *tx.input.ModelID == util.ModelIDEmpty:
		return errp.Errorf("invalid mid")
	case tx.input.Threshold == nil && tx.input.Approver == nil:
		return errp.Errorf("nothing to update")
	}
	return nil
}

func (tx *TxUpdateModelKeepers) Verify(bctx BlockContext, bs BlockState) error {
	var err error
	errp := util.ErrPrefix("TxUpdateModelKeepers.Verify error: ")

	if err = tx.TxBase.Verify(bctx, bs); err != nil {
		return errp.ErrorIf(err)
	}

	tx.mi, err = bs.LoadModel(*tx.input.ModelID)
	switch {
	case err != nil:
		return errp.ErrorIf(err)

	case !util.SatisfySigningPlus(tx.mi.Threshold, tx.mi.Keepers, tx.signers):
		return errp.Errorf("invalid signatures for keepers")

	case tx.ld.NeedApprove(tx.mi.Approver, nil) && !tx.signers.Has(*tx.mi.Approver):
		return errp.Errorf("invalid signature for approver")
	}
	return nil
}

func (tx *TxUpdateModelKeepers) Accept(bctx BlockContext, bs BlockState) error {
	var err error
	errp := util.ErrPrefix("TxUpdateModelKeepers.Accept error: ")

	if tx.input.Approver != nil {
		if *tx.input.Approver == util.EthIDEmpty {
			tx.mi.Approver = nil
		} else {
			tx.mi.Approver = tx.input.Approver
		}
	}
	if tx.input.Threshold != nil {
		tx.mi.Threshold = *tx.input.Threshold
		tx.mi.Keepers = *tx.input.Keepers
	}
	if err = bs.SaveModel(*tx.input.ModelID, tx.mi); err != nil {
		return errp.ErrorIf(err)
	}
	return errp.ErrorIf(tx.TxBase.Accept(bctx, bs))
}
