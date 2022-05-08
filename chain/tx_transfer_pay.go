// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package chain

import (
	"encoding/json"
	"fmt"

	"github.com/ldclabs/ldvm/ld"
	"github.com/ldclabs/ldvm/util"
)

type TxTransferPay struct {
	TxBase
	exSigners util.EthIDs
	data      *ld.TxTransfer
}

func (tx *TxTransferPay) MarshalJSON() ([]byte, error) {
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

func (tx *TxTransferPay) SyntacticVerify() error {
	var err error
	if err = tx.TxBase.SyntacticVerify(); err != nil {
		return err
	}

	if len(tx.ld.Data) == 0 {
		return fmt.Errorf("TxTransferPay invalid")
	}

	tx.exSigners, err = util.DeriveSigners(tx.ld.Data, tx.ld.ExSignatures)
	if err != nil {
		return fmt.Errorf("TxTransferPay invalid exSignatures: %v", err)
	}
	if !tx.exSigners.Has(tx.ld.To) {
		return fmt.Errorf("TxTransferPay invalid exSignatures, not from recipient")
	}

	tx.data = &ld.TxTransfer{}
	if err = tx.data.Unmarshal(tx.ld.Data); err != nil {
		return fmt.Errorf("TxTransferPay unmarshal data failed: %v", err)
	}
	if err = tx.data.SyntacticVerify(); err != nil {
		return fmt.Errorf("TxTransferPay SyntacticVerify failed: %v", err)
	}
	if tx.data.To == nil || *tx.data.To != tx.ld.To {
		return fmt.Errorf("TxTransferPay invalid recipient")
	}
	if tx.data.Expire < tx.ld.Timestamp {
		return fmt.Errorf("TxTransferPay expired")
	}
	if tx.data.Amount == nil || tx.data.Amount.Cmp(tx.ld.Amount) != 0 {
		return fmt.Errorf("TxTransferPay invalid amount")
	}
	return nil
}
