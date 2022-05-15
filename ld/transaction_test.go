// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package ld

import (
	"bytes"
	"encoding/json"
	"math/big"
	"testing"

	"github.com/ldclabs/ldvm/constants"
	"github.com/ldclabs/ldvm/util"
	"github.com/stretchr/testify/assert"
)

func TestTxData(t *testing.T) {
	assert := assert.New(t)

	var tx *TxData
	assert.ErrorContains(tx.SyntacticVerify(), "nil pointer")

	tx = &TxData{Type: TypeDeleteData + 1}
	assert.ErrorContains(tx.SyntacticVerify(), "invalid type")

	tx = &TxData{Type: TypeTransfer, ChainID: 1000}
	assert.ErrorContains(tx.SyntacticVerify(), "invalid ChainID")

	tx = &TxData{Type: TypeTransfer, ChainID: gChainID, Token: &util.TokenSymbol{'a', 'b', 'c'}}
	assert.ErrorContains(tx.SyntacticVerify(), "invalid token symbol")

	tx = &TxData{Type: TypeTransfer, ChainID: gChainID, Amount: big.NewInt(0)}
	assert.ErrorContains(tx.SyntacticVerify(), "invalid amount")

	tx = &TxData{Type: TypeTransfer, ChainID: gChainID, Data: RawData{}}
	assert.ErrorContains(tx.SyntacticVerify(), "empty data")

	tx = &TxData{Type: TypeTransfer, ChainID: gChainID, Signatures: []util.Signature{}}
	assert.ErrorContains(tx.SyntacticVerify(), "empty signatures")

	tx = &TxData{Type: TypeTransfer, ChainID: gChainID, ExSignatures: []util.Signature{}}
	assert.ErrorContains(tx.SyntacticVerify(), "empty exSignatures")

	tx = &TxData{ChainID: gChainID}
	assert.NoError(tx.SyntacticVerify())

	data := [2022]byte{}
	to := util.Signer2.Address()
	tx = &TxData{
		Type:    TypeTransfer,
		ChainID: gChainID,
		From:    util.Signer1.Address(),
		To:      &to,
		Amount:  big.NewInt(1),
		Data:    data[:],
	}
	assert.NoError(tx.SyntacticVerify())

	unsignedBytes, err := tx.Marshal()
	assert.NoError(err)
	assert.Equal(unsignedBytes, tx.UnsignedBytes())

	assert.Equal(uint64(8751), tx.RequiredGas(1000))
	assert.Equal(uint64(2142), tx.RequiredGas(3000))

	tx2 := &TxData{}
	assert.NoError(tx2.Unmarshal(unsignedBytes))
	assert.NoError(tx2.SyntacticVerify())

	cbordata := tx2.Bytes()
	assert.Equal(unsignedBytes, cbordata)

	sig, err := util.Signer1.Sign(tx.UnsignedBytes())
	assert.NoError(err)
	tx.Signatures = append(tx.Signatures, sig)
	assert.NoError(tx.SyntacticVerify())
	assert.Equal(unsignedBytes, tx.UnsignedBytes())
	assert.NotEqual(unsignedBytes, tx.Bytes())
	assert.Equal(uint64(8751), tx.RequiredGas(1000))

	tx2 = &TxData{}
	assert.NoError(tx2.Unmarshal(tx.Bytes()))
	assert.NoError(tx2.SyntacticVerify())
	assert.Equal(unsignedBytes, tx2.UnsignedBytes())
	assert.Equal(tx.Bytes(), tx2.Bytes())

	txx := tx.ToTransaction()
	assert.NoError(txx.SyntacticVerify())
	assert.Equal(tx.ID(), txx.ID)

	jsondata, err := json.Marshal(txx)
	assert.NoError(err)
	assert.Contains(string(jsondata), `"from":"0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC"`)
	assert.Contains(string(jsondata), `"id":"Av2arMc7hB2RYsfFqdsYu9tPnqvdzmxqLMceQ7BbKHyzf1VgF"`)
	assert.NotContains(string(jsondata), `"exSignatures":null`)
	assert.Contains(string(jsondata), `"gas":0`)
	assert.Contains(string(jsondata), `"type":3`)
	assert.Contains(string(jsondata), `"name":"TransferTx"`)

	txx2 := &Transaction{}
	assert.NoError(txx2.Unmarshal(txx.Bytes()))
	assert.NoError(txx2.SyntacticVerify())
	assert.Equal(txx.Bytes(), txx2.Bytes())
	assert.Equal(txx.ID, txx2.ID)

	jsondata2, err := json.Marshal(txx2)
	assert.NoError(err)
	assert.Equal(string(jsondata), string(jsondata2))
}

