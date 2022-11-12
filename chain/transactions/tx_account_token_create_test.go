// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package transactions

import (
	"math/big"
	"strconv"
	"testing"

	"github.com/ldclabs/ldvm/constants"
	"github.com/ldclabs/ldvm/ld"
	"github.com/ldclabs/ldvm/util"
	"github.com/ldclabs/ldvm/util/signer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTxCreateToken(t *testing.T) {
	assert := assert.New(t)

	// SyntacticVerify
	tx := &TxCreateToken{}
	assert.ErrorContains(tx.SyntacticVerify(), "nil pointer")
	_, err := tx.MarshalJSON()
	require.NoError(t, err)

	ctx := NewMockChainContext()
	cs := ctx.MockChainState()
	token := ld.MustNewToken("$LDC")
	tokenid := util.Address(token)
	sender := signer.Signer1.Key().Address()

	ltx := &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeCreateToken,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      sender,
	}}
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "no signatures")

	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeCreateToken,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      sender,
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "nil to as token account")

	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeCreateToken,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      sender,
		To:        &tokenid,
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "nil amount")

	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeCreateToken,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      sender,
		To:        &tokenid,
		Amount:    new(big.Int).SetUint64(100),
		Token:     token.Ptr(),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "invalid token, should be nil")

	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeCreateToken,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      sender,
		To:        &tokenid,
		Amount:    new(big.Int).SetUint64(100),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "invalid data")

	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeCreateToken,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      sender,
		To:        &tokenid,
		Amount:    new(big.Int).SetUint64(100),
		Data:      []byte("你好👋"),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "cbor: unexpected following extraneous data")

	input := &ld.TxAccounter{}
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeCreateToken,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      sender,
		To:        signer.Signer2.Key().Address().Ptr(),
		Amount:    new(big.Int).SetUint64(100),
		Data:      input.Bytes(),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "invalid token 0x44171C37Ff5D7B7bb8Dcad5C81f16284A229E641")

	input = &ld.TxAccounter{}
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeCreateToken,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      sender,
		To:        &tokenid,
		Amount:    new(big.Int).SetUint64(100),
		Data:      input.Bytes(),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "invalid threshold, expected >= 1")

	input = &ld.TxAccounter{
		Threshold: ld.Uint16Ptr(0),
		Keepers:   &signer.Keys{signer.Signer1.Key()},
	}
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeCreateToken,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      sender,
		To:        &tokenid,
		Amount:    new(big.Int).SetUint64(100),
		Data:      input.Bytes(),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "invalid threshold, expected >= 1")

	input = &ld.TxAccounter{
		Threshold: ld.Uint16Ptr(1),
		Keepers:   &signer.Keys{signer.Signer1.Key()},
	}
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeCreateToken,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      sender,
		To:        &tokenid,
		Amount:    new(big.Int).SetUint64(100),
		Data:      input.Bytes(),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "invalid amount, expected >= 1")

	input = &ld.TxAccounter{
		Threshold: ld.Uint16Ptr(1),
		Keepers:   &signer.Keys{signer.Signer1.Key()},
		Amount:    new(big.Int).SetUint64(constants.LDC * 10),
		Name:      "LDC\nToken",
	}
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeCreateToken,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      sender,
		To:        &tokenid,
		Amount:    new(big.Int).SetUint64(100),
		Data:      input.Bytes(),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, `invalid name "LDC\nToken"`)

	input = &ld.TxAccounter{
		Threshold: ld.Uint16Ptr(1),
		Keepers:   &signer.Keys{signer.Signer1.Key()},
		Amount:    new(big.Int).SetUint64(constants.LDC * 10),
		Approver:  signer.Signer2.Key().Ptr(),
		Name:      "LD",
	}
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeCreateToken,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      sender,
		To:        &tokenid,
		Amount:    new(big.Int).SetUint64(100),
		Data:      input.Bytes(),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, `invalid name "LD", expected length >= 3`)

	input = &ld.TxAccounter{
		Threshold: ld.Uint16Ptr(1),
		Keepers:   &signer.Keys{signer.Signer1.Key()},
		Amount:    new(big.Int).SetUint64(constants.LDC * 10),
		Approver:  &signer.Key{},
		Name:      "LDC",
	}
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeCreateToken,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      sender,
		To:        &tokenid,
		Amount:    new(big.Int).SetUint64(100),
		Data:      input.Bytes(),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err,
		"invalid approver, signer.Key.Valid: empty key")

	input = &ld.TxAccounter{
		Threshold: ld.Uint16Ptr(1),
		Keepers:   &signer.Keys{signer.Signer1.Key()},
		Amount:    new(big.Int).SetUint64(constants.LDC * 10),
		Approver:  signer.Signer2.Key().Ptr(),
		Name:      "LDC",
	}
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeCreateToken,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      sender,
		To:        &tokenid,
		Amount:    new(big.Int).SetUint64(100),
		Data:      input.Bytes(),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	itx, err := NewTx(ltx)
	require.NoError(t, err)

	cs.CommitAccounts()
	assert.ErrorContains(itx.Apply(ctx, cs),
		`insufficient NativeLDC balance, expected 2179200, got 0`)
	cs.CheckoutAccounts()

	senderAcc := cs.MustAccount(sender)
	senderAcc.Add(constants.NativeToken, new(big.Int).SetUint64(constants.LDC))
	cs.CommitAccounts()
	assert.ErrorContains(itx.Apply(ctx, cs),
		`invalid amount, expected >= 10000000000000, got 100`)
	cs.CheckoutAccounts()

	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeCreateToken,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      sender,
		To:        &tokenid,
		Amount:    new(big.Int).SetUint64(10000000000000),
		Data:      input.Bytes(),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	itx, err = NewTx(ltx)
	require.NoError(t, err)
	senderAcc.Add(constants.NativeToken, new(big.Int).SetUint64(10000000000000))
	cs.CommitAccounts()
	assert.ErrorContains(itx.Apply(ctx, cs),
		`insufficient transferable NativeLDC balance, expected 10000000000000, got 9999997790100`)
	cs.CheckoutAccounts()

	senderAcc.Add(constants.NativeToken, new(big.Int).SetUint64(constants.LDC))
	assert.NoError(itx.Apply(ctx, cs))

	tokenAcc := cs.MustAccount(tokenid)
	senderGas := ltx.Gas()
	assert.Equal(senderGas*ctx.Price,
		itx.(*TxCreateToken).ldc.Balance().Uint64())
	assert.Equal(senderGas*100,
		itx.(*TxCreateToken).miner.Balance().Uint64())
	assert.Equal(uint64(0), tokenAcc.Balance().Uint64())
	assert.Equal(uint64(10000000000000), tokenAcc.BalanceOfAll(constants.NativeToken).Uint64())
	assert.Equal(constants.LDC*10, tokenAcc.BalanceOf(token).Uint64())
	assert.Equal(constants.LDC-senderGas*(ctx.Price+100),
		senderAcc.Balance().Uint64())
	assert.Equal(constants.LDC*2-senderGas*(ctx.Price+100),
		senderAcc.BalanceOfAll(constants.NativeToken).Uint64())

	assert.Equal(uint64(0), tokenAcc.Nonce())
	assert.Equal(uint16(1), tokenAcc.Threshold())
	assert.Equal(signer.Keys{signer.Signer1.Key()}, tokenAcc.Keepers())
	assert.Equal(signer.Signer2.Key(), tokenAcc.LD().Approver)
	assert.Equal(constants.LDC*10, tokenAcc.LD().MaxTotalSupply.Uint64())
	assert.Equal(constants.LDC*10, tokenAcc.LD().Tokens[token.AsKey()].Uint64())

	jsondata, err := itx.MarshalJSON()
	require.NoError(t, err)
	// fmt.Println(string(jsondata))
	assert.Equal(`{"tx":{"type":"TypeCreateToken","chainID":2357,"nonce":0,"gasTip":100,"gasFeeCap":1000,"from":"0x8db97c7cECe249C2b98bdc0226cc4C2A57bF52fc","to":"0x00000000000000000000000000000000244C4443","amount":10000000000000,"data":{"threshold":1,"keepers":["jbl8fOziScK5i9wCJsxMKle_UvwKxwPH"],"approver":"RBccN_9de3u43K1cgfFihKIp5kE1lmGG","amount":10000000000,"name":"LDC"}},"sigs":["8htMbeZH3FXJvx1-vCF_JNDw2U5VYz3PxWl_Nvd654s5Tr1tM4mmbGRKb_Nwq2fgZfiyP-J517t3P3NoCMYA3QC8upym"],"id":"sS-MItyigKpK8_aqbgoz9STjakL974USNZaCQdj1t1sZN1Xu"}`, string(jsondata))

	// create again
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeCreateToken,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     1,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      sender,
		To:        &tokenid,
		Amount:    new(big.Int).SetUint64(10000000000000),
		Data:      input.Bytes(),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	itx, err = NewTx(ltx)
	require.NoError(t, err)
	senderAcc.Add(constants.NativeToken, new(big.Int).SetUint64(10000000000000))
	cs.CommitAccounts()
	assert.ErrorContains(itx.Apply(ctx, cs), "token account $LDC exists")
	cs.CheckoutAccounts()

	// destroy and create again
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeDestroyToken,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
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
		"insufficient NativeLDC balance, expected 2113100, got 0")
	cs.CheckoutAccounts()
	tokenAcc.Add(constants.NativeToken, new(big.Int).SetUint64(constants.LDC))
	assert.NoError(itx.Apply(ctx, cs))

	senderGas += ltx.Gas()
	assert.Equal(senderGas*ctx.Price,
		itx.(*TxDestroyToken).ldc.Balance().Uint64())
	assert.Equal(senderGas*100,
		itx.(*TxDestroyToken).miner.Balance().Uint64())
	assert.Equal(uint64(1), tokenAcc.Nonce())
	assert.Equal(uint16(0), tokenAcc.Threshold())
	assert.Equal(signer.Keys{}, tokenAcc.Keepers())
	assert.Equal(ld.NativeAccount, tokenAcc.LD().Type)
	assert.Nil(tokenAcc.LD().Approver)
	assert.Nil(tokenAcc.LD().MaxTotalSupply)
	assert.Nil(tokenAcc.LD().Tokens[token.AsKey()])

	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeCreateToken,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     1,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      sender,
		To:        &tokenid,
		Amount:    new(big.Int).SetUint64(10000000000000),
		Data:      input.Bytes(),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	itx, err = NewTx(ltx)
	require.NoError(t, err)
	senderAcc.Add(constants.NativeToken, new(big.Int).SetUint64(10000000000000))
	assert.NoError(itx.Apply(ctx, cs))

	senderGas += ltx.Gas()
	assert.Equal(senderGas*ctx.Price,
		itx.(*TxCreateToken).ldc.Balance().Uint64())
	assert.Equal(senderGas*100,
		itx.(*TxCreateToken).miner.Balance().Uint64())
	assert.Equal(uint64(1), tokenAcc.Nonce())
	assert.Equal(uint16(1), tokenAcc.Threshold())
	assert.Equal(signer.Keys{signer.Signer1.Key()}, tokenAcc.Keepers())
	assert.Equal(signer.Signer2.Key(), tokenAcc.LD().Approver)
	assert.Equal(constants.LDC*10, tokenAcc.LD().MaxTotalSupply.Uint64())
	assert.Equal(constants.LDC*10, tokenAcc.LD().Tokens[token.AsKey()].Uint64())
	assert.Equal(uint64(0), tokenAcc.Balance().Uint64())
	assert.Equal(uint64(10000000000000), tokenAcc.BalanceOfAll(constants.NativeToken).Uint64())
	assert.Equal(constants.LDC*10, tokenAcc.BalanceOf(token).Uint64())

	assert.NoError(cs.VerifyState())
}

