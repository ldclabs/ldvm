// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package chain

import (
	"fmt"
	"math/big"

	"github.com/ldclabs/ldvm/constants"
	"github.com/ldclabs/ldvm/ld"
	"github.com/ldclabs/ldvm/util"
)

type TxCreateTokenAccount struct {
	*TxBase
	data *ld.TxAccounter
}

func (tx *TxCreateTokenAccount) MarshalJSON() ([]byte, error) {
	if tx == nil || tx.ld == nil {
		return util.Null, nil
	}
	v := tx.ld.Copy()
	if tx.data == nil {
		return nil, fmt.Errorf("MarshalJSON failed: data not exists")
	}
	d, err := tx.data.MarshalJSON()
	if err != nil {
		return nil, err
	}
	v.Data = d
	return v.MarshalJSON()
}

func (tx *TxCreateTokenAccount) SyntacticVerify() error {
	var err error
	if err = tx.TxBase.SyntacticVerify(); err != nil {
		return err
	}

	if tx.ld.Token != constants.LDCAccount {
		return fmt.Errorf("invalid token %s, required LDC", util.EthID(tx.ld.Token))
	}

	if token := util.TokenSymbol(tx.ld.To); token.String() == "" {
		return fmt.Errorf("TxCreateTokenAccount invalid token: %s", token)
	}

	if len(tx.ld.Data) == 0 {
		return fmt.Errorf("TxCreateTokenAccount invalid")
	}
	tx.data = &ld.TxAccounter{}
	if err = tx.data.Unmarshal(tx.ld.Data); err != nil {
		return fmt.Errorf("TxCreateTokenAccount unmarshal data failed: %v", err)
	}
	if err = tx.data.SyntacticVerify(); err != nil {
		return fmt.Errorf("TxCreateTokenAccount SyntacticVerify failed: %v", err)
	}

	if tx.data.Threshold == 0 {
		return fmt.Errorf("TxCreateTokenAccount invalid threshold")
	}
	if len(tx.data.Keepers) == 0 {
		return fmt.Errorf("TxCreateTokenAccount invalid keepers")
	}
	if tx.data.Amount == nil || tx.data.Amount.Sign() <= 0 {
		return fmt.Errorf("TxCreateTokenAccount invalid amount")
	}
	return nil
}

func (tx *TxCreateTokenAccount) Verify(blk *Block, bs BlockState) error {
	var err error
	if err = tx.TxBase.Verify(blk, bs); err != nil {
		return err
	}

	feeCfg := blk.FeeConfig()
	if tx.ld.Amount.Cmp(feeCfg.MinTokenPledge) < 0 {
		return fmt.Errorf("TxCreateStakeAccount invalid amount, expected >= %v, got %v",
			feeCfg.MinTokenPledge, tx.ld.Amount)
	}
	return tx.to.CheckCreateToken(tx.ld.To, tx.data)
}

// VerifyGenesis skipping signature verification
func (tx *TxCreateTokenAccount) VerifyGenesis(blk *Block, bs BlockState) error {
	var err error
	tx.data = &ld.TxAccounter{}
	if err = tx.data.Unmarshal(tx.ld.Data); err != nil {
		return fmt.Errorf("TxCreateTokenAccount unmarshal data failed: %v", err)
	}
	if err = tx.data.SyntacticVerify(); err != nil {
		return fmt.Errorf("TxCreateTokenAccount SyntacticVerify failed: %v", err)
	}

	tx.tip = new(big.Int)
	tx.fee = new(big.Int)
	tx.cost = new(big.Int)

	tx.from, err = bs.LoadAccount(tx.ld.From)
	if err != nil {
		return err
	}

	tx.from.Add(constants.LDCAccount, blk.ctx.Chain().MaxTotalSupply)
	if tx.ldc, err = bs.LoadAccount(constants.LDCAccount); err != nil {
		return err
	}
	if tx.miner, err = blk.Miner(); err != nil {
		return err
	}
	tx.to, err = bs.LoadAccount(tx.ld.To)
	return nil
}

func (tx *TxCreateTokenAccount) Accept(blk *Block, bs BlockState) error {
	var err error
	if err = tx.to.CreateToken(tx.ld.To, tx.data); err != nil {
		return err
	}
	return tx.TxBase.Accept(blk, bs)
}
