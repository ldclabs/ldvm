// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package chain

import (
	"fmt"
	"math/big"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ldclabs/ldvm/constants"
	"github.com/ldclabs/ldvm/ld"
	"github.com/ldclabs/ldvm/util"
)

type TxTransferExchange struct {
	*TxBase
	exSigners []ids.ShortID
	data      *ld.TxExchanger
	quantity  *big.Int
}

func (tx *TxTransferExchange) MarshalJSON() ([]byte, error) {
	if tx == nil || tx.ld == nil {
		return util.Null, nil
	}
	v := tx.ld.Copy()
	if tx.data == nil {
		tx.data = &ld.TxExchanger{}
		if err := tx.data.Unmarshal(tx.ld.Data); err != nil {
			return nil, fmt.Errorf("TxTransferExchange unmarshal data failed: %v", err)
		}
	}
	d, err := tx.data.MarshalJSON()
	if err != nil {
		return nil, err
	}
	v.Data = d
	return v.MarshalJSON()
}

func (tx *TxTransferExchange) SyntacticVerify() error {
	var err error
	if err = tx.TxBase.SyntacticVerify(); err != nil {
		return err
	}

	if len(tx.ld.Data) == 0 {
		return fmt.Errorf("TxTransferExchange invalid")
	}

	tx.exSigners, err = util.DeriveSigners(tx.ld.Data, tx.ld.ExSignatures)
	if err != nil {
		return fmt.Errorf("TxTransferExchange invalid exSignatures: %v", err)
	}

	tx.data = &ld.TxExchanger{}
	if err = tx.data.Unmarshal(tx.ld.Data); err != nil {
		return fmt.Errorf("TxTransferExchange unmarshal data failed: %v", err)
	}
	if err = tx.data.SyntacticVerify(); err != nil {
		return fmt.Errorf("TxTransferExchange SyntacticVerify failed: %v", err)
	}
	if tx.data.Nonce == 0 {
		return fmt.Errorf("TxTransferExchange invalid nonce")
	}
	if tx.data.Seller != tx.ld.To {
		return fmt.Errorf("TxTransferExchange invalid to")
	}
	if tx.data.To != ids.ShortEmpty && tx.data.To != tx.ld.From {
		return fmt.Errorf("TxTransferExchange invalid from")
	}
	if tx.data.Receive != tx.ld.Token {
		return fmt.Errorf("TxTransferExchange invalid token")
	}
	if tx.ld.Amount == nil || tx.ld.Amount.Sign() < 1 {
		return fmt.Errorf("TxTransferExchange invalid amount")
	}
	// quantity = amount * 1_000_000_000 / price
	tx.quantity = new(big.Int).SetUint64(constants.LDC)
	tx.quantity.Mul(tx.quantity, tx.ld.Amount)
	tx.quantity.Quo(tx.quantity, tx.data.Price)
	if tx.quantity.Cmp(tx.data.Minimum) < 0 || tx.quantity.Cmp(tx.data.Quota) > 0 {
		return fmt.Errorf("TxTransferExchange invalid amount")
	}
	return nil
}

func (tx *TxTransferExchange) Verify(blk *Block) error {
	var err error
	if err = tx.TxBase.Verify(blk); err != nil {
		return err
	}

	if tx.to.Nonce() != tx.data.Nonce {
		return fmt.Errorf("TxTransferExchange invalid seller nonce")
	}
	// verify seller's signatures
	if !tx.to.SatisfySigning(tx.exSigners) {
		return fmt.Errorf("TxTransferExchange account seller need more signers")
	}
	tokenB := tx.to.BalanceOf(tx.data.Sell)
	if tx.quantity.Cmp(tokenB) > 0 {
		return fmt.Errorf("TxTransferExchange seller %s insufficient balance, expected %v, got %v",
			tx.data.Seller, tx.quantity, tokenB)
	}
	return err
}

func (tx *TxTransferExchange) Accept(blk *Block) error {
	var err error
	if err = tx.to.SubByNonce(tx.data.Sell, tx.data.Nonce, tx.quantity); err != nil {
		return err
	}
	if err = tx.from.Add(tx.data.Sell, tx.quantity); err != nil {
		return err
	}
	return tx.TxBase.Accept(blk)
}
