// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package txn

import (
	"encoding/json"

	"github.com/ldclabs/ldvm/ids"
	"github.com/ldclabs/ldvm/ld"
	"github.com/ldclabs/ldvm/util/erring"
)

type TxResetStake struct {
	TxBase
	input *ld.StakeConfig
}

func (tx *TxResetStake) MarshalJSON() ([]byte, error) {
	if tx == nil || tx.ld == nil {
		return []byte("null"), nil
	}

	v := tx.ld.Copy()
	errp := erring.ErrPrefix("txn.TxResetStake.MarshalJSON: ")
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

func (tx *TxResetStake) SyntacticVerify() error {
	var err error
	errp := erring.ErrPrefix("txn.TxResetStake.SyntacticVerify: ")

	if err = tx.TxBase.SyntacticVerify(); err != nil {
		return errp.ErrorIf(err)
	}

	switch {
	case tx.ld.Tx.To != nil:
		return errp.Errorf("invalid to, should be nil")

	case tx.ld.Tx.Token != nil:
		return errp.Errorf("invalid token, should be nil")

	case tx.ld.Tx.Amount != nil:
		return errp.Errorf("invalid amount, should be nil")

	case len(tx.ld.Tx.Data) == 0:
		return errp.Errorf("invalid data")
	}

	if stake := ids.StakeSymbol(tx.ld.Tx.From); !stake.Valid() {
		return errp.Errorf("invalid stake account %s", stake.GoString())
	}

	tx.input = &ld.StakeConfig{}
	if err = tx.input.Unmarshal(tx.ld.Tx.Data); err != nil {
		return errp.ErrorIf(err)
	}
	if err = tx.input.SyntacticVerify(); err != nil {
		return errp.ErrorIf(err)
	}

	if tx.input.LockTime > 0 && tx.input.LockTime <= tx.ld.Timestamp {
		return errp.Errorf("invalid lockTime, expected > %d, got %d",
			tx.ld.Timestamp, tx.input.LockTime)
	}
	return nil
}

func (tx *TxResetStake) Apply(ctx ChainContext, cs ChainState) error {
	var err error
	errp := erring.ErrPrefix("txn.TxResetStake.Apply: ")

	if err = tx.TxBase.verify(ctx, cs); err != nil {
		return errp.ErrorIf(err)
	}

	if !tx.from.VerifyPlus(tx.ld.TxHash(), tx.ld.Signatures, nil) {
		return errp.Errorf("invalid signatures for stake keepers")
	}

	if err = cs.LoadLedger(tx.from); err != nil {
		return errp.ErrorIf(err)
	}

	if err := tx.TxBase.accept(ctx, cs); err != nil {
		return errp.ErrorIf(err)
	}
	// do it after TxBase.Accept
	return errp.ErrorIf(tx.from.ResetStake(tx.input))
}
