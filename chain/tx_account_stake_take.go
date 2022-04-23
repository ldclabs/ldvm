// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package chain

import (
	"fmt"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ldclabs/ldvm/constants"
	"github.com/ldclabs/ldvm/util"
)

type TxTakeStake struct {
	*TxBase
	exSigners []ids.ShortID
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
	tx.exSigners, err = util.DeriveSigners(tx.ld.UnsignedBytes(), tx.ld.ExSignatures)
	if err != nil {
		return fmt.Errorf("TxUpdateData invalid exSignatures")
	}
	return nil
}

func (tx *TxTakeStake) Verify(blk *Block) error {
	var err error
	if err = tx.TxBase.Verify(blk); err != nil {
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

func (tx *TxTakeStake) Accept(blk *Block) error {
	var err error
	if err = tx.to.TakeStake(tx.ld.From, tx.ld.Amount, blk.FeeConfig().MaxValidatorStake); err != nil {
		return err
	}
	return tx.TxBase.Accept(blk)
}
