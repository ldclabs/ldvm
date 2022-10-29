// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package transactions

import (
	"math/big"
	"testing"

	"github.com/ldclabs/ldvm/constants"
	"github.com/ldclabs/ldvm/ld"
	"github.com/ldclabs/ldvm/util/signer"
	"github.com/stretchr/testify/assert"
)

func TestTxRepay(t *testing.T) {
	assert := assert.New(t)

	// SyntacticVerify
	tx := &TxRepay{}
	assert.ErrorContains(tx.SyntacticVerify(), "nil pointer")
	_, err := tx.MarshalJSON()
	assert.NoError(err)

	ctx := NewMockChainContext()
	cs := ctx.MockChainState()
	token := ld.MustNewToken("$LDC")
	borrower := signer.Signer1.Key().Address()
	lender := signer.Signer2.Key().Address()

	ltx := &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeRepay,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      borrower,
	}}
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "no signatures")

	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "nil to as lender")

	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeRepay,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      borrower,
		To:        &lender,
		Amount:    new(big.Int).SetUint64(0),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "invalid amount, expected > 0, got 0")

	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeRepay,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      borrower,
		To:        &lender,
		Amount:    new(big.Int).SetUint64(constants.LDC),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	itx, err := NewTx(ltx)
	assert.NoError(err)

	cs.CommitAccounts()
	assert.ErrorContains(itx.Apply(ctx, cs),
		"insufficient NativeLDC balance, expected 1001222100, got 0")
	cs.CheckoutAccounts()

	borrowerAcc := cs.MustAccount(borrower)
	borrowerAcc.Add(constants.NativeToken, new(big.Int).SetUint64(constants.LDC*2))

	cs.CommitAccounts()
	assert.ErrorContains(itx.Apply(ctx, cs),
		"Account(0x44171C37Ff5D7B7bb8Dcad5C81f16284A229E641).Repay: invalid lending")
	cs.CheckoutAccounts()

	// open lending
	lcfg := &ld.LendingConfig{
		Token:           token,
		DailyInterest:   10_000,
		OverdueInterest: 10_000,
		MinAmount:       new(big.Int).SetUint64(constants.LDC),
		MaxAmount:       new(big.Int).SetUint64(constants.LDC * 10),
	}
	assert.NoError(lcfg.SyntacticVerify())
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeOpenLending,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      lender,
		Data:      ld.MustMarshal(lcfg),
	}}
	assert.NoError(ltx.SignWith(signer.Signer2))
	assert.NoError(ltx.SyntacticVerify())
	ltx.Timestamp = cs.Timestamp()
	itx, err = NewTx(ltx)
	assert.NoError(err)

	lenderAcc := cs.MustAccount(lender)
	lenderAcc.Add(constants.NativeToken, new(big.Int).SetUint64(constants.LDC))
	assert.NoError(itx.Apply(ctx, cs))

	lenderGas := ltx.Gas()
	tx2 := itx.(*TxOpenLending)
	assert.Equal(lenderGas*ctx.Price, tx2.ldc.Balance().Uint64())
	assert.Equal(lenderGas*100, tx2.miner.Balance().Uint64())
	assert.Equal(constants.LDC-lenderGas*(ctx.Price+100),
		lenderAcc.Balance().Uint64())
	assert.NotNil(lenderAcc.ld.Lending)
	assert.NotNil(lenderAcc.ledger)
	assert.Equal(0, len(lenderAcc.ledger.Lending))

	// repay again
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeRepay,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      borrower,
		To:        &lender,
		Amount:    new(big.Int).SetUint64(constants.LDC),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	ltx.Timestamp = cs.Timestamp()
	itx, err = NewTx(ltx)
	assert.NoError(err)

	cs.CommitAccounts()
	assert.ErrorContains(itx.Apply(ctx, cs),
		"Account(0x44171C37Ff5D7B7bb8Dcad5C81f16284A229E641).Repay: invalid token, expected $LDC, got NativeLDC")
	cs.CheckoutAccounts()

	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeRepay,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      borrower,
		To:        &lender,
		Token:     &token,
		Amount:    new(big.Int).SetUint64(constants.LDC),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	ltx.Timestamp = cs.Timestamp()
	itx, err = NewTx(ltx)
	assert.NoError(err)

	cs.CommitAccounts()
	assert.ErrorContains(itx.Apply(ctx, cs),
		"insufficient $LDC balance, expected 1000000000, got 0")
	cs.CheckoutAccounts()

	borrowerAcc.Add(token, new(big.Int).SetUint64(constants.LDC))
	cs.CommitAccounts()
	assert.ErrorContains(itx.Apply(ctx, cs),
		"TxRepay.Apply: Account(0x44171C37Ff5D7B7bb8Dcad5C81f16284A229E641).Repay: don't need to repay")
	cs.CheckoutAccounts()

	// borrow
	input := &ld.TxTransfer{
		Nonce:  1,
		From:   &lender,
		To:     &borrower,
		Token:  &token,
		Amount: new(big.Int).SetUint64(constants.LDC),
		Expire: cs.Timestamp() + 1,
	}
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeBorrow,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      borrower,
		To:        &lender,
		Token:     &token,
		Data:      input.Bytes(),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.ExSignWith(signer.Signer2))
	assert.NoError(ltx.SyntacticVerify())
	ltx.Timestamp = cs.Timestamp()
	itx, err = NewTx(ltx)
	assert.NoError(err)

	cs.CommitAccounts()
	assert.ErrorContains(itx.Apply(ctx, cs),
		"TxBorrow.Apply: Account(0x44171C37Ff5D7B7bb8Dcad5C81f16284A229E641).Borrow: insufficient $LDC balance, expected 1000000000, got 0")
	cs.CheckoutAccounts()

	assert.NoError(lenderAcc.Add(token, new(big.Int).SetUint64(constants.LDC)))
	assert.NoError(lenderAcc.UpdateNonceTable(cs.Timestamp()+1, []uint64{1}))
	assert.NoError(itx.Apply(ctx, cs))

	borrowerGas := ltx.Gas()
	tx3 := itx.(*TxBorrow)
	assert.Equal((lenderGas+borrowerGas)*ctx.Price, tx3.ldc.Balance().Uint64())
	assert.Equal((lenderGas+borrowerGas)*100, tx3.miner.Balance().Uint64())
	assert.Equal(constants.LDC*2-borrowerGas*(ctx.Price+100),
		borrowerAcc.Balance().Uint64())
	assert.Equal(constants.LDC-lenderGas*(ctx.Price+100),
		lenderAcc.Balance().Uint64())
	assert.Equal(uint64(0), lenderAcc.balanceOf(token).Uint64())
	assert.Equal(constants.LDC*2, borrowerAcc.balanceOf(token).Uint64())

	assert.Equal(1, len(lenderAcc.ledger.Lending))
	assert.Equal(0, len(lenderAcc.ld.NonceTable))
	assert.NotNil(lenderAcc.ledger.Lending[borrower.AsKey()])
	entry := lenderAcc.ledger.Lending[borrower.AsKey()]
	assert.Equal(constants.LDC, entry.Amount.Uint64())
	assert.Equal(cs.Timestamp(), entry.UpdateAt)
	assert.Equal(uint64(0), entry.DueTime)

	// repay after a day
	cs.CommitAccounts()
	ctx.height++
	ctx.timestamp += 3600 * 24
	cs.CheckoutAccounts()

	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeRepay,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     1,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      borrower,
		To:        &lender,
		Token:     &token,
		Amount:    new(big.Int).SetUint64(constants.LDC),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	ltx.Timestamp = cs.Timestamp()
	itx, err = NewTx(ltx)
	assert.NoError(err)
	assert.NoError(itx.Apply(ctx, cs))

	borrowerGas += ltx.Gas()
	assert.Equal((lenderGas+borrowerGas)*ctx.Price,
		itx.(*TxRepay).ldc.Balance().Uint64())
	assert.Equal((lenderGas+borrowerGas)*100,
		itx.(*TxRepay).miner.Balance().Uint64())
	assert.Equal(constants.LDC*2-borrowerGas*(ctx.Price+100),
		borrowerAcc.Balance().Uint64())
	assert.Equal(constants.LDC-lenderGas*(ctx.Price+100),
		lenderAcc.Balance().Uint64())
	assert.Equal(constants.LDC, lenderAcc.balanceOf(token).Uint64())
	assert.Equal(constants.LDC, borrowerAcc.balanceOf(token).Uint64())

	assert.NotNil(lenderAcc.ledger.Lending[borrower.AsKey()])
	entry = lenderAcc.ledger.Lending[borrower.AsKey()]

	interest := uint64(float64(constants.LDC) * 10_000 / 1_000_000)
	assert.Equal(interest, entry.Amount.Uint64(), "with 1 day interest")
	assert.Equal(cs.Timestamp(), entry.UpdateAt)
	assert.Equal(uint64(0), entry.DueTime)

	jsondata, err := itx.MarshalJSON()
	assert.NoError(err)
	// fmt.Println(string(jsondata))
	assert.Equal(`{"tx":{"type":"TypeRepay","chainID":2357,"nonce":1,"gasTip":100,"gasFeeCap":1000,"from":"0x8db97c7cECe249C2b98bdc0226cc4C2A57bF52fc","to":"0x44171C37Ff5D7B7bb8Dcad5C81f16284A229E641","token":"$LDC","amount":1000000000},"sigs":["DJpQlyTwsgCRYOsGgGVxeKZhG5di0e028DPM4YSGV8ZK-4iLRAxTDGVQ1o8_Dzgsz-ZsY0ajb4yjfs_uIs5_pwBYjYXp"],"id":"VkUgbcrSqCsYIjk14O90UqOr95Jy0tfVFaZzJg5wH-R29AwO"}`, string(jsondata))

	// repay again
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeRepay,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     2,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      borrower,
		To:        &lender,
		Token:     &token,
		Amount:    new(big.Int).SetUint64(constants.MilliLDC * 20),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	ltx.Timestamp = cs.Timestamp()
	itx, err = NewTx(ltx)
	assert.NoError(err)
	assert.NoError(itx.Apply(ctx, cs))

	borrowerGas += ltx.Gas()
	assert.Equal((lenderGas+borrowerGas)*ctx.Price,
		itx.(*TxRepay).ldc.Balance().Uint64())
	assert.Equal((lenderGas+borrowerGas)*100,
		itx.(*TxRepay).miner.Balance().Uint64())
	assert.Equal(constants.LDC*2-borrowerGas*(ctx.Price+100),
		borrowerAcc.Balance().Uint64())
	assert.Equal(constants.LDC-lenderGas*(ctx.Price+100),
		lenderAcc.Balance().Uint64())
	assert.Equal(constants.LDC+interest, lenderAcc.balanceOf(token).Uint64())
	assert.Equal(constants.LDC-interest, borrowerAcc.balanceOf(token).Uint64())
	assert.Nil(lenderAcc.ledger.Lending[borrower.AsKey()], "clear entry when repay all")

	assert.NoError(cs.VerifyState())
}
