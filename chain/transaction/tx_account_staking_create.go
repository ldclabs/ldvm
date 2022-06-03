// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package transaction

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/ldclabs/ldvm/ld"
	"github.com/ldclabs/ldvm/util"
)

type TxCreateStakeAccount struct {
	TxBase
	input *ld.TxAccounter
	stake *ld.StakeConfig
}

func (tx *TxCreateStakeAccount) MarshalJSON() ([]byte, error) {
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

func (tx *TxCreateStakeAccount) SyntacticVerify() error {
	var err error
	if err = tx.TxBase.SyntacticVerify(); err != nil {
		return err
	}

	if tx.ld.Token != nil {
		return fmt.Errorf("invalid token, expected NativeToken, got %s",
			strconv.Quote(tx.ld.Token.GoString()))
	}
	if tx.ld.To == nil {
		return fmt.Errorf("TxCreateStakeAccount invalid to")
	}
	if token := util.StakeSymbol(*tx.ld.To); !token.Valid() {
		return fmt.Errorf("TxCreateStakeAccount invalid stake address: %s", token.GoString())
	}
	if len(tx.ld.Data) == 0 {
		return fmt.Errorf("TxCreateStakeAccount invalid")
	}

	tx.input = &ld.TxAccounter{}
	if err = tx.input.Unmarshal(tx.ld.Data); err != nil {
		return fmt.Errorf("TxCreateStakeAccount unmarshal data failed: %v", err)
	}
	if err = tx.input.SyntacticVerify(); err != nil {
		return fmt.Errorf("TxCreateStakeAccount SyntacticVerify failed: %v", err)
	}

	if tx.input.Threshold == nil {
		return fmt.Errorf("TxCreateStakeAccount invalid threshold")
	}
	if len(tx.input.Keepers) == 0 {
		return fmt.Errorf("TxCreateStakeAccount invalid keepers")
	}
	if tx.input.Amount != nil {
		return fmt.Errorf("TxCreateStakeAccount invalid amount, please take stake after created")
	}

	tx.stake = &ld.StakeConfig{}
	if err = tx.stake.Unmarshal(tx.input.Data); err != nil {
		return fmt.Errorf("TxCreateStakeAccount unmarshal data failed: %v", err)
	}
	if err = tx.stake.SyntacticVerify(); err != nil {
		return fmt.Errorf("TxCreateStakeAccount SyntacticVerify failed: %v", err)
	}
	if tx.stake.LockTime < tx.ld.Timestamp {
		return fmt.Errorf("TxCreateStakeAccount invalid lockTime")
	}
	return nil
}

func (tx *TxCreateStakeAccount) Verify(bctx BlockContext, bs BlockState) error {
	var err error
	if err = tx.TxBase.Verify(bctx, bs); err != nil {
		return err
	}

	feeCfg := bctx.FeeConfig()
	if tx.ld.Amount.Cmp(feeCfg.MinStakePledge) < 0 {
		return fmt.Errorf("TxCreateStakeAccount invalid amount, expected >= %v, got %v",
			feeCfg.MinStakePledge, tx.ld.Amount)
	}
	return tx.to.CheckCreateStake(tx.ld.From, tx.ld.Amount, tx.input, tx.stake)
}

func (tx *TxCreateStakeAccount) Accept(bctx BlockContext, bs BlockState) error {
	var err error
	if err = tx.to.CreateStake(tx.ld.From, tx.ld.Amount, tx.input, tx.stake); err != nil {
		return err
	}
	return tx.TxBase.Accept(bctx, bs)
}
