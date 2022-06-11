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
		return nil, fmt.Errorf("TxTransferPay.MarshalJSON failed: invalid tx.input")
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
	errPrefix := "TxTransferPay.SyntacticVerify failed:"
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

	tx.input = &ld.TxTransfer{}
	if err = tx.input.Unmarshal(tx.ld.Data); err != nil {
		return fmt.Errorf("%s %v", errPrefix, err)
	}

	if err = tx.input.SyntacticVerify(); err != nil {
		return fmt.Errorf("%s %v", errPrefix, err)
	}

	switch {
	case tx.input.From != nil && *tx.input.From != tx.ld.From:
		return fmt.Errorf("%s invalid sender, expected %s, got %s",
			errPrefix, *tx.input.From, tx.ld.From)

	case tx.input.To == nil:
		return fmt.Errorf("%s nil recipient", errPrefix)

	case *tx.input.To != *tx.ld.To:
		return fmt.Errorf("%s invalid recipient, expected %s, got %s",
			errPrefix, tx.input.To, *tx.ld.To)

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

	case tx.input.Expire > 0 && tx.input.Expire < tx.ld.Timestamp:
		return fmt.Errorf("%s data expired", errPrefix)
	}

	tx.exSigners, err = tx.ld.ExSigners()
	if err != nil {
		return fmt.Errorf("%s invalid exSignatures: %v", errPrefix, err)
	}
	return nil
}

func (tx *TxTransferPay) Verify(bctx BlockContext, bs BlockState) error {
	var err error
	if err = tx.TxBase.Verify(bctx, bs); err != nil {
		return fmt.Errorf("TxTransferPay.Verify failed: %v", err)
	}

	if !tx.to.SatisfySigning(tx.exSigners) {
		return fmt.Errorf("TxTransferPay.Verify failed: invalid exSignatures for recipient")
	}
	return nil
}
