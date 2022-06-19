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
	errPrefix := "TxUpdateModelKeepers.SyntacticVerify failed:"
	if err = tx.TxBase.SyntacticVerify(); err != nil {
		return fmt.Errorf("%s %v", errPrefix, err)
	}

	switch {
	case tx.ld.To != nil:
		return fmt.Errorf("%s invalid to, should be nil", errPrefix)
	case tx.ld.Token != nil:
		return fmt.Errorf("%s invalid token, should be nil", errPrefix)
	case tx.ld.Amount != nil:
		return fmt.Errorf("%s invalid amount, should be nil", errPrefix)
	case len(tx.ld.Data) == 0:
		return fmt.Errorf("%s invalid data", errPrefix)
	}

	tx.input = &ld.TxUpdater{}
	if err = tx.input.Unmarshal(tx.ld.Data); err != nil {
		return fmt.Errorf("%s %v", errPrefix, err)
	}
	if err = tx.input.SyntacticVerify(); err != nil {
		return fmt.Errorf("%s %v", errPrefix, err)
	}

	switch {
	case tx.input.ModelID == nil || *tx.input.ModelID == util.ModelIDEmpty:
		return fmt.Errorf("%s invalid mid", errPrefix)
	case tx.input.Threshold == nil && tx.input.Approver == nil:
		return fmt.Errorf("%s nothing to update", errPrefix)
	}
	return nil
}

func (tx *TxUpdateModelKeepers) Verify(bctx BlockContext, bs BlockState) error {
	var err error
	errPrefix := "TxUpdateModelKeepers.Verify failed:"
	if err = tx.TxBase.Verify(bctx, bs); err != nil {
		return fmt.Errorf("%s %v", errPrefix, err)
	}

	tx.mi, err = bs.LoadModel(*tx.input.ModelID)
	switch {
	case err != nil:
		return fmt.Errorf("%s %v", errPrefix, err)

	case !util.SatisfySigningPlus(tx.mi.Threshold, tx.mi.Keepers, tx.signers):
		return fmt.Errorf("%s invalid signatures for keepers", errPrefix)

	case tx.ld.NeedApprove(tx.mi.Approver, nil) && !tx.signers.Has(*tx.mi.Approver):
		return fmt.Errorf("%s invalid signature for approver", errPrefix)
	}
	return nil
}

func (tx *TxUpdateModelKeepers) Accept(bctx BlockContext, bs BlockState) error {
	var err error

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
		return err
	}
	return tx.TxBase.Accept(bctx, bs)
}
