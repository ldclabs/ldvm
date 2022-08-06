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

	tx = &TxData{Type: TypeTransfer, ChainID: gChainID, Amount: big.NewInt(-1)}
	assert.ErrorContains(tx.SyntacticVerify(), "invalid amount")

	tx = &TxData{Type: TypeTransfer, ChainID: gChainID, Amount: big.NewInt(0)}
	assert.ErrorContains(tx.SyntacticVerify(), "nil to together with amount")

	tx = &TxData{Type: TypeTransfer, ChainID: gChainID, Data: util.RawData{}}
	assert.ErrorContains(tx.SyntacticVerify(), "empty data")

	tx = &TxData{Type: TypeTransfer, ChainID: gChainID, Signatures: []util.Signature{}}
	assert.ErrorContains(tx.SyntacticVerify(), "empty signatures")

	tx = &TxData{Type: TypeTransfer, ChainID: gChainID, ExSignatures: []util.Signature{}}
	assert.ErrorContains(tx.SyntacticVerify(), "empty exSignatures")

	tx = &TxData{ChainID: gChainID}
	assert.NoError(tx.SyntacticVerify())

	// data := [1024 * 128]byte{}
	to := util.Signer2.Address()
	tx = &TxData{
		Type:    TypeTransfer,
		ChainID: gChainID,
		From:    util.Signer1.Address(),
		To:      &to,
		Amount:  big.NewInt(1),
	}
	assert.NoError(tx.SyntacticVerify())

	data := tx.Bytes()
	assert.Equal(data, tx.UnsignedBytes())

	assert.NoError(tx.SignWith(util.Signer1))
	assert.NoError(tx.SyntacticVerify())
	assert.Equal(data, tx.UnsignedBytes())
	assert.NotEqual(data, tx.Bytes())
	assert.Equal(73, len(tx.UnsignedBytes()))
	assert.Equal(144, len(tx.Bytes()))
	assert.Equal(uint64(598), tx.Gas(), "a very small gas transaction")

	jsondata, err := json.Marshal(tx)
	assert.NoError(err)
	// fmt.Println(string(jsondata))
	assert.Equal(`{"type":"TypeTransfer","chainID":2357,"nonce":0,"gasTip":0,"gasFeeCap":0,"from":"0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC","to":"0x44171C37Ff5D7B7bb8dcad5C81f16284A229e641","amount":1,"signatures":["c137a94f3bcc4dd10a0b21472bd1e1af5744f3c8f3fe63d25ae23d379926c0c9223c5fc9fdd5d4de8e41c9426f49a397ed14033e0865cccfa1ee39a01404e57d01"]}`, string(jsondata))

	tx = &TxData{
		Type:    TypeTransfer,
		ChainID: gChainID,
		From:    util.Signer1.Address(),
		To:      &to,
		Amount:  big.NewInt(1),
		Data:    GenJSONData(1024 * 256),
	}
	assert.ErrorContains(tx.SyntacticVerify(),
		"TxData.SyntacticVerify error: size too large, expected <= 262144, got 262224")

	tx = &TxData{
		Type:    TypeTransfer,
		ChainID: gChainID,
		From:    util.Signer1.Address(),
		To:      &to,
		Amount:  big.NewInt(1),
		Data:    GenJSONData(1024 * 255),
	}
	assert.NoError(tx.SyntacticVerify())
	data = tx.Bytes()
	assert.Equal(data, tx.UnsignedBytes())

	tx2 := &TxData{}
	assert.NoError(tx2.Unmarshal(data))
	assert.NoError(tx2.SyntacticVerify())

	cbordata := tx2.Bytes()
	assert.Equal(data, cbordata)

	assert.NoError(tx.SignWith(util.Signer1))
	assert.NoError(tx.SyntacticVerify())
	assert.Equal(data, tx.UnsignedBytes())
	assert.NotEqual(data, tx.Bytes())
	assert.Equal(261200, len(tx.UnsignedBytes()))
	assert.Equal(261271, len(tx.Bytes()))
	assert.Equal(uint64(7774076), tx.Gas(), "a very big gas transaction")

	tx2 = &TxData{}
	assert.NoError(tx2.Unmarshal(tx.Bytes()))
	assert.NoError(tx2.SyntacticVerify())
	assert.Equal(data, tx2.UnsignedBytes())
	assert.Equal(tx.Bytes(), tx2.Bytes())

	txx := tx.ToTransaction()
	assert.NoError(txx.SyntacticVerify())
	assert.Equal(tx.ID(), txx.ID)
	assert.Equal(tx.Gas(), txx.Gas())
	jsondata, err = json.Marshal(txx)
	assert.NoError(err)

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
	// fmt.Println(string(jsondata))
	assert.Equal(`{"type":"TypeTransfer","chainID":2357,"nonce":1,"gasTip":0,"gasFeeCap":1000,"from":"0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC","to":"0x44171C37Ff5D7B7bb8dcad5C81f16284A229e641","amount":1000000,"signatures":["7db3ec16b7970728f2d20d32d1640b5034f62aaca20480b645b32cd87594f5536b238186d4624c8fef63fcd7f442e31756f51710883792c38e952065df45c0dd00"],"id":"E7ML6WgNZowbGX63GfSA2u5niXSnLA61a1o8SgaumKz6n9qqH"}`, string(jsondata))

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

	assert.Equal(uint64(618), tx.Gas())

	tx3.GasFeeCap++
	assert.NoError(tx3.SyntacticVerify())
	assert.NotEqual(tx.ID, tx3.ID)
	assert.NotEqual(tx.Bytes(), tx3.Bytes())
	assert.NotEqual(tx.ShortID(), tx3.ShortID())

	signers, err := tx.Signers()
	assert.NoError(err)
	assert.Equal(util.EthIDs{util.Signer1.Address()}, signers)
	_, err = tx.ExSigners()
	assert.ErrorContains(err, `DeriveSigners error: empty data`)

	assert.False(tx.IsBatched())
	assert.False(tx.NeedApprove(nil, nil))
	assert.True(tx.NeedApprove(&constants.GenesisAccount, nil))
	assert.True(tx.NeedApprove(&constants.GenesisAccount, TxTypes{TypeTransfer}))
	assert.False(tx.NeedApprove(&constants.GenesisAccount, TxTypes{TypeUpdateAccountInfo}))
}

