// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package transaction

import (
	"encoding/json"
	"fmt"

	"github.com/ldclabs/ldvm/ld"
	"github.com/ldclabs/ldvm/util"
)

type TxTransferCash struct {
	TxBase
	issuer    *Account
	exSigners util.EthIDs
	data      *ld.TxTransfer
}

func (tx *TxTransferCash) MarshalJSON() ([]byte, error) {
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

func (tx *TxTransferCash) SyntacticVerify() error {
	var err error
	if err = tx.TxBase.SyntacticVerify(); err != nil {
		return err
	}
	if tx.ld.To == nil {
		return fmt.Errorf("TxTransferCash invalid to")
	}

	if tx.ld.Amount != nil {
		return fmt.Errorf("TxTransferCash invalid amount")
	}

	if len(tx.ld.Data) == 0 {
		return fmt.Errorf("TxTransferCash invalid")
	}

	tx.exSigners, err = tx.ld.ExSigners()
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
	if tx.data.Token != nil && *tx.data.Token != tx.token {
		return fmt.Errorf("TxTransferCash invalid token")
	}
	if tx.data.To == nil || *tx.data.To != tx.ld.From {
		return fmt.Errorf("TxTransferCash invalid recipient")
	}
	if tx.data.From == nil || *tx.data.From != *tx.ld.To {
		return fmt.Errorf("TxTransferCash invalid issuer")
	}

	if tx.data.Expire < tx.ld.Timestamp {
		return fmt.Errorf("TxTransferCash expired")
	}

	if tx.data.Amount == nil {
		return fmt.Errorf("TxTransferCash invalid data amount")
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
		return fmt.Errorf("TxTransferCash account issuer need more signers")
	}

	if err = tx.to.CheckSubByNonceTable(tx.token, tx.data.Expire, tx.data.Nonce, tx.data.Amount); err != nil {
		return err
	}
	return err
}

func (tx *TxTransferCash) Accept(bctx BlockContext, bs BlockState) error {
	var err error
	if err = tx.to.SubByNonceTable(tx.token, tx.data.Expire, tx.data.Nonce, tx.data.Amount); err != nil {
		return err
	}
	if err = tx.from.Add(tx.token, tx.data.Amount); err != nil {
		return err
	}
	return tx.TxBase.Accept(bctx, bs)
}
