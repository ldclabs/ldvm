// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package transaction

import (
	"encoding/json"
	"fmt"

	"github.com/ldclabs/ldvm/ld"
	"github.com/ldclabs/ldvm/util"
)

type TxDeleteData struct {
	TxBase
	input *ld.TxUpdater
	dm    *ld.DataMeta
}

func (tx *TxDeleteData) MarshalJSON() ([]byte, error) {
	if tx == nil || tx.ld == nil {
		return []byte("null"), nil
	}
	v := tx.ld.Copy()
	if tx.input == nil {
		return nil, fmt.Errorf("TxDeleteData.MarshalJSON failed: invalid tx.input")
	}
	d, err := json.Marshal(tx.input)
	if err != nil {
		return nil, err
	}
	v.Data = d
	return json.Marshal(v)
}

func (tx *TxDeleteData) SyntacticVerify() error {
	var err error
	if err = tx.TxBase.SyntacticVerify(); err != nil {
		return err
	}

	switch {
	case tx.ld.To != nil:
		return fmt.Errorf("TxDeleteData.SyntacticVerify failed: invalid to, should be nil")
	case tx.ld.Token != nil:
		return fmt.Errorf("TxDeleteData.SyntacticVerify failed: invalid token, should be nil")
	case tx.ld.Amount != nil:
		return fmt.Errorf("TxDeleteData.SyntacticVerify failed: invalid amount, should be nil")
	case len(tx.ld.Data) == 0:
		return fmt.Errorf("TxDeleteData.SyntacticVerify failed: invalid data")
	}

	tx.input = &ld.TxUpdater{}
	if err = tx.input.Unmarshal(tx.ld.Data); err != nil {
		return fmt.Errorf("TxDeleteData.SyntacticVerify failed: %v", err)
	}
	if err = tx.input.SyntacticVerify(); err != nil {
		return fmt.Errorf("TxDeleteData.SyntacticVerify failed: %v", err)
	}

	switch {
	case tx.input.ID == nil || *tx.input.ID == util.DataIDEmpty:
		return fmt.Errorf("TxDeleteData.SyntacticVerify failed: invalid data id")
	case tx.input.Version == 0:
		return fmt.Errorf("TxDeleteData.SyntacticVerify failed: invalid data version")
	}
	return nil
}

func (tx *TxDeleteData) Verify(bctx BlockContext, bs BlockState) error {
	var err error
	if err = tx.TxBase.Verify(bctx, bs); err != nil {
		return err
	}

	tx.dm, err = bs.LoadData(*tx.input.ID)
	switch {
	case err != nil:
		return fmt.Errorf("TxDeleteData.Verify failed: %v", err)
	case tx.dm.Version != tx.input.Version:
		return fmt.Errorf("TxDeleteData.Verify failed: invalid version, expected %d, got %d",
			tx.dm.Version, tx.input.Version)
	case !util.SatisfySigning(tx.dm.Threshold, tx.dm.Keepers, tx.signers, false):
		return fmt.Errorf("TxDeleteData.Verify failed: invalid signatures for data keepers")
	case tx.ld.NeedApprove(tx.dm.Approver, tx.dm.ApproveList) && !tx.signers.Has(*tx.dm.Approver):
		return fmt.Errorf("TxDeleteData.Verify failed: invalid signature for data approver")
	}
	return nil
}

func (tx *TxDeleteData) Accept(bctx BlockContext, bs BlockState) error {
	var err error

	if err = bs.DeleteData(*tx.input.ID, tx.dm, tx.input.Data); err != nil {
		return err
	}
	return tx.TxBase.Accept(bctx, bs)
}