func TestTransaction(t *testing.T) {
	assert := assert.New(t)

	var tx *Transaction
	assert.ErrorContains(tx.SyntacticVerify(), "nil pointer")

	to := util.Signer2.Address()
	txData := &TxData{
		Type:      TypeTransfer,
		ChainID:   gChainID,
		Nonce:     1,
		GasTip:    0,
		GasFeeCap: 1000,
		From:      util.Signer1.Address(),
		To:        &to,
		Amount:    big.NewInt(1_000_000),
	}
	sig1, err := util.Signer1.Sign(txData.UnsignedBytes())
	assert.NoError(err)
	txData.Signatures = append(txData.Signatures, sig1)
	tx = txData.ToTransaction()
	assert.NoError(tx.SyntacticVerify())

	jsondata, err := json.Marshal(tx)
	assert.NoError(err)
	assert.Equal(`{"type":3,"chainID":2357,"nonce":1,"gasTip":0,"gasFeeCap":1000,"from":"0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC","to":"0x44171C37Ff5D7B7bb8dcad5C81f16284A229e641","amount":1000000,"signatures":["070a1d67010bfecec1309e0d30f62f9f73f339ad8fa726c3b70d43066089a92660a2104180dd0f2335fcd3c599f641ed8e9bc6ce88d7b1b71285120fb3fa1d1c01"],"gas":0,"name":"TransferTx","id":"12sX66xsbsSAZCN6ZWv2bRgua9EXEwwyx5eeBrRC16hwjRnce"}`, string(jsondata))

	ctx := tx.Copy()
	assert.NoError(ctx.SyntacticVerify())
	jsondata2, err := json.Marshal(tx)
	assert.NoError(err)
	assert.Equal(jsondata, jsondata2)
	ctx.Data = []byte(`"Hello, world!"`)
	jsondata2, err = json.Marshal(ctx)
	assert.NoError(err)
	assert.Contains(string(jsondata2), `"data":"Hello, world!"`)

	tx2 := &Transaction{}
	assert.NoError(tx2.UnmarshalTx(txData.Bytes()))
	assert.NoError(tx2.SyntacticVerify())
	assert.Equal(tx.Bytes(), tx2.Bytes())
	assert.Equal(tx.ShortID(), tx2.ShortID())

	tx3 := &Transaction{}
	assert.NoError(tx3.Unmarshal(tx2.Bytes()))
	assert.NoError(tx3.SyntacticVerify())
	assert.Equal(tx.Bytes(), tx3.Bytes())
	assert.Equal(tx.ShortID(), tx3.ShortID())

	assert.Equal(uint64(119), tx.RequiredGas(1000))
	assert.Equal(uint64(0), tx.GasUnits().Uint64())

	signers, err := tx.Signers()
	assert.NoError(err)
	assert.Equal(util.EthIDs{util.Signer1.Address()}, signers)
	_, err = tx.ExSigners()
	assert.ErrorContains(err, `DeriveSigners: empty data or signature`)

	assert.False(tx.IsBatched())
	assert.False(tx.NeedApprove(nil, nil))
	assert.True(tx.NeedApprove(&constants.GenesisAccount, nil))
	assert.True(tx.NeedApprove(&constants.GenesisAccount, []TxType{TypeTransfer}))
	assert.False(tx.NeedApprove(&constants.GenesisAccount, []TxType{TypeUpdateAccountKeepers}))
}

func TestTxs(t *testing.T) {
	assert := assert.New(t)

	testTx := (&TxData{
		Type:    TypeTest,
		ChainID: gChainID,
	}).ToTransaction()

	_, err := NewBatchTx(testTx)
	assert.ErrorContains(err, "NewBatchTx: not batch transactions")

	to := util.Signer2.Address()
	txData := &TxData{
		Type:      TypeTransfer,
		ChainID:   gChainID,
		Nonce:     1,
		GasTip:    0,
		GasFeeCap: 1000,
		From:      util.Signer1.Address(),
		To:        &to,
		Amount:    big.NewInt(1_000_000),
	}
	sig1, err := util.Signer1.Sign(txData.UnsignedBytes())
	assert.NoError(err)
	txData.Signatures = append(txData.Signatures, sig1)
	tx1 := txData.ToTransaction()

	txData = &TxData{
		Type:      TypeTransfer,
		ChainID:   gChainID,
		Nonce:     2,
		GasTip:    0,
		GasFeeCap: 1000,
		From:      util.Signer1.Address(),
		To:        &to,
		Amount:    big.NewInt(1_000_000),
		Data:      []byte(`"👋"`),
	}
	sig1, err = util.Signer1.Sign(txData.UnsignedBytes())
	assert.NoError(err)
	txData.Signatures = append(txData.Signatures, sig1)
	tx2 := txData.ToTransaction()

	txs, err := NewBatchTx(testTx, tx1, tx2)
	assert.NoError(err)

	assert.True(txs.IsBatched())
	assert.Equal(tx2.ID, txs.ID)
	assert.Equal(tx2.Bytes(), txs.Bytes())

	data, err := txs.Txs().Marshal()
	assert.NoError(err)
	txs2 := Txs{}
	assert.NoError(txs2.Unmarshal(data))
	assert.Equal(3, len(txs2))
	assert.Equal(testTx.Bytes(), txs2[0].Bytes())
	assert.Equal(tx1.Bytes(), txs2[1].Bytes())
	assert.Equal(tx2.Bytes(), txs2[2].Bytes())
}