func TestTxCreateTokenGenesis(t *testing.T) {
	assert := assert.New(t)

	ctx := NewMockChainContext()
	cs := ctx.MockChainState()

	sender := signer.Signer1.Key().Address()

	// can not create the NativeToken
	input := &ld.TxAccounter{
		Threshold: ld.Uint16Ptr(1),
		Keepers:   &signer.Keys{signer.Signer1.Key()},
		Amount:    new(big.Int).SetUint64(constants.LDC * 10),
		Name:      "NativeToken",
	}
	ltx := &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeCreateToken,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      sender,
		To:        &constants.LDCAccount,
		Amount:    new(big.Int).SetUint64(10000000000000),
		Data:      input.Bytes(),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err := NewTx(ltx)
	assert.ErrorContains(err,
		"invalid to as token account, expected not 0x0000000000000000000000000000000000000000")

	// create the NativeToken in GenesisTx
	input = &ld.TxAccounter{
		Amount: ctx.ChainConfig().MaxTotalSupply,
		Name:   "Linked Data Chain",
		Data:   []byte(strconv.Quote(ctx.ChainConfig().Message)),
	}
	assert.NoError(input.SyntacticVerify())
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:    ld.TypeCreateToken,
		ChainID: ctx.ChainConfig().ChainID,
		From:    constants.GenesisAccount,
		To:      &constants.LDCAccount,
		Data:    input.Bytes(),
	}}
	itx, err := NewGenesisTx(ltx)
	require.NoError(t, err)

	assert.NoError(itx.(*TxCreateToken).ApplyGenesis(ctx, cs))

	ldcAcc := cs.MustAccount(constants.LDCAccount)
	assert.Equal(ctx.ChainConfig().MaxTotalSupply.Uint64(),
		ldcAcc.Balance().Uint64())
	assert.Equal(ctx.ChainConfig().MaxTotalSupply.Uint64(),
		ldcAcc.BalanceOfAll(constants.NativeToken).Uint64())
	assert.Equal(uint64(0), itx.(*TxCreateToken).miner.Balance().Uint64())
	assert.Equal(uint64(0), itx.(*TxCreateToken).from.Balance().Uint64())
	assert.Equal(uint64(1), itx.(*TxCreateToken).from.Nonce())

	assert.Equal(uint64(0), ldcAcc.Nonce())
	assert.Equal(uint16(0), ldcAcc.Threshold())
	assert.Equal(signer.Keys{}, ldcAcc.Keepers())
	assert.Nil(ldcAcc.LD().Approver)
	assert.Nil(ldcAcc.LD().ApproveList)
	assert.Equal(ctx.ChainConfig().MaxTotalSupply.Uint64(), ldcAcc.LD().MaxTotalSupply.Uint64())
	assert.Equal(0, len(ldcAcc.LD().Tokens))

	jsondata, err := itx.MarshalJSON()
	require.NoError(t, err)
	// fmt.Println(string(jsondata))
	assert.Equal(`{"tx":{"type":"TypeCreateToken","chainID":2357,"nonce":0,"gasTip":0,"gasFeeCap":0,"from":"0xFFfFFFfFfffFFfFFffFFFfFfFffFFFfffFfFFFff","to":"0x0000000000000000000000000000000000000000","data":{"amount":1000000000000000000,"name":"Linked Data Chain","data":"Hello, LDVM!"}},"id":"85P0P5wJoPXpZF7BSALe1ronSxmtj0PkWzx5wlWYo52WX3ri"}`, string(jsondata))

	// NativeToken cannot be destroy
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeDestroyToken,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      constants.LDCAccount,
		To:        constants.GenesisAccount.Ptr(),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "TxBase.SyntacticVerify: invalid from")

	assert.NoError(cs.VerifyState())
}
