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

func TestTxUpdateModelInfo(t *testing.T) {
	assert := assert.New(t)

	// SyntacticVerify
	tx := &TxUpdateModelInfo{}
	assert.ErrorContains(tx.SyntacticVerify(), "nil pointer")
	_, err := tx.MarshalJSON()
	require.NoError(t, err)

	ctx := NewMockChainContext()
	cs := ctx.MockChainState()
	token := ld.MustNewToken("$LDC")
	owner := signer.Signer1.Key().Address()

	ltx := &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeUpdateModelInfo,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      owner,
	}}
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "no signatures")

	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeUpdateModelInfo,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      owner,
		To:        ids.GenesisAccount.Ptr(),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "invalid to, should be nil")

	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeUpdateModelInfo,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      owner,
		Token:     token.Ptr(),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "invalid token, should be nil")

	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeUpdateModelInfo,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      owner,
		Amount:    big.NewInt(1),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.ErrorContains(ltx.SyntacticVerify(), "nil \"to\" together with amount")

	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeUpdateModelInfo,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      owner,
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "invalid data")

	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeUpdateModelInfo,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      owner,
		Data:      []byte("你好👋"),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "cbor: unexpected following extraneous data")

	input := ld.TxUpdater{}
	assert.NoError(input.SyntacticVerify())
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeUpdateModelInfo,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      owner,
		Data:      input.Bytes(),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "invalid mid")

	input = ld.TxUpdater{ModelID: &ids.EmptyModelID}
	assert.NoError(input.SyntacticVerify())
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeUpdateModelInfo,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      owner,
		Data:      input.Bytes(),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "invalid mid")

	mid := ids.ModelID{'1', '2', '3', '4', '5', '6'}
	input = ld.TxUpdater{ModelID: &mid, Keepers: &signer.Keys{}}
	assert.ErrorContains(input.SyntacticVerify(), "invalid threshold")
	input = ld.TxUpdater{ModelID: &mid, Threshold: ld.Uint16Ptr(0)}
	assert.ErrorContains(input.SyntacticVerify(), "no keepers, threshold should be nil")
	input = ld.TxUpdater{ModelID: &mid, Threshold: ld.Uint16Ptr(1), Keepers: &signer.Keys{}}
	assert.ErrorContains(input.SyntacticVerify(), "invalid threshold, expected <= 0, got 1")

	mid = ids.ModelID{'1', '2', '3', '4', '5', '6'}
	input = ld.TxUpdater{ModelID: &mid}
	assert.NoError(input.SyntacticVerify())
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeUpdateModelInfo,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      owner,
		Data:      input.Bytes(),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "nothing to update")

	mid = ids.ModelID{'1', '2', '3', '4', '5', '6'}
	input = ld.TxUpdater{
		ModelID:  &mid,
		Approver: signer.Signer2.Key().Ptr(),
	}
	assert.NoError(input.SyntacticVerify())
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeUpdateModelInfo,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      owner,
		Data:      input.Bytes(),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	itx, err := NewTx(ltx)
	require.NoError(t, err)

	cs.CommitAccounts()
	assert.ErrorContains(itx.Apply(ctx, cs),
		"insufficient NativeLDC balance, expected 1339800, got 0")
	cs.CheckoutAccounts()
	ownerAcc := cs.MustAccount(owner)
	ownerAcc.Add(ids.NativeToken, new(big.Int).SetUint64(unit.LDC))
	cs.CommitAccounts()
	assert.ErrorContains(itx.Apply(ctx, cs),
		"MTIzNDU2AAAAAAAAAAAAAAAAAABQtLNs not found")
	cs.CheckoutAccounts()

	ipldm, err := service.ProfileModel()
	require.NoError(t, err)
	mi := &ld.ModelInfo{
		Name:      ipldm.Name(),
		Threshold: 1,
		Keepers:   signer.Keys{signer.Signer1.Key()},
		Schema:    ipldm.Schema(),
		ID:        mid,
	}
	assert.NoError(mi.SyntacticVerify())
	assert.NoError(cs.SaveModel(mi))
	assert.NoError(itx.Apply(ctx, cs))

	assert.Equal(ltx.Gas()*ctx.Price, itx.(*TxUpdateModelInfo).ldc.Balance().Uint64())
	assert.Equal(ltx.Gas()*100, itx.(*TxUpdateModelInfo).miner.Balance().Uint64())
	assert.Equal(unit.LDC-ltx.Gas()*(ctx.Price+100),
		ownerAcc.BalanceOfAll(ids.NativeToken).Uint64())
	assert.Equal(uint64(1), ownerAcc.Nonce())

	mi, err = cs.LoadModel(mid)
	require.NoError(t, err)
	require.NotNil(t, mi.Approver)
	assert.True(signer.Signer2.Key().Equal(mi.Approver))

	jsondata, err := itx.MarshalJSON()
	require.NoError(t, err)
	// fmt.Println(string(jsondata))
	assert.Equal(`{"tx":{"type":"TypeUpdateModelInfo","chainID":2357,"nonce":0,"gasTip":100,"gasFeeCap":1000,"from":"0x8db97c7cECe249C2b98bdc0226cc4C2A57bF52fc","data":{"mid":"MTIzNDU2AAAAAAAAAAAAAAAAAABQtLNs","approver":"RBccN_9de3u43K1cgfFihKIp5kE1lmGG"}},"sigs":["RUuVcEakQT_mucfMBvnWss53sKCleyNrZtlmkX6OKrtvdj6ez5JVsS4OmnwT-CxQBHM98KydOCZhmNRlsBRfoQD6iQKn"],"id":"n_HYkPVptSXKQltq2Qc9kYK89ddJwjngjVGRr_FOMd-lIYwO"}`, string(jsondata))

	assert.NoError(cs.VerifyState())

	// approver sign and clear approver
	input = ld.TxUpdater{
		ModelID:   &mid,
		Approver:  &signer.Key{},
		Threshold: ld.Uint16Ptr(1),
		Keepers:   &signer.Keys{signer.Signer1.Key(), signer.Signer2.Key()},
	}
	assert.NoError(input.SyntacticVerify())
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeUpdateModelInfo,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     1,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      owner,
		Data:      input.Bytes(),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	itx, err = NewTx(ltx)
	require.NoError(t, err)
	cs.CommitAccounts()
	assert.ErrorContains(itx.Apply(ctx, cs), "invalid signature for approver")
	cs.CheckoutAccounts()

	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeUpdateModelInfo,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     1,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      owner,
		Data:      input.Bytes(),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1, signer.Signer2))
	assert.NoError(ltx.SyntacticVerify())
	itx, err = NewTx(ltx)
	require.NoError(t, err)
	assert.NoError(itx.Apply(ctx, cs))

	mi, err = cs.LoadModel(mid)
	require.NoError(t, err)
	assert.Nil(mi.Approver)
	assert.Equal(signer.Keys{signer.Signer1.Key(), signer.Signer2.Key()}, mi.Keepers)

	// check SatisfySigningPlus
	input = ld.TxUpdater{
		ModelID:   &mid,
		Threshold: ld.Uint16Ptr(0),
		Keepers:   &signer.Keys{signer.Signer2.Key()},
	}
	assert.NoError(input.SyntacticVerify())
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeUpdateModelInfo,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     2,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      owner,
		Data:      input.Bytes(),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	itx, err = NewTx(ltx)
	require.NoError(t, err)

	cs.CommitAccounts()
	assert.ErrorContains(itx.Apply(ctx, cs), "invalid signatures for keepers")
	cs.CheckoutAccounts()

	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeUpdateModelInfo,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     2,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      owner,
		Data:      input.Bytes(),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1, signer.Signer2))
	assert.NoError(ltx.SyntacticVerify())
	itx, err = NewTx(ltx)
	require.NoError(t, err)
	assert.NoError(itx.Apply(ctx, cs))

	mi, err = cs.LoadModel(mid)
	require.NoError(t, err)
	assert.Nil(mi.Approver)
	assert.Equal(uint16(0), mi.Threshold)
	assert.Equal(signer.Keys{signer.Signer2.Key()}, mi.Keepers)

	assert.NoError(cs.VerifyState())
}
