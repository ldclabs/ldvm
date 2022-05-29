// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package transaction

import (
	"encoding/json"
	"fmt"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/snow/choices"
	"github.com/ldclabs/ldvm/ld"
)

type TxTest struct {
	ld     *ld.Transaction
	status choices.Status
}

func (tx *TxTest) MarshalJSON() ([]byte, error) {
	if tx == nil || tx.ld == nil {
		return []byte("null"), nil
	}

	return json.Marshal(tx.ld)
}

func (tx *TxTest) LD() *ld.Transaction {
	return tx.ld
}

func (tx *TxTest) ID() ids.ID {
	return tx.ld.ID
}

func (tx *TxTest) Type() ld.TxType {
	return tx.ld.Type
}

func (tx *TxTest) Bytes() []byte {
	return tx.ld.Bytes()
}

func (tx *TxTest) Status() string {
	return tx.status.String()
}

func (tx *TxTest) SetStatus(s choices.Status) {
	tx.status = s
}

func (tx *TxTest) SyntacticVerify() error {
	if tx == nil || tx.ld == nil {
		return fmt.Errorf("tx is nil")
	}
	return nil
}

// call after SyntacticVerify
func (tx *TxTest) Verify(bctx BlockContext, bs BlockState) error {
	return fmt.Errorf("TODO: not implemented")
}

func (tx *TxTest) Accept(bctx BlockContext, bs BlockState) error {
	return nil
}
