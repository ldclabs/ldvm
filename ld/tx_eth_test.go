// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package ld

import (
	"encoding/json"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ldclabs/ldvm/util"
	"github.com/stretchr/testify/assert"
)

func TestTxEth(t *testing.T) {
	assert := assert.New(t)

	testTo := common.HexToAddress("b94f5374fce5edbc8e2a8697c15331677e6ebf0b")

	eip2718Tx := types.NewTx(&types.AccessListTx{
		ChainID:  new(big.Int).SetUint64(gChainID),
		Nonce:    0,
		To:       &testTo,
		Value:    ToEthBalance(big.NewInt(10)),
		Gas:      25000,
		GasPrice: big.NewInt(1000),
		Data:     common.FromHex("5544"),
	})
	eip2718TxHash := EthSigner.Hash(eip2718Tx)
	signedEip2718Tx, err := types.SignTx(eip2718Tx, EthSigner, util.Signer1.PK)
	assert.NoError(err)
	signedEip2718TxBinary, err := signedEip2718Tx.MarshalBinary()
	assert.NoError(err)

	txe := &TxEth{}
	assert.NoError(txe.Unmarshal(signedEip2718TxBinary))
	assert.NoError(txe.SyntacticVerify())
	assert.Equal(common.FromHex("5544"), txe.Data())
	assert.Equal(signedEip2718TxBinary, txe.Bytes())
	pk, err := util.DerivePublicKey(eip2718TxHash[:], txe.sigs[0][:])
	assert.NoError(err)
	assert.Equal(util.Signer1.Address(), util.EthID(crypto.PubkeyToAddress(*pk)))

	tx := txe.ToTransaction()
	assert.NoError(tx.SyntacticVerify())
	jsondata, err := json.Marshal(tx)
	assert.NoError(err)

	assert.Equal(TxType(1), tx.Type)
	assert.Equal(gChainID, tx.ChainID)
	assert.Equal(uint64(0), tx.Nonce)
	assert.Equal(uint64(0), tx.GasTip)
	assert.Equal(uint64(1000), tx.GasFeeCap)
	assert.Equal(uint64(1227), tx.Gas())
	assert.Equal(util.Signer1.Address(), tx.From)
	assert.Equal(testTo[:], tx.To[:])
	assert.Equal(big.NewInt(10), tx.Amount)
	assert.Equal(txe.Bytes(), []byte(tx.Data))
	assert.Equal(txe.sigs[0], tx.Signatures[0])

	signers, err := tx.Signers()
	assert.NoError(err)
	assert.Equal(util.EthIDs{tx.From}, signers)

	cbordata, err := tx.Marshal()
	assert.NoError(err)

	tx2 := &Transaction{}
	assert.NoError(tx2.Unmarshal(cbordata))
	assert.NoError(tx2.SyntacticVerify())
	cbordata2 := tx2.Bytes()
	jsondata2, err := json.Marshal(tx2)
	assert.Equal(string(jsondata), string(jsondata2))
	assert.Equal(cbordata, cbordata2)
}

func TestTxEthLegacy(t *testing.T) {
	assert := assert.New(t)

	testTo := common.HexToAddress("b94f5374fce5edbc8e2a8697c15331677e6ebf0b")
	signer := types.HomesteadSigner{}

	legacyTx := types.NewTx(&types.LegacyTx{
		Nonce:    3,
		To:       &testTo,
		Value:    ToEthBalance(big.NewInt(10)),
		Gas:      25000,
		GasPrice: big.NewInt(1000),
		Data:     common.FromHex("abcd"),
	})
	legacyTxHash := signer.Hash(legacyTx)
	signedLegacyTx, err := types.SignTx(legacyTx, signer, util.Signer1.PK)
	assert.NoError(err)
	signedLegacyTxBinary, err := signedLegacyTx.MarshalBinary()
	assert.NoError(err)

	txe := &TxEth{}
	assert.NoError(txe.Unmarshal(signedLegacyTxBinary))
	assert.NoError(txe.SyntacticVerify())
	assert.Equal(common.FromHex("abcd"), txe.Data())
	assert.Equal(signedLegacyTxBinary, txe.Bytes())
	pk, err := util.DerivePublicKey(legacyTxHash[:], txe.sigs[0][:])
	assert.NoError(err)
	assert.Equal(util.Signer1.Address(), util.EthID(crypto.PubkeyToAddress(*pk)))

	tx := txe.ToTransaction()
	assert.NoError(tx.SyntacticVerify())
	jsondata, err := json.Marshal(tx)
	assert.NoError(err)

	assert.Equal(TxType(1), tx.Type)
	assert.Equal(gChainID, tx.ChainID)
	assert.Equal(uint64(3), tx.Nonce)
	assert.Equal(uint64(0), tx.GasTip)
	assert.Equal(uint64(1000), tx.GasFeeCap)
	assert.Equal(uint64(1198), tx.Gas())
	assert.Equal(util.Signer1.Address(), tx.From)
	assert.Equal(testTo[:], tx.To[:])
	assert.Equal(big.NewInt(10), tx.Amount)
	assert.Equal(txe.Bytes(), []byte(tx.Data))
	assert.Equal(txe.sigs[0], tx.Signatures[0])

	cbordata, err := tx.Marshal()
	assert.NoError(err)

	tx2 := &Transaction{}
	assert.NoError(tx2.Unmarshal(cbordata))
	assert.NoError(tx2.SyntacticVerify())
	cbordata2 := tx2.Bytes()
	jsondata2, err := json.Marshal(tx2)
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
	signedEip2718Tx, err := types.SignTx(eip2718Tx, EthSigner, util.Signer1.PK)
	assert.NoError(err)
	signedEip2718TxBinary, err := signedEip2718Tx.MarshalBinary()
	assert.NoError(err)

	txe := &TxEth{}
	assert.NoError(txe.Unmarshal(signedEip2718TxBinary))
	assert.ErrorContains(txe.SyntacticVerify(), "invalid recipient")
}
