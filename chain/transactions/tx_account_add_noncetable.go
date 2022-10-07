// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package transactions

import (
	"encoding/json"

	"github.com/ldclabs/ldvm/util"
)

type TxAddNonceTable struct {
	TxBase
	input []uint64
}

func (tx *TxAddNonceTable) MarshalJSON() ([]byte, error) {
	if tx == nil || tx.ld == nil {
		return []byte("null"), nil
	}

	v := tx.ld.Copy()
	errp := util.ErrPrefix("transactions.TxAddNonceTable.MarshalJSON: ")
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

// ApplyGenesis skipping signature verification
func (tx *TxAddNonceTable) SyntacticVerify() error {
	var err error
	errp := util.ErrPrefix("transactions.TxAddNonceTable.SyntacticVerify: ")

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

	tx.input = make([]uint64, 0)
	if err = util.UnmarshalCBOR(tx.ld.Tx.Data, &tx.input); err != nil {
		return errp.Errorf("invalid data, %v", err)
	}
	switch {
	case len(tx.input) < 2:
		return errp.Errorf("no nonce")

	case len(tx.input) > 1025:
		return errp.Errorf("too many nonces, expected <= 1024, got %d", len(tx.input)-1)

	case tx.input[0] <= tx.ld.Timestamp:
		return errp.Errorf("invalid expire time, expected > %d, got %d",
			tx.ld.Timestamp, tx.input[0])

	case tx.input[0] > (tx.ld.Timestamp + 3600*24*30):
		return errp.Errorf("invalid expire time, expected <= %d, got %d",
			tx.ld.Timestamp+3600*24*30, tx.input[0])
	}
	return nil
}

func (tx *TxAddNonceTable) Apply(ctx ChainContext, cs ChainState) error {
	var err error
	errp := util.ErrPrefix("transactions.TxAddNonceTable.Apply: ")

	if err = tx.TxBase.verify(ctx, cs); err != nil {
		return errp.ErrorIf(err)
	}

	if err = tx.from.AddNonceTable(tx.input[0], tx.input[1:]); err != nil {
		return errp.ErrorIf(err)
	}
	return errp.ErrorIf(tx.TxBase.accept(ctx, cs))
}
