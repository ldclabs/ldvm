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
	errPrefix := "TxEth.SyntacticVerify failed:"
	if err = tx.TxBase.SyntacticVerify(); err != nil {
		return fmt.Errorf("%s %v", errPrefix, err)
	}

	switch {
	case tx.ld.To == nil:
		return fmt.Errorf("%s invalid to", errPrefix)
	case tx.ld.Amount == nil:
		return fmt.Errorf("%s invalid amount", errPrefix)
	case len(tx.ld.Data) == 0:
		return fmt.Errorf("%s invalid data", errPrefix)
	}

	tx.input = &ld.TxEth{}
	if err = tx.input.Unmarshal(tx.ld.Data); err != nil {
		return fmt.Errorf("%s %v", errPrefix, err)
	}
	if err = tx.input.SyntacticVerify(); err != nil {
		return fmt.Errorf("%s %v", errPrefix, err)
	}

	txData := tx.input.TxData(nil)
	switch {
	case tx.ld.ChainID != txData.ChainID:
		return fmt.Errorf("%s invalid chainID", errPrefix)

	case tx.ld.Nonce != txData.Nonce:
		return fmt.Errorf("%s invalid nonce", errPrefix)

	case tx.ld.GasTip != txData.GasTip:
		return fmt.Errorf("%s invalid gasTip", errPrefix)

	case tx.ld.GasFeeCap != txData.GasFeeCap:
		return fmt.Errorf("%s invalid gasFeeCap", errPrefix)

	case tx.ld.From != txData.From:
		return fmt.Errorf("%s invalid from", errPrefix)

	case *tx.ld.To != *txData.To:
		return fmt.Errorf("%s invalid to", errPrefix)

	case tx.ld.Token != nil:
		return fmt.Errorf("%s invalid token", errPrefix)

	case tx.ld.Amount.Cmp(txData.Amount) != 0:
		return fmt.Errorf("%s invalid amount", errPrefix)

	case len(tx.ld.Signatures) != 1 || tx.ld.Signatures[0] != txData.Signatures[0]:
		return fmt.Errorf("%s invalid signatures", errPrefix)

	case tx.ld.ExSignatures != nil:
		return fmt.Errorf("%s invalid exSignatures", errPrefix)
	}

	tx.signers = util.EthIDs{tx.ld.From}
	return nil
}
