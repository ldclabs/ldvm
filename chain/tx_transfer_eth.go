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

type TxEth struct {
	TxBase
	data *ld.TxEth
}

func (tx *TxEth) MarshalJSON() ([]byte, error) {
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

func (tx *TxEth) SyntacticVerify() error {
	if tx == nil || tx.ld == nil {
		return fmt.Errorf("TxEth is nil")
	}

	tx.data = &ld.TxEth{}
	if err := tx.data.Unmarshal(tx.ld.Data); err != nil {
		return fmt.Errorf("TxEth unmarshal data failed: %v", err)
	}
	if len(tx.ld.Signatures) != 1 {
		return fmt.Errorf("TxEth invalid signature")
	}
	copy(tx.data.Signature[:], tx.ld.Signatures[0][:])
	if err := tx.data.SyntacticVerify(); err != nil {
		return err
	}
	if tx.ld.Token != constants.NativeToken {
		return fmt.Errorf("invalid token %s, required LDC", tx.ld.Token)
	}
	if tx.ld.ChainID != tx.data.ChainID {
		return fmt.Errorf("TxEth invalid chainID")
	}
	if tx.ld.Nonce != tx.data.Nonce {
		return fmt.Errorf("TxEth invalid nonce")
	}
	if tx.ld.GasTip != tx.data.GasTipCap {
		return fmt.Errorf("TxEth invalid gasTipCap")
	}
	if tx.ld.GasFeeCap != tx.data.GasFeeCap {
		return fmt.Errorf("TxEth invalid gasFeeCap")
	}
	if tx.ld.From != tx.data.From {
		return fmt.Errorf("TxEth invalid from")
	}
	if tx.ld.To != tx.data.To {
		return fmt.Errorf("TxEth invalid to")
	}
	if tx.ld.Amount == nil || tx.data.Value == nil || tx.ld.Amount.Cmp(tx.data.Value) != 0 {
		return fmt.Errorf("TxEth invalid amount")
	}

	tx.signers = []util.EthID{tx.ld.From}
	return nil
}
