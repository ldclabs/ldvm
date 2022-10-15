// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package transactions

import (
	"math/big"
	"testing"

	"github.com/ldclabs/ldvm/constants"
	"github.com/ldclabs/ldvm/ld"
	"github.com/ldclabs/ldvm/ld/service"
	"github.com/ldclabs/ldvm/util"
	"github.com/ldclabs/ldvm/util/signer"
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
	assert.NoError(from.UpdateKeepers(ld.Uint16Ptr(1), &signer.Keys{signer.Signer1.Key()}, nil, nil))

	to, err := cs.LoadAccount(signer.Signer2.Key().Address())
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
	assert.ErrorContains(err, "no signatures")

	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypePunish,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      to.id,
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err,
		"invalid from, expected GenesisAccount, got 0x44171C37Ff5D7B7bb8Dcad5C81f16284A229E641")

	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypePunish,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      from.id,
		To:        &to.id,
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
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
	assert.NoError(ltx.SignWith(signer.Signer1))
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
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.ErrorContains(ltx.SyntacticVerify(), "nil \"to\" together with amount")

	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypePunish,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      from.id,
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
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
	assert.NoError(ltx.SignWith(signer.Signer1))
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
	assert.NoError(ltx.SignWith(signer.Signer1))
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
	assert.NoError(ltx.SignWith(signer.Signer1))
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
	assert.NoError(ltx.SignWith(signer.Signer1))
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
		"YWJjZGVmAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAACuCNU1 not found")
	cs.CheckoutAccounts()

	di := &ld.DataInfo{
		Version:   1,
		Threshold: 1,
		Keepers:   signer.Keys{signer.Signer2.Key()},
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
	assert.Equal(`{"tx":{"type":"TypePunish","chainID":2357,"nonce":0,"gasTip":100,"gasFeeCap":1000,"from":"0xFFfFFFfFfffFFfFFffFFFfFfFffFFFfffFfFFFff","data":{"id":"YWJjZGVmAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAACuCNU1","data":"Illegal content"}},"sigs":["UPUNmg0N7UxdpDwR4cwnnLGvuRLUBJ5x6lthPexKUjVGJ7Mn6a8KJ8Rq8xdVUFOeWxr6W8oASqZMebETWPysGAHDgCjD"],"id":"y1ilDeTEerfrXkeEwjUmP02og4hw8u3mDsnMFNr_wsP2sKNu"}`, string(jsondata))

	assert.NoError(cs.VerifyState())
}

func TestTxPunishNameServiceData(t *testing.T) {
	assert := assert.New(t)

	ctx := NewMockChainContext()
	cs := ctx.MockChainState()

	sender := signer.Signer1.Key().Address()
	recipient := signer.Signer2.Key().Address()

	nm, err := service.NameModel()
	assert.NoError(err)
	mi := &ld.ModelInfo{
		Name:      nm.Name(),
		Threshold: 1,
		Keepers:   signer.Keys{signer.Signer2.Key()},
		Schema:    nm.Schema(),
		ID:        ctx.ChainConfig().NameServiceID,
	}

	name := &service.Name{
		Name:       "ldc:to",
		Records:    []string{"ldc:to IN A 10.0.0.1"},
		Extensions: service.Extensions{},
	}
	assert.NoError(name.SyntacticVerify())

	input := &ld.TxUpdater{
		ModelID:   &mi.ID,
		Version:   1,
		Threshold: ld.Uint16Ptr(1),
		Keepers:   &signer.Keys{signer.Signer1.Key()},
		Data:      name.Bytes(),
		To:        &recipient,
		Expire:    100,
		Amount:    new(big.Int).SetUint64(constants.MilliLDC),
	}
	assert.NoError(input.SyntacticVerify())
	ltx := &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeCreateData,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      sender,
		To:        &recipient,
		Amount:    new(big.Int).SetUint64(constants.MilliLDC),
		Data:      input.Bytes(),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.ExSignWith(signer.Signer2))
	assert.NoError(ltx.SyntacticVerify())

	senderAcc := cs.MustAccount(sender)
	assert.NoError(senderAcc.Add(constants.NativeToken, new(big.Int).SetUint64(constants.LDC)))
	assert.NoError(cs.SaveModel(mi))

	ltx.Timestamp = 10
	itx, err := NewTx(ltx)
	assert.NoError(err)

	_, err = cs.LoadDataByName("ldc:to")
	assert.ErrorContains(err, `"ldc:to" not found`)
	assert.NoError(itx.Apply(ctx, cs))
	di, err := cs.LoadDataByName("ldc:to")
	assert.NoError(err)
	assert.Equal(mi.ID, di.ModelID)

	genesis := cs.MustAccount(constants.GenesisAccount)
	assert.NoError(genesis.UpdateKeepers(ld.Uint16Ptr(1), &signer.Keys{signer.Signer1.Key()}, nil, nil))
	genesis.Add(constants.NativeToken, new(big.Int).SetUint64(constants.LDC))

	input = &ld.TxUpdater{ID: &di.ID}
	assert.NoError(input.SyntacticVerify())
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypePunish,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      genesis.id,
		Data:      input.Bytes(),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	itx, err = NewTx(ltx)
	assert.NoError(err)
	assert.NoError(itx.Apply(ctx, cs))

	_, err = cs.LoadDataByName("ldc:to")
	assert.ErrorContains(err, `"ldc:to" not found`)

	di2, err := cs.LoadData(di.ID)
	assert.NoError(err)
	assert.Equal(uint64(0), di2.Version)
	assert.Equal(mi.ID, di2.ModelID)
	assert.Nil(di2.Payload)

	assert.NoError(cs.VerifyState())
}
