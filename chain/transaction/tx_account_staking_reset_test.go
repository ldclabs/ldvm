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

func TestTxResetStake(t *testing.T) {
	assert := assert.New(t)

	// SyntacticVerify
	tx := &TxResetStake{}
	assert.ErrorContains(tx.SyntacticVerify(), "nil pointer")
	_, err := tx.MarshalJSON()
	assert.NoError(err)

	bctx := NewMockBCtx()
	bs := bctx.MockBS()
	stake := ld.MustNewStake("#TEST")
	stakeid := util.EthID(stake)
	token := ld.MustNewToken("$TEST")
	sender := util.Signer1.Address()
	keeper := util.Signer2.Address()

	txData := &ld.TxData{
		Type:      ld.TypeResetStake,
		ChainID:   bctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      sender,
	}
	assert.NoError(txData.SyntacticVerify())
	_, err = NewTx(txData.ToTransaction(), true)
	assert.ErrorContains(err, "DeriveSigners error: no signature")

	txData = &ld.TxData{
		Type:      ld.TypeResetStake,
		ChainID:   bctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      sender,
		To:        &keeper,
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx(txData.ToTransaction(), true)
	assert.ErrorContains(err, "invalid to, should be nil")

	txData = &ld.TxData{
		Type:      ld.TypeResetStake,
		ChainID:   bctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      sender,
		Token:     &token,
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx(txData.ToTransaction(), true)
	assert.ErrorContains(err, "invalid token, should be nil")

	txData = &ld.TxData{
		Type:      ld.TypeResetStake,
		ChainID:   bctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      sender,
		Amount:    big.NewInt(0),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx(txData.ToTransaction(), true)
	assert.ErrorContains(err,
		"Transaction.SyntacticVerify error: TxData.SyntacticVerify error: nil to together with amount")

	txData = &ld.TxData{
		Type:      ld.TypeResetStake,
		ChainID:   bctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      sender,
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx(txData.ToTransaction(), true)
	assert.ErrorContains(err, "invalid data")

	txData = &ld.TxData{
		Type:      ld.TypeResetStake,
		ChainID:   bctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      sender,
		Data:      []byte("你好👋"),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx(txData.ToTransaction(), true)
	assert.ErrorContains(err, "invalid stake account 0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC")

	txData = &ld.TxData{
		Type:      ld.TypeResetStake,
		ChainID:   bctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      stakeid,
		Data:      []byte("你好👋"),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx(txData.ToTransaction(), true)
	assert.ErrorContains(err,
		"StakeConfig.Unmarshal error: cbor: cannot unmarshal primitives into Go value of type ld.StakeConfig")

	// create a stake account for testing
	scfg := &ld.StakeConfig{
		LockTime:    bs.Timestamp() + 100,
		WithdrawFee: 100_000,
		MinAmount:   new(big.Int).SetUint64(constants.LDC * 10),
		MaxAmount:   new(big.Int).SetUint64(constants.LDC * 100),
	}
	sinput := &ld.TxAccounter{
		Threshold: ld.Uint16Ptr(1),
		Keepers:   &util.EthIDs{util.Signer1.Address(), util.Signer2.Address()},
		Data:      ld.MustMarshal(scfg),
	}
	txData = &ld.TxData{
		Type:      ld.TypeCreateStake,
		ChainID:   bctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      keeper,
		To:        &stakeid,
		Amount:    new(big.Int).Set(bctx.FeeConfig().MinStakePledge),
		Data:      sinput.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer2))
	tt := txData.ToTransaction()
	tt.Timestamp = bs.Timestamp()
	itx, err := NewTx(tt, true)
	assert.NoError(err)

	keeperAcc := bs.MustAccount(keeper)
	keeperAcc.Add(constants.NativeToken,
		new(big.Int).SetUint64(bctx.FeeConfig().MinStakePledge.Uint64()+constants.LDC))
	assert.NoError(itx.Apply(bctx, bs))

	keeperGas := tt.Gas()
	stakeAcc := bs.MustAccount(stakeid)
	assert.Equal(keeperGas*bctx.Price,
		itx.(*TxCreateStake).ldc.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(keeperGas*100,
		itx.(*TxCreateStake).miner.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(constants.LDC*0, stakeAcc.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(bctx.FeeConfig().MinStakePledge.Uint64(),
		stakeAcc.balanceOfAll(constants.NativeToken).Uint64())
	assert.Equal(constants.LDC-keeperGas*(bctx.Price+100),
		keeperAcc.balanceOf(constants.NativeToken).Uint64())

	assert.NotNil(stakeAcc.ledger)
	keeperEntry := stakeAcc.ledger.Stake[keeper.AsKey()]
	assert.NotNil(keeperEntry)
	assert.Equal(bctx.FeeConfig().MinStakePledge.Uint64(), keeperEntry.Amount.Uint64())
	assert.Equal(uint64(0), keeperEntry.LockTime)
	assert.Nil(keeperEntry.Approver)

	input := &ld.StakeConfig{
		LockTime:    bs.Timestamp(),
		WithdrawFee: 1_000,
		MinAmount:   new(big.Int).SetUint64(constants.LDC * 100),
		MaxAmount:   new(big.Int).SetUint64(constants.LDC * 100),
	}
	txData = &ld.TxData{
		Type:      ld.TypeResetStake,
		ChainID:   bctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      stakeid,
		Data:      ld.MustMarshal(input),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	tt = txData.ToTransaction()
	tt.Timestamp = bs.Timestamp()
	_, err = NewTx(tt, true)
	assert.ErrorContains(err,
		"TxResetStake.SyntacticVerify error: invalid lockTime, expected > 1000, got 1000")

	input = &ld.StakeConfig{
		LockTime:    bs.Timestamp() + 1,
		WithdrawFee: 1_000,
		MinAmount:   new(big.Int).SetUint64(constants.LDC * 100),
		MaxAmount:   new(big.Int).SetUint64(constants.LDC * 100),
	}
	txData = &ld.TxData{
		Type:      ld.TypeResetStake,
		ChainID:   bctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      stakeid,
		Data:      ld.MustMarshal(input),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	tt = txData.ToTransaction()
	tt.Timestamp = bs.Timestamp()
	itx, err = NewTx(tt, true)
	assert.NoError(err)

	bs.CommitAccounts()
	assert.ErrorContains(itx.Apply(bctx, bs),
		"insufficient NativeLDC balance, expected 1832600, got 0")
	bs.CheckoutAccounts()

	stakeAcc.Add(constants.NativeToken, new(big.Int).SetUint64(constants.LDC))
	bs.CommitAccounts()
	assert.ErrorContains(itx.Apply(bctx, bs),
		"TxResetStake.Apply error: invalid signatures for stake keepers")
	bs.CheckoutAccounts()

	input = &ld.StakeConfig{
		Type:        1,
		LockTime:    bs.Timestamp() + 1,
		WithdrawFee: 1_000,
		MinAmount:   new(big.Int).SetUint64(constants.LDC * 100),
		MaxAmount:   new(big.Int).SetUint64(constants.LDC * 100),
	}
	txData = &ld.TxData{
		Type:      ld.TypeResetStake,
		ChainID:   bctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      stakeid,
		Data:      ld.MustMarshal(input),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	assert.NoError(txData.SignWith(util.Signer2))
	tt = txData.ToTransaction()
	tt.Timestamp = bs.Timestamp()
	itx, err = NewTx(tt, true)
	assert.NoError(err)
	bs.CommitAccounts()
	assert.ErrorContains(itx.Apply(bctx, bs),
		"TxResetStake.Apply error: Account(0x0000000000000000000000000000002354455354).ResetStake error: can't change stake type, expected 0, got 1")
	bs.CheckoutAccounts()

	input = &ld.StakeConfig{
		Token:       token,
		LockTime:    bs.Timestamp() + 1,
		WithdrawFee: 1_000,
		MinAmount:   new(big.Int).SetUint64(constants.LDC * 100),
		MaxAmount:   new(big.Int).SetUint64(constants.LDC * 100),
	}
	txData = &ld.TxData{
		Type:      ld.TypeResetStake,
		ChainID:   bctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      stakeid,
		Data:      ld.MustMarshal(input),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	assert.NoError(txData.SignWith(util.Signer2))
	tt = txData.ToTransaction()
	tt.Timestamp = bs.Timestamp()
	itx, err = NewTx(tt, true)
	assert.NoError(err)
	bs.CommitAccounts()
	assert.ErrorContains(itx.Apply(bctx, bs),
		"TxResetStake.Apply error: Account(0x0000000000000000000000000000002354455354).ResetStake error: can't change stake token, expected NativeLDC, got $TEST")
	bs.CheckoutAccounts()

	input = &ld.StakeConfig{
		LockTime:    bs.Timestamp() + 1,
		WithdrawFee: 1_000,
		MinAmount:   new(big.Int).SetUint64(constants.LDC * 100),
		MaxAmount:   new(big.Int).SetUint64(constants.LDC * 100),
	}
	txData = &ld.TxData{
		Type:      ld.TypeResetStake,
		ChainID:   bctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      stakeid,
		Data:      ld.MustMarshal(input),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	assert.NoError(txData.SignWith(util.Signer2))
	tt = txData.ToTransaction()
	tt.Timestamp = bs.Timestamp()
	itx, err = NewTx(tt, true)
	assert.NoError(err)
	bs.CommitAccounts()
	assert.ErrorContains(itx.Apply(bctx, bs),
		"TxResetStake.Apply error: Account(0x0000000000000000000000000000002354455354).ResetStake error: stake in lock, please retry after lockTime, Unix(1100)")

	bctx.timestamp += 101
	bs.CheckoutAccounts()
	input = &ld.StakeConfig{
		LockTime:    bs.Timestamp() + 1,
		WithdrawFee: 1_000,
		MinAmount:   new(big.Int).SetUint64(constants.LDC * 100),
		MaxAmount:   new(big.Int).SetUint64(constants.LDC * 100),
	}
	txData = &ld.TxData{
		Type:      ld.TypeResetStake,
		ChainID:   bctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      stakeid,
		Data:      ld.MustMarshal(input),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	assert.NoError(txData.SignWith(util.Signer2))
	tt = txData.ToTransaction()
	tt.Timestamp = bs.Timestamp()
	itx, err = NewTx(tt, true)
	assert.NoError(err)
	bs.CommitAccounts()
	assert.NoError(itx.Apply(bctx, bs))
	bs.CheckoutAccounts()

	// take a stake for testing
	input2 := &ld.TxTransfer{
		Nonce:  0,
		From:   &sender,
		To:     &stakeid,
		Amount: new(big.Int).SetUint64(constants.LDC * 10),
		Expire: bs.Timestamp(),
	}
	txData = &ld.TxData{
		Type:      ld.TypeTakeStake,
		ChainID:   bctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      sender,
		To:        &stakeid,
		Amount:    new(big.Int).SetUint64(constants.LDC * 10),
		Data:      input2.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	assert.NoError(txData.ExSignWith(util.Signer2))
	tt = txData.ToTransaction()
	tt.Timestamp = bs.Timestamp()
	itx, err = NewTx(tt, true)
	assert.NoError(err)

	senderAcc := bs.MustAccount(sender)
	senderAcc.Add(constants.NativeToken, new(big.Int).SetUint64(constants.LDC*11))
	assert.NoError(itx.Apply(bctx, bs))

	senderGas := tt.Gas()
	assert.Equal((keeperGas+senderGas)*bctx.Price,
		itx.(*TxTakeStake).ldc.balanceOf(constants.NativeToken).Uint64())
	assert.Equal((keeperGas+senderGas)*100,
		itx.(*TxTakeStake).miner.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(bctx.FeeConfig().MinStakePledge.Uint64()+constants.LDC*11,
		stakeAcc.balanceOfAll(constants.NativeToken).Uint64())
	assert.Equal(constants.LDC-senderGas*(bctx.Price+100),
		senderAcc.balanceOf(constants.NativeToken).Uint64())
	senderEntry := stakeAcc.ledger.Stake[sender.AsKey()]
	assert.NotNil(senderEntry)
	assert.Equal(constants.LDC*10, senderEntry.Amount.Uint64())
	assert.Equal(uint64(0), senderEntry.LockTime)
	assert.Nil(senderEntry.Approver)

	// add stake approver for testing
	input3 := &ld.TxAccounter{Approver: &keeper}
	txData = &ld.TxData{
		Type:      ld.TypeUpdateStakeApprover,
		ChainID:   bctx.ChainConfig().ChainID,
		Nonce:     1,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      sender,
		To:        &stakeid,
		Data:      input3.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	tt = txData.ToTransaction()
	itx, err = NewTx(tt, true)
	assert.NoError(err)
	assert.NoError(itx.Apply(bctx, bs))

	senderGas += tt.Gas()
	assert.Equal((keeperGas+senderGas)*bctx.Price,
		itx.(*TxUpdateStakeApprover).ldc.balanceOf(constants.NativeToken).Uint64())
	assert.Equal((keeperGas+senderGas)*100,
		itx.(*TxUpdateStakeApprover).miner.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(constants.LDC-senderGas*(bctx.Price+100),
		senderAcc.balanceOf(constants.NativeToken).Uint64())
	senderEntry = stakeAcc.ledger.Stake[sender.AsKey()]
	assert.NotNil(senderEntry)
	assert.NotNil(senderEntry.Approver)
	assert.Equal(keeper, *senderEntry.Approver)

	// reset again
	input = &ld.StakeConfig{
		LockTime:    bs.Timestamp() + 1,
		WithdrawFee: 1_000,
		MinAmount:   new(big.Int).SetUint64(constants.LDC * 100),
		MaxAmount:   new(big.Int).SetUint64(constants.LDC * 100),
	}
	txData = &ld.TxData{
		Type:      ld.TypeResetStake,
		ChainID:   bctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      stakeid,
		Data:      ld.MustMarshal(input),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	assert.NoError(txData.SignWith(util.Signer2))
	tt = txData.ToTransaction()
	tt.Timestamp = bs.Timestamp()
	itx, err = NewTx(tt, true)
	assert.NoError(err)
	bs.CommitAccounts()
	assert.ErrorContains(itx.Apply(bctx, bs),
		"TxResetStake.Apply error: Account(0x0000000000000000000000000000002354455354).ResetStake error: stake holders should not more than 1")
	bs.CheckoutAccounts()

	input2 = &ld.TxTransfer{Amount: new(big.Int).SetUint64(constants.LDC * 10)}
	txData = &ld.TxData{
		Type:      ld.TypeWithdrawStake,
		ChainID:   bctx.ChainConfig().ChainID,
		Nonce:     2,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      sender,
		To:        &stakeid,
		Data:      input2.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	assert.NoError(txData.SignWith(util.Signer2))
	tt = txData.ToTransaction()
	tt.Timestamp = bs.Timestamp()
	itx, err = NewTx(tt, true)
	assert.NoError(err)
	assert.NoError(itx.Apply(bctx, bs))

	senderGas += tt.Gas()
	assert.Equal((keeperGas+senderGas)*bctx.Price,
		itx.(*TxWithdrawStake).ldc.balanceOf(constants.NativeToken).Uint64())
	assert.Equal((keeperGas+senderGas)*100,
		itx.(*TxWithdrawStake).miner.balanceOf(constants.NativeToken).Uint64())

	withdrawFee := constants.LDC * 10 * scfg.WithdrawFee / 1_000_000
	assert.Equal(constants.LDC*11-withdrawFee-senderGas*(bctx.Price+100),
		senderAcc.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(constants.LDC+withdrawFee, stakeAcc.balanceOf(constants.NativeToken).Uint64())
	assert.NotNil(stakeAcc.ledger.Stake[sender.AsKey()])
	assert.Equal(constants.LDC*0, stakeAcc.ledger.Stake[sender.AsKey()].Amount.Uint64())

	input = &ld.StakeConfig{
		LockTime:    bs.Timestamp() + 1,
		WithdrawFee: 1_000,
		MinAmount:   new(big.Int).SetUint64(constants.LDC * 100),
		MaxAmount:   new(big.Int).SetUint64(constants.LDC * 100),
	}
	txData = &ld.TxData{
		Type:      ld.TypeResetStake,
		ChainID:   bctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      stakeid,
		Data:      ld.MustMarshal(input),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	assert.NoError(txData.SignWith(util.Signer2))
	tt = txData.ToTransaction()
	tt.Timestamp = bs.Timestamp()
	itx, err = NewTx(tt, true)
	assert.NoError(err)
	assert.NoError(itx.Apply(bctx, bs))

	stakeGas := tt.Gas()
	assert.Equal((keeperGas+senderGas+stakeGas)*bctx.Price,
		itx.(*TxResetStake).ldc.balanceOf(constants.NativeToken).Uint64())
	assert.Equal((keeperGas+senderGas+stakeGas)*100,
		itx.(*TxResetStake).miner.balanceOf(constants.NativeToken).Uint64())

	assert.Equal(input.LockTime, stakeAcc.ld.Stake.LockTime)
	assert.Equal(input.WithdrawFee, stakeAcc.ld.Stake.WithdrawFee)
	assert.Equal(constants.LDC*100, stakeAcc.ld.Stake.MinAmount.Uint64())
	assert.Equal(constants.LDC*100, stakeAcc.ld.Stake.MaxAmount.Uint64())
	assert.Equal(2, len(stakeAcc.ledger.Stake))
	assert.Equal(constants.LDC*0, stakeAcc.ledger.Stake[sender.AsKey()].Amount.Uint64())

	jsondata, err := itx.MarshalJSON()
	assert.NoError(err)
	// fmt.Println(string(jsondata))
	assert.Equal(`{"type":"TypeResetStake","chainID":2357,"nonce":0,"gasTip":100,"gasFeeCap":1000,"from":"0x0000000000000000000000000000002354455354","data":{"token":"","type":0,"lockTime":1102,"withdrawFee":1000,"minAmount":100000000000,"maxAmount":100000000000},"signatures":["50fefc4c1b98028b8f2885532130f2c8f63d2eaff6a77c47a4acbef86f09dae3668a49adcd572fde208a902bb663b1eddbfcdcd3b412345e918c5695ec55207301","97d6cc372fdedb259e19f5cf7eb293d7408a551ef46695b6db76c6bc73bbacd20f3c273c9430d07a00ea19fb50fccf5a1df0a862acf5c1adfd0a62a2572c9b1e00"],"id":"UkNj9cPkLrx99DgMA36bfQdTSyEb1pKRpmKYbUtDg26PAmSJb"}`, string(jsondata))

	assert.NoError(bs.VerifyState())
}
