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
	"github.com/stretchr/testify/assert"
)

func TestTxCreateToken(t *testing.T) {
	assert := assert.New(t)

	// SyntacticVerify
	tx := &TxCreateToken{}
	assert.ErrorContains(tx.SyntacticVerify(), "nil pointer")
	_, err := tx.MarshalJSON()
	assert.NoError(err)

	ctx := NewMockChainContext()
	cs := ctx.MockChainState()
	token := ld.MustNewToken("$LDC")
	tokenid := util.EthID(token)

	sender := util.Signer1.Address()
	approver := util.Signer2.Address()

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
	assert.ErrorContains(err, "DeriveSigners error: no signature")

	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeCreateToken,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      sender,
	}}
	assert.NoError(ltx.SignWith(util.Signer1))
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
	assert.NoError(ltx.SignWith(util.Signer1))
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
		Token:     &token,
	}}
	assert.NoError(ltx.SignWith(util.Signer1))
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
	assert.NoError(ltx.SignWith(util.Signer1))
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
	assert.NoError(ltx.SignWith(util.Signer1))
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
		To:        &approver,
		Amount:    new(big.Int).SetUint64(100),
		Data:      input.Bytes(),
	}}
	assert.NoError(ltx.SignWith(util.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "invalid token 0x44171C37Ff5D7B7bb8dcad5C81f16284A229e641")

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
	assert.NoError(ltx.SignWith(util.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "invalid threshold, expected >= 1")

	input = &ld.TxAccounter{
		Threshold: ld.Uint16Ptr(0),
		Keepers:   &util.EthIDs{util.Signer1.Address()},
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
	assert.NoError(ltx.SignWith(util.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "invalid threshold, expected >= 1")

	input = &ld.TxAccounter{
		Threshold: ld.Uint16Ptr(1),
		Keepers:   &util.EthIDs{util.Signer1.Address()},
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
	assert.NoError(ltx.SignWith(util.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "invalid amount, expected >= 1")

	input = &ld.TxAccounter{
		Threshold: ld.Uint16Ptr(1),
		Keepers:   &util.EthIDs{util.Signer1.Address()},
		Amount:    new(big.Int).SetUint64(constants.LDC * 10),
		Approver:  &util.EthIDEmpty,
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
	assert.NoError(ltx.SignWith(util.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err,
		"invalid approver, expected not 0x0000000000000000000000000000000000000000")

	input = &ld.TxAccounter{
		Threshold: ld.Uint16Ptr(1),
		Keepers:   &util.EthIDs{util.Signer1.Address()},
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
	assert.NoError(ltx.SignWith(util.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, `invalid name "LDC\nToken"`)

	input = &ld.TxAccounter{
		Threshold: ld.Uint16Ptr(1),
		Keepers:   &util.EthIDs{util.Signer1.Address()},
		Amount:    new(big.Int).SetUint64(constants.LDC * 10),
		Approver:  &approver,
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
	assert.NoError(ltx.SignWith(util.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, `invalid name "LD", expected length >= 3`)

	input = &ld.TxAccounter{
		Threshold: ld.Uint16Ptr(1),
		Keepers:   &util.EthIDs{util.Signer1.Address()},
		Amount:    new(big.Int).SetUint64(constants.LDC * 10),
		Approver:  &approver,
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
	assert.NoError(ltx.SignWith(util.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	itx, err := NewTx(ltx)
	assert.NoError(err)

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
	assert.NoError(ltx.SignWith(util.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	itx, err = NewTx(ltx)
	assert.NoError(err)
	senderAcc.Add(constants.NativeToken, new(big.Int).SetUint64(10000000000000))
	assert.NoError(itx.Apply(ctx, cs))

	tokenAcc := cs.MustAccount(tokenid)
	senderGas := ltx.Gas()
	assert.Equal(senderGas*ctx.Price,
		itx.(*TxCreateToken).ldc.Balance().Uint64())
	assert.Equal(senderGas*100,
		itx.(*TxCreateToken).miner.Balance().Uint64())
	assert.Equal(uint64(0), tokenAcc.Balance().Uint64())
	assert.Equal(uint64(10000000000000), tokenAcc.balanceOfAll(constants.NativeToken).Uint64())
	assert.Equal(constants.LDC*10, tokenAcc.balanceOf(token).Uint64())
	assert.Equal(constants.LDC-senderGas*(ctx.Price+100),
		senderAcc.Balance().Uint64())

	assert.Equal(uint64(0), tokenAcc.Nonce())
	assert.Equal(uint16(1), tokenAcc.Threshold())
	assert.Equal(util.EthIDs{util.Signer1.Address()}, tokenAcc.Keepers())
	assert.Equal(approver, *tokenAcc.ld.Approver)
	assert.Equal(constants.LDC*10, tokenAcc.ld.MaxTotalSupply.Uint64())
	assert.Equal(constants.LDC*10, tokenAcc.ld.Tokens[token.AsKey()].Uint64())

	jsondata, err := itx.MarshalJSON()
	assert.NoError(err)
	// fmt.Println(string(jsondata))
	assert.Equal(`{"tx":{"type":"TypeCreateToken","chainID":2357,"nonce":0,"gasTip":100,"gasFeeCap":1000,"from":"0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC","to":"0x00000000000000000000000000000000244C4443","amount":10000000000000,"data":{"threshold":1,"keepers":["0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC"],"approver":"0x44171C37Ff5D7B7bb8dcad5C81f16284A229e641","amount":10000000000,"name":"LDC"}},"sigs":["f21b4c6de647dc55c9bf1d7ebc217f24d0f0d94e55633dcfc5697f36f77ae78b394ebd6d3389a66c644a6ff370ab67e065f8b23fe279d7bb773f736808c600dd00"],"id":"2M2yMrerj2vm7kaRitpY4VXTkvAg52ocTQwRpcfceWw3frJaNY"}`, string(jsondata))

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
	assert.NoError(ltx.SignWith(util.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	itx, err = NewTx(ltx)
	assert.NoError(err)
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
		To:        &sender,
	}}
	assert.NoError(ltx.SignWith(util.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	itx, err = NewTx(ltx)
	assert.NoError(err)
	cs.CommitAccounts()
	assert.ErrorContains(itx.Apply(ctx, cs),
		"TxDestroyToken.Apply error: invalid signature for approver")
	cs.CheckoutAccounts()

	assert.NoError(ltx.SignWith(util.Signer1, util.Signer2))
	assert.NoError(ltx.SyntacticVerify())
	itx, err = NewTx(ltx)
	assert.NoError(err)
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
	assert.Equal(util.EthIDs{}, tokenAcc.Keepers())
	assert.Equal(ld.NativeAccount, tokenAcc.ld.Type)
	assert.Nil(tokenAcc.ld.Approver)
	assert.Nil(tokenAcc.ld.MaxTotalSupply)
	assert.Nil(tokenAcc.ld.Tokens[token.AsKey()])

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
	assert.NoError(ltx.SignWith(util.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	itx, err = NewTx(ltx)
	assert.NoError(err)
	senderAcc.Add(constants.NativeToken, new(big.Int).SetUint64(10000000000000))
	assert.NoError(itx.Apply(ctx, cs))

	senderGas += ltx.Gas()
	assert.Equal(senderGas*ctx.Price,
		itx.(*TxCreateToken).ldc.Balance().Uint64())
	assert.Equal(senderGas*100,
		itx.(*TxCreateToken).miner.Balance().Uint64())
	assert.Equal(uint64(1), tokenAcc.Nonce())
	assert.Equal(uint16(1), tokenAcc.Threshold())
	assert.Equal(util.EthIDs{util.Signer1.Address()}, tokenAcc.Keepers())
	assert.Equal(approver, *tokenAcc.ld.Approver)
	assert.Equal(constants.LDC*10, tokenAcc.ld.MaxTotalSupply.Uint64())
	assert.Equal(constants.LDC*10, tokenAcc.ld.Tokens[token.AsKey()].Uint64())
	assert.Equal(uint64(0), tokenAcc.Balance().Uint64())
	assert.Equal(uint64(10000000000000), tokenAcc.balanceOfAll(constants.NativeToken).Uint64())
	assert.Equal(constants.LDC*10, tokenAcc.balanceOf(token).Uint64())

	assert.NoError(cs.VerifyState())
}

func TestTxCreateTokenGenesis(t *testing.T) {
	assert := assert.New(t)

	ctx := NewMockChainContext()
	cs := ctx.MockChainState()

	sender := util.Signer1.Address()

	// can not create the NativeToken
	input := &ld.TxAccounter{
		Threshold: ld.Uint16Ptr(1),
		Keepers:   &util.EthIDs{util.Signer1.Address()},
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
	assert.NoError(ltx.SignWith(util.Signer1))
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
	assert.NoError(err)

	assert.NoError(itx.(*TxCreateToken).ApplyGenesis(ctx, cs))

	ldcAcc := cs.MustAccount(constants.LDCAccount)
	assert.Equal(ctx.ChainConfig().MaxTotalSupply.Uint64(),
		ldcAcc.Balance().Uint64())
	assert.Equal(ctx.ChainConfig().MaxTotalSupply.Uint64(),
		ldcAcc.balanceOfAll(constants.NativeToken).Uint64())
	assert.Equal(uint64(0), itx.(*TxCreateToken).miner.Balance().Uint64())
	assert.Equal(uint64(0), itx.(*TxCreateToken).from.Balance().Uint64())
	assert.Equal(uint64(1), itx.(*TxCreateToken).from.Nonce())

	assert.Equal(uint64(0), ldcAcc.Nonce())
	assert.Equal(uint16(0), ldcAcc.Threshold())
	assert.Equal(util.EthIDs{}, ldcAcc.Keepers())
	assert.Nil(ldcAcc.ld.Approver)
	assert.Nil(ldcAcc.ld.ApproveList)
	assert.Equal(ctx.ChainConfig().MaxTotalSupply.Uint64(), ldcAcc.ld.MaxTotalSupply.Uint64())
	assert.Equal(0, len(ldcAcc.ld.Tokens))

	jsondata, err := itx.MarshalJSON()
	assert.NoError(err)
	// fmt.Println(string(jsondata))
	assert.Equal(`{"tx":{"type":"TypeCreateToken","chainID":2357,"nonce":0,"gasTip":0,"gasFeeCap":0,"from":"0xFFfFfFffFFfffFFfFFfFFFFFffFFFffffFfFFFfF","to":"0x0000000000000000000000000000000000000000","data":{"amount":1000000000000000000,"name":"Linked Data Chain","data":"Hello, LDVM!"}},"id":"2rGsZRNdiUHWVmPpsSEXzYA4wnr3YUdsTRLTMw2jSswtPdLky3"}`, string(jsondata))

	// NativeToken cannot be destroy
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeDestroyToken,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      constants.LDCAccount,
		To:        &constants.GenesisAccount,
	}}
	assert.NoError(ltx.SignWith(util.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "TxBase.SyntacticVerify error: invalid from")

	assert.NoError(cs.VerifyState())
}
