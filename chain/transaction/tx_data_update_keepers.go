// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package transaction

import (
	"encoding/json"
	"fmt"

	"github.com/ldclabs/ldvm/ld"
	"github.com/ldclabs/ldvm/util"
)

type TxUpdateDataKeepers struct {
	TxBase
	input *ld.TxUpdater
	dm    *ld.DataInfo
}

func (tx *TxUpdateDataKeepers) MarshalJSON() ([]byte, error) {
	if tx == nil || tx.ld == nil {
		return []byte("null"), nil
	}
	v := tx.ld.Copy()
	if tx.input == nil {
		return nil, fmt.Errorf("TxUpdateDataKeepers.MarshalJSON failed: invalid tx.input")
	}
	d, err := json.Marshal(tx.input)
	if err != nil {
		return nil, err
	}
	v.Data = d
	return json.Marshal(v)
}

func (tx *TxUpdateDataKeepers) SyntacticVerify() error {
	var err error
	errPrefix := "TxUpdateDataKeepers.SyntacticVerify failed:"
	if err = tx.TxBase.SyntacticVerify(); err != nil {
		return fmt.Errorf("%s %v", errPrefix, err)
	}

	switch {
	case tx.ld.To != nil:
		return fmt.Errorf("%s invalid to, should be nil", errPrefix)

	case tx.ld.Token != nil:
		return fmt.Errorf("%s invalid token, should be nil", errPrefix)

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
	case tx.input.ID == nil || *tx.input.ID == util.DataIDEmpty:
		return fmt.Errorf("%s invalid data id", errPrefix)

	case tx.input.Version == 0:
		return fmt.Errorf("%s invalid data version", errPrefix)

	case tx.input.Threshold == nil && tx.input.Approver == nil && tx.input.ApproveList == nil:
		return fmt.Errorf("%s no thing to update", errPrefix)
	}

	return nil
}

func (tx *TxUpdateDataKeepers) Verify(bctx BlockContext, bs BlockState) error {
	var err error
	errPrefix := "TxUpdateDataKeepers.Verify failed:"
	if err = tx.TxBase.Verify(bctx, bs); err != nil {
		return fmt.Errorf("%s %v", errPrefix, err)
	}

	tx.dm, err = bs.LoadData(*tx.input.ID)
	switch {
	case err != nil:
		return fmt.Errorf("%s %v", errPrefix, err)

	case tx.dm.Version != tx.input.Version:
		return fmt.Errorf("%s invalid version, expected %d, got %d",
			errPrefix, tx.dm.Version, tx.input.Version)

	case !util.SatisfySigningPlus(tx.dm.Threshold, tx.dm.Keepers, tx.signers):
		return fmt.Errorf("%s invalid signatures for data keepers", errPrefix)

	case tx.ld.NeedApprove(tx.dm.Approver, tx.dm.ApproveList) &&
		!tx.signers.Has(*tx.dm.Approver):
		return fmt.Errorf("%s invalid signature for data approver", errPrefix)
	}

	if tx.input.KSig != nil {
		kSigner, err := util.DeriveSigner(tx.dm.Data, (*tx.input.KSig)[:])
		if err != nil {
			return fmt.Errorf("%s invalid kSig: %v", errPrefix, err)
		}
		keepers := tx.dm.Keepers
		if tx.input.Keepers != nil {
			keepers = *tx.input.Keepers
		}
		if !keepers.Has(kSigner) {
			return fmt.Errorf("%s invalid kSig", errPrefix)
		}
	}
	return nil
}

func (tx *TxUpdateDataKeepers) Accept(bctx BlockContext, bs BlockState) error {
	var err error

	tx.dm.Version++
	if tx.input.Approver != nil {
		if *tx.input.Approver == util.EthIDEmpty {
			tx.dm.Approver = nil
			tx.dm.ApproveList = nil
		} else {
			tx.dm.Approver = tx.input.Approver
		}
	}
	if tx.input.ApproveList != nil {
		tx.dm.ApproveList = tx.input.ApproveList
	}
	if tx.input.Threshold != nil {
		tx.dm.Threshold = *tx.input.Threshold
		tx.dm.Keepers = *tx.input.Keepers
	}
	if tx.input.KSig != nil {
		tx.dm.KSig = *tx.input.KSig
	}
	if err = bs.SaveData(*tx.input.ID, tx.dm); err != nil {
		return err
	}
	return tx.TxBase.Accept(bctx, bs)
}
