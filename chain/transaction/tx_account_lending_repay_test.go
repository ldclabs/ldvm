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
	bs := NewMockBS(bctx)
	token := ld.MustNewToken("$LDC")

	ldc, err := bs.LoadAccount(constants.LDCAccount)
	assert.NoError(err)
	miner, err := bs.LoadMiner(bctx.Miner())
	assert.NoError(err)
	borrower, err := bs.LoadAccount(util.Signer1.Address())
	assert.NoError(err)
	lender, err := bs.LoadAccount(util.Signer2.Address())
	assert.NoError(err)

	txData := &ld.TxData{
		Type:      ld.TypeRepay,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      borrower.id,
	}
	assert.NoError(txData.SyntacticVerify())
	_, err = NewTx(txData.ToTransaction(), true)
	assert.ErrorContains(err, "DeriveSigners: no signature")

	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx(txData.ToTransaction(), true)
	assert.ErrorContains(err, "nil to as lender")

	txData = &ld.TxData{
		Type:      ld.TypeRepay,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      borrower.id,
		To:        &lender.id,
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
		From:      borrower.id,
		To:        &lender.id,
		Amount:    new(big.Int).SetUint64(constants.LDC),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	tt := txData.ToTransaction()
	itx, err := NewTx(tt, true)
	assert.NoError(err)
	assert.ErrorContains(itx.Verify(bctx, bs), "invalid gas, expected 580, got 0")

	tt.Gas = tt.RequiredGas(bctx.FeeConfig().ThresholdGas)
	itx, err = NewTx(tt, true)
	assert.NoError(err)
	assert.ErrorContains(itx.Verify(bctx, bs),
		"insufficient NativeLDC balance, expected 1000638000, got 0")
	borrower.Add(constants.NativeToken, new(big.Int).SetUint64(constants.LDC*2))
	assert.ErrorContains(itx.Verify(bctx, bs), "CheckRepay failed: invalid lending")

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
		From:      lender.id,
		Data:      ld.MustMarshal(lcfg),
	}
	assert.NoError(txData.SignWith(util.Signer2))
	tt = txData.ToTransaction()
	tt.Timestamp = bs.Timestamp()
	tt.Gas = tt.RequiredGas(bctx.FeeConfig().ThresholdGas)
	lenderGas := tt.Gas
	itx, err = NewTx(tt, true)
	assert.NoError(err)
	lender.Add(constants.NativeToken, new(big.Int).SetUint64(constants.LDC))
	assert.NoError(itx.Verify(bctx, bs))
	assert.NoError(itx.Accept(bctx, bs))

	assert.Equal(lenderGas*bctx.Price, ldc.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(lenderGas*100, miner.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(constants.LDC-lenderGas*(bctx.Price+100),
		lender.balanceOf(constants.NativeToken).Uint64())
	assert.NotNil(lender.ld.Lending)
	assert.NotNil(lender.ld.LendingLedger)
	assert.Equal(0, len(lender.ld.LendingLedger))

	// repay again
	txData = &ld.TxData{
		Type:      ld.TypeRepay,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      borrower.id,
		To:        &lender.id,
		Amount:    new(big.Int).SetUint64(constants.LDC),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	tt = txData.ToTransaction()
	tt.Timestamp = bs.Timestamp()
	tt.Gas = tt.RequiredGas(bctx.FeeConfig().ThresholdGas)
	itx, err = NewTx(tt, true)
	assert.NoError(err)
	assert.ErrorContains(itx.Verify(bctx, bs),
		"CheckRepay failed: invalid token, expected $LDC, got NativeLDC")

	txData = &ld.TxData{
		Type:      ld.TypeRepay,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      borrower.id,
		To:        &lender.id,
		Token:     &token,
		Amount:    new(big.Int).SetUint64(constants.LDC),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	tt = txData.ToTransaction()
	tt.Timestamp = bs.Timestamp()
	tt.Gas = tt.RequiredGas(bctx.FeeConfig().ThresholdGas)
	itx, err = NewTx(tt, true)
	assert.NoError(err)
	assert.ErrorContains(itx.Verify(bctx, bs),
		"insufficient $LDC balance, expected 1000000000, got 0")

	borrower.Add(token, new(big.Int).SetUint64(constants.LDC))
	assert.ErrorContains(itx.Verify(bctx, bs), "CheckRepay failed: don't need to repay")

	// borrow
	input := &ld.TxTransfer{
		Nonce:  1,
		From:   &lender.id,
		To:     &borrower.id,
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
		From:      borrower.id,
		To:        &lender.id,
		Token:     &token,
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	assert.NoError(txData.ExSignWith(util.Signer2))
	tt = txData.ToTransaction()
	tt.Timestamp = bs.Timestamp()
	tt.Gas = tt.RequiredGas(bctx.FeeConfig().ThresholdGas)
	borrowerGas := tt.Gas
	itx, err = NewTx(tt, true)
	assert.NoError(err)
	assert.ErrorContains(itx.Verify(bctx, bs),
		"CheckBorrow failed: insufficient $LDC balance, expected 1000000000, got 0")

	assert.NoError(lender.Add(token, new(big.Int).SetUint64(constants.LDC)))
	assert.NoError(lender.AddNonceTable(bs.Timestamp()+1, []uint64{1}))

	assert.NoError(itx.Verify(bctx, bs))
	assert.NoError(itx.Accept(bctx, bs))

	assert.Equal((lenderGas+borrowerGas)*bctx.Price, ldc.balanceOf(constants.NativeToken).Uint64())
	assert.Equal((lenderGas+borrowerGas)*100, miner.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(constants.LDC*2-borrowerGas*(bctx.Price+100),
		borrower.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(constants.LDC-lenderGas*(bctx.Price+100),
		lender.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(uint64(0), lender.balanceOf(token).Uint64())
	assert.Equal(constants.LDC*2, borrower.balanceOf(token).Uint64())

	assert.Equal(1, len(lender.ld.LendingLedger))
	assert.Equal(0, len(lender.ld.NonceTable))
	assert.NotNil(lender.ld.LendingLedger[borrower.id])
	entry := lender.ld.LendingLedger[borrower.id]
	assert.Equal(constants.LDC, entry.Amount.Uint64())
	assert.Equal(bs.Timestamp(), entry.UpdateAt)
	assert.Equal(uint64(0), entry.DueTime)

	// repay after a day
	bctx.height++
	bctx.timestamp += 3600 * 24
	bs.height++
	bs.timestamp = bctx.timestamp
	lender.ld.Height++
	lender.ld.Timestamp = bs.Timestamp()
	borrower.ld.Height++
	borrower.ld.Timestamp = bs.Timestamp()

	txData = &ld.TxData{
		Type:      ld.TypeRepay,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     1,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      borrower.id,
		To:        &lender.id,
		Token:     &token,
		Amount:    new(big.Int).SetUint64(constants.LDC),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	tt = txData.ToTransaction()
	tt.Timestamp = bs.Timestamp()
	tt.Gas = tt.RequiredGas(bctx.FeeConfig().ThresholdGas)
	borrowerGas += tt.Gas
	itx, err = NewTx(tt, true)
	assert.NoError(err)
	assert.NoError(itx.Verify(bctx, bs))
	assert.NoError(itx.Accept(bctx, bs))

	assert.Equal((lenderGas+borrowerGas)*bctx.Price, ldc.balanceOf(constants.NativeToken).Uint64())
	assert.Equal((lenderGas+borrowerGas)*100, miner.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(constants.LDC*2-borrowerGas*(bctx.Price+100),
		borrower.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(constants.LDC-lenderGas*(bctx.Price+100),
		lender.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(constants.LDC, lender.balanceOf(token).Uint64())
	assert.Equal(constants.LDC, borrower.balanceOf(token).Uint64())

	assert.NotNil(lender.ld.LendingLedger[borrower.id])
	entry = lender.ld.LendingLedger[borrower.id]

	interest := uint64(float64(constants.LDC) * 10_000 / 1_000_000)
	assert.Equal(interest, entry.Amount.Uint64(), "with 1 day interest")
	assert.Equal(bs.Timestamp(), entry.UpdateAt)
	assert.Equal(uint64(0), entry.DueTime)

	jsondata, err := itx.MarshalJSON()
	assert.NoError(err)
	// fmt.Println(string(jsondata))
	assert.Equal(`{"type":"TypeRepay","chainID":2357,"nonce":1,"gasTip":100,"gasFeeCap":1000,"from":"0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC","to":"0x44171C37Ff5D7B7bb8dcad5C81f16284A229e641","token":"$LDC","amount":1000000000,"signatures":["0c9a509724f0b2009160eb0680657178a6611b9762d1ed36f033cce1848657c64afb888b440c530c6550d68f3f0f382ccfe66c6346a36f8ca37ecfee22ce7fa700"],"gas":604,"id":"LkWrkmjvWd6HUMGEhdn7jXqxLS5feFTPVQE9uRuzrAx4nogQW"}`, string(jsondata))

	// repay again
	txData = &ld.TxData{
		Type:      ld.TypeRepay,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     2,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      borrower.id,
		To:        &lender.id,
		Token:     &token,
		Amount:    new(big.Int).SetUint64(constants.MilliLDC * 20),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	tt = txData.ToTransaction()
	tt.Timestamp = bs.Timestamp()
	tt.Gas = tt.RequiredGas(bctx.FeeConfig().ThresholdGas)
	borrowerGas += tt.Gas
	itx, err = NewTx(tt, true)
	assert.NoError(err)
	assert.NoError(itx.Verify(bctx, bs))
	assert.NoError(itx.Accept(bctx, bs))

	assert.Equal((lenderGas+borrowerGas)*bctx.Price, ldc.balanceOf(constants.NativeToken).Uint64())
	assert.Equal((lenderGas+borrowerGas)*100, miner.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(constants.LDC*2-borrowerGas*(bctx.Price+100),
		borrower.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(constants.LDC-lenderGas*(bctx.Price+100),
		lender.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(constants.LDC+interest, lender.balanceOf(token).Uint64())
	assert.Equal(constants.LDC-interest, borrower.balanceOf(token).Uint64())
	assert.Nil(lender.ld.LendingLedger[borrower.id], "clear entry when repay all")

	assert.NoError(bs.VerifyState())
}
