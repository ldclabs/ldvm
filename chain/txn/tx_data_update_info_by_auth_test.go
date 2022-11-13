// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package txn

import (
	"math/big"
	"testing"

	"github.com/ldclabs/ldvm/ids"
	"github.com/ldclabs/ldvm/ld"
	"github.com/ldclabs/ldvm/signer"
	"github.com/ldclabs/ldvm/unit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTxUpdateDataInfoByAuth(t *testing.T) {
	assert := assert.New(t)

	// SyntacticVerify
	tx := &TxUpdateDataInfoByAuth{}
	assert.ErrorContains(tx.SyntacticVerify(), "nil pointer")
	_, err := tx.MarshalJSON()
	require.NoError(t, err)

	ctx := NewMockChainContext()
	cs := ctx.MockChainState()
	token := ld.MustNewToken("$LDC")

	buyer := signer.Signer1.Key().Address()
	owner := signer.Signer2.Key().Address()

	ltx := &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeUpdateDataInfoByAuth,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      buyer,
	}}
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "no signatures")

	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeUpdateDataInfoByAuth,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      buyer,
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "nil to")

	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeUpdateDataInfoByAuth,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      buyer,
		To:        &owner,
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "invalid data")

	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeUpdateDataInfoByAuth,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      buyer,
		To:        &owner,
		Data:      []byte("你好👋"),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "no exSignatures")

	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeUpdateDataInfoByAuth,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      buyer,
		To:        &owner,
		Data:      []byte("你好👋"),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.ExSignWith(signer.Signer2))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "cbor: unexpected following extraneous data")

	input := &ld.TxUpdater{}
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeUpdateDataInfoByAuth,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      buyer,
		To:        &owner,
		Data:      input.Bytes(),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.ExSignWith(signer.Signer2))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "invalid data id")

	input = &ld.TxUpdater{ID: &ids.EmptyDataID}
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeUpdateDataInfoByAuth,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      buyer,
		To:        &owner,
		Data:      input.Bytes(),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.ExSignWith(signer.Signer2))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "invalid data id")

	did := ids.DataID{1, 2, 3, 4}
	input = &ld.TxUpdater{ID: &did}
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeUpdateDataInfoByAuth,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      buyer,
		To:        &owner,
		Data:      input.Bytes(),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.ExSignWith(signer.Signer2))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "invalid data version")

	input = &ld.TxUpdater{ID: &did, Version: 1,
		Keepers:   &signer.Keys{signer.Signer1.Key()},
		Threshold: ld.Uint16Ptr(1)}
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeUpdateDataInfoByAuth,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      buyer,
		To:        &owner,
		Data:      input.Bytes(),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.ExSignWith(signer.Signer2))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "invalid keepers, should be nil")

	sig := make(signer.Sig, 65)
	input = &ld.TxUpdater{
		ID: &did, Version: 1,
		Sig: &sig,
		SigClaims: &ld.SigClaims{
			Issuer:     ids.DataID{1, 2, 3, 4},
			Subject:    ids.DataID{5, 6, 7, 8},
			Audience:   ld.RawModelID,
			Expiration: 100,
			IssuedAt:   1,
			CWTID:      ids.ID32{9, 10, 11, 12},
		},
	}
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeUpdateDataInfoByAuth,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      buyer,
		To:        &owner,
		Data:      input.Bytes(),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.ExSignWith(signer.Signer2))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "invalid sigClaims, should be nil")

	input = &ld.TxUpdater{ID: &did, Version: 1, Approver: &signer.Key{}}
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeUpdateDataInfoByAuth,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      buyer,
		To:        &owner,
		Data:      input.Bytes(),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.ExSignWith(signer.Signer2))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "invalid approver, should be nil")

	input = &ld.TxUpdater{ID: &did, Version: 1,
		ApproveList: &ld.TxTypes{ld.TypeUpdateDataInfoByAuth}}
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeUpdateDataInfoByAuth,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      buyer,
		To:        &owner,
		Data:      input.Bytes(),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.ExSignWith(signer.Signer2))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "invalid approveList, should be nil")

	input = &ld.TxUpdater{ID: &did, Version: 1}
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeUpdateDataInfoByAuth,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      buyer,
		To:        &owner,
		Data:      input.Bytes(),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.ExSignWith(signer.Signer2))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "nil to")

	input = &ld.TxUpdater{ID: &did, Version: 1, To: &owner}
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeUpdateDataInfoByAuth,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      buyer,
		To:        ids.GenesisAccount.Ptr(),
		Data:      input.Bytes(),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.ExSignWith(signer.Signer2))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err,
		"invalid to, expected 0x44171C37Ff5D7B7bb8Dcad5C81f16284A229E641, got 0xFFfFFFfFfffFFfFFffFFFfFfFffFFFfffFfFFFff")

	input = &ld.TxUpdater{ID: &did, Version: 1, To: &owner}
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeUpdateDataInfoByAuth,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      buyer,
		To:        &owner,
		Data:      input.Bytes(),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.ExSignWith(signer.Signer2))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "nil amount")

	input = &ld.TxUpdater{ID: &did, Version: 1, To: &owner,
		Amount: new(big.Int).SetUint64(unit.MilliLDC)}
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeUpdateDataInfoByAuth,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      buyer,
		To:        &owner,
		Amount:    new(big.Int).SetUint64(1),
		Data:      input.Bytes(),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.ExSignWith(signer.Signer2))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "invalid amount, expected 1000000, got 1")

	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeUpdateDataInfoByAuth,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      buyer,
		To:        &owner,
		Token:     token.Ptr(),
		Amount:    new(big.Int).SetUint64(unit.MilliLDC),
		Data:      input.Bytes(),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.ExSignWith(signer.Signer2))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "invalid token, expected NativeToken, got $LDC")

	input = &ld.TxUpdater{ID: &did, Version: 1, To: &owner,
		Amount: new(big.Int).SetUint64(unit.MilliLDC), Token: &token}

	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeUpdateDataInfoByAuth,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      buyer,
		To:        &owner,
		Amount:    new(big.Int).SetUint64(unit.MilliLDC),
		Data:      input.Bytes(),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.ExSignWith(signer.Signer2))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "invalid token, expected $LDC, got NativeLDC")

	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeUpdateDataInfoByAuth,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      buyer,
		To:        &owner,
		Token:     token.Ptr(),
		Amount:    new(big.Int).SetUint64(unit.MilliLDC),
		Data:      input.Bytes(),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.ExSignWith(signer.Signer2))
	assert.NoError(ltx.SyntacticVerify())
	itx, err := NewTx(ltx)
	require.NoError(t, err)

	cs.CommitAccounts()
	assert.ErrorContains(itx.Apply(ctx, cs),
		"insufficient NativeLDC balance, expected 2099900, got 0")
	cs.CheckoutAccounts()

	buyerAcc := cs.MustAccount(buyer)
	buyerAcc.Add(ids.NativeToken, new(big.Int).SetUint64(unit.LDC))

	cs.CommitAccounts()
	assert.ErrorContains(itx.Apply(ctx, cs),
		"insufficient $LDC balance, expected 1000000, got 0")
	cs.CheckoutAccounts()
	buyerAcc.Add(token, new(big.Int).SetUint64(unit.LDC))
	cs.CommitAccounts()
	assert.ErrorContains(itx.Apply(ctx, cs),
		"no keepers on sender account")
	cs.CheckoutAccounts()

	assert.NoError(buyerAcc.UpdateKeepers(ld.Uint16Ptr(1), &signer.Keys{signer.Signer1.Key()}, nil, nil))

	cs.CommitAccounts()
	assert.ErrorContains(itx.Apply(ctx, cs),
		"AQIDBAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAACs148t not found")
	cs.CheckoutAccounts()

	di := &ld.DataInfo{
		ModelID:   ld.RawModelID,
		Version:   2,
		Threshold: 1,
		Keepers:   signer.Keys{signer.Signer2.Key()},
		Payload:   []byte(`42`),
		Approver:  signer.Signer1.Key(),
		ID:        did,
	}
	assert.NoError(di.SyntacticVerify())
	assert.NoError(cs.SaveData(di))
	cs.CommitAccounts()
	assert.ErrorContains(itx.Apply(ctx, cs), "invalid version, expected 2, got 1")
	cs.CheckoutAccounts()

	input = &ld.TxUpdater{ID: &did, Version: 2, To: &owner,
		Amount: new(big.Int).SetUint64(unit.MilliLDC), Token: &token}
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeUpdateDataInfoByAuth,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      buyer,
		To:        &owner,
		Token:     token.Ptr(),
		Amount:    new(big.Int).SetUint64(unit.MilliLDC),
		Data:      input.Bytes(),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.ExSignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	itx, err = NewTx(ltx)
	require.NoError(t, err)
	cs.CommitAccounts()
	assert.ErrorContains(itx.Apply(ctx, cs), "invalid exSignatures for data keepers")
	cs.CheckoutAccounts()

	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeUpdateDataInfoByAuth,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      buyer,
		To:        &owner,
		Token:     token.Ptr(),
		Amount:    new(big.Int).SetUint64(unit.MilliLDC),
		Data:      input.Bytes(),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.ExSignWith(signer.Signer2))
	assert.NoError(ltx.SyntacticVerify())
	itx, err = NewTx(ltx)
	require.NoError(t, err)
	cs.CommitAccounts()
	assert.ErrorContains(itx.Apply(ctx, cs), "invalid exSignature for data approver")
	cs.CheckoutAccounts()

	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeUpdateDataInfoByAuth,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      buyer,
		To:        &owner,
		Token:     token.Ptr(),
		Amount:    new(big.Int).SetUint64(unit.MilliLDC),
		Data:      input.Bytes(),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.ExSignWith(signer.Signer1, signer.Signer2))
	assert.NoError(ltx.SyntacticVerify())

	itx, err = NewTx(ltx)
	require.NoError(t, err)
	assert.NoError(itx.Apply(ctx, cs))

	buyerGas := ltx.Gas()
	assert.Equal(buyerGas*ctx.Price,
		itx.(*TxUpdateDataInfoByAuth).ldc.Balance().Uint64())
	assert.Equal(buyerGas*100,
		itx.(*TxUpdateDataInfoByAuth).miner.Balance().Uint64())
	assert.Equal(unit.LDC-buyerGas*(ctx.Price+100),
		buyerAcc.BalanceOfAll(ids.NativeToken).Uint64())
	assert.Equal(unit.LDC-unit.MilliLDC,
		buyerAcc.BalanceOf(token).Uint64())
	assert.Equal(unit.MilliLDC,
		itx.(*TxUpdateDataInfoByAuth).to.BalanceOf(token).Uint64())
	assert.Equal(uint64(1), buyerAcc.Nonce())

	di2, err := cs.LoadData(di.ID)
	require.NoError(t, err)
	assert.Equal(di.Version+1, di2.Version)
	assert.Equal(uint16(1), di2.Threshold)
	assert.Equal(signer.Keys{signer.Signer1.Key()}, di2.Keepers)
	assert.Equal(di.Payload, di2.Payload)
	assert.Nil(di2.Sig)
	assert.Nil(di2.Approver)
	assert.Nil(di2.ApproveList)

	jsondata, err := itx.MarshalJSON()
	require.NoError(t, err)
	// fmt.Println(string(jsondata))
	assert.Equal(`{"tx":{"type":"TypeUpdateDataInfoByAuth","chainID":2357,"nonce":0,"gasTip":100,"gasFeeCap":1000,"from":"0x8db97c7cECe249C2b98bdc0226cc4C2A57bF52fc","to":"0x44171C37Ff5D7B7bb8Dcad5C81f16284A229E641","token":"$LDC","amount":1000000,"data":{"id":"AQIDBAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAACs148t","version":2,"token":"$LDC","to":"0x44171C37Ff5D7B7bb8Dcad5C81f16284A229E641","amount":1000000}},"sigs":["jIwbZj66JDXhroiCUW7Tc4qLLFsXM2Z8Q9ZTedRIgntDRhrK1NIafbBirpC0MVBm5AsODBbd-mmSDnIvE30wFwC6g34A"],"exSigs":["GyB7Ag9nn-wXjmQwlg9YYm61X1a7bgVjUfNbPbNOnLdz6abXIBdOLG6Bc40RyNMrXS173wjy5cP1mIJRgA6vUQDGgC_s","6J2VngrdbCekT7SPUDC93WA9vTKYxtvrKBVpLCBnxvJckCBmTr4dlBANCmWH4OxDWUiTE_yItuStFEz_YfWKOAAcS6UV"],"id":"P2sJmW5wOkP6j6UCOATAHtAoFj8bLdhUIvGIuWUgHQ39JV_y"}`, string(jsondata))

	assert.NoError(cs.VerifyState())
}
