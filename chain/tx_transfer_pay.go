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

type TxTransferPay struct {
	*TxBase
	exSigners []ids.ShortID
	data      *ld.TxTransfer
}

func (tx *TxTransferPay) MarshalJSON() ([]byte, error) {
	if tx == nil || tx.ld == nil {
		return util.Null, nil
	}
	v := tx.ld.Copy()
	if tx.data == nil {
		tx.data = &ld.TxTransfer{}
		if err := tx.data.Unmarshal(tx.ld.Data); err != nil {
			return nil, fmt.Errorf("TxTransferPay unmarshal data failed: %v", err)
		}
	}
	d, err := tx.data.MarshalJSON()
	if err != nil {
		return nil, err
	}
	v.Data = d
	return v.MarshalJSON()
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
	if !util.ShortIDs(tx.exSigners).Has(tx.ld.To) {
		return fmt.Errorf("TxTransferPay invalid exSignatures, not from recipient")
	}

	tx.data = &ld.TxTransfer{}
	if err = tx.data.Unmarshal(tx.ld.Data); err != nil {
		return fmt.Errorf("TxTransferPay unmarshal data failed: %v", err)
	}
	if err = tx.data.SyntacticVerify(); err != nil {
		return fmt.Errorf("TxTransferPay SyntacticVerify failed: %v", err)
	}
	if tx.data.To != tx.ld.To {
		return fmt.Errorf("TxTransferPay invalid recipient")
	}
	if tx.data.Expire > 0 && tx.data.Expire < uint64(time.Now().Unix()) {
		return fmt.Errorf("TxTransferPay expired")
	}
	if tx.data.Amount != nil && tx.data.Amount.Cmp(tx.ld.Amount) != 0 {
		return fmt.Errorf("TxTransferPay invalid amount")
	}
	return nil
}
