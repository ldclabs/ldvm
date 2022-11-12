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

func TestTxResetStake(t *testing.T) {
	assert := assert.New(t)

	// SyntacticVerify
	tx := &TxResetStake{}
	assert.ErrorContains(tx.SyntacticVerify(), "nil pointer")
	_, err := tx.MarshalJSON()
	require.NoError(t, err)

	ctx := NewMockChainContext()
	cs := ctx.MockChainState()
	stake := ld.MustNewStake("#TEST")
	stakeid := util.Address(stake)
	token := ld.MustNewToken("$TEST")
	sender := signer.Signer1.Key().Address()
	keeper := signer.Signer2.Key().Address()

	ltx := &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeResetStake,
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
		Type:      ld.TypeResetStake,
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
	assert.ErrorContains(err, "invalid to, should be nil")

	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeResetStake,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      sender,
		Token:     token.Ptr(),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "invalid token, should be nil")

	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeResetStake,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      sender,
		Amount:    big.NewInt(0),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.ErrorContains(ltx.SyntacticVerify(),
		"nil \"to\" together with amount")

	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeResetStake,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      sender,
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "invalid data")

	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeResetStake,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      sender,
		Data:      []byte("你好👋"),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "invalid stake account 0x8db97c7cECe249C2b98bdc0226cc4C2A57bF52fc")

	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeResetStake,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      stakeid,
		Data:      []byte("你好👋"),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err,
		"StakeConfig.Unmarshal: cbor")

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
	itx, err := NewTx(ltx)
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
	assert.Equal(uint64(0), keeperEntry.LockTime)
	assert.Nil(keeperEntry.Approver)

	input := &ld.StakeConfig{
		LockTime:    cs.Timestamp(),
		WithdrawFee: 1_000,
		MinAmount:   new(big.Int).SetUint64(constants.LDC * 100),
		MaxAmount:   new(big.Int).SetUint64(constants.LDC * 100),
	}
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeResetStake,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      stakeid,
		Data:      ld.MustMarshal(input),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	ltx.Timestamp = cs.Timestamp()
	_, err = NewTx(ltx)
	assert.ErrorContains(err,
		"TxResetStake.SyntacticVerify: invalid lockTime, expected > 1000, got 1000")

	input = &ld.StakeConfig{
		LockTime:    cs.Timestamp() + 1,
		WithdrawFee: 1_000,
		MinAmount:   new(big.Int).SetUint64(constants.LDC * 100),
		MaxAmount:   new(big.Int).SetUint64(constants.LDC * 100),
	}
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeResetStake,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      stakeid,
		Data:      ld.MustMarshal(input),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	ltx.Timestamp = cs.Timestamp()
	itx, err = NewTx(ltx)
	require.NoError(t, err)

	cs.CommitAccounts()
	assert.ErrorContains(itx.Apply(ctx, cs),
		"insufficient NativeLDC balance, expected 1855700, got 0")
	cs.CheckoutAccounts()

	stakeAcc.Add(constants.NativeToken, new(big.Int).SetUint64(constants.LDC))
	cs.CommitAccounts()
	assert.ErrorContains(itx.Apply(ctx, cs),
		"TxResetStake.Apply: invalid signatures for stake keepers")
	cs.CheckoutAccounts()

	input = &ld.StakeConfig{
		Type:        1,
		LockTime:    cs.Timestamp() + 1,
		WithdrawFee: 1_000,
		MinAmount:   new(big.Int).SetUint64(constants.LDC * 100),
		MaxAmount:   new(big.Int).SetUint64(constants.LDC * 100),
	}
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeResetStake,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      stakeid,
		Data:      ld.MustMarshal(input),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1, signer.Signer2))
	assert.NoError(ltx.SyntacticVerify())
	ltx.Timestamp = cs.Timestamp()
	itx, err = NewTx(ltx)
	require.NoError(t, err)
	cs.CommitAccounts()
	assert.ErrorContains(itx.Apply(ctx, cs),
		"can't change stake type, expected 0, got 1")
	cs.CheckoutAccounts()

	input = &ld.StakeConfig{
		Token:       token,
		LockTime:    cs.Timestamp() + 1,
		WithdrawFee: 1_000,
		MinAmount:   new(big.Int).SetUint64(constants.LDC * 100),
		MaxAmount:   new(big.Int).SetUint64(constants.LDC * 100),
	}
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeResetStake,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      stakeid,
		Data:      ld.MustMarshal(input),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1, signer.Signer2))
	assert.NoError(ltx.SyntacticVerify())
	ltx.Timestamp = cs.Timestamp()
	itx, err = NewTx(ltx)
	require.NoError(t, err)
	cs.CommitAccounts()
	assert.ErrorContains(itx.Apply(ctx, cs),
		"can't change stake token, expected NativeLDC, got $TEST")
	cs.CheckoutAccounts()

	input = &ld.StakeConfig{
		LockTime:    cs.Timestamp() + 1,
		WithdrawFee: 1_000,
		MinAmount:   new(big.Int).SetUint64(constants.LDC * 100),
		MaxAmount:   new(big.Int).SetUint64(constants.LDC * 100),
	}
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeResetStake,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      stakeid,
		Data:      ld.MustMarshal(input),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1, signer.Signer2))
	assert.NoError(ltx.SyntacticVerify())
	ltx.Timestamp = cs.Timestamp()
	itx, err = NewTx(ltx)
	require.NoError(t, err)
	cs.CommitAccounts()
	assert.ErrorContains(itx.Apply(ctx, cs),
		"stake in lock, please retry after lockTime, Unix(1100)")

	ctx.timestamp += 101
	cs.CheckoutAccounts()
	input = &ld.StakeConfig{
		LockTime:    cs.Timestamp() + 1,
		WithdrawFee: 1_000,
		MinAmount:   new(big.Int).SetUint64(constants.LDC * 100),
		MaxAmount:   new(big.Int).SetUint64(constants.LDC * 100),
	}
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeResetStake,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      stakeid,
		Data:      ld.MustMarshal(input),
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
	input3 := &ld.TxAccounter{Approver: signer.Signer2.Key().Ptr()}
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

	// reset again
	input = &ld.StakeConfig{
		LockTime:    cs.Timestamp() + 1,
		WithdrawFee: 1_000,
		MinAmount:   new(big.Int).SetUint64(constants.LDC * 100),
		MaxAmount:   new(big.Int).SetUint64(constants.LDC * 100),
	}
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeResetStake,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      stakeid,
		Data:      ld.MustMarshal(input),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1, signer.Signer2))
	assert.NoError(ltx.SyntacticVerify())
	ltx.Timestamp = cs.Timestamp()
	itx, err = NewTx(ltx)
	require.NoError(t, err)
	cs.CommitAccounts()
	assert.ErrorContains(itx.Apply(ctx, cs),
		"stake holders should not more than 1")
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

	input = &ld.StakeConfig{
		LockTime:    cs.Timestamp() + 1,
		WithdrawFee: 1_000,
		MinAmount:   new(big.Int).SetUint64(constants.LDC * 100),
		MaxAmount:   new(big.Int).SetUint64(constants.LDC * 100),
	}
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeResetStake,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      stakeid,
		Data:      ld.MustMarshal(input),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1, signer.Signer2))
	assert.NoError(ltx.SyntacticVerify())
	ltx.Timestamp = cs.Timestamp()
	itx, err = NewTx(ltx)
	require.NoError(t, err)
	assert.NoError(itx.Apply(ctx, cs))

	stakeGas := ltx.Gas()
	assert.Equal((keeperGas+senderGas+stakeGas)*ctx.Price,
		itx.(*TxResetStake).ldc.Balance().Uint64())
	assert.Equal((keeperGas+senderGas+stakeGas)*100,
		itx.(*TxResetStake).miner.Balance().Uint64())

	assert.Equal(input.LockTime, stakeAcc.LD().Stake.LockTime)
	assert.Equal(input.WithdrawFee, stakeAcc.LD().Stake.WithdrawFee)
	assert.Equal(constants.LDC*100, stakeAcc.LD().Stake.MinAmount.Uint64())
	assert.Equal(constants.LDC*100, stakeAcc.LD().Stake.MaxAmount.Uint64())
	assert.Equal(2, len(stakeAcc.Ledger().Stake))
	assert.Equal(constants.LDC*0, stakeAcc.Ledger().Stake[sender.AsKey()].Amount.Uint64())

	jsondata, err := itx.MarshalJSON()
	require.NoError(t, err)
	// fmt.Println(string(jsondata))
	assert.Equal(`{"tx":{"type":"TypeResetStake","chainID":2357,"nonce":0,"gasTip":100,"gasFeeCap":1000,"from":"0x0000000000000000000000000000002354455354","data":{"token":"","type":0,"lockTime":1102,"withdrawFee":1000,"minAmount":100000000000,"maxAmount":100000000000}},"sigs":["UP78TBuYAouPKIVTITDyyPY9Lq_2p3xHpKy--G8J2uNmikmtzVcv3iCKkCu2Y7Ht2_zc07QSNF6RjFaV7FUgcwE-HkXb","l9bMNy_e2yWeGfXPfrKT10CKVR70ZpW223bGvHO7rNIPPCc8lDDQegDqGftQ_M9aHfCoYqz1wa39CmKiVyybHgDjMNRy"],"id":"selx2k_hGWS7rk2yujcPa6xT7_OIIB-HEVmpHXtrYnEDkeZ2"}`, string(jsondata))

	assert.NoError(cs.VerifyState())
}
