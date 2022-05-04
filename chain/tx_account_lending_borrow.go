// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package chain

import (
	"encoding/json"
	"fmt"

	"github.com/ldclabs/ldvm/constants"
	"github.com/ldclabs/ldvm/ld"
	"github.com/ldclabs/ldvm/util"
)

type TxBorrow struct {
	TxBase
	exSigners []util.EthID
	data      *ld.TxTransfer
	dueTime   uint64
}

func (tx *TxBorrow) MarshalJSON() ([]byte, error) {
	if tx == nil || tx.ld == nil {
		return util.Null, nil
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

func (tx *TxBorrow) SyntacticVerify() error {
	var err error
	if err = tx.TxBase.SyntacticVerify(); err != nil {
		return err
	}

	if tx.ld.Token != constants.NativeToken {
		return fmt.Errorf("invalid token %s, required LDC", tx.ld.Token)
	}
	if tx.ld.To == util.EthIDEmpty {
		return fmt.Errorf("TxBorrow invalid to: %s", tx.ld.To)
	}
	if tx.ld.Amount.Sign() != 0 {
		return fmt.Errorf("TxBorrow invalid amount, expected 0, got %v", tx.ld.Amount)
	}

	if len(tx.ld.Data) == 0 {
		return fmt.Errorf("TxBorrow invalid")
	}
	tx.exSigners, err = util.DeriveSigners(tx.ld.Data, tx.ld.ExSignatures)
	if err != nil {
		return fmt.Errorf("TxBorrow invalid exSignatures")
	}

	tx.data = &ld.TxTransfer{}
	if err = tx.data.Unmarshal(tx.ld.Data); err != nil {
		return fmt.Errorf("TxBorrow unmarshal data failed: %v", err)
	}
	if err = tx.data.SyntacticVerify(); err != nil {
		return fmt.Errorf("TxBorrow SyntacticVerify failed: %v", err)
	}

	if tx.data.From != tx.ld.To {
		return fmt.Errorf("TxBorrow invalid lender")
	}
	if tx.data.To != tx.ld.From {
		return fmt.Errorf("TxBorrow invalid recipient")
	}
	if tx.data.Expire < tx.ld.Timestamp {
		return fmt.Errorf("TxBorrow expired")
	}
	if tx.data.Amount.Sign() <= 0 {
		return fmt.Errorf("TxBorrow invalid amount")
	}

	if len(tx.data.Data) > 0 {
		u := uint64(0)
		if err = ld.DecMode.Unmarshal(tx.ld.Data, &u); err != nil {
			return fmt.Errorf("TxBorrow unmarshal dueTime failed: %v", err)
		}
		tx.dueTime = u
	}
	return nil
}

func (tx *TxBorrow) Verify(blk *Block, bs BlockState) error {
	var err error
	if err = tx.TxBase.Verify(blk, bs); err != nil {
		return err
	}
	if err = tx.to.CheckSubByNonceTable(tx.data.Token, tx.data.Expire, tx.data.Nonce, tx.data.Amount); err != nil {
		return err
	}
	// verify lender's signatures
	if !tx.to.SatisfySigning(tx.exSigners) {
		return fmt.Errorf("TxBorrow account lender need more signers")
	}
	return tx.to.CheckBorrow(tx.data.Token, tx.ld.From, tx.data.Amount, tx.dueTime)
}

func (tx *TxBorrow) Accept(blk *Block, bs BlockState) error {
	var err error
	if err = tx.to.Borrow(tx.data.Token, tx.ld.From, tx.data.Amount, tx.dueTime); err != nil {
		return err
	}
	if err = tx.to.SubByNonceTable(tx.data.Token, tx.data.Expire, tx.data.Nonce, tx.data.Amount); err != nil {
		return err
	}
	if err = tx.from.Add(tx.data.Token, tx.data.Amount); err != nil {
		return err
	}
	return tx.TxBase.Accept(blk, bs)
}
