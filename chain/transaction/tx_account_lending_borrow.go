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

type TxBorrow struct {
	TxBase
	exSigners util.EthIDs
	input     *ld.TxTransfer
	dueTime   uint64
}

func (tx *TxBorrow) MarshalJSON() ([]byte, error) {
	if tx == nil || tx.ld == nil {
		return []byte("null"), nil
	}
	v := tx.ld.Copy()
	if tx.input == nil {
		return nil, fmt.Errorf("TxBorrow.MarshalJSON failed: invalid tx.input")
	}
	d, err := json.Marshal(tx.input)
	if err != nil {
		return nil, err
	}
	v.Data = d
	return json.Marshal(v)
}

func (tx *TxBorrow) SyntacticVerify() error {
	var err error
	errPrefix := "TxBorrow.SyntacticVerify failed:"
	if err = tx.TxBase.SyntacticVerify(); err != nil {
		return fmt.Errorf("%s %v", errPrefix, err)
	}

	switch {
	case tx.ld.To == nil:
		return fmt.Errorf("%s nil to as lender", errPrefix)

	case tx.ld.Amount != nil:
		return fmt.Errorf("%s invalid amount, should be nil", errPrefix)

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
	case tx.input.From == nil:
		return fmt.Errorf("%s nil from as lender", errPrefix)

	case *tx.input.From != *tx.ld.To:
		return fmt.Errorf("%s invalid to, expected %s, got %s", errPrefix, tx.input.From, tx.ld.To)

	case tx.input.To == nil:
		return fmt.Errorf("%s nil to as borrower", errPrefix)

	case *tx.input.To != tx.ld.From:
		return fmt.Errorf("%s invalid from, expected %s, got %s", errPrefix, tx.input.To, tx.ld.From)

	case tx.input.Token == nil && tx.token != constants.NativeToken:
		return fmt.Errorf("%s invalid token, expected %s, got %s",
			errPrefix, constants.NativeToken.GoString(), tx.token.GoString())

	case tx.input.Token != nil && tx.token != *tx.input.Token:
		return fmt.Errorf("%s invalid token, expected %s, got %s",
			errPrefix, tx.input.Token.GoString(), tx.token.GoString())

	case tx.input.Amount == nil || tx.input.Amount.Sign() <= 0:
		return fmt.Errorf("%s invalid amount, expected >= 1", errPrefix)

	case tx.input.Expire < tx.ld.Timestamp:
		return fmt.Errorf("%s data expired", errPrefix)
	}

	if len(tx.input.Data) > 0 {
		u := uint64(0)
		if err = ld.UnmarshalCBOR(tx.input.Data, &u); err != nil {
			return fmt.Errorf("%s invalid dueTime, %v", errPrefix, err)
		}
		if u <= tx.ld.Timestamp {
			return fmt.Errorf("%s invalid dueTime, expected > %d, got %d",
				errPrefix, tx.ld.Timestamp, u)
		}
		tx.dueTime = u
	}

	tx.exSigners, err = tx.ld.ExSigners()
	if err != nil {
		return fmt.Errorf("%s invalid exSignatures, %v", errPrefix, err)
	}
	return nil
}

func (tx *TxBorrow) Verify(bctx BlockContext, bs BlockState) error {
	var err error
	if err = tx.TxBase.Verify(bctx, bs); err != nil {
		return fmt.Errorf("TxBorrow.Verify failed: %v", err)
	}
	// verify lender's signatures
	if !tx.to.SatisfySigning(tx.exSigners) {
		return fmt.Errorf("TxBorrow.Verify failed: invalid exSignatures for lending keepers")
	}
	if err = tx.to.CheckBorrow(tx.token, tx.ld.From, tx.input.Amount, tx.dueTime); err != nil {
		return fmt.Errorf("TxBorrow.Verify failed: %v", err)
	}
	if err = tx.to.CheckSubByNonceTable(
		tx.token, tx.input.Expire, tx.input.Nonce, tx.input.Amount); err != nil {
		return fmt.Errorf("TxBorrow.Verify failed: %v", err)
	}
	return nil
}

func (tx *TxBorrow) Accept(bctx BlockContext, bs BlockState) error {
	var err error
	if err = tx.to.Borrow(
		tx.token, tx.ld.From, tx.input.Amount, tx.dueTime); err != nil {
		return err
	}
	if err = tx.to.SubByNonceTable(
		tx.token, tx.input.Expire, tx.input.Nonce, tx.input.Amount); err != nil {
		return err
	}
	if err = tx.from.Add(tx.token, tx.input.Amount); err != nil {
		return err
	}
	return tx.TxBase.Accept(bctx, bs)
}
