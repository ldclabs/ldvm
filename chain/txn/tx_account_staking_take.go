// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package txn

import (
	"encoding/json"

	"github.com/ldclabs/ldvm/ids"
	"github.com/ldclabs/ldvm/ld"
	"github.com/ldclabs/ldvm/util/encoding"
	"github.com/ldclabs/ldvm/util/erring"
)

type TxTakeStake struct {
	TxBase
	input    *ld.TxTransfer
	lockTime uint64
}

func (tx *TxTakeStake) MarshalJSON() ([]byte, error) {
	if tx == nil || tx.ld == nil {
		return []byte("null"), nil
	}

	v := tx.ld.Copy()
	errp := erring.ErrPrefix("txn.TxTakeStake.MarshalJSON: ")
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

func (tx *TxTakeStake) SyntacticVerify() error {
	var err error
	errp := erring.ErrPrefix("txn.TxTakeStake.SyntacticVerify: ")

	if err = tx.TxBase.SyntacticVerify(); err != nil {
		return errp.ErrorIf(err)
	}

	switch {
	case tx.ld.Tx.To == nil:
		return errp.Errorf("nil to as stake account")

	case tx.ld.Tx.Amount == nil:
		return errp.Errorf("nil amount")

	case len(tx.ld.Tx.Data) == 0:
		return errp.Errorf("invalid data")

	case len(tx.ld.ExSignatures) == 0:
		return errp.Errorf("no exSignatures")
	}

	if stake := ids.StakeSymbol(*tx.ld.Tx.To); !stake.Valid() {
		return errp.Errorf("invalid stake account %s", stake.GoString())
	}

	tx.input = &ld.TxTransfer{}
	if err = tx.input.Unmarshal(tx.ld.Tx.Data); err != nil {
		return errp.ErrorIf(err)
	}
	if err = tx.input.SyntacticVerify(); err != nil {
		return errp.ErrorIf(err)
	}

	switch {
	case tx.input.Nonce != tx.ld.Tx.Nonce:
		return errp.Errorf("invalid nonce, expected %d, got %d",
			tx.input.Nonce, tx.ld.Tx.Nonce)

	case tx.input.From == nil:
		return errp.Errorf("nil from")

	case *tx.input.From != tx.ld.Tx.From:
		return errp.Errorf("invalid from, expected %s, got %s",
			tx.input.From, tx.ld.Tx.From)

	case tx.input.To == nil:
		return errp.Errorf("nil to")

	case *tx.input.To != *tx.ld.Tx.To:
		return errp.Errorf("invalid to, expected %s, got %s",
			tx.input.To, tx.ld.Tx.To)

	case tx.input.Token == nil && tx.token != ids.NativeToken:
		return errp.Errorf("invalid token, expected %s, got %s",
			ids.NativeToken.GoString(), tx.token.GoString())

	case tx.input.Token != nil && tx.token != *tx.input.Token:
		return errp.Errorf("invalid token, expected %s, got %s",
			tx.input.Token.GoString(), tx.token.GoString())

	case tx.input.Amount == nil:
		return errp.Errorf("nil amount")

	case tx.input.Amount.Cmp(tx.ld.Tx.Amount) != 0:
		return errp.Errorf("invalid amount, expected %v, got %v",
			tx.input.Amount, tx.ld.Tx.Amount)

	case tx.input.Expire < tx.ld.Timestamp:
		return errp.Errorf("data expired, expected >= %d, got %v", tx.ld.Timestamp, tx.input.Expire)
	}

	if len(tx.input.Data) > 0 {
		u := uint64(0)
		if err = encoding.UnmarshalCBOR(tx.input.Data, &u); err != nil {
			return errp.Errorf("invalid lockTime, %v", err)
		}
		tx.lockTime = u
	}

	return nil
}

func (tx *TxTakeStake) Apply(ctx ChainContext, cs ChainState) error {
	var err error
	errp := erring.ErrPrefix("txn.TxTakeStake.Apply: ")

	if err = tx.TxBase.verify(ctx, cs); err != nil {
		return errp.ErrorIf(err)
	}

	if err = cs.LoadLedger(tx.to); err != nil {
		return errp.ErrorIf(err)
	}

	// must TakeStake and then Accept
	if err = tx.to.TakeStake(tx.token, tx.ld.Tx.From, tx.ld.Tx.Amount, tx.lockTime); err != nil {
		return errp.ErrorIf(err)
	}
	if !tx.to.Verify(tx.ld.ExHash(), tx.ld.ExSignatures, tx.to.IDKey()) {
		return errp.Errorf("invalid exSignatures for stake keepers")
	}
	return errp.ErrorIf(tx.TxBase.accept(ctx, cs))
}
