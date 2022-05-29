// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package transaction

import (
	"encoding/json"
	"fmt"

	"github.com/ldclabs/ldvm/ld"
	"github.com/ldclabs/ldvm/util"
)

type TxUpdateDataKeepersByAuth struct {
	TxBase
	exSigners util.EthIDs
	data      *ld.TxUpdater
	dm        *ld.DataMeta
}

func (tx *TxUpdateDataKeepersByAuth) MarshalJSON() ([]byte, error) {
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

func (tx *TxUpdateDataKeepersByAuth) SyntacticVerify() error {
	var err error
	if err = tx.TxBase.SyntacticVerify(); err != nil {
		return err
	}

	if tx.ld.To == nil {
		return fmt.Errorf("TxUpdateModelKeepers invalid to")
	}

	if len(tx.ld.Data) == 0 {
		return fmt.Errorf("TxUpdateModelKeepers invalid")
	}

	tx.exSigners, err = tx.ld.ExSigners()
	if err != nil {
		return fmt.Errorf("invalid exSignatures: %v", err)
	}

	tx.data = &ld.TxUpdater{}
	if err = tx.data.Unmarshal(tx.ld.Data); err != nil {
		return fmt.Errorf("TxUpdateDataKeepersByAuth unmarshal data failed: %v", err)
	}
	if err = tx.data.SyntacticVerify(); err != nil {
		return fmt.Errorf("TxUpdateDataKeepersByAuth syntacticVerify failed: %v", err)
	}
	if tx.data.ID == nil ||
		tx.data.Version == 0 ||
		tx.data.Amount == nil ||
		tx.data.Amount.Cmp(tx.ld.Amount) != 0 ||
		tx.data.To == nil ||
		*tx.data.To != *tx.ld.To {
		return fmt.Errorf("TxUpdateDataKeepersByAuth invalid TxUpdater")
	}
	if tx.data.Token != nil && *tx.data.Token != tx.token {
		return fmt.Errorf("TxUpdateDataKeepersByAuth invalid TxUpdater token")
	}
	return nil
}

func (tx *TxUpdateDataKeepersByAuth) Verify(bctx BlockContext, bs BlockState) error {
	var err error
	if err = tx.TxBase.Verify(bctx, bs); err != nil {
		return err
	}

	tx.dm, err = bs.LoadData(*tx.data.ID)
	if err != nil {
		return fmt.Errorf("TxUpdateDataKeepersByAuth load data failed: %v", err)
	}
	if tx.dm.Version != tx.data.Version {
		return fmt.Errorf("TxUpdateDataKeepersByAuth version mismatch, expected %v, got %v",
			tx.dm.Version, tx.data.Version)
	}
	// verify seller's signatures
	if !util.SatisfySigningPlus(tx.dm.Threshold, tx.dm.Keepers, tx.exSigners) {
		return fmt.Errorf("TxUpdateDataKeepersByAuth need more exSignatures")
	}
	if tx.ld.NeedApprove(tx.dm.Approver, tx.dm.ApproveList) && !tx.exSigners.Has(*tx.dm.Approver) {
		return fmt.Errorf("TxUpdateDataKeepersByAuth.Verify failed: no approver signing")
	}
	if tx.data.KSig != nil {
		kSigner, err := util.DeriveSigner(tx.dm.Data, (*tx.data.KSig)[:])
		if err != nil {
			return fmt.Errorf("TxUpdateDataKeepersByAuth.Verify failed: invalid kSig: %v", err)
		}
		keepers := tx.data.Keepers
		if len(keepers) == 0 {
			keepers = tx.dm.Keepers
		}
		if !keepers.Has(kSigner) {
			return fmt.Errorf("TxUpdateDataKeepersByAuth.Verify failed: invalid kSig")
		}
	}
	return nil
}

func (tx *TxUpdateDataKeepersByAuth) Accept(bctx BlockContext, bs BlockState) error {
	var err error

	tx.dm.Version++
	tx.dm.Threshold = tx.data.Threshold
	tx.dm.Keepers = tx.data.Keepers
	if len(tx.dm.Keepers) == 0 {
		tx.dm.Threshold = tx.from.Threshold()
		tx.dm.Keepers = tx.from.Keepers()
	}
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
	if tx.data.KSig != nil {
		tx.dm.KSig = *tx.data.KSig
	}
	if err = bs.SaveData(*tx.data.ID, tx.dm); err != nil {
		return err
	}
	return tx.TxBase.Accept(bctx, bs)
}
