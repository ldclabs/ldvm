// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package chain

import (
	"fmt"
	"time"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ldclabs/ldvm/ld"
	"github.com/ldclabs/ldvm/util"
)

type TxTransferCash struct {
	*TxBase
	issuer    *Account
	exSigners []ids.ShortID
	data      *ld.TxTransfer
}

func (tx *TxTransferCash) MarshalJSON() ([]byte, error) {
	if tx == nil || tx.ld == nil {
		return util.Null, nil
	}
	v := tx.ld.Copy()
	if tx.data == nil {
		tx.data = &ld.TxTransfer{}
		if err := tx.data.Unmarshal(tx.ld.Data); err != nil {
			return nil, fmt.Errorf("TxTransferCash unmarshal data failed: %v", err)
		}
	}
	d, err := tx.data.MarshalJSON()
	if err != nil {
		return nil, err
	}
	v.Data = d
	return v.MarshalJSON()
}

func (tx *TxTransferCash) SyntacticVerify() error {
	var err error
	if err = tx.TxBase.SyntacticVerify(); err != nil {
		return err
	}

	if len(tx.ld.Data) == 0 {
		return fmt.Errorf("TxTransferCash invalid")
	}

	tx.exSigners, err = util.DeriveSigners(tx.ld.Data, tx.ld.ExSignatures)
	if err != nil {
		return fmt.Errorf("TxTransferCash invalid exSignatures: %v", err)
	}

	tx.data = &ld.TxTransfer{}
	if err = tx.data.Unmarshal(tx.ld.Data); err != nil {
		return fmt.Errorf("TxTransferCash unmarshal data failed: %v", err)
	}
	if err = tx.data.SyntacticVerify(); err != nil {
		return fmt.Errorf("TxTransferCash SyntacticVerify failed: %v", err)
	}
	if tx.data.Nonce == 0 {
		return fmt.Errorf("TxTransferCash invalid nonce")
	}
	if tx.data.Token != tx.ld.Token {
		return fmt.Errorf("TxTransferCash invalid token")
	}
	if tx.data.To != tx.ld.From {
		return fmt.Errorf("TxTransferCash invalid recipient")
	}
	if tx.data.To != tx.ld.To {
		return fmt.Errorf("TxTransferCash invalid recipient")
	}
	if tx.data.Expire > 0 && tx.data.Expire < uint64(time.Now().Unix()) {
		return fmt.Errorf("TxTransferCash expired")
	}

	// tx.ld.Amount can be less than tx.data.Amount
	if tx.data.Amount == nil || tx.ld.Amount == nil || tx.data.Amount.Cmp(tx.ld.Amount) < 0 {
		return fmt.Errorf("TxTransferCash invalid amount")
	}
	return nil
}

func (tx *TxTransferCash) Verify(blk *Block) error {
	var err error
	if err = tx.TxBase.Verify(blk); err != nil {
		return err
	}

	if tx.issuer, err = blk.State().LoadAccount(tx.data.From); err != nil {
		return err
	}
	if tx.issuer.Nonce() != tx.data.Nonce {
		return fmt.Errorf("TxTransferCash invalid issuer nonce")
	}
	// verify issuer's signatures
	if !tx.issuer.SatisfySigning(tx.exSigners) {
		return fmt.Errorf("TxTransferCash account issuer need more signers")
	}
	tokenB := tx.issuer.BalanceOf(tx.ld.Token)
	if tx.data.Amount.Cmp(tokenB) > 0 {
		return fmt.Errorf("TxTransferCash issuer %s insufficient balance, expected %v, got %v",
			tx.data.From, tx.data.Amount, tokenB)
	}
	return err
}

func (tx *TxTransferCash) Accept(blk *Block) error {
	var err error
	if err = tx.issuer.SubByNonce(tx.ld.Token, tx.data.Nonce, tx.ld.Amount); err != nil {
		return err
	}
	if err = tx.to.Add(tx.ld.Token, tx.ld.Amount); err != nil {
		return err
	}
	return tx.TxBase.Accept(blk)
}
