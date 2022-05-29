// go:build test

// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package ld

import (
	"math/big"

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
		GasTip:    constants.MilliLDC / 100,
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
