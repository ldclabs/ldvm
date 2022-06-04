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
	input     *ld.TxUpdater
	dm        *ld.DataMeta
}

func (tx *TxUpdateDataKeepersByAuth) MarshalJSON() ([]byte, error) {
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

	tx.input = &ld.TxUpdater{}
	if err = tx.input.Unmarshal(tx.ld.Data); err != nil {
		return fmt.Errorf("TxUpdateDataKeepersByAuth unmarshal data failed: %v", err)
	}
	if err = tx.input.SyntacticVerify(); err != nil {
		return fmt.Errorf("TxUpdateDataKeepersByAuth syntacticVerify failed: %v", err)
	}
	if tx.input.ID == nil ||
		tx.input.Version == 0 ||
		tx.input.Amount == nil ||
		tx.input.Amount.Cmp(tx.ld.Amount) != 0 ||
		tx.input.To == nil ||
		*tx.input.To != *tx.ld.To {
		return fmt.Errorf("TxUpdateDataKeepersByAuth invalid TxUpdater")
	}
	if tx.input.Token != nil && *tx.input.Token != tx.token {
		return fmt.Errorf("TxUpdateDataKeepersByAuth invalid TxUpdater token")
	}
	return nil
}

func (tx *TxUpdateDataKeepersByAuth) Verify(bctx BlockContext, bs BlockState) error {
	var err error
	if err = tx.TxBase.Verify(bctx, bs); err != nil {
		return err
	}

	tx.dm, err = bs.LoadData(*tx.input.ID)
	if err != nil {
		return fmt.Errorf("TxUpdateDataKeepersByAuth load data failed: %v", err)
	}
	if tx.dm.Version != tx.input.Version {
		return fmt.Errorf("TxUpdateDataKeepersByAuth version mismatch, expected %v, got %v",
			tx.dm.Version, tx.input.Version)
	}
	// verify seller's signatures
	if !util.SatisfySigningPlus(tx.dm.Threshold, tx.dm.Keepers, tx.exSigners) {
		return fmt.Errorf("TxUpdateDataKeepersByAuth need more exSignatures")
	}
	if tx.ld.NeedApprove(tx.dm.Approver, tx.dm.ApproveList) && !tx.exSigners.Has(*tx.dm.Approver) {
		return fmt.Errorf("TxUpdateDataKeepersByAuth.Verify failed: no approver signing")
	}
	if tx.input.KSig != nil {
		kSigner, err := util.DeriveSigner(tx.dm.Data, (*tx.input.KSig)[:])
		if err != nil {
			return fmt.Errorf("TxUpdateDataKeepersByAuth.Verify failed: invalid kSig: %v", err)
		}
		keepers := tx.dm.Keepers
		if tx.input.Keepers != nil {
			keepers = *tx.input.Keepers
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
	tx.dm.Threshold = *tx.input.Threshold
	tx.dm.Keepers = *tx.input.Keepers
	if len(tx.dm.Keepers) == 0 {
		tx.dm.Threshold = tx.from.Threshold()
		tx.dm.Keepers = tx.from.Keepers()
	}
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
	if tx.input.KSig != nil {
		tx.dm.KSig = *tx.input.KSig
	}
	if err = bs.SaveData(*tx.input.ID, tx.dm); err != nil {
		return err
	}
	return tx.TxBase.Accept(bctx, bs)
}
