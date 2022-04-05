// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package chain

import (
	"fmt"
	"math/big"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ldclabs/ldvm/ld"
)

type Transaction interface {
	ID() ids.ID
	Type() ld.TxType
	Gas() *big.Int
	Bytes() []byte
	SyntacticVerify() error
	Verify(blk *Block) error
	Accept(blk *Block) error
}

func NewTx(tx *ld.Transaction, syntacticVerifyLD bool) (Transaction, error) {
	if syntacticVerifyLD {
		if err := tx.SyntacticVerify(); err != nil {
			return nil, err
		}
	}

	var tt Transaction
	switch tx.Type {
	case ld.TypeMintFee:
		tt = &TxMintFee{ld: tx}
	case ld.TypeTransfer:
		tt = &TxTransfer{ld: tx}
	case ld.TypeUpdateAccountGuardians:
		tt = &TxUpdateAccountGuardians{ld: tx}
	case ld.TypeCreateModel:
		tt = &TxCreateModel{ld: tx}
	case ld.TypeUpdateModelKeepers:
		tt = &TxUpdateModelKeepers{ld: tx}
	case ld.TypeCreateData:
		tt = &TxCreateData{ld: tx}
	case ld.TypeUpdateData:
		tt = &TxUpdateData{ld: tx}
	case ld.TypeUpdateDataOwners:
		tt = &TxUpdateDataOwners{ld: tx}
	case ld.TypeUpdateDataOwnersByAuth:
		tt = &TxUpdateDataOwnersByAuth{ld: tx}
	case ld.TypeDeleteData:
		tt = &TxDeleteData{ld: tx}
	default:
		return nil, fmt.Errorf("unknown tx type: %d", tx.Type)
	}

	if err := tt.SyntacticVerify(); err != nil {
		return nil, err
	}
	return tt, nil
}
