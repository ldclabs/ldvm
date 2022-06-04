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
	mm    *ld.ModelMeta
}

func (tx *TxUpdateModelKeepers) MarshalJSON() ([]byte, error) {
	if tx == nil || tx.ld == nil {
		return []byte("null"), nil
	}
	v := tx.ld.Copy()
	if tx.input == nil {
		return nil, fmt.Errorf("TxUpdateModelKeepers.MarshalJSON failed: invalid tx.input")
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
	if err = tx.TxBase.SyntacticVerify(); err != nil {
		return err
	}

	switch {
	case tx.ld.To != nil:
		return fmt.Errorf("TxUpdateModelKeepers.SyntacticVerify failed: invalid to, should be nil")
	case tx.ld.Token != nil:
		return fmt.Errorf("TxUpdateModelKeepers.SyntacticVerify failed: invalid token, should be nil")
	case tx.ld.Amount != nil:
		return fmt.Errorf("TxUpdateModelKeepers.SyntacticVerify failed: invalid amount, should be nil")
	case len(tx.ld.Data) == 0:
		return fmt.Errorf("TxUpdateModelKeepers.SyntacticVerify failed: invalid data")
	}

	tx.input = &ld.TxUpdater{}
	if err = tx.input.Unmarshal(tx.ld.Data); err != nil {
		return fmt.Errorf("TxUpdateModelKeepers.SyntacticVerify failed: %v", err)
	}
	if err = tx.input.SyntacticVerify(); err != nil {
		return fmt.Errorf("TxUpdateModelKeepers.SyntacticVerify failed: %v", err)
	}

	switch {
	case tx.input.ModelID == nil || *tx.input.ModelID == util.ModelIDEmpty:
		return fmt.Errorf("TxUpdateModelKeepers.SyntacticVerify failed: invalid mid")
	case tx.input.Threshold == nil && tx.input.Approver == nil:
		return fmt.Errorf("TxUpdateModelKeepers.SyntacticVerify failed: nothing to update")
	}
	return nil
}

func (tx *TxUpdateModelKeepers) Verify(bctx BlockContext, bs BlockState) error {
	var err error
	if err = tx.TxBase.Verify(bctx, bs); err != nil {
		return err
	}

	tx.mm, err = bs.LoadModel(*tx.input.ModelID)
	if err != nil {
		return fmt.Errorf("TxUpdateModelKeepers.Verify failed: %v", err)
	}

	if !util.SatisfySigningPlus(tx.mm.Threshold, tx.mm.Keepers, tx.signers) {
		return fmt.Errorf("TxUpdateModelKeepers.Verify failed: invalid signature for keepers")
	}

	if tx.ld.NeedApprove(tx.mm.Approver, nil) && !tx.signers.Has(*tx.mm.Approver) {
		return fmt.Errorf("TxUpdateModelKeepers.Verify failed: invalid signature for approver")
	}
	return nil
}

func (tx *TxUpdateModelKeepers) Accept(bctx BlockContext, bs BlockState) error {
	var err error

	if tx.input.Approver != nil {
		if *tx.input.Approver == util.EthIDEmpty {
			tx.mm.Approver = nil
		} else {
			tx.mm.Approver = tx.input.Approver
		}
	}
	if tx.input.Threshold != nil {
		tx.mm.Threshold = *tx.input.Threshold
		tx.mm.Keepers = *tx.input.Keepers
	}
	if err = bs.SaveModel(*tx.input.ModelID, tx.mm); err != nil {
		return err
	}
	return tx.TxBase.Accept(bctx, bs)
}
