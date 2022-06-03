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
		return nil, fmt.Errorf("MarshalJSON failed: data not exists")
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

	if tx.ld.Token != nil {
		return fmt.Errorf("invalid token, expected NativeToken, got %s",
			strconv.Quote(tx.ld.Token.GoString()))
	}
	if len(tx.ld.Data) == 0 {
		return fmt.Errorf("TxDeleteData invalid")
	}
	tx.input = &ld.TxUpdater{}
	if err = tx.input.Unmarshal(tx.ld.Data); err != nil {
		return fmt.Errorf("TxDeleteData unmarshal data failed: %v", err)
	}
	if err = tx.input.SyntacticVerify(); err != nil {
		return fmt.Errorf("TxDeleteData SyntacticVerify failed: %v", err)
	}
	if tx.input.ID == nil ||
		tx.input.Version == 0 {
		return fmt.Errorf("TxDeleteData invalid TxUpdater")
	}
	return nil
}

func (tx *TxDeleteData) Verify(bctx BlockContext, bs BlockState) error {
	var err error
	if err = tx.TxBase.Verify(bctx, bs); err != nil {
		return err
	}

	tx.dm, err = bs.LoadData(*tx.input.ID)
	if err != nil {
		return fmt.Errorf("TxDeleteData load data failed: %v", err)
	}
	if tx.dm.Version != tx.input.Version {
		return fmt.Errorf("TxDeleteData version mismatch, expected %v, got %v",
			tx.dm.Version, tx.input.Version)
	}
	if !util.SatisfySigning(tx.dm.Threshold, tx.dm.Keepers, tx.signers, false) {
		return fmt.Errorf("TxDeleteData need more signatures")
	}
	if tx.ld.NeedApprove(tx.dm.Approver, tx.dm.ApproveList) && !tx.signers.Has(*tx.dm.Approver) {
		return fmt.Errorf("TxDeleteData.Verify failed: no approver signing")
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
