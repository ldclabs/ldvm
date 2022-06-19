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
		return nil, fmt.Errorf("TxCreateStakeAccount.MarshalJSON error: invalid tx.input")
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
	errp := util.ErrPrefix("TxCreateStakeAccount.SyntacticVerify error: ")

	if err = tx.TxBase.SyntacticVerify(); err != nil {
		return errp.ErrorIf(err)
	}

	switch {
	case tx.ld.To == nil:
		return errp.Errorf("nil to as stake account")

	case tx.ld.Amount == nil:
		return errp.Errorf("nil amount")

	case tx.ld.Token != nil:
		return errp.Errorf("invalid token, should be nil")

	case len(tx.ld.Data) == 0:
		return errp.Errorf("invalid data")
	}

	if stake := util.StakeSymbol(*tx.ld.To); !stake.Valid() {
		return errp.Errorf("invalid stake account %s", stake.GoString())
	}

	tx.input = &ld.TxAccounter{}
	if err = tx.input.Unmarshal(tx.ld.Data); err != nil {
		return errp.ErrorIf(err)
	}
	if err = tx.input.SyntacticVerify(); err != nil {
		return errp.ErrorIf(err)
	}

	switch {
	case tx.input.Threshold == nil || *tx.input.Threshold == 0:
		return errp.Errorf("invalid threshold, expected >= 1")

	case tx.input.Amount != nil:
		return errp.Errorf("invalid amount, should be nil")

	case tx.input.Approver != nil && *tx.input.Approver == util.EthIDEmpty:
		return errp.Errorf("invalid approver, expected not %s", tx.input.Approver)

	case len(tx.input.Data) == 0:
		return errp.Errorf("invalid TxAccounter data")
	}

	tx.stake = &ld.StakeConfig{}
	if err = tx.stake.Unmarshal(tx.input.Data); err != nil {
		return errp.ErrorIf(err)
	}
	if err = tx.stake.SyntacticVerify(); err != nil {
		return errp.ErrorIf(err)
	}

	switch {
	case tx.stake.LockTime > 0 && tx.stake.LockTime <= tx.ld.Timestamp:
		return errp.Errorf("invalid lockTime, expected 0 or >= %d", tx.ld.Timestamp)
	}
	return nil
}

func (tx *TxCreateStakeAccount) Verify(bctx BlockContext, bs BlockState) error {
	var err error
	errp := util.ErrPrefix("TxCreateStakeAccount.Verify error: ")

	if err = tx.TxBase.Verify(bctx, bs); err != nil {
		return errp.ErrorIf(err)
	}

	feeCfg := bctx.FeeConfig()
	if tx.amount.Cmp(feeCfg.MinStakePledge) < 0 {
		return errp.Errorf("invalid amount, expected >= %v, got %v",
			feeCfg.MinStakePledge, tx.ld.Amount)
	}
	if err = tx.to.CheckCreateStake(tx.ld.From, tx.ld.Amount, tx.input, tx.stake); err != nil {
		return errp.ErrorIf(err)
	}
	return nil
}

func (tx *TxCreateStakeAccount) Accept(bctx BlockContext, bs BlockState) error {
	var err error
	errp := util.ErrPrefix("TxCreateStakeAccount.Accept error: ")

	if err = tx.to.CreateStake(tx.ld.From, tx.ld.Amount, tx.input, tx.stake); err != nil {
		return errp.ErrorIf(err)
	}

	pledge := new(big.Int).Set(bctx.FeeConfig().MinStakePledge)
	tx.to.Init(pledge, bs.Height(), bs.Timestamp())
	return errp.ErrorIf(tx.TxBase.Accept(bctx, bs))
}
