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
	stakeid := util.EthID(stake)
	token := ld.MustNewToken("$TEST")

	sender := util.Signer1.Address()
	keeper := util.Signer2.Address()

	txData := &ld.TxData{
		Type:      ld.TypeWithdrawStake,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      sender,
	}
	assert.NoError(txData.SyntacticVerify())
	_, err = NewTx2(txData.ToTransaction())
	assert.ErrorContains(err, "DeriveSigners error: no signature")

	txData = &ld.TxData{
		Type:      ld.TypeWithdrawStake,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      sender,
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx2(txData.ToTransaction())
	assert.ErrorContains(err, "nil to as stake account")

	txData = &ld.TxData{
		Type:      ld.TypeWithdrawStake,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      sender,
		To:        &stakeid,
		Token:     &token,
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx2(txData.ToTransaction())
	assert.ErrorContains(err, "invalid token, should be nil")

	txData = &ld.TxData{
		Type:      ld.TypeWithdrawStake,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      sender,
		To:        &stakeid,
		Amount:    big.NewInt(1),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx2(txData.ToTransaction())
	assert.ErrorContains(err, "invalid amount, should be nil")

	txData = &ld.TxData{
		Type:      ld.TypeWithdrawStake,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      sender,
		To:        &stakeid,
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx2(txData.ToTransaction())
	assert.ErrorContains(err, "invalid data")

	txData = &ld.TxData{
		Type:      ld.TypeWithdrawStake,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      sender,
		To:        &constants.GenesisAccount,
		Data:      []byte("你好👋"),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx2(txData.ToTransaction())
	assert.ErrorContains(err, "invalid stake account 0xFFfFfFffFFfffFFfFFfFFFFFffFFFffffFfFFFfF")

	txData = &ld.TxData{
		Type:      ld.TypeWithdrawStake,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      sender,
		To:        &stakeid,
		Data:      []byte("你好👋"),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx2(txData.ToTransaction())
	assert.ErrorContains(err, "cbor: unexpected following extraneous data")

	input := &ld.TxTransfer{Nonce: 1}
	txData = &ld.TxData{
		Type:      ld.TypeWithdrawStake,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      sender,
		To:        &stakeid,
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx2(txData.ToTransaction())
	assert.ErrorContains(err, "invalid nonce, expected 0, got 1")

	input = &ld.TxTransfer{From: &sender}
	txData = &ld.TxData{
		Type:      ld.TypeWithdrawStake,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      sender,
		To:        &stakeid,
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx2(txData.ToTransaction())
	assert.ErrorContains(err,
		"invalid from, expected nil, got 0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC")

	input = &ld.TxTransfer{To: &keeper}
	txData = &ld.TxData{
		Type:      ld.TypeWithdrawStake,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      sender,
		To:        &stakeid,
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx2(txData.ToTransaction())
	assert.ErrorContains(err,
		"invalid to, expected nil, got 0x44171C37Ff5D7B7bb8dcad5C81f16284A229e641")

	input = &ld.TxTransfer{Amount: nil}
	txData = &ld.TxData{
		Type:      ld.TypeWithdrawStake,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      sender,
		To:        &stakeid,
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx2(txData.ToTransaction())
	assert.ErrorContains(err, "nil amount, expected >= 0")

	input = &ld.TxTransfer{Token: &token, Amount: new(big.Int).SetUint64(constants.LDC)}
	txData = &ld.TxData{
		Type:      ld.TypeWithdrawStake,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      sender,
		To:        &stakeid,
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	tt := txData.ToTransaction()
	itx, err := NewTx2(tt)
	assert.NoError(err)

	cs.CommitAccounts()
	assert.ErrorContains(itx.Apply(ctx, cs),
		"TxWithdrawStake.Apply error: Account(0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC).CheckBalance error: insufficient NativeLDC balance, expected 1032900, got 0")
	cs.CheckoutAccounts()

	senderAcc := cs.MustAccount(sender)
	senderAcc.Add(constants.NativeToken, new(big.Int).SetUint64(constants.LDC))
	cs.CommitAccounts()
	assert.ErrorContains(itx.Apply(ctx, cs),
		"TxWithdrawStake.Apply error: Account(0x0000000000000000000000000000002354455354).WithdrawStake error: invalid stake account")
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
		Keepers:   &util.EthIDs{util.Signer2.Address()},
		Data:      ld.MustMarshal(scfg),
	}
	txData = &ld.TxData{
		Type:      ld.TypeCreateStake,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      keeper,
		To:        &stakeid,
		Amount:    new(big.Int).Set(ctx.FeeConfig().MinStakePledge),
		Data:      sinput.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer2))
	tt = txData.ToTransaction()
	tt.Timestamp = cs.Timestamp()
	itx, err = NewTx2(tt)
	assert.NoError(err)

	keeperAcc := cs.MustAccount(keeper)
	keeperAcc.Add(constants.NativeToken,
		new(big.Int).SetUint64(ctx.FeeConfig().MinStakePledge.Uint64()+constants.LDC))
	assert.NoError(itx.Apply(ctx, cs))

	keeperGas := tt.Gas()
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
	txData = &ld.TxData{
		Type:      ld.TypeWithdrawStake,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      sender,
		To:        &stakeid,
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	tt = txData.ToTransaction()
	itx, err = NewTx2(tt)
	cs.CommitAccounts()
	assert.ErrorContains(itx.Apply(ctx, cs),
		"TxWithdrawStake.Apply error: Account(0x0000000000000000000000000000002354455354).WithdrawStake error: invalid token, expected $TEST, got NativeLDC")

	ctx.timestamp = cs.Timestamp() + 1
	cs.CheckoutAccounts()

	input = &ld.TxTransfer{Token: &token, Amount: new(big.Int).SetUint64(constants.LDC)}
	txData = &ld.TxData{
		Type:      ld.TypeWithdrawStake,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      sender,
		To:        &stakeid,
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	tt = txData.ToTransaction()
	tt.Timestamp = cs.Timestamp()
	itx, err = NewTx2(tt)
	cs.CommitAccounts()
	assert.ErrorContains(itx.Apply(ctx, cs),
		"TxWithdrawStake.Apply error: Account(0x0000000000000000000000000000002354455354).WithdrawStake error: stake in lock, please retry after lockTime")

	ctx.timestamp += 1
	cs.CheckoutAccounts()
	input = &ld.TxTransfer{Token: &token, Amount: new(big.Int).SetUint64(constants.LDC)}
	txData = &ld.TxData{
		Type:      ld.TypeWithdrawStake,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      sender,
		To:        &stakeid,
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	tt = txData.ToTransaction()
	tt.Timestamp = cs.Timestamp()
	itx, err = NewTx2(tt)
	cs.CommitAccounts()
	assert.ErrorContains(itx.Apply(ctx, cs),
		"TxWithdrawStake.Apply error: Account(0x0000000000000000000000000000002354455354).WithdrawStake error: 0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC has no stake to withdraw")

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
	txData = &ld.TxData{
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
	}
	assert.NoError(txData.SignWith(util.Signer1))
	assert.NoError(txData.ExSignWith(util.Signer2))
	tt = txData.ToTransaction()
	tt.Timestamp = cs.Timestamp()
	itx, err = NewTx2(tt)
	assert.NoError(err)

	senderAcc.Add(token, new(big.Int).SetUint64(constants.LDC*10))
	assert.NoError(itx.Apply(ctx, cs))

	senderGas := tt.Gas()
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
	input2 := &ld.TxAccounter{Approver: &keeper}
	txData = &ld.TxData{
		Type:      ld.TypeUpdateStakeApprover,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     1,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      sender,
		To:        &stakeid,
		Data:      input2.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	tt = txData.ToTransaction()
	itx, err = NewTx2(tt)
	assert.NoError(err)
	assert.NoError(itx.Apply(ctx, cs))

	senderGas += tt.Gas()
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
	assert.Equal(keeper, *senderEntry.Approver)

	input = &ld.TxTransfer{Token: &token, Amount: new(big.Int).SetUint64(constants.LDC)}
	txData = &ld.TxData{
		Type:      ld.TypeWithdrawStake,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     2,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      sender,
		To:        &stakeid,
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	tt = txData.ToTransaction()
	tt.Timestamp = cs.Timestamp()
	itx, err = NewTx2(tt)
	assert.NoError(err)
	cs.CommitAccounts()
	assert.ErrorContains(itx.Apply(ctx, cs),
		"TxWithdrawStake.Apply error: Account(0x0000000000000000000000000000002354455354).WithdrawStake error: stake in lock, please retry after lockTime, Unix(1002)")

	ctx.timestamp += 1
	cs.CheckoutAccounts()
	input = &ld.TxTransfer{Token: &token, Amount: new(big.Int).SetUint64(constants.LDC)}
	txData = &ld.TxData{
		Type:      ld.TypeWithdrawStake,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     2,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      sender,
		To:        &stakeid,
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	tt = txData.ToTransaction()
	tt.Timestamp = cs.Timestamp()
	itx, err = NewTx2(tt)
	assert.NoError(err)
	cs.CommitAccounts()
	assert.ErrorContains(itx.Apply(ctx, cs),
		"TxWithdrawStake.Apply error: Account(0x0000000000000000000000000000002354455354).WithdrawStake error: 0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC need approver signing")

	cs.CheckoutAccounts()
	input = &ld.TxTransfer{Token: &token, Amount: new(big.Int).SetUint64(constants.LDC * 20)}
	txData = &ld.TxData{
		Type:      ld.TypeWithdrawStake,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     2,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      sender,
		To:        &stakeid,
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	assert.NoError(txData.SignWith(util.Signer2))
	tt = txData.ToTransaction()
	tt.Timestamp = cs.Timestamp()
	itx, err = NewTx2(tt)
	assert.NoError(err)
	cs.CommitAccounts()
	assert.ErrorContains(itx.Apply(ctx, cs),
		"TxWithdrawStake.Apply error: Account(0x0000000000000000000000000000002354455354).WithdrawStake error: 0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC has an insufficient stake to withdraw, expected 10000000000, got 20000000000")

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
	txData = &ld.TxData{
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
	}
	assert.NoError(txData.SignWith(util.Signer2))
	assert.NoError(txData.ExSignWith(util.Signer2))
	tt = txData.ToTransaction()
	tt.Timestamp = cs.Timestamp()
	itx, err = NewTx2(tt)
	assert.NoError(err)

	keeperAcc.Add(token, new(big.Int).SetUint64(constants.LDC*10))
	assert.NoError(itx.Apply(ctx, cs))

	keeperGas += tt.Gas()
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
	txData = &ld.TxData{
		Type:      ld.TypeWithdrawStake,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     2,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      sender,
		To:        &stakeid,
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	assert.NoError(txData.SignWith(util.Signer2))
	tt = txData.ToTransaction()
	tt.Timestamp = cs.Timestamp()
	itx, err = NewTx2(tt)
	assert.NoError(err)
	cs.CommitAccounts()
	assert.ErrorContains(itx.Apply(ctx, cs),
		"TxWithdrawStake.Apply error: Account(0x0000000000000000000000000000002354455354).WithdrawStake error: insufficient $TEST balance for withdraw, expected 20000000000, got 15000000000")

	cs.CheckoutAccounts()
	input = &ld.TxTransfer{
		Nonce:  2,
		From:   &keeper,
		To:     &stakeid,
		Token:  &token,
		Amount: new(big.Int).SetUint64(constants.LDC * 5),
		Expire: cs.Timestamp(),
	}
	txData = &ld.TxData{
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
	}
	assert.NoError(txData.SignWith(util.Signer2))
	assert.NoError(txData.ExSignWith(util.Signer2))
	tt = txData.ToTransaction()
	tt.Timestamp = cs.Timestamp()
	itx, err = NewTx2(tt)
	assert.NoError(err)
	assert.NoError(itx.Apply(ctx, cs))

	keeperGas += tt.Gas()
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
	txData = &ld.TxData{
		Type:      ld.TypeWithdrawStake,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     2,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      sender,
		To:        &stakeid,
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	assert.NoError(txData.SignWith(util.Signer2))
	tt = txData.ToTransaction()
	tt.Timestamp = cs.Timestamp()
	itx, err = NewTx2(tt)
	assert.NoError(err)
	assert.NoError(itx.Apply(ctx, cs))

	senderGas += tt.Gas()
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
	assert.Equal(`{"type":"TypeWithdrawStake","chainID":2357,"nonce":2,"gasTip":100,"gasFeeCap":1000,"from":"0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC","to":"0x0000000000000000000000000000002354455354","data":{"token":"$TEST","amount":20000000000},"signatures":["182a2ebf98042157bb3120f8be02b9cc231fb3d5777f0b919f07d022fb179da609e343e7e0e2491bd10f056bc4c1baf7f73b5790c37600cd8e6767761160e9a201","ce64723d37705370e54e86091fc30b59cf8ba8bacbd8358daa54eee8d57c25114b774c02b77fb292ad9b728f3ced3ce32db7ed35d7da5aa2fc819c6f80a79fe000"],"id":"2NcanD819AXwpLLkdWgr4otjbWWCcAGfFH7K4SferyRGYyNMhY"}`, string(jsondata))

	// clear up sender' stake entry when no stake and approver.
	stakeAcc.ledger.Stake[sender.AsKey()].Approver = nil
	input = &ld.TxTransfer{Token: &token, Amount: new(big.Int).SetUint64(0)}
	txData = &ld.TxData{
		Type:      ld.TypeWithdrawStake,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     3,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      sender,
		To:        &stakeid,
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	tt = txData.ToTransaction()
	tt.Timestamp = cs.Timestamp()
	itx, err = NewTx2(tt)
	assert.NoError(err)
	assert.NoError(itx.Apply(ctx, cs))

	senderGas += tt.Gas()
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
	txData = &ld.TxData{
		Type:      ld.TypeWithdrawStake,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     3,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      keeper,
		To:        &stakeid,
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer2))
	tt = txData.ToTransaction()
	tt.Timestamp = cs.Timestamp()
	itx, err = NewTx2(tt)
	assert.NoError(err)
	assert.NoError(itx.Apply(ctx, cs))

	keeperGas += tt.Gas()
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
