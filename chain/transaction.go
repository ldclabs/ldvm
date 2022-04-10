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
	Bytes() []byte
	SyntacticVerify() error
	Verify(blk *Block) error
	Accept(blk *Block) error
	Event(ts int64) *Event
	MarshalJSON() ([]byte, error)
}

type GenesisTx interface {
	VerifyGenesis(blk *Block) error
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
	case ld.TypeTransferReply:
		tt = &TxTransferReply{ld: tx}
	case ld.TypeTransferCash:
		tt = &TxTransferCash{ld: tx}
	case ld.TypeUpdateAccountKeepers:
		tt = &TxUpdateAccountKeepers{ld: tx}
	case ld.TypeCreateModel:
		tt = &TxCreateModel{ld: tx}
	case ld.TypeUpdateModelKeepers:
		tt = &TxUpdateModelKeepers{ld: tx}
	case ld.TypeCreateData:
		tt = &TxCreateData{ld: tx}
	case ld.TypeUpdateData:
		tt = &TxUpdateData{ld: tx}
	case ld.TypeUpdateDataKeepers:
		tt = &TxUpdateDataKeepers{ld: tx}
	case ld.TypeUpdateDataKeepersByAuth:
		tt = &TxUpdateDataKeepersByAuth{ld: tx}
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

func newGenesisTx(tx *ld.Transaction) (Transaction, error) {
	var tt Transaction
	switch tx.Type {
	case ld.TypeMintFee:
		tt = &TxMintFee{ld: tx}
	case ld.TypeTransfer:
		tt = &TxTransfer{ld: tx}
	case ld.TypeUpdateAccountKeepers:
		tt = &TxUpdateAccountKeepers{ld: tx}
	case ld.TypeCreateModel:
		tt = &TxCreateModel{ld: tx}
	case ld.TypeCreateData:
		tt = &TxCreateData{ld: tx}
	default:
		return nil, fmt.Errorf("unknown tx type: %d", tx.Type)
	}
	return tt, nil
}

func verifyBase(blk *Block, tx *ld.Transaction, signers []ids.ShortID) (*Account, error) {
	bs := blk.State()
	if err := bs.ChainConfig().CheckChainID(tx.ChainID); err != nil {
		return nil, err
	}

	requireGas := tx.RequireGas(blk.bs.FeeConfig().ThresholdGas)
	if tx.Gas < requireGas || requireGas > tx.GasFeeCap {
		return nil, fmt.Errorf("tx gas not matching, require %d", requireGas)
	}

	from, err := bs.LoadAccount(tx.From)
	if err != nil {
		return nil, err
	}

	if tx.Nonce != from.Nonce() {
		return nil, fmt.Errorf("account nonce not matching")
	}
	if !from.SatisfySigning(signers) {
		return nil, fmt.Errorf("need more account signatures")
	}

	cost := new(big.Int).Mul(tx.BigIntGas(), blk.GasPrice())
	if tx.Amount != nil {
		if tx.Amount.Sign() < 0 {
			return nil, fmt.Errorf("invalid amount %d", tx.Amount)
		} else if tx.Amount.Sign() > 0 && tx.To == ids.ShortEmpty {
			return nil, fmt.Errorf("required recipient to recive %d", tx.Amount)
		}

		cost = cost.Add(cost, tx.Amount)
	}
	if from.Balance().Cmp(cost) < 0 {
		return nil, fmt.Errorf("insufficient balance %d of account %s, required %d",
			from.Balance(), tx.From, cost)
	}
	return from, nil
}
