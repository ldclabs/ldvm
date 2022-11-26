// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package txn

import (
	"math/big"
	"testing"

	"github.com/ldclabs/ldvm/ids"
	"github.com/ldclabs/ldvm/ld"
	"github.com/ldclabs/ldvm/signer"
	"github.com/ldclabs/ldvm/unit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTxOpenLending(t *testing.T) {
	assert := assert.New(t)

	// SyntacticVerify
	tx := &TxOpenLending{}
	assert.ErrorContains(tx.SyntacticVerify(), "nil pointer")
	_, err := tx.MarshalJSON()
	require.NoError(t, err)

	ctx := NewMockChainContext()
	cs := ctx.MockChainState()
	token := ld.MustNewToken("$LDC")
	sender := signer.Signer1.Key().Address()

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
	assert.ErrorContains(err, "no signatures")

	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeOpenLending,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      sender,
		To:        ids.GenesisAccount.Ptr(),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
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
		Token:     ids.NativeToken.Ptr(),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
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
	assert.NoError(ltx.SignWith(signer.Signer1))
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
	assert.NoError(ltx.SignWith(signer.Signer1))
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
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "cbor: unexpected following extraneous data")

	input := &ld.LendingConfig{
		DailyInterest:   10,
		OverdueInterest: 1,
		MinAmount:       new(big.Int).SetUint64(unit.LDC),
		MaxAmount:       new(big.Int).SetUint64(unit.LDC),
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
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	ltx.Timestamp = cs.Timestamp()
	itx, err := NewTx(ltx)
	require.NoError(t, err)

	cs.CommitAccounts()
	assert.ErrorContains(itx.Apply(ctx, cs), "insufficient NativeLDC balance, expected 1816100, got 0")
	cs.CheckoutAccounts()

	senderAcc := cs.MustAccount(sender)
	senderAcc.Add(ids.NativeToken, new(big.Int).SetUint64(unit.LDC*2))
	assert.NoError(itx.Apply(ctx, cs))

	senderGas := ltx.Gas()
	tx = itx.(*TxOpenLending)
	assert.Equal(senderGas*ctx.Price, tx.ldc.Balance().Uint64())
	assert.Equal(senderGas*100, tx.miner.Balance().Uint64())
	assert.Equal(unit.LDC-senderGas*(ctx.Price+100),
		senderAcc.Balance().Uint64())
	assert.Equal(uint64(1), senderAcc.Nonce())
	require.NotNil(t, senderAcc.LD().Lending)
	assert.Equal(ids.NativeToken, senderAcc.LD().Lending.Token)
	assert.Equal(uint64(10), senderAcc.LD().Lending.DailyInterest)
	assert.Equal(uint64(1), senderAcc.LD().Lending.OverdueInterest)
	assert.Equal(unit.LDC, senderAcc.LD().Lending.MinAmount.Uint64())
	assert.Equal(unit.LDC, senderAcc.LD().Lending.MaxAmount.Uint64())
	assert.Equal(make(map[string]*ld.LendingEntry), senderAcc.Ledger().Lending)

	jsondata, err := itx.MarshalJSON()
	require.NoError(t, err)
	// fmt.Println(string(jsondata))
	assert.Equal(`{"tx":{"type":"TypeOpenLending","chainID":2357,"nonce":0,"gasTip":100,"gasFeeCap":1000,"from":"0x8db97c7cECe249C2b98bdc0226cc4C2A57bF52fc","data":{"token":"","dailyInterest":10,"overdueInterest":1,"minAmount":1000000000,"maxAmount":1000000000}},"sigs":["ZZ2aLGhz_-T0BHAhU-K7ls9CQ07EmvR4jHCAqq28SecdHQB2EDBKKibUI0W74oejQ5q78LdBhbNcmZ_CswtJWADJzBvv"],"id":"BomskC8OyfMUjpxBQSecj3OYmyTtobhE6ybc9e-1P_WGWo7z"}`, string(jsondata))

	// openLending again
	input = &ld.LendingConfig{
		Token:           token,
		DailyInterest:   100,
		OverdueInterest: 10,
		MinAmount:       new(big.Int).SetUint64(unit.LDC),
		MaxAmount:       new(big.Int).SetUint64(unit.LDC * 10),
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
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	ltx.Timestamp = cs.Timestamp()
	itx, err = NewTx(ltx)
	require.NoError(t, err)
	cs.CommitAccounts()
	assert.ErrorContains(itx.Apply(ctx, cs),
		"lending exists")
	cs.CheckoutAccounts()

	assert.NoError(senderAcc.UpdateKeepers(nil, nil, signer.Signer2.Key().Ptr(), &ld.TxTypes{ld.TypeOpenLending}))
	// close lending
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeCloseLending,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     1,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      sender,
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	ltx.Timestamp = cs.Timestamp()
	itx, err = NewTx(ltx)
	require.NoError(t, err)
	assert.NoError(itx.Apply(ctx, cs))

	senderGas += ltx.Gas()
	assert.Equal(senderGas*ctx.Price,
		itx.(*TxCloseLending).ldc.Balance().Uint64())
	assert.Equal(senderGas*100,
		itx.(*TxCloseLending).miner.Balance().Uint64())
	assert.Nil(senderAcc.LD().Lending)
	assert.Equal(0, len(senderAcc.Ledger().Lending))

	input = &ld.LendingConfig{
		Token:           token,
		DailyInterest:   100,
		OverdueInterest: 10,
		MinAmount:       new(big.Int).SetUint64(unit.LDC),
		MaxAmount:       new(big.Int).SetUint64(unit.LDC * 10),
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
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	ltx.Timestamp = cs.Timestamp()
	itx, err = NewTx(ltx)
	require.NoError(t, err)
	cs.CommitAccounts()
	assert.ErrorContains(itx.Apply(ctx, cs), "invalid signature for approver")
	cs.CheckoutAccounts()

	assert.NoError(ltx.SignWith(signer.Signer1, signer.Signer2))
	assert.NoError(ltx.SyntacticVerify())
	ltx.Timestamp = cs.Timestamp()
	itx, err = NewTx(ltx)
	require.NoError(t, err)
	assert.NoError(itx.Apply(ctx, cs))

	senderGas += ltx.Gas()
	assert.Equal(senderGas*ctx.Price,
		itx.(*TxOpenLending).ldc.Balance().Uint64())
	assert.Equal(senderGas*100,
		itx.(*TxOpenLending).miner.Balance().Uint64())
	assert.Equal(uint64(3), senderAcc.Nonce())
	require.NotNil(t, senderAcc.LD().Lending)
	assert.Equal(token, senderAcc.LD().Lending.Token)
	assert.Equal(uint64(100), senderAcc.LD().Lending.DailyInterest)
	assert.Equal(uint64(10), senderAcc.LD().Lending.OverdueInterest)
	assert.Equal(unit.LDC, senderAcc.LD().Lending.MinAmount.Uint64())
	assert.Equal(unit.LDC*10, senderAcc.LD().Lending.MaxAmount.Uint64())
	assert.Equal(make(map[string]*ld.LendingEntry), senderAcc.Ledger().Lending)

	assert.NoError(cs.VerifyState())
}
