// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package transaction

import (
	"encoding/json"
	"fmt"
	"math/big"

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
		return nil, fmt.Errorf("TxCreateStakeAccount.MarshalJSON failed: invalid tx.input")
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
	errPrefix := "TxCreateStakeAccount.SyntacticVerify failed:"
	if err = tx.TxBase.SyntacticVerify(); err != nil {
		return fmt.Errorf("%s %v", errPrefix, err)
	}

	switch {
	case tx.ld.To == nil:
		return fmt.Errorf("%s nil to as stake account", errPrefix)

	case tx.ld.Amount == nil:
		return fmt.Errorf("%s nil amount", errPrefix)

	case tx.ld.Token != nil:
		return fmt.Errorf("%s invalid token, should be nil", errPrefix)

	case len(tx.ld.Data) == 0:
		return fmt.Errorf("%s invalid data", errPrefix)
	}

	if stake := util.StakeSymbol(*tx.ld.To); !stake.Valid() {
		return fmt.Errorf("%s invalid stake account %s", errPrefix, stake.GoString())
	}

	tx.input = &ld.TxAccounter{}
	if err = tx.input.Unmarshal(tx.ld.Data); err != nil {
		return fmt.Errorf("%s %v", errPrefix, err)
	}
	if err = tx.input.SyntacticVerify(); err != nil {
		return fmt.Errorf("%s %v", errPrefix, err)
	}

	switch {
	case tx.input.Threshold == nil || *tx.input.Threshold == 0:
		return fmt.Errorf("%s invalid threshold, expected >= 1", errPrefix)

	case tx.input.Amount != nil:
		return fmt.Errorf("%s invalid amount, should be nil", errPrefix)

	case tx.input.Approver != nil && *tx.input.Approver == util.EthIDEmpty:
		return fmt.Errorf("%s invalid approver, expected not %s", errPrefix, tx.input.Approver)

	case len(tx.input.Data) == 0:
		return fmt.Errorf("%s invalid TxAccounter data", errPrefix)
	}

	tx.stake = &ld.StakeConfig{}
	if err = tx.stake.Unmarshal(tx.input.Data); err != nil {
		return fmt.Errorf("%s %v", errPrefix, err)
	}
	if err = tx.stake.SyntacticVerify(); err != nil {
		return fmt.Errorf("%s %v", errPrefix, err)
	}

	switch {
	case tx.stake.LockTime > 0 && tx.stake.LockTime <= tx.ld.Timestamp:
		return fmt.Errorf("%s invalid lockTime, expected 0 or >= %d", errPrefix, tx.ld.Timestamp)
	}
	return nil
}

func (tx *TxCreateStakeAccount) Verify(bctx BlockContext, bs BlockState) error {
	var err error
	if err = tx.TxBase.Verify(bctx, bs); err != nil {
		return fmt.Errorf("TxCreateStakeAccount.Verify failed: %v", err)
	}

	feeCfg := bctx.FeeConfig()
	if tx.amount.Cmp(feeCfg.MinStakePledge) < 0 {
		return fmt.Errorf("TxCreateStakeAccount.SyntacticVerify Verify: invalid amount, expected >= %v, got %v",
			feeCfg.MinStakePledge, tx.ld.Amount)
	}
	if err = tx.to.CheckCreateStake(tx.ld.From, tx.ld.Amount, tx.input, tx.stake); err != nil {
		return fmt.Errorf("TxCreateStakeAccount.Verify failed: %v", err)
	}
	return nil
}

func (tx *TxCreateStakeAccount) Accept(bctx BlockContext, bs BlockState) error {
	var err error
	if err = tx.to.CreateStake(tx.ld.From, tx.ld.Amount, tx.input, tx.stake); err != nil {
		return err
	}

	pledge := new(big.Int).Set(bctx.FeeConfig().MinStakePledge)
	tx.to.Init(pledge, bs.Height(), bs.Timestamp())
	return tx.TxBase.Accept(bctx, bs)
}
