// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package transaction

import (
	"encoding/json"
	"fmt"

	"github.com/ldclabs/ldvm/ld"
	"github.com/ldclabs/ldvm/util"
)

type TxResetStakeAccount struct {
	TxBase
	input *ld.StakeConfig
}

func (tx *TxResetStakeAccount) MarshalJSON() ([]byte, error) {
	if tx == nil || tx.ld == nil {
		return []byte("null"), nil
	}
	v := tx.ld.Copy()
	if tx.input == nil {
		return nil, fmt.Errorf("TxResetStakeAccount.MarshalJSON failed: invalid tx.input")
	}
	d, err := json.Marshal(tx.input)
	if err != nil {
		return nil, err
	}
	v.Data = d
	return json.Marshal(v)
}

func (tx *TxResetStakeAccount) SyntacticVerify() error {
	var err error
	errPrefix := "TxResetStakeAccount.SyntacticVerify failed:"
	if err = tx.TxBase.SyntacticVerify(); err != nil {
		return fmt.Errorf("%s %v", errPrefix, err)
	}

	switch {
	case tx.ld.To != nil:
		return fmt.Errorf("%s invalid to, should be nil", errPrefix)

	case tx.ld.Token != nil:
		return fmt.Errorf("%s invalid token, should be nil", errPrefix)

	case tx.ld.Amount != nil:
		return fmt.Errorf("%s invalid amount, should be nil", errPrefix)

	case len(tx.ld.Data) == 0:
		return fmt.Errorf("%s invalid data", errPrefix)
	}

	if stake := util.StakeSymbol(tx.ld.From); !stake.Valid() {
		return fmt.Errorf("%s invalid stake account %s", errPrefix, stake.GoString())
	}

	tx.input = &ld.StakeConfig{}
	if err = tx.input.Unmarshal(tx.ld.Data); err != nil {
		return fmt.Errorf("%s %v", errPrefix, err)
	}
	if err = tx.input.SyntacticVerify(); err != nil {
		return fmt.Errorf("%s %v", errPrefix, err)
	}

	if tx.input.LockTime < tx.ld.Timestamp {
		return fmt.Errorf("%s invalid lockTime, expected >= %d", errPrefix, tx.ld.Timestamp)
	}
	return nil
}

func (tx *TxResetStakeAccount) Verify(bctx BlockContext, bs BlockState) error {
	var err error
	if err = tx.TxBase.Verify(bctx, bs); err != nil {
		return fmt.Errorf("TxResetStakeAccount.Verify failed: %v", err)
	}
	if !tx.from.SatisfySigningPlus(tx.signers) {
		return fmt.Errorf("TxResetStakeAccount.Verify failed: invalid signatures for stake keepers")
	}
	if err = tx.from.CheckResetStake(tx.input); err != nil {
		return fmt.Errorf("TxResetStakeAccount.Verify failed: %v", err)
	}
	return nil
}

func (tx *TxResetStakeAccount) Accept(bctx BlockContext, bs BlockState) error {
	if err := tx.TxBase.Accept(bctx, bs); err != nil {
		return err
	}
	// do it after TxBase.Accept
	return tx.from.ResetStake(tx.input)
}
