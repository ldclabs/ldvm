// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package transactions

import (
	"encoding/json"
	"math/big"

	"github.com/ldclabs/ldvm/ld"
	"github.com/ldclabs/ldvm/util"
)

type TxCreateStake struct {
	TxBase
	input *ld.TxAccounter
	stake *ld.StakeConfig
}

func (tx *TxCreateStake) MarshalJSON() ([]byte, error) {
	if tx == nil || tx.ld == nil {
		return []byte("null"), nil
	}

	v := tx.ld.Copy()
	errp := util.ErrPrefix("TxCreateStake.MarshalJSON error: ")
	if tx.input == nil {
		return nil, errp.Errorf("nil tx.input")
	}
	d, err := json.Marshal(tx.input)
	if err != nil {
		return nil, errp.ErrorIf(err)
	}
	v.Tx.Data = d
	return errp.ErrorMap(json.Marshal(v))
}

func (tx *TxCreateStake) SyntacticVerify() error {
	var err error
	errp := util.ErrPrefix("TxCreateStake.SyntacticVerify error: ")

	if err = tx.TxBase.SyntacticVerify(); err != nil {
		return errp.ErrorIf(err)
	}

	switch {
	case tx.ld.Tx.To == nil:
		return errp.Errorf("nil to as stake account")

	case tx.ld.Tx.Amount == nil:
		return errp.Errorf("nil amount")

	case tx.ld.Tx.Token != nil:
		return errp.Errorf("invalid token, should be nil")

	case len(tx.ld.Tx.Data) == 0:
		return errp.Errorf("invalid data")
	}

	if stake := util.StakeSymbol(*tx.ld.Tx.To); !stake.Valid() {
		return errp.Errorf("invalid stake account %s", stake.GoString())
	}

	tx.input = &ld.TxAccounter{}
	if err = tx.input.Unmarshal(tx.ld.Tx.Data); err != nil {
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
		return errp.Errorf("invalid input data")
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
		return errp.Errorf("invalid lockTime, expected 0 or >= %d, got %d", tx.ld.Timestamp, tx.stake.LockTime)
	}
	return nil
}

func (tx *TxCreateStake) Apply(ctx ChainContext, cs ChainState) error {
	var err error
	errp := util.ErrPrefix("TxCreateStake.Apply error: ")

	if err = tx.TxBase.verify(ctx, cs); err != nil {
		return errp.ErrorIf(err)
	}

	pledge := new(big.Int).Set(ctx.FeeConfig().MinStakePledge)
	if tx.amount.Cmp(pledge) < 0 {
		return errp.Errorf("invalid amount, expected >= %v, got %v",
			pledge, tx.ld.Tx.Amount)
	}

	if err = cs.LoadLedger(tx.to); err != nil {
		return errp.ErrorIf(err)
	}

	if err = tx.to.CreateStake(tx.ld.Tx.From, pledge, tx.input, tx.stake); err != nil {
		return errp.ErrorIf(err)
	}

	tx.to.Init(pledge, cs.Height(), cs.Timestamp())
	return errp.ErrorIf(tx.TxBase.accept(ctx, cs))
}
