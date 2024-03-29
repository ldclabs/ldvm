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

func TestTxCreateStake(t *testing.T) {
	assert := assert.New(t)

	// SyntacticVerify
	tx := &TxCreateStake{}
	assert.ErrorContains(tx.SyntacticVerify(), "nil pointer")
	_, err := tx.MarshalJSON()
	require.NoError(t, err)

	ctx := NewMockChainContext()
	cs := ctx.MockChainState()
	stake := ld.MustNewStake("#TEST")
	stakeid := ids.Address(stake)
	token := ld.MustNewToken("$TEST")

	from := cs.MustAccount(signer.Signer1.Key().Address())
	approver := signer.Signer2.Key()

	ltx := &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeCreateStake,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      from.ID(),
	}}
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "no signatures")

	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeCreateStake,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      from.ID(),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "nil to as stake account")

	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeCreateStake,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      from.ID(),
		To:        &stakeid,
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "nil amount")

	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeCreateStake,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      from.ID(),
		To:        &stakeid,
		Amount:    new(big.Int).SetUint64(100),
		Token:     token.Ptr(),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "invalid token, should be nil")

	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeCreateStake,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      from.ID(),
		To:        &stakeid,
		Amount:    new(big.Int).SetUint64(100),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "invalid data")

	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeCreateStake,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      from.ID(),
		To:        &stakeid,
		Amount:    new(big.Int).SetUint64(100),
		Data:      []byte("你好👋"),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "extraneous data")

	input := &ld.TxAccounter{}
	approverAddr := approver.Address()
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeCreateStake,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      from.ID(),
		To:        &approverAddr,
		Amount:    new(big.Int).SetUint64(100),
		Data:      input.Bytes(),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "invalid stake account 0x44171C37Ff5D7B7bb8Dcad5C81f16284A229E641")

	input = &ld.TxAccounter{}
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeCreateStake,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      from.ID(),
		To:        &stakeid,
		Amount:    new(big.Int).SetUint64(100),
		Data:      input.Bytes(),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "invalid threshold, expected >= 1")

	input = &ld.TxAccounter{
		Threshold: ld.Uint16Ptr(0),
		Keepers:   &signer.Keys{signer.Signer1.Key()},
	}
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeCreateStake,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      from.ID(),
		To:        &stakeid,
		Amount:    new(big.Int).SetUint64(100),
		Data:      input.Bytes(),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "invalid threshold, expected >= 1")

	input = &ld.TxAccounter{
		Threshold: ld.Uint16Ptr(1),
		Keepers:   &signer.Keys{signer.Signer1.Key()},
		Amount:    new(big.Int).SetUint64(unit.LDC * 10),
	}
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeCreateStake,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      from.ID(),
		To:        &stakeid,
		Amount:    new(big.Int).SetUint64(100),
		Data:      input.Bytes(),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "invalid amount, should be nil")

	input = &ld.TxAccounter{
		Threshold: ld.Uint16Ptr(1),
		Keepers:   &signer.Keys{signer.Signer1.Key()},
		Approver:  &approver,
	}
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeCreateStake,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      from.ID(),
		To:        &stakeid,
		Amount:    new(big.Int).SetUint64(100),
		Data:      input.Bytes(),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "invalid input data")

	scfg := &ld.StakeConfig{
		LockTime:    cs.Timestamp(),
		WithdrawFee: 100_000,
		MinAmount:   big.NewInt(100),
		MaxAmount:   big.NewInt(1000),
	}
	input = &ld.TxAccounter{
		Threshold: ld.Uint16Ptr(1),
		Keepers:   &signer.Keys{signer.Signer1.Key()},
		Approver:  &signer.Key{},
		Data:      ld.MustMarshal(scfg),
	}
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeCreateStake,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      from.ID(),
		To:        &stakeid,
		Amount:    new(big.Int).SetUint64(100),
		Data:      input.Bytes(),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err,
		"invalid approver, signer.Key.Valid: empty key")

	input = &ld.TxAccounter{
		Threshold: ld.Uint16Ptr(1),
		Keepers:   &signer.Keys{signer.Signer1.Key()},
		Approver:  &approver,
		Data:      ld.MustMarshal(scfg),
	}
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeCreateStake,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      from.ID(),
		To:        &stakeid,
		Amount:    new(big.Int).SetUint64(100),
		Data:      input.Bytes(),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	ltx.Timestamp = cs.Timestamp()
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "invalid lockTime, expected 0 or >= 1000")

	scfg = &ld.StakeConfig{
		LockTime:    cs.Timestamp() + 1000,
		WithdrawFee: 100_000,
		MinAmount:   new(big.Int).SetUint64(unit.LDC),
		MaxAmount:   new(big.Int).SetUint64(unit.LDC),
	}
	input = &ld.TxAccounter{
		Threshold: ld.Uint16Ptr(1),
		Keepers:   &signer.Keys{signer.Signer1.Key()},
		Approver:  &approver,
		Data:      ld.MustMarshal(scfg),
	}
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeCreateStake,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      from.ID(),
		To:        &stakeid,
		Amount:    new(big.Int).SetUint64(100),
		Data:      input.Bytes(),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	ltx.Timestamp = cs.Timestamp()
	itx, err := NewTx(ltx)
	require.NoError(t, err)

	cs.CommitAccounts()
	assert.ErrorContains(itx.Apply(ctx, cs), "insufficient NativeLDC balance, expected 2378300, got 0")
	cs.CheckoutAccounts()

	from.Add(ids.NativeToken, new(big.Int).SetUint64(unit.LDC))
	cs.CommitAccounts()
	assert.ErrorContains(itx.Apply(ctx, cs), "invalid amount, expected >= 1000000000000, got 100")
	cs.CheckoutAccounts()

	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeCreateStake,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      from.ID(),
		To:        &stakeid,
		Amount:    new(big.Int).SetUint64(unit.LDC * 1001),
		Data:      input.Bytes(),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	ltx.Timestamp = cs.Timestamp()
	itx, err = NewTx(ltx)
	require.NoError(t, err)
	cs.CommitAccounts()
	assert.ErrorContains(itx.Apply(ctx, cs),
		"insufficient NativeLDC balance, expected 1001002403500, got 1000000000")
	cs.CheckoutAccounts()

	from.Add(ids.NativeToken, new(big.Int).SetUint64(unit.LDC*1002))
	assert.NoError(itx.Apply(ctx, cs))

	stakeAcc := cs.MustAccount(stakeid)
	ldc := cs.MustAccount(ids.LDCAccount)
	miner, err := cs.LoadAccount(ctx.Builder())
	require.NoError(t, err)

	assert.Equal(ltx.Gas()*ctx.Price,
		ldc.Balance().Uint64())
	assert.Equal(ltx.Gas()*100,
		miner.Balance().Uint64())
	assert.Equal(unit.LDC, stakeAcc.Balance().Uint64())
	assert.Equal(unit.LDC*1001, stakeAcc.BalanceOfAll(ids.NativeToken).Uint64())
	assert.Equal(unit.LDC-ltx.Gas()*(ctx.Price+100),
		from.Balance().Uint64())

	assert.Equal(uint64(0), stakeAcc.Nonce())
	assert.Equal(uint16(1), stakeAcc.Threshold())
	assert.Equal(signer.Keys{signer.Signer1.Key()}, stakeAcc.Keepers())
	assert.Equal(approver.Address(), stakeAcc.LD().Approver.Address())
	assert.Equal(ld.StakeAccount, stakeAcc.LD().Type)
	assert.Nil(stakeAcc.LD().MaxTotalSupply)
	require.NotNil(t, stakeAcc.LD().Stake)
	require.NotNil(t, stakeAcc.Ledger())
	require.NotNil(t, stakeAcc.Ledger().Stake[from.ID().AsKey()])
	assert.Equal(unit.LDC*1000, stakeAcc.Ledger().Stake[from.ID().AsKey()].Amount.Uint64())

	jsondata, err := itx.MarshalJSON()
	require.NoError(t, err)
	// fmt.Println(string(jsondata))
	assert.Equal(`{"tx":{"type":"TypeCreateStake","chainID":2357,"nonce":0,"gasTip":100,"gasFeeCap":1000,"from":"0x8db97c7cECe249C2b98bdc0226cc4C2A57bF52fc","to":"0x0000000000000000000000000000002354455354","amount":1001000000000,"data":{"threshold":1,"keepers":["jbl8fOziScK5i9wCJsxMKle_UvwKxwPH"],"approver":"RBccN_9de3u43K1cgfFihKIp5kE1lmGG","data":"hlQAAAAAAAAAAAAAAAAAAAAAAAAAAAAZB9AaAAGGoMJEO5rKAMJEO5rKAMz9ac8"}},"sigs":["zTENIbPu7D6OBdAhXFiZw1SYEEnb6vNgeCHFjG60e1BZ9W6BAgDFGoCVhm08qiaiFx225Kr_1vnu42hrTe8zfQCJhI_H"],"id":"PnBb3GPc7IfR-RE4XITuCa6TjwmvJxJ7HnaUITmmwwQJhBKB"}`, string(jsondata))

	// create again
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeCreateStake,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     1,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      from.ID(),
		To:        &stakeid,
		Amount:    new(big.Int).SetUint64(unit.LDC * 1001),
		Data:      input.Bytes(),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	ltx.Timestamp = cs.Timestamp()
	itx, err = NewTx(ltx)
	require.NoError(t, err)
	from.Add(ids.NativeToken, new(big.Int).SetUint64(unit.LDC*1001))

	cs.CommitAccounts()
	assert.ErrorContains(itx.Apply(ctx, cs),
		"stake account #TEST exists")
	cs.CheckoutAccounts()

	// destroy and create again
	ctx.timestamp += 1001
	stakeAcc.LD().Timestamp = cs.Timestamp()
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeDestroyStake,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      stakeid,
		To:        from.ID().Ptr(),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	ltx.Timestamp = cs.Timestamp()
	itx, err = NewTx(ltx)
	require.NoError(t, err)

	cs.CommitAccounts()
	assert.ErrorContains(itx.Apply(ctx, cs),
		"TxDestroyStake.Apply: invalid signature for approver")
	cs.CheckoutAccounts()

	assert.NoError(ltx.SignWith(signer.Signer1, signer.Signer2))
	assert.NoError(ltx.SyntacticVerify())
	ltx.Timestamp = cs.Timestamp()
	itx, err = NewTx(ltx)
	require.NoError(t, err)
	assert.NoError(itx.Apply(ctx, cs))

	assert.Equal(uint64(1), stakeAcc.Nonce())
	assert.Equal(uint16(0), stakeAcc.Threshold())
	assert.Equal(signer.Keys{}, stakeAcc.Keepers())
	assert.Nil(stakeAcc.LD().Approver)
	assert.Equal(ld.NativeAccount, stakeAcc.LD().Type)
	assert.Nil(stakeAcc.LD().MaxTotalSupply)
	assert.Nil(stakeAcc.LD().Stake)
	assert.Equal(0, len(stakeAcc.Ledger().Stake))
	assert.Equal(uint64(0), stakeAcc.BalanceOfAll(ids.NativeToken).Uint64())

	// creat again.
	scfg = &ld.StakeConfig{
		Token:       token,
		LockTime:    0,
		WithdrawFee: 100_000,
		MinAmount:   big.NewInt(100),
		MaxAmount:   big.NewInt(1000),
	}
	input = &ld.TxAccounter{
		Threshold: ld.Uint16Ptr(1),
		Keepers:   &signer.Keys{signer.Signer1.Key()},
		Data:      ld.MustMarshal(scfg),
	}
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeCreateStake,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     1,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      from.ID(),
		To:        &stakeid,
		Amount:    new(big.Int).SetUint64(unit.LDC * 1000),
		Data:      input.Bytes(),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	ltx.Timestamp = cs.Timestamp()
	itx, err = NewTx(ltx)
	require.NoError(t, err)
	assert.NoError(itx.Apply(ctx, cs))

	assert.Equal(unit.LDC*0, stakeAcc.Balance().Uint64())
	assert.Equal(unit.LDC*1000, stakeAcc.BalanceOfAll(ids.NativeToken).Uint64())

	assert.Equal(uint64(1), stakeAcc.Nonce())
	assert.Equal(uint16(1), stakeAcc.Threshold())
	assert.Equal(signer.Keys{signer.Signer1.Key()}, stakeAcc.Keepers())
	assert.Nil(stakeAcc.LD().Approver)
	assert.Equal(ld.StakeAccount, stakeAcc.LD().Type)
	assert.Nil(stakeAcc.LD().MaxTotalSupply)
	require.NotNil(t, stakeAcc.LD().Stake)
	require.NotNil(t, stakeAcc.Ledger().Stake)
	assert.Nil(stakeAcc.Ledger().Stake[from.ID().AsKey()])
	require.NotNil(t, stakeAcc.LD().Tokens[token.AsKey()])
	assert.Equal(uint64(0), stakeAcc.LD().Tokens[token.AsKey()].Uint64())

	assert.NoError(cs.VerifyState())
}
