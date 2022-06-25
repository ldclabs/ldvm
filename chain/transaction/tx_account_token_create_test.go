// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package transaction

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

	bctx := NewMockBCtx()
	bs := bctx.MockBS()
	token := ld.MustNewToken("$LDC")
	tokenid := util.EthID(token)

	sender := util.Signer1.Address()
	approver := util.Signer2.Address()

	txData := &ld.TxData{
		Type:      ld.TypeCreateToken,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      sender,
	}
	assert.NoError(txData.SyntacticVerify())
	_, err = NewTx(txData.ToTransaction(), true)
	assert.ErrorContains(err, "DeriveSigners error: no signature")

	txData = &ld.TxData{
		Type:      ld.TypeCreateToken,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      sender,
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx(txData.ToTransaction(), true)
	assert.ErrorContains(err, "nil to as token account")

	txData = &ld.TxData{
		Type:      ld.TypeCreateToken,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      sender,
		To:        &tokenid,
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx(txData.ToTransaction(), true)
	assert.ErrorContains(err, "nil amount")

	txData = &ld.TxData{
		Type:      ld.TypeCreateToken,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      sender,
		To:        &tokenid,
		Amount:    new(big.Int).SetUint64(100),
		Token:     &token,
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx(txData.ToTransaction(), true)
	assert.ErrorContains(err, "invalid token, should be nil")

	txData = &ld.TxData{
		Type:      ld.TypeCreateToken,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      sender,
		To:        &tokenid,
		Amount:    new(big.Int).SetUint64(100),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx(txData.ToTransaction(), true)
	assert.ErrorContains(err, "invalid data")

	txData = &ld.TxData{
		Type:      ld.TypeCreateToken,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      sender,
		To:        &tokenid,
		Amount:    new(big.Int).SetUint64(100),
		Data:      []byte("你好👋"),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx(txData.ToTransaction(), true)
	assert.ErrorContains(err, "cbor: cannot unmarshal")

	input := &ld.TxAccounter{}
	txData = &ld.TxData{
		Type:      ld.TypeCreateToken,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      sender,
		To:        &approver,
		Amount:    new(big.Int).SetUint64(100),
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx(txData.ToTransaction(), true)
	assert.ErrorContains(err, "invalid token 0x44171C37Ff5D7B7bb8dcad5C81f16284A229e641")

	input = &ld.TxAccounter{}
	txData = &ld.TxData{
		Type:      ld.TypeCreateToken,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      sender,
		To:        &tokenid,
		Amount:    new(big.Int).SetUint64(100),
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx(txData.ToTransaction(), true)
	assert.ErrorContains(err, "invalid threshold, expected >= 1")

	input = &ld.TxAccounter{
		Threshold: ld.Uint16Ptr(0),
		Keepers:   &util.EthIDs{util.Signer1.Address()},
	}
	txData = &ld.TxData{
		Type:      ld.TypeCreateToken,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      sender,
		To:        &tokenid,
		Amount:    new(big.Int).SetUint64(100),
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx(txData.ToTransaction(), true)
	assert.ErrorContains(err, "invalid threshold, expected >= 1")

	input = &ld.TxAccounter{
		Threshold: ld.Uint16Ptr(1),
		Keepers:   &util.EthIDs{util.Signer1.Address()},
	}
	txData = &ld.TxData{
		Type:      ld.TypeCreateToken,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      sender,
		To:        &tokenid,
		Amount:    new(big.Int).SetUint64(100),
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx(txData.ToTransaction(), true)
	assert.ErrorContains(err, "invalid amount, expected >= 1")

	input = &ld.TxAccounter{
		Threshold: ld.Uint16Ptr(1),
		Keepers:   &util.EthIDs{util.Signer1.Address()},
		Amount:    new(big.Int).SetUint64(constants.LDC * 10),
		Approver:  &util.EthIDEmpty,
	}
	txData = &ld.TxData{
		Type:      ld.TypeCreateToken,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      sender,
		To:        &tokenid,
		Amount:    new(big.Int).SetUint64(100),
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx(txData.ToTransaction(), true)
	assert.ErrorContains(err,
		"invalid approver, expected not 0x0000000000000000000000000000000000000000")

	input = &ld.TxAccounter{
		Threshold: ld.Uint16Ptr(1),
		Keepers:   &util.EthIDs{util.Signer1.Address()},
		Amount:    new(big.Int).SetUint64(constants.LDC * 10),
		Name:      "LDC\nToken",
	}
	txData = &ld.TxData{
		Type:      ld.TypeCreateToken,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      sender,
		To:        &tokenid,
		Amount:    new(big.Int).SetUint64(100),
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx(txData.ToTransaction(), true)
	assert.ErrorContains(err, `invalid name "LDC\nToken"`)

	input = &ld.TxAccounter{
		Threshold: ld.Uint16Ptr(1),
		Keepers:   &util.EthIDs{util.Signer1.Address()},
		Amount:    new(big.Int).SetUint64(constants.LDC * 10),
		Approver:  &approver,
		Name:      "LD",
	}
	txData = &ld.TxData{
		Type:      ld.TypeCreateToken,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      sender,
		To:        &tokenid,
		Amount:    new(big.Int).SetUint64(100),
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx(txData.ToTransaction(), true)
	assert.ErrorContains(err, `invalid name "LD", expected length >= 3`)

	input = &ld.TxAccounter{
		Threshold: ld.Uint16Ptr(1),
		Keepers:   &util.EthIDs{util.Signer1.Address()},
		Amount:    new(big.Int).SetUint64(constants.LDC * 10),
		Approver:  &approver,
		Name:      "LDC",
	}
	txData = &ld.TxData{
		Type:      ld.TypeCreateToken,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      sender,
		To:        &tokenid,
		Amount:    new(big.Int).SetUint64(100),
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	tt := txData.ToTransaction()
	itx, err := NewTx(tt, true)
	assert.NoError(err)
	assert.ErrorContains(itx.Apply(bctx, bs), `invalid gas, expected 1586, got 0`)

	tt.Gas = tt.RequiredGas(bctx.FeeConfig().ThresholdGas)
	itx, err = NewTx(tt, true)
	assert.NoError(err)
	bs.CommitAccounts()
	assert.ErrorContains(itx.Apply(bctx, bs),
		`insufficient NativeLDC balance, expected 1744700, got 0`)
	bs.CheckoutAccounts()

	senderAcc := bs.MustAccount(sender)
	senderAcc.Add(constants.NativeToken, new(big.Int).SetUint64(constants.LDC))
	bs.CommitAccounts()
	assert.ErrorContains(itx.Apply(bctx, bs),
		`invalid amount, expected >= 10000000000000, got 100`)
	bs.CheckoutAccounts()

	txData = &ld.TxData{
		Type:      ld.TypeCreateToken,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      sender,
		To:        &tokenid,
		Amount:    new(big.Int).SetUint64(10000000000000),
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	tt = txData.ToTransaction()
	tt.Gas = tt.RequiredGas(bctx.FeeConfig().ThresholdGas)
	senderGas := tt.Gas
	itx, err = NewTx(tt, true)
	assert.NoError(err)
	senderAcc.Add(constants.NativeToken, new(big.Int).SetUint64(10000000000000))
	assert.NoError(itx.Apply(bctx, bs))

	tokenAcc := bs.MustAccount(tokenid)

	assert.Equal(senderGas*bctx.Price,
		itx.(*TxCreateToken).ldc.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(senderGas*100,
		itx.(*TxCreateToken).miner.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(uint64(0), tokenAcc.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(uint64(10000000000000), tokenAcc.balanceOfAll(constants.NativeToken).Uint64())
	assert.Equal(constants.LDC*10, tokenAcc.balanceOf(token).Uint64())
	assert.Equal(constants.LDC-senderGas*(bctx.Price+100),
		senderAcc.balanceOf(constants.NativeToken).Uint64())

	assert.Equal(uint64(0), tokenAcc.Nonce())
	assert.Equal(uint16(1), tokenAcc.Threshold())
	assert.Equal(util.EthIDs{util.Signer1.Address()}, tokenAcc.Keepers())
	assert.Equal(approver, *tokenAcc.ld.Approver)
	assert.Equal(constants.LDC*10, tokenAcc.ld.MaxTotalSupply.Uint64())
	assert.Equal(constants.LDC*10, tokenAcc.ld.Tokens[token].Uint64())

	jsondata, err := itx.MarshalJSON()
	assert.NoError(err)
	// fmt.Println(string(jsondata))
	assert.Equal(`{"type":"TypeCreateToken","chainID":2357,"nonce":0,"gasTip":100,"gasFeeCap":1000,"from":"0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC","to":"0x00000000000000000000000000000000244C4443","amount":10000000000000,"data":{"threshold":1,"keepers":["0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC"],"approver":"0x44171C37Ff5D7B7bb8dcad5C81f16284A229e641","amount":10000000000,"name":"LDC"},"signatures":["f21b4c6de647dc55c9bf1d7ebc217f24d0f0d94e55633dcfc5697f36f77ae78b394ebd6d3389a66c644a6ff370ab67e065f8b23fe279d7bb773f736808c600dd00"],"gas":1611,"id":"25sqrKnJcWpahrL6M6YUygiXXWvJ2iUVsYgzLHyqMCWmDsLGLo"}`, string(jsondata))

	// create again
	txData = &ld.TxData{
		Type:      ld.TypeCreateToken,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     1,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      sender,
		To:        &tokenid,
		Amount:    new(big.Int).SetUint64(10000000000000),
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	tt = txData.ToTransaction()
	tt.Gas = tt.RequiredGas(bctx.FeeConfig().ThresholdGas)
	itx, err = NewTx(tt, true)
	assert.NoError(err)
	senderAcc.Add(constants.NativeToken, new(big.Int).SetUint64(10000000000000))
	bs.CommitAccounts()
	assert.ErrorContains(itx.Apply(bctx, bs), "token account $LDC exists")
	bs.CheckoutAccounts()

	// destroy and create again
	txData = &ld.TxData{
		Type:      ld.TypeDestroyToken,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      tokenid,
		To:        &sender,
	}
	assert.NoError(txData.SignWith(util.Signer1))
	tt = txData.ToTransaction()
	tt.Gas = tt.RequiredGas(bctx.FeeConfig().ThresholdGas)
	itx, err = NewTx(tt, true)
	assert.NoError(err)
	bs.CommitAccounts()
	assert.ErrorContains(itx.Apply(bctx, bs),
		"TxDestroyToken.Apply error: invalid signature for approver")
	bs.CheckoutAccounts()

	assert.NoError(txData.SignWith(util.Signer2))
	tt = txData.ToTransaction()
	tt.Gas = tt.RequiredGas(bctx.FeeConfig().ThresholdGas)
	senderGas += tt.Gas
	itx, err = NewTx(tt, true)
	assert.NoError(err)
	bs.CommitAccounts()
	assert.ErrorContains(itx.Apply(bctx, bs),
		"insufficient NativeLDC balance, expected 1353000, got 0")
	bs.CheckoutAccounts()
	tokenAcc.Add(constants.NativeToken, new(big.Int).SetUint64(constants.LDC))
	assert.NoError(itx.Apply(bctx, bs))

	assert.Equal(senderGas*bctx.Price,
		itx.(*TxDestroyToken).ldc.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(senderGas*100,
		itx.(*TxDestroyToken).miner.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(uint64(1), tokenAcc.Nonce())
	assert.Equal(uint16(0), tokenAcc.Threshold())
	assert.Equal(util.EthIDs{}, tokenAcc.Keepers())
	assert.Equal(ld.NativeAccount, tokenAcc.ld.Type)
	assert.Nil(tokenAcc.ld.Approver)
	assert.Nil(tokenAcc.ld.MaxTotalSupply)
	assert.Nil(tokenAcc.ld.Tokens[token])

	txData = &ld.TxData{
		Type:      ld.TypeCreateToken,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     1,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      sender,
		To:        &tokenid,
		Amount:    new(big.Int).SetUint64(10000000000000),
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	tt = txData.ToTransaction()
	tt.Gas = tt.RequiredGas(bctx.FeeConfig().ThresholdGas)
	senderGas += tt.Gas
	itx, err = NewTx(tt, true)
	assert.NoError(err)
	senderAcc.Add(constants.NativeToken, new(big.Int).SetUint64(10000000000000))
	assert.NoError(itx.Apply(bctx, bs))

	assert.Equal(senderGas*bctx.Price,
		itx.(*TxCreateToken).ldc.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(senderGas*100,
		itx.(*TxCreateToken).miner.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(uint64(1), tokenAcc.Nonce())
	assert.Equal(uint16(1), tokenAcc.Threshold())
	assert.Equal(util.EthIDs{util.Signer1.Address()}, tokenAcc.Keepers())
	assert.Equal(approver, *tokenAcc.ld.Approver)
	assert.Equal(constants.LDC*10, tokenAcc.ld.MaxTotalSupply.Uint64())
	assert.Equal(constants.LDC*10, tokenAcc.ld.Tokens[token].Uint64())
	assert.Equal(uint64(0), tokenAcc.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(uint64(10000000000000), tokenAcc.balanceOfAll(constants.NativeToken).Uint64())
	assert.Equal(constants.LDC*10, tokenAcc.balanceOf(token).Uint64())

	assert.NoError(bs.VerifyState())
}

func TestTxCreateTokenGenesis(t *testing.T) {
	assert := assert.New(t)

	bctx := NewMockBCtx()
	bs := bctx.MockBS()

	sender := util.Signer1.Address()

	// can not create the NativeToken
	input := &ld.TxAccounter{
		Threshold: ld.Uint16Ptr(1),
		Keepers:   &util.EthIDs{util.Signer1.Address()},
		Amount:    new(big.Int).SetUint64(constants.LDC * 10),
		Name:      "NativeToken",
	}
	txData := &ld.TxData{
		Type:      ld.TypeCreateToken,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      sender,
		To:        &constants.LDCAccount,
		Amount:    new(big.Int).SetUint64(10000000000000),
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	tt := txData.ToTransaction()
	tt.Gas = tt.RequiredGas(bctx.FeeConfig().ThresholdGas)
	_, err := NewTx(tt, true)
	assert.ErrorContains(err,
		"invalid to as token account, expected not 0x0000000000000000000000000000000000000000")

	// create the NativeToken in GenesisTx
	input = &ld.TxAccounter{
		Amount: bctx.Chain().MaxTotalSupply,
		Name:   "Linked Data Chain",
		Data:   []byte(strconv.Quote(bctx.Chain().Message)),
	}
	assert.NoError(input.SyntacticVerify())
	txData = &ld.TxData{
		Type:    ld.TypeCreateToken,
		ChainID: bctx.Chain().ChainID,
		From:    constants.GenesisAccount,
		To:      &constants.LDCAccount,
		Data:    input.Bytes(),
	}
	itx, err := NewGenesisTx(txData.ToTransaction())
	assert.NoError(err)

	assert.NoError(itx.(*TxCreateToken).ApplyGenesis(bctx, bs))

	ldcAcc := bs.MustAccount(constants.LDCAccount)
	assert.Equal(bctx.Chain().MaxTotalSupply.Uint64(),
		ldcAcc.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(bctx.Chain().MaxTotalSupply.Uint64(),
		ldcAcc.balanceOfAll(constants.NativeToken).Uint64())
	assert.Equal(uint64(0), itx.(*TxCreateToken).miner.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(uint64(0), itx.(*TxCreateToken).from.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(uint64(1), itx.(*TxCreateToken).from.Nonce())

	assert.Equal(uint64(0), ldcAcc.Nonce())
	assert.Equal(uint16(0), ldcAcc.Threshold())
	assert.Equal(util.EthIDs{}, ldcAcc.Keepers())
	assert.Nil(ldcAcc.ld.Approver)
	assert.Nil(ldcAcc.ld.ApproveList)
	assert.Equal(bctx.Chain().MaxTotalSupply.Uint64(), ldcAcc.ld.MaxTotalSupply.Uint64())
	assert.Equal(0, len(ldcAcc.ld.Tokens))

	jsondata, err := itx.MarshalJSON()
	assert.NoError(err)
	// fmt.Println(string(jsondata))
	assert.Equal(`{"type":"TypeCreateToken","chainID":2357,"nonce":0,"gasTip":0,"gasFeeCap":0,"from":"0xFFfFfFffFFfffFFfFFfFFFFFffFFFffffFfFFFfF","to":"0x0000000000000000000000000000000000000000","data":{"amount":1000000000000000000,"name":"Linked Data Chain","data":"Hello, LDVM!"},"gas":0,"id":"2df9TxdMzdFaZWnBxwSeQrcznCsu5Xg7vQqcPPJavU1cUa3CC5"}`, string(jsondata))

	// NativeToken cannot be destroy
	txData = &ld.TxData{
		Type:      ld.TypeDestroyToken,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      constants.LDCAccount,
		To:        &constants.GenesisAccount,
	}
	assert.NoError(txData.SignWith(util.Signer1))
	tt = txData.ToTransaction()
	tt.Gas = tt.RequiredGas(bctx.FeeConfig().ThresholdGas)
	itx, err = NewTx(tt, true)
	assert.ErrorContains(err, "TxBase.SyntacticVerify error: invalid from")

	assert.NoError(bs.VerifyState())
}
