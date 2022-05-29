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
	data      *ld.TxExchanger
	quantity  *big.Int
}

func (tx *TxTransferExchange) MarshalJSON() ([]byte, error) {
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

func (tx *TxTransferExchange) SyntacticVerify() error {
	var err error
	if err = tx.TxBase.SyntacticVerify(); err != nil {
		return err
	}

	if tx.ld.To == nil {
		return fmt.Errorf("TxTransferExchange invalid to")
	}

	if len(tx.ld.Data) == 0 {
		return fmt.Errorf("TxTransferExchange invalid")
	}

	tx.exSigners, err = tx.ld.ExSigners()
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
	if tx.data.Payee != *tx.ld.To {
		return fmt.Errorf("TxTransferExchange invalid to")
	}
	if tx.data.Purchaser != nil && *tx.data.Purchaser != tx.ld.From {
		return fmt.Errorf("TxTransferExchange invalid from")
	}
	if tx.data.Receive != tx.token {
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

func (tx *TxTransferExchange) Verify(bctx BlockContext, bs BlockState) error {
	var err error
	if err = tx.TxBase.Verify(bctx, bs); err != nil {
		return err
	}
	// verify seller's signatures
	if !tx.to.SatisfySigning(tx.exSigners) {
		return fmt.Errorf("TxTransferExchange account seller need more signers")
	}
	if err = tx.to.CheckSubByNonceTable(tx.data.Sell, tx.data.Expire, tx.data.Nonce, tx.quantity); err != nil {
		return err
	}
	return err
}

func (tx *TxTransferExchange) Accept(bctx BlockContext, bs BlockState) error {
	var err error
	if err = tx.to.SubByNonceTable(tx.data.Sell, tx.data.Expire, tx.data.Nonce, tx.quantity); err != nil {
		return err
	}
	if err = tx.from.Add(tx.data.Sell, tx.quantity); err != nil {
		return err
	}
	return tx.TxBase.Accept(bctx, bs)
}
