// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package ld

import (
	"encoding/json"
	"math/big"
	"testing"

	"github.com/ldclabs/ldvm/constants"
	"github.com/ldclabs/ldvm/util"
	"github.com/ldclabs/ldvm/util/signer"
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
	assert.ErrorContains(tx.SyntacticVerify(), "nil \"to\" together with amount")

	tx = &TxData{Type: TypeTransfer, ChainID: gChainID, Data: util.RawData{}}
	assert.ErrorContains(tx.SyntacticVerify(), "empty data")

	tx = &TxData{ChainID: gChainID}
	assert.NoError(tx.SyntacticVerify())

	tx = &TxData{
		Type:    TypeTransfer,
		ChainID: gChainID,
		From:    signer.Signer1.Key().Address(),
		To:      signer.Signer2.Key().Address().Ptr(),
		Amount:  big.NewInt(1),
	}
	assert.NoError(tx.SyntacticVerify())

	jsondata, err := json.Marshal(tx)
	assert.NoError(err)
	// fmt.Println(string(jsondata))
	assert.Equal(`{"type":"TypeTransfer","chainID":2357,"nonce":0,"gasTip":0,"gasFeeCap":0,"from":"0x8db97c7cECe249C2b98bdc0226cc4C2A57bF52fc","to":"0x44171C37Ff5D7B7bb8Dcad5C81f16284A229E641","amount":1}`, string(jsondata))

	tx2 := &TxData{}
	assert.NoError(tx2.Unmarshal(tx.Bytes()))
	assert.NoError(tx2.SyntacticVerify())
	assert.Equal(tx.Bytes(), tx2.Bytes())
}

func TestTransaction(t *testing.T) {
	assert := assert.New(t)

	var tx *Transaction
	assert.ErrorContains(tx.SyntacticVerify(), "nil pointer")

	tx = &Transaction{Tx: TxData{Type: TypeDeleteData + 1}}
	assert.ErrorContains(tx.SyntacticVerify(), "invalid type")

	tx = &Transaction{Tx: TxData{Type: TypeTransfer, ChainID: 1000}}
	assert.ErrorContains(tx.SyntacticVerify(), "invalid ChainID")

	tx = &Transaction{Tx: TxData{Type: TypeTransfer, ChainID: gChainID, Token: &util.TokenSymbol{'a', 'b', 'c'}}}
	assert.ErrorContains(tx.SyntacticVerify(), "invalid token symbol")

	tx = &Transaction{Tx: TxData{Type: TypeTransfer, ChainID: gChainID, Amount: big.NewInt(-1)}}
	assert.ErrorContains(tx.SyntacticVerify(), "invalid amount")

	tx = &Transaction{Tx: TxData{Type: TypeTransfer, ChainID: gChainID, Amount: big.NewInt(0)}}
	assert.ErrorContains(tx.SyntacticVerify(), "nil \"to\" together with amount")

	tx = &Transaction{Tx: TxData{Type: TypeTransfer, ChainID: gChainID, Data: util.RawData{}}}
	assert.ErrorContains(tx.SyntacticVerify(), "empty data")

	tx = &Transaction{Tx: TxData{Type: TypeTransfer, ChainID: gChainID}, Signatures: signer.Sigs{}}
	assert.ErrorContains(tx.SyntacticVerify(), "empty signatures")

	tx = &Transaction{Tx: TxData{Type: TypeTransfer, ChainID: gChainID}, ExSignatures: signer.Sigs{}}
	assert.ErrorContains(tx.SyntacticVerify(), "empty exSignatures")

	tx = &Transaction{Tx: TxData{
		Type:    TypeTransfer,
		ChainID: gChainID,
		From:    signer.Signer1.Key().Address(),
		To:      signer.Signer2.Key().Address().Ptr(),
		Amount:  big.NewInt(1),
		Data:    GenJSONData(1024 * 256),
	}}
	assert.ErrorContains(tx.SyntacticVerify(),
		"Transaction.SyntacticVerify: size too large, expected <= 262144, got 262228")

	tx = &Transaction{Tx: TxData{ChainID: gChainID}}
	assert.NoError(tx.SyntacticVerify())

	assert.NoError(tx.SignWith(signer.Signer1))
	assert.NoError(tx.SyntacticVerify())
	assert.Equal(44, len(tx.Tx.Bytes()))
	assert.Equal(119, len(tx.Bytes()))
	assert.Equal(uint64(436), tx.Gas(), "a very small gas transaction")

	tx = &Transaction{Tx: TxData{
		Type:    TypeTransfer,
		ChainID: gChainID,
		From:    signer.Signer1.Key().Address(),
		To:      signer.Signer2.Key().Address().Ptr(),
		Amount:  big.NewInt(1),
		Data:    GenJSONData(1024 * 255),
	}}

	assert.NoError(tx.SignWith(signer.Signer1))
	assert.NoError(tx.SyntacticVerify())
	assert.Equal(261200, len(tx.Tx.Bytes()))
	assert.Equal(261275, len(tx.Bytes()))
	assert.Equal(uint64(7774227), tx.Gas(), "a very big gas transaction")

	tx = &Transaction{
		Tx: TxData{
			Type:      TypeTransfer,
			ChainID:   gChainID,
			Nonce:     1,
			GasTip:    0,
			GasFeeCap: 1000,
			From:      signer.Signer1.Key().Address(),
			To:        signer.Signer2.Key().Address().Ptr(),
			Amount:    big.NewInt(1_000_000),
		},
	}
	assert.NoError(tx.SignWith(signer.Signer1))
	assert.NoError(tx.SyntacticVerify())
	assert.Equal(len(tx.Bytes()), tx.BytesSize())
	assert.Equal(uint64(638), tx.Gas())

	approver := signer.Key(constants.GenesisAccount[:])
	assert.False(tx.IsBatched())
	assert.False(tx.needApprove(nil, nil))
	assert.False(tx.needApprove(signer.Key{}, nil))
	assert.True(tx.needApprove(approver, nil))
	assert.True(tx.needApprove(approver, TxTypes{TypeTransfer}))
	assert.False(tx.needApprove(approver, TxTypes{TypeUpdateAccountInfo}))

	jsondata, err := json.Marshal(tx)
	assert.NoError(err)
	// fmt.Println(string(jsondata))
	assert.Equal(`{"tx":{"type":"TypeTransfer","chainID":2357,"nonce":1,"gasTip":0,"gasFeeCap":1000,"from":"0x8db97c7cECe249C2b98bdc0226cc4C2A57bF52fc","to":"0x44171C37Ff5D7B7bb8Dcad5C81f16284A229E641","amount":1000000},"sigs":["fbPsFreXByjy0g0y0WQLUDT2KqyiBIC2RbMs2HWU9VNrI4GG1GJMj-9j_Nf0QuMXVvUXEIg3ksOOlSBl30XA3QAgiCJt"],"id":"aLokjgaVT95weTdJmhe2T1VjnvqfqaDNx7JHtRuo8TAsHAps"}`, string(jsondata))

	tx1 := tx.Copy()
	assert.NoError(tx1.SyntacticVerify())
	assert.Equal(tx.Tx.Bytes(), tx1.Tx.Bytes())

	tx1 = tx.Copy()
	tx1.Tx.Data = []byte(`"Hello, world!"`)
	assert.Nil(tx.Tx.Data)
	assert.NotEqual(tx.Tx.Bytes(), tx1.Tx.Bytes())
	jsondata, err = json.Marshal(tx1)
	assert.NoError(err)
	// fmt.Println(string(jsondata))
	assert.Equal(`{"tx":{"type":"TypeTransfer","chainID":2357,"nonce":1,"gasTip":0,"gasFeeCap":1000,"from":"0x8db97c7cECe249C2b98bdc0226cc4C2A57bF52fc","to":"0x44171C37Ff5D7B7bb8Dcad5C81f16284A229E641","amount":1000000,"data":"Hello, world!"},"sigs":["fbPsFreXByjy0g0y0WQLUDT2KqyiBIC2RbMs2HWU9VNrI4GG1GJMj-9j_Nf0QuMXVvUXEIg3ksOOlSBl30XA3QAgiCJt"],"id":"aLokjgaVT95weTdJmhe2T1VjnvqfqaDNx7JHtRuo8TAsHAps"}`, string(jsondata))

	tx2 := &Transaction{}
	assert.NoError(tx2.Unmarshal(tx.Bytes()))
	assert.NoError(tx2.SyntacticVerify())
	assert.Equal(tx.ID, tx2.ID)
	assert.Equal(tx.Bytes(), tx2.Bytes())

	assert.Equal(uint64(638), tx2.Gas())

	tx2.Tx.GasFeeCap++
	assert.NoError(tx2.SyntacticVerify())
	assert.NotEqual(tx.ID, tx2.ID)
	assert.NotEqual(tx.Bytes(), tx2.Bytes())
}

