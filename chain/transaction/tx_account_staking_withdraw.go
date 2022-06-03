// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package transaction

import (
	"encoding/json"
	"fmt"

	"github.com/ldclabs/ldvm/ld"
	"github.com/ldclabs/ldvm/util"
)

type TxWithdrawStake struct {
	TxBase
	input *ld.TxTransfer
}

func (tx *TxWithdrawStake) MarshalJSON() ([]byte, error) {
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
	tx.input = &ld.TxTransfer{}
	if err = tx.input.Unmarshal(tx.ld.Data); err != nil {
		return fmt.Errorf("TxWithdrawStake unmarshal data failed: %v", err)
	}
	if err = tx.input.SyntacticVerify(); err != nil {
		return fmt.Errorf("TxWithdrawStake SyntacticVerify failed: %v", err)
	}

	if tx.input.Amount == nil || tx.input.Amount.Sign() <= 0 {
		return fmt.Errorf("TxWithdrawStake invalid amount")
	}
	return nil
}

func (tx *TxWithdrawStake) Verify(bctx BlockContext, bs BlockState) error {
	var err error
	if err = tx.TxBase.Verify(bctx, bs); err != nil {
		return err
	}
	return tx.to.CheckWithdrawStake(tx.token, tx.ld.From, tx.signers, tx.input.Amount)
}

func (tx *TxWithdrawStake) Accept(bctx BlockContext, bs BlockState) error {
	// must WithdrawStake and then Accept
	withdraw, err := tx.to.WithdrawStake(tx.token, tx.ld.From, tx.signers, tx.input.Amount)
	if err != nil {
		return err
	}
	if err = tx.to.Sub(tx.token, withdraw); err != nil {
		return err
	}
	if err = tx.from.Add(tx.token, withdraw); err != nil {
		return err
	}
	return tx.TxBase.Accept(bctx, bs)
}
