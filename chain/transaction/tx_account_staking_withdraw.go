// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package transaction

import (
	"encoding/json"
	"fmt"

	"github.com/ldclabs/ldvm/constants"
	"github.com/ldclabs/ldvm/ld"
	"github.com/ldclabs/ldvm/util"
)

type TxWithdrawStake struct {
	TxBase
	input *ld.TxTransfer
}

func (tx *TxWithdrawStake) MarshalJSON() ([]byte, error) {
	if tx == nil || tx.ld == nil {
		return []byte("null"), nil
	}
	v := tx.ld.Copy()
	if tx.input == nil {
		return nil, fmt.Errorf("TxWithdrawStake.MarshalJSON failed: invalid tx.input")
	}
	d, err := json.Marshal(tx.input)
	if err != nil {
		return nil, err
	}
	v.Data = d
	return json.Marshal(v)
}

func (tx *TxWithdrawStake) SyntacticVerify() error {
	var err error
	errPrefix := "TxWithdrawStake.SyntacticVerify failed:"
	if err = tx.TxBase.SyntacticVerify(); err != nil {
		return fmt.Errorf("%s %v", errPrefix, err)
	}

	switch {
	case tx.ld.To == nil:
		return fmt.Errorf("%s nil to as stake account", errPrefix)

	case tx.ld.Amount != nil:
		return fmt.Errorf("%s invalid amount, should be nil", errPrefix)

	case len(tx.ld.Data) == 0:
		return fmt.Errorf("%s invalid data", errPrefix)
	}

	if stake := util.StakeSymbol(*tx.ld.To); !stake.Valid() {
		return fmt.Errorf("%s invalid stake account %s", errPrefix, stake.GoString())
	}

	tx.input = &ld.TxTransfer{}
	if err = tx.input.Unmarshal(tx.ld.Data); err != nil {
		return fmt.Errorf("%s %v", errPrefix, err)
	}
	if err = tx.input.SyntacticVerify(); err != nil {
		return fmt.Errorf("%s %v", errPrefix, err)
	}

	switch {
	case tx.input.Token == nil && tx.token != constants.NativeToken:
		return fmt.Errorf("%s invalid token, expected %s, got %s",
			errPrefix, constants.NativeToken.GoString(), tx.token.GoString())

	case tx.input.Token != nil && tx.token != *tx.input.Token:
		return fmt.Errorf("%s invalid token, expected %s, got %s",
			errPrefix, tx.input.Token.GoString(), tx.token.GoString())

	case tx.input.Amount == nil || tx.input.Amount.Sign() <= 0:
		return fmt.Errorf("%s invalid amount, expected >= 1", errPrefix)
	}
	return nil
}

func (tx *TxWithdrawStake) Verify(bctx BlockContext, bs BlockState) error {
	var err error
	if err = tx.TxBase.Verify(bctx, bs); err != nil {
		return fmt.Errorf("TxWithdrawStake.Verify failed: %v", err)
	}
	if err = tx.to.CheckWithdrawStake(tx.token, tx.ld.From, tx.signers, tx.input.Amount); err != nil {
		return fmt.Errorf("TxWithdrawStake.Verify failed: %v", err)
	}
	return nil
}

func (tx *TxWithdrawStake) Accept(bctx BlockContext, bs BlockState) error {
	// must WithdrawStake and then Accept
	withdraw, err := tx.to.WithdrawStake(tx.token, tx.ld.From, tx.signers, tx.input.Amount)
	if err != nil {
		return err
	}
	if err = tx.to.Sub(tx.token, withdraw); err != nil {
		return err
	}
	if err = tx.from.Add(tx.token, withdraw); err != nil {
		return err
	}
	return tx.TxBase.Accept(bctx, bs)
}
