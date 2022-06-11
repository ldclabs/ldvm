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

type TxTakeStake struct {
	TxBase
	exSigners util.EthIDs
	input     *ld.TxTransfer
	lockTime  uint64
}

func (tx *TxTakeStake) MarshalJSON() ([]byte, error) {
	if tx == nil || tx.ld == nil {
		return []byte("null"), nil
	}
	v := tx.ld.Copy()
	if tx.input == nil {
		return nil, fmt.Errorf("TxTakeStake.MarshalJSON failed: invalid tx.input")
	}
	d, err := json.Marshal(tx.input)
	if err != nil {
		return nil, err
	}
	v.Data = d
	return json.Marshal(v)
}

func (tx *TxTakeStake) SyntacticVerify() error {
	var err error
	errPrefix := "TxTakeStake.SyntacticVerify failed:"
	if err = tx.TxBase.SyntacticVerify(); err != nil {
		return fmt.Errorf("%s %v", errPrefix, err)
	}

	switch {
	case tx.ld.To == nil:
		return fmt.Errorf("%s nil to as stake account", errPrefix)

	case tx.ld.Amount == nil:
		return fmt.Errorf("%s nil amount", errPrefix)

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
	case tx.input.From == nil:
		return fmt.Errorf("%s nil from", errPrefix)

	case *tx.input.From != tx.ld.From:
		return fmt.Errorf("%s invalid from, expected %s, got %s",
			errPrefix, tx.input.From, tx.ld.From)

	case tx.input.Token == nil && tx.token != constants.NativeToken:
		return fmt.Errorf("%s invalid token, expected %s, got %s",
			errPrefix, constants.NativeToken.GoString(), tx.token.GoString())

	case tx.input.Token != nil && tx.token != *tx.input.Token:
		return fmt.Errorf("%s invalid token, expected %s, got %s",
			errPrefix, tx.input.Token.GoString(), tx.token.GoString())

	case tx.input.Amount == nil:
		return fmt.Errorf("%s nil amount", errPrefix)

	case tx.input.Amount.Cmp(tx.ld.Amount) != 0:
		return fmt.Errorf("%s invalid amount, expected %v, got %v",
			errPrefix, tx.input.Amount, tx.ld.Amount)

	case tx.input.Expire < tx.ld.Timestamp:
		return fmt.Errorf("%s data expired", errPrefix)
	}

	if len(tx.input.Data) > 0 {
		u := uint64(0)
		if err = ld.DecMode.Unmarshal(tx.input.Data, &u); err != nil {
			return fmt.Errorf("%s invalid lockTime, %v", errPrefix, err)
		}
		tx.lockTime = u
	}
	tx.exSigners, err = tx.ld.ExSigners()
	if err != nil {
		return fmt.Errorf("%s invalid exSignatures, %v", errPrefix, err)
	}
	return nil
}

func (tx *TxTakeStake) Verify(bctx BlockContext, bs BlockState) error {
	var err error
	if err = tx.TxBase.Verify(bctx, bs); err != nil {
		return fmt.Errorf("TxTakeStake.Verify failed: %v", err)
	}
	if !tx.to.SatisfySigning(tx.exSigners) {
		return fmt.Errorf("TxTakeStake.Verify failed: invalid exSignatures for stake keepers")
	}
	if err = tx.to.CheckTakeStake(tx.token, tx.ld.From, tx.ld.Amount, tx.lockTime); err != nil {
		return fmt.Errorf("TxTakeStake.Verify failed: %v", err)
	}
	return nil
}

func (tx *TxTakeStake) Accept(bctx BlockContext, bs BlockState) error {
	var err error
	// must TakeStake and then Accept
	if err = tx.to.TakeStake(tx.token, tx.ld.From, tx.ld.Amount, tx.lockTime); err != nil {
		return err
	}
	return tx.TxBase.Accept(bctx, bs)
}