func TestTxs(t *testing.T) {
	assert := assert.New(t)

	testTx := (&TxData{
		Type:    TypeTest,
		ChainID: gChainID,
	}).ToTransaction()

	assert.NoError(testTx.SyntacticVerify())
	assert.Equal(48, testTx.BytesSize())

	_, err := NewBatchTx(testTx)
	assert.ErrorContains(err, "NewBatchTx: not batch transactions")

	tx1 := &Transaction{Tx: TxData{
		Type:      TypeTransfer,
		ChainID:   gChainID,
		Nonce:     1,
		GasTip:    0,
		GasFeeCap: 1000,
		From:      signer.Signer1.Key().Address(),
		To:        signer.Signer2.Key().Address().Ptr(),
		Amount:    big.NewInt(1_000_000),
	}}
	assert.NoError(tx1.SignWith(signer.Signer1))

	tx2 := &Transaction{Tx: TxData{
		Type:      TypeTransfer,
		ChainID:   gChainID,
		Nonce:     2,
		GasTip:    0,
		GasFeeCap: 1000,
		From:      signer.Signer1.Key().Address(),
		To:        signer.Signer2.Key().Address().Ptr(),
		Amount:    big.NewInt(1_000_000),
		Data:      []byte(`"👋"`),
	}}
	assert.NoError(tx2.SignWith(signer.Signer1))

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

	to := signer.Signer1.Key().Address()
	s0 := signer.NewSigner()
	s1 := signer.NewSigner()
	s2 := signer.NewSigner()
	s3 := signer.NewSigner()

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
	assert.Equal(uint64(455), stx0.priority)
	assert.Equal(uint64(1216), stx1.priority)
	assert.Equal(uint64(1820), stx2.priority)
	assert.Equal(uint64(8826), stx3.priority)
	assert.Equal(uint64(7949), stx4.priority)

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
	assert.Equal(uint64(455), stx0.priority)
	assert.Equal(uint64(1216), stx1.priority)
	assert.Equal(uint64(1820), stx2.priority)
	assert.Equal(uint64(8826), stx3.priority)
	assert.Equal(uint64(7949), stx4.priority)

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
