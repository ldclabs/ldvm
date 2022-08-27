// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package transactions

import (
	"encoding/json"
	"math/big"

	"github.com/ldclabs/ldvm/constants"
	"github.com/ldclabs/ldvm/ld"
	"github.com/ldclabs/ldvm/util"
)

type TxCreateModel struct {
	TxBase
	input *ld.ModelInfo
}

func (tx *TxCreateModel) MarshalJSON() ([]byte, error) {
	if tx == nil || tx.ld == nil {
		return []byte("null"), nil
	}

	v := tx.ld.Copy()
	errp := util.ErrPrefix("TxCreateModel.MarshalJSON error: ")
	if tx.input == nil {
		return nil, errp.Errorf("nil tx.input")
	}
	d, err := json.Marshal(tx.input)
	if err != nil {
		return nil, errp.ErrorIf(err)
	}
	v.Data = d
	return errp.ErrorMap(json.Marshal(v))
}

func (tx *TxCreateModel) SyntacticVerify() error {
	var err error
	errp := util.ErrPrefix("TxCreateModel.SyntacticVerify error: ")

	if err = tx.TxBase.SyntacticVerify(); err != nil {
		return errp.ErrorIf(err)
	}

	switch {
	case tx.ld.To != nil:
		return errp.Errorf("invalid to, should be nil")

	case tx.ld.Token != nil:
		return errp.Errorf("invalid token, should be nil")

	case tx.ld.Amount != nil:
		return errp.Errorf("invalid amount, should be nil")

	case len(tx.ld.Data) == 0:
		return errp.Errorf("invalid data")
	}

	tx.input = &ld.ModelInfo{}
	if err = tx.input.Unmarshal(tx.ld.Data); err != nil {
		return errp.ErrorIf(err)
	}
	if err = tx.input.SyntacticVerify(); err != nil {
		return errp.ErrorIf(err)
	}
	tx.input.ID = util.ModelID(tx.ld.ShortID())
	return nil
}

// ApplyGenesis skipping signature verification
func (tx *TxCreateModel) ApplyGenesis(ctx ChainContext, cs ChainState) error {
	var err error
	errp := util.ErrPrefix("TxCreateModel.ApplyGenesis error: ")

	tx.input = &ld.ModelInfo{}
	if err = tx.input.Unmarshal(tx.ld.Data); err != nil {
		return errp.ErrorIf(err)
	}
	if err = tx.input.SyntacticVerify(); err != nil {
		return errp.ErrorIf(err)
	}
	tx.input.ID = util.ModelID(tx.ld.ShortID())

	tx.amount = new(big.Int)
	tx.tip = new(big.Int)
	tx.fee = new(big.Int)
	tx.cost = new(big.Int)

	if tx.ldc, err = cs.LoadAccount(constants.LDCAccount); err != nil {
		return errp.ErrorIf(err)
	}

	if tx.miner, err = cs.LoadMiner(ctx.Miner()); err != nil {
		return errp.ErrorIf(err)
	}

	if tx.from, err = cs.LoadAccount(tx.ld.From); err != nil {
		return errp.ErrorIf(err)
	}

	if err = cs.SaveModel(tx.input); err != nil {
		return errp.ErrorIf(err)
	}

	// create model account
	if len(tx.input.Keepers) > 0 {
		modelAcc, err := cs.LoadAccount(util.EthID(tx.input.ID))
		if err != nil {
			return errp.ErrorIf(err)
		}
		if !modelAcc.IsEmpty() {
			return errp.Errorf("model account %s exists", modelAcc.ID())
		}

		threshold := tx.input.Threshold
		if threshold == 0 {
			threshold = 1
		}
		if err = modelAcc.UpdateKeepers(
			&threshold, &tx.input.Keepers, nil, nil); err != nil {
			return errp.ErrorIf(err)
		}
	}
	return errp.ErrorIf(tx.TxBase.accept(ctx, cs))
}

func (tx *TxCreateModel) Apply(ctx ChainContext, cs ChainState) error {
	var err error
	errp := util.ErrPrefix("TxCreateModel.Apply error: ")

	if err = tx.TxBase.verify(ctx, cs); err != nil {
		return errp.ErrorIf(err)
	}
	if err = cs.SaveModel(tx.input); err != nil {
		return errp.ErrorIf(err)
	}

	// create model account
	if len(tx.input.Keepers) > 0 {
		modelAcc, err := cs.LoadAccount(util.EthID(tx.input.ID))
		if err != nil {
			return errp.ErrorIf(err)
		}

		if !modelAcc.IsEmpty() {
			return errp.Errorf("model account %s exists", modelAcc.ID())
		}

		threshold := tx.input.Threshold
		if threshold == 0 {
			threshold = 1
		}
		if err = modelAcc.UpdateKeepers(
			&threshold, &tx.input.Keepers, nil, nil); err != nil {
			return errp.ErrorIf(err)
		}
	}
	return errp.ErrorIf(tx.TxBase.accept(ctx, cs))
}