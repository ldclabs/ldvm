// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package transaction

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/ldclabs/ldvm/ld"
)

type TxAddAccountNonceTable struct {
	TxBase
	data []uint64
}

func (tx *TxAddAccountNonceTable) MarshalJSON() ([]byte, error) {
	if tx == nil || tx.ld == nil {
		return []byte("null"), nil
	}
	v := tx.ld.Copy()
	if tx.data == nil {
		return nil, fmt.Errorf("MarshalJSON failed: data not exists")
	}
	d, err := json.Marshal(tx.data)
	if err != nil {
		return nil, err
	}
	v.Data = d
	return json.Marshal(v)
}

// VerifyGenesis skipping signature verification
func (tx *TxAddAccountNonceTable) SyntacticVerify() error {
	var err error
	if err = tx.TxBase.SyntacticVerify(); err != nil {
		return err
	}

	if tx.ld.Token != nil {
		return fmt.Errorf("invalid token, expected NativeToken, got %s",
			strconv.Quote(tx.ld.Token.GoString()))
	}
	if tx.ld.To != nil {
		return fmt.Errorf("TxAddAccountNonceTable invalid to")
	}
	if tx.ld.Amount != nil {
		return fmt.Errorf("TxAddAccountNonceTable invalid amount")
	}
	if len(tx.ld.Data) == 0 {
		return fmt.Errorf("TxAddAccountNonceTable invalid")
	}
	tx.data = make([]uint64, 0)
	if err = ld.DecMode.Unmarshal(tx.ld.Data, &tx.data); err != nil {
		return fmt.Errorf("TxAddAccountNonceTable unmarshal data failed: %v", err)
	}
	if len(tx.data) < 2 {
		return fmt.Errorf("TxAddAccountNonceTable numbers empty")
	}
	if len(tx.data) > 1025 {
		return fmt.Errorf("TxAddAccountNonceTable too many numbers")
	}
	if tx.data[0] < tx.ld.Timestamp || tx.data[0] > (tx.ld.Timestamp+3600*24*7) {
		return fmt.Errorf("TxAddAccountNonceTable invalid expire")
	}
	return nil
}

func (tx *TxAddAccountNonceTable) Verify(bctx BlockContext, bs BlockState) error {
	var err error
	if err = tx.TxBase.Verify(bctx, bs); err != nil {
		return err
	}
	if err = tx.from.CheckNonceTable(tx.data[0], tx.data[1:]); err != nil {
		return err
	}
	return nil
}

func (tx *TxAddAccountNonceTable) Accept(bctx BlockContext, bs BlockState) error {
	var err error
	if err = tx.from.AddNonceTable(tx.data[0], tx.data[1:]); err != nil {
		return err
	}

	return tx.TxBase.Accept(bctx, bs)
}