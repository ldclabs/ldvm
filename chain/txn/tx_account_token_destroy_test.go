// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package txn

import (
	"math/big"
	"testing"

	"github.com/ldclabs/ldvm/chain/acct"
	"github.com/ldclabs/ldvm/ids"
	"github.com/ldclabs/ldvm/ld"
	"github.com/ldclabs/ldvm/signer"
	"github.com/ldclabs/ldvm/unit"
	"github.com/ldclabs/ldvm/util/encoding"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTxDestroyToken(t *testing.T) {
	assert := assert.New(t)

	// SyntacticVerify
	tx := &TxDestroyToken{}
	assert.ErrorContains(tx.SyntacticVerify(), "nil pointer")
	_, err := tx.MarshalJSON()
	require.NoError(t, err)

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
		To:        recipient.Ptr(),
		Token:     token.Ptr(),
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
		To:        recipient.Ptr(),
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
		To:        recipient.Ptr(),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "invalid token 0x8db97c7cECe249C2b98bdc0226cc4C2A57bF52fc")

	cfg := &ld.TxAccounter{
		Threshold: ld.Uint16Ptr(1),
		Keepers:   &signer.Keys{signer.Signer1.Key(), signer.Signer2.Key()},
		Amount:    new(big.Int).SetUint64(unit.LDC * 10),
	}
	testToken := acct.NewAccount(ids.Address(token))
	testToken.Init(big.NewInt(0), ctx.FeeConfig().MinTokenPledge, 0, 0)
	assert.NoError(testToken.CreateToken(cfg))
	assert.Equal(false, testToken.Valid(ld.TokenAccount))
	testToken.Add(ids.NativeToken, ctx.FeeConfig().MinTokenPledge)
	assert.Equal(true, testToken.Valid(ld.TokenAccount))
	cs.AC[testToken.ID()] = testToken

	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeTransfer,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      testToken.ID(),
		To:        recipient.Ptr(),
		Token:     token.Ptr(),
		Amount:    new(big.Int).SetUint64(unit.LDC),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	itx, err := NewTx(ltx)
	require.NoError(t, err)

	cs.CommitAccounts()
	assert.ErrorContains(itx.Apply(ctx, cs),
		"insufficient NativeLDC balance, expected 847000, got 0")
	cs.CheckoutAccounts()
	testToken.Add(ids.NativeToken, new(big.Int).SetUint64(unit.LDC))
	assert.NoError(itx.Apply(ctx, cs))

	recipientAcc := cs.MustAccount(recipient)

	tokenGas := ltx.Gas()
	assert.Equal(tokenGas*ctx.Price,
		itx.(*TxTransfer).ldc.Balance().Uint64())
	assert.Equal(tokenGas*100,
		itx.(*TxTransfer).miner.Balance().Uint64())
	assert.Equal(unit.LDC-tokenGas*(ctx.Price+100),
		testToken.Balance().Uint64())
	assert.Equal(unit.LDC*9,
		testToken.BalanceOf(token).Uint64())
	assert.Equal(uint64(0),
		recipientAcc.Balance().Uint64())
	assert.Equal(unit.LDC,
		recipientAcc.BalanceOf(token).Uint64())
	assert.Equal(uint64(1), testToken.Nonce())

	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeDestroyToken,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     1,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      testToken.ID(),
		To:        recipient.Ptr(),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	itx, err = NewTx(ltx)
	require.NoError(t, err)

	cs.CommitAccounts()
	assert.ErrorContains(itx.Apply(ctx, cs), "invalid signature for keepers")
	cs.CheckoutAccounts()

	assert.NoError(ltx.SignWith(signer.Signer1, signer.Signer2))
	assert.NoError(ltx.SyntacticVerify())
	itx, err = NewTx(ltx)
	require.NoError(t, err)

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
		To:        testToken.ID().Ptr(),
		Token:     token.Ptr(),
		Amount:    new(big.Int).SetUint64(unit.LDC),
	}}
	assert.NoError(ltx.SignWith(signer.Signer2))
	assert.NoError(ltx.SyntacticVerify())
	itx, err = NewTx(ltx)
	require.NoError(t, err)
	recipientAcc.Add(ids.NativeToken, new(big.Int).SetUint64(unit.LDC))
	assert.NoError(itx.Apply(ctx, cs))

	recipientGas := ltx.Gas()
	assert.Equal((tokenGas+recipientGas)*ctx.Price,
		itx.(*TxTransfer).ldc.Balance().Uint64())
	assert.Equal((tokenGas+recipientGas)*100,
		itx.(*TxTransfer).miner.Balance().Uint64())
	assert.Equal(ctx.FeeConfig().MinTokenPledge.Uint64()+unit.LDC-tokenGas*(ctx.Price+100),
		testToken.BalanceOfAll(ids.NativeToken).Uint64())
	assert.Equal(unit.LDC*10, testToken.BalanceOf(token).Uint64())
	assert.Equal(unit.LDC-recipientGas*(ctx.Price+100),
		recipientAcc.BalanceOfAll(ids.NativeToken).Uint64())
	assert.Equal(uint64(0), recipientAcc.BalanceOf(token).Uint64())

	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeDestroyToken,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     1,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      testToken.ID(),
		To:        recipient.Ptr(),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1, signer.Signer2))
	assert.NoError(ltx.SyntacticVerify())
	itx, err = NewTx(ltx)
	require.NoError(t, err)
	assert.NoError(itx.Apply(ctx, cs))

	tokenGas += ltx.Gas()
	assert.Equal((tokenGas+recipientGas)*ctx.Price,
		itx.(*TxDestroyToken).ldc.Balance().Uint64())
	assert.Equal((tokenGas+recipientGas)*100,
		itx.(*TxDestroyToken).miner.Balance().Uint64())

	assert.Equal(uint64(0), testToken.Balance().Uint64())
	assert.Equal(uint64(0), testToken.BalanceOfAll(ids.NativeToken).Uint64())
	assert.Equal(uint64(0), testToken.BalanceOf(token).Uint64())
	assert.Equal(uint64(0), testToken.BalanceOfAll(token).Uint64())
	assert.Equal(ctx.FeeConfig().MinTokenPledge.Uint64()+unit.LDC*2-(tokenGas+recipientGas)*(ctx.Price+100),
		recipientAcc.BalanceOfAll(ids.NativeToken).Uint64())
	assert.Equal(uint64(0), recipientAcc.BalanceOf(token).Uint64())
	assert.Equal(uint64(0), recipientAcc.BalanceOfAll(token).Uint64())

	assert.Equal(uint64(2), testToken.Nonce())
	assert.Equal(uint16(0), testToken.Threshold())
	assert.Equal(signer.Keys{}, testToken.Keepers())

	jsondata, err := itx.MarshalJSON()
	require.NoError(t, err)
	// fmt.Println(string(jsondata))
	assert.Equal(`{"tx":{"type":"TypeDestroyToken","chainID":2357,"nonce":1,"gasTip":100,"gasFeeCap":1000,"from":"0x00000000000000000000000000000000244C4443","to":"0x44171C37Ff5D7B7bb8Dcad5C81f16284A229E641"},"sigs":["lHgSlVzFykfOwC9oQ6zQy_bBYA2M956P89SZ1Kbof79p22zGK2W41CvX3pCrvxnTQlMPH-qtkSPfpnJ857EC4wG6gsN8","09Sl-st2HTiztAZ3wApmvN_sdRFBFXrF419p_FAkxwRg0IWHLYzsiiJwhIkBJtl4XmLbg0ZWol4bPqcH3ev7ZAFMSGV0"],"id":"j_xIfHsXX7FqTRExLzyENfU4hqK4YpKl1EGHvdx_7_d-tz7e"}`, string(jsondata))

	assert.NoError(cs.VerifyState())
}

