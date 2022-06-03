// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package transaction

import (
	"encoding/json"
	"fmt"

	"github.com/ldclabs/ldvm/ld"
	"github.com/ldclabs/ldvm/util"
)

type TxTakeStake struct {
	TxBase
	exSigners util.EthIDs
	input     *ld.TxTransfer
	lockTime  uint64
}

func (tx *TxTakeStake) MarshalJSON() ([]byte, error) {
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

func (tx *TxTakeStake) SyntacticVerify() error {
	var err error
	if err = tx.TxBase.SyntacticVerify(); err != nil {
		return err
	}

	if tx.ld.To == nil {
		return fmt.Errorf("TxTakeStake invalid to")
	}

	if token := util.StakeSymbol(*tx.ld.To); !token.Valid() {
		return fmt.Errorf("TxTakeStake invalid stake address: %s", token.GoString())
	}
	if len(tx.ld.Data) == 0 {
		return fmt.Errorf("TxTakeStake invalid")
	}
	tx.exSigners, err = tx.ld.ExSigners()
	if err != nil {
		return fmt.Errorf("TxTakeStake invalid exSignatures")
	}

	tx.input = &ld.TxTransfer{}
	if err = tx.input.Unmarshal(tx.ld.Data); err != nil {
		return fmt.Errorf("TxTakeStake unmarshal data failed: %v", err)
	}
	if err = tx.input.SyntacticVerify(); err != nil {
		return fmt.Errorf("TxTakeStake SyntacticVerify failed: %v", err)
	}
	if tx.input.Token == nil || *tx.input.Token != tx.token {
		return fmt.Errorf("TxTakeStake invalid token")
	}
	if tx.input.From == nil || *tx.input.From != tx.ld.From {
		return fmt.Errorf("TxTakeStake invalid sender")
	}
	if tx.input.To == nil || *tx.input.To != *tx.ld.To {
		return fmt.Errorf("TxTakeStake invalid recipient")
	}
	if tx.input.Expire < tx.ld.Timestamp {
		return fmt.Errorf("TxTakeStake expired")
	}
	if tx.input.Amount == nil || tx.input.Amount.Cmp(tx.ld.Amount) != 0 {
		return fmt.Errorf("TxTransferCash invalid amount")
	}
	if len(tx.input.Data) > 0 {
		u := uint64(0)
		if err = ld.DecMode.Unmarshal(tx.input.Data, &u); err != nil {
			return fmt.Errorf("TxTransferCash unmarshal lockTime failed: %v", err)
		}
		tx.lockTime = u
	}
	return nil
}

func (tx *TxTakeStake) Verify(bctx BlockContext, bs BlockState) error {
	var err error
	if err = tx.TxBase.Verify(bctx, bs); err != nil {
		return err
	}
	if !tx.to.SatisfySigning(tx.exSigners) {
		return fmt.Errorf("stake account need more signers")
	}
	return tx.to.CheckTakeStake(tx.token, tx.ld.From, tx.ld.Amount, tx.lockTime)
}

func (tx *TxTakeStake) Accept(bctx BlockContext, bs BlockState) error {
	var err error
	// must TakeStake and then Accept
	if err = tx.to.TakeStake(tx.token, tx.ld.From, tx.ld.Amount, tx.lockTime); err != nil {
		return err
	}
	return tx.TxBase.Accept(bctx, bs)
}
