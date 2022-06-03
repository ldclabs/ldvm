// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package transaction

import (
	"encoding/json"
	"fmt"

	"github.com/ldclabs/ldvm/constants"
	"github.com/ldclabs/ldvm/ld"
	"github.com/ldclabs/ldvm/util"
)

type TxTransferCash struct {
	TxBase
	issuer    *Account
	exSigners util.EthIDs
	input     *ld.TxTransfer
}

func (tx *TxTransferCash) MarshalJSON() ([]byte, error) {
	if tx == nil || tx.ld == nil {
		return []byte("null"), nil
	}
	v := tx.ld.Copy()
	if tx.input == nil {
		return nil, fmt.Errorf("TxTransferCash.MarshalJSON failed: invalid tx.input")
	}
	d, err := json.Marshal(tx.input)
	if err != nil {
		return nil, err
	}
	v.Data = d
	return json.Marshal(v)
}

func (tx *TxTransferCash) SyntacticVerify() error {
	var err error
	if err = tx.TxBase.SyntacticVerify(); err != nil {
		return err
	}
	switch {
	case tx.ld.To == nil:
		return fmt.Errorf("TxTransferCash.SyntacticVerify failed: invalid to")
	case tx.ld.Amount != nil:
		return fmt.Errorf("TxTransferCash.SyntacticVerify failed: invalid amount, should be nil")
	case len(tx.ld.Data) == 0:
		return fmt.Errorf("TxTransferCash.SyntacticVerify failed: invalid data")
	}

	tx.input = &ld.TxTransfer{}
	if err = tx.input.Unmarshal(tx.ld.Data); err != nil {
		return fmt.Errorf("TxTransferCash.SyntacticVerify failed: %v", err)
	}
	if err = tx.input.SyntacticVerify(); err != nil {
		return fmt.Errorf("TxTransferCash.SyntacticVerify failed: %v", err)
	}

	switch {
	case tx.input.From == nil:
		return fmt.Errorf("TxTransferCash.SyntacticVerify failed: nil issuer")
	case *tx.input.From != *tx.ld.To:
		return fmt.Errorf("TxTransferCash.SyntacticVerify failed: invalid issuer, expected %s, got %s",
			tx.input.From, tx.ld.To)
	case tx.input.To == nil:
		return fmt.Errorf("TxTransferCash.SyntacticVerify failed: nil recipient")
	case *tx.input.To != tx.ld.From:
		return fmt.Errorf("TxTransferCash.SyntacticVerify failed: invalid recipient, expected %s, got %s",
			tx.input.To, tx.ld.From)
	case tx.input.Token == nil && tx.token != constants.NativeToken:
		return fmt.Errorf("TxTransferCash.SyntacticVerify failed: invalid token, expected %s, got %s",
			constants.NativeToken.GoString(), tx.token.GoString())
	case tx.input.Token != nil && tx.token != *tx.input.Token:
		return fmt.Errorf("TxTransferCash.SyntacticVerify failed: invalid token, expected %s, got %s",
			tx.input.Token.GoString(), tx.token.GoString())
	case tx.input.Amount == nil:
		return fmt.Errorf("TxTransferCash.SyntacticVerify failed: nil amount")
	case tx.input.Expire < tx.ld.Timestamp:
		return fmt.Errorf("TxTransferCash.SyntacticVerify failed: data expired")
	}

	tx.exSigners, err = tx.ld.ExSigners()
	if err != nil {
		return fmt.Errorf("TxTransferPay.SyntacticVerify failed: %v", err)
	}
	return nil
}

func (tx *TxTransferCash) Verify(bctx BlockContext, bs BlockState) error {
	var err error
	if err = tx.TxBase.Verify(bctx, bs); err != nil {
		return err
	}
	// verify issuer's signatures
	if !tx.to.SatisfySigning(tx.exSigners) {
		return fmt.Errorf("TxTransferPay.Verify failed: invalid signature for issuer")
	}

	if err = tx.to.CheckSubByNonceTable(
		tx.token, tx.input.Expire, tx.input.Nonce, tx.input.Amount); err != nil {
		return err
	}
	return err
}

func (tx *TxTransferCash) Accept(bctx BlockContext, bs BlockState) error {
	var err error
	if err = tx.to.SubByNonceTable(
		tx.token, tx.input.Expire, tx.input.Nonce, tx.input.Amount); err != nil {
		return err
	}
	if err = tx.from.Add(tx.token, tx.input.Amount); err != nil {
		return err
	}
	return tx.TxBase.Accept(bctx, bs)
}
