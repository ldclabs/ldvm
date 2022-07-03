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

func TestTxRepay(t *testing.T) {
	assert := assert.New(t)

	// SyntacticVerify
	tx := &TxRepay{}
	assert.ErrorContains(tx.SyntacticVerify(), "nil pointer")
	_, err := tx.MarshalJSON()
	assert.NoError(err)

	bctx := NewMockBCtx()
	bs := bctx.MockBS()
	token := ld.MustNewToken("$LDC")
	borrower := util.Signer1.Address()
	lender := util.Signer2.Address()

	txData := &ld.TxData{
		Type:      ld.TypeRepay,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      borrower,
	}
	assert.NoError(txData.SyntacticVerify())
	_, err = NewTx(txData.ToTransaction(), true)
	assert.ErrorContains(err, "DeriveSigners error: no signature")

	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx(txData.ToTransaction(), true)
	assert.ErrorContains(err, "nil to as lender")

	txData = &ld.TxData{
		Type:      ld.TypeRepay,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      borrower,
		To:        &lender,
		Amount:    new(big.Int).SetUint64(0),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx(txData.ToTransaction(), true)
	assert.ErrorContains(err, "invalid amount, expected > 0, got 0")

	txData = &ld.TxData{
		Type:      ld.TypeRepay,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      borrower,
		To:        &lender,
		Amount:    new(big.Int).SetUint64(constants.LDC),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	tt := txData.ToTransaction()
	itx, err := NewTx(tt, true)
	assert.NoError(err)

	bs.CommitAccounts()
	assert.ErrorContains(itx.Apply(bctx, bs),
		"insufficient NativeLDC balance, expected 1001200100, got 0")
	bs.CheckoutAccounts()

	borrowerAcc := bs.MustAccount(borrower)
	borrowerAcc.Add(constants.NativeToken, new(big.Int).SetUint64(constants.LDC*2))

	bs.CommitAccounts()
	assert.ErrorContains(itx.Apply(bctx, bs),
		"Account(0x44171C37Ff5D7B7bb8dcad5C81f16284A229e641).Repay error: invalid lending")
	bs.CheckoutAccounts()

	// open lending
	lcfg := &ld.LendingConfig{
		Token:           token,
		DailyInterest:   10_000,
		OverdueInterest: 10_000,
		MinAmount:       new(big.Int).SetUint64(constants.LDC),
		MaxAmount:       new(big.Int).SetUint64(constants.LDC * 10),
	}
	assert.NoError(lcfg.SyntacticVerify())
	txData = &ld.TxData{
		Type:      ld.TypeOpenLending,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      lender,
		Data:      ld.MustMarshal(lcfg),
	}
	assert.NoError(txData.SignWith(util.Signer2))
	tt = txData.ToTransaction()
	tt.Timestamp = bs.Timestamp()
	itx, err = NewTx(tt, true)
	assert.NoError(err)

	lenderAcc := bs.MustAccount(lender)
	lenderAcc.Add(constants.NativeToken, new(big.Int).SetUint64(constants.LDC))
	assert.NoError(itx.Apply(bctx, bs))

	lenderGas := tt.Gas()
	tx2 := itx.(*TxOpenLending)
	assert.Equal(lenderGas*bctx.Price, tx2.ldc.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(lenderGas*100, tx2.miner.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(constants.LDC-lenderGas*(bctx.Price+100),
		lenderAcc.balanceOf(constants.NativeToken).Uint64())
	assert.NotNil(lenderAcc.ld.Lending)
	assert.NotNil(lenderAcc.ledger)
	assert.Equal(0, len(lenderAcc.ledger.Lending))

	// repay again
	txData = &ld.TxData{
		Type:      ld.TypeRepay,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      borrower,
		To:        &lender,
		Amount:    new(big.Int).SetUint64(constants.LDC),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	tt = txData.ToTransaction()
	tt.Timestamp = bs.Timestamp()
	itx, err = NewTx(tt, true)
	assert.NoError(err)

	bs.CommitAccounts()
	assert.ErrorContains(itx.Apply(bctx, bs),
		"Account(0x44171C37Ff5D7B7bb8dcad5C81f16284A229e641).Repay error: invalid token, expected $LDC, got NativeLDC")
	bs.CheckoutAccounts()

	txData = &ld.TxData{
		Type:      ld.TypeRepay,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      borrower,
		To:        &lender,
		Token:     &token,
		Amount:    new(big.Int).SetUint64(constants.LDC),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	tt = txData.ToTransaction()
	tt.Timestamp = bs.Timestamp()
	itx, err = NewTx(tt, true)
	assert.NoError(err)

	bs.CommitAccounts()
	assert.ErrorContains(itx.Apply(bctx, bs),
		"insufficient $LDC balance, expected 1000000000, got 0")
	bs.CheckoutAccounts()

	borrowerAcc.Add(token, new(big.Int).SetUint64(constants.LDC))
	bs.CommitAccounts()
	assert.ErrorContains(itx.Apply(bctx, bs),
		"TxRepay.Apply error: Account(0x44171C37Ff5D7B7bb8dcad5C81f16284A229e641).Repay error: don't need to repay")
	bs.CheckoutAccounts()

	// borrow
	input := &ld.TxTransfer{
		Nonce:  1,
		From:   &lender,
		To:     &borrower,
		Token:  &token,
		Amount: new(big.Int).SetUint64(constants.LDC),
		Expire: bs.Timestamp() + 1,
	}
	txData = &ld.TxData{
		Type:      ld.TypeBorrow,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      borrower,
		To:        &lender,
		Token:     &token,
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	assert.NoError(txData.ExSignWith(util.Signer2))
	tt = txData.ToTransaction()
	tt.Timestamp = bs.Timestamp()
	itx, err = NewTx(tt, true)
	assert.NoError(err)

	bs.CommitAccounts()
	assert.ErrorContains(itx.Apply(bctx, bs),
		"TxBorrow.Apply error: Account(0x44171C37Ff5D7B7bb8dcad5C81f16284A229e641).Borrow error: insufficient $LDC balance, expected 1000000000, got 0")
	bs.CheckoutAccounts()

	assert.NoError(lenderAcc.Add(token, new(big.Int).SetUint64(constants.LDC)))
	assert.NoError(lenderAcc.AddNonceTable(bs.Timestamp()+1, []uint64{1}))
	assert.NoError(itx.Apply(bctx, bs))

	borrowerGas := tt.Gas()
	tx3 := itx.(*TxBorrow)
	assert.Equal((lenderGas+borrowerGas)*bctx.Price, tx3.ldc.balanceOf(constants.NativeToken).Uint64())
	assert.Equal((lenderGas+borrowerGas)*100, tx3.miner.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(constants.LDC*2-borrowerGas*(bctx.Price+100),
		borrowerAcc.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(constants.LDC-lenderGas*(bctx.Price+100),
		lenderAcc.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(uint64(0), lenderAcc.balanceOf(token).Uint64())
	assert.Equal(constants.LDC*2, borrowerAcc.balanceOf(token).Uint64())

	assert.Equal(1, len(lenderAcc.ledger.Lending))
	assert.Equal(0, len(lenderAcc.ld.NonceTable))
	assert.NotNil(lenderAcc.ledger.Lending[borrower.AsKey()])
	entry := lenderAcc.ledger.Lending[borrower.AsKey()]
	assert.Equal(constants.LDC, entry.Amount.Uint64())
	assert.Equal(bs.Timestamp(), entry.UpdateAt)
	assert.Equal(uint64(0), entry.DueTime)

	// repay after a day
	bs.CommitAccounts()
	bctx.height++
	bctx.timestamp += 3600 * 24
	bs.CheckoutAccounts()

	txData = &ld.TxData{
		Type:      ld.TypeRepay,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     1,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      borrower,
		To:        &lender,
		Token:     &token,
		Amount:    new(big.Int).SetUint64(constants.LDC),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	tt = txData.ToTransaction()
	tt.Timestamp = bs.Timestamp()
	itx, err = NewTx(tt, true)
	assert.NoError(err)
	assert.NoError(itx.Apply(bctx, bs))

	borrowerGas += tt.Gas()
	assert.Equal((lenderGas+borrowerGas)*bctx.Price,
		itx.(*TxRepay).ldc.balanceOf(constants.NativeToken).Uint64())
	assert.Equal((lenderGas+borrowerGas)*100,
		itx.(*TxRepay).miner.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(constants.LDC*2-borrowerGas*(bctx.Price+100),
		borrowerAcc.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(constants.LDC-lenderGas*(bctx.Price+100),
		lenderAcc.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(constants.LDC, lenderAcc.balanceOf(token).Uint64())
	assert.Equal(constants.LDC, borrowerAcc.balanceOf(token).Uint64())

	assert.NotNil(lenderAcc.ledger.Lending[borrower.AsKey()])
	entry = lenderAcc.ledger.Lending[borrower.AsKey()]

	interest := uint64(float64(constants.LDC) * 10_000 / 1_000_000)
	assert.Equal(interest, entry.Amount.Uint64(), "with 1 day interest")
	assert.Equal(bs.Timestamp(), entry.UpdateAt)
	assert.Equal(uint64(0), entry.DueTime)

	jsondata, err := itx.MarshalJSON()
	assert.NoError(err)
	// fmt.Println(string(jsondata))
	assert.Equal(`{"type":"TypeRepay","chainID":2357,"nonce":1,"gasTip":100,"gasFeeCap":1000,"from":"0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC","to":"0x44171C37Ff5D7B7bb8dcad5C81f16284A229e641","token":"$LDC","amount":1000000000,"signatures":["0c9a509724f0b2009160eb0680657178a6611b9762d1ed36f033cce1848657c64afb888b440c530c6550d68f3f0f382ccfe66c6346a36f8ca37ecfee22ce7fa700"],"id":"LkWrkmjvWd6HUMGEhdn7jXqxLS5feFTPVQE9uRuzrAx4nogQW"}`, string(jsondata))

	// repay again
	txData = &ld.TxData{
		Type:      ld.TypeRepay,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     2,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      borrower,
		To:        &lender,
		Token:     &token,
		Amount:    new(big.Int).SetUint64(constants.MilliLDC * 20),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	tt = txData.ToTransaction()
	tt.Timestamp = bs.Timestamp()
	itx, err = NewTx(tt, true)
	assert.NoError(err)
	assert.NoError(itx.Apply(bctx, bs))

	borrowerGas += tt.Gas()
	assert.Equal((lenderGas+borrowerGas)*bctx.Price,
		itx.(*TxRepay).ldc.balanceOf(constants.NativeToken).Uint64())
	assert.Equal((lenderGas+borrowerGas)*100,
		itx.(*TxRepay).miner.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(constants.LDC*2-borrowerGas*(bctx.Price+100),
		borrowerAcc.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(constants.LDC-lenderGas*(bctx.Price+100),
		lenderAcc.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(constants.LDC+interest, lenderAcc.balanceOf(token).Uint64())
	assert.Equal(constants.LDC-interest, borrowerAcc.balanceOf(token).Uint64())
	assert.Nil(lenderAcc.ledger.Lending[borrower.AsKey()], "clear entry when repay all")

	assert.NoError(bs.VerifyState())
}
