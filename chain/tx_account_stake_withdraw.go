// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package chain

import (
	"encoding/json"
	"fmt"

	"github.com/ldclabs/ldvm/ld"
	"github.com/ldclabs/ldvm/util"
)

type TxWithdrawStake struct {
	TxBase
	data *ld.TxTransfer
}

func (tx *TxWithdrawStake) MarshalJSON() ([]byte, error) {
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

func (tx *TxWithdrawStake) SyntacticVerify() error {
	var err error
	if err = tx.TxBase.SyntacticVerify(); err != nil {
		return err
	}

	if tx.ld.To == nil {
		return fmt.Errorf("TxWithdrawStake invalid to")
	}

	if tx.ld.Amount.Sign() != 0 {
		return fmt.Errorf("TxWithdrawStake invalid amount")
	}
	if token := util.StakeSymbol(*tx.ld.To); !token.Valid() {
		return fmt.Errorf("TxWithdrawStake invalid stake address: %s", token.GoString())
	}
	if len(tx.ld.Data) == 0 {
		return fmt.Errorf("TxWithdrawStake invalid")
	}
	tx.data = &ld.TxTransfer{}
	if err = tx.data.Unmarshal(tx.ld.Data); err != nil {
		return fmt.Errorf("TxWithdrawStake unmarshal data failed: %v", err)
	}
	if err = tx.data.SyntacticVerify(); err != nil {
		return fmt.Errorf("TxWithdrawStake SyntacticVerify failed: %v", err)
	}

	if tx.data.Amount == nil || tx.data.Amount.Sign() <= 0 {
		return fmt.Errorf("TxWithdrawStake invalid amount")
	}
	return nil
}

func (tx *TxWithdrawStake) Verify(blk *Block, bs BlockState) error {
	var err error
	if err = tx.TxBase.Verify(blk, bs); err != nil {
		return err
	}
	_, err = tx.to.CheckWithdrawStake(tx.token, tx.ld.From, tx.signers, tx.data.Amount)
	return err
}

func (tx *TxWithdrawStake) Accept(blk *Block, bs BlockState) error {
	withdraw, err := tx.to.WithdrawStake(tx.token, tx.ld.From, tx.signers, tx.data.Amount)
	if err != nil {
		return err
	}
	if err = tx.to.Sub(tx.token, withdraw); err != nil {
		return err
	}
	if err = tx.from.Add(tx.token, withdraw); err != nil {
		return err
	}
	return tx.TxBase.Accept(blk, bs)
}
