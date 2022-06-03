// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package transaction

import (
	"encoding/json"
	"fmt"

	"github.com/ldclabs/ldvm/ld"
	"github.com/ldclabs/ldvm/util"
)

type TxBorrow struct {
	TxBase
	exSigners util.EthIDs
	input     *ld.TxTransfer
	dueTime   uint64
}

func (tx *TxBorrow) MarshalJSON() ([]byte, error) {
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

func (tx *TxBorrow) SyntacticVerify() error {
	var err error
	if err = tx.TxBase.SyntacticVerify(); err != nil {
		return err
	}
	if tx.ld.To == nil {
		return fmt.Errorf("TxBorrow invalid to")
	}
	if tx.ld.Amount.Sign() != 0 {
		return fmt.Errorf("TxBorrow invalid amount, expected 0, got %v", tx.ld.Amount)
	}

	if len(tx.ld.Data) == 0 {
		return fmt.Errorf("TxBorrow invalid")
	}
	tx.exSigners, err = tx.ld.ExSigners()
	if err != nil {
		return fmt.Errorf("TxBorrow invalid exSignatures")
	}

	tx.input = &ld.TxTransfer{}
	if err = tx.input.Unmarshal(tx.ld.Data); err != nil {
		return fmt.Errorf("TxBorrow unmarshal data failed: %v", err)
	}
	if err = tx.input.SyntacticVerify(); err != nil {
		return fmt.Errorf("TxBorrow SyntacticVerify failed: %v", err)
	}

	if tx.input.From == nil || *tx.input.From != *tx.ld.To {
		return fmt.Errorf("TxBorrow invalid lender")
	}
	if tx.input.To == nil || *tx.input.To != tx.ld.From {
		return fmt.Errorf("TxBorrow invalid recipient")
	}
	if tx.input.Token != nil && *tx.input.Token != tx.token {
		return fmt.Errorf("TxBorrow invalid token")
	}
	if tx.input.Expire < tx.ld.Timestamp {
		return fmt.Errorf("TxBorrow expired")
	}
	if tx.input.Amount == nil {
		return fmt.Errorf("TxBorrow invalid amount")
	}

	if len(tx.input.Data) > 0 {
		u := uint64(0)
		if err = ld.DecMode.Unmarshal(tx.input.Data, &u); err != nil {
			return fmt.Errorf("TxBorrow unmarshal dueTime failed: %v", err)
		}
		tx.dueTime = u
	}
	return nil
}

func (tx *TxBorrow) Verify(bctx BlockContext, bs BlockState) error {
	var err error
	if err = tx.TxBase.Verify(bctx, bs); err != nil {
		return err
	}
	if err = tx.to.CheckSubByNonceTable(tx.token, tx.input.Expire, tx.input.Nonce, tx.input.Amount); err != nil {
		return err
	}
	// verify lender's signatures
	if !tx.to.SatisfySigning(tx.exSigners) {
		return fmt.Errorf("TxBorrow account lender need more signers")
	}
	return tx.to.CheckBorrow(tx.token, tx.ld.From, tx.input.Amount, tx.dueTime)
}

func (tx *TxBorrow) Accept(bctx BlockContext, bs BlockState) error {
	var err error
	if err = tx.to.Borrow(tx.token, tx.ld.From, tx.input.Amount, tx.dueTime); err != nil {
		return err
	}
	if err = tx.to.SubByNonceTable(tx.token, tx.input.Expire, tx.input.Nonce, tx.input.Amount); err != nil {
		return err
	}
	if err = tx.from.Add(tx.token, tx.input.Amount); err != nil {
		return err
	}
	return tx.TxBase.Accept(bctx, bs)
}
