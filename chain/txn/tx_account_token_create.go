// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package txn

import (
	"encoding/json"
	"math/big"

	"github.com/ldclabs/ldvm/ids"
	"github.com/ldclabs/ldvm/ld"
	"github.com/ldclabs/ldvm/util/erring"
)

type TxCreateToken struct {
	TxBase
	input *ld.TxAccounter
}

func (tx *TxCreateToken) MarshalJSON() ([]byte, error) {
	if tx == nil || tx.ld == nil {
		return []byte("null"), nil
	}

	v := tx.ld.Copy()
	errp := erring.ErrPrefix("txn.TxCreateToken.MarshalJSON: ")
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

func (tx *TxCreateToken) SyntacticVerify() error {
	var err error
	errp := erring.ErrPrefix("txn.TxCreateToken.SyntacticVerify: ")

	if err = tx.TxBase.SyntacticVerify(); err != nil {
		return errp.ErrorIf(err)
	}

	switch {
	case tx.ld.Tx.To == nil:
		return errp.Errorf("nil to as token account")

	case *tx.ld.Tx.To == ids.EmptyAddress:
		return errp.Errorf("invalid to as token account, expected not %s", tx.ld.Tx.To)

	case tx.ld.Tx.Amount == nil:
		return errp.Errorf("nil amount")

	case tx.ld.Tx.Token != nil:
		return errp.Errorf("invalid token, should be nil")

	case len(tx.ld.Tx.Data) == 0:
		return errp.Errorf("invalid data")
	}

	if token := ids.TokenSymbol(*tx.ld.Tx.To); !token.Valid() {
		return errp.Errorf("invalid token %s", token.GoString())
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

	case tx.input.Amount == nil || tx.input.Amount.Sign() <= 0:
		return errp.Errorf("invalid amount, expected >= 1")

	case len(tx.input.Name) < 3:
		return errp.Errorf("invalid name %q, expected length >= 3", tx.input.Name)
	}

	if tx.input.Approver != nil {
		if err = tx.input.Approver.Valid(); err != nil {
			return errp.Errorf("invalid approver, %v", err)
		}
	}
	return nil
}

// ApplyGenesis skipping signature verification
func (tx *TxCreateToken) ApplyGenesis(ctx ChainContext, cs ChainState) error {
	var err error
	errp := erring.ErrPrefix("txn.TxCreateToken.ApplyGenesis: ")

	tx.input = &ld.TxAccounter{}
	if err = tx.input.Unmarshal(tx.ld.Tx.Data); err != nil {
		return errp.ErrorIf(err)
	}
	if err = tx.input.SyntacticVerify(); err != nil {
		return errp.ErrorIf(err)
	}

	tx.amount = new(big.Int)
	tx.tip = new(big.Int)
	tx.fee = new(big.Int)
	tx.cost = new(big.Int)

	tx.from, err = cs.LoadAccount(tx.ld.Tx.From)
	if err != nil {
		return errp.ErrorIf(err)
	}
	if tx.ldc, err = cs.LoadAccount(ids.LDCAccount); err != nil {
		return errp.ErrorIf(err)
	}
	if tx.miner, err = cs.LoadBuilder(ctx.Builder()); err != nil {
		return errp.ErrorIf(err)
	}
	if tx.to, err = cs.LoadAccount(*tx.ld.Tx.To); err != nil {
		return errp.ErrorIf(err)
	}

	if err = tx.to.CreateToken(tx.input); err != nil {
		return errp.ErrorIf(err)
	}
	if tx.to.ID() != ids.LDCAccount {
		feeCfg := ctx.FeeConfig()
		tx.to.Init(big.NewInt(0), feeCfg.MinTokenPledge, cs.Height(), cs.Timestamp())
	}
	return errp.ErrorIf(tx.TxBase.accept(ctx, cs))
}

func (tx *TxCreateToken) Apply(ctx ChainContext, cs ChainState) error {
	var err error
	errp := erring.ErrPrefix("txn.TxCreateToken.Apply: ")

	if err = tx.TxBase.verify(ctx, cs); err != nil {
		return errp.ErrorIf(err)
	}

	feeCfg := ctx.FeeConfig()
	if tx.ld.Tx.Amount.Cmp(feeCfg.MinTokenPledge) < 0 {
		return errp.Errorf("invalid amount, expected >= %v, got %v",
			feeCfg.MinTokenPledge, tx.ld.Tx.Amount)
	}

	if err = tx.to.CreateToken(tx.input); err != nil {
		return errp.ErrorIf(err)
	}
	if tx.to.ID() != ids.LDCAccount {
		tx.to.Init(big.NewInt(0), feeCfg.MinTokenPledge, cs.Height(), cs.Timestamp())
	}
	return errp.ErrorIf(tx.TxBase.accept(ctx, cs))
}
