// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package chain

import (
	"fmt"
	"time"

	"github.com/ava-labs/avalanchego/ids"

	"github.com/ldclabs/ldvm/constants"
	"github.com/ldclabs/ldvm/ld"
	"github.com/ldclabs/ldvm/util"
)

type TxTakeStake struct {
	*TxBase
	exSigners []ids.ShortID
	data      *ld.TxTransfer
}

func (tx *TxTakeStake) MarshalJSON() ([]byte, error) {
	if tx == nil || tx.ld == nil {
		return util.Null, nil
	}
	v := tx.ld.Copy()
	if tx.data == nil {
		tx.data = &ld.TxTransfer{}
		if err := tx.data.Unmarshal(tx.ld.Data); err != nil {
			return nil, fmt.Errorf("TxTakeStake unmarshal failed: %v", err)
		}
	}

	d, err := tx.data.MarshalJSON()
	if err != nil {
		return nil, err
	}
	v.Data = d
	return v.MarshalJSON()
}

func (tx *TxTakeStake) SyntacticVerify() error {
	var err error
	if err = tx.TxBase.SyntacticVerify(); err != nil {
		return err
	}

	if tx.ld.Token != constants.LDCAccount {
		return fmt.Errorf("invalid token %s, required LDC", util.EthID(tx.ld.Token))
	}
	if !util.ValidStakeAddress(tx.ld.To) {
		return fmt.Errorf("TxTakeStake invalid stake address: %s", util.EthID(tx.ld.To))
	}
	if len(tx.ld.Data) == 0 {
		return fmt.Errorf("TxTakeStake invalid")
	}
	tx.exSigners, err = util.DeriveSigners(tx.ld.Data, tx.ld.ExSignatures)
	if err != nil {
		return fmt.Errorf("TxTakeStake invalid exSignatures")
	}

	tx.data = &ld.TxTransfer{}
	if err = tx.data.Unmarshal(tx.ld.Data); err != nil {
		return fmt.Errorf("TxTakeStake unmarshal data failed: %v", err)
	}
	if err = tx.data.SyntacticVerify(); err != nil {
		return fmt.Errorf("TxTakeStake SyntacticVerify failed: %v", err)
	}
	if tx.data.Nonce == tx.ld.Nonce {
		return fmt.Errorf("TxTakeStake invalid nonce")
	}
	if tx.data.Token != constants.LDCAccount {
		return fmt.Errorf("TxTakeStake invalid token")
	}
	if tx.data.From != tx.ld.From {
		return fmt.Errorf("TxTakeStake invalid sender")
	}
	if tx.data.To != tx.ld.To {
		return fmt.Errorf("TxTakeStake invalid recipient")
	}
	if tx.data.Expire > 0 && tx.data.Expire < uint64(time.Now().Unix()) {
		return fmt.Errorf("TxTakeStake expired")
	}
	// tx.ld.Amount can be less than tx.data.Amount
	if tx.data.Amount == nil || tx.ld.Amount == nil || tx.data.Amount.Cmp(tx.ld.Amount) < 0 {
		return fmt.Errorf("TxTransferCash invalid amount")
	}
	return nil
}

func (tx *TxTakeStake) Verify(blk *Block, bs BlockState) error {
	var err error
	if err = tx.TxBase.Verify(blk, bs); err != nil {
		return err
	}
	if tx.to.IsEmpty() {
		return fmt.Errorf("TxTakeStake invalid address, stake account %s not exists", util.EthID(tx.ld.To))
	}
	feeCfg := blk.FeeConfig()
	if tx.ld.Amount.Cmp(feeCfg.MinDelegatorStake) < 0 {
		return fmt.Errorf("TxTakeStake invalid amount, expected >= %v, got %v",
			feeCfg.MinDelegatorStake, tx.ld.Amount)
	}
	if tx.ld.Amount.Cmp(feeCfg.MaxValidatorStake) > 0 {
		return fmt.Errorf("TxTakeStake invalid amount, expected <= %v, got %v",
			feeCfg.MaxValidatorStake, tx.ld.Amount)
	}
	if !tx.to.SatisfySigning(tx.exSigners) {
		return fmt.Errorf("stake account need more signers")
	}
	return nil
}

func (tx *TxTakeStake) Accept(blk *Block, bs BlockState) error {
	var err error
	if err = tx.to.TakeStake(tx.ld.From, tx.ld.Amount, blk.FeeConfig().MaxValidatorStake); err != nil {
		return err
	}
	return tx.TxBase.Accept(blk, bs)
}
