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
		return nil, fmt.Errorf("TxCreateTokenAccount.MarshalJSON failed: invalid tx.input")
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
	errPrefix := "TxCreateTokenAccount.SyntacticVerify failed:"
	if err = tx.TxBase.SyntacticVerify(); err != nil {
		return fmt.Errorf("%s %v", errPrefix, err)
	}

	switch {
	case tx.ld.To == nil:
		return fmt.Errorf("%s nil to as token account", errPrefix)

	case *tx.ld.To == util.EthIDEmpty:
		return fmt.Errorf("%s invalid to as token account, expected not %s", errPrefix, tx.ld.To)

	case tx.ld.Amount == nil:
		return fmt.Errorf("%s nil amount", errPrefix)

	case tx.ld.Token != nil:
		return fmt.Errorf("%s invalid token, should be nil", errPrefix)

	case len(tx.ld.Data) == 0:
		return fmt.Errorf("%s invalid data", errPrefix)
	}

	if token := util.TokenSymbol(*tx.ld.To); !token.Valid() {
		return fmt.Errorf("%s invalid token %s", errPrefix, token.GoString())
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

	case tx.input.Amount == nil || tx.input.Amount.Sign() <= 0:
		return fmt.Errorf("%s invalid amount, expected >= 1", errPrefix)

	case tx.input.Approver != nil && *tx.input.Approver == util.EthIDEmpty:
		return fmt.Errorf("%s invalid approver, expected not %s", errPrefix, tx.input.Approver)

	case len(tx.input.Name) < 3:
		return fmt.Errorf("%s invalid name %s, expected length >= 3", errPrefix, strconv.Quote(tx.input.Name))
	}
	return nil
}

func (tx *TxCreateTokenAccount) Verify(bctx BlockContext, bs BlockState) error {
	var err error
	if err = tx.TxBase.Verify(bctx, bs); err != nil {
		return fmt.Errorf("TxCreateTokenAccount.Verify failed: %v", err)
	}

	feeCfg := bctx.FeeConfig()
	if tx.ld.Amount.Cmp(feeCfg.MinTokenPledge) < 0 {
		return fmt.Errorf("TxCreateTokenAccount.Verify failed: invalid amount, expected >= %v, got %v",
			feeCfg.MinTokenPledge, tx.ld.Amount)
	}
	if err = tx.to.CheckCreateToken(tx.input); err != nil {
		return fmt.Errorf("TxCreateTokenAccount.Verify failed: %v", err)
	}
	return nil
}

// VerifyGenesis skipping signature verification
func (tx *TxCreateTokenAccount) VerifyGenesis(bctx BlockContext, bs BlockState) error {
	var err error
	tx.input = &ld.TxAccounter{}
	if err = tx.input.Unmarshal(tx.ld.Data); err != nil {
		return fmt.Errorf("TxCreateTokenAccount.VerifyGenesis failed: %v", err)
	}
	if err = tx.input.SyntacticVerify(); err != nil {
		return fmt.Errorf("TxCreateTokenAccount.VerifyGenesis failed: %v", err)
	}

	tx.amount = new(big.Int)
	tx.tip = new(big.Int)
	tx.fee = new(big.Int)
	tx.cost = new(big.Int)

	tx.from, err = bs.LoadAccount(tx.ld.From)
	if err != nil {
		return err
	}
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
	if tx.to.id != constants.LDCAccount {
		pledge := new(big.Int).Set(bctx.FeeConfig().MinTokenPledge)
		tx.to.Init(pledge, bs.Height(), bs.Timestamp())
	}
	return tx.TxBase.Accept(bctx, bs)
}
