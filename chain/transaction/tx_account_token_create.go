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
		return nil, fmt.Errorf("TxCreateTokenAccount.MarshalJSON error: invalid tx.input")
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
	errp := util.ErrPrefix("TxCreateTokenAccount.SyntacticVerify error: ")

	if err = tx.TxBase.SyntacticVerify(); err != nil {
		return errp.ErrorIf(err)
	}

	switch {
	case tx.ld.To == nil:
		return errp.Errorf("nil to as token account")

	case *tx.ld.To == util.EthIDEmpty:
		return errp.Errorf("invalid to as token account, expected not %s", tx.ld.To)

	case tx.ld.Amount == nil:
		return errp.Errorf("nil amount")

	case tx.ld.Token != nil:
		return errp.Errorf("invalid token, should be nil")

	case len(tx.ld.Data) == 0:
		return errp.Errorf("invalid data")
	}

	if token := util.TokenSymbol(*tx.ld.To); !token.Valid() {
		return errp.Errorf("invalid token %s", token.GoString())
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

	case tx.input.Amount == nil || tx.input.Amount.Sign() <= 0:
		return errp.Errorf("invalid amount, expected >= 1")

	case tx.input.Approver != nil && *tx.input.Approver == util.EthIDEmpty:
		return errp.Errorf("invalid approver, expected not %s", tx.input.Approver)

	case len(tx.input.Name) < 3:
		return errp.Errorf("invalid name %s, expected length >= 3", strconv.Quote(tx.input.Name))
	}
	return nil
}

func (tx *TxCreateTokenAccount) Verify(bctx BlockContext, bs BlockState) error {
	var err error
	errp := util.ErrPrefix("TxCreateTokenAccount.Verify error: ")

	if err = tx.TxBase.Verify(bctx, bs); err != nil {
		return errp.ErrorIf(err)
	}

	feeCfg := bctx.FeeConfig()
	if tx.ld.Amount.Cmp(feeCfg.MinTokenPledge) < 0 {
		return errp.Errorf("invalid amount, expected >= %v, got %v",
			feeCfg.MinTokenPledge, tx.ld.Amount)
	}
	if err = tx.to.CheckCreateToken(tx.input); err != nil {
		return errp.ErrorIf(err)
	}
	return nil
}

// VerifyGenesis skipping signature verification
func (tx *TxCreateTokenAccount) VerifyGenesis(bctx BlockContext, bs BlockState) error {
	var err error
	errp := util.ErrPrefix("TxCreateTokenAccount.VerifyGenesis error: ")

	tx.input = &ld.TxAccounter{}
	if err = tx.input.Unmarshal(tx.ld.Data); err != nil {
		return errp.ErrorIf(err)
	}
	if err = tx.input.SyntacticVerify(); err != nil {
		return errp.ErrorIf(err)
	}

	tx.amount = new(big.Int)
	tx.tip = new(big.Int)
	tx.fee = new(big.Int)
	tx.cost = new(big.Int)

	tx.from, err = bs.LoadAccount(tx.ld.From)
	if err != nil {
		return errp.ErrorIf(err)
	}
	if tx.ldc, err = bs.LoadAccount(constants.LDCAccount); err != nil {
		return errp.ErrorIf(err)
	}
	if tx.miner, err = bs.LoadMiner(bctx.Miner()); err != nil {
		return errp.ErrorIf(err)
	}
	tx.to, err = bs.LoadAccount(*tx.ld.To)
	return errp.ErrorIf(err)
}

func (tx *TxCreateTokenAccount) Accept(bctx BlockContext, bs BlockState) error {
	var err error
	errp := util.ErrPrefix("TxCreateTokenAccount.Accept error: ")

	if err = tx.to.CreateToken(tx.input); err != nil {
		return errp.ErrorIf(err)
	}
	if tx.to.id != constants.LDCAccount {
		pledge := new(big.Int).Set(bctx.FeeConfig().MinTokenPledge)
		tx.to.Init(pledge, bs.Height(), bs.Timestamp())
	}
	return errp.ErrorIf(tx.TxBase.Accept(bctx, bs))
}
