// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package transaction

import (
	"github.com/ldclabs/ldvm/ld"
	"github.com/ldclabs/ldvm/util"
)

type TxEth struct {
	TxBase
	input *ld.TxEth
}

func (tx *TxEth) SyntacticVerify() error {
	var err error
	errp := util.ErrPrefix("TxEth.SyntacticVerify error: ")

	if err = tx.TxBase.SyntacticVerify(); err != nil {
		return errp.ErrorIf(err)
	}

	switch {
	case tx.ld.To == nil:
		return errp.Errorf("invalid to")
	case tx.ld.Amount == nil:
		return errp.Errorf("invalid amount")
	case len(tx.ld.Data) == 0:
		return errp.Errorf("invalid data")
	}

	tx.input = &ld.TxEth{}
	if err = tx.input.Unmarshal(tx.ld.Data); err != nil {
		return errp.ErrorIf(err)
	}
	if err = tx.input.SyntacticVerify(); err != nil {
		return errp.ErrorIf(err)
	}

	txData := tx.input.TxData(nil)
	switch {
	case tx.ld.ChainID != txData.ChainID:
		return errp.Errorf("invalid chainID")

	case tx.ld.Nonce != txData.Nonce:
		return errp.Errorf("invalid nonce")

	case tx.ld.GasTip != txData.GasTip:
		return errp.Errorf("invalid gasTip")

	case tx.ld.GasFeeCap != txData.GasFeeCap:
		return errp.Errorf("invalid gasFeeCap")

	case tx.ld.From != txData.From:
		return errp.Errorf("invalid from")

	case *tx.ld.To != *txData.To:
		return errp.Errorf("invalid to")

	case tx.ld.Token != nil:
		return errp.Errorf("invalid token")

	case tx.ld.Amount.Cmp(txData.Amount) != 0:
		return errp.Errorf("invalid amount")

	case len(tx.ld.Signatures) != 1 || tx.ld.Signatures[0] != txData.Signatures[0]:
		return errp.Errorf("invalid signatures")

	case tx.ld.ExSignatures != nil:
		return errp.Errorf("invalid exSignatures")
	}

	tx.signers = util.EthIDs{tx.ld.From}
	return nil
}
