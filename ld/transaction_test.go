// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package ld

import (
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

	tx = &TxData{Type: TypeTransfer, ChainID: gChainID, Amount: big.NewInt(1)}
	assert.ErrorContains(tx.SyntacticVerify(), "invalid to")

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
	assert.Equal(tx.ID, txx.ID)

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
	assert.Equal(len(tx.Bytes()), tx.BytesSize())

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
	assert.Equal(tx.ID, tx2.ID)
	assert.Equal(tx.Bytes(), tx2.Bytes())
	assert.Equal(tx.ShortID(), tx2.ShortID())

	tx3 := &Transaction{}
	assert.NoError(tx3.Unmarshal(tx2.Bytes()))
	assert.NoError(tx3.SyntacticVerify())
	assert.Equal(tx.ID, tx3.ID)
	assert.Equal(tx.Bytes(), tx3.Bytes())
	assert.Equal(tx.ShortID(), tx3.ShortID())

	assert.Equal(uint64(119), tx.RequiredGas(1000))
	assert.Equal(uint64(0), tx.GasUnits().Uint64())

	tx3.GasFeeCap++
	assert.NoError(tx3.SyntacticVerify())
	assert.NotEqual(tx.ID, tx3.ID)
	assert.NotEqual(tx.Bytes(), tx3.Bytes())
	assert.NotEqual(tx.ShortID(), tx3.ShortID())

	signers, err := tx.Signers()
	assert.NoError(err)
	assert.Equal(util.EthIDs{util.Signer1.Address()}, signers)
	_, err = tx.ExSigners()
	assert.ErrorContains(err, `DeriveSigners: empty data`)

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

	assert.NoError(testTx.SyntacticVerify())
	assert.Equal(0, testTx.BytesSize())

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
	assert.Equal(len(tx1.Bytes())+len(tx2.Bytes()), txs.BytesSize())

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

	to := util.Signer1.Address()
	s0 := util.NewSigner()
	s1 := util.NewSigner()
	s2 := util.NewSigner()
	s3 := util.NewSigner()

	stx0 := MustNewTestTx(s0, TypeTest, nil, nil)
	stx1 := MustNewTestTx(s0, TypeTransfer, &to, GenJSONData(100))
	stx2 := MustNewTestTx(s1, TypeTransfer, &to, GenJSONData(200))
	stx3 := MustNewTestTx(s2, TypeTransfer, &to, GenJSONData(1100))
	stx4 := MustNewTestTx(s1, TypeTransfer, &to, GenJSONData(1000))
	btx, err := NewBatchTx(stx0, stx1, stx2, stx3, stx4)
	assert.NoError(err)
	assert.Equal(stx3.ID, btx.ID)
	assert.Equal(stx3.RequiredGas(1000), btx.RequiredGas(1000))
	assert.Equal(len(stx1.Bytes())+len(stx2.Bytes())+len(stx3.Bytes())+len(stx4.Bytes()), btx.BytesSize())
	assert.Equal(uint64(0), stx0.priority)
	assert.Equal(uint64(0), stx1.priority)
	assert.Equal(uint64(0), stx2.priority)
	assert.Equal(uint64(0), stx3.priority)
	assert.Equal(uint64(0), stx4.priority)
	assert.Equal(uint64(0), btx.priority)

	btx.SetPriority(1000, 0)
	assert.Equal(uint64(0), stx0.priority)
	assert.True(stx2.priority == stx1.priority, "small bytes size txs has the same priority")
	assert.True(stx3.priority > stx2.priority)
	assert.True(stx4.priority > stx2.priority)
	assert.True(stx3.priority > stx4.priority)
	assert.Equal(stx3.priority, btx.priority)

	tx0 := MustNewTestTx(s0, TypeTransfer, &to, nil)
	tx1 := MustNewTestTx(s1, TypeTransfer, &to, GenJSONData(1000))
	tx2 := MustNewTestTx(s2, TypeTransfer, &to, GenJSONData(1200))
	tx3 := MustNewTestTx(s3, TypeTransfer, &to, GenJSONData(1500))
	txs := Txs{tx0, tx1, tx2, tx3}
	txs.SortWith(1000, 0)
	assert.Equal(tx3.ID, txs[0].ID)
	assert.Equal(tx2.ID, txs[1].ID)
	assert.Equal(tx1.ID, txs[2].ID)
	assert.Equal(tx0.ID, txs[3].ID)

	txs = append(txs, btx)
	txs.SortWith(1000, 0)
	assert.Equal(tx3.ID, txs[0].ID)
	assert.Equal(btx.ID, txs[1].ID)
	assert.Equal(tx2.ID, txs[2].ID)
	assert.Equal(tx1.ID, txs[3].ID)
	assert.Equal(tx0.ID, txs[4].ID)

	assert.Equal(stx0.ID, btx.batch[0].ID)
	assert.Equal(stx1.ID, btx.batch[1].ID)
	assert.Equal(stx2.ID, btx.batch[2].ID)
	assert.Equal(stx3.ID, btx.batch[3].ID)
	assert.Equal(stx4.ID, btx.batch[4].ID)

	tx0.AddedTime = 121
	tx1.AddedTime = 121
	tx2.AddedTime = 121
	tx3.AddedTime = 121
	txs.SortWith(1000, 120)
	assert.Equal(btx.ID, txs[0].ID, "delay should be feedback into priority")
	assert.Equal(tx1.ID, txs[1].ID, "delay should be feedback into priority")
	assert.Equal(tx2.ID, txs[2].ID, "delay should be feedback into priority")
	assert.Equal(tx0.ID, txs[3].ID, "delay should be feedback into priority")
	assert.Equal(tx3.ID, txs[4].ID)
}
