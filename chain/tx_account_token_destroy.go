// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package chain

import (
	"fmt"
	"strconv"

	"github.com/ldclabs/ldvm/util"
)

type TxDestroyTokenAccount struct {
	TxBase
}

func (tx *TxDestroyTokenAccount) SyntacticVerify() error {
	var err error
	if err = tx.TxBase.SyntacticVerify(); err != nil {
		return err
	}
	if tx.ld.Token != nil {
		return fmt.Errorf("invalid token, expected NativeToken, got %s",
			strconv.Quote(tx.ld.Token.GoString()))
	}
	if tx.ld.To == nil {
		return fmt.Errorf("TxDestroyTokenAccount invalid recipient")
	}
	if token := util.TokenSymbol(tx.ld.From); !token.Valid() {
		return fmt.Errorf("TxDestroyTokenAccount invalid token: %s", token.GoString())
	}
	if tx.ld.Amount == nil || tx.ld.Amount.Sign() != 0 {
		return fmt.Errorf("TxCreateTokenAccount invalid amount")
	}
	return nil
}

func (tx *TxDestroyTokenAccount) Verify(blk *Block, bs BlockState) error {
	var err error
	if err = tx.TxBase.Verify(blk, bs); err != nil {
		return err
	}
	if !tx.from.SatisfySigningPlus(tx.signers) {
		return fmt.Errorf("sender account need more signers")
	}
	return tx.from.CheckDestroyToken(util.TokenSymbol(tx.ld.From), tx.to)
}

func (tx *TxDestroyTokenAccount) Accept(blk *Block, bs BlockState) error {
	if err := tx.TxBase.Accept(blk, bs); err != nil {
		return err
	}
	// DestroyToken after TxBase.Accept
	return tx.from.DestroyToken(util.TokenSymbol(tx.ld.From), tx.to)
}
