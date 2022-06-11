// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package transaction

import (
	"fmt"

	"github.com/ldclabs/ldvm/util"
)

type TxDestroyStakeAccount struct {
	TxBase
}

func (tx *TxDestroyStakeAccount) SyntacticVerify() error {
	var err error
	errPrefix := "TxDestroyStakeAccount.SyntacticVerify failed:"
	if err = tx.TxBase.SyntacticVerify(); err != nil {
		return fmt.Errorf("%s %v", errPrefix, err)
	}

	switch {
	case tx.ld.To == nil:
		return fmt.Errorf("%s nil to as pledge recipient", errPrefix)

	case tx.ld.Token != nil:
		return fmt.Errorf("%s invalid token, should be nil", errPrefix)

	case tx.ld.Amount != nil:
		return fmt.Errorf("%s invalid amount, should be nil", errPrefix)
	}

	if stake := util.StakeSymbol(tx.ld.From); !stake.Valid() {
		return fmt.Errorf("%s invalid stake account %s", errPrefix, stake.GoString())
	}
	return nil
}

func (tx *TxDestroyStakeAccount) Verify(bctx BlockContext, bs BlockState) error {
	var err error
	if err = tx.TxBase.Verify(bctx, bs); err != nil {
		return fmt.Errorf("TxDestroyStakeAccount.Verify failed: %v", err)
	}
	if !tx.from.SatisfySigningPlus(tx.signers) {
		return fmt.Errorf("TxDestroyStakeAccount.Verify failed: invalid signatures for stake keepers, need more")
	}
	if err = tx.from.CheckDestroyStake(tx.to); err != nil {
		return fmt.Errorf("TxDestroyStakeAccount.Verify failed: %v", err)
	}
	return nil
}

func (tx *TxDestroyStakeAccount) Accept(bctx BlockContext, bs BlockState) error {
	if err := tx.TxBase.Accept(bctx, bs); err != nil {
		return err
	}
	// do it after TxBase.Accept
	tx.from.pledge.SetUint64(0)
	return tx.from.DestroyStake(tx.to)
}
