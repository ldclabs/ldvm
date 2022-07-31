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

func TestTxBorrow(t *testing.T) {
	assert := assert.New(t)

	// SyntacticVerify
	tx := &TxBorrow{}
	assert.ErrorContains(tx.SyntacticVerify(), "nil pointer")
	_, err := tx.MarshalJSON()
	assert.NoError(err)

	bctx := NewMockBCtx()
	bs := bctx.MockBS()
	token := ld.MustNewToken("$LDC")
	borrower := util.Signer1.Address()
	lender := util.Signer2.Address()

	txData := &ld.TxData{
		Type:      ld.TypeBorrow,
		ChainID:   bctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
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
		ChainID:   bctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      borrower,
		To:        &lender,
		Amount:    new(big.Int).SetUint64(1),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx2(txData.ToTransaction())
	assert.ErrorContains(err, "invalid amount, should be nil")

	txData = &ld.TxData{
		Type:      ld.TypeBorrow,
		ChainID:   bctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      borrower,
		To:        &lender,
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx2(txData.ToTransaction())
	assert.ErrorContains(err, "invalid data")

	txData = &ld.TxData{
		Type:      ld.TypeBorrow,
		ChainID:   bctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      borrower,
		To:        &lender,
		Data:      []byte("你好👋"),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx2(txData.ToTransaction())
	assert.ErrorContains(err, "cbor: cannot unmarshal")

	input := &ld.TxTransfer{}
	txData = &ld.TxData{
		Type:      ld.TypeBorrow,
		ChainID:   bctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
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
		ChainID:   bctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
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
		ChainID:   bctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
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
		ChainID:   bctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
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
		ChainID:   bctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
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
		ChainID:   bctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
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
		ChainID:   bctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
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
		ChainID:   bctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      borrower,
		To:        &lender,
		Token:     &token,
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	tt := txData.ToTransaction()
	tt.Timestamp = bs.Timestamp()
	_, err = NewTx2(tt)
	assert.ErrorContains(err, "data expired")

	dueTime := bs.Timestamp()
	dueTimeData, err := util.MarshalCBOR(dueTime)
	assert.NoError(err)
	input = &ld.TxTransfer{
		From:   &lender,
		To:     &borrower,
		Token:  &token,
		Amount: new(big.Int).SetUint64(constants.LDC),
		Expire: bs.Timestamp(),
		Data:   dueTimeData,
	}
	txData = &ld.TxData{
		Type:      ld.TypeBorrow,
		ChainID:   bctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      borrower,
		To:        &lender,
		Token:     &token,
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	tt = txData.ToTransaction()
	tt.Timestamp = bs.Timestamp()
	_, err = NewTx2(tt)
	assert.ErrorContains(err, "invalid dueTime, expected > 1000, got 1000")

	dueTime = bs.Timestamp() + 3600*24
	dueTimeData, err = util.MarshalCBOR(dueTime)
	assert.NoError(err)
	input = &ld.TxTransfer{
		From:   &lender,
		To:     &borrower,
		Token:  &token,
		Amount: new(big.Int).SetUint64(constants.LDC),
		Expire: bs.Timestamp(),
		Data:   dueTimeData,
	}
	txData = &ld.TxData{
		Type:      ld.TypeBorrow,
		ChainID:   bctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      borrower,
		To:        &lender,
		Token:     &token,
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	tt = txData.ToTransaction()
	tt.Timestamp = bs.Timestamp()
	_, err = NewTx2(tt)
	assert.ErrorContains(err,
		"invalid exSignatures, Transaction.ExSigners error: DeriveSigners error: no signature")

	assert.NoError(txData.ExSignWith(util.Signer2))
	tt = txData.ToTransaction()
	tt.Timestamp = bs.Timestamp()
	itx, err := NewTx2(tt)
	assert.NoError(err)

	bs.CommitAccounts()
	assert.ErrorContains(itx.Apply(bctx, bs),
		"insufficient NativeLDC balance, expected 2348500, got 0")
	bs.CheckoutAccounts()

	bs.MustAccount(borrower).Add(constants.NativeToken, new(big.Int).SetUint64(constants.LDC))
	bs.CommitAccounts()
	assert.ErrorContains(itx.Apply(bctx, bs),
		"Account(0x44171C37Ff5D7B7bb8dcad5C81f16284A229e641).Borrow error: invalid lending")
	bs.CheckoutAccounts()

	lcfg := &ld.LendingConfig{
		DailyInterest:   10_000,
		OverdueInterest: 10_000,
		MinAmount:       new(big.Int).SetUint64(constants.LDC),
		MaxAmount:       new(big.Int).SetUint64(constants.LDC * 10),
	}
	assert.NoError(lcfg.SyntacticVerify())
	txData = &ld.TxData{
		Type:      ld.TypeOpenLending,
		ChainID:   bctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      lender,
		Data:      ld.MustMarshal(lcfg),
	}
	assert.NoError(txData.SignWith(util.Signer2))
	tt = txData.ToTransaction()
	tt.Timestamp = bs.Timestamp()
	itx, err = NewTx2(tt)
	assert.NoError(err)

	lenderAcc := bs.MustAccount(lender)
	lenderAcc.Add(constants.NativeToken, new(big.Int).SetUint64(constants.LDC))
	assert.NoError(itx.Apply(bctx, bs))

	lenderGas := tt.Gas()
	assert.Equal(lenderGas*bctx.Price,
		itx.(*TxOpenLending).ldc.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(lenderGas*100,
		itx.(*TxOpenLending).miner.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(constants.LDC-lenderGas*(bctx.Price+100),
		lenderAcc.balanceOf(constants.NativeToken).Uint64())
	assert.NotNil(lenderAcc.ld.Lending)
	assert.NotNil(lenderAcc.ledger)
	assert.Equal(0, len(lenderAcc.ledger.Lending))

	txData = &ld.TxData{
		Type:      ld.TypeBorrow,
		ChainID:   bctx.ChainConfig().ChainID,
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
	itx, err = NewTx2(tt)
	assert.NoError(err)

	bs.CommitAccounts()
	assert.ErrorContains(itx.Apply(bctx, bs),
		"TxBorrow.Apply error: Account(0x44171C37Ff5D7B7bb8dcad5C81f16284A229e641).Borrow error: invalid token, expected NativeLDC, got $LDC")
	bs.CheckoutAccounts()

	input = &ld.TxTransfer{
		From:   &lender,
		To:     &borrower,
		Amount: new(big.Int).SetUint64(constants.LDC),
		Expire: bs.Timestamp(),
		Data:   dueTimeData,
	}
	txData = &ld.TxData{
		Type:      ld.TypeBorrow,
		ChainID:   bctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      borrower,
		To:        &lender,
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	assert.NoError(txData.ExSignWith(util.Signer2))
	tt = txData.ToTransaction()
	tt.Timestamp = bs.Timestamp()
	itx, err = NewTx2(tt)
	assert.NoError(err)

	bs.CommitAccounts()
	assert.ErrorContains(itx.Apply(bctx, bs),
		"Account(0x44171C37Ff5D7B7bb8dcad5C81f16284A229e641).Borrow error: insufficient NativeLDC balance, expected 1000000000, got 998178400")
	bs.CheckoutAccounts()

	assert.NoError(lenderAcc.Add(constants.NativeToken, new(big.Int).SetUint64(constants.LDC*2)))
	bs.CommitAccounts()
	assert.ErrorContains(itx.Apply(bctx, bs),
		"Account(0x44171C37Ff5D7B7bb8dcad5C81f16284A229e641).SubByNonceTable error: nonce 0 not exists at 1000")
	bs.CheckoutAccounts()

	assert.NoError(lenderAcc.AddNonceTable(bs.Timestamp(), []uint64{0, 1}))
	assert.NoError(itx.Apply(bctx, bs))
	bs.CommitAccounts()

	borrowerGas := tt.Gas()
	borrowerAcc := bs.MustAccount(borrower)
	assert.Equal((lenderGas+borrowerGas)*bctx.Price,
		itx.(*TxBorrow).ldc.balanceOf(constants.NativeToken).Uint64())
	assert.Equal((lenderGas+borrowerGas)*100,
		itx.(*TxBorrow).miner.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(constants.LDC*2-borrowerGas*(bctx.Price+100),
		borrowerAcc.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(uint64(1), borrowerAcc.Nonce())
	assert.Equal(1, len(lenderAcc.ledger.Lending))
	assert.NotNil(lenderAcc.ledger.Lending[borrowerAcc.id.AsKey()])
	entry := lenderAcc.ledger.Lending[borrowerAcc.id.AsKey()]
	assert.Equal(constants.LDC, entry.Amount.Uint64())
	assert.Equal(bs.Timestamp(), entry.UpdateAt)
	assert.Equal(bs.Timestamp()+3600*24, entry.DueTime)

	jsondata, err := itx.MarshalJSON()
	assert.NoError(err)
	// fmt.Println(string(jsondata))
	assert.Equal(`{"type":"TypeBorrow","chainID":2357,"nonce":0,"gasTip":100,"gasFeeCap":1000,"from":"0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC","to":"0x44171C37Ff5D7B7bb8dcad5C81f16284A229e641","data":{"from":"0x44171C37Ff5D7B7bb8dcad5C81f16284A229e641","to":"0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC","amount":1000000000,"expire":1000,"data":"0x1a00015568c54d609b"},"signatures":["d23a9c587172465060d4fce7cf73770e3914a2d8f35622856083dab0fdd0bb4a1a8b65671392c6bd15d4eff1ae4cd79516ad5d5ea25571b6278571967c87458701"],"exSignatures":["b82d77943d8761685f7ca432cf5059455a278bb34089cbb02d85b20cd87c430500002e111cb2f693c0c75ba2c96085fce89827c39d9c827b349fb873fac32eaa01"],"id":"2qNonZbDwMaPFV3K8fWhCe27DVerF5WsjDp5cMCKnuzmxwXT6M"}`, string(jsondata))

	// borrow again after a day
	bctx.height++
	bctx.timestamp += 3600 * 24
	bs.CheckoutAccounts()

	assert.NoError(lenderAcc.AddNonceTable(bs.Timestamp(), []uint64{999}))
	input = &ld.TxTransfer{
		Nonce:  999,
		From:   &lender,
		To:     &borrower,
		Amount: new(big.Int).SetUint64(constants.LDC),
		Expire: bs.Timestamp(),
	}
	txData = &ld.TxData{
		Type:      ld.TypeBorrow,
		ChainID:   bctx.ChainConfig().ChainID,
		Nonce:     1,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      borrower,
		To:        &lender,
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	assert.NoError(txData.ExSignWith(util.Signer2))
	tt = txData.ToTransaction()
	tt.Timestamp = bs.Timestamp()
	itx, err = NewTx2(tt)
	assert.NoError(err)
	assert.NoError(itx.Apply(bctx, bs))

	borrowerGas += tt.Gas()
	assert.Equal((lenderGas+borrowerGas)*bctx.Price,
		itx.(*TxBorrow).ldc.balanceOf(constants.NativeToken).Uint64())
	assert.Equal((lenderGas+borrowerGas)*100,
		itx.(*TxBorrow).miner.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(constants.LDC*3-borrowerGas*(bctx.Price+100),
		borrowerAcc.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(constants.LDC-lenderGas*(bctx.Price+100),
		lenderAcc.balanceOf(constants.NativeToken).Uint64())
	assert.NotNil(lenderAcc.ledger.Lending[borrower.AsKey()])
	entry = lenderAcc.ledger.Lending[borrower.AsKey()]
	rate := 1 + float64(10_000)/1_000_000
	assert.Equal(uint64(float64(constants.LDC)*rate)+constants.LDC, entry.Amount.Uint64(),
		"with 1 day interest")
	assert.Equal(bs.Timestamp(), entry.UpdateAt)
	assert.Equal(uint64(0), entry.DueTime, "overwrite to 0")

	assert.NoError(bs.VerifyState())
}
