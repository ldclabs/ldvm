// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package chain

import (
	"math/big"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ldclabs/ldvm/ld"
)

type TxUpdateDataOwnersByAuth struct {
	ld *ld.Transaction
	id ids.ID
}

func (tx *TxUpdateDataOwnersByAuth) ID() ids.ID {
	return tx.ld.ID()
}

func (tx *TxUpdateDataOwnersByAuth) Type() ld.TxType {
	return tx.ld.Type
}

func (tx *TxUpdateDataOwnersByAuth) Gas() *big.Int {
	return new(big.Int).SetUint64(tx.ld.Gas)
}

func (tx *TxUpdateDataOwnersByAuth) Bytes() []byte {
	return tx.ld.Bytes()
}

func (tx *TxUpdateDataOwnersByAuth) SyntacticVerify() error {
	return nil
}

func (tx *TxUpdateDataOwnersByAuth) Verify(blk *Block) error {
	return nil
}

func (tx *TxUpdateDataOwnersByAuth) Accept(blk *Block) error {
	return nil
}
