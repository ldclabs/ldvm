// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package transaction

import (
	"math/big"
	"testing"

	"github.com/ldclabs/ldvm/constants"
	"github.com/ldclabs/ldvm/ld"
	"github.com/ldclabs/ldvm/util"
	"github.com/stretchr/testify/assert"
)

func TestTxDestroyTokenAccount(t *testing.T) {
	assert := assert.New(t)

	// SyntacticVerify
	tx := &TxDestroyTokenAccount{}
	assert.ErrorContains(tx.SyntacticVerify(), "nil pointer")
	_, err := tx.MarshalJSON()
	assert.NoError(err)

	bctx := NewMockBCtx()
	bs := NewMockBS(bctx)
	token := ld.MustNewToken("$LDC")

	from, err := bs.LoadAccount(util.Signer1.Address())
	assert.NoError(err)
	recipient := util.Signer2.Address()

	txData := &ld.TxData{
		Type:      ld.TypeDestroyToken,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      from.id,
	}
	assert.NoError(txData.SyntacticVerify())
	_, err = NewTx(txData.ToTransaction(), true)
	assert.ErrorContains(err, "DeriveSigners: no signature")

	txData = &ld.TxData{
		Type:      ld.TypeDestroyToken,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      from.id,
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx(txData.ToTransaction(), true)
	assert.ErrorContains(err, "nil to as pledge recipient")

	txData = &ld.TxData{
		Type:      ld.TypeDestroyToken,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      from.id,
		To:        &recipient,
		Token:     &token,
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx(txData.ToTransaction(), true)
	assert.ErrorContains(err, "invalid token, should be nil")

	txData = &ld.TxData{
		Type:      ld.TypeDestroyToken,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      from.id,
		To:        &recipient,
		Amount:    big.NewInt(1),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx(txData.ToTransaction(), true)
	assert.ErrorContains(err, "invalid amount, should be nil")

	txData = &ld.TxData{
		Type:      ld.TypeDestroyToken,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      from.id,
		To:        &recipient,
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx(txData.ToTransaction(), true)
	assert.ErrorContains(err, "invalid token 0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC")

	cfg := &ld.TxAccounter{
		Threshold: ld.Uint16Ptr(1),
		Keepers:   &util.EthIDs{util.Signer1.Address(), util.Signer2.Address()},
		Amount:    new(big.Int).SetUint64(constants.LDC * 10),
	}
	testToken := NewAccount(util.EthID(token))
	testToken.Init(new(big.Int).SetUint64(constants.LDC), 0, 0)
	assert.NoError(testToken.CheckCreateToken(cfg))
	assert.NoError(testToken.CreateToken(cfg))
	assert.Equal(false, testToken.valid(ld.TokenAccount))
	testToken.Add(constants.NativeToken, new(big.Int).SetUint64(constants.LDC))
	assert.Equal(true, testToken.valid(ld.TokenAccount))
	bs.AC[testToken.id] = testToken

	txData = &ld.TxData{
		Type:      ld.TypeTransfer,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      testToken.id,
		To:        &recipient,
		Token:     &token,
		Amount:    new(big.Int).SetUint64(constants.LDC),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	tt := txData.ToTransaction()
	itx, err := NewTx(tt, true)
	assert.NoError(err)
	assert.ErrorContains(itx.Verify(bctx, bs), "invalid gas, expected 145, got 0")
	tt.Gas = tt.RequiredGas(bctx.FeeConfig().ThresholdGas)
	itx, err = NewTx(tt, true)
	assert.NoError(err)
	assert.ErrorContains(itx.Verify(bctx, bs),
		"insufficient NativeLDC balance, expected 159500, got 0")
	testToken.Add(constants.NativeToken, new(big.Int).SetUint64(constants.LDC))
	assert.NoError(itx.Verify(bctx, bs))
	assert.NoError(itx.Accept(bctx, bs))

	acc, err := bs.LoadAccount(recipient)
	assert.NoError(err)
	ldc, err := bs.LoadAccount(constants.LDCAccount)
	assert.NoError(err)
	miner, err := bs.LoadMiner(bctx.Miner())
	assert.NoError(err)

	ldcBa := ldc.balanceOf(constants.NativeToken).Uint64()
	minerBa := miner.balanceOf(constants.NativeToken).Uint64()
	tokenBa := testToken.balanceOf(constants.NativeToken).Uint64()

	tx2 := itx.(*TxTransfer)
	assert.Equal(tx2.ld.Gas*bctx.Price, ldcBa)
	assert.Equal(tx2.ld.Gas*100, minerBa)
	assert.Equal(constants.LDC-tx2.ld.Gas*(bctx.Price+100), tokenBa)
	assert.Equal(constants.LDC*9,
		testToken.balanceOf(token).Uint64())
	assert.Equal(uint64(0),
		acc.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(constants.LDC,
		acc.balanceOf(token).Uint64())
	assert.Equal(uint64(1), testToken.Nonce())

	txData = &ld.TxData{
		Type:      ld.TypeDestroyToken,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     1,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      testToken.id,
		To:        &recipient,
	}
	assert.NoError(txData.SignWith(util.Signer1))
	tt = txData.ToTransaction()
	tt.Gas = tt.RequiredGas(bctx.FeeConfig().ThresholdGas)
	itx, err = NewTx(tt, true)
	assert.NoError(err)
	assert.ErrorContains(itx.Verify(bctx, bs), "invalid signature for keepers")

	assert.NoError(txData.SignWith(util.Signer2))
	tt = txData.ToTransaction()
	tt.Gas = tt.RequiredGas(bctx.FeeConfig().ThresholdGas)
	itx, err = NewTx(tt, true)
	assert.NoError(err)
	assert.ErrorContains(itx.Verify(bctx, bs),
		"some token in the use, expected 10000000000, got 9000000000")

	txData = &ld.TxData{
		Type:      ld.TypeTransfer,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      recipient,
		To:        &testToken.id,
		Token:     &token,
		Amount:    new(big.Int).SetUint64(constants.LDC),
	}
	assert.NoError(txData.SignWith(util.Signer2))
	tt = txData.ToTransaction()
	tt.Gas = tt.RequiredGas(bctx.FeeConfig().ThresholdGas)
	itx, err = NewTx(tt, true)
	assert.NoError(err)
	acc.Add(constants.NativeToken, new(big.Int).SetUint64(constants.LDC))
	assert.NoError(itx.Verify(bctx, bs))
	assert.NoError(itx.Accept(bctx, bs))

	tx2 = itx.(*TxTransfer)
	assert.Equal(tx2.ld.Gas*bctx.Price+ldcBa,
		ldc.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(tx2.ld.Gas*100+minerBa,
		miner.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(tokenBa, testToken.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(constants.LDC+tokenBa, testToken.balanceOfAll(constants.NativeToken).Uint64())
	assert.Equal(constants.LDC*10, testToken.balanceOf(token).Uint64())
	assert.Equal(constants.LDC-tx2.ld.Gas*(bctx.Price+100),
		acc.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(uint64(0), acc.balanceOf(token).Uint64())

	ldcBa = ldc.balanceOf(constants.NativeToken).Uint64()
	minerBa = miner.balanceOf(constants.NativeToken).Uint64()
	tokenBa = testToken.balanceOf(constants.NativeToken).Uint64()
	accBa := acc.balanceOf(constants.NativeToken).Uint64()

	txData = &ld.TxData{
		Type:      ld.TypeDestroyToken,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     1,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      testToken.id,
		To:        &recipient,
	}
	assert.NoError(txData.SignWith(util.Signer1))
	assert.NoError(txData.SignWith(util.Signer2))
	tt = txData.ToTransaction()
	tt.Gas = tt.RequiredGas(bctx.FeeConfig().ThresholdGas)
	itx, err = NewTx(tt, true)
	assert.NoError(err)
	assert.NoError(itx.Verify(bctx, bs))
	assert.NoError(itx.Accept(bctx, bs))

	tx = itx.(*TxDestroyTokenAccount)
	assert.Equal(tx.ld.Gas*bctx.Price+ldcBa,
		ldc.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(tx.ld.Gas*100+minerBa,
		miner.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(uint64(0), testToken.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(uint64(0), testToken.balanceOfAll(constants.NativeToken).Uint64())
	assert.Equal(uint64(0), testToken.balanceOf(token).Uint64())
	assert.Equal(constants.LDC+accBa+tokenBa-tx.ld.Gas*(bctx.Price+100),
		acc.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(uint64(0), acc.balanceOf(token).Uint64())

	assert.Equal(uint64(2), testToken.Nonce())
	assert.Equal(uint16(0), testToken.Threshold())
	assert.Equal(util.EthIDs{}, testToken.Keepers())

	jsondata, err := itx.MarshalJSON()
	assert.NoError(err)
	// fmt.Println(string(jsondata))
	assert.Equal(`{"type":"TypeDestroyToken","chainID":2357,"nonce":1,"gasTip":100,"gasFeeCap":1000,"from":"0x00000000000000000000000000000000244C4443","to":"0x44171C37Ff5D7B7bb8dcad5C81f16284A229e641","signatures":["947812955cc5ca47cec02f6843acd0cbf6c1600d8cf79e8ff3d499d4a6e87fbf69db6cc62b65b8d42bd7de90abbf19d342530f1feaad9123dfa6727ce7b102e301","d3d4a5facb761d38b3b40677c00a66bcdfec751141157ac5e35f69fc5024c70460d085872d8cec8a227084890126d9785e62db834656a25e1b3ea707ddebfb6401"],"gas":1230,"id":"yzxEZNCXs37uZhQmpVejzsAmmofqCTmptx1mcACX3H3HoTPFy"}`, string(jsondata))

	assert.NoError(bs.VerifyState())
}

func TestTxDestroyTokenAccountWithApproverAndLending(t *testing.T) {
	assert := assert.New(t)

	bctx := NewMockBCtx()
	bs := NewMockBS(bctx)
	token := ld.MustNewToken("$LDC")
	tokenid := util.EthID(token)

	from, err := bs.LoadAccount(util.Signer1.Address())
	assert.NoError(err)
	approver := util.Signer2.Address()

	// CreateToken
	input := &ld.TxAccounter{
		Threshold:   ld.Uint16Ptr(1),
		Keepers:     &util.EthIDs{util.Signer1.Address()},
		Approver:    &approver,
		ApproveList: ld.TxTypes{ld.TypeOpenLending, ld.TypeDestroyToken},
		Amount:      new(big.Int).SetUint64(constants.LDC * 10),
		Name:        "LDC Token",
	}
	assert.NoError(input.SyntacticVerify())
	txData := &ld.TxData{
		Type:      ld.TypeCreateToken,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      from.id,
		To:        &tokenid,
		Amount:    bctx.FeeConfig().MinTokenPledge,
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	tt := txData.ToTransaction()
	tt.Gas = tt.RequiredGas(bctx.FeeConfig().ThresholdGas)
	itx, err := NewTx(tt, true)
	assert.NoError(err)
	from.Add(constants.NativeToken, bctx.FeeConfig().MinTokenPledge)
	from.Add(constants.NativeToken, new(big.Int).SetUint64(constants.LDC))
	assert.NoError(itx.Verify(bctx, bs))
	assert.NoError(itx.Accept(bctx, bs))

	tokenAcc, err := bs.LoadAccount(tokenid)
	assert.NoError(err)
	ldc, err := bs.LoadAccount(constants.LDCAccount)
	assert.NoError(err)
	miner, err := bs.LoadMiner(bctx.Miner())
	assert.NoError(err)

	assert.Equal(tt.Gas*bctx.Price,
		ldc.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(tt.Gas*100,
		miner.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(uint64(0), tokenAcc.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(bctx.FeeConfig().MinTokenPledge.Uint64(),
		tokenAcc.balanceOfAll(constants.NativeToken).Uint64())
	assert.Equal(constants.LDC*10, tokenAcc.balanceOf(token).Uint64())
	assert.Equal(constants.LDC-tt.Gas*(bctx.Price+100),
		from.balanceOf(constants.NativeToken).Uint64())

	assert.Equal(uint16(1), tokenAcc.Threshold())
	assert.Equal(util.EthIDs{util.Signer1.Address()}, tokenAcc.Keepers())
	assert.Equal(approver, *tokenAcc.ld.Approver)
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
	txData = &ld.TxData{
		Type:      ld.TypeOpenLending,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      tokenid,
		Data:      ld.MustMarshal(lcfg),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	tt = txData.ToTransaction()
	tt.Gas = tt.RequiredGas(bctx.FeeConfig().ThresholdGas)
	itx, err = NewTx(tt, true)
	assert.NoError(err)
	from.Add(constants.NativeToken, new(big.Int).SetUint64(constants.LDC))
	assert.ErrorContains(itx.Verify(bctx, bs),
		"TxBase.Verify failed: invalid signature for approver")

	assert.NoError(txData.SignWith(util.Signer2))
	tt = txData.ToTransaction()
	tt.Gas = tt.RequiredGas(bctx.FeeConfig().ThresholdGas)
	itx, err = NewTx(tt, true)
	assert.NoError(err)
	assert.ErrorContains(itx.Verify(bctx, bs),
		"insufficient NativeLDC balance, expected 1416800, got 0")
	tokenAcc.Add(constants.NativeToken, new(big.Int).SetUint64(constants.LDC))
	assert.NoError(itx.Verify(bctx, bs))
	assert.NoError(itx.Accept(bctx, bs))

	assert.NotNil(tokenAcc.ld.Lending)
	assert.NotNil(tokenAcc.ld.LendingLedger)
	assert.Equal(uint64(1), tokenAcc.Nonce())

	// AddNonceTable
	ns := []uint64{bs.Timestamp() + 1, 1, 2, 3}
	ndData, err := ld.MarshalCBOR(ns)
	assert.NoError(err)
	txData = &ld.TxData{
		Type:      ld.TypeAddNonceTable,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     1,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      tokenid,
		Data:      ndData,
	}
	assert.NoError(txData.SignWith(util.Signer1))
	tt = txData.ToTransaction()
	tt.Timestamp = bs.Timestamp()
	tt.Gas = tt.RequiredGas(bctx.FeeConfig().ThresholdGas)
	itx, err = NewTx(tt, true)
	assert.NoError(err)
	assert.NoError(itx.Verify(bctx, bs))
	assert.NoError(itx.Accept(bctx, bs))

	assert.Equal([]uint64{1, 2, 3}, tokenAcc.ld.NonceTable[bs.Timestamp()+1])
	assert.Equal(uint64(2), tokenAcc.Nonce())

	// Borrow
	tf := &ld.TxTransfer{
		Nonce:  3,
		From:   &tokenAcc.id,
		To:     &from.id,
		Token:  &token,
		Amount: new(big.Int).SetUint64(constants.LDC),
		Expire: bs.Timestamp() + 1,
	}
	assert.NoError(tf.SyntacticVerify())
	txData = &ld.TxData{
		Type:      ld.TypeBorrow,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     1,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      from.id,
		To:        &tokenid,
		Token:     &token,
		Data:      tf.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	assert.NoError(txData.ExSignWith(util.Signer1))
	tt = txData.ToTransaction()
	tt.Timestamp = bs.Timestamp()
	tt.Gas = tt.RequiredGas(bctx.FeeConfig().ThresholdGas)
	itx, err = NewTx(tt, true)
	assert.NoError(err)
	assert.NoError(itx.Verify(bctx, bs))
	assert.NoError(itx.Accept(bctx, bs))

	assert.Equal([]uint64{1, 2}, tokenAcc.ld.NonceTable[bs.Timestamp()+1])
	assert.Equal(constants.LDC*9, tokenAcc.balanceOf(token).Uint64())
	assert.Equal(constants.LDC, from.balanceOf(token).Uint64())

	// DestroyToken
	txData = &ld.TxData{
		Type:      ld.TypeDestroyToken,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     2,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      tokenid,
		To:        &from.id,
	}
	assert.NoError(txData.SignWith(util.Signer1))
	tt = txData.ToTransaction()
	tt.Gas = tt.RequiredGas(bctx.FeeConfig().ThresholdGas)
	itx, err = NewTx(tt, true)
	assert.NoError(err)
	assert.ErrorContains(itx.Verify(bctx, bs),
		"TxBase.Verify failed: invalid signature for approver")

	assert.NoError(txData.SignWith(util.Signer2))
	tt = txData.ToTransaction()
	tt.Gas = tt.RequiredGas(bctx.FeeConfig().ThresholdGas)
	itx, err = NewTx(tt, true)
	assert.NoError(err)
	assert.ErrorContains(itx.Verify(bctx, bs),
		"please repay all before close")

	// TypeRepay
	txData = &ld.TxData{
		Type:      ld.TypeRepay,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     2,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      from.id,
		To:        &tokenid,
		Token:     &token,
		Amount:    new(big.Int).SetUint64(constants.LDC),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	tt = txData.ToTransaction()
	tt.Gas = tt.RequiredGas(bctx.FeeConfig().ThresholdGas)
	itx, err = NewTx(tt, true)
	assert.NoError(err)
	assert.NoError(itx.Verify(bctx, bs))
	assert.NoError(itx.Accept(bctx, bs))

	assert.Equal(constants.LDC*10, tokenAcc.balanceOf(token).Uint64())
	assert.Equal(uint64(0), from.balanceOf(token).Uint64())

	// DestroyToken
	txData = &ld.TxData{
		Type:      ld.TypeDestroyToken,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     2,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      tokenid,
		To:        &from.id,
	}
	assert.NoError(txData.SignWith(util.Signer1))
	assert.NoError(txData.SignWith(util.Signer2))
	tt = txData.ToTransaction()
	tt.Gas = tt.RequiredGas(bctx.FeeConfig().ThresholdGas)
	itx, err = NewTx(tt, true)
	assert.NoError(err)
	assert.NoError(itx.Verify(bctx, bs))
	assert.NoError(itx.Accept(bctx, bs))

	assert.Equal(uint64(3), tokenAcc.Nonce())
	assert.Equal(uint16(0), tokenAcc.Threshold())
	assert.Equal(util.EthIDs{}, tokenAcc.Keepers())
	assert.Equal(make(map[uint64][]uint64), tokenAcc.ld.NonceTable)
	assert.Equal(ld.NativeAccount, tokenAcc.ld.Type)
	assert.Nil(tokenAcc.ld.Approver)
	assert.Nil(tokenAcc.ld.ApproveList)
	assert.Nil(tokenAcc.ld.Lending)
	assert.Nil(tokenAcc.ld.LendingLedger)
	assert.Nil(tokenAcc.ld.MaxTotalSupply)
	assert.Nil(tokenAcc.ld.Tokens[token])
	assert.Equal(uint64(0), tokenAcc.balanceOfAll(token).Uint64())
	assert.Equal(uint64(0), tokenAcc.balanceOfAll(constants.NativeToken).Uint64())

	jsondata, err := itx.MarshalJSON()
	assert.NoError(err)
	// fmt.Println(string(jsondata))
	assert.Equal(`{"type":"TypeDestroyToken","chainID":2357,"nonce":2,"gasTip":100,"gasFeeCap":1000,"from":"0x00000000000000000000000000000000244C4443","to":"0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC","signatures":["421ff11ef9716a8e3a39e1540901e60c5afc5f9264cb67e8a560e307643ddc024d7b268017924304a372817581d68a2a3b21a874c390825cd7197326f5e8971601","ca327eff261fb9383025cb5763cba84acf70b11c004610aef4d7ff89bc6fa3307e95964358708cf4f96d7ee77eb6fdbd3012d1b696cbf6514b1fdb4d29ce88aa00"],"gas":1230,"id":"726954uuMm8zLHRw6ZH2dV9gPTuea4V5sPZ6hkeEVUfszCqv5"}`, string(jsondata))

	// DestroyToken again
	txData = &ld.TxData{
		Type:      ld.TypeDestroyToken,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     3,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      tokenid,
		To:        &from.id,
	}
	assert.NoError(txData.SignWith(util.Signer1))
	assert.NoError(txData.SignWith(util.Signer2))
	tt = txData.ToTransaction()
	tt.Gas = tt.RequiredGas(bctx.FeeConfig().ThresholdGas)
	itx, err = NewTx(tt, true)
	assert.NoError(err)
	assert.ErrorContains(itx.Verify(bctx, bs),
		"TxBase.Verify failed: invalid signatures for sender")

	assert.NoError(bs.VerifyState())
}
