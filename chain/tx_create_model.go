// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package chain

import (
	"math/big"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ldclabs/ldvm/ld"
)

type TxCreateModel struct {
	ld *ld.Transaction
	id ids.ID
}

func (tx *TxCreateModel) ID() ids.ID {
	return tx.ld.ID()
}

func (tx *TxCreateModel) Type() ld.TxType {
	return tx.ld.Type
}

func (tx *TxCreateModel) Gas() *big.Int {
	return new(big.Int).SetUint64(tx.ld.Gas)
}

func (tx *TxCreateModel) Bytes() []byte {
	return tx.ld.Bytes()
}

func (tx *TxCreateModel) SyntacticVerify() error {
	return nil
}

func (tx *TxCreateModel) Verify(blk *Block) error {
	return nil
}

func (tx *TxCreateModel) Accept(blk *Block) error {
	return nil
}
