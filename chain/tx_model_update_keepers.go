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

type TxUpdateModelKeepers struct {
	TxBase
	data *ld.TxUpdater
	mm   *ld.ModelMeta
}

func (tx *TxUpdateModelKeepers) MarshalJSON() ([]byte, error) {
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

func (tx *TxUpdateModelKeepers) SyntacticVerify() error {
	var err error
	if err = tx.TxBase.SyntacticVerify(); err != nil {
		return err
	}

	if tx.ld.Token != nil {
		return fmt.Errorf("invalid token, expected NativeToken, got %s",
			strconv.Quote(tx.ld.Token.GoString()))
	}
	if len(tx.ld.Data) == 0 {
		return fmt.Errorf("TxUpdateModelKeepers invalid")
	}

	tx.data = &ld.TxUpdater{}
	if err = tx.data.Unmarshal(tx.ld.Data); err != nil {
		return fmt.Errorf("TxUpdateModelKeepers unmarshal data failed: %v", err)
	}
	if err = tx.data.SyntacticVerify(); err != nil {
		return fmt.Errorf("TxUpdateModelKeepers SyntacticVerify failed: %v", err)
	}
	if tx.data.ModelID == nil {
		return fmt.Errorf("TxUpdateModelKeepers invalid TxUpdater")
	}
	if len(tx.data.Keepers) == 0 && tx.data.Approver == nil {
		return fmt.Errorf("TxUpdateModelKeepers no keepers nor approver")
	}
	return nil
}

func (tx *TxUpdateModelKeepers) Verify(blk *Block, bs BlockState) error {
	var err error
	if err = tx.TxBase.Verify(blk, bs); err != nil {
		return err
	}

	tx.mm, err = bs.LoadModel(*tx.data.ModelID)
	if err != nil {
		return fmt.Errorf("TxUpdateModelKeepers load model failed: %v", err)
	}

	if !util.SatisfySigningPlus(tx.mm.Threshold, tx.mm.Keepers, tx.signers) {
		return fmt.Errorf("TxUpdateModelKeepers need more signatures")
	}

	if tx.ld.NeedApprove(tx.mm.Approver, nil) && !tx.signers.Has(*tx.mm.Approver) {
		return fmt.Errorf("TxUpdateModelKeepers.Verify failed: no approver signing")
	}
	return nil
}

func (tx *TxUpdateModelKeepers) Accept(blk *Block, bs BlockState) error {
	var err error

	if tx.data.Approver != nil {
		if *tx.data.Approver == util.EthIDEmpty {
			tx.mm.Approver = nil
		} else {
			tx.mm.Approver = tx.data.Approver
		}
	}
	if len(tx.data.Keepers) > 0 {
		tx.mm.Threshold = tx.data.Threshold
		tx.mm.Keepers = tx.data.Keepers
	}
	if err = bs.SaveModel(*tx.data.ModelID, tx.mm); err != nil {
		return err
	}
	return tx.TxBase.Accept(blk, bs)
}
