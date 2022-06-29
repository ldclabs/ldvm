// go:build test

// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package ld

import (
	"math/big"

	"github.com/ethereum/go-ethereum/core/types"

	"github.com/ldclabs/ldvm/constants"
	"github.com/ldclabs/ldvm/util"
)

func MustNewTestTx(signer *util.Signer, ty TxType, to *util.EthID, data []byte) *Transaction {
	tx, err := NewTestTx(signer, ty, to, data)
	if err != nil {
		panic(err)
	}
	return tx
}

func NewTestTx(signer *util.Signer, ty TxType, to *util.EthID, data []byte) (*Transaction, error) {
	var err error
	txData := &TxData{
		Type:      ty,
		ChainID:   gChainID,
		Nonce:     signer.Nonce(),
		GasTip:    0,
		GasFeeCap: constants.LDC,
		From:      signer.Address(),
		To:        to,
		Data:      data,
	}
	if to != nil {
		txData.Amount = new(big.Int).SetUint64(constants.LDC)
	}
	if err = txData.SyntacticVerify(); err != nil {
		return nil, err
	}
	if err := txData.SignWith(signer); err != nil {
		return nil, err
	}
	tx := txData.ToTransaction()
	if err = tx.SyntacticVerify(); err != nil {
		return nil, err
	}
	return tx, nil
}

func NewEthTx(innerTx *types.AccessListTx) (*TxEth, error) {
	eip2718Tx := types.NewTx(innerTx)
	signedEip2718Tx, err := types.SignTx(eip2718Tx, EthSigner, util.Signer1.PK)
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

func MustNewToken(str string) util.TokenSymbol {
	s, err := util.NewToken(str)
	if err != nil {
		panic(err)
	}
	return s
}

func MustNewStake(str string) util.StakeSymbol {
	s, err := util.NewStake(str)
	if err != nil {
		panic(err)
	}
	return s
}
