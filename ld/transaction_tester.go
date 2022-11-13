//go:build test

// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package ld

import (
	"encoding/json"
	"math/big"

	"github.com/ethereum/go-ethereum/core/types"

	"github.com/ldclabs/ldvm/ids"
	"github.com/ldclabs/ldvm/signer"
	"github.com/ldclabs/ldvm/unit"
)

func MustNewTestTx(signer1 *signer.SignerTester, ty TxType, to *ids.Address, data []byte) *Transaction {
	tx, err := NewTestTx(signer1, ty, to, data)
	if err != nil {
		panic(err)
	}
	return tx
}

func NewTestTx(signer1 *signer.SignerTester, ty TxType, to *ids.Address, data []byte) (*Transaction, error) {
	var err error
	tx := &Transaction{
		Tx: TxData{
			Type:      ty,
			ChainID:   gChainID,
			Nonce:     signer1.Nonce(),
			GasTip:    0,
			GasFeeCap: unit.LDC,
			From:      signer1.Key().Address(),
			To:        to,
			Data:      data,
		},
	}
	if to != nil {
		tx.Tx.Amount = new(big.Int).SetUint64(unit.LDC)
	}
	if err := tx.SignWith(signer1); err != nil {
		return nil, err
	}
	if err = tx.SyntacticVerify(); err != nil {
		return nil, err
	}
	return tx, nil
}

func NewEthTx(innerTx *types.AccessListTx) (*TxEth, error) {
	eip2718Tx := types.NewTx(innerTx)
	signedEip2718Tx, err := types.SignTx(eip2718Tx, EthSigner, signer.Signer1.PK)
	if err != nil {
		return nil, err
	}
	signedEip2718TxBinary, err := signedEip2718Tx.MarshalBinary()
	if err != nil {
		return nil, err
	}

	txe := &TxEth{}
	if err = txe.Unmarshal(signedEip2718TxBinary); err != nil {
		return nil, err
	}
	if err = txe.SyntacticVerify(); err != nil {
		return nil, err
	}
	return txe, nil
}

func GenJSONData(n int) []byte {
	switch {
	case n <= 0:
		return nil
	case n == 1:
		return []byte(`0`)
	case n == 2:
		return []byte(`""`)
	default:
		rt := make([]byte, n)
		rt[0] = '"'
		for i := 1; i < n-1; i++ {
			rt[i] = '1'
		}
		rt[n-1] = '"'
		return rt
	}
}

func MustNewToken(str string) ids.TokenSymbol {
	s, err := ids.TokenFromStr(str)
	if err != nil {
		panic(err)
	}
	return s
}

func MustNewStake(str string) ids.StakeSymbol {
	s, err := ids.StakeFromStr(str)
	if err != nil {
		panic(err)
	}
	return s
}

func MustMarshalJSON(v interface{}) []byte {
	data, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return data
}

func Uint16Ptr(u uint16) *uint16 {
	return &u
}
