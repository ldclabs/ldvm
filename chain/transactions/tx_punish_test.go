// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package transactions

import (
	"math/big"
	"testing"

	"github.com/ldclabs/ldvm/constants"
	"github.com/ldclabs/ldvm/ld"
	"github.com/ldclabs/ldvm/util"
	"github.com/stretchr/testify/assert"
)

func TestTxPunish(t *testing.T) {
	assert := assert.New(t)

	// SyntacticVerify
	tx := &TxPunish{}
	assert.ErrorContains(tx.SyntacticVerify(), "nil pointer")
	_, err := tx.MarshalJSON()
	assert.NoError(err)

	ctx := NewMockChainContext()
	cs := ctx.MockChainState()

	from := cs.MustAccount(constants.GenesisAccount)
	assert.NoError(err)
	singer1 := util.Signer1.Address()
	assert.NoError(from.UpdateKeepers(ld.Uint16Ptr(1), &util.EthIDs{singer1}, nil, nil))

	to, err := cs.LoadAccount(util.Signer2.Address())
	assert.NoError(err)

	ltx := &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypePunish,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      to.id,
	}}
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "DeriveSigners error: no signature")

	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypePunish,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      to.id,
	}}
	assert.NoError(ltx.SignWith(util.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err,
		"invalid from, expected GenesisAccount, got 0x44171C37Ff5D7B7bb8dcad5C81f16284A229e641")

	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypePunish,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      from.id,
		To:        &to.id,
	}}
	assert.NoError(ltx.SignWith(util.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "invalid to, should be nil")

	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypePunish,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      from.id,
		Token:     &constants.NativeToken,
	}}
	assert.NoError(ltx.SignWith(util.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "invalid token, should be nil")

	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypePunish,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      from.id,
		Amount:    big.NewInt(1),
	}}
	assert.NoError(ltx.SignWith(util.Signer1))
	assert.ErrorContains(ltx.SyntacticVerify(), "nil \"to\" together with amount")

	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypePunish,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      from.id,
	}}
	assert.NoError(ltx.SignWith(util.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "invalid data")

	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypePunish,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      from.id,
		Data:      []byte("你好👋"),
	}}
	assert.NoError(ltx.SignWith(util.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "cbor: unexpected following extraneous data")

	input := ld.TxUpdater{}
	assert.NoError(input.SyntacticVerify())
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypePunish,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      from.id,
		Data:      input.Bytes(),
	}}
	assert.NoError(ltx.SignWith(util.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "invalid data id")

	input = ld.TxUpdater{ID: &util.DataIDEmpty}
	assert.NoError(input.SyntacticVerify())
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypePunish,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      from.id,
		Data:      input.Bytes(),
	}}
	assert.NoError(ltx.SignWith(util.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "invalid data id")

	did := util.DataID{'a', 'b', 'c', 'd', 'e', 'f'}
	input = ld.TxUpdater{ID: &did, Data: []byte(`"Illegal content"`)}
	assert.NoError(input.SyntacticVerify())
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypePunish,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      from.id,
		Data:      input.Bytes(),
	}}
	assert.NoError(ltx.SignWith(util.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	itx, err := NewTx(ltx)
	assert.NoError(err)

	cs.CommitAccounts()
	assert.ErrorContains(itx.Apply(ctx, cs),
		"insufficient NativeLDC balance, expected 1056000, got 0")
	cs.CheckoutAccounts()

	from.Add(constants.NativeToken, new(big.Int).SetUint64(constants.LDC))
	cs.CommitAccounts()
	assert.ErrorContains(itx.Apply(ctx, cs),
		"jtZ1sadmr49B1MauiwmwEtuMte25vqm9kq1eHnccb3X1QRDAk not found")
	cs.CheckoutAccounts()

	di := &ld.DataInfo{
		Version:   1,
		Threshold: 1,
		Keepers:   util.EthIDs{util.Signer2.Address()},
		Payload:   []byte(`"test...."`),
		ID:        did,
	}
	assert.NoError(err)
	assert.NoError(di.SyntacticVerify())
	assert.NoError(cs.SaveData(di))
	assert.NoError(cs.SavePrevData(di))
	assert.NoError(itx.Apply(ctx, cs))

	assert.Equal(ltx.Gas()*ctx.Price,
		itx.(*TxPunish).ldc.Balance().Uint64())
	assert.Equal(ltx.Gas()*100,
		itx.(*TxPunish).miner.Balance().Uint64())
	assert.Equal(constants.LDC-ltx.Gas()*(ctx.Price+100),
		from.Balance().Uint64())
	assert.Equal(uint64(1), from.Nonce())

	di, err = cs.LoadData(did)
	assert.NoError(err)
	assert.Equal(uint64(0), di.Version)
	assert.Equal(input.Data, di.Payload)

	jsondata, err := itx.MarshalJSON()
	assert.NoError(err)
	// fmt.Println(string(jsondata))
	assert.Equal(`{"tx":{"type":"TypePunish","chainID":2357,"nonce":0,"gasTip":100,"gasFeeCap":1000,"from":"0xFFfFfFffFFfffFFfFFfFFFFFffFFFffffFfFFFfF","data":{"id":"jtZ1sadmr49B1MauiwmwEtuMte25vqm9kq1eHnccb3X1QRDAk","data":"Illegal content"}},"sigs":["50f50d9a0d0ded4c5da43c11e1cc279cb1afb912d4049e71ea5b613dec4a52354627b327e9af0a27c46af3175550539e5b1afa5bca004aa64c79b11358fcac1801"],"id":"2YZD34h4uf4fLGobKryT89FdYw7bRN5A4uPWgNBiV5wF3NMBer"}`, string(jsondata))

	assert.NoError(cs.VerifyState())
}
