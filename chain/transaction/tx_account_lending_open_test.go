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

func TestTxOpenLending(t *testing.T) {
	assert := assert.New(t)

	// SyntacticVerify
	tx := &TxOpenLending{}
	assert.ErrorContains(tx.SyntacticVerify(), "nil pointer")
	_, err := tx.MarshalJSON()
	assert.NoError(err)

	bctx := NewMockBCtx()
	bs := bctx.MockBS()
	token := ld.MustNewToken("$LDC")
	sender := util.Signer1.Address()
	approver := util.Signer2.Address()

	txData := &ld.TxData{
		Type:      ld.TypeOpenLending,
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
		Type:      ld.TypeOpenLending,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      sender,
		To:        &constants.GenesisAccount,
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx(txData.ToTransaction(), true)
	assert.ErrorContains(err, "invalid to, should be nil")

	txData = &ld.TxData{
		Type:      ld.TypeOpenLending,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      sender,
		Token:     &constants.NativeToken,
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx(txData.ToTransaction(), true)
	assert.ErrorContains(err, "invalid token, should be nil")

	txData = &ld.TxData{
		Type:      ld.TypeOpenLending,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      sender,
		Amount:    big.NewInt(1),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx(txData.ToTransaction(), true)
	assert.ErrorContains(err, "nil to together with amount")

	txData = &ld.TxData{
		Type:      ld.TypeOpenLending,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      sender,
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx(txData.ToTransaction(), true)
	assert.ErrorContains(err, "invalid data")

	txData = &ld.TxData{
		Type:      ld.TypeOpenLending,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      sender,
		Data:      []byte("你好👋"),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx(txData.ToTransaction(), true)
	assert.ErrorContains(err, "cbor: cannot unmarshal")

	input := &ld.LendingConfig{
		DailyInterest:   10,
		OverdueInterest: 1,
		MinAmount:       new(big.Int).SetUint64(constants.LDC),
		MaxAmount:       new(big.Int).SetUint64(constants.LDC),
	}
	assert.NoError(input.SyntacticVerify())
	txData = &ld.TxData{
		Type:      ld.TypeOpenLending,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      sender,
		Data:      ld.MustMarshal(input),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	tt := txData.ToTransaction()
	tt.Timestamp = bs.Timestamp()
	itx, err := NewTx(tt, true)
	assert.NoError(err)

	bs.CommitAccounts()
	assert.ErrorContains(itx.Apply(bctx, bs), "insufficient NativeLDC balance, expected 1794100, got 0")
	bs.CheckoutAccounts()

	senderAcc := bs.MustAccount(sender)
	senderAcc.Add(constants.NativeToken, new(big.Int).SetUint64(constants.LDC))
	assert.NoError(itx.Apply(bctx, bs))

	senderGas := tt.Gas()
	tx = itx.(*TxOpenLending)
	assert.Equal(senderGas*bctx.Price, tx.ldc.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(senderGas*100, tx.miner.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(constants.LDC-senderGas*(bctx.Price+100),
		senderAcc.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(uint64(1), senderAcc.Nonce())
	assert.NotNil(senderAcc.ld.Lending)
	assert.Equal(constants.NativeToken, senderAcc.ld.Lending.Token)
	assert.Equal(uint64(10), senderAcc.ld.Lending.DailyInterest)
	assert.Equal(uint64(1), senderAcc.ld.Lending.OverdueInterest)
	assert.Equal(constants.LDC, senderAcc.ld.Lending.MinAmount.Uint64())
	assert.Equal(constants.LDC, senderAcc.ld.Lending.MaxAmount.Uint64())
	assert.Equal(make(map[string]*ld.LendingEntry), senderAcc.ledger.Lending)

	jsondata, err := itx.MarshalJSON()
	assert.NoError(err)
	// fmt.Println(string(jsondata))
	assert.Equal(`{"type":"TypeOpenLending","chainID":2357,"nonce":0,"gasTip":100,"gasFeeCap":1000,"from":"0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC","data":{"token":"","dailyInterest":10,"overdueInterest":1,"minAmount":1000000000,"maxAmount":1000000000},"signatures":["659d9a2c6873ffe4f404702153e2bb96cf42434ec49af4788c7080aaadbc49e71d1d007610304a2a26d42345bbe287a3439abbf0b74185b35c999fc2b30b495800"],"id":"2kdoe4A18gqBWtYvEE9oJUh5AiWU4vkr9NfKznptTMeGDFpBJV"}`, string(jsondata))

	// openLending again
	input = &ld.LendingConfig{
		Token:           token,
		DailyInterest:   100,
		OverdueInterest: 10,
		MinAmount:       new(big.Int).SetUint64(constants.LDC),
		MaxAmount:       new(big.Int).SetUint64(constants.LDC * 10),
	}
	assert.NoError(input.SyntacticVerify())
	txData = &ld.TxData{
		Type:      ld.TypeOpenLending,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     1,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      sender,
		Data:      ld.MustMarshal(input),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	tt = txData.ToTransaction()
	tt.Timestamp = bs.Timestamp()
	itx, err = NewTx(tt, true)
	assert.NoError(err)
	bs.CommitAccounts()
	assert.ErrorContains(itx.Apply(bctx, bs),
		"Account(0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC).OpenLending error: lending exists")
	bs.CheckoutAccounts()

	assert.NoError(senderAcc.UpdateKeepers(nil, nil, &approver, ld.TxTypes{ld.TypeOpenLending}))
	// close lending
	txData = &ld.TxData{
		Type:      ld.TypeCloseLending,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     1,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      sender,
	}
	assert.NoError(txData.SignWith(util.Signer1))
	tt = txData.ToTransaction()
	tt.Timestamp = bs.Timestamp()
	itx, err = NewTx(tt, true)
	assert.NoError(err)
	assert.NoError(itx.Apply(bctx, bs))

	senderGas += tt.Gas()
	assert.Equal(senderGas*bctx.Price,
		itx.(*TxCloseLending).ldc.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(senderGas*100,
		itx.(*TxCloseLending).miner.balanceOf(constants.NativeToken).Uint64())
	assert.Nil(senderAcc.ld.Lending)
	assert.Equal(0, len(senderAcc.ledger.Lending))

	input = &ld.LendingConfig{
		Token:           token,
		DailyInterest:   100,
		OverdueInterest: 10,
		MinAmount:       new(big.Int).SetUint64(constants.LDC),
		MaxAmount:       new(big.Int).SetUint64(constants.LDC * 10),
	}
	assert.NoError(input.SyntacticVerify())
	txData = &ld.TxData{
		Type:      ld.TypeOpenLending,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     2,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      sender,
		Data:      ld.MustMarshal(input),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	tt = txData.ToTransaction()
	tt.Timestamp = bs.Timestamp()
	itx, err = NewTx(tt, true)
	assert.NoError(err)
	bs.CommitAccounts()
	assert.ErrorContains(itx.Apply(bctx, bs), "invalid signature for approver")
	bs.CheckoutAccounts()

	assert.NoError(txData.SignWith(util.Signer2))
	tt = txData.ToTransaction()
	tt.Timestamp = bs.Timestamp()
	itx, err = NewTx(tt, true)
	assert.NoError(err)
	assert.NoError(itx.Apply(bctx, bs))

	senderGas += tt.Gas()
	assert.Equal(senderGas*bctx.Price,
		itx.(*TxOpenLending).ldc.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(senderGas*100,
		itx.(*TxOpenLending).miner.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(uint64(3), senderAcc.Nonce())
	assert.NotNil(senderAcc.ld.Lending)
	assert.Equal(token, senderAcc.ld.Lending.Token)
	assert.Equal(uint64(100), senderAcc.ld.Lending.DailyInterest)
	assert.Equal(uint64(10), senderAcc.ld.Lending.OverdueInterest)
	assert.Equal(constants.LDC, senderAcc.ld.Lending.MinAmount.Uint64())
	assert.Equal(constants.LDC*10, senderAcc.ld.Lending.MaxAmount.Uint64())
	assert.Equal(make(map[string]*ld.LendingEntry), senderAcc.ledger.Lending)

	assert.NoError(bs.VerifyState())
}
