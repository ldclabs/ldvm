// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package txn

import (
	"encoding/json"

	"github.com/ldclabs/ldvm/ids"
	"github.com/ldclabs/ldvm/ld"
	"github.com/ldclabs/ldvm/util/erring"
)

type TxTransferCash struct {
	TxBase
	input *ld.TxTransfer
}

func (tx *TxTransferCash) MarshalJSON() ([]byte, error) {
	if tx == nil || tx.ld == nil {
		return []byte("null"), nil
	}

	v := tx.ld.Copy()
	errp := erring.ErrPrefix("txn.TxTransferCash.MarshalJSON: ")
	if tx.input == nil {
		return nil, errp.Errorf("nil tx.input")
	}
	d, err := json.Marshal(tx.input)
	if err != nil {
		return nil, errp.ErrorIf(err)
	}
	v.Tx.Data = d
	return errp.ErrorMap(json.Marshal(v))
}

func (tx *TxTransferCash) SyntacticVerify() error {
	var err error
	errp := erring.ErrPrefix("txn.TxTransferCash.SyntacticVerify: ")
	if err = tx.TxBase.SyntacticVerify(); err != nil {
		return errp.ErrorIf(err)
	}

	switch {
	case tx.ld.Tx.To == nil:
		return errp.Errorf("invalid to")

	case tx.ld.Tx.Amount != nil:
		return errp.Errorf("invalid amount, should be nil")

	case len(tx.ld.Tx.Data) == 0:
		return errp.Errorf("invalid data")

	case len(tx.ld.ExSignatures) == 0:
		return errp.Errorf("no exSignatures")
	}

	tx.input = &ld.TxTransfer{}
	if err = tx.input.Unmarshal(tx.ld.Tx.Data); err != nil {
		return errp.ErrorIf(err)
	}
	if err = tx.input.SyntacticVerify(); err != nil {
		return errp.ErrorIf(err)
	}

	switch {
	case tx.input.From == nil:
		return errp.Errorf("nil issuer")

	case *tx.input.From != *tx.ld.Tx.To:
		return errp.Errorf("invalid issuer, expected %s, got %s",
			tx.input.From, tx.ld.Tx.To)

	case tx.input.To == nil:
		return errp.Errorf("nil recipient")

	case *tx.input.To != tx.ld.Tx.From:
		return errp.Errorf("invalid recipient, expected %s, got %s",
			tx.input.To, tx.ld.Tx.From)

	case tx.input.Token == nil && tx.token != ids.NativeToken:
		return errp.Errorf("invalid token, expected %s, got %s",
			ids.NativeToken.GoString(), tx.token.GoString())

	case tx.input.Token != nil && tx.token != *tx.input.Token:
		return errp.Errorf("invalid token, expected %s, got %s",
			tx.input.Token.GoString(), tx.token.GoString())

	case tx.input.Amount == nil || tx.input.Amount.Sign() <= 0:
		return errp.Errorf("invalid amount, expected >= 1")

	case tx.input.Expire < tx.ld.Timestamp:
		return errp.Errorf("data expired")
	}

	return nil
}

func (tx *TxTransferCash) Apply(ctx ChainContext, cs ChainState) error {
	var err error
	errp := erring.ErrPrefix("txn.TxTransferCash.Apply: ")

	if err = tx.TxBase.verify(ctx, cs); err != nil {
		return errp.ErrorIf(err)
	}
	// verify issuer's signatures
	if !tx.to.Verify(tx.ld.ExHash(), tx.ld.ExSignatures, tx.to.IDKey()) {
		return errp.Errorf("invalid signature for issuer")
	}

	if err = tx.to.SubByNonceTable(
		tx.token, tx.input.Expire, tx.input.Nonce, tx.input.Amount); err != nil {
		return errp.ErrorIf(err)
	}
	if err = tx.from.Add(tx.token, tx.input.Amount); err != nil {
		return errp.ErrorIf(err)
	}
	return errp.ErrorIf(tx.TxBase.accept(ctx, cs))
}
