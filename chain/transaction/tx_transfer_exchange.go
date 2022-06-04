// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package transaction

import (
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/ldclabs/ldvm/constants"
	"github.com/ldclabs/ldvm/ld"
	"github.com/ldclabs/ldvm/util"
)

type TxTransferExchange struct {
	TxBase
	exSigners util.EthIDs
	input     *ld.TxExchanger
	quantity  *big.Int
}

func (tx *TxTransferExchange) MarshalJSON() ([]byte, error) {
	if tx == nil || tx.ld == nil {
		return []byte("null"), nil
	}
	v := tx.ld.Copy()
	if tx.input == nil {
		return nil, fmt.Errorf("TxTransferExchange.MarshalJSON failed: invalid tx.input")
	}
	d, err := json.Marshal(tx.input)
	if err != nil {
		return nil, err
	}
	v.Data = d
	return json.Marshal(v)
}

func (tx *TxTransferExchange) SyntacticVerify() error {
	var err error
	if err = tx.TxBase.SyntacticVerify(); err != nil {
		return err
	}

	switch {
	case tx.ld.To == nil:
		return fmt.Errorf("TxTransferExchange.SyntacticVerify failed: invalid to")
	case tx.ld.Amount == nil:
		return fmt.Errorf("TxTransferExchange.SyntacticVerify failed: invalid amount")
	case len(tx.ld.Data) == 0:
		return fmt.Errorf("TxTransferExchange.SyntacticVerify failed: invalid data")
	}

	tx.input = &ld.TxExchanger{}
	if err = tx.input.Unmarshal(tx.ld.Data); err != nil {
		return fmt.Errorf("TxTransferExchange.SyntacticVerify failed: %v", err)
	}
	if err = tx.input.SyntacticVerify(); err != nil {
		return fmt.Errorf("TxTransferExchange.SyntacticVerify failed: %v", err)
	}

	// quantity = amount * 1_000_000_000 / price
	tx.quantity = new(big.Int).SetUint64(constants.LDC)
	tx.quantity.Mul(tx.quantity, tx.ld.Amount)
	tx.quantity.Quo(tx.quantity, tx.input.Price)

	switch {
	case tx.quantity.Cmp(tx.input.Minimum) < 0:
		min := new(big.Int).Mul(tx.input.Minimum, tx.input.Price)
		min.Quo(min, new(big.Int).SetUint64(constants.LDC))
		return fmt.Errorf(
			"TxTransferExchange.SyntacticVerify failed: invalid amount, expected >=%v, got %v",
			min, tx.ld.Amount)
	case tx.quantity.Cmp(tx.input.Quota) > 0:
		max := new(big.Int).Mul(tx.input.Quota, tx.input.Price)
		max.Quo(max, new(big.Int).SetUint64(constants.LDC))
		return fmt.Errorf(
			"TxTransferExchange.SyntacticVerify failed: invalid amount, expected <=%v, got %v",
			max, tx.ld.Amount)
	case tx.input.Purchaser != nil && *tx.input.Purchaser != tx.ld.From:
		return fmt.Errorf(
			"TxTransferExchange.SyntacticVerify failed: invalid from, expected %s, got %s",
			*tx.input.Purchaser, tx.ld.From)
	case tx.input.Payee != *tx.ld.To:
		return fmt.Errorf(
			"TxTransferExchange.SyntacticVerify failed: invalid to, expected %s, got %s",
			tx.input.Payee, tx.ld.To)
	case tx.input.Receive != tx.token:
		return fmt.Errorf(
			"TxTransferExchange.SyntacticVerify failed: invalid token, expected %s, got %s",
			tx.input.Receive.GoString(), tx.token.GoString())
	case tx.input.Expire < tx.ld.Timestamp:
		return fmt.Errorf("TxTransferExchange.SyntacticVerify failed: data expired")
	}

	tx.exSigners, err = tx.ld.ExSigners()
	if err != nil {
		return fmt.Errorf("TxTransferExchange.SyntacticVerify failed: invalid exSignatures: %v", err)
	}
	return nil
}

func (tx *TxTransferExchange) Verify(bctx BlockContext, bs BlockState) error {
	var err error
	if err = tx.TxBase.Verify(bctx, bs); err != nil {
		return err
	}
	// verify seller's signatures
	if !tx.to.SatisfySigning(tx.exSigners) {
		return fmt.Errorf("TxTransferExchange.Verify failed: invalid signatures for seller")
	}
	if err = tx.to.CheckSubByNonceTable(
		tx.input.Sell, tx.input.Expire, tx.input.Nonce, tx.quantity); err != nil {
		return err
	}
	return err
}

func (tx *TxTransferExchange) Accept(bctx BlockContext, bs BlockState) error {
	var err error
	if err = tx.to.SubByNonceTable(tx.input.Sell, tx.input.Expire, tx.input.Nonce, tx.quantity); err != nil {
		return err
	}
	if err = tx.from.Add(tx.input.Sell, tx.quantity); err != nil {
		return err
	}
	return tx.TxBase.Accept(bctx, bs)
}