func TestTxDestroyTokenWithApproverAndLending(t *testing.T) {
	assert := assert.New(t)

	ctx := NewMockChainContext()
	cs := ctx.MockChainState()
	token := ld.MustNewToken("$LDC")
	tokenid := ids.Address(token)
	sender := signer.Signer1.Key().Address()

	// CreateToken
	input := &ld.TxAccounter{
		Threshold:   ld.Uint16Ptr(1),
		Keepers:     &signer.Keys{signer.Signer1.Key()},
		Approver:    signer.Signer2.Key().Ptr(),
		ApproveList: &ld.TxTypes{ld.TypeOpenLending, ld.TypeDestroyToken},
		Amount:      new(big.Int).SetUint64(unit.LDC * 10),
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
	require.NoError(t, err)

	tokenAcc := cs.MustAccount(tokenid)
	senderAcc := cs.MustAccount(sender)
	senderAcc.Add(ids.NativeToken, ctx.FeeConfig().MinTokenPledge)
	senderAcc.Add(ids.NativeToken, new(big.Int).SetUint64(unit.LDC))
	cs.CommitAccounts()
	assert.ErrorContains(itx.Apply(ctx, cs),
		"insufficient transferable NativeLDC balance, expected 10000000000000, got 9999997696600")
	cs.CheckoutAccounts()

	senderAcc.Add(ids.NativeToken, new(big.Int).SetUint64(unit.LDC))
	assert.NoError(itx.Apply(ctx, cs))

	senderGas := ltx.Gas()
	assert.Equal(senderGas*ctx.Price,
		itx.(*TxCreateToken).ldc.Balance().Uint64())
	assert.Equal(senderGas*100,
		itx.(*TxCreateToken).miner.Balance().Uint64())
	assert.Equal(uint64(0),
		tokenAcc.Balance().Uint64())
	assert.Equal(ctx.FeeConfig().MinTokenPledge.Uint64(),
		tokenAcc.BalanceOfAll(ids.NativeToken).Uint64())
	assert.Equal(unit.LDC*10,
		tokenAcc.BalanceOf(token).Uint64())
	assert.Equal(unit.LDC*2-senderGas*(ctx.Price+100),
		senderAcc.BalanceOfAll(ids.NativeToken).Uint64())

	assert.Equal(uint16(1), tokenAcc.Threshold())
	assert.Equal(signer.Keys{signer.Signer1.Key()}, tokenAcc.Keepers())
	assert.Equal(signer.Signer2.Key(), tokenAcc.LD().Approver)
	assert.Equal(ld.TxTypes{ld.TypeOpenLending, ld.TypeDestroyToken}, tokenAcc.LD().ApproveList)

	// OpenLending
	lcfg := &ld.LendingConfig{
		Token:           token,
		DailyInterest:   10,
		OverdueInterest: 10,
		MinAmount:       big.NewInt(1000),
		MaxAmount:       new(big.Int).SetUint64(unit.LDC),
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
	require.NoError(t, err)
	senderAcc.Add(ids.NativeToken, new(big.Int).SetUint64(unit.LDC))

	cs.CommitAccounts()
	assert.ErrorContains(itx.Apply(ctx, cs),
		"TxOpenLending.Apply: invalid signature for approver")
	cs.CheckoutAccounts()

	assert.NoError(ltx.SignWith(signer.Signer1, signer.Signer2))
	assert.NoError(ltx.SyntacticVerify())
	itx, err = NewTx(ltx)
	require.NoError(t, err)

	cs.CommitAccounts()
	assert.ErrorContains(itx.Apply(ctx, cs),
		"insufficient NativeLDC balance, expected 2197800, got 0")
	cs.CheckoutAccounts()

	tokenAcc.Add(ids.NativeToken, new(big.Int).SetUint64(unit.LDC))
	assert.NoError(itx.Apply(ctx, cs))

	tokenGas := ltx.Gas()
	assert.Equal((tokenGas+senderGas)*ctx.Price,
		itx.(*TxOpenLending).ldc.Balance().Uint64())
	assert.Equal((tokenGas+senderGas)*100,
		itx.(*TxOpenLending).miner.Balance().Uint64())
	require.NotNil(t, tokenAcc.LD().Lending)
	assert.Equal(uint64(1), tokenAcc.Nonce())

	// UpdateNonceTable
	ns := []uint64{cs.Timestamp() + 1, 1, 2, 3}
	ndData, err := encoding.MarshalCBOR(ns)
	require.NoError(t, err)
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeUpdateNonceTable,
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
	require.NoError(t, err)
	assert.NoError(itx.Apply(ctx, cs))

	tokenGas += ltx.Gas()
	assert.Equal((tokenGas+senderGas)*ctx.Price,
		itx.(*TxUpdateNonceTable).ldc.Balance().Uint64())
	assert.Equal((tokenGas+senderGas)*100,
		itx.(*TxUpdateNonceTable).miner.Balance().Uint64())
	assert.Equal([]uint64{1, 2, 3}, tokenAcc.LD().NonceTable[cs.Timestamp()+1])
	assert.Equal(uint64(2), tokenAcc.Nonce())

	// Borrow
	tf := &ld.TxTransfer{
		Nonce:  3,
		From:   tokenAcc.ID().Ptr(),
		To:     sender.Ptr(),
		Token:  token.Ptr(),
		Amount: new(big.Int).SetUint64(unit.LDC),
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
		Token:     token.Ptr(),
		Data:      tf.Bytes(),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.ExSignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	ltx.Timestamp = cs.Timestamp()
	itx, err = NewTx(ltx)
	require.NoError(t, err)
	assert.NoError(itx.Apply(ctx, cs))

	senderGas += ltx.Gas()
	assert.Equal((tokenGas+senderGas)*ctx.Price,
		itx.(*TxBorrow).ldc.Balance().Uint64())
	assert.Equal((tokenGas+senderGas)*100,
		itx.(*TxBorrow).miner.Balance().Uint64())
	assert.Equal([]uint64{1, 2}, tokenAcc.LD().NonceTable[cs.Timestamp()+1])
	assert.Equal(unit.LDC*9, tokenAcc.BalanceOf(token).Uint64())
	assert.Equal(unit.LDC, senderAcc.BalanceOf(token).Uint64())

	// DestroyToken
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeDestroyToken,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     2,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      tokenid,
		To:        sender.Ptr(),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	itx, err = NewTx(ltx)
	require.NoError(t, err)

	cs.CommitAccounts()
	assert.ErrorContains(itx.Apply(ctx, cs),
		"TxDestroyToken.Apply: invalid signature for approver")
	cs.CheckoutAccounts()

	assert.NoError(ltx.SignWith(signer.Signer1, signer.Signer2))
	assert.NoError(ltx.SyntacticVerify())
	itx, err = NewTx(ltx)
	require.NoError(t, err)
	cs.CommitAccounts()
	assert.ErrorContains(itx.Apply(ctx, cs),
		"some token in the use, maxTotalSupply expected 10000000000, got 9000000000")
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
		Token:     token.Ptr(),
		Amount:    new(big.Int).SetUint64(unit.LDC),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	itx, err = NewTx(ltx)
	require.NoError(t, err)
	assert.NoError(itx.Apply(ctx, cs))

	senderGas += ltx.Gas()
	assert.Equal((tokenGas+senderGas)*ctx.Price,
		itx.(*TxRepay).ldc.Balance().Uint64())
	assert.Equal((tokenGas+senderGas)*100,
		itx.(*TxRepay).miner.Balance().Uint64())
	assert.Equal(unit.LDC*10, tokenAcc.BalanceOf(token).Uint64())
	assert.Equal(uint64(0), senderAcc.BalanceOf(token).Uint64())

	// DestroyToken
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeDestroyToken,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     2,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      tokenid,
		To:        sender.Ptr(),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1, signer.Signer2))
	assert.NoError(ltx.SyntacticVerify())
	itx, err = NewTx(ltx)
	require.NoError(t, err)
	assert.NoError(itx.Apply(ctx, cs))

	tokenGas += ltx.Gas()
	assert.Equal((tokenGas+senderGas)*ctx.Price,
		itx.(*TxDestroyToken).ldc.Balance().Uint64())
	assert.Equal((tokenGas+senderGas)*100,
		itx.(*TxDestroyToken).miner.Balance().Uint64())
	assert.Equal(uint64(3), tokenAcc.Nonce())
	assert.Equal(uint16(0), tokenAcc.Threshold())
	assert.Equal(signer.Keys{}, tokenAcc.Keepers())
	assert.Equal(make(map[uint64][]uint64), tokenAcc.LD().NonceTable)
	assert.Equal(ld.NativeAccount, tokenAcc.LD().Type)
	assert.Nil(tokenAcc.LD().Approver)
	assert.Nil(tokenAcc.LD().ApproveList)
	assert.Nil(tokenAcc.LD().Lending)
	assert.Nil(tokenAcc.LD().MaxTotalSupply)
	assert.Nil(tokenAcc.LD().Tokens[token.AsKey()])
	assert.Equal(uint64(0), tokenAcc.BalanceOfAll(token).Uint64())
	assert.Equal(uint64(0), tokenAcc.BalanceOfAll(ids.NativeToken).Uint64())

	jsondata, err := itx.MarshalJSON()
	require.NoError(t, err)
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
		To:        sender.Ptr(),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1, signer.Signer2))
	assert.NoError(ltx.SyntacticVerify())
	itx, err = NewTx(ltx)
	require.NoError(t, err)
	assert.ErrorContains(itx.Apply(ctx, cs),
		"TxDestroyToken.Apply: invalid signatures for sender")

	assert.NoError(cs.VerifyState())
}
