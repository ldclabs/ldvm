// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package txn

import (
	"github.com/ldclabs/ldvm/ld"
	"github.com/ldclabs/ldvm/util/erring"
)

type TxEth struct {
	TxBase
	input *ld.TxEth
}

func (tx *TxEth) SyntacticVerify() error {
	var err error
	errp := erring.ErrPrefix("txn.TxEth.SyntacticVerify: ")

	if err = tx.TxBase.SyntacticVerify(); err != nil {
		return errp.ErrorIf(err)
	}

	switch {
	case tx.ld.Tx.To == nil:
		return errp.Errorf("invalid to")
	case tx.ld.Tx.Amount == nil:
		return errp.Errorf("invalid amount")
	case len(tx.ld.Tx.Data) == 0:
		return errp.Errorf("invalid data")
	}

	tx.input = &ld.TxEth{}
	if err = tx.input.Unmarshal(tx.ld.Tx.Data); err != nil {
		return errp.ErrorIf(err)
	}
	if err = tx.input.SyntacticVerify(); err != nil {
		return errp.ErrorIf(err)
	}

	etx := tx.input.ToTransaction()
	switch {
	case tx.ld.Tx.ChainID != etx.Tx.ChainID:
		return errp.Errorf("invalid chainID")

	case tx.ld.Tx.Nonce != etx.Tx.Nonce:
		return errp.Errorf("invalid nonce")

	case tx.ld.Tx.GasTip != etx.Tx.GasTip:
		return errp.Errorf("invalid gasTip")

	case tx.ld.Tx.GasFeeCap != etx.Tx.GasFeeCap:
		return errp.Errorf("invalid gasFeeCap")

	case tx.ld.Tx.From != etx.Tx.From:
		return errp.Errorf("invalid from")

	case *tx.ld.Tx.To != *etx.Tx.To:
		return errp.Errorf("invalid to")

	case tx.ld.Tx.Token != nil:
		return errp.Errorf("invalid token")

	case tx.ld.Tx.Amount.Cmp(etx.Tx.Amount) != 0:
		return errp.Errorf("invalid amount")

	case len(tx.ld.Signatures) != 1 || len(etx.Signatures) != 1 || !tx.ld.Signatures[0].Equal(etx.Signatures[0]):
		return errp.Errorf("invalid signatures")

	case tx.ld.ExSignatures != nil:
		return errp.Errorf("invalid exSignatures")
	}

	return nil
}
