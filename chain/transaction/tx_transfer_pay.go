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

type TxTransferPay struct {
	TxBase
	exSigners util.EthIDs
	input     *ld.TxTransfer
}

func (tx *TxTransferPay) MarshalJSON() ([]byte, error) {
	if tx == nil || tx.ld == nil {
		return []byte("null"), nil
	}
	v := tx.ld.Copy()
	if tx.input == nil {
		return nil, fmt.Errorf("TxTransferPay.MarshalJSON error: invalid tx.input")
	}
	d, err := json.Marshal(tx.input)
	if err != nil {
		return nil, err
	}
	v.Data = d
	return json.Marshal(v)
}

func (tx *TxTransferPay) SyntacticVerify() error {
	var err error
	errp := util.ErrPrefix("TxTransferPay.SyntacticVerify error: ")

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

	tx.input = &ld.TxTransfer{}
	if err = tx.input.Unmarshal(tx.ld.Data); err != nil {
		return errp.ErrorIf(err)
	}

	if err = tx.input.SyntacticVerify(); err != nil {
		return errp.ErrorIf(err)
	}

	switch {
	case tx.input.From != nil && *tx.input.From != tx.ld.From:
		return errp.Errorf("invalid sender, expected %s, got %s",
			*tx.input.From, tx.ld.From)

	case tx.input.To == nil:
		return errp.Errorf("nil recipient")

	case *tx.input.To != *tx.ld.To:
		return errp.Errorf("invalid recipient, expected %s, got %s",
			tx.input.To, *tx.ld.To)

	case tx.input.Token == nil && tx.token != constants.NativeToken:
		return errp.Errorf("invalid token, expected %s, got %s",
			constants.NativeToken.GoString(), tx.token.GoString())

	case tx.input.Token != nil && tx.token != *tx.input.Token:
		return errp.Errorf("invalid token, expected %s, got %s",
			tx.input.Token.GoString(), tx.token.GoString())

	case tx.input.Amount == nil:
		return errp.Errorf("nil amount")

	case tx.input.Amount.Cmp(tx.ld.Amount) != 0:
		return errp.Errorf("invalid amount, expected %v, got %v",
			tx.input.Amount, tx.ld.Amount)

	case tx.input.Expire > 0 && tx.input.Expire < tx.ld.Timestamp:
		return errp.Errorf("data expired")
	}

	tx.exSigners, err = tx.ld.ExSigners()
	if err != nil {
		return errp.Errorf("invalid exSignatures, %v", err)
	}
	return nil
}

func (tx *TxTransferPay) Verify(bctx BlockContext, bs BlockState) error {
	var err error
	errp := util.ErrPrefix("TxTransferPay.Verify error: ")

	if err = tx.TxBase.Verify(bctx, bs); err != nil {
		return errp.ErrorIf(err)
	}

	if !tx.to.SatisfySigning(tx.exSigners) {
		return errp.Errorf("invalid exSignatures for recipient")
	}
	return nil
}
