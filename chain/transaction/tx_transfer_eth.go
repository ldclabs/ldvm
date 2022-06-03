// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package transaction

import (
	"fmt"

	"github.com/ldclabs/ldvm/ld"
	"github.com/ldclabs/ldvm/util"
)

type TxEth struct {
	TxBase
	input *ld.TxEth
}

func (tx *TxEth) SyntacticVerify() error {
	var err error
	if err = tx.TxBase.SyntacticVerify(); err != nil {
		return err
	}

	switch {
	case tx.ld.To == nil:
		return fmt.Errorf("TxEth.SyntacticVerify failed: invalid to")
	case tx.ld.Amount == nil:
		return fmt.Errorf("TxEth.SyntacticVerify failed: invalid amount")
	case len(tx.ld.Data) == 0:
		return fmt.Errorf("TxEth.SyntacticVerify failed: invalid data")
	}

	tx.input = &ld.TxEth{}
	if err = tx.input.Unmarshal(tx.ld.Data); err != nil {
		return fmt.Errorf("TxEth.SyntacticVerify failed: %v", err)
	}
	if err = tx.input.SyntacticVerify(); err != nil {
		return fmt.Errorf("TxEth.SyntacticVerify failed: %v", err)
	}

	txData := tx.input.TxData(nil)
	switch {
	case tx.ld.ChainID != txData.ChainID:
		return fmt.Errorf("TxEth.SyntacticVerify failed: invalid chainID")
	case tx.ld.Nonce != txData.Nonce:
		return fmt.Errorf("TxEth.SyntacticVerify failed: invalid nonce")
	case tx.ld.GasTip != txData.GasTip:
		return fmt.Errorf("TxEth.SyntacticVerify failed: invalid gasTip")
	case tx.ld.GasFeeCap != txData.GasFeeCap:
		return fmt.Errorf("TxEth.SyntacticVerify failed: invalid gasFeeCap")
	case tx.ld.From != txData.From:
		return fmt.Errorf("TxEth.SyntacticVerify failed: invalid from")
	case *tx.ld.To != *txData.To:
		return fmt.Errorf("TxEth.SyntacticVerify failed: invalid to")
	case tx.ld.Token != nil:
		return fmt.Errorf("TxEth.SyntacticVerify failed: invalid token")
	case tx.ld.Amount.Cmp(txData.Amount) != 0:
		return fmt.Errorf("TxEth.SyntacticVerify failed: invalid amount")
	case len(tx.ld.Signatures) != 1 || tx.ld.Signatures[0] != txData.Signatures[0]:
		return fmt.Errorf("TxEth.SyntacticVerify failed: invalid signatures")
	case tx.ld.ExSignatures != nil:
		return fmt.Errorf("TxEth.SyntacticVerify failed: invalid exSignatures")
	}

	tx.signers = []util.EthID{tx.ld.From}
	return nil
}
