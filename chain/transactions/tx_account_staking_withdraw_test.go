// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package transactions

import (
	"math/big"
	"testing"

	"github.com/ldclabs/ldvm/constants"
	"github.com/ldclabs/ldvm/ld"
	"github.com/ldclabs/ldvm/util"
	"github.com/ldclabs/ldvm/util/signer"
	"github.com/stretchr/testify/assert"
)

func TestTxWithdrawStake(t *testing.T) {
	assert := assert.New(t)

	// SyntacticVerify
	tx := &TxWithdrawStake{}
	assert.ErrorContains(tx.SyntacticVerify(), "nil pointer")
	_, err := tx.MarshalJSON()
	assert.NoError(err)

	ctx := NewMockChainContext()
	cs := ctx.MockChainState()
	stake := ld.MustNewStake("#TEST")
	stakeid := util.Address(stake)
	token := ld.MustNewToken("$TEST")

	sender := signer.Signer1.Key().Address()
	keeper := signer.Signer2.Key().Address()

	ltx := &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeWithdrawStake,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      sender,
	}}
	assert.NoError(ltx.SyntacticVerify())
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "no signatures")

	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeWithdrawStake,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      sender,
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "nil to as stake account")

	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeWithdrawStake,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      sender,
		To:        &stakeid,
		Token:     &token,
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "invalid token, should be nil")

	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeWithdrawStake,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      sender,
		To:        &stakeid,
		Amount:    big.NewInt(1),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "invalid amount, should be nil")

	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeWithdrawStake,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      sender,
		To:        &stakeid,
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "invalid data")

	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeWithdrawStake,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      sender,
		To:        &constants.GenesisAccount,
		Data:      []byte("你好👋"),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "invalid stake account 0xFFfFFFfFfffFFfFFffFFFfFfFffFFFfffFfFFFff")

	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeWithdrawStake,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      sender,
		To:        &stakeid,
		Data:      []byte("你好👋"),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "cbor: unexpected following extraneous data")

	input := &ld.TxTransfer{Nonce: 1}
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeWithdrawStake,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      sender,
		To:        &stakeid,
		Data:      input.Bytes(),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "invalid nonce, expected 0, got 1")

	input = &ld.TxTransfer{From: &sender}
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeWithdrawStake,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      sender,
		To:        &stakeid,
		Data:      input.Bytes(),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err,
		"invalid from, expected nil, got 0x8db97c7cECe249C2b98bdc0226cc4C2A57bF52fc")

	input = &ld.TxTransfer{To: &keeper}
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeWithdrawStake,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      sender,
		To:        &stakeid,
		Data:      input.Bytes(),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err,
		"invalid to, expected nil, got 0x44171C37Ff5D7B7bb8Dcad5C81f16284A229E641")

	input = &ld.TxTransfer{Amount: nil}
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeWithdrawStake,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      sender,
		To:        &stakeid,
		Data:      input.Bytes(),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "nil amount, expected >= 0")

	input = &ld.TxTransfer{Token: &token, Amount: new(big.Int).SetUint64(constants.LDC)}
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeWithdrawStake,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      sender,
		To:        &stakeid,
		Data:      input.Bytes(),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	itx, err := NewTx(ltx)
	assert.NoError(err)

	cs.CommitAccounts()
	assert.ErrorContains(itx.Apply(ctx, cs),
		"TxWithdrawStake.Apply: Account(0x8db97c7cECe249C2b98bdc0226cc4C2A57bF52fc).CheckBalance: insufficient NativeLDC balance, expected 1056000, got 0")
	cs.CheckoutAccounts()

	senderAcc := cs.MustAccount(sender)
	senderAcc.Add(constants.NativeToken, new(big.Int).SetUint64(constants.LDC))
	cs.CommitAccounts()
	assert.ErrorContains(itx.Apply(ctx, cs),
		"TxWithdrawStake.Apply: Account(0x0000000000000000000000000000002354455354).WithdrawStake: invalid stake account")
	cs.CheckoutAccounts()

	// create a new stake account for testing
	scfg := &ld.StakeConfig{
		Token:       token,
		LockTime:    cs.Timestamp() + 1,
		WithdrawFee: 100_000,
		MinAmount:   new(big.Int).SetUint64(constants.LDC * 1),
		MaxAmount:   new(big.Int).SetUint64(constants.LDC * 10),
	}
	sinput := &ld.TxAccounter{
		Threshold: ld.Uint16Ptr(1),
		Keepers:   &signer.Keys{signer.Signer2.Key()},
		Data:      ld.MustMarshal(scfg),
	}
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeCreateStake,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      keeper,
		To:        &stakeid,
		Amount:    new(big.Int).Set(ctx.FeeConfig().MinStakePledge),
		Data:      sinput.Bytes(),
	}}
	assert.NoError(ltx.SignWith(signer.Signer2))
	assert.NoError(ltx.SyntacticVerify())
	ltx.Timestamp = cs.Timestamp()
	itx, err = NewTx(ltx)
	assert.NoError(err)

	keeperAcc := cs.MustAccount(keeper)
	keeperAcc.Add(constants.NativeToken,
		new(big.Int).SetUint64(ctx.FeeConfig().MinStakePledge.Uint64()+constants.LDC))
	assert.NoError(itx.Apply(ctx, cs))

	keeperGas := ltx.Gas()
	stakeAcc := cs.MustAccount(stakeid)
	assert.Equal(keeperGas*ctx.Price,
		itx.(*TxCreateStake).ldc.Balance().Uint64())
	assert.Equal(keeperGas*100,
		itx.(*TxCreateStake).miner.Balance().Uint64())
	assert.Equal(constants.LDC*0, stakeAcc.Balance().Uint64())
	assert.Equal(ctx.FeeConfig().MinStakePledge.Uint64(),
		stakeAcc.balanceOfAll(constants.NativeToken).Uint64())
	assert.Equal(constants.LDC-keeperGas*(ctx.Price+100),
		keeperAcc.Balance().Uint64())

	assert.Nil(stakeAcc.ld.Approver)
	assert.Equal(ld.StakeAccount, stakeAcc.ld.Type)
	assert.Nil(stakeAcc.ld.MaxTotalSupply)
	assert.NotNil(stakeAcc.ld.Stake)
	assert.NotNil(stakeAcc.ledger)
	assert.Nil(stakeAcc.ledger.Stake[keeper.AsKey()])

	input = &ld.TxTransfer{Amount: new(big.Int).SetUint64(constants.LDC)}
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeWithdrawStake,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      sender,
		To:        &stakeid,
		Data:      input.Bytes(),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	itx, err = NewTx(ltx)
	assert.NoError(err)
	cs.CommitAccounts()
	assert.ErrorContains(itx.Apply(ctx, cs),
		"TxWithdrawStake.Apply: Account(0x0000000000000000000000000000002354455354).WithdrawStake: invalid token, expected $TEST, got NativeLDC")

	ctx.timestamp = cs.Timestamp() + 1
	cs.CheckoutAccounts()

	input = &ld.TxTransfer{Token: &token, Amount: new(big.Int).SetUint64(constants.LDC)}
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeWithdrawStake,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      sender,
		To:        &stakeid,
		Data:      input.Bytes(),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	ltx.Timestamp = cs.Timestamp()
	itx, err = NewTx(ltx)
	assert.NoError(err)
	cs.CommitAccounts()
	assert.ErrorContains(itx.Apply(ctx, cs),
		"TxWithdrawStake.Apply: Account(0x0000000000000000000000000000002354455354).WithdrawStake: stake in lock, please retry after lockTime")

	ctx.timestamp += 1
	cs.CheckoutAccounts()
	input = &ld.TxTransfer{Token: &token, Amount: new(big.Int).SetUint64(constants.LDC)}
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeWithdrawStake,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      sender,
		To:        &stakeid,
		Data:      input.Bytes(),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	ltx.Timestamp = cs.Timestamp()
	itx, err = NewTx(ltx)
	assert.NoError(err)
	cs.CommitAccounts()
	assert.ErrorContains(itx.Apply(ctx, cs),
		"TxWithdrawStake.Apply: Account(0x0000000000000000000000000000002354455354).WithdrawStake: 0x8db97c7cECe249C2b98bdc0226cc4C2A57bF52fc has no stake to withdraw")

	// take a stake for testing
	input = &ld.TxTransfer{
		Nonce:  0,
		From:   &sender,
		To:     &stakeid,
		Token:  &token,
		Amount: new(big.Int).SetUint64(constants.LDC * 10),
		Expire: cs.Timestamp(),
		Data:   util.MustMarshalCBOR(cs.Timestamp()),
	}
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeTakeStake,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      sender,
		To:        &stakeid,
		Token:     &token,
		Amount:    new(big.Int).SetUint64(constants.LDC * 10),
		Data:      input.Bytes(),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.ExSignWith(signer.Signer2))
	assert.NoError(ltx.SyntacticVerify())
	ltx.Timestamp = cs.Timestamp()
	itx, err = NewTx(ltx)
	assert.NoError(err)

	senderAcc.Add(token, new(big.Int).SetUint64(constants.LDC*10))
	assert.NoError(itx.Apply(ctx, cs))

	senderGas := ltx.Gas()
	assert.Equal((keeperGas+senderGas)*ctx.Price,
		itx.(*TxTakeStake).ldc.Balance().Uint64())
	assert.Equal((keeperGas+senderGas)*100,
		itx.(*TxTakeStake).miner.Balance().Uint64())
	assert.Equal(constants.LDC*10, stakeAcc.balanceOf(token).Uint64())
	assert.Equal(ctx.FeeConfig().MinStakePledge.Uint64(),
		stakeAcc.balanceOfAll(constants.NativeToken).Uint64())
	assert.Equal(constants.LDC-senderGas*(ctx.Price+100),
		senderAcc.Balance().Uint64())
	assert.Equal(constants.LDC*0, senderAcc.balanceOf(token).Uint64())
	senderEntry := stakeAcc.ledger.Stake[sender.AsKey()]
	assert.NotNil(senderEntry)
	assert.Equal(constants.LDC*10, senderEntry.Amount.Uint64())
	assert.Equal(cs.Timestamp(), senderEntry.LockTime)
	assert.Nil(senderEntry.Approver)

	// add stake approver for testing
	input2 := &ld.TxAccounter{Approver: signer.Signer2.Key().Ptr()}
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeUpdateStakeApprover,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     1,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      sender,
		To:        &stakeid,
		Data:      input2.Bytes(),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	itx, err = NewTx(ltx)
	assert.NoError(err)
	assert.NoError(itx.Apply(ctx, cs))

	senderGas += ltx.Gas()
	assert.Equal((keeperGas+senderGas)*ctx.Price,
		itx.(*TxUpdateStakeApprover).ldc.Balance().Uint64())
	assert.Equal((keeperGas+senderGas)*100,
		itx.(*TxUpdateStakeApprover).miner.Balance().Uint64())
	assert.Equal(constants.LDC*10, stakeAcc.balanceOf(token).Uint64())
	assert.Equal(ctx.FeeConfig().MinStakePledge.Uint64(),
		stakeAcc.balanceOfAll(constants.NativeToken).Uint64())
	assert.Equal(constants.LDC-senderGas*(ctx.Price+100),
		senderAcc.Balance().Uint64())
	senderEntry = stakeAcc.ledger.Stake[sender.AsKey()]
	assert.NotNil(senderEntry)
	assert.NotNil(senderEntry.Approver)
	assert.Equal(keeper, senderEntry.Approver.Address())

	input = &ld.TxTransfer{Token: &token, Amount: new(big.Int).SetUint64(constants.LDC)}
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeWithdrawStake,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     2,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      sender,
		To:        &stakeid,
		Data:      input.Bytes(),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	ltx.Timestamp = cs.Timestamp()
	itx, err = NewTx(ltx)
	assert.NoError(err)
	cs.CommitAccounts()
	assert.ErrorContains(itx.Apply(ctx, cs),
		"TxWithdrawStake.Apply: Account(0x0000000000000000000000000000002354455354).WithdrawStake: stake in lock, please retry after lockTime, Unix(1002)")

	ctx.timestamp += 1
	cs.CheckoutAccounts()
	input = &ld.TxTransfer{Token: &token, Amount: new(big.Int).SetUint64(constants.LDC)}
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeWithdrawStake,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     2,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      sender,
		To:        &stakeid,
		Data:      input.Bytes(),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	ltx.Timestamp = cs.Timestamp()
	itx, err = NewTx(ltx)
	assert.NoError(err)
	cs.CommitAccounts()
	assert.ErrorContains(itx.Apply(ctx, cs),
		"TxWithdrawStake.Apply: Account(0x0000000000000000000000000000002354455354).WithdrawStake: 0x8db97c7cECe249C2b98bdc0226cc4C2A57bF52fc need approver signing")

	cs.CheckoutAccounts()
	input = &ld.TxTransfer{Token: &token, Amount: new(big.Int).SetUint64(constants.LDC * 20)}
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeWithdrawStake,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     2,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      sender,
		To:        &stakeid,
		Data:      input.Bytes(),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1, signer.Signer2))
	assert.NoError(ltx.SyntacticVerify())
	ltx.Timestamp = cs.Timestamp()
	itx, err = NewTx(ltx)
	assert.NoError(err)
	cs.CommitAccounts()
	assert.ErrorContains(itx.Apply(ctx, cs),
		"TxWithdrawStake.Apply: Account(0x0000000000000000000000000000002354455354).WithdrawStake: 0x8db97c7cECe249C2b98bdc0226cc4C2A57bF52fc has an insufficient stake to withdraw, expected 10000000000, got 20000000000")

	stakeAcc.Add(token, new(big.Int).SetUint64(constants.LDC*10))
	assert.Equal(constants.LDC*20, stakeAcc.balanceOf(token).Uint64())
	assert.Equal(constants.LDC*10, stakeAcc.ledger.Stake[sender.AsKey()].Amount.Uint64())

	// keeper: take a stake for testing
	input = &ld.TxTransfer{
		Nonce:  1,
		From:   &keeper,
		To:     &stakeid,
		Token:  &token,
		Amount: new(big.Int).SetUint64(constants.LDC * 5),
		Expire: cs.Timestamp(),
	}
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeTakeStake,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     1,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      keeper,
		To:        &stakeid,
		Token:     &token,
		Amount:    new(big.Int).SetUint64(constants.LDC * 5),
		Data:      input.Bytes(),
	}}
	assert.NoError(ltx.SignWith(signer.Signer2))
	assert.NoError(ltx.ExSignWith(signer.Signer2))
	assert.NoError(ltx.SyntacticVerify())
	ltx.Timestamp = cs.Timestamp()
	itx, err = NewTx(ltx)
	assert.NoError(err)

	keeperAcc.Add(token, new(big.Int).SetUint64(constants.LDC*10))
	assert.NoError(itx.Apply(ctx, cs))

	keeperGas += ltx.Gas()
	assert.Equal((keeperGas+senderGas)*ctx.Price,
		itx.(*TxTakeStake).ldc.Balance().Uint64())
	assert.Equal((keeperGas+senderGas)*100,
		itx.(*TxTakeStake).miner.Balance().Uint64())
	assert.Equal(ctx.FeeConfig().MinStakePledge.Uint64(),
		stakeAcc.balanceOfAll(constants.NativeToken).Uint64())
	assert.Equal(constants.LDC-senderGas*(ctx.Price+100),
		senderAcc.Balance().Uint64())
	assert.Equal(constants.LDC*0, senderAcc.balanceOf(token).Uint64())
	assert.Equal(constants.LDC*5, keeperAcc.balanceOf(token).Uint64())
	assert.Equal(constants.LDC*25, stakeAcc.balanceOf(token).Uint64())
	assert.Equal(constants.LDC*20, stakeAcc.ledger.Stake[sender.AsKey()].Amount.Uint64())
	assert.Equal(constants.LDC*5, stakeAcc.ledger.Stake[keeper.AsKey()].Amount.Uint64())

	stakeAcc.Sub(token, new(big.Int).SetUint64(constants.LDC*10))
	input = &ld.TxTransfer{Token: &token, Amount: new(big.Int).SetUint64(constants.LDC * 20)}
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeWithdrawStake,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     2,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      sender,
		To:        &stakeid,
		Data:      input.Bytes(),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1, signer.Signer2))
	assert.NoError(ltx.SyntacticVerify())
	ltx.Timestamp = cs.Timestamp()
	itx, err = NewTx(ltx)
	assert.NoError(err)
	cs.CommitAccounts()
	assert.ErrorContains(itx.Apply(ctx, cs),
		"TxWithdrawStake.Apply: Account(0x0000000000000000000000000000002354455354).WithdrawStake: insufficient $TEST balance for withdraw, expected 20000000000, got 15000000000")

	cs.CheckoutAccounts()
	input = &ld.TxTransfer{
		Nonce:  2,
		From:   &keeper,
		To:     &stakeid,
		Token:  &token,
		Amount: new(big.Int).SetUint64(constants.LDC * 5),
		Expire: cs.Timestamp(),
	}
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeTakeStake,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     2,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      keeper,
		To:        &stakeid,
		Token:     &token,
		Amount:    new(big.Int).SetUint64(constants.LDC * 5),
		Data:      input.Bytes(),
	}}
	assert.NoError(ltx.SignWith(signer.Signer2))
	assert.NoError(ltx.ExSignWith(signer.Signer2))
	assert.NoError(ltx.SyntacticVerify())
	ltx.Timestamp = cs.Timestamp()
	itx, err = NewTx(ltx)
	assert.NoError(err)
	assert.NoError(itx.Apply(ctx, cs))

	keeperGas += ltx.Gas()
	assert.Equal((keeperGas+senderGas)*ctx.Price,
		itx.(*TxTakeStake).ldc.Balance().Uint64())
	assert.Equal((keeperGas+senderGas)*100,
		itx.(*TxTakeStake).miner.Balance().Uint64())
	assert.Equal(ctx.FeeConfig().MinStakePledge.Uint64(),
		stakeAcc.balanceOfAll(constants.NativeToken).Uint64())
	assert.Equal(constants.LDC-senderGas*(ctx.Price+100),
		senderAcc.Balance().Uint64())
	assert.Equal(constants.LDC*0, senderAcc.balanceOf(token).Uint64())
	assert.Equal(constants.LDC*0, keeperAcc.balanceOf(token).Uint64())
	assert.Equal(constants.LDC*20, stakeAcc.balanceOf(token).Uint64())
	assert.Equal(constants.LDC*20, stakeAcc.ledger.Stake[sender.AsKey()].Amount.Uint64())
	assert.Equal(constants.LDC*10, stakeAcc.ledger.Stake[keeper.AsKey()].Amount.Uint64())

	input = &ld.TxTransfer{Token: &token, Amount: new(big.Int).SetUint64(constants.LDC * 20)}
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeWithdrawStake,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     2,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      sender,
		To:        &stakeid,
		Data:      input.Bytes(),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1, signer.Signer2))
	assert.NoError(ltx.SyntacticVerify())
	ltx.Timestamp = cs.Timestamp()
	itx, err = NewTx(ltx)
	assert.NoError(err)
	assert.NoError(itx.Apply(ctx, cs))

	senderGas += ltx.Gas()
	assert.Equal((keeperGas+senderGas)*ctx.Price,
		itx.(*TxWithdrawStake).ldc.Balance().Uint64())
	assert.Equal((keeperGas+senderGas)*100,
		itx.(*TxWithdrawStake).miner.Balance().Uint64())
	assert.Equal(ctx.FeeConfig().MinStakePledge.Uint64(),
		stakeAcc.balanceOfAll(constants.NativeToken).Uint64())
	assert.Equal(constants.LDC-senderGas*(ctx.Price+100),
		senderAcc.Balance().Uint64())

	withdrawFee := constants.LDC * 20 * scfg.WithdrawFee / 1_000_000
	assert.Equal(constants.LDC*20-withdrawFee, senderAcc.balanceOf(token).Uint64())
	assert.Equal(constants.LDC*0, keeperAcc.balanceOf(token).Uint64())
	assert.Equal(withdrawFee, stakeAcc.balanceOf(token).Uint64())
	assert.NotNil(stakeAcc.ledger.Stake[sender.AsKey()])
	assert.Equal(constants.LDC*0, stakeAcc.ledger.Stake[sender.AsKey()].Amount.Uint64())
	assert.Equal(constants.LDC*10, stakeAcc.ledger.Stake[keeper.AsKey()].Amount.Uint64())

	jsondata, err := itx.MarshalJSON()
	assert.NoError(err)
	// fmt.Println(string(jsondata))
	assert.Equal(`{"tx":{"type":"TypeWithdrawStake","chainID":2357,"nonce":2,"gasTip":100,"gasFeeCap":1000,"from":"0x8db97c7cECe249C2b98bdc0226cc4C2A57bF52fc","to":"0x0000000000000000000000000000002354455354","data":{"token":"$TEST","amount":20000000000}},"sigs":["GCouv5gEIVe7MSD4vgK5zCMfs9V3fwuRnwfQIvsXnaYJ40Pn4OJJG9EPBWvEwbr39ztXkMN2AM2OZ2d2EWDpogHNKD5a","zmRyPTdwU3DlToYJH8MLWc-LqLrL2DWNqlTu6NV8JRFLd0wCt3-ykq2bco887TzjLbftNdfaWqL8gZxvgKef4ABl0GM0"],"id":"r9xN8GGi48Rn_lfcDhl9EBBtCV6h3Ds2tEDnqg1ZKhKd1ElX"}`, string(jsondata))

	// clear up sender' stake entry when no stake and approver.
	stakeAcc.ledger.Stake[sender.AsKey()].Approver = nil
	input = &ld.TxTransfer{Token: &token, Amount: new(big.Int).SetUint64(0)}
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeWithdrawStake,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     3,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      sender,
		To:        &stakeid,
		Data:      input.Bytes(),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	ltx.Timestamp = cs.Timestamp()
	itx, err = NewTx(ltx)
	assert.NoError(err)
	assert.NoError(itx.Apply(ctx, cs))

	senderGas += ltx.Gas()
	assert.Equal((keeperGas+senderGas)*ctx.Price,
		itx.(*TxWithdrawStake).ldc.Balance().Uint64())
	assert.Equal((keeperGas+senderGas)*100,
		itx.(*TxWithdrawStake).miner.Balance().Uint64())
	assert.Equal(ctx.FeeConfig().MinStakePledge.Uint64(),
		stakeAcc.balanceOfAll(constants.NativeToken).Uint64())
	assert.Equal(constants.LDC-senderGas*(ctx.Price+100),
		senderAcc.Balance().Uint64())

	assert.Equal(withdrawFee, stakeAcc.balanceOf(token).Uint64())
	assert.Nil(stakeAcc.ledger.Stake[sender.AsKey()])
	assert.Equal(constants.LDC*10, stakeAcc.ledger.Stake[keeper.AsKey()].Amount.Uint64())

	// keeper: withdraw all stake
	stakeAcc.Add(token, new(big.Int).SetUint64(constants.LDC*20-withdrawFee))
	assert.Equal(constants.LDC*20, stakeAcc.balanceOf(token).Uint64())
	input = &ld.TxTransfer{Token: &token, Amount: new(big.Int).SetUint64(constants.LDC * 20)}
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeWithdrawStake,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     3,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      keeper,
		To:        &stakeid,
		Data:      input.Bytes(),
	}}
	assert.NoError(ltx.SignWith(signer.Signer2))
	assert.NoError(ltx.SyntacticVerify())
	ltx.Timestamp = cs.Timestamp()
	itx, err = NewTx(ltx)
	assert.NoError(err)
	assert.NoError(itx.Apply(ctx, cs))

	keeperGas += ltx.Gas()
	assert.Equal((keeperGas+senderGas)*ctx.Price,
		itx.(*TxWithdrawStake).ldc.Balance().Uint64())
	assert.Equal((keeperGas+senderGas)*100,
		itx.(*TxWithdrawStake).miner.Balance().Uint64())
	assert.Equal(ctx.FeeConfig().MinStakePledge.Uint64(),
		stakeAcc.balanceOfAll(constants.NativeToken).Uint64())

	withdrawFee = constants.LDC * 20 * scfg.WithdrawFee / 1_000_000
	assert.Equal(constants.LDC*20-withdrawFee, keeperAcc.balanceOf(token).Uint64())
	assert.Equal(withdrawFee, stakeAcc.balanceOf(token).Uint64())
	assert.Nil(stakeAcc.ledger.Stake[keeper.AsKey()])
	assert.Equal(0, len(stakeAcc.ledger.Stake))

	assert.NoError(cs.VerifyState())
}