func TestTxsSort(t *testing.T) {
	assert := assert.New(t)

	to := util.Signer2.Address()
	txData := &TxData{
		Type:      TypeTransfer,
		ChainID:   gChainID,
		Nonce:     1,
		GasTip:    200,
		GasFeeCap: 1000,
		From:      util.Signer1.Address(),
		To:        &to,
		Amount:    big.NewInt(1_000_000),
		Data:      []byte(`"Hello, world!"`),
	}
	sig, err := util.Signer1.Sign(txData.UnsignedBytes())
	assert.NoError(err)
	txData.Signatures = append(txData.Signatures, sig)
	tx1 := txData.ToTransaction()
	assert.NoError(tx1.SyntacticVerify())

	txData = &TxData{
		Type:      TypeTransfer,
		ChainID:   gChainID,
		Nonce:     2,
		GasTip:    200,
		GasFeeCap: 1000,
		From:      util.Signer1.Address(),
		To:        &to,
		Amount:    big.NewInt(1_000_000),
	}
	sig, err = util.Signer1.Sign(txData.UnsignedBytes())
	assert.NoError(err)
	txData.Signatures = append(txData.Signatures, sig)
	tx2 := txData.ToTransaction()
	assert.NoError(tx2.SyntacticVerify())

	data := [1024]byte{}
	kSig, err := util.Signer2.Sign(data[:])
	assert.NoError(err)

	dm := &DataMeta{
		Version:   1,
		Threshold: 1,
		Keepers:   util.EthIDs{util.Signer1.Address(), util.Signer2.Address()},
		Data:      data[:],
		KSig:      kSig,
	}
	assert.NoError(dm.SyntacticVerify())
	cbordata, err := dm.Marshal()
	assert.NoError(err)
	txData = &TxData{
		Type:      TypeCreateData,
		ChainID:   gChainID,
		Nonce:     1,
		GasTip:    100,
		GasFeeCap: 1000,
		From:      util.Signer2.Address(),
		Data:      cbordata,
	}
	sig, err = util.Signer2.Sign(txData.UnsignedBytes())
	assert.NoError(err)
	txData.Signatures = append(txData.Signatures, sig)
	tx3 := txData.ToTransaction()
	assert.NoError(tx3.SyntacticVerify())

	txs := Txs{tx2, tx1, tx3}
	txs.Sort()
	assert.Equal(-1, bytes.Compare(tx3.ID[:], tx1.ID[:]))
	assert.Equal(tx3.ID, txs[0].ID)
	assert.Equal(tx1.ID, txs[1].ID)
	assert.Equal(tx2.ID, txs[2].ID)

	tx3.Priority = 1
	tx1.Priority = 2
	tx2.Priority = 3
	txs.Sort()

	assert.Equal(tx1.ID, txs[0].ID)
	assert.Equal(tx2.ID, txs[1].ID)
	assert.Equal(tx3.ID, txs[2].ID)

	txs.UpdatePriority(1000, 3)
	assert.True(tx3.Priority > tx2.Priority)
	txs.Sort()
	assert.Equal(tx3.ID, txs[0].ID)
	assert.Equal(tx1.ID, txs[1].ID)
	assert.Equal(tx2.ID, txs[2].ID)

	tx1.AddedTime = 10
	tx2.AddedTime = 20
	tx3.AddedTime = 110
	txs.UpdatePriority(1000, 120)
	assert.True(tx3.Priority < tx2.Priority)
	txs.Sort()
	assert.Equal(tx1.ID, txs[0].ID)
	assert.Equal(tx2.ID, txs[1].ID)
	assert.Equal(tx3.ID, txs[2].ID)
}
