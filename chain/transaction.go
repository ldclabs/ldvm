// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package chain

import (
	"fmt"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/snow/choices"
	"github.com/ldclabs/ldvm/ld"
)

type Transaction interface {
	LD() *ld.Transaction
	ID() ids.ID
	Type() ld.TxType
	Bytes() []byte
	Status() string
	SetStatus(choices.Status)
	SyntacticVerify() error
	Verify(blk *Block, bs BlockState) error
	Accept(blk *Block, bs BlockState) error
	Event(ts int64) *Event
	MarshalJSON() ([]byte, error)
}

func NewTx(tx *ld.Transaction, syntacticVerifyLD bool) (Transaction, error) {
	if syntacticVerifyLD {
		if err := tx.SyntacticVerify(); err != nil {
			return nil, err
		}
	}
	var tt Transaction
	switch tx.Type {
	case ld.TypeTest:
		tt = &TxTest{ld: tx}

	case ld.TypeEth:
		tt = &TxEth{TxBase: TxBase{ld: tx}}
	case ld.TypeTransfer:
		tt = &TxTransfer{TxBase: TxBase{ld: tx}}
	case ld.TypeTransferPay:
		tt = &TxTransferPay{TxBase: TxBase{ld: tx}}
	case ld.TypeTransferCash:
		tt = &TxTransferCash{TxBase: TxBase{ld: tx}}
	case ld.TypeExchange:
		tt = &TxTransferExchange{TxBase: TxBase{ld: tx}}

	case ld.TypeAddNonceTable:
		tt = &TxAddAccountNonceTable{TxBase: TxBase{ld: tx}}
	case ld.TypeUpdateAccountKeepers:
		tt = &TxUpdateAccountKeepers{TxBase: TxBase{ld: tx}}
	case ld.TypeCreateTokenAccount:
		tt = &TxCreateTokenAccount{TxBase: TxBase{ld: tx}}
	case ld.TypeDestroyTokenAccount:
		tt = &TxDestroyTokenAccount{TxBase: TxBase{ld: tx}}
	case ld.TypeCreateStakeAccount:
		tt = &TxCreateStakeAccount{TxBase: TxBase{ld: tx}}
	case ld.TypeResetStakeAccount:
		tt = &TxResetStakeAccount{TxBase: TxBase{ld: tx}}
	case ld.TypeTakeStake:
		tt = &TxTakeStake{TxBase: TxBase{ld: tx}}
	case ld.TypeWithdrawStake:
		tt = &TxWithdrawStake{TxBase: TxBase{ld: tx}}
	case ld.TypeOpenLending:
		tt = &TxOpenLending{TxBase: TxBase{ld: tx}}
	case ld.TypeCloseLending:
		tt = &TxCloseLending{TxBase: TxBase{ld: tx}}
	case ld.TypeBorrow:
		tt = &TxBorrow{TxBase: TxBase{ld: tx}}
	case ld.TypeRepay:
		tt = &TxRepay{TxBase: TxBase{ld: tx}}

	case ld.TypeCreateModel:
		tt = &TxCreateModel{TxBase: TxBase{ld: tx}}
	case ld.TypeUpdateModelKeepers:
		tt = &TxUpdateModelKeepers{TxBase: TxBase{ld: tx}}

	case ld.TypeCreateData:
		tt = &TxCreateData{TxBase: TxBase{ld: tx}}
	case ld.TypeUpdateData:
		tt = &TxUpdateData{TxBase: TxBase{ld: tx}}
	case ld.TypeUpdateDataKeepers:
		tt = &TxUpdateDataKeepers{TxBase: TxBase{ld: tx}}
	case ld.TypeUpdateDataKeepersByAuth:
		tt = &TxUpdateDataKeepersByAuth{TxBase: TxBase{ld: tx}}
	case ld.TypeDeleteData:
		tt = &TxDeleteData{TxBase: TxBase{ld: tx}}
	case ld.TypePunish:
		tt = &TxPunish{TxBase: TxBase{ld: tx}}
	default:
		return nil, fmt.Errorf("unknown tx type: %d", tx.Type)
	}

	if err := tt.SyntacticVerify(); err != nil {
		return nil, err
	}
	return tt, nil
}

type GenesisTx interface {
	VerifyGenesis(blk *Block, bs BlockState) error
}

func NewGenesisTx(tx *ld.Transaction) (Transaction, error) {
	var tt Transaction
	switch tx.Type {
	case ld.TypeTransfer:
		tt = &TxTransfer{TxBase: TxBase{ld: tx}}
	case ld.TypeUpdateAccountKeepers:
		tt = &TxUpdateAccountKeepers{TxBase: TxBase{ld: tx}}
	case ld.TypeCreateTokenAccount:
		tt = &TxCreateTokenAccount{TxBase: TxBase{ld: tx}}
	case ld.TypeCreateModel:
		tt = &TxCreateModel{TxBase: TxBase{ld: tx}}
	case ld.TypeCreateData:
		tt = &TxCreateData{TxBase: TxBase{ld: tx}}
	default:
		return nil, fmt.Errorf("not support genesis tx type: %d", tx.Type)
	}
	return tt, nil
}
