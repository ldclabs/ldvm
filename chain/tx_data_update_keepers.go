// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package chain

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/ldclabs/ldvm/ld"
	"github.com/ldclabs/ldvm/util"
)

type TxUpdateDataKeepers struct {
	TxBase
	data *ld.TxUpdater
	dm   *ld.DataMeta
}

func (tx *TxUpdateDataKeepers) MarshalJSON() ([]byte, error) {
	if tx == nil || tx.ld == nil {
		return []byte("null"), nil
	}
	v := tx.ld.Copy()
	if tx.data == nil {
		return nil, fmt.Errorf("MarshalJSON failed: data not exists")
	}
	d, err := json.Marshal(tx.data)
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
	tx.data = &ld.TxUpdater{}
	if err = tx.data.Unmarshal(tx.ld.Data); err != nil {
		return fmt.Errorf("TxUpdateDataKeepers unmarshal data failed: %v", err)
	}
	if err = tx.data.SyntacticVerify(); err != nil {
		return fmt.Errorf("TxUpdateDataKeepers SyntacticVerify failed: %v", err)
	}
	if tx.data.ID == nil ||
		tx.data.Version == 0 {
		return fmt.Errorf("TxUpdateDataKeepers invalid txUpdater")
	}
	if len(tx.data.Keepers) == 0 && tx.data.Approver == nil && tx.data.ApproveList == nil && tx.data.KSig == nil {
		return fmt.Errorf("TxUpdateDataKeepers no thing to update")
	}
	return nil
}

func (tx *TxUpdateDataKeepers) Verify(blk *Block, bs BlockState) error {
	var err error
	if err = tx.TxBase.Verify(blk, bs); err != nil {
		return err
	}

	tx.dm, err = bs.LoadData(*tx.data.ID)
	if err != nil {
		return fmt.Errorf("TxUpdateDataKeepers load data failed: %v", err)
	}
	if tx.dm.Version != tx.data.Version {
		return fmt.Errorf("TxUpdateDataKeepers version mismatch, expected %v, got %v",
			tx.dm.Version, tx.data.Version)
	}
	if !util.SatisfySigningPlus(tx.dm.Threshold, tx.dm.Keepers, tx.signers) {
		return fmt.Errorf("TxUpdateDataKeepers need more signatures")
	}
	if tx.ld.NeedApprove(tx.dm.Approver, tx.dm.ApproveList) && !tx.signers.Has(*tx.dm.Approver) {
		return fmt.Errorf("TxUpdateDataKeepers.Verify failed: no approver signing")
	}
	if tx.data.KSig != nil {
		kSigner, err := util.DeriveSigner(tx.dm.Data, (*tx.data.KSig)[:])
		if err != nil {
			return fmt.Errorf("TxUpdateDataKeepers.Verify failed: invalid kSig: %v", err)
		}
		keepers := tx.data.Keepers
		if len(keepers) == 0 {
			keepers = tx.dm.Keepers
		}
		if !keepers.Has(kSigner) {
			return fmt.Errorf("TxUpdateDataKeepers.Verify failed: invalid kSig")
		}
	}
	return nil
}

func (tx *TxUpdateDataKeepers) Accept(blk *Block, bs BlockState) error {
	var err error

	tx.dm.Version++
	if tx.data.Approver != nil {
		if *tx.data.Approver == util.EthIDEmpty {
			tx.dm.Approver = nil
		} else {
			tx.dm.Approver = tx.data.Approver
		}
	}
	if tx.data.ApproveList != nil {
		tx.dm.ApproveList = tx.data.ApproveList
	}
	if len(tx.data.Keepers) > 0 {
		tx.dm.Threshold = tx.data.Threshold
		tx.dm.Keepers = tx.data.Keepers
	}
	if tx.data.KSig != nil {
		tx.dm.KSig = *tx.data.KSig
	}
	if err = bs.SaveData(*tx.data.ID, tx.dm); err != nil {
		return err
	}
	return tx.TxBase.Accept(blk, bs)
}
