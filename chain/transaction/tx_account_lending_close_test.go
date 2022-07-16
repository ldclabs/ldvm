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

func TestTxCloseLending(t *testing.T) {
	assert := assert.New(t)

	// SyntacticVerify
	tx := &TxCloseLending{}
	assert.ErrorContains(tx.SyntacticVerify(), "nil pointer")
	_, err := tx.MarshalJSON()
	assert.NoError(err)

	bctx := NewMockBCtx()
	bs := bctx.MockBS()
	token := ld.MustNewToken("$LDC")
	sender := util.Signer1.Address()

	txData := &ld.TxData{
		Type:      ld.TypeCloseLending,
		ChainID:   bctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      sender,
	}
	assert.NoError(txData.SyntacticVerify())
	_, err = NewTx(txData.ToTransaction(), true)
	assert.ErrorContains(err, "DeriveSigners error: no signature")

	txData = &ld.TxData{
		Type:      ld.TypeCloseLending,
		ChainID:   bctx.ChainConfig().ChainID,
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
		Type:      ld.TypeCloseLending,
		ChainID:   bctx.ChainConfig().ChainID,
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
		Type:      ld.TypeCloseLending,
		ChainID:   bctx.ChainConfig().ChainID,
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
		Type:      ld.TypeCloseLending,
		ChainID:   bctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      sender,
	}
	assert.NoError(txData.SignWith(util.Signer1))
	tt := txData.ToTransaction()
	tt.Timestamp = bs.Timestamp()
	itx, err := NewTx(tt, true)
	assert.NoError(err)

	bs.CommitAccounts()
	assert.ErrorContains(itx.Apply(bctx, bs),
		"insufficient NativeLDC balance, expected 1579600, got 0")
	bs.CheckoutAccounts()

	senderAcc := bs.MustAccount(sender)
	senderAcc.Add(constants.NativeToken, new(big.Int).SetUint64(constants.LDC))
	bs.CommitAccounts()
	assert.ErrorContains(itx.Apply(bctx, bs),
		"Account(0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC).CloseLending error: invalid lending")
	bs.CheckoutAccounts()

	input := &ld.LendingConfig{
		Token:           token,
		DailyInterest:   100,
		OverdueInterest: 10,
		MinAmount:       new(big.Int).SetUint64(constants.LDC),
		MaxAmount:       new(big.Int).SetUint64(constants.LDC * 10),
	}
	assert.NoError(input.SyntacticVerify())
	txData = &ld.TxData{
		Type:      ld.TypeOpenLending,
		ChainID:   bctx.ChainConfig().ChainID,
		Nonce:     0,
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
	assert.NoError(itx.Apply(bctx, bs))

	senderGas := tt.Gas()
	assert.Equal(uint64(1), senderAcc.Nonce())
	assert.NotNil(senderAcc.ld.Lending)
	assert.Equal(token, senderAcc.ld.Lending.Token)
	assert.Equal(make(map[string]*ld.LendingEntry), senderAcc.ledger.Lending)

	txData = &ld.TxData{
		Type:      ld.TypeCloseLending,
		ChainID:   bctx.ChainConfig().ChainID,
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
	assert.Equal(constants.LDC-senderGas*(bctx.Price+100),
		senderAcc.balanceOf(constants.NativeToken).Uint64())
	assert.Nil(senderAcc.ld.Lending)
	assert.Equal(0, len(senderAcc.ledger.Lending))

	jsondata, err := itx.MarshalJSON()
	assert.NoError(err)
	// fmt.Println(string(jsondata))
	assert.Equal(`{"type":"TypeCloseLending","chainID":2357,"nonce":1,"gasTip":100,"gasFeeCap":1000,"from":"0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC","signatures":["ceb99eaf8a7825a8c9d26eeeea7fe3672eede4e0a4d0c18cb50638c00fca73831d50efafc6533b29258b3938811d5fe02a15688911f9958b121c7e14f6e01e8500"],"id":"242SMs7S5niL4YYNsaNrQgASCKh4Brcnep9sjhxv7toj8R1PNB"}`, string(jsondata))

	txData = &ld.TxData{
		Type:      ld.TypeCloseLending,
		ChainID:   bctx.ChainConfig().ChainID,
		Nonce:     2,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      sender,
	}
	assert.NoError(txData.SignWith(util.Signer1))
	tt = txData.ToTransaction()
	tt.Timestamp = bs.Timestamp()
	itx, err = NewTx(tt, true)
	assert.NoError(err)
	assert.ErrorContains(itx.Apply(bctx, bs),
		"Account(0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC).CloseLending error: invalid lending")

	assert.NoError(bs.VerifyState())
}
