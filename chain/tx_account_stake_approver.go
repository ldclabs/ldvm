// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package chain

import (
	"encoding/json"
	"fmt"

	"github.com/ldclabs/ldvm/ld"
	"github.com/ldclabs/ldvm/util"
)

type TxUpdateStakeApprover struct {
	TxBase
	data *ld.TxAccounter
}

func (tx *TxUpdateStakeApprover) MarshalJSON() ([]byte, error) {
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

func (tx *TxUpdateStakeApprover) SyntacticVerify() error {
	var err error
	if err = tx.TxBase.SyntacticVerify(); err != nil {
		return err
	}
	if tx.ld.To == nil {
		return fmt.Errorf("TxUpdateStakeApprover invalid to")
	}

	if token := util.StakeSymbol(*tx.ld.To); !token.Valid() {
		return fmt.Errorf("TxUpdateStakeApprover invalid stake address: %s", token.GoString())
	}
	if tx.ld.Amount.Sign() != 0 {
		return fmt.Errorf("TxUpdateStakeApprover invalid amount")
	}
	if len(tx.ld.Data) == 0 {
		return fmt.Errorf("TxUpdateStakeApprover invalid")
	}
	tx.data = &ld.TxAccounter{}
	if err = tx.data.Unmarshal(tx.ld.Data); err != nil {
		return fmt.Errorf("TxUpdateStakeApprover unmarshal data failed: %v", err)
	}
	if err = tx.data.SyntacticVerify(); err != nil {
		return fmt.Errorf("TxUpdateStakeApprover SyntacticVerify failed: %v", err)
	}
	if tx.data.Approver == nil {
		return fmt.Errorf("TxUpdateStakeApprover invalid approver")
	}
	return nil
}

func (tx *TxUpdateStakeApprover) Verify(blk *Block, bs BlockState) error {
	var err error
	if err = tx.TxBase.Verify(blk, bs); err != nil {
		return err
	}
	return tx.to.CheckUpdateApprover(tx.ld.From, *tx.data.Approver, tx.signers)
}

func (tx *TxUpdateStakeApprover) Accept(blk *Block, bs BlockState) error {
	err := tx.to.UpdateApprover(tx.ld.From, *tx.data.Approver, tx.signers)
	if err != nil {
		return err
	}
	return tx.TxBase.Accept(blk, bs)
}
