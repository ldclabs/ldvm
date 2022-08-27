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

func TestTxBorrow(t *testing.T) {
	assert := assert.New(t)

	// SyntacticVerify
	tx := &TxBorrow{}
	assert.ErrorContains(tx.SyntacticVerify(), "nil pointer")
	_, err := tx.MarshalJSON()
	assert.NoError(err)

	ctx := NewMockChainContext()
	cs := ctx.MockChainState()
	token := ld.MustNewToken("$LDC")
	borrower := util.Signer1.Address()
	lender := util.Signer2.Address()

	txData := &ld.TxData{
		Type:      ld.TypeBorrow,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      borrower,
	}
	assert.NoError(txData.SyntacticVerify())
	_, err = NewTx2(txData.ToTransaction())
	assert.ErrorContains(err, "DeriveSigners error: no signature")

	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx2(txData.ToTransaction())
	assert.ErrorContains(err, "nil to as lender")

	txData = &ld.TxData{
		Type:      ld.TypeBorrow,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      borrower,
		To:        &lender,
		Amount:    new(big.Int).SetUint64(1),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx2(txData.ToTransaction())
	assert.ErrorContains(err, "invalid amount, should be nil")

	txData = &ld.TxData{
		Type:      ld.TypeBorrow,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      borrower,
		To:        &lender,
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx2(txData.ToTransaction())
	assert.ErrorContains(err, "invalid data")

	txData = &ld.TxData{
		Type:      ld.TypeBorrow,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      borrower,
		To:        &lender,
		Data:      []byte("你好👋"),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx2(txData.ToTransaction())
	assert.ErrorContains(err, "cbor: unexpected following extraneous data")

	input := &ld.TxTransfer{}
	txData = &ld.TxData{
		Type:      ld.TypeBorrow,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      borrower,
		To:        &lender,
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx2(txData.ToTransaction())
	assert.ErrorContains(err, "nil from as lender")

	input = &ld.TxTransfer{
		From: &lender,
	}
	txData = &ld.TxData{
		Type:      ld.TypeBorrow,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      borrower,
		To:        &constants.GenesisAccount,
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx2(txData.ToTransaction())
	assert.ErrorContains(err,
		"invalid to as borrower, expected 0x44171C37Ff5D7B7bb8dcad5C81f16284A229e641, got 0xFFfFfFffFFfffFFfFFfFFFFFffFFFffffFfFFFfF")

	input = &ld.TxTransfer{
		From: &lender,
	}
	txData = &ld.TxData{
		Type:      ld.TypeBorrow,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      borrower,
		To:        &lender,
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx2(txData.ToTransaction())
	assert.ErrorContains(err, "nil to as borrower")

	input = &ld.TxTransfer{
		From: &lender,
		To:   &constants.GenesisAccount,
	}
	txData = &ld.TxData{
		Type:      ld.TypeBorrow,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      borrower,
		To:        &lender,
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx2(txData.ToTransaction())
	assert.ErrorContains(err,
		"invalid from as lender, expected 0xFFfFfFffFFfffFFfFFfFFFFFffFFFffffFfFFFfF, got 0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC")

	input = &ld.TxTransfer{
		From: &lender,
		To:   &borrower,
	}
	txData = &ld.TxData{
		Type:      ld.TypeBorrow,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      borrower,
		To:        &lender,
		Token:     &token,
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx2(txData.ToTransaction())
	assert.ErrorContains(err,
		"invalid token, expected NativeLDC, got $LDC")

	input = &ld.TxTransfer{
		From:  &lender,
		To:    &borrower,
		Token: &token,
	}
	txData = &ld.TxData{
		Type:      ld.TypeBorrow,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      borrower,
		To:        &lender,
		Token:     &constants.NativeToken,
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx2(txData.ToTransaction())
	assert.ErrorContains(err,
		"invalid token, expected $LDC, got NativeLDC")

	input = &ld.TxTransfer{
		From:  &lender,
		To:    &borrower,
		Token: &token,
	}
	txData = &ld.TxData{
		Type:      ld.TypeBorrow,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      borrower,
		To:        &lender,
		Token:     &token,
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx2(txData.ToTransaction())
	assert.ErrorContains(err, "invalid amount, expected >= 1")

	input = &ld.TxTransfer{
		From:   &lender,
		To:     &borrower,
		Token:  &token,
		Amount: new(big.Int).SetUint64(constants.LDC),
	}
	txData = &ld.TxData{
		Type:      ld.TypeBorrow,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      borrower,
		To:        &lender,
		Token:     &token,
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	tt := txData.ToTransaction()
	tt.Timestamp = cs.Timestamp()
	_, err = NewTx2(tt)
	assert.ErrorContains(err, "data expired")

	dueTime := cs.Timestamp()
	dueTimeData, err := util.MarshalCBOR(dueTime)
	assert.NoError(err)
	input = &ld.TxTransfer{
		From:   &lender,
		To:     &borrower,
		Token:  &token,
		Amount: new(big.Int).SetUint64(constants.LDC),
		Expire: cs.Timestamp(),
		Data:   dueTimeData,
	}
	txData = &ld.TxData{
		Type:      ld.TypeBorrow,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      borrower,
		To:        &lender,
		Token:     &token,
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	tt = txData.ToTransaction()
	tt.Timestamp = cs.Timestamp()
	_, err = NewTx2(tt)
	assert.ErrorContains(err, "invalid dueTime, expected > 1000, got 1000")

	dueTime = cs.Timestamp() + 3600*24
	dueTimeData, err = util.MarshalCBOR(dueTime)
	assert.NoError(err)
	input = &ld.TxTransfer{
		From:   &lender,
		To:     &borrower,
		Token:  &token,
		Amount: new(big.Int).SetUint64(constants.LDC),
		Expire: cs.Timestamp(),
		Data:   dueTimeData,
	}
	txData = &ld.TxData{
		Type:      ld.TypeBorrow,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      borrower,
		To:        &lender,
		Token:     &token,
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	tt = txData.ToTransaction()
	tt.Timestamp = cs.Timestamp()
	_, err = NewTx2(tt)
	assert.ErrorContains(err,
		"invalid exSignatures, Transaction.ExSigners error: DeriveSigners error: no signature")

	assert.NoError(txData.ExSignWith(util.Signer2))
	tt = txData.ToTransaction()
	tt.Timestamp = cs.Timestamp()
	itx, err := NewTx2(tt)
	assert.NoError(err)

	cs.CommitAccounts()
	assert.ErrorContains(itx.Apply(ctx, cs),
		"insufficient NativeLDC balance, expected 2348500, got 0")
	cs.CheckoutAccounts()

	cs.MustAccount(borrower).Add(constants.NativeToken, new(big.Int).SetUint64(constants.LDC))
	cs.CommitAccounts()
	assert.ErrorContains(itx.Apply(ctx, cs),
		"Account(0x44171C37Ff5D7B7bb8dcad5C81f16284A229e641).Borrow error: invalid lending")
	cs.CheckoutAccounts()

	lcfg := &ld.LendingConfig{
		DailyInterest:   10_000,
		OverdueInterest: 10_000,
		MinAmount:       new(big.Int).SetUint64(constants.LDC),
		MaxAmount:       new(big.Int).SetUint64(constants.LDC * 10),
	}
	assert.NoError(lcfg.SyntacticVerify())
	txData = &ld.TxData{
		Type:      ld.TypeOpenLending,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      lender,
		Data:      ld.MustMarshal(lcfg),
	}
	assert.NoError(txData.SignWith(util.Signer2))
	tt = txData.ToTransaction()
	tt.Timestamp = cs.Timestamp()
	itx, err = NewTx2(tt)
	assert.NoError(err)

	lenderAcc := cs.MustAccount(lender)
	lenderAcc.Add(constants.NativeToken, new(big.Int).SetUint64(constants.LDC))
	assert.NoError(itx.Apply(ctx, cs))

	lenderGas := tt.Gas()
	assert.Equal(lenderGas*ctx.Price,
		itx.(*TxOpenLending).ldc.Balance().Uint64())
	assert.Equal(lenderGas*100,
		itx.(*TxOpenLending).miner.Balance().Uint64())
	assert.Equal(constants.LDC-lenderGas*(ctx.Price+100),
		lenderAcc.Balance().Uint64())
	assert.NotNil(lenderAcc.ld.Lending)
	assert.NotNil(lenderAcc.ledger)
	assert.Equal(0, len(lenderAcc.ledger.Lending))

	txData = &ld.TxData{
		Type:      ld.TypeBorrow,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      borrower,
		To:        &lender,
		Token:     &token,
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	assert.NoError(txData.ExSignWith(util.Signer2))
	tt = txData.ToTransaction()
	tt.Timestamp = cs.Timestamp()
	itx, err = NewTx2(tt)
	assert.NoError(err)

	cs.CommitAccounts()
	assert.ErrorContains(itx.Apply(ctx, cs),
		"TxBorrow.Apply error: Account(0x44171C37Ff5D7B7bb8dcad5C81f16284A229e641).Borrow error: invalid token, expected NativeLDC, got $LDC")
	cs.CheckoutAccounts()

	input = &ld.TxTransfer{
		From:   &lender,
		To:     &borrower,
		Amount: new(big.Int).SetUint64(constants.LDC),
		Expire: cs.Timestamp(),
		Data:   dueTimeData,
	}
	txData = &ld.TxData{
		Type:      ld.TypeBorrow,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      borrower,
		To:        &lender,
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	assert.NoError(txData.ExSignWith(util.Signer2))
	tt = txData.ToTransaction()
	tt.Timestamp = cs.Timestamp()
	itx, err = NewTx2(tt)
	assert.NoError(err)

	cs.CommitAccounts()
	assert.ErrorContains(itx.Apply(ctx, cs),
		"Account(0x44171C37Ff5D7B7bb8dcad5C81f16284A229e641).Borrow error: insufficient NativeLDC balance, expected 1000000000, got 998178400")
	cs.CheckoutAccounts()

	assert.NoError(lenderAcc.Add(constants.NativeToken, new(big.Int).SetUint64(constants.LDC*2)))
	cs.CommitAccounts()
	assert.ErrorContains(itx.Apply(ctx, cs),
		"Account(0x44171C37Ff5D7B7bb8dcad5C81f16284A229e641).SubByNonceTable error: nonce 0 not exists at 1000")
	cs.CheckoutAccounts()

	assert.NoError(lenderAcc.AddNonceTable(cs.Timestamp(), []uint64{0, 1}))
	assert.NoError(itx.Apply(ctx, cs))
	cs.CommitAccounts()

	borrowerGas := tt.Gas()
	borrowerAcc := cs.MustAccount(borrower)
	assert.Equal((lenderGas+borrowerGas)*ctx.Price,
		itx.(*TxBorrow).ldc.Balance().Uint64())
	assert.Equal((lenderGas+borrowerGas)*100,
		itx.(*TxBorrow).miner.Balance().Uint64())
	assert.Equal(constants.LDC*2-borrowerGas*(ctx.Price+100),
		borrowerAcc.Balance().Uint64())
	assert.Equal(uint64(1), borrowerAcc.Nonce())
	assert.Equal(1, len(lenderAcc.ledger.Lending))
	assert.NotNil(lenderAcc.ledger.Lending[borrowerAcc.id.AsKey()])
	entry := lenderAcc.ledger.Lending[borrowerAcc.id.AsKey()]
	assert.Equal(constants.LDC, entry.Amount.Uint64())
	assert.Equal(cs.Timestamp(), entry.UpdateAt)
	assert.Equal(cs.Timestamp()+3600*24, entry.DueTime)

	jsondata, err := itx.MarshalJSON()
	assert.NoError(err)
	// fmt.Println(string(jsondata))
	assert.Equal(`{"type":"TypeBorrow","chainID":2357,"nonce":0,"gasTip":100,"gasFeeCap":1000,"from":"0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC","to":"0x44171C37Ff5D7B7bb8dcad5C81f16284A229e641","data":{"from":"0x44171C37Ff5D7B7bb8dcad5C81f16284A229e641","to":"0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC","amount":1000000000,"expire":1000,"data":"0x1a00015568c54d609b"},"signatures":["d23a9c587172465060d4fce7cf73770e3914a2d8f35622856083dab0fdd0bb4a1a8b65671392c6bd15d4eff1ae4cd79516ad5d5ea25571b6278571967c87458701"],"exSignatures":["b82d77943d8761685f7ca432cf5059455a278bb34089cbb02d85b20cd87c430500002e111cb2f693c0c75ba2c96085fce89827c39d9c827b349fb873fac32eaa01"],"id":"2qNonZbDwMaPFV3K8fWhCe27DVerF5WsjDp5cMCKnuzmxwXT6M"}`, string(jsondata))

	// borrow again after a day
	ctx.height++
	ctx.timestamp += 3600 * 24
	cs.CheckoutAccounts()

	assert.NoError(lenderAcc.AddNonceTable(cs.Timestamp(), []uint64{999}))
	input = &ld.TxTransfer{
		Nonce:  999,
		From:   &lender,
		To:     &borrower,
		Amount: new(big.Int).SetUint64(constants.LDC),
		Expire: cs.Timestamp(),
	}
	txData = &ld.TxData{
		Type:      ld.TypeBorrow,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     1,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      borrower,
		To:        &lender,
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	assert.NoError(txData.ExSignWith(util.Signer2))
	tt = txData.ToTransaction()
	tt.Timestamp = cs.Timestamp()
	itx, err = NewTx2(tt)
	assert.NoError(err)
	assert.NoError(itx.Apply(ctx, cs))

	borrowerGas += tt.Gas()
	assert.Equal((lenderGas+borrowerGas)*ctx.Price,
		itx.(*TxBorrow).ldc.Balance().Uint64())
	assert.Equal((lenderGas+borrowerGas)*100,
		itx.(*TxBorrow).miner.Balance().Uint64())
	assert.Equal(constants.LDC*3-borrowerGas*(ctx.Price+100),
		borrowerAcc.Balance().Uint64())
	assert.Equal(constants.LDC-lenderGas*(ctx.Price+100),
		lenderAcc.Balance().Uint64())
	assert.NotNil(lenderAcc.ledger.Lending[borrower.AsKey()])
	entry = lenderAcc.ledger.Lending[borrower.AsKey()]
	rate := 1 + float64(10_000)/1_000_000
	assert.Equal(uint64(float64(constants.LDC)*rate)+constants.LDC, entry.Amount.Uint64(),
		"with 1 day interest")
	assert.Equal(cs.Timestamp(), entry.UpdateAt)
	assert.Equal(uint64(0), entry.DueTime, "overwrite to 0")

	assert.NoError(cs.VerifyState())
}
