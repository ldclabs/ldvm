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
	"github.com/ldclabs/ldvm/util/encoding"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTxBorrow(t *testing.T) {
	assert := assert.New(t)

	// SyntacticVerify
	tx := &TxBorrow{}
	assert.ErrorContains(tx.SyntacticVerify(), "nil pointer")
	_, err := tx.MarshalJSON()
	require.NoError(t, err)

	ctx := NewMockChainContext()
	cs := ctx.MockChainState()
	token := ld.MustNewToken("$LDC")
	borrower := signer.Signer1.Key().Address()
	lender := signer.Signer2.Key().Address()

	ltx := &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeBorrow,
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
		Type:      ld.TypeBorrow,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      borrower,
		To:        &lender,
		Amount:    new(big.Int).SetUint64(1),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "invalid amount, should be nil")

	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeBorrow,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      borrower,
		To:        &lender,
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "invalid data")

	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeBorrow,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      borrower,
		To:        &lender,
		Data:      []byte("你好👋"),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "no exSignatures")

	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeBorrow,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      borrower,
		To:        &lender,
		Data:      []byte("你好👋"),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.ExSignWith(signer.Signer2))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "cbor: unexpected following extraneous data")

	input := &ld.TxTransfer{}
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeBorrow,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      borrower,
		To:        &lender,
		Data:      input.Bytes(),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.ExSignWith(signer.Signer2))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "nil from as lender")

	input = &ld.TxTransfer{
		From: &lender,
	}
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeBorrow,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      borrower,
		To:        ids.GenesisAccount.Ptr(),
		Data:      input.Bytes(),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.ExSignWith(signer.Signer2))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err,
		"invalid to as borrower, expected 0x44171C37Ff5D7B7bb8Dcad5C81f16284A229E641, got 0xFFfFFFfFfffFFfFFffFFFfFfFffFFFfffFfFFFff")

	input = &ld.TxTransfer{
		From: &lender,
	}
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeBorrow,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      borrower,
		To:        &lender,
		Data:      input.Bytes(),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.ExSignWith(signer.Signer2))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "nil to as borrower")

	input = &ld.TxTransfer{
		From: &lender,
		To:   ids.GenesisAccount.Ptr(),
	}
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeBorrow,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      borrower,
		To:        &lender,
		Data:      input.Bytes(),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.ExSignWith(signer.Signer2))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err,
		"invalid from as lender, expected 0xFFfFFFfFfffFFfFFffFFFfFfFffFFFfffFfFFFff, got 0x8db97c7cECe249C2b98bdc0226cc4C2A57bF52fc")

	input = &ld.TxTransfer{
		From: &lender,
		To:   &borrower,
	}
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeBorrow,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      borrower,
		To:        &lender,
		Token:     token.Ptr(),
		Data:      input.Bytes(),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.ExSignWith(signer.Signer2))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err,
		"invalid token, expected NativeLDC, got $LDC")

	input = &ld.TxTransfer{
		From:  &lender,
		To:    &borrower,
		Token: token.Ptr(),
	}
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeBorrow,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      borrower,
		To:        &lender,
		Token:     &ids.NativeToken,
		Data:      input.Bytes(),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.ExSignWith(signer.Signer2))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err,
		"invalid token, expected $LDC, got NativeLDC")

	input = &ld.TxTransfer{
		From:  &lender,
		To:    &borrower,
		Token: token.Ptr(),
	}
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeBorrow,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      borrower,
		To:        &lender,
		Token:     token.Ptr(),
		Data:      input.Bytes(),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.ExSignWith(signer.Signer2))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "invalid amount, expected >= 1")

	input = &ld.TxTransfer{
		From:   &lender,
		To:     &borrower,
		Token:  token.Ptr(),
		Amount: new(big.Int).SetUint64(unit.LDC),
	}
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeBorrow,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      borrower,
		To:        &lender,
		Token:     token.Ptr(),
		Data:      input.Bytes(),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.ExSignWith(signer.Signer2))
	assert.NoError(ltx.SyntacticVerify())
	ltx.Timestamp = cs.Timestamp()
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "data expired")

	dueTime := cs.Timestamp()
	dueTimeData, err := encoding.MarshalCBOR(dueTime)
	require.NoError(t, err)
	input = &ld.TxTransfer{
		From:   &lender,
		To:     &borrower,
		Token:  token.Ptr(),
		Amount: new(big.Int).SetUint64(unit.LDC),
		Expire: cs.Timestamp(),
		Data:   dueTimeData,
	}
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeBorrow,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      borrower,
		To:        &lender,
		Token:     token.Ptr(),
		Data:      input.Bytes(),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.ExSignWith(signer.Signer2))
	assert.NoError(ltx.SyntacticVerify())
	ltx.Timestamp = cs.Timestamp()
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "invalid dueTime, expected > 1000, got 1000")

	dueTime = cs.Timestamp() + 3600*24
	dueTimeData, err = encoding.MarshalCBOR(dueTime)
	require.NoError(t, err)
	input = &ld.TxTransfer{
		From:   &lender,
		To:     &borrower,
		Token:  token.Ptr(),
		Amount: new(big.Int).SetUint64(unit.LDC),
		Expire: cs.Timestamp(),
		Data:   dueTimeData,
	}
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeBorrow,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      borrower,
		To:        &lender,
		Token:     token.Ptr(),
		Data:      input.Bytes(),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.ExSignWith(signer.Signer2))
	assert.NoError(ltx.SyntacticVerify())
	ltx.Timestamp = cs.Timestamp()
	itx, err := NewTx(ltx)
	require.NoError(t, err)

	cs.CommitAccounts()
	assert.ErrorContains(itx.Apply(ctx, cs),
		"insufficient NativeLDC balance, expected 2374900, got 0")
	cs.CheckoutAccounts()

	cs.MustAccount(borrower).Add(ids.NativeToken, new(big.Int).SetUint64(unit.LDC))
	cs.CommitAccounts()
	assert.ErrorContains(itx.Apply(ctx, cs),
		"invalid lending")
	cs.CheckoutAccounts()

	lcfg := &ld.LendingConfig{
		DailyInterest:   10_000,
		OverdueInterest: 10_000,
		MinAmount:       new(big.Int).SetUint64(unit.LDC),
		MaxAmount:       new(big.Int).SetUint64(unit.LDC * 10),
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
	require.NoError(t, err)

	lenderAcc := cs.MustAccount(lender)
	lenderAcc.Add(ids.NativeToken, new(big.Int).SetUint64(unit.LDC*2))
	assert.NoError(itx.Apply(ctx, cs))

	lenderGas := ltx.Gas()
	assert.Equal(lenderGas*ctx.Price,
		itx.(*TxOpenLending).ldc.Balance().Uint64())
	assert.Equal(lenderGas*100,
		itx.(*TxOpenLending).miner.Balance().Uint64())
	assert.Equal(unit.LDC-lenderGas*(ctx.Price+100),
		lenderAcc.Balance().Uint64())
	require.NotNil(t, lenderAcc.LD().Lending)
	require.NotNil(t, lenderAcc.Ledger())
	assert.Equal(0, len(lenderAcc.Ledger().Lending))

	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeBorrow,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      borrower,
		To:        &lender,
		Token:     token.Ptr(),
		Data:      input.Bytes(),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.ExSignWith(signer.Signer2))
	assert.NoError(ltx.SyntacticVerify())
	ltx.Timestamp = cs.Timestamp()
	itx, err = NewTx(ltx)
	require.NoError(t, err)

	cs.CommitAccounts()
	assert.ErrorContains(itx.Apply(ctx, cs),
		"invalid token, expected NativeLDC, got $LDC")
	cs.CheckoutAccounts()

	input = &ld.TxTransfer{
		From:   &lender,
		To:     &borrower,
		Amount: new(big.Int).SetUint64(unit.LDC),
		Expire: cs.Timestamp() + 1,
		Data:   dueTimeData,
	}
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeBorrow,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      borrower,
		To:        &lender,
		Data:      input.Bytes(),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.ExSignWith(signer.Signer2))
	assert.NoError(ltx.SyntacticVerify())
	ltx.Timestamp = cs.Timestamp()
	itx, err = NewTx(ltx)
	require.NoError(t, err)

	cs.CommitAccounts()
	assert.ErrorContains(itx.Apply(ctx, cs),
		"insufficient transferable NativeLDC balance, expected 1000000000, got 998155300")
	cs.CheckoutAccounts()

	assert.NoError(lenderAcc.Add(ids.NativeToken, new(big.Int).SetUint64(unit.LDC*2)))
	cs.CommitAccounts()
	assert.ErrorContains(itx.Apply(ctx, cs),
		"nonce 0 not exists at 1001")
	cs.CheckoutAccounts()

	assert.NoError(lenderAcc.UpdateNonceTable(cs.Timestamp()+1, []uint64{0, 1}))
	assert.NoError(itx.Apply(ctx, cs))
	cs.CommitAccounts()

	borrowerGas := ltx.Gas()
	borrowerAcc := cs.MustAccount(borrower)
	assert.Equal((lenderGas+borrowerGas)*ctx.Price,
		itx.(*TxBorrow).ldc.Balance().Uint64())
	assert.Equal((lenderGas+borrowerGas)*100,
		itx.(*TxBorrow).miner.Balance().Uint64())
	assert.Equal(unit.LDC*2-borrowerGas*(ctx.Price+100),
		borrowerAcc.BalanceOfAll(ids.NativeToken).Uint64())
	assert.Equal(uint64(1), borrowerAcc.Nonce())
	assert.Equal(1, len(lenderAcc.Ledger().Lending))
	require.NotNil(t, lenderAcc.Ledger().Lending[borrowerAcc.ID().AsKey()])
	entry := lenderAcc.Ledger().Lending[borrowerAcc.ID().AsKey()]
	assert.Equal(unit.LDC, entry.Amount.Uint64())
	assert.Equal(cs.Timestamp(), entry.UpdateAt)
	assert.Equal(cs.Timestamp()+3600*24, entry.DueTime)

	jsondata, err := itx.MarshalJSON()
	require.NoError(t, err)
	// fmt.Println(string(jsondata))
	assert.Equal(`{"tx":{"type":"TypeBorrow","chainID":2357,"nonce":0,"gasTip":100,"gasFeeCap":1000,"from":"0x8db97c7cECe249C2b98bdc0226cc4C2A57bF52fc","to":"0x44171C37Ff5D7B7bb8Dcad5C81f16284A229E641","data":{"from":"0x44171C37Ff5D7B7bb8Dcad5C81f16284A229E641","to":"0x8db97c7cECe249C2b98bdc0226cc4C2A57bF52fc","amount":1000000000,"expire":1001,"data":"GgABVWhrgiP-"}},"sigs":["M-ev3GGkRy8KvHmbAJoENKO0y8e1WVmuPCJ3mWNcUBtDMyU2gpu9pKwp_R-kdFt8d6xU_u1cxnHYygUvfV0mbwA2Ef2B"],"exSigs":["NgAL6RskeRV3yquqE6Qd5nUHpgWFrnf2hVzyqTwypyovOsyEaBt_rIa_ZwYxkGCmJvnkAih6sLeRoSwyd0U38wFTlQpE"],"id":"rg9f2lvhqTHfn4TE6cbz51IXE-C2vLGvBLR-c_bfAcp-LouL"}`, string(jsondata))

	// borrow again after a day
	ctx.height++
	ctx.timestamp += 3600 * 24
	cs.CheckoutAccounts()

	assert.NoError(lenderAcc.UpdateNonceTable(cs.Timestamp()+1, []uint64{999}))
	input = &ld.TxTransfer{
		Nonce:  999,
		From:   &lender,
		To:     &borrower,
		Amount: new(big.Int).SetUint64(unit.LDC),
		Expire: cs.Timestamp() + 1,
	}
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeBorrow,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     1,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      borrower,
		To:        &lender,
		Data:      input.Bytes(),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.ExSignWith(signer.Signer2))
	assert.NoError(ltx.SyntacticVerify())
	ltx.Timestamp = cs.Timestamp()
	itx, err = NewTx(ltx)
	require.NoError(t, err)
	assert.NoError(itx.Apply(ctx, cs))

	borrowerGas += ltx.Gas()
	assert.Equal((lenderGas+borrowerGas)*ctx.Price,
		itx.(*TxBorrow).ldc.Balance().Uint64())
	assert.Equal((lenderGas+borrowerGas)*100,
		itx.(*TxBorrow).miner.Balance().Uint64())
	assert.Equal(unit.LDC*3-borrowerGas*(ctx.Price+100),
		borrowerAcc.BalanceOfAll(ids.NativeToken).Uint64())
	assert.Equal(unit.LDC-lenderGas*(ctx.Price+100),
		lenderAcc.Balance().Uint64())
	require.NotNil(t, lenderAcc.Ledger().Lending[borrower.AsKey()])
	entry = lenderAcc.Ledger().Lending[borrower.AsKey()]
	rate := 1 + float64(10_000)/1_000_000
	assert.Equal(uint64(float64(unit.LDC)*rate)+unit.LDC, entry.Amount.Uint64(),
		"with 1 day interest")
	assert.Equal(cs.Timestamp(), entry.UpdateAt)
	assert.Equal(uint64(0), entry.DueTime, "overwrite to 0")

	assert.NoError(cs.VerifyState())
}
