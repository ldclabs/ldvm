// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package transactions

import (
	"math/big"
	"testing"

	"github.com/ldclabs/ldvm/constants"
	"github.com/ldclabs/ldvm/ld"
	"github.com/ldclabs/ldvm/util"
	"github.com/ldclabs/ldvm/util/signer"
	"github.com/stretchr/testify/assert"
)

func TestTxDestroyToken(t *testing.T) {
	assert := assert.New(t)

	// SyntacticVerify
	tx := &TxDestroyToken{}
	assert.ErrorContains(tx.SyntacticVerify(), "nil pointer")
	_, err := tx.MarshalJSON()
	assert.NoError(err)

	ctx := NewMockChainContext()
	cs := ctx.MockChainState()
	token := ld.MustNewToken("$LDC")
	tokenid := token.Address()
	sender := signer.Signer1.Key().Address()
	recipient := signer.Signer2.Key().Address()

	ltx := &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeDestroyToken,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      tokenid,
	}}
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "no signatures")

	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeDestroyToken,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      tokenid,
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "nil to as pledge recipient")

	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeDestroyToken,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      tokenid,
		To:        &recipient,
		Token:     &token,
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "invalid token, should be nil")

	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeDestroyToken,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      tokenid,
		To:        &recipient,
		Amount:    big.NewInt(1),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "invalid amount, should be nil")

	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeDestroyToken,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      sender,
		To:        &recipient,
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "invalid token 0x8db97c7cECe249C2b98bdc0226cc4C2A57bF52fc")

	cfg := &ld.TxAccounter{
		Threshold: ld.Uint16Ptr(1),
		Keepers:   &signer.Keys{signer.Signer1.Key(), signer.Signer2.Key()},
		Amount:    new(big.Int).SetUint64(constants.LDC * 10),
	}
	testToken := NewAccount(util.Address(token))
	testToken.Init(ctx.FeeConfig().MinTokenPledge, 0, 0)
	assert.NoError(testToken.CreateToken(cfg))
	assert.Equal(false, testToken.valid(ld.TokenAccount))
	testToken.Add(constants.NativeToken, ctx.FeeConfig().MinTokenPledge)
	assert.Equal(true, testToken.valid(ld.TokenAccount))
	cs.AC[testToken.id] = testToken

	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeTransfer,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      testToken.id,
		To:        &recipient,
		Token:     &token,
		Amount:    new(big.Int).SetUint64(constants.LDC),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	itx, err := NewTx(ltx)
	assert.NoError(err)

	cs.CommitAccounts()
	assert.ErrorContains(itx.Apply(ctx, cs),
		"insufficient NativeLDC balance, expected 847000, got 0")
	cs.CheckoutAccounts()
	testToken.Add(constants.NativeToken, new(big.Int).SetUint64(constants.LDC))
	assert.NoError(itx.Apply(ctx, cs))

	recipientAcc := cs.MustAccount(recipient)

	tokenGas := ltx.Gas()
	assert.Equal(tokenGas*ctx.Price,
		itx.(*TxTransfer).ldc.Balance().Uint64())
	assert.Equal(tokenGas*100,
		itx.(*TxTransfer).miner.Balance().Uint64())
	assert.Equal(constants.LDC-tokenGas*(ctx.Price+100),
		testToken.Balance().Uint64())
	assert.Equal(constants.LDC*9,
		testToken.balanceOf(token).Uint64())
	assert.Equal(uint64(0),
		recipientAcc.Balance().Uint64())
	assert.Equal(constants.LDC,
		recipientAcc.balanceOf(token).Uint64())
	assert.Equal(uint64(1), testToken.Nonce())

	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeDestroyToken,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     1,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      testToken.id,
		To:        &recipient,
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	itx, err = NewTx(ltx)
	assert.NoError(err)

	cs.CommitAccounts()
	assert.ErrorContains(itx.Apply(ctx, cs), "invalid signature for keepers")
	cs.CheckoutAccounts()

	assert.NoError(ltx.SignWith(signer.Signer1, signer.Signer2))
	assert.NoError(ltx.SyntacticVerify())
	itx, err = NewTx(ltx)
	assert.NoError(err)

	cs.CommitAccounts()
	assert.ErrorContains(itx.Apply(ctx, cs),
		"some token in the use, maxTotalSupply expected 10000000000, got 9000000000")
	cs.CheckoutAccounts()

	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeTransfer,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      recipient,
		To:        &testToken.id,
		Token:     &token,
		Amount:    new(big.Int).SetUint64(constants.LDC),
	}}
	assert.NoError(ltx.SignWith(signer.Signer2))
	assert.NoError(ltx.SyntacticVerify())
	itx, err = NewTx(ltx)
	assert.NoError(err)
	recipientAcc.Add(constants.NativeToken, new(big.Int).SetUint64(constants.LDC))
	assert.NoError(itx.Apply(ctx, cs))

	recipientGas := ltx.Gas()
	assert.Equal((tokenGas+recipientGas)*ctx.Price,
		itx.(*TxTransfer).ldc.Balance().Uint64())
	assert.Equal((tokenGas+recipientGas)*100,
		itx.(*TxTransfer).miner.Balance().Uint64())
	assert.Equal(ctx.FeeConfig().MinTokenPledge.Uint64()+constants.LDC-tokenGas*(ctx.Price+100),
		testToken.balanceOfAll(constants.NativeToken).Uint64())
	assert.Equal(constants.LDC*10, testToken.balanceOf(token).Uint64())
	assert.Equal(constants.LDC-recipientGas*(ctx.Price+100),
		recipientAcc.Balance().Uint64())
	assert.Equal(uint64(0), recipientAcc.balanceOf(token).Uint64())

	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeDestroyToken,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     1,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      testToken.id,
		To:        &recipient,
	}}
	assert.NoError(ltx.SignWith(signer.Signer1, signer.Signer2))
	assert.NoError(ltx.SyntacticVerify())
	itx, err = NewTx(ltx)
	assert.NoError(err)
	assert.NoError(itx.Apply(ctx, cs))

	tokenGas += ltx.Gas()
	assert.Equal((tokenGas+recipientGas)*ctx.Price,
		itx.(*TxDestroyToken).ldc.Balance().Uint64())
	assert.Equal((tokenGas+recipientGas)*100,
		itx.(*TxDestroyToken).miner.Balance().Uint64())

	assert.Equal(uint64(0), testToken.Balance().Uint64())
	assert.Equal(uint64(0), testToken.balanceOfAll(constants.NativeToken).Uint64())
	assert.Equal(uint64(0), testToken.balanceOf(token).Uint64())
	assert.Equal(uint64(0), testToken.balanceOfAll(token).Uint64())
	assert.Equal(ctx.FeeConfig().MinTokenPledge.Uint64()+constants.LDC*2-(tokenGas+recipientGas)*(ctx.Price+100),
		recipientAcc.Balance().Uint64())
	assert.Equal(uint64(0), recipientAcc.balanceOf(token).Uint64())
	assert.Equal(uint64(0), recipientAcc.balanceOfAll(token).Uint64())

	assert.Equal(uint64(2), testToken.Nonce())
	assert.Equal(uint16(0), testToken.Threshold())
	assert.Equal(signer.Keys{}, testToken.Keepers())

	jsondata, err := itx.MarshalJSON()
	assert.NoError(err)
	// fmt.Println(string(jsondata))
	assert.Equal(`{"tx":{"type":"TypeDestroyToken","chainID":2357,"nonce":1,"gasTip":100,"gasFeeCap":1000,"from":"0x00000000000000000000000000000000244C4443","to":"0x44171C37Ff5D7B7bb8Dcad5C81f16284A229E641"},"sigs":["lHgSlVzFykfOwC9oQ6zQy_bBYA2M956P89SZ1Kbof79p22zGK2W41CvX3pCrvxnTQlMPH-qtkSPfpnJ857EC4wG6gsN8","09Sl-st2HTiztAZ3wApmvN_sdRFBFXrF419p_FAkxwRg0IWHLYzsiiJwhIkBJtl4XmLbg0ZWol4bPqcH3ev7ZAFMSGV0"],"id":"j_xIfHsXX7FqTRExLzyENfU4hqK4YpKl1EGHvdx_7_d-tz7e"}`, string(jsondata))

	assert.NoError(cs.VerifyState())
}

