// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package transactions

import (
	"fmt"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ldclabs/ldvm/ld"
)

type Transaction interface {
	LD() *ld.Transaction
	ID() ids.ID
	Type() ld.TxType
	Bytes() []byte
	SyntacticVerify() error
	Apply(ctx ChainContext, cs ChainState) error
	MarshalJSON() ([]byte, error)
}

// NewTx returns a stateful transaction from a ld.Transaction.
func NewTx(tx *ld.Transaction) (Transaction, error) {
	if tx.ID == ids.Empty {
		return nil, fmt.Errorf("NewTx: transaction should be syntactic verified")
	}

	var tt Transaction
	switch tx.Tx.Type {
	case ld.TypeTest:
		tt = &TxTest{TxBase: TxBase{ld: tx}}
	case ld.TypeEth:
		tt = &TxEth{TxBase: TxBase{ld: tx}}
	case ld.TypeTransfer:
		tt = &TxTransfer{TxBase: TxBase{ld: tx}}
	case ld.TypeTransferPay:
		tt = &TxTransferPay{TxBase: TxBase{ld: tx}}
	case ld.TypeTransferCash:
		tt = &TxTransferCash{TxBase: TxBase{ld: tx}}
	case ld.TypeExchange:
		tt = &TxExchange{TxBase: TxBase{ld: tx}}

	case ld.TypeAddNonceTable:
		tt = &TxAddNonceTable{TxBase: TxBase{ld: tx}}
	case ld.TypeUpdateAccountInfo:
		tt = &TxUpdateAccountInfo{TxBase: TxBase{ld: tx}}
	case ld.TypeCreateToken:
		tt = &TxCreateToken{TxBase: TxBase{ld: tx}}
	case ld.TypeDestroyToken:
		tt = &TxDestroyToken{TxBase: TxBase{ld: tx}}
	case ld.TypeCreateStake:
		tt = &TxCreateStake{TxBase: TxBase{ld: tx}}
	case ld.TypeResetStake:
		tt = &TxResetStake{TxBase: TxBase{ld: tx}}
	case ld.TypeDestroyStake:
		tt = &TxDestroyStake{TxBase: TxBase{ld: tx}}
	case ld.TypeTakeStake:
		tt = &TxTakeStake{TxBase: TxBase{ld: tx}}
	case ld.TypeWithdrawStake:
		tt = &TxWithdrawStake{TxBase: TxBase{ld: tx}}
	case ld.TypeUpdateStakeApprover:
		tt = &TxUpdateStakeApprover{TxBase: TxBase{ld: tx}}
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
	case ld.TypeUpdateModelInfo:
		tt = &TxUpdateModelInfo{TxBase: TxBase{ld: tx}}

	case ld.TypeCreateData:
		tt = &TxCreateData{TxBase: TxBase{ld: tx}}
	case ld.TypeUpdateData:
		tt = &TxUpdateData{TxBase: TxBase{ld: tx}}
	case ld.TypeUpgradeData:
		tt = &TxUpgradeData{TxBase: TxBase{ld: tx}}
	case ld.TypeUpdateDataInfo:
		tt = &TxUpdateDataInfo{TxBase: TxBase{ld: tx}}
	case ld.TypeUpdateDataInfoByAuth:
		tt = &TxUpdateDataInfoByAuth{TxBase: TxBase{ld: tx}}
	case ld.TypeDeleteData:
		tt = &TxDeleteData{TxBase: TxBase{ld: tx}}
	case ld.TypePunish:
		tt = &TxPunish{TxBase: TxBase{ld: tx}}
	default:
		return nil, fmt.Errorf("NewTx: unknown tx type %d", tx.Tx.Type)
	}

	if err := tt.SyntacticVerify(); err != nil {
		return nil, err
	}
	return tt, nil
}

type GenesisTx interface {
	ApplyGenesis(ctx ChainContext, cs ChainState) error
}

func NewGenesisTx(tx *ld.Transaction) (Transaction, error) {
	if err := tx.SyntacticVerify(); err != nil {
		return nil, err
	}

	var tt Transaction
	switch tx.Tx.Type {
	case ld.TypeTransfer:
		tt = &TxTransfer{TxBase: TxBase{ld: tx}}
	case ld.TypeUpdateAccountInfo:
		tt = &TxUpdateAccountInfo{TxBase: TxBase{ld: tx}}
	case ld.TypeCreateToken:
		tt = &TxCreateToken{TxBase: TxBase{ld: tx}}
	case ld.TypeCreateModel:
		tt = &TxCreateModel{TxBase: TxBase{ld: tx}}
	case ld.TypeCreateData:
		tt = &TxCreateData{TxBase: TxBase{ld: tx}}
	default:
		return nil, fmt.Errorf("NewGenesisTx: unsupport TxType: %s", tx.Tx.Type)
	}
	return tt, nil
}
