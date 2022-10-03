// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package transactions

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

	ctx := NewMockChainContext()
	cs := ctx.MockChainState()
	token := ld.MustNewToken("$LDC")
	sender := util.Signer1.Address()
	approver := util.Signer2.Address()

	ltx := &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeOpenLending,
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
		Type:      ld.TypeOpenLending,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      sender,
		To:        &constants.GenesisAccount,
	}}
	assert.NoError(ltx.SignWith(util.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "invalid to, should be nil")

	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeOpenLending,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      sender,
		Token:     &constants.NativeToken,
	}}
	assert.NoError(ltx.SignWith(util.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "invalid token, should be nil")

	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeOpenLending,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      sender,
		Amount:    big.NewInt(1),
	}}
	assert.NoError(ltx.SignWith(util.Signer1))
	_, err = NewTx(ltx)
	assert.ErrorContains(ltx.SyntacticVerify(), "nil \"to\" together with amount")

	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeOpenLending,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      sender,
	}}
	assert.NoError(ltx.SignWith(util.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "invalid data")

	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeOpenLending,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      sender,
		Data:      []byte("你好👋"),
	}}
	assert.NoError(ltx.SignWith(util.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "cbor: unexpected following extraneous data")

	input := &ld.LendingConfig{
		DailyInterest:   10,
		OverdueInterest: 1,
		MinAmount:       new(big.Int).SetUint64(constants.LDC),
		MaxAmount:       new(big.Int).SetUint64(constants.LDC),
	}
	assert.NoError(input.SyntacticVerify())
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeOpenLending,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      sender,
		Data:      ld.MustMarshal(input),
	}}
	assert.NoError(ltx.SignWith(util.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	ltx.Timestamp = cs.Timestamp()
	itx, err := NewTx(ltx)
	assert.NoError(err)

	cs.CommitAccounts()
	assert.ErrorContains(itx.Apply(ctx, cs), "insufficient NativeLDC balance, expected 1816100, got 0")
	cs.CheckoutAccounts()

	senderAcc := cs.MustAccount(sender)
	senderAcc.Add(constants.NativeToken, new(big.Int).SetUint64(constants.LDC))
	assert.NoError(itx.Apply(ctx, cs))

	senderGas := ltx.Gas()
	tx = itx.(*TxOpenLending)
	assert.Equal(senderGas*ctx.Price, tx.ldc.Balance().Uint64())
	assert.Equal(senderGas*100, tx.miner.Balance().Uint64())
	assert.Equal(constants.LDC-senderGas*(ctx.Price+100),
		senderAcc.Balance().Uint64())
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
	assert.Equal(`{"tx":{"type":"TypeOpenLending","chainID":2357,"nonce":0,"gasTip":100,"gasFeeCap":1000,"from":"0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC","data":{"token":"","dailyInterest":10,"overdueInterest":1,"minAmount":1000000000,"maxAmount":1000000000}},"sigs":["659d9a2c6873ffe4f404702153e2bb96cf42434ec49af4788c7080aaadbc49e71d1d007610304a2a26d42345bbe287a3439abbf0b74185b35c999fc2b30b495800"],"id":"3szwpWhba7z3YCc4uNcmh5HCEvwfu2i8yvSC48sm6GpZEgFQt"}`, string(jsondata))

	// openLending again
	input = &ld.LendingConfig{
		Token:           token,
		DailyInterest:   100,
		OverdueInterest: 10,
		MinAmount:       new(big.Int).SetUint64(constants.LDC),
		MaxAmount:       new(big.Int).SetUint64(constants.LDC * 10),
	}
	assert.NoError(input.SyntacticVerify())
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeOpenLending,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     1,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      sender,
		Data:      ld.MustMarshal(input),
	}}
	assert.NoError(ltx.SignWith(util.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	ltx.Timestamp = cs.Timestamp()
	itx, err = NewTx(ltx)
	assert.NoError(err)
	cs.CommitAccounts()
	assert.ErrorContains(itx.Apply(ctx, cs),
		"Account(0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC).OpenLending error: lending exists")
	cs.CheckoutAccounts()

	assert.NoError(senderAcc.UpdateKeepers(nil, nil, &approver, ld.TxTypes{ld.TypeOpenLending}))
	// close lending
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeCloseLending,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     1,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      sender,
	}}
	assert.NoError(ltx.SignWith(util.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	ltx.Timestamp = cs.Timestamp()
	itx, err = NewTx(ltx)
	assert.NoError(err)
	assert.NoError(itx.Apply(ctx, cs))

	senderGas += ltx.Gas()
	assert.Equal(senderGas*ctx.Price,
		itx.(*TxCloseLending).ldc.Balance().Uint64())
	assert.Equal(senderGas*100,
		itx.(*TxCloseLending).miner.Balance().Uint64())
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
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeOpenLending,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     2,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      sender,
		Data:      ld.MustMarshal(input),
	}}
	assert.NoError(ltx.SignWith(util.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	ltx.Timestamp = cs.Timestamp()
	itx, err = NewTx(ltx)
	assert.NoError(err)
	cs.CommitAccounts()
	assert.ErrorContains(itx.Apply(ctx, cs), "invalid signature for approver")
	cs.CheckoutAccounts()

	assert.NoError(ltx.SignWith(util.Signer1, util.Signer2))
	assert.NoError(ltx.SyntacticVerify())
	ltx.Timestamp = cs.Timestamp()
	itx, err = NewTx(ltx)
	assert.NoError(err)
	assert.NoError(itx.Apply(ctx, cs))

	senderGas += ltx.Gas()
	assert.Equal(senderGas*ctx.Price,
		itx.(*TxOpenLending).ldc.Balance().Uint64())
	assert.Equal(senderGas*100,
		itx.(*TxOpenLending).miner.Balance().Uint64())
	assert.Equal(uint64(3), senderAcc.Nonce())
	assert.NotNil(senderAcc.ld.Lending)
	assert.Equal(token, senderAcc.ld.Lending.Token)
	assert.Equal(uint64(100), senderAcc.ld.Lending.DailyInterest)
	assert.Equal(uint64(10), senderAcc.ld.Lending.OverdueInterest)
	assert.Equal(constants.LDC, senderAcc.ld.Lending.MinAmount.Uint64())
	assert.Equal(constants.LDC*10, senderAcc.ld.Lending.MaxAmount.Uint64())
	assert.Equal(make(map[string]*ld.LendingEntry), senderAcc.ledger.Lending)

	assert.NoError(cs.VerifyState())
}
