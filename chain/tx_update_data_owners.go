// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package chain

import (
	"math/big"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ldclabs/ldvm/ld"
)

type TxUpdateDataOwners struct {
	ld *ld.Transaction
	id ids.ID
}

func (tx *TxUpdateDataOwners) ID() ids.ID {
	return tx.ld.ID()
}

func (tx *TxUpdateDataOwners) Type() ld.TxType {
	return tx.ld.Type
}

func (tx *TxUpdateDataOwners) Gas() *big.Int {
	return new(big.Int).SetUint64(tx.ld.Gas)
}

func (tx *TxUpdateDataOwners) Bytes() []byte {
	return tx.ld.Bytes()
}

func (tx *TxUpdateDataOwners) SyntacticVerify() error {
	return nil
}

func (tx *TxUpdateDataOwners) Verify(blk *Block) error {
	return nil
}

func (tx *TxUpdateDataOwners) Accept(blk *Block) error {
	return nil
}
