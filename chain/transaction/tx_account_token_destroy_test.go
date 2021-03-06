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

func TestTxDestroyToken(t *testing.T) {
	assert := assert.New(t)

	// SyntacticVerify
	tx := &TxDestroyToken{}
	assert.ErrorContains(tx.SyntacticVerify(), "nil pointer")
	_, err := tx.MarshalJSON()
	assert.NoError(err)

	bctx := NewMockBCtx()
	bs := bctx.MockBS()
	token := ld.MustNewToken("$LDC")
	tokenid := token.EthID()
	sender := util.Signer1.Address()
	recipient := util.Signer2.Address()

	txData := &ld.TxData{
		Type:      ld.TypeDestroyToken,
		ChainID:   bctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      tokenid,
	}
	assert.NoError(txData.SyntacticVerify())
	_, err = NewTx2(txData.ToTransaction())
	assert.ErrorContains(err, "DeriveSigners error: no signature")

	txData = &ld.TxData{
		Type:      ld.TypeDestroyToken,
		ChainID:   bctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      tokenid,
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx2(txData.ToTransaction())
	assert.ErrorContains(err, "nil to as pledge recipient")

	txData = &ld.TxData{
		Type:      ld.TypeDestroyToken,
		ChainID:   bctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      tokenid,
		To:        &recipient,
		Token:     &token,
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx2(txData.ToTransaction())
	assert.ErrorContains(err, "invalid token, should be nil")

	txData = &ld.TxData{
		Type:      ld.TypeDestroyToken,
		ChainID:   bctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      tokenid,
		To:        &recipient,
		Amount:    big.NewInt(1),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx2(txData.ToTransaction())
	assert.ErrorContains(err, "invalid amount, should be nil")

	txData = &ld.TxData{
		Type:      ld.TypeDestroyToken,
		ChainID:   bctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      sender,
		To:        &recipient,
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx2(txData.ToTransaction())
	assert.ErrorContains(err, "invalid token 0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC")

	cfg := &ld.TxAccounter{
		Threshold: ld.Uint16Ptr(1),
		Keepers:   &util.EthIDs{util.Signer1.Address(), util.Signer2.Address()},
		Amount:    new(big.Int).SetUint64(constants.LDC * 10),
	}
	testToken := NewAccount(util.EthID(token))
	testToken.Init(bctx.FeeConfig().MinTokenPledge, 0, 0)
	assert.NoError(testToken.CreateToken(cfg))
	assert.Equal(false, testToken.valid(ld.TokenAccount))
	testToken.Add(constants.NativeToken, bctx.FeeConfig().MinTokenPledge)
	assert.Equal(true, testToken.valid(ld.TokenAccount))
	bs.AC[testToken.id] = testToken

	txData = &ld.TxData{
		Type:      ld.TypeTransfer,
		ChainID:   bctx.ChainConfig().ChainID,
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
	itx, err := NewTx2(tt)
	assert.NoError(err)

	bs.CommitAccounts()
	assert.ErrorContains(itx.Apply(bctx, bs),
		"insufficient NativeLDC balance, expected 823900, got 0")
	bs.CheckoutAccounts()
	testToken.Add(constants.NativeToken, new(big.Int).SetUint64(constants.LDC))
	assert.NoError(itx.Apply(bctx, bs))

	recipientAcc := bs.MustAccount(recipient)

	tokenGas := tt.Gas()
	assert.Equal(tokenGas*bctx.Price,
		itx.(*TxTransfer).ldc.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(tokenGas*100,
		itx.(*TxTransfer).miner.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(constants.LDC-tokenGas*(bctx.Price+100),
		testToken.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(constants.LDC*9,
		testToken.balanceOf(token).Uint64())
	assert.Equal(uint64(0),
		recipientAcc.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(constants.LDC,
		recipientAcc.balanceOf(token).Uint64())
	assert.Equal(uint64(1), testToken.Nonce())

	txData = &ld.TxData{
		Type:      ld.TypeDestroyToken,
		ChainID:   bctx.ChainConfig().ChainID,
		Nonce:     1,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      testToken.id,
		To:        &recipient,
	}
	assert.NoError(txData.SignWith(util.Signer1))
	tt = txData.ToTransaction()
	itx, err = NewTx2(tt)
	assert.NoError(err)

	bs.CommitAccounts()
	assert.ErrorContains(itx.Apply(bctx, bs), "invalid signature for keepers")
	bs.CheckoutAccounts()

	assert.NoError(txData.SignWith(util.Signer2))
	tt = txData.ToTransaction()
	itx, err = NewTx2(tt)
	assert.NoError(err)

	bs.CommitAccounts()
	assert.ErrorContains(itx.Apply(bctx, bs),
		"some token in the use, maxTotalSupply expected 10000000000, got 9000000000")
	bs.CheckoutAccounts()

	txData = &ld.TxData{
		Type:      ld.TypeTransfer,
		ChainID:   bctx.ChainConfig().ChainID,
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
	itx, err = NewTx2(tt)
	assert.NoError(err)
	recipientAcc.Add(constants.NativeToken, new(big.Int).SetUint64(constants.LDC))
	assert.NoError(itx.Apply(bctx, bs))

	recipientGas := tt.Gas()
	assert.Equal((tokenGas+recipientGas)*bctx.Price,
		itx.(*TxTransfer).ldc.balanceOf(constants.NativeToken).Uint64())
	assert.Equal((tokenGas+recipientGas)*100,
		itx.(*TxTransfer).miner.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(bctx.FeeConfig().MinTokenPledge.Uint64()+constants.LDC-tokenGas*(bctx.Price+100),
		testToken.balanceOfAll(constants.NativeToken).Uint64())
	assert.Equal(constants.LDC*10, testToken.balanceOf(token).Uint64())
	assert.Equal(constants.LDC-recipientGas*(bctx.Price+100),
		recipientAcc.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(uint64(0), recipientAcc.balanceOf(token).Uint64())

	txData = &ld.TxData{
		Type:      ld.TypeDestroyToken,
		ChainID:   bctx.ChainConfig().ChainID,
		Nonce:     1,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      testToken.id,
		To:        &recipient,
	}
	assert.NoError(txData.SignWith(util.Signer1))
	assert.NoError(txData.SignWith(util.Signer2))
	tt = txData.ToTransaction()
	itx, err = NewTx2(tt)
	assert.NoError(err)
	assert.NoError(itx.Apply(bctx, bs))

	tokenGas += tt.Gas()
	assert.Equal((tokenGas+recipientGas)*bctx.Price,
		itx.(*TxDestroyToken).ldc.balanceOf(constants.NativeToken).Uint64())
	assert.Equal((tokenGas+recipientGas)*100,
		itx.(*TxDestroyToken).miner.balanceOf(constants.NativeToken).Uint64())

	assert.Equal(uint64(0), testToken.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(uint64(0), testToken.balanceOfAll(constants.NativeToken).Uint64())
	assert.Equal(uint64(0), testToken.balanceOf(token).Uint64())
	assert.Equal(uint64(0), testToken.balanceOfAll(token).Uint64())
	assert.Equal(bctx.FeeConfig().MinTokenPledge.Uint64()+constants.LDC*2-(tokenGas+recipientGas)*(bctx.Price+100),
		recipientAcc.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(uint64(0), recipientAcc.balanceOf(token).Uint64())
	assert.Equal(uint64(0), recipientAcc.balanceOfAll(token).Uint64())

	assert.Equal(uint64(2), testToken.Nonce())
	assert.Equal(uint16(0), testToken.Threshold())
	assert.Equal(util.EthIDs{}, testToken.Keepers())

	jsondata, err := itx.MarshalJSON()
	assert.NoError(err)
	// fmt.Println(string(jsondata))
	assert.Equal(`{"type":"TypeDestroyToken","chainID":2357,"nonce":1,"gasTip":100,"gasFeeCap":1000,"from":"0x00000000000000000000000000000000244C4443","to":"0x44171C37Ff5D7B7bb8dcad5C81f16284A229e641","signatures":["947812955cc5ca47cec02f6843acd0cbf6c1600d8cf79e8ff3d499d4a6e87fbf69db6cc62b65b8d42bd7de90abbf19d342530f1feaad9123dfa6727ce7b102e301","d3d4a5facb761d38b3b40677c00a66bcdfec751141157ac5e35f69fc5024c70460d085872d8cec8a227084890126d9785e62db834656a25e1b3ea707ddebfb6401"],"id":"yzxEZNCXs37uZhQmpVejzsAmmofqCTmptx1mcACX3H3HoTPFy"}`, string(jsondata))

	assert.NoError(bs.VerifyState())
}

func TestTxDestroyTokenWithApproverAndLending(t *testing.T) {
	assert := assert.New(t)

	bctx := NewMockBCtx()
	bs := bctx.MockBS()
	token := ld.MustNewToken("$LDC")
	tokenid := util.EthID(token)

	sender := util.Signer1.Address()
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
		ChainID:   bctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      sender,
		To:        &tokenid,
		Amount:    bctx.FeeConfig().MinTokenPledge,
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	tt := txData.ToTransaction()
	itx, err := NewTx2(tt)
	assert.NoError(err)

	tokenAcc := bs.MustAccount(tokenid)
	senderAcc := bs.MustAccount(sender)
	senderAcc.Add(constants.NativeToken, bctx.FeeConfig().MinTokenPledge)
	senderAcc.Add(constants.NativeToken, new(big.Int).SetUint64(constants.LDC))
	assert.NoError(itx.Apply(bctx, bs))

	senderGas := tt.Gas()
	assert.Equal(senderGas*bctx.Price,
		itx.(*TxCreateToken).ldc.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(senderGas*100,
		itx.(*TxCreateToken).miner.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(uint64(0),
		tokenAcc.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(bctx.FeeConfig().MinTokenPledge.Uint64(),
		tokenAcc.balanceOfAll(constants.NativeToken).Uint64())
	assert.Equal(constants.LDC*10,
		tokenAcc.balanceOf(token).Uint64())
	assert.Equal(constants.LDC-senderGas*(bctx.Price+100),
		senderAcc.balanceOf(constants.NativeToken).Uint64())

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
		ChainID:   bctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      tokenid,
		Data:      ld.MustMarshal(lcfg),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	tt = txData.ToTransaction()
	itx, err = NewTx2(tt)
	assert.NoError(err)
	senderAcc.Add(constants.NativeToken, new(big.Int).SetUint64(constants.LDC))

	bs.CommitAccounts()
	assert.ErrorContains(itx.Apply(bctx, bs),
		"TxOpenLending.Apply error: invalid signature for approver")
	bs.CheckoutAccounts()

	assert.NoError(txData.SignWith(util.Signer2))
	tt = txData.ToTransaction()
	itx, err = NewTx2(tt)
	assert.NoError(err)

	bs.CommitAccounts()
	assert.ErrorContains(itx.Apply(bctx, bs),
		"insufficient NativeLDC balance, expected 2173600, got 0")
	bs.CheckoutAccounts()

	tokenAcc.Add(constants.NativeToken, new(big.Int).SetUint64(constants.LDC))
	assert.NoError(itx.Apply(bctx, bs))

	tokenGas := tt.Gas()
	assert.Equal((tokenGas+senderGas)*bctx.Price,
		itx.(*TxOpenLending).ldc.balanceOf(constants.NativeToken).Uint64())
	assert.Equal((tokenGas+senderGas)*100,
		itx.(*TxOpenLending).miner.balanceOf(constants.NativeToken).Uint64())
	assert.NotNil(tokenAcc.ld.Lending)
	assert.Equal(uint64(1), tokenAcc.Nonce())

	// AddNonceTable
	ns := []uint64{bs.Timestamp() + 1, 1, 2, 3}
	ndData, err := util.MarshalCBOR(ns)
	assert.NoError(err)
	txData = &ld.TxData{
		Type:      ld.TypeAddNonceTable,
		ChainID:   bctx.ChainConfig().ChainID,
		Nonce:     1,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      tokenid,
		Data:      ndData,
	}
	assert.NoError(txData.SignWith(util.Signer1))
	tt = txData.ToTransaction()
	tt.Timestamp = bs.Timestamp()
	itx, err = NewTx2(tt)
	assert.NoError(err)
	assert.NoError(itx.Apply(bctx, bs))

	tokenGas += tt.Gas()
	assert.Equal((tokenGas+senderGas)*bctx.Price,
		itx.(*TxAddNonceTable).ldc.balanceOf(constants.NativeToken).Uint64())
	assert.Equal((tokenGas+senderGas)*100,
		itx.(*TxAddNonceTable).miner.balanceOf(constants.NativeToken).Uint64())
	assert.Equal([]uint64{1, 2, 3}, tokenAcc.ld.NonceTable[bs.Timestamp()+1])
	assert.Equal(uint64(2), tokenAcc.Nonce())

	// Borrow
	tf := &ld.TxTransfer{
		Nonce:  3,
		From:   &tokenAcc.id,
		To:     &sender,
		Token:  &token,
		Amount: new(big.Int).SetUint64(constants.LDC),
		Expire: bs.Timestamp() + 1,
	}
	assert.NoError(tf.SyntacticVerify())
	txData = &ld.TxData{
		Type:      ld.TypeBorrow,
		ChainID:   bctx.ChainConfig().ChainID,
		Nonce:     1,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      sender,
		To:        &tokenid,
		Token:     &token,
		Data:      tf.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	assert.NoError(txData.ExSignWith(util.Signer1))
	tt = txData.ToTransaction()
	tt.Timestamp = bs.Timestamp()
	itx, err = NewTx2(tt)
	assert.NoError(err)
	assert.NoError(itx.Apply(bctx, bs))

	senderGas += tt.Gas()
	assert.Equal((tokenGas+senderGas)*bctx.Price,
		itx.(*TxBorrow).ldc.balanceOf(constants.NativeToken).Uint64())
	assert.Equal((tokenGas+senderGas)*100,
		itx.(*TxBorrow).miner.balanceOf(constants.NativeToken).Uint64())
	assert.Equal([]uint64{1, 2}, tokenAcc.ld.NonceTable[bs.Timestamp()+1])
	assert.Equal(constants.LDC*9, tokenAcc.balanceOf(token).Uint64())
	assert.Equal(constants.LDC, senderAcc.balanceOf(token).Uint64())

	// DestroyToken
	txData = &ld.TxData{
		Type:      ld.TypeDestroyToken,
		ChainID:   bctx.ChainConfig().ChainID,
		Nonce:     2,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      tokenid,
		To:        &sender,
	}
	assert.NoError(txData.SignWith(util.Signer1))
	tt = txData.ToTransaction()
	itx, err = NewTx2(tt)
	assert.NoError(err)

	bs.CommitAccounts()
	assert.ErrorContains(itx.Apply(bctx, bs),
		"TxDestroyToken.Apply error: invalid signature for approver")
	bs.CheckoutAccounts()

	assert.NoError(txData.SignWith(util.Signer2))
	tt = txData.ToTransaction()
	itx, err = NewTx2(tt)
	assert.NoError(err)
	bs.CommitAccounts()
	assert.ErrorContains(itx.Apply(bctx, bs),
		"TxDestroyToken.Apply error: Account(0x00000000000000000000000000000000244C4443).DestroyToken error: some token in the use, maxTotalSupply expected 10000000000, got 9000000000")
	bs.CheckoutAccounts()

	// TypeRepay
	txData = &ld.TxData{
		Type:      ld.TypeRepay,
		ChainID:   bctx.ChainConfig().ChainID,
		Nonce:     2,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      sender,
		To:        &tokenid,
		Token:     &token,
		Amount:    new(big.Int).SetUint64(constants.LDC),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	tt = txData.ToTransaction()
	itx, err = NewTx2(tt)
	assert.NoError(err)
	assert.NoError(itx.Apply(bctx, bs))

	senderGas += tt.Gas()
	assert.Equal((tokenGas+senderGas)*bctx.Price,
		itx.(*TxRepay).ldc.balanceOf(constants.NativeToken).Uint64())
	assert.Equal((tokenGas+senderGas)*100,
		itx.(*TxRepay).miner.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(constants.LDC*10, tokenAcc.balanceOf(token).Uint64())
	assert.Equal(uint64(0), senderAcc.balanceOf(token).Uint64())

	// DestroyToken
	txData = &ld.TxData{
		Type:      ld.TypeDestroyToken,
		ChainID:   bctx.ChainConfig().ChainID,
		Nonce:     2,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      tokenid,
		To:        &sender,
	}
	assert.NoError(txData.SignWith(util.Signer1))
	assert.NoError(txData.SignWith(util.Signer2))
	tt = txData.ToTransaction()
	itx, err = NewTx2(tt)
	assert.NoError(err)
	assert.NoError(itx.Apply(bctx, bs))

	tokenGas += tt.Gas()
	assert.Equal((tokenGas+senderGas)*bctx.Price,
		itx.(*TxDestroyToken).ldc.balanceOf(constants.NativeToken).Uint64())
	assert.Equal((tokenGas+senderGas)*100,
		itx.(*TxDestroyToken).miner.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(uint64(3), tokenAcc.Nonce())
	assert.Equal(uint16(0), tokenAcc.Threshold())
	assert.Equal(util.EthIDs{}, tokenAcc.Keepers())
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
	assert.Equal(`{"type":"TypeDestroyToken","chainID":2357,"nonce":2,"gasTip":100,"gasFeeCap":1000,"from":"0x00000000000000000000000000000000244C4443","to":"0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC","signatures":["421ff11ef9716a8e3a39e1540901e60c5afc5f9264cb67e8a560e307643ddc024d7b268017924304a372817581d68a2a3b21a874c390825cd7197326f5e8971601","ca327eff261fb9383025cb5763cba84acf70b11c004610aef4d7ff89bc6fa3307e95964358708cf4f96d7ee77eb6fdbd3012d1b696cbf6514b1fdb4d29ce88aa00"],"id":"726954uuMm8zLHRw6ZH2dV9gPTuea4V5sPZ6hkeEVUfszCqv5"}`, string(jsondata))

	// DestroyToken again
	txData = &ld.TxData{
		Type:      ld.TypeDestroyToken,
		ChainID:   bctx.ChainConfig().ChainID,
		Nonce:     3,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      tokenid,
		To:        &sender,
	}
	assert.NoError(txData.SignWith(util.Signer1))
	assert.NoError(txData.SignWith(util.Signer2))
	tt = txData.ToTransaction()
	itx, err = NewTx2(tt)
	assert.NoError(err)
	assert.ErrorContains(itx.Apply(bctx, bs),
		"TxDestroyToken.Apply error: invalid signatures for sender")

	assert.NoError(bs.VerifyState())
}
