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
	"github.com/stretchr/testify/require"
)

func TestTxDestroyStake(t *testing.T) {
	assert := assert.New(t)

	// SyntacticVerify
	tx := &TxDestroyStake{}
	assert.ErrorContains(tx.SyntacticVerify(), "nil pointer")
	_, err := tx.MarshalJSON()
	require.NoError(t, err)

	ctx := NewMockChainContext()
	cs := ctx.MockChainState()
	stake := ld.MustNewStake("#TEST")
	stakeid := util.Address(stake)
	token := ld.MustNewToken("$TEST")
	sender := signer.Signer1.Key().Address()
	approver := signer.Signer2.Key()
	keeper := approver.Address()

	ltx := &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeDestroyStake,
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
		Type:      ld.TypeDestroyStake,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      sender,
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "nil to as pledge recipient")

	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeDestroyStake,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      sender,
		To:        &keeper,
		Token:     token.Ptr(),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "invalid token, should be nil")

	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeDestroyStake,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      sender,
		To:        &keeper,
		Amount:    big.NewInt(1),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "invalid amount, should be nil")

	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeDestroyStake,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      sender,
		To:        &keeper,
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err,
		"TxDestroyStake.SyntacticVerify: invalid stake account 0x8db97c7cECe249C2b98bdc0226cc4C2A57bF52fc")

	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeDestroyStake,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      stakeid,
		To:        &keeper,
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	ltx.Timestamp = cs.Timestamp()
	itx, err := NewTx(ltx)
	require.NoError(t, err)

	cs.CommitAccounts()
	assert.ErrorContains(itx.Apply(ctx, cs),
		"TxDestroyStake.Apply: invalid signatures for sender")
	cs.CheckoutAccounts()

	// create a stake account for testing
	scfg := &ld.StakeConfig{
		LockTime:    cs.Timestamp() + 100,
		WithdrawFee: 100_000,
		MinAmount:   new(big.Int).SetUint64(constants.LDC * 10),
		MaxAmount:   new(big.Int).SetUint64(constants.LDC * 100),
	}
	sinput := &ld.TxAccounter{
		Threshold: ld.Uint16Ptr(1),
		Keepers:   &signer.Keys{signer.Signer1.Key(), signer.Signer2.Key()},
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
	require.NoError(t, err)

	keeperAcc := cs.MustAccount(keeper)
	keeperAcc.Add(constants.NativeToken,
		new(big.Int).SetUint64(ctx.FeeConfig().MinStakePledge.Uint64()+constants.LDC*2))
	assert.NoError(itx.Apply(ctx, cs))

	keeperGas := ltx.Gas()
	stakeAcc := cs.MustAccount(stakeid)
	assert.Equal(keeperGas*ctx.Price,
		itx.(*TxCreateStake).ldc.Balance().Uint64())
	assert.Equal(keeperGas*100,
		itx.(*TxCreateStake).miner.Balance().Uint64())
	assert.Equal(constants.LDC*0, stakeAcc.Balance().Uint64())
	assert.Equal(ctx.FeeConfig().MinStakePledge.Uint64(),
		stakeAcc.BalanceOfAll(constants.NativeToken).Uint64())
	assert.Equal(constants.LDC-keeperGas*(ctx.Price+100),
		keeperAcc.Balance().Uint64())

	require.NotNil(t, stakeAcc.Ledger())
	keeperEntry := stakeAcc.Ledger().Stake[keeper.AsKey()]
	require.NotNil(t, keeperEntry)
	assert.Equal(ctx.FeeConfig().MinStakePledge.Uint64(), keeperEntry.Amount.Uint64())

	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeDestroyStake,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      stakeid,
		To:        &keeper,
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	ltx.Timestamp = cs.Timestamp()
	itx, err = NewTx(ltx)
	require.NoError(t, err)

	cs.CommitAccounts()
	assert.ErrorContains(itx.Apply(ctx, cs),
		"insufficient NativeLDC balance, expected 1728100, got 0")
	cs.CheckoutAccounts()

	stakeAcc.Add(constants.NativeToken, new(big.Int).SetUint64(constants.LDC))
	cs.CommitAccounts()
	assert.ErrorContains(itx.Apply(ctx, cs),
		"TxDestroyStake.Apply: invalid signatures for stake keepers")
	cs.CheckoutAccounts()

	assert.NoError(ltx.SignWith(signer.Signer1, signer.Signer2))
	assert.NoError(ltx.SyntacticVerify())
	ltx.Timestamp = cs.Timestamp()
	itx, err = NewTx(ltx)
	require.NoError(t, err)
	cs.CommitAccounts()
	assert.ErrorContains(itx.Apply(ctx, cs),
		"stake in lock, please retry after lockTime, Unix(1100)")
	cs.CheckoutAccounts()

	ctx.timestamp += 101
	cs.CheckoutAccounts()
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeDestroyStake,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      stakeid,
		To:        &keeper,
	}}
	assert.NoError(ltx.SignWith(signer.Signer1, signer.Signer2))
	assert.NoError(ltx.SyntacticVerify())
	ltx.Timestamp = cs.Timestamp()
	itx, err = NewTx(ltx)
	require.NoError(t, err)
	cs.CommitAccounts()
	assert.NoError(itx.Apply(ctx, cs))
	cs.CheckoutAccounts()

	// take a stake for testing
	input2 := &ld.TxTransfer{
		Nonce:  0,
		From:   sender.Ptr(),
		To:     &stakeid,
		Amount: new(big.Int).SetUint64(constants.LDC * 10),
		Expire: cs.Timestamp(),
	}
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeTakeStake,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      sender,
		To:        &stakeid,
		Amount:    new(big.Int).SetUint64(constants.LDC * 10),
		Data:      input2.Bytes(),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.ExSignWith(signer.Signer2))
	assert.NoError(ltx.SyntacticVerify())
	ltx.Timestamp = cs.Timestamp()
	itx, err = NewTx(ltx)
	require.NoError(t, err)

	senderAcc := cs.MustAccount(sender)
	senderAcc.Add(constants.NativeToken, new(big.Int).SetUint64(constants.LDC*12))
	assert.NoError(itx.Apply(ctx, cs))

	senderGas := ltx.Gas()
	assert.Equal((keeperGas+senderGas)*ctx.Price,
		itx.(*TxTakeStake).ldc.Balance().Uint64())
	assert.Equal((keeperGas+senderGas)*100,
		itx.(*TxTakeStake).miner.Balance().Uint64())
	assert.Equal(ctx.FeeConfig().MinStakePledge.Uint64()+constants.LDC*11,
		stakeAcc.BalanceOfAll(constants.NativeToken).Uint64())
	assert.Equal(constants.LDC-senderGas*(ctx.Price+100),
		senderAcc.Balance().Uint64())
	senderEntry := stakeAcc.Ledger().Stake[sender.AsKey()]
	require.NotNil(t, senderEntry)
	assert.Equal(constants.LDC*10, senderEntry.Amount.Uint64())
	assert.Equal(uint64(0), senderEntry.LockTime)
	assert.Nil(senderEntry.Approver)

	// add stake approver for testing
	input3 := &ld.TxAccounter{Approver: &approver}
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeUpdateStakeApprover,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     1,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      sender,
		To:        &stakeid,
		Data:      input3.Bytes(),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	itx, err = NewTx(ltx)
	require.NoError(t, err)
	assert.NoError(itx.Apply(ctx, cs))

	senderGas += ltx.Gas()
	assert.Equal((keeperGas+senderGas)*ctx.Price,
		itx.(*TxUpdateStakeApprover).ldc.Balance().Uint64())
	assert.Equal((keeperGas+senderGas)*100,
		itx.(*TxUpdateStakeApprover).miner.Balance().Uint64())
	assert.Equal(constants.LDC-senderGas*(ctx.Price+100),
		senderAcc.Balance().Uint64())
	senderEntry = stakeAcc.Ledger().Stake[sender.AsKey()]
	require.NotNil(t, senderEntry)
	require.NotNil(t, senderEntry.Approver)
	assert.Equal(keeper, senderEntry.Approver.Address())

	// destroy again
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeDestroyStake,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      stakeid,
		To:        &keeper,
	}}
	assert.NoError(ltx.SignWith(signer.Signer1, signer.Signer2))
	assert.NoError(ltx.SyntacticVerify())
	ltx.Timestamp = cs.Timestamp()
	itx, err = NewTx(ltx)
	require.NoError(t, err)
	cs.CommitAccounts()
	assert.ErrorContains(itx.Apply(ctx, cs),
		"stake ledger not empty, please withdraw all except recipient")
	cs.CheckoutAccounts()

	input2 = &ld.TxTransfer{Amount: new(big.Int).SetUint64(constants.LDC * 10)}
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeWithdrawStake,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     2,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      sender,
		To:        &stakeid,
		Data:      input2.Bytes(),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1, signer.Signer2))
	assert.NoError(ltx.SyntacticVerify())
	ltx.Timestamp = cs.Timestamp()
	itx, err = NewTx(ltx)
	require.NoError(t, err)
	assert.NoError(itx.Apply(ctx, cs))

	senderGas += ltx.Gas()
	assert.Equal((keeperGas+senderGas)*ctx.Price,
		itx.(*TxWithdrawStake).ldc.Balance().Uint64())
	assert.Equal((keeperGas+senderGas)*100,
		itx.(*TxWithdrawStake).miner.Balance().Uint64())

	withdrawFee := constants.LDC * 10 * scfg.WithdrawFee / 1_000_000
	assert.Equal(constants.LDC*11-withdrawFee-senderGas*(ctx.Price+100),
		senderAcc.Balance().Uint64())
	assert.Equal(constants.LDC+withdrawFee, stakeAcc.Balance().Uint64())
	require.NotNil(t, stakeAcc.Ledger().Stake[sender.AsKey()])
	assert.Equal(constants.LDC*0, stakeAcc.Ledger().Stake[sender.AsKey()].Amount.Uint64())

	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeDestroyStake,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      stakeid,
		To:        &keeper,
	}}
	assert.NoError(ltx.SignWith(signer.Signer1, signer.Signer2))
	assert.NoError(ltx.SyntacticVerify())
	ltx.Timestamp = cs.Timestamp()
	itx, err = NewTx(ltx)
	require.NoError(t, err)
	assert.NoError(itx.Apply(ctx, cs))

	stakeGas := ltx.Gas()
	assert.Equal((keeperGas+senderGas+stakeGas)*ctx.Price,
		itx.(*TxDestroyStake).ldc.Balance().Uint64())
	assert.Equal((keeperGas+senderGas+stakeGas)*100,
		itx.(*TxDestroyStake).miner.Balance().Uint64())

	assert.Equal(ctx.FeeConfig().MinStakePledge.Uint64()+constants.LDC*2+withdrawFee-(keeperGas+stakeGas)*(ctx.Price+100),
		keeperAcc.Balance().Uint64())
	assert.Equal(ld.AccountType(0), stakeAcc.LD().Type)
	assert.Equal(uint16(0), stakeAcc.LD().Threshold)
	assert.Equal(uint64(1), stakeAcc.LD().Nonce)
	assert.Equal(signer.Keys{}, stakeAcc.LD().Keepers)
	assert.Equal(make(map[uint64][]uint64), stakeAcc.LD().NonceTable)
	assert.Nil(stakeAcc.LD().Approver)
	assert.Nil(stakeAcc.LD().ApproveList)
	assert.Nil(stakeAcc.LD().Stake)
	assert.Equal(0, len(stakeAcc.Ledger().Stake))

	jsondata, err := itx.MarshalJSON()
	require.NoError(t, err)
	// fmt.Println(string(jsondata))
	assert.Equal(`{"tx":{"type":"TypeDestroyStake","chainID":2357,"nonce":0,"gasTip":100,"gasFeeCap":1000,"from":"0x0000000000000000000000000000002354455354","to":"0x44171C37Ff5D7B7bb8Dcad5C81f16284A229E641"},"sigs":["48OVRuaf_QHvxdUNOnZDXnUAMKKBDcxHiqHazqfp0MACRmEv7xe3M4ukta17alZ_t1e2vYczaIZW4aRhtYJ3gQHUm7Pu","bhJRPt29cy0CdFHl8Wn2wgIylCWOgZH70Fi1k9ZFiS5v4-FFCl2AwnvmMp1NHeXf4gQJscEkRyYs--nxfO22WQFWhz36"],"id":"F-OK20ASGhw1N823p7O5ajH15gHb2vFUi4Ow0GVsbkBWNnAw"}`, string(jsondata))

	// create stake account again
	scfg = &ld.StakeConfig{
		Token:       token,
		Type:        2,
		WithdrawFee: 100_000,
		MinAmount:   new(big.Int).SetUint64(constants.LDC * 10),
		MaxAmount:   new(big.Int).SetUint64(constants.LDC * 100),
	}
	sinput = &ld.TxAccounter{
		Threshold: ld.Uint16Ptr(1),
		Keepers:   &signer.Keys{signer.Signer2.Key()},
		Data:      ld.MustMarshal(scfg),
	}
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeCreateStake,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     1,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      keeper,
		To:        &stakeid,
		Amount:    new(big.Int).SetUint64(ctx.FeeConfig().MinStakePledge.Uint64() + constants.LDC),
		Data:      sinput.Bytes(),
	}}
	assert.NoError(ltx.SignWith(signer.Signer2))
	assert.NoError(ltx.SyntacticVerify())
	ltx.Timestamp = cs.Timestamp()
	itx, err = NewTx(ltx)
	require.NoError(t, err)
	assert.NoError(itx.Apply(ctx, cs))

	keeperGas += ltx.Gas()
	assert.Equal((keeperGas+senderGas+stakeGas)*ctx.Price,
		itx.(*TxCreateStake).ldc.Balance().Uint64())
	assert.Equal((keeperGas+senderGas+stakeGas)*100,
		itx.(*TxCreateStake).miner.Balance().Uint64())

	assert.Equal(constants.LDC, stakeAcc.Balance().Uint64())
	assert.Equal(ctx.FeeConfig().MinStakePledge.Uint64()+constants.LDC,
		stakeAcc.BalanceOfAll(constants.NativeToken).Uint64())
	assert.Equal(uint64(0), keeperAcc.BalanceOf(token).Uint64())

	require.NotNil(t, stakeAcc.Ledger().Stake)
	assert.Equal(0, len(stakeAcc.Ledger().Stake))

	stakeAcc.Add(token, new(big.Int).SetUint64(constants.LDC*9))
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeDestroyStake,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     1,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      stakeid,
		To:        &keeper,
	}}
	assert.NoError(ltx.SignWith(signer.Signer2))
	assert.NoError(ltx.SyntacticVerify())
	ltx.Timestamp = cs.Timestamp()
	itx, err = NewTx(ltx)
	require.NoError(t, err)
	assert.NoError(itx.Apply(ctx, cs))

	stakeGas += ltx.Gas()
	assert.Equal((keeperGas+senderGas+stakeGas)*ctx.Price,
		itx.(*TxDestroyStake).ldc.Balance().Uint64())
	assert.Equal((keeperGas+senderGas+stakeGas)*100,
		itx.(*TxDestroyStake).miner.Balance().Uint64())

	assert.Equal(ctx.FeeConfig().MinStakePledge.Uint64()+constants.LDC*2+withdrawFee-(keeperGas+stakeGas)*(ctx.Price+100),
		keeperAcc.Balance().Uint64())
	assert.Equal(constants.LDC*9,
		keeperAcc.BalanceOf(token).Uint64())
	assert.Equal(ld.AccountType(0), stakeAcc.LD().Type)
	assert.Equal(uint16(0), stakeAcc.LD().Threshold)
	assert.Equal(uint64(2), stakeAcc.LD().Nonce)
	assert.Equal(signer.Keys{}, stakeAcc.LD().Keepers)
	assert.Equal(make(map[uint64][]uint64), stakeAcc.LD().NonceTable)
	assert.Nil(stakeAcc.LD().Approver)
	assert.Nil(stakeAcc.LD().ApproveList)
	assert.Nil(stakeAcc.LD().Stake)
	assert.Equal(0, len(stakeAcc.Ledger().Stake))

	assert.NoError(cs.VerifyState())
}

