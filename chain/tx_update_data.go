// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package chain

import (
	"math/big"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ldclabs/ldvm/ld"
)

type TxUpdateData struct {
	ld *ld.Transaction
	id ids.ID
}

func (tx *TxUpdateData) ID() ids.ID {
	return tx.ld.ID()
}

func (tx *TxUpdateData) Type() ld.TxType {
	return tx.ld.Type
}

func (tx *TxUpdateData) Gas() *big.Int {
	return new(big.Int).SetUint64(tx.ld.Gas)
}

func (tx *TxUpdateData) Bytes() []byte {
	return tx.ld.Bytes()
}

func (tx *TxUpdateData) SyntacticVerify() error {
	return nil
}

func (tx *TxUpdateData) Verify(blk *Block) error {
	return nil
}

func (tx *TxUpdateData) Accept(blk *Block) error {
	return nil
}
