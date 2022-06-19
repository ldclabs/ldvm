// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package transaction

import (
	"encoding/json"
	"fmt"

	"github.com/ldclabs/ldvm/util"
)

type TxAddAccountNonceTable struct {
	TxBase
	input []uint64
}

func (tx *TxAddAccountNonceTable) MarshalJSON() ([]byte, error) {
	if tx == nil || tx.ld == nil {
		return []byte("null"), nil
	}
	v := tx.ld.Copy()
	if tx.input == nil {
		return nil, fmt.Errorf("TxAddAccountNonceTable.MarshalJSON error: invalid tx.input")
	}
	d, err := json.Marshal(tx.input)
	if err != nil {
		return nil, err
	}
	v.Data = d
	return json.Marshal(v)
}

// VerifyGenesis skipping signature verification
func (tx *TxAddAccountNonceTable) SyntacticVerify() error {
	var err error
	errp := util.ErrPrefix("TxAddAccountNonceTable.SyntacticVerify error: ")

	if err = tx.TxBase.SyntacticVerify(); err != nil {
		return errp.ErrorIf(err)
	}

	switch {
	case tx.ld.To != nil:
		return errp.Errorf("invalid to, should be nil")

	case tx.ld.Token != nil:
		return errp.Errorf("invalid token, should be nil")

	case tx.ld.Amount != nil:
		return errp.Errorf("invalid amount, should be nil")

	case len(tx.ld.Data) == 0:
		return errp.Errorf("invalid data")
	}

	tx.input = make([]uint64, 0)
	if err = util.UnmarshalCBOR(tx.ld.Data, &tx.input); err != nil {
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

func (tx *TxAddAccountNonceTable) Verify(bctx BlockContext, bs BlockState) error {
	var err error
	errp := util.ErrPrefix("TxAddAccountNonceTable.Verify error: ")

	if err = tx.TxBase.Verify(bctx, bs); err != nil {
		return errp.ErrorIf(err)
	}
	if err = tx.from.CheckNonceTable(tx.input[0], tx.input[1:]); err != nil {
		return errp.ErrorIf(err)
	}
	return nil
}

func (tx *TxAddAccountNonceTable) Accept(bctx BlockContext, bs BlockState) error {
	var err error
	errp := util.ErrPrefix("TxAddAccountNonceTable.Accept error: ")

	if err = tx.from.AddNonceTable(tx.input[0], tx.input[1:]); err != nil {
		return errp.ErrorIf(err)
	}

	return errp.ErrorIf(tx.TxBase.Accept(bctx, bs))
}
