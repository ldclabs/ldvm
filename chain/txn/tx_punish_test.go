// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package txn

import (
	"math/big"
	"testing"

	"github.com/ldclabs/ldvm/ids"
	"github.com/ldclabs/ldvm/ld"
	"github.com/ldclabs/ldvm/ld/service"
	"github.com/ldclabs/ldvm/signer"
	"github.com/ldclabs/ldvm/unit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTxPunish(t *testing.T) {
	assert := assert.New(t)

	// SyntacticVerify
	tx := &TxPunish{}
	assert.ErrorContains(tx.SyntacticVerify(), "nil pointer")
	_, err := tx.MarshalJSON()
	require.NoError(t, err)

	ctx := NewMockChainContext()
	cs := ctx.MockChainState()

	from := cs.MustAccount(ids.GenesisAccount)
	assert.NoError(from.UpdateKeepers(ld.Uint16Ptr(1), &signer.Keys{signer.Signer1.Key()}, nil, nil))

	to := cs.MustAccount(signer.Signer2.Key().Address())

	ltx := &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypePunish,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      to.ID(),
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
		From:      to.ID(),
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
		From:      from.ID(),
		To:        to.ID().Ptr(),
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
		From:      from.ID(),
		Token:     &ids.NativeToken,
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
		From:      from.ID(),
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
		From:      from.ID(),
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
		From:      from.ID(),
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
		From:      from.ID(),
		Data:      input.Bytes(),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "invalid data id")

	input = ld.TxUpdater{ID: &ids.EmptyDataID}
	assert.NoError(input.SyntacticVerify())
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypePunish,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      from.ID(),
		Data:      input.Bytes(),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "invalid data id")

	did := ids.DataID{'a', 'b', 'c', 'd', 'e', 'f'}
	input = ld.TxUpdater{ID: &did, Data: []byte(`"Illegal content"`)}
	assert.NoError(input.SyntacticVerify())
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypePunish,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      from.ID(),
		Data:      input.Bytes(),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	itx, err := NewTx(ltx)
	require.NoError(t, err)

	cs.CommitAccounts()
	assert.ErrorContains(itx.Apply(ctx, cs),
		"insufficient NativeLDC balance, expected 1056000, got 0")
	cs.CheckoutAccounts()

	from.Add(ids.NativeToken, new(big.Int).SetUint64(unit.LDC))
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
	require.NoError(t, err)
	assert.NoError(di.SyntacticVerify())
	assert.NoError(cs.SaveData(di))
	assert.NoError(cs.SavePrevData(di))
	assert.NoError(itx.Apply(ctx, cs))

	assert.Equal(ltx.Gas()*ctx.Price,
		itx.(*TxPunish).ldc.Balance().Uint64())
	assert.Equal(ltx.Gas()*100,
		itx.(*TxPunish).miner.Balance().Uint64())
	assert.Equal(unit.LDC-ltx.Gas()*(ctx.Price+100),
		from.Balance().Uint64())
	assert.Equal(uint64(1), from.Nonce())

	di, err = cs.LoadData(did)
	require.NoError(t, err)
	assert.Equal(uint64(0), di.Version)
	assert.Equal(input.Data, di.Payload)

	jsondata, err := itx.MarshalJSON()
	require.NoError(t, err)
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
	require.NoError(t, err)
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
		To:        recipient.Ptr(),
		Expire:    100,
		Amount:    new(big.Int).SetUint64(unit.MilliLDC),
	}
	assert.NoError(input.SyntacticVerify())
	ltx := &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeCreateData,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      sender,
		To:        recipient.Ptr(),
		Amount:    new(big.Int).SetUint64(unit.MilliLDC),
		Data:      input.Bytes(),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.ExSignWith(signer.Signer2))
	assert.NoError(ltx.SyntacticVerify())

	senderAcc := cs.MustAccount(sender)
	assert.NoError(senderAcc.Add(ids.NativeToken, new(big.Int).SetUint64(unit.LDC*2)))
	assert.NoError(cs.SaveModel(mi))

	ltx.Timestamp = 10
	itx, err := NewTx(ltx)
	require.NoError(t, err)

	_, err = cs.LoadDataByName("ldc:to")
	assert.ErrorContains(err, `"ldc:to" not found`)
	assert.NoError(itx.Apply(ctx, cs))
	di, err := cs.LoadDataByName("ldc:to")
	require.NoError(t, err)
	assert.Equal(mi.ID, di.ModelID)

	genesis := cs.MustAccount(ids.GenesisAccount)
	assert.NoError(genesis.UpdateKeepers(ld.Uint16Ptr(1), &signer.Keys{signer.Signer1.Key()}, nil, nil))
	genesis.Add(ids.NativeToken, new(big.Int).SetUint64(unit.LDC))

	input = &ld.TxUpdater{ID: &di.ID}
	assert.NoError(input.SyntacticVerify())
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypePunish,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      genesis.ID(),
		Data:      input.Bytes(),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	itx, err = NewTx(ltx)
	require.NoError(t, err)
	assert.NoError(itx.Apply(ctx, cs))

	_, err = cs.LoadDataByName("ldc:to")
	assert.ErrorContains(err, `"ldc:to" not found`)

	di2, err := cs.LoadData(di.ID)
	require.NoError(t, err)
	assert.Equal(uint64(0), di2.Version)
	assert.Equal(mi.ID, di2.ModelID)
	assert.Nil(di2.Payload)

	assert.NoError(cs.VerifyState())
}