func TestTxDestroyTokenWithApproverAndLending(t *testing.T) {
	assert := assert.New(t)

	ctx := NewMockChainContext()
	cs := ctx.MockChainState()
	token := ld.MustNewToken("$LDC")
	tokenid := util.Address(token)

	sender := signer.Signer1.Key().Address()
	approver := signer.Signer2.Key()

	// CreateToken
	input := &ld.TxAccounter{
		Threshold:   ld.Uint16Ptr(1),
		Keepers:     &signer.Keys{signer.Signer1.Key()},
		Approver:    &approver,
		ApproveList: &ld.TxTypes{ld.TypeOpenLending, ld.TypeDestroyToken},
		Amount:      new(big.Int).SetUint64(constants.LDC * 10),
		Name:        "LDC Token",
	}
	assert.NoError(input.SyntacticVerify())
	ltx := &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeCreateToken,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      sender,
		To:        &tokenid,
		Amount:    ctx.FeeConfig().MinTokenPledge,
		Data:      input.Bytes(),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	itx, err := NewTx(ltx)
	assert.NoError(err)

	tokenAcc := cs.MustAccount(tokenid)
	senderAcc := cs.MustAccount(sender)
	senderAcc.Add(constants.NativeToken, ctx.FeeConfig().MinTokenPledge)
	senderAcc.Add(constants.NativeToken, new(big.Int).SetUint64(constants.LDC))
	assert.NoError(itx.Apply(ctx, cs))

	senderGas := ltx.Gas()
	assert.Equal(senderGas*ctx.Price,
		itx.(*TxCreateToken).ldc.Balance().Uint64())
	assert.Equal(senderGas*100,
		itx.(*TxCreateToken).miner.Balance().Uint64())
	assert.Equal(uint64(0),
		tokenAcc.Balance().Uint64())
	assert.Equal(ctx.FeeConfig().MinTokenPledge.Uint64(),
		tokenAcc.balanceOfAll(constants.NativeToken).Uint64())
	assert.Equal(constants.LDC*10,
		tokenAcc.balanceOf(token).Uint64())
	assert.Equal(constants.LDC-senderGas*(ctx.Price+100),
		senderAcc.Balance().Uint64())

	assert.Equal(uint16(1), tokenAcc.Threshold())
	assert.Equal(signer.Keys{signer.Signer1.Key()}, tokenAcc.Keepers())
	assert.Equal(approver, tokenAcc.ld.Approver)
	assert.Equal(ld.TxTypes{ld.TypeOpenLending, ld.TypeDestroyToken}, tokenAcc.ld.ApproveList)

	// OpenLending
	lcfg := &ld.LendingConfig{
		Token:           token,
		DailyInterest:   10,
		OverdueInterest: 10,
		MinAmount:       big.NewInt(1000),
		MaxAmount:       new(big.Int).SetUint64(constants.LDC),
	}
	assert.NoError(lcfg.SyntacticVerify())
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeOpenLending,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      tokenid,
		Data:      ld.MustMarshal(lcfg),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	itx, err = NewTx(ltx)
	assert.NoError(err)
	senderAcc.Add(constants.NativeToken, new(big.Int).SetUint64(constants.LDC))

	cs.CommitAccounts()
	assert.ErrorContains(itx.Apply(ctx, cs),
		"TxOpenLending.Apply: invalid signature for approver")
	cs.CheckoutAccounts()

	assert.NoError(ltx.SignWith(signer.Signer1, signer.Signer2))
	assert.NoError(ltx.SyntacticVerify())
	itx, err = NewTx(ltx)
	assert.NoError(err)

	cs.CommitAccounts()
	assert.ErrorContains(itx.Apply(ctx, cs),
		"insufficient NativeLDC balance, expected 2197800, got 0")
	cs.CheckoutAccounts()

	tokenAcc.Add(constants.NativeToken, new(big.Int).SetUint64(constants.LDC))
	assert.NoError(itx.Apply(ctx, cs))

	tokenGas := ltx.Gas()
	assert.Equal((tokenGas+senderGas)*ctx.Price,
		itx.(*TxOpenLending).ldc.Balance().Uint64())
	assert.Equal((tokenGas+senderGas)*100,
		itx.(*TxOpenLending).miner.Balance().Uint64())
	assert.NotNil(tokenAcc.ld.Lending)
	assert.Equal(uint64(1), tokenAcc.Nonce())

	// AddNonceTable
	ns := []uint64{cs.Timestamp() + 1, 1, 2, 3}
	ndData, err := util.MarshalCBOR(ns)
	assert.NoError(err)
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeAddNonceTable,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     1,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      tokenid,
		Data:      ndData,
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	ltx.Timestamp = cs.Timestamp()
	itx, err = NewTx(ltx)
	assert.NoError(err)
	assert.NoError(itx.Apply(ctx, cs))

	tokenGas += ltx.Gas()
	assert.Equal((tokenGas+senderGas)*ctx.Price,
		itx.(*TxAddNonceTable).ldc.Balance().Uint64())
	assert.Equal((tokenGas+senderGas)*100,
		itx.(*TxAddNonceTable).miner.Balance().Uint64())
	assert.Equal([]uint64{1, 2, 3}, tokenAcc.ld.NonceTable[cs.Timestamp()+1])
	assert.Equal(uint64(2), tokenAcc.Nonce())

	// Borrow
	tf := &ld.TxTransfer{
		Nonce:  3,
		From:   &tokenAcc.id,
		To:     &sender,
		Token:  &token,
		Amount: new(big.Int).SetUint64(constants.LDC),
		Expire: cs.Timestamp() + 1,
	}
	assert.NoError(tf.SyntacticVerify())
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeBorrow,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     1,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      sender,
		To:        &tokenid,
		Token:     &token,
		Data:      tf.Bytes(),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.ExSignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	ltx.Timestamp = cs.Timestamp()
	itx, err = NewTx(ltx)
	assert.NoError(err)
	assert.NoError(itx.Apply(ctx, cs))

	senderGas += ltx.Gas()
	assert.Equal((tokenGas+senderGas)*ctx.Price,
		itx.(*TxBorrow).ldc.Balance().Uint64())
	assert.Equal((tokenGas+senderGas)*100,
		itx.(*TxBorrow).miner.Balance().Uint64())
	assert.Equal([]uint64{1, 2}, tokenAcc.ld.NonceTable[cs.Timestamp()+1])
	assert.Equal(constants.LDC*9, tokenAcc.balanceOf(token).Uint64())
	assert.Equal(constants.LDC, senderAcc.balanceOf(token).Uint64())

	// DestroyToken
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeDestroyToken,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     2,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      tokenid,
		To:        &sender,
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	itx, err = NewTx(ltx)
	assert.NoError(err)

	cs.CommitAccounts()
	assert.ErrorContains(itx.Apply(ctx, cs),
		"TxDestroyToken.Apply: invalid signature for approver")
	cs.CheckoutAccounts()

	assert.NoError(ltx.SignWith(signer.Signer1, signer.Signer2))
	assert.NoError(ltx.SyntacticVerify())
	itx, err = NewTx(ltx)
	assert.NoError(err)
	cs.CommitAccounts()
	assert.ErrorContains(itx.Apply(ctx, cs),
		"TxDestroyToken.Apply: Account(0x00000000000000000000000000000000244C4443).DestroyToken: some token in the use, maxTotalSupply expected 10000000000, got 9000000000")
	cs.CheckoutAccounts()

	// TypeRepay
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeRepay,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     2,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      sender,
		To:        &tokenid,
		Token:     &token,
		Amount:    new(big.Int).SetUint64(constants.LDC),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	itx, err = NewTx(ltx)
	assert.NoError(err)
	assert.NoError(itx.Apply(ctx, cs))

	senderGas += ltx.Gas()
	assert.Equal((tokenGas+senderGas)*ctx.Price,
		itx.(*TxRepay).ldc.Balance().Uint64())
	assert.Equal((tokenGas+senderGas)*100,
		itx.(*TxRepay).miner.Balance().Uint64())
	assert.Equal(constants.LDC*10, tokenAcc.balanceOf(token).Uint64())
	assert.Equal(uint64(0), senderAcc.balanceOf(token).Uint64())

	// DestroyToken
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeDestroyToken,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     2,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      tokenid,
		To:        &sender,
	}}
	assert.NoError(ltx.SignWith(signer.Signer1, signer.Signer2))
	assert.NoError(ltx.SyntacticVerify())
	itx, err = NewTx(ltx)
	assert.NoError(err)
	assert.NoError(itx.Apply(ctx, cs))

	tokenGas += ltx.Gas()
	assert.Equal((tokenGas+senderGas)*ctx.Price,
		itx.(*TxDestroyToken).ldc.Balance().Uint64())
	assert.Equal((tokenGas+senderGas)*100,
		itx.(*TxDestroyToken).miner.Balance().Uint64())
	assert.Equal(uint64(3), tokenAcc.Nonce())
	assert.Equal(uint16(0), tokenAcc.Threshold())
	assert.Equal(signer.Keys{}, tokenAcc.Keepers())
	assert.Equal(make(map[uint64][]uint64), tokenAcc.ld.NonceTable)
	assert.Equal(ld.NativeAccount, tokenAcc.ld.Type)
	assert.Nil(tokenAcc.ld.Approver)
	assert.Nil(tokenAcc.ld.ApproveList)
	assert.Nil(tokenAcc.ld.Lending)
	assert.Nil(tokenAcc.ld.MaxTotalSupply)
	assert.Nil(tokenAcc.ld.Tokens[token.AsKey()])
	assert.Equal(uint64(0), tokenAcc.balanceOfAll(token).Uint64())
	assert.Equal(uint64(0), tokenAcc.balanceOfAll(constants.NativeToken).Uint64())

	jsondata, err := itx.MarshalJSON()
	assert.NoError(err)
	// fmt.Println(string(jsondata))
	assert.Equal(`{"tx":{"type":"TypeDestroyToken","chainID":2357,"nonce":2,"gasTip":100,"gasFeeCap":1000,"from":"0x00000000000000000000000000000000244C4443","to":"0x8db97c7cECe249C2b98bdc0226cc4C2A57bF52fc"},"sigs":["Qh_xHvlxao46OeFUCQHmDFr8X5Jky2fopWDjB2Q93AJNeyaAF5JDBKNygXWB1ooqOyGodMOQglzXGXMm9eiXFgHzyXKX","yjJ-_yYfuTgwJctXY8uoSs9wsRwARhCu9Nf_ibxvozB-lZZDWHCM9Pltfud-tv29MBLRtpbL9lFLH9tNKc6IqgDFCL71"],"id":"e6nJPIAV6ym0S60lk37AfOlzhYtHqp1hGt9trGBdoUYGYXJN"}`, string(jsondata))

	// DestroyToken again
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeDestroyToken,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     3,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      tokenid,
		To:        &sender,
	}}
	assert.NoError(ltx.SignWith(signer.Signer1, signer.Signer2))
	assert.NoError(ltx.SyntacticVerify())
	itx, err = NewTx(ltx)
	assert.NoError(err)
	assert.ErrorContains(itx.Apply(ctx, cs),
		"TxDestroyToken.Apply: invalid signatures for sender")

	assert.NoError(cs.VerifyState())
}
