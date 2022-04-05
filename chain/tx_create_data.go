// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package chain

import (
	"math/big"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ldclabs/ldvm/ld"
)

type TxCreateData struct {
	ld *ld.Transaction
	id ids.ID
}

func (tx *TxCreateData) ID() ids.ID {
	return tx.ld.ID()
}

func (tx *TxCreateData) Type() ld.TxType {
	return tx.ld.Type
}

func (tx *TxCreateData) Gas() *big.Int {
	return new(big.Int).SetUint64(tx.ld.Gas)
}

func (tx *TxCreateData) Bytes() []byte {
	return tx.ld.Bytes()
}

func (tx *TxCreateData) SyntacticVerify() error {
	return nil
}

func (tx *TxCreateData) Verify(blk *Block) error {
	return nil
}

func (tx *TxCreateData) Accept(blk *Block) error {
	return nil
}
