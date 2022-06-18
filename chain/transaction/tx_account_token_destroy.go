// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package transaction

import (
	"fmt"

	"github.com/ldclabs/ldvm/util"
)

type TxDestroyTokenAccount struct {
	TxBase
}

func (tx *TxDestroyTokenAccount) SyntacticVerify() error {
	var err error
	errPrefix := "TxDestroyTokenAccount.SyntacticVerify failed:"
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

	if token := util.TokenSymbol(tx.ld.From); !token.Valid() {
		return fmt.Errorf("%s invalid token %s", errPrefix, token.GoString())
	}
	return nil
}

func (tx *TxDestroyTokenAccount) Verify(bctx BlockContext, bs BlockState) error {
	var err error
	errPrefix := "TxDestroyTokenAccount.Verify failed:"

	if err = tx.TxBase.Verify(bctx, bs); err != nil {
		return fmt.Errorf("%s %v", errPrefix, err)
	}
	if !tx.from.SatisfySigningPlus(tx.signers) {
		return fmt.Errorf("%s invalid signature for keepers", errPrefix)
	}
	if err = tx.from.CheckDestroyToken(tx.to); err != nil {
		return fmt.Errorf("%s %v", errPrefix, err)
	}
	return nil
}

func (tx *TxDestroyTokenAccount) Accept(bctx BlockContext, bs BlockState) error {
	if err := tx.TxBase.Accept(bctx, bs); err != nil {
		return err
	}
	// DestroyToken after TxBase.Accept
	tx.from.pledge.SetUint64(0)
	return tx.from.DestroyToken(tx.to)
}
