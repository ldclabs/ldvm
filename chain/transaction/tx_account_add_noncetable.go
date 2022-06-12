// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package transaction

import (
	"encoding/json"
	"fmt"

	"github.com/ldclabs/ldvm/ld"
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
		return nil, fmt.Errorf("TxAddAccountNonceTable.MarshalJSON failed: invalid tx.input")
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
	errPrefix := "TxAddAccountNonceTable.SyntacticVerify failed:"
	if err = tx.TxBase.SyntacticVerify(); err != nil {
		return fmt.Errorf("%s %v", errPrefix, err)
	}

	switch {
	case tx.ld.To != nil:
		return fmt.Errorf("%s invalid to, should be nil", errPrefix)

	case tx.ld.Token != nil:
		return fmt.Errorf("%s invalid token, should be nil", errPrefix)

	case tx.ld.Amount != nil:
		return fmt.Errorf("%s invalid amount, should be nil", errPrefix)

	case len(tx.ld.Data) == 0:
		return fmt.Errorf("%s invalid data", errPrefix)
	}

	tx.input = make([]uint64, 0)
	if err = ld.DecMode.Unmarshal(tx.ld.Data, &tx.input); err != nil {
		return fmt.Errorf("%s invalid data, %v", errPrefix, err)
	}
	switch {
	case len(tx.input) < 2:
		return fmt.Errorf("%s no nonce", errPrefix)

	case len(tx.input) > 1025:
		return fmt.Errorf("%s too many nonces, expected <= 1024, got %d",
			errPrefix, len(tx.input)-1)

	case tx.input[0] <= tx.ld.Timestamp:
		return fmt.Errorf("%s invalid expire time, expected > %d, got %d",
			errPrefix, tx.ld.Timestamp, tx.input[0])

	case tx.input[0] > (tx.ld.Timestamp + 3600*24*30):
		return fmt.Errorf("%s invalid expire time, expected <= %d, got %d",
			errPrefix, tx.ld.Timestamp+3600*24*30, tx.input[0])
	}
	return nil
}

func (tx *TxAddAccountNonceTable) Verify(bctx BlockContext, bs BlockState) error {
	var err error
	if err = tx.TxBase.Verify(bctx, bs); err != nil {
		return fmt.Errorf("TxAddAccountNonceTable.Verify failed: %v", err)
	}
	if err = tx.from.CheckNonceTable(tx.input[0], tx.input[1:]); err != nil {
		return fmt.Errorf("TxAddAccountNonceTable.Verify failed: %v", err)
	}
	return nil
}

func (tx *TxAddAccountNonceTable) Accept(bctx BlockContext, bs BlockState) error {
	var err error
	if err = tx.from.AddNonceTable(tx.input[0], tx.input[1:]); err != nil {
		return err
	}

	return tx.TxBase.Accept(bctx, bs)
}