func TestTxs(t *testing.T) {
	assert := assert.New(t)

	testTx := (&TxData{
		Type:    TypeTest,
		ChainID: gChainID,
	}).ToTransaction()

	assert.NoError(testTx.SyntacticVerify())
	assert.Equal(81, testTx.BytesSize())

	_, err := NewBatchTx(testTx)
	assert.ErrorContains(err, "NewBatchTx error: not batch transactions")

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
	assert.Equal(len(testTx.Bytes())+len(tx1.Bytes())+len(tx2.Bytes()), txs.BytesSize())

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
	assert.Equal(stx3.Gas(), btx.Gas())
	assert.Equal(
		len(stx0.Bytes())+len(stx1.Bytes())+len(stx2.Bytes())+len(stx3.Bytes())+len(stx4.Bytes()),
		btx.BytesSize())
	assert.Equal(stx3.priority, btx.priority)
	assert.Equal(uint64(436), stx0.priority)
	assert.Equal(uint64(1193), stx1.priority)
	assert.Equal(uint64(1795), stx2.priority)
	assert.Equal(uint64(8791), stx3.priority)
	assert.Equal(uint64(7915), stx4.priority)

	txs := Txs{stx0, stx1, stx2, stx3, stx4}
	txs.Sort()
	assert.Equal(stx3.ID, txs[0].ID, "because of high priority")
	assert.Equal(stx2.ID, txs[1].ID, "because of low nonce than stx4 from the same sender")
	assert.Equal(stx4.ID, txs[2].ID)
	assert.Equal(stx0.ID, txs[3].ID)
	assert.Equal(stx1.ID, txs[4].ID)

	// sort again should not change the order
	txs = Txs{stx4, stx3, stx2, stx1, stx0}
	txs.Sort()
	assert.Equal(stx3.ID, txs[0].ID)
	assert.Equal(stx2.ID, txs[1].ID)
	assert.Equal(stx4.ID, txs[2].ID)
	assert.Equal(stx0.ID, txs[3].ID)
	assert.Equal(stx1.ID, txs[4].ID)
	assert.Equal(stx3.priority, btx.priority)
	assert.Equal(uint64(436), stx0.priority)
	assert.Equal(uint64(1193), stx1.priority)
	assert.Equal(uint64(1795), stx2.priority)
	assert.Equal(uint64(8791), stx3.priority)
	assert.Equal(uint64(7915), stx4.priority)

	tx0 := MustNewTestTx(s0, TypeTransfer, &to, nil)
	tx1 := MustNewTestTx(s1, TypeTransfer, &to, GenJSONData(1000))
	tx2 := MustNewTestTx(s2, TypeTransfer, &to, GenJSONData(1200))
	tx3 := MustNewTestTx(s3, TypeTransfer, &to, GenJSONData(1500))
	txs = Txs{tx0, tx1, tx2, tx3}
	txs.Sort()
	assert.Equal(tx3.ID, txs[0].ID)
	assert.Equal(tx2.ID, txs[1].ID)
	assert.Equal(tx1.ID, txs[2].ID)
	assert.Equal(tx0.ID, txs[3].ID)

	txs = append(txs, btx)
	txs.Sort()
	assert.Equal(tx3.ID, txs[0].ID)
	assert.Equal(btx.ID, txs[1].ID)
	assert.Equal(tx2.ID, txs[2].ID)
	assert.Equal(tx1.ID, txs[3].ID)
	assert.Equal(tx0.ID, txs[4].ID)

	// should keep the origin order in batch txs
	assert.Equal(stx0.ID, btx.batch[0].ID)
	assert.Equal(stx1.ID, btx.batch[1].ID)
	assert.Equal(stx2.ID, btx.batch[2].ID)
	assert.Equal(stx3.ID, btx.batch[3].ID)
	assert.Equal(stx4.ID, btx.batch[4].ID)

	// sort again should not change the order
	txs[1], txs[3] = txs[3], txs[1]
	txs.Sort()
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
}
