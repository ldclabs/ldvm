// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package transaction

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/ldclabs/ldvm/ld"
	"github.com/ldclabs/ldvm/util"
)

type TxUpdateDataKeepers struct {
	TxBase
	input *ld.TxUpdater
	dm    *ld.DataMeta
}

func (tx *TxUpdateDataKeepers) MarshalJSON() ([]byte, error) {
	if tx == nil || tx.ld == nil {
		return []byte("null"), nil
	}
	v := tx.ld.Copy()
	if tx.input == nil {
		return nil, fmt.Errorf("MarshalJSON failed: data not exists")
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
	if err = tx.TxBase.SyntacticVerify(); err != nil {
		return err
	}

	if tx.ld.Token != nil {
		return fmt.Errorf("invalid token, expected NativeToken, got %s",
			strconv.Quote(tx.ld.Token.GoString()))
	}
	if len(tx.ld.Data) == 0 {
		return fmt.Errorf("TxUpdateDataKeepers invalid")
	}
	tx.input = &ld.TxUpdater{}
	if err = tx.input.Unmarshal(tx.ld.Data); err != nil {
		return fmt.Errorf("TxUpdateDataKeepers unmarshal data failed: %v", err)
	}
	if err = tx.input.SyntacticVerify(); err != nil {
		return fmt.Errorf("TxUpdateDataKeepers SyntacticVerify failed: %v", err)
	}
	if tx.input.ID == nil ||
		tx.input.Version == 0 {
		return fmt.Errorf("TxUpdateDataKeepers invalid txUpdater")
	}
	if tx.input.Threshold == nil && tx.input.Approver == nil && tx.input.ApproveList == nil && tx.input.KSig == nil {
		return fmt.Errorf("TxUpdateDataKeepers no thing to update")
	}
	return nil
}

func (tx *TxUpdateDataKeepers) Verify(bctx BlockContext, bs BlockState) error {
	var err error
	if err = tx.TxBase.Verify(bctx, bs); err != nil {
		return err
	}

	tx.dm, err = bs.LoadData(*tx.input.ID)
	if err != nil {
		return fmt.Errorf("TxUpdateDataKeepers load data failed: %v", err)
	}
	if tx.dm.Version != tx.input.Version {
		return fmt.Errorf("TxUpdateDataKeepers version mismatch, expected %v, got %v",
			tx.dm.Version, tx.input.Version)
	}
	if !util.SatisfySigningPlus(tx.dm.Threshold, tx.dm.Keepers, tx.signers) {
		return fmt.Errorf("TxUpdateDataKeepers need more signatures")
	}
	if tx.ld.NeedApprove(tx.dm.Approver, tx.dm.ApproveList) && !tx.signers.Has(*tx.dm.Approver) {
		return fmt.Errorf("TxUpdateDataKeepers.Verify failed: no approver signing")
	}
	if tx.input.KSig != nil {
		kSigner, err := util.DeriveSigner(tx.dm.Data, (*tx.input.KSig)[:])
		if err != nil {
			return fmt.Errorf("TxUpdateDataKeepers.Verify failed: invalid kSig: %v", err)
		}
		keepers := tx.input.Keepers
		if len(keepers) == 0 {
			keepers = tx.dm.Keepers
		}
		if !keepers.Has(kSigner) {
			return fmt.Errorf("TxUpdateDataKeepers.Verify failed: invalid kSig")
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
		} else {
			tx.dm.Approver = tx.input.Approver
		}
	}
	if tx.input.ApproveList != nil {
		tx.dm.ApproveList = tx.input.ApproveList
	}
	if tx.input.Threshold != nil {
		tx.dm.Threshold = *tx.input.Threshold
		tx.dm.Keepers = tx.input.Keepers
	}
	if tx.input.KSig != nil {
		tx.dm.KSig = *tx.input.KSig
	}
	if err = bs.SaveData(*tx.input.ID, tx.dm); err != nil {
		return err
	}
	return tx.TxBase.Accept(bctx, bs)
}
