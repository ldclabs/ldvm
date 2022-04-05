// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package chain

import (
	"math/big"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ldclabs/ldvm/ld"
)

type TxUpdateModelKeepers struct {
	ld *ld.Transaction
	id ids.ID
}

func (tx *TxUpdateModelKeepers) ID() ids.ID {
	return tx.ld.ID()
}

func (tx *TxUpdateModelKeepers) Type() ld.TxType {
	return tx.ld.Type
}

func (tx *TxUpdateModelKeepers) Gas() *big.Int {
	return new(big.Int).SetUint64(tx.ld.Gas)
}

func (tx *TxUpdateModelKeepers) Bytes() []byte {
	return tx.ld.Bytes()
}

func (tx *TxUpdateModelKeepers) SyntacticVerify() error {
	return nil
}

func (tx *TxUpdateModelKeepers) Verify(blk *Block) error {
	return nil
}

func (tx *TxUpdateModelKeepers) Accept(blk *Block) error {
	return nil
}