func TestTxDestroyStakeWithApproverAndLending(t *testing.T) {
	assert := assert.New(t)

	ctx := NewMockChainContext()
	cs := ctx.MockChainState()
	stake := ld.MustNewStake("#TEST")
	stakeid := util.Address(stake)
	token := ld.MustNewToken("$TEST")
	approver := signer.Signer1.Key()
	keeper := signer.Signer2.Key().Address()

	scfg := &ld.StakeConfig{
		Token:       token,
		Type:        1,
		WithdrawFee: 100_000,
		MinAmount:   new(big.Int).SetUint64(constants.LDC * 10),
		MaxAmount:   new(big.Int).SetUint64(constants.LDC * 100),
	}
	input := &ld.TxAccounter{
		Threshold:   ld.Uint16Ptr(1),
		Keepers:     &signer.Keys{signer.Signer2.Key()},
		Approver:    &approver,
		ApproveList: &ld.TxTypes{ld.TypeOpenLending, ld.TypeDestroyStake},
		Data:        ld.MustMarshal(scfg),
	}
	ltx := &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeCreateStake,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      keeper,
		To:        &stakeid,
		Amount:    new(big.Int).SetUint64(ctx.FeeConfig().MinStakePledge.Uint64() + constants.LDC),
		Data:      input.Bytes(),
	}}
	assert.NoError(ltx.SignWith(signer.Signer2))
	assert.NoError(ltx.SyntacticVerify())
	ltx.Timestamp = cs.Timestamp()
	itx, err := NewTx(ltx)
	require.NoError(t, err)

	keeperAcc := cs.MustAccount(keeper)
	keeperAcc.Add(constants.NativeToken,
		new(big.Int).SetUint64(ctx.FeeConfig().MinStakePledge.Uint64()+constants.LDC*3))
	assert.NoError(itx.Apply(ctx, cs))

	keeperGas := ltx.Gas()
	stakeAcc := cs.MustAccount(stakeid)
	assert.Equal((keeperGas)*ctx.Price,
		itx.(*TxCreateStake).ldc.Balance().Uint64())
	assert.Equal((keeperGas)*100,
		itx.(*TxCreateStake).miner.Balance().Uint64())

	assert.Equal(constants.LDC, stakeAcc.Balance().Uint64())
	assert.Equal(ctx.FeeConfig().MinStakePledge.Uint64()+constants.LDC,
		stakeAcc.BalanceOfAll(constants.NativeToken).Uint64())
	assert.Equal(uint64(0), stakeAcc.BalanceOf(token).Uint64())

	require.NotNil(t, stakeAcc.Ledger().Stake)
	assert.Equal(0, len(stakeAcc.Ledger().Stake))
	require.NotNil(t, stakeAcc.LD().Approver)
	assert.Equal(approver.Address(), stakeAcc.LD().Approver.Address())
	assert.Equal(ld.TxTypes{ld.TypeOpenLending, ld.TypeDestroyStake}, stakeAcc.LD().ApproveList)
	stakeAcc.Add(token, new(big.Int).SetUint64(constants.LDC*10))

	// OpenLending
	lcfg := &ld.LendingConfig{
		Token:           token,
		DailyInterest:   10,
		OverdueInterest: 10,
		MinAmount:       big.NewInt(1000),
		MaxAmount:       new(big.Int).SetUint64(constants.LDC),
	}
	assert.NoError(lcfg.SyntacticVerify())
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeOpenLending,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      stakeid,
		Data:      ld.MustMarshal(lcfg),
	}}
	assert.NoError(ltx.SignWith(signer.Signer2))
	assert.NoError(ltx.SyntacticVerify())
	ltx.Timestamp = cs.Timestamp()
	itx, err = NewTx(ltx)
	require.NoError(t, err)

	cs.CommitAccounts()
	assert.ErrorContains(itx.Apply(ctx, cs),
		"TxOpenLending.Apply: invalid signature for approver")
	cs.CheckoutAccounts()

	assert.NoError(ltx.SignWith(signer.Signer1, signer.Signer2))
	assert.NoError(ltx.SyntacticVerify())
	ltx.Timestamp = cs.Timestamp()
	itx, err = NewTx(ltx)
	require.NoError(t, err)
	assert.NoError(itx.Apply(ctx, cs))

	stakeGas := ltx.Gas()
	assert.Equal((keeperGas+stakeGas)*ctx.Price,
		itx.(*TxOpenLending).ldc.Balance().Uint64())
	assert.Equal((keeperGas+stakeGas)*100,
		itx.(*TxOpenLending).miner.Balance().Uint64())
	require.NotNil(t, stakeAcc.LD().Lending)
	assert.Equal(uint64(1), stakeAcc.Nonce())

	// UpdateNonceTable
	ns := []uint64{cs.Timestamp() + 1, 1, 2, 3}
	ndData, err := util.MarshalCBOR(ns)
	require.NoError(t, err)
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeUpdateNonceTable,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     1,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      stakeid,
		Data:      ndData,
	}}
	assert.NoError(ltx.SignWith(signer.Signer2))
	assert.NoError(ltx.SyntacticVerify())
	ltx.Timestamp = cs.Timestamp()
	itx, err = NewTx(ltx)
	require.NoError(t, err)
	assert.NoError(itx.Apply(ctx, cs))

	stakeGas += ltx.Gas()
	assert.Equal((keeperGas+stakeGas)*ctx.Price,
		itx.(*TxUpdateNonceTable).ldc.Balance().Uint64())
	assert.Equal((keeperGas+stakeGas)*100,
		itx.(*TxUpdateNonceTable).miner.Balance().Uint64())
	assert.Equal([]uint64{1, 2, 3}, stakeAcc.LD().NonceTable[cs.Timestamp()+1])
	assert.Equal(uint64(2), stakeAcc.Nonce())

	// Borrow
	approverAddr := approver.Address()
	tf := &ld.TxTransfer{
		Nonce:  3,
		From:   &stakeid,
		To:     &approverAddr,
		Token:  token.Ptr(),
		Amount: new(big.Int).SetUint64(constants.LDC),
		Expire: cs.Timestamp() + 1,
	}
	assert.NoError(tf.SyntacticVerify())
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeBorrow,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      approverAddr,
		To:        &stakeid,
		Token:     token.Ptr(),
		Data:      tf.Bytes(),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.ExSignWith(signer.Signer2))
	assert.NoError(ltx.SyntacticVerify())
	ltx.Timestamp = cs.Timestamp()
	itx, err = NewTx(ltx)
	require.NoError(t, err)
	approverAcc := cs.MustAccount(approverAddr)
	approverAcc.Add(constants.NativeToken, new(big.Int).SetUint64(constants.LDC))
	assert.NoError(itx.Apply(ctx, cs))

	approverGas := ltx.Gas()
	assert.Equal((keeperGas+stakeGas+approverGas)*ctx.Price,
		itx.(*TxBorrow).ldc.Balance().Uint64())
	assert.Equal((keeperGas+stakeGas+approverGas)*100,
		itx.(*TxBorrow).miner.Balance().Uint64())

	assert.Equal([]uint64{1, 2}, stakeAcc.LD().NonceTable[cs.Timestamp()+1])
	assert.Equal(constants.LDC*9, stakeAcc.BalanceOf(token).Uint64())
	assert.Equal(constants.LDC, approverAcc.BalanceOf(token).Uint64())

	// DestroyStake
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeDestroyStake,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     2,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      stakeid,
		To:        &keeper,
	}}
	assert.NoError(ltx.SignWith(signer.Signer1, signer.Signer2))
	assert.NoError(ltx.SyntacticVerify())
	ltx.Timestamp = cs.Timestamp()
	itx, err = NewTx(ltx)
	require.NoError(t, err)
	cs.CommitAccounts()
	assert.ErrorContains(itx.Apply(ctx, cs),
		"please repay all before close")
	cs.CheckoutAccounts()

	// TypeRepay
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeRepay,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     1,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      approverAddr,
		To:        &stakeid,
		Token:     token.Ptr(),
		Amount:    new(big.Int).SetUint64(constants.LDC),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	itx, err = NewTx(ltx)
	require.NoError(t, err)
	assert.NoError(itx.Apply(ctx, cs))

	approverGas += ltx.Gas()
	assert.Equal((keeperGas+stakeGas+approverGas)*ctx.Price,
		itx.(*TxRepay).ldc.Balance().Uint64())
	assert.Equal((keeperGas+stakeGas+approverGas)*100,
		itx.(*TxRepay).miner.Balance().Uint64())

	assert.Equal(constants.LDC*10, stakeAcc.BalanceOf(token).Uint64())
	assert.Equal(uint64(0), approverAcc.BalanceOf(token).Uint64())

	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeDestroyStake,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     2,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      stakeid,
		To:        &keeper,
	}}
	assert.NoError(ltx.SignWith(signer.Signer1, signer.Signer2))
	assert.NoError(ltx.SyntacticVerify())
	ltx.Timestamp = cs.Timestamp()
	itx, err = NewTx(ltx)
	require.NoError(t, err)
	assert.NoError(itx.Apply(ctx, cs))

	stakeGas += ltx.Gas()
	assert.Equal((keeperGas+stakeGas+approverGas)*ctx.Price,
		itx.(*TxDestroyStake).ldc.Balance().Uint64())
	assert.Equal((keeperGas+stakeGas+approverGas)*100,
		itx.(*TxDestroyStake).miner.Balance().Uint64())
	assert.Equal(ctx.FeeConfig().MinStakePledge.Uint64()+constants.LDC*2-(keeperGas+stakeGas)*(ctx.Price+100),
		keeperAcc.Balance().Uint64())
	assert.Equal(constants.LDC*0, stakeAcc.BalanceOf(token).Uint64())
	assert.Equal(constants.LDC*10, keeperAcc.BalanceOf(token).Uint64())

	assert.Equal(ld.AccountType(0), stakeAcc.LD().Type)
	assert.Equal(uint16(0), stakeAcc.LD().Threshold)
	assert.Equal(uint64(3), stakeAcc.LD().Nonce)
	assert.Equal(signer.Keys{}, stakeAcc.LD().Keepers)
	assert.Equal(make(map[uint64][]uint64), stakeAcc.LD().NonceTable)
	assert.Nil(stakeAcc.LD().Approver)
	assert.Nil(stakeAcc.LD().ApproveList)
	assert.Nil(stakeAcc.LD().Stake)
	assert.Nil(stakeAcc.LD().Lending)

	assert.NoError(cs.VerifyState())
}
