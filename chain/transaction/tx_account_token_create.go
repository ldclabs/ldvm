// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package transaction

import (
	"encoding/json"
	"fmt"
	"math/big"
	"strconv"

	"github.com/ldclabs/ldvm/constants"
	"github.com/ldclabs/ldvm/ld"
	"github.com/ldclabs/ldvm/util"
)

type TxCreateTokenAccount struct {
	TxBase
	input *ld.TxAccounter
}

func (tx *TxCreateTokenAccount) MarshalJSON() ([]byte, error) {
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

func (tx *TxCreateTokenAccount) SyntacticVerify() error {
	var err error
	if err = tx.TxBase.SyntacticVerify(); err != nil {
		return err
	}

	if tx.ld.Token != nil {
		return fmt.Errorf("invalid token, expected NativeToken, got %s",
			strconv.Quote(tx.ld.Token.GoString()))
	}
	if tx.ld.To == nil {
		return fmt.Errorf("TxCreateTokenAccount invalid to")
	}

	if token := util.TokenSymbol(*tx.ld.To); !token.Valid() {
		return fmt.Errorf("TxCreateTokenAccount invalid token: %s", token.GoString())
	}

	if len(tx.ld.Data) == 0 {
		return fmt.Errorf("TxCreateTokenAccount invalid")
	}
	tx.input = &ld.TxAccounter{}
	if err = tx.input.Unmarshal(tx.ld.Data); err != nil {
		return fmt.Errorf("TxCreateTokenAccount unmarshal data failed: %v", err)
	}
	if err = tx.input.SyntacticVerify(); err != nil {
		return fmt.Errorf("TxCreateTokenAccount SyntacticVerify failed: %v", err)
	}

	if tx.input.Threshold == nil {
		return fmt.Errorf("TxCreateTokenAccount invalid threshold")
	}
	if len(tx.input.Keepers) == 0 {
		return fmt.Errorf("TxCreateTokenAccount invalid keepers")
	}
	if tx.input.Amount == nil || tx.input.Amount.Sign() <= 0 {
		return fmt.Errorf("TxCreateTokenAccount invalid amount")
	}
	return nil
}

func (tx *TxCreateTokenAccount) Verify(bctx BlockContext, bs BlockState) error {
	var err error
	if err = tx.TxBase.Verify(bctx, bs); err != nil {
		return err
	}

	feeCfg := bctx.FeeConfig()
	if tx.ld.Amount.Cmp(feeCfg.MinTokenPledge) < 0 {
		return fmt.Errorf("TxCreateStakeAccount invalid amount, expected >= %v, got %v",
			feeCfg.MinTokenPledge, tx.ld.Amount)
	}
	return tx.to.CheckCreateToken(tx.input)
}

// VerifyGenesis skipping signature verification
func (tx *TxCreateTokenAccount) VerifyGenesis(bctx BlockContext, bs BlockState) error {
	var err error
	tx.input = &ld.TxAccounter{}
	if err = tx.input.Unmarshal(tx.ld.Data); err != nil {
		return fmt.Errorf("TxCreateTokenAccount unmarshal data failed: %v", err)
	}
	if err = tx.input.SyntacticVerify(); err != nil {
		return fmt.Errorf("TxCreateTokenAccount SyntacticVerify failed: %v", err)
	}

	tx.tip = new(big.Int)
	tx.fee = new(big.Int)
	tx.cost = new(big.Int)

	tx.from, err = bs.LoadAccount(tx.ld.From)
	if err != nil {
		return err
	}

	tx.from.Add(constants.NativeToken, bctx.Chain().MaxTotalSupply)
	if tx.ldc, err = bs.LoadAccount(constants.LDCAccount); err != nil {
		return err
	}
	if tx.miner, err = bs.LoadMiner(bctx.Miner()); err != nil {
		return err
	}
	tx.to, err = bs.LoadAccount(*tx.ld.To)
	return nil
}

func (tx *TxCreateTokenAccount) Accept(bctx BlockContext, bs BlockState) error {
	var err error
	if err = tx.to.CreateToken(tx.input); err != nil {
		return err
	}
	return tx.TxBase.Accept(bctx, bs)
}
