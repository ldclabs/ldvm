// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package ld

import (
	"encoding/json"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ldclabs/ldvm/util/signer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTxEth(t *testing.T) {
	assert := assert.New(t)

	testTo := common.HexToAddress("b94F5374FCE5eDBC8E2A8697C15331677E6EBF0b")

	eip2718Tx := types.NewTx(&types.AccessListTx{
		ChainID:  new(big.Int).SetUint64(gChainID),
		Nonce:    0,
		To:       &testTo,
		Value:    ToEthBalance(big.NewInt(10)),
		Gas:      25000,
		GasPrice: ToEthBalance(big.NewInt(1000)),
		Data:     common.FromHex("5544"),
	})
	eip2718TxHash := EthSigner.Hash(eip2718Tx)
	signedEip2718Tx, err := types.SignTx(eip2718Tx, EthSigner, signer.Signer1.PK)
	require.NoError(t, err)
	signedEip2718TxBinary, err := signedEip2718Tx.MarshalBinary()
	require.NoError(t, err)

	txe := &TxEth{}
	assert.NoError(txe.Unmarshal(signedEip2718TxBinary))
	assert.NoError(txe.SyntacticVerify())
	assert.Equal(common.FromHex("5544"), txe.Data())
	assert.Equal(signedEip2718TxBinary, txe.Bytes())
	assert.Equal(0, txe.sig.FindKey(eip2718TxHash[:], signer.Signer1.Key()))

	tx := txe.ToTransaction()
	assert.NoError(tx.SyntacticVerify())
	jsondata, err := json.Marshal(tx)
	require.NoError(t, err)

	assert.Equal(TxType(1), tx.Tx.Type)
	assert.Equal(gChainID, tx.Tx.ChainID)
	assert.Equal(uint64(0), tx.Tx.Nonce)
	assert.Equal(uint64(0), tx.Tx.GasTip)
	assert.Equal(uint64(1000), tx.Tx.GasFeeCap)
	assert.Equal(uint64(1268), tx.Gas())
	assert.Equal(signer.Signer1.Key().Address(), tx.Tx.From)
	assert.Equal(testTo[:], tx.Tx.To[:])
	assert.Equal(big.NewInt(10), tx.Tx.Amount)
	assert.Equal(txe.Bytes(), []byte(tx.Tx.Data))
	assert.Equal(txe.sig, tx.Signatures[0])

	cbordata, err := tx.Marshal()
	require.NoError(t, err)
	// fmt.Println(string(jsondata))
	assert.Equal(`{"tx":{"type":"TypeEth","chainID":2357,"nonce":0,"gasTip":0,"gasFeeCap":1000,"from":"0x8db97c7cECe249C2b98bdc0226cc4C2A57bF52fc","to":"0xb94F5374FCE5eDBC8E2A8697C15331677E6EBF0b","amount":10,"data":"Afhvggk1gIXo1KUQAIJhqJS5T1N0_OXtvI4qhpfBUzFnfm6_C4UCVAvkAIJVRMCAoCyt0-a5DEePGkPlTvTT6xSe8uCoMpj_NkFWXeRfPzLmoHjk5vg2chuCcxQFq8omO4q6XkbsXrLqg78OEqUdO2UEgeq2Uw"},"sigs":["LK3T5rkMR48aQ-VO9NPrFJ7y4KgymP82QVZd5F8_MuZ45Ob4NnIbgnMUBavKJjuKul5G7F6y6oO_DhKlHTtlBACygGdJ"],"id":"WAcArva0XB9MUVi7NTVfF82tl1AypM7qE85Z_zs-HUoDGXMI"}`, string(jsondata))

	tx2 := &Transaction{}
	assert.NoError(tx2.Unmarshal(cbordata))
	assert.NoError(tx2.SyntacticVerify())
	cbordata2 := tx2.Bytes()
	jsondata2, _ := json.Marshal(tx2)
	assert.Equal(string(jsondata), string(jsondata2))
	assert.Equal(cbordata, cbordata2)
}

func TestTxEthLegacy(t *testing.T) {
	assert := assert.New(t)

	testTo := common.HexToAddress("b94f5374fce5edbc8e2a8697c15331677e6ebf0b")
	ethSigner := types.HomesteadSigner{}

	legacyTx := types.NewTx(&types.LegacyTx{
		Nonce:    3,
		To:       &testTo,
		Value:    ToEthBalance(big.NewInt(10)),
		Gas:      25000,
		GasPrice: ToEthBalance(big.NewInt(1000)),
		Data:     common.FromHex("abcd"),
	})
	legacyTxHash := ethSigner.Hash(legacyTx)
	signedLegacyTx, err := types.SignTx(legacyTx, ethSigner, signer.Signer1.PK)
	require.NoError(t, err)
	signedLegacyTxBinary, err := signedLegacyTx.MarshalBinary()
	require.NoError(t, err)

	txe := &TxEth{}
	assert.NoError(txe.Unmarshal(signedLegacyTxBinary))
	assert.NoError(txe.SyntacticVerify())
	assert.Equal(common.FromHex("abcd"), txe.Data())
	assert.Equal(signedLegacyTxBinary, txe.Bytes())
	assert.Equal(0, txe.sig.FindKey(legacyTxHash[:], signer.Signer1.Key()))

	tx := txe.ToTransaction()
	assert.NoError(tx.SyntacticVerify())
	jsondata, err := json.Marshal(tx)
	require.NoError(t, err)

	assert.Equal(TxType(1), tx.Tx.Type)
	assert.Equal(gChainID, tx.Tx.ChainID)
	assert.Equal(uint64(3), tx.Tx.Nonce)
	assert.Equal(uint64(0), tx.Tx.GasTip)
	assert.Equal(uint64(1000), tx.Tx.GasFeeCap)
	assert.Equal(uint64(1239), tx.Gas())
	assert.Equal(signer.Signer1.Key().Address(), tx.Tx.From)
	assert.Equal(testTo[:], tx.Tx.To[:])
	assert.Equal(big.NewInt(10), tx.Tx.Amount)
	assert.Equal(txe.Bytes(), []byte(tx.Tx.Data))
	assert.Equal(txe.sig, tx.Signatures[0])

	cbordata, err := tx.Marshal()
	require.NoError(t, err)

	tx2 := &Transaction{}
	assert.NoError(tx2.Unmarshal(cbordata))
	assert.NoError(tx2.SyntacticVerify())
	cbordata2 := tx2.Bytes()
	jsondata2, _ := json.Marshal(tx2)
	assert.Equal(string(jsondata), string(jsondata2))
	assert.Equal(cbordata, cbordata2)
}

func TestTxEthErr(t *testing.T) {
	assert := assert.New(t)

	// empty to
	eip2718Tx := types.NewTx(&types.AccessListTx{
		ChainID:  new(big.Int).SetUint64(gChainID),
		Nonce:    3,
		To:       &common.Address{},
		Value:    ToEthBalance(big.NewInt(10)),
		Gas:      25000,
		GasPrice: big.NewInt(1000),
		Data:     common.FromHex("5544"),
	})
	signedEip2718Tx, err := types.SignTx(eip2718Tx, EthSigner, signer.Signer1.PK)
	require.NoError(t, err)
	signedEip2718TxBinary, err := signedEip2718Tx.MarshalBinary()
	require.NoError(t, err)

	txe := &TxEth{}
	assert.NoError(txe.Unmarshal(signedEip2718TxBinary))
	assert.ErrorContains(txe.SyntacticVerify(), "invalid recipient")
}
