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

func TestTxDestroyStake(t *testing.T) {
	assert := assert.New(t)

	// SyntacticVerify
	tx := &TxDestroyStake{}
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
		Type:      ld.TypeDestroyStake,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      sender,
	}
	assert.NoError(txData.SyntacticVerify())
	_, err = NewTx(txData.ToTransaction(), true)
	assert.ErrorContains(err, "DeriveSigners error: no signature")

	txData = &ld.TxData{
		Type:      ld.TypeDestroyStake,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      sender,
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx(txData.ToTransaction(), true)
	assert.ErrorContains(err, "nil to as pledge recipient")

	txData = &ld.TxData{
		Type:      ld.TypeDestroyStake,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      sender,
		To:        &keeper,
		Token:     &token,
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx(txData.ToTransaction(), true)
	assert.ErrorContains(err, "invalid token, should be nil")

	txData = &ld.TxData{
		Type:      ld.TypeDestroyStake,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      sender,
		To:        &keeper,
		Amount:    big.NewInt(1),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx(txData.ToTransaction(), true)
	assert.ErrorContains(err, "invalid amount, should be nil")

	txData = &ld.TxData{
		Type:      ld.TypeDestroyStake,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      sender,
		To:        &keeper,
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx(txData.ToTransaction(), true)
	assert.ErrorContains(err,
		"TxDestroyStake.SyntacticVerify error: invalid stake account 0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC")

	txData = &ld.TxData{
		Type:      ld.TypeDestroyStake,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      stakeid,
		To:        &keeper,
	}
	assert.NoError(txData.SignWith(util.Signer1))
	tt := txData.ToTransaction()
	itx, err := NewTx(tt, true)
	assert.NoError(err)

	bs.CommitAccounts()
	assert.ErrorContains(itx.Apply(bctx, bs), "invalid gas, expected 1230, got 0")
	bs.CheckoutAccounts()

	tt.Timestamp = bs.Timestamp()
	tt.Gas = tt.RequiredGas(bctx.FeeConfig().ThresholdGas)
	itx, err = NewTx(tt, true)
	assert.NoError(err)
	bs.CommitAccounts()
	assert.ErrorContains(itx.Apply(bctx, bs),
		"TxDestroyStake.Apply error: invalid signatures for sender")
	bs.CheckoutAccounts()

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
		ChainID:   bctx.Chain().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      keeper,
		To:        &stakeid,
		Amount:    new(big.Int).Set(bctx.FeeConfig().MinStakePledge),
		Data:      sinput.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer2))
	tt = txData.ToTransaction()
	tt.Timestamp = bs.Timestamp()
	tt.Gas = tt.RequiredGas(bctx.FeeConfig().ThresholdGas)
	keeperGas := tt.Gas
	itx, err = NewTx(tt, true)
	assert.NoError(err)

	keeperAcc := bs.MustAccount(keeper)
	keeperAcc.Add(constants.NativeToken,
		new(big.Int).SetUint64(bctx.FeeConfig().MinStakePledge.Uint64()+constants.LDC))
	assert.NoError(itx.Apply(bctx, bs))

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

	assert.NotNil(stakeAcc.ld.StakeLedger)
	keeperEntry := stakeAcc.ld.StakeLedger[keeper]
	assert.NotNil(keeperEntry)
	assert.Equal(bctx.FeeConfig().MinStakePledge.Uint64(), keeperEntry.Amount.Uint64())

	txData = &ld.TxData{
		Type:      ld.TypeDestroyStake,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      stakeid,
		To:        &keeper,
	}
	assert.NoError(txData.SignWith(util.Signer1))
	tt = txData.ToTransaction()
	tt.Timestamp = bs.Timestamp()
	tt.Gas = tt.RequiredGas(bctx.FeeConfig().ThresholdGas)
	itx, err = NewTx(tt, true)
	assert.NoError(err)

	bs.CommitAccounts()
	assert.ErrorContains(itx.Apply(bctx, bs),
		"insufficient NativeLDC balance, expected 1353000, got 0")
	bs.CheckoutAccounts()

	stakeAcc.Add(constants.NativeToken, new(big.Int).SetUint64(constants.LDC))
	bs.CommitAccounts()
	assert.ErrorContains(itx.Apply(bctx, bs),
		"TxDestroyStake.Apply error: invalid signatures for stake keepers")
	bs.CheckoutAccounts()

	assert.NoError(txData.SignWith(util.Signer2))
	tt = txData.ToTransaction()
	tt.Timestamp = bs.Timestamp()
	tt.Gas = tt.RequiredGas(bctx.FeeConfig().ThresholdGas)
	itx, err = NewTx(tt, true)
	assert.NoError(err)
	bs.CommitAccounts()
	assert.ErrorContains(itx.Apply(bctx, bs),
		"TxDestroyStake.Apply error: Account(0x0000000000000000000000000000002354455354).DestroyStake error: stake in lock, please retry after lockTime, Unix(1100)")
	bs.CheckoutAccounts()

	bctx.timestamp += 101
	bs.CheckoutAccounts()
	txData = &ld.TxData{
		Type:      ld.TypeDestroyStake,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      stakeid,
		To:        &keeper,
	}
	assert.NoError(txData.SignWith(util.Signer1))
	assert.NoError(txData.SignWith(util.Signer2))
	tt = txData.ToTransaction()
	tt.Timestamp = bs.Timestamp()
	tt.Gas = tt.RequiredGas(bctx.FeeConfig().ThresholdGas)
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
		ChainID:   bctx.Chain().ChainID,
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
	tt.Gas = tt.RequiredGas(bctx.FeeConfig().ThresholdGas)
	senderGas := tt.Gas
	tt.Timestamp = bs.Timestamp()
	itx, err = NewTx(tt, true)
	assert.NoError(err)

	senderAcc := bs.MustAccount(sender)
	senderAcc.Add(constants.NativeToken, new(big.Int).SetUint64(constants.LDC*11))
	assert.NoError(itx.Apply(bctx, bs))

	assert.Equal((keeperGas+senderGas)*bctx.Price,
		itx.(*TxTakeStake).ldc.balanceOf(constants.NativeToken).Uint64())
	assert.Equal((keeperGas+senderGas)*100,
		itx.(*TxTakeStake).miner.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(bctx.FeeConfig().MinStakePledge.Uint64()+constants.LDC*11,
		stakeAcc.balanceOfAll(constants.NativeToken).Uint64())
	assert.Equal(constants.LDC-senderGas*(bctx.Price+100),
		senderAcc.balanceOf(constants.NativeToken).Uint64())
	senderEntry := stakeAcc.ld.StakeLedger[sender]
	assert.NotNil(senderEntry)
	assert.Equal(constants.LDC*10, senderEntry.Amount.Uint64())
	assert.Equal(uint64(0), senderEntry.LockTime)
	assert.Nil(senderEntry.Approver)

	// add stake approver for testing
	input3 := &ld.TxAccounter{Approver: &keeper}
	txData = &ld.TxData{
		Type:      ld.TypeUpdateStakeApprover,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     1,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      sender,
		To:        &stakeid,
		Data:      input3.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	tt = txData.ToTransaction()
	tt.Gas = tt.RequiredGas(bctx.FeeConfig().ThresholdGas)
	senderGas += tt.Gas
	itx, err = NewTx(tt, true)
	assert.NoError(err)
	assert.NoError(itx.Apply(bctx, bs))

	assert.Equal((keeperGas+senderGas)*bctx.Price,
		itx.(*TxUpdateStakeApprover).ldc.balanceOf(constants.NativeToken).Uint64())
	assert.Equal((keeperGas+senderGas)*100,
		itx.(*TxUpdateStakeApprover).miner.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(constants.LDC-senderGas*(bctx.Price+100),
		senderAcc.balanceOf(constants.NativeToken).Uint64())
	senderEntry = stakeAcc.ld.StakeLedger[sender]
	assert.NotNil(senderEntry)
	assert.NotNil(senderEntry.Approver)
	assert.Equal(keeper, *senderEntry.Approver)

	// destroy again
	txData = &ld.TxData{
		Type:      ld.TypeDestroyStake,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      stakeid,
		To:        &keeper,
	}
	assert.NoError(txData.SignWith(util.Signer1))
	assert.NoError(txData.SignWith(util.Signer2))
	tt = txData.ToTransaction()
	tt.Timestamp = bs.Timestamp()
	tt.Gas = tt.RequiredGas(bctx.FeeConfig().ThresholdGas)
	itx, err = NewTx(tt, true)
	assert.NoError(err)
	bs.CommitAccounts()
	assert.ErrorContains(itx.Apply(bctx, bs),
		"TxDestroyStake.Apply error: Account(0x0000000000000000000000000000002354455354).DestroyStake error: stake ledger not empty, please withdraw all except recipient")
	bs.CheckoutAccounts()

	input2 = &ld.TxTransfer{Amount: new(big.Int).SetUint64(constants.LDC * 10)}
	txData = &ld.TxData{
		Type:      ld.TypeWithdrawStake,
		ChainID:   bctx.Chain().ChainID,
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
	tt.Gas = tt.RequiredGas(bctx.FeeConfig().ThresholdGas)
	senderGas += tt.Gas
	tt.Timestamp = bs.Timestamp()
	itx, err = NewTx(tt, true)
	assert.NoError(err)
	assert.NoError(itx.Apply(bctx, bs))

	assert.Equal((keeperGas+senderGas)*bctx.Price,
		itx.(*TxWithdrawStake).ldc.balanceOf(constants.NativeToken).Uint64())
	assert.Equal((keeperGas+senderGas)*100,
		itx.(*TxWithdrawStake).miner.balanceOf(constants.NativeToken).Uint64())

	withdrawFee := constants.LDC * 10 * scfg.WithdrawFee / 1_000_000
	assert.Equal(constants.LDC*11-withdrawFee-senderGas*(bctx.Price+100),
		senderAcc.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(constants.LDC+withdrawFee, stakeAcc.balanceOf(constants.NativeToken).Uint64())
	assert.NotNil(stakeAcc.ld.StakeLedger[sender])
	assert.Equal(constants.LDC*0, stakeAcc.ld.StakeLedger[sender].Amount.Uint64())

	txData = &ld.TxData{
		Type:      ld.TypeDestroyStake,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      stakeid,
		To:        &keeper,
	}
	assert.NoError(txData.SignWith(util.Signer1))
	assert.NoError(txData.SignWith(util.Signer2))
	tt = txData.ToTransaction()
	tt.Timestamp = bs.Timestamp()
	tt.Gas = tt.RequiredGas(bctx.FeeConfig().ThresholdGas)
	stakeGas := tt.Gas
	itx, err = NewTx(tt, true)
	assert.NoError(err)
	assert.NoError(itx.Apply(bctx, bs))

	assert.Equal((keeperGas+senderGas+stakeGas)*bctx.Price,
		itx.(*TxDestroyStake).ldc.balanceOf(constants.NativeToken).Uint64())
	assert.Equal((keeperGas+senderGas+stakeGas)*100,
		itx.(*TxDestroyStake).miner.balanceOf(constants.NativeToken).Uint64())

	assert.Equal(bctx.FeeConfig().MinStakePledge.Uint64()+constants.LDC*2+withdrawFee-(keeperGas+stakeGas)*(bctx.Price+100),
		keeperAcc.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(ld.AccountType(0), stakeAcc.ld.Type)
	assert.Equal(uint16(0), stakeAcc.ld.Threshold)
	assert.Equal(uint64(1), stakeAcc.ld.Nonce)
	assert.Equal(util.EthIDs{}, stakeAcc.ld.Keepers)
	assert.Equal(make(map[uint64][]uint64), stakeAcc.ld.NonceTable)
	assert.Nil(stakeAcc.ld.Approver)
	assert.Nil(stakeAcc.ld.ApproveList)
	assert.Nil(stakeAcc.ld.Stake)
	assert.Nil(stakeAcc.ld.StakeLedger)

	jsondata, err := itx.MarshalJSON()
	assert.NoError(err)
	// fmt.Println(string(jsondata))
	assert.Equal(`{"type":"TypeDestroyStake","chainID":2357,"nonce":0,"gasTip":100,"gasFeeCap":1000,"from":"0x0000000000000000000000000000002354455354","to":"0x44171C37Ff5D7B7bb8dcad5C81f16284A229e641","signatures":["e3c39546e69ffd01efc5d50d3a76435e750030a2810dcc478aa1dacea7e9d0c00246612fef17b7338ba4b5ad7b6a567fb757b6bd8733688656e1a461b582778101","6e12513eddbd732d027451e5f169f6c2023294258e8191fbd058b593d645892e6fe3e1450a5d80c27be6329d4d1de5dfe20409b1c12447262cfbe9f17cedb65901"],"gas":1230,"id":"G49AutnLGHtMQ2atyZ12MxFRidWBaXBgLDVbvY597vXpcTzta"}`, string(jsondata))

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
		Keepers:   &util.EthIDs{util.Signer2.Address()},
		Data:      ld.MustMarshal(scfg),
	}
	txData = &ld.TxData{
		Type:      ld.TypeCreateStake,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     1,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      keeper,
		To:        &stakeid,
		Amount:    new(big.Int).SetUint64(bctx.FeeConfig().MinStakePledge.Uint64() + constants.LDC),
		Data:      sinput.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer2))
	tt = txData.ToTransaction()
	tt.Timestamp = bs.Timestamp()
	tt.Gas = tt.RequiredGas(bctx.FeeConfig().ThresholdGas)
	keeperGas += tt.Gas
	itx, err = NewTx(tt, true)
	assert.NoError(err)
	assert.NoError(itx.Apply(bctx, bs))

	assert.Equal((keeperGas+senderGas+stakeGas)*bctx.Price,
		itx.(*TxCreateStake).ldc.balanceOf(constants.NativeToken).Uint64())
	assert.Equal((keeperGas+senderGas+stakeGas)*100,
		itx.(*TxCreateStake).miner.balanceOf(constants.NativeToken).Uint64())

	assert.Equal(constants.LDC, stakeAcc.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(bctx.FeeConfig().MinStakePledge.Uint64()+constants.LDC,
		stakeAcc.balanceOfAll(constants.NativeToken).Uint64())
	assert.Equal(uint64(0), keeperAcc.balanceOf(token).Uint64())

	assert.NotNil(stakeAcc.ld.StakeLedger)
	assert.Equal(0, len(stakeAcc.ld.StakeLedger))

	stakeAcc.Add(token, new(big.Int).SetUint64(constants.LDC*9))
	txData = &ld.TxData{
		Type:      ld.TypeDestroyStake,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     1,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      stakeid,
		To:        &keeper,
	}
	assert.NoError(txData.SignWith(util.Signer2))
	tt = txData.ToTransaction()
	tt.Timestamp = bs.Timestamp()
	tt.Gas = tt.RequiredGas(bctx.FeeConfig().ThresholdGas)
	stakeGas += tt.Gas
	itx, err = NewTx(tt, true)
	assert.NoError(err)
	assert.NoError(itx.Apply(bctx, bs))

	assert.Equal((keeperGas+senderGas+stakeGas)*bctx.Price,
		itx.(*TxDestroyStake).ldc.balanceOf(constants.NativeToken).Uint64())
	assert.Equal((keeperGas+senderGas+stakeGas)*100,
		itx.(*TxDestroyStake).miner.balanceOf(constants.NativeToken).Uint64())

	assert.Equal(bctx.FeeConfig().MinStakePledge.Uint64()+constants.LDC*2+withdrawFee-(keeperGas+stakeGas)*(bctx.Price+100),
		keeperAcc.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(constants.LDC*9,
		keeperAcc.balanceOf(token).Uint64())
	assert.Equal(ld.AccountType(0), stakeAcc.ld.Type)
	assert.Equal(uint16(0), stakeAcc.ld.Threshold)
	assert.Equal(uint64(2), stakeAcc.ld.Nonce)
	assert.Equal(util.EthIDs{}, stakeAcc.ld.Keepers)
	assert.Equal(make(map[uint64][]uint64), stakeAcc.ld.NonceTable)
	assert.Nil(stakeAcc.ld.Approver)
	assert.Nil(stakeAcc.ld.ApproveList)
	assert.Nil(stakeAcc.ld.Stake)
	assert.Nil(stakeAcc.ld.StakeLedger)

	assert.NoError(bs.VerifyState())
}

func TestTxDestroyStakeWithApproverAndLending(t *testing.T) {
	assert := assert.New(t)

	bctx := NewMockBCtx()
	bs := bctx.MockBS()
	stake := ld.MustNewStake("#TEST")
	stakeid := util.EthID(stake)
	token := ld.MustNewToken("$TEST")
	approver := util.Signer1.Address()
	keeper := util.Signer2.Address()

	scfg := &ld.StakeConfig{
		Token:       token,
		Type:        1,
		WithdrawFee: 100_000,
		MinAmount:   new(big.Int).SetUint64(constants.LDC * 10),
		MaxAmount:   new(big.Int).SetUint64(constants.LDC * 100),
	}
	input := &ld.TxAccounter{
		Threshold:   ld.Uint16Ptr(1),
		Keepers:     &util.EthIDs{util.Signer2.Address()},
		Approver:    &approver,
		ApproveList: ld.TxTypes{ld.TypeOpenLending, ld.TypeDestroyStake},
		Data:        ld.MustMarshal(scfg),
	}
	txData := &ld.TxData{
		Type:      ld.TypeCreateStake,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      keeper,
		To:        &stakeid,
		Amount:    new(big.Int).SetUint64(bctx.FeeConfig().MinStakePledge.Uint64() + constants.LDC),
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer2))
	tt := txData.ToTransaction()
	tt.Timestamp = bs.Timestamp()
	tt.Gas = tt.RequiredGas(bctx.FeeConfig().ThresholdGas)
	keeperGas := tt.Gas
	itx, err := NewTx(tt, true)
	assert.NoError(err)

	keeperAcc := bs.MustAccount(keeper)
	keeperAcc.Add(constants.NativeToken,
		new(big.Int).SetUint64(bctx.FeeConfig().MinStakePledge.Uint64()+constants.LDC*2))
	assert.NoError(itx.Apply(bctx, bs))

	stakeAcc := bs.MustAccount(stakeid)
	assert.Equal((keeperGas)*bctx.Price,
		itx.(*TxCreateStake).ldc.balanceOf(constants.NativeToken).Uint64())
	assert.Equal((keeperGas)*100,
		itx.(*TxCreateStake).miner.balanceOf(constants.NativeToken).Uint64())

	assert.Equal(constants.LDC, stakeAcc.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(bctx.FeeConfig().MinStakePledge.Uint64()+constants.LDC,
		stakeAcc.balanceOfAll(constants.NativeToken).Uint64())
	assert.Equal(uint64(0), stakeAcc.balanceOf(token).Uint64())

	assert.NotNil(stakeAcc.ld.StakeLedger)
	assert.Equal(0, len(stakeAcc.ld.StakeLedger))
	assert.NotNil(stakeAcc.ld.Approver)
	assert.Equal(approver, *stakeAcc.ld.Approver)
	assert.Equal(ld.TxTypes{ld.TypeOpenLending, ld.TypeDestroyStake}, stakeAcc.ld.ApproveList)
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
	txData = &ld.TxData{
		Type:      ld.TypeOpenLending,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      stakeid,
		Data:      ld.MustMarshal(lcfg),
	}
	assert.NoError(txData.SignWith(util.Signer2))
	tt = txData.ToTransaction()
	tt.Timestamp = bs.Timestamp()
	tt.Gas = tt.RequiredGas(bctx.FeeConfig().ThresholdGas)
	itx, err = NewTx(tt, true)
	assert.NoError(err)

	bs.CommitAccounts()
	assert.ErrorContains(itx.Apply(bctx, bs),
		"TxOpenLending.Apply error: invalid signature for approver")
	bs.CheckoutAccounts()

	assert.NoError(txData.SignWith(util.Signer1))
	tt = txData.ToTransaction()
	tt.Timestamp = bs.Timestamp()
	tt.Gas = tt.RequiredGas(bctx.FeeConfig().ThresholdGas)
	stakeGas := tt.Gas
	itx, err = NewTx(tt, true)
	assert.NoError(err)
	assert.NoError(itx.Apply(bctx, bs))

	assert.Equal((keeperGas+stakeGas)*bctx.Price,
		itx.(*TxOpenLending).ldc.balanceOf(constants.NativeToken).Uint64())
	assert.Equal((keeperGas+stakeGas)*100,
		itx.(*TxOpenLending).miner.balanceOf(constants.NativeToken).Uint64())
	assert.NotNil(stakeAcc.ld.Lending)
	assert.NotNil(stakeAcc.ld.LendingLedger)
	assert.Equal(uint64(1), stakeAcc.Nonce())

	// AddNonceTable
	ns := []uint64{bs.Timestamp() + 1, 1, 2, 3}
	ndData, err := util.MarshalCBOR(ns)
	assert.NoError(err)
	txData = &ld.TxData{
		Type:      ld.TypeAddNonceTable,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     1,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      stakeid,
		Data:      ndData,
	}
	assert.NoError(txData.SignWith(util.Signer2))
	tt = txData.ToTransaction()
	tt.Timestamp = bs.Timestamp()
	tt.Gas = tt.RequiredGas(bctx.FeeConfig().ThresholdGas)
	stakeGas += tt.Gas
	itx, err = NewTx(tt, true)
	assert.NoError(err)
	assert.NoError(itx.Apply(bctx, bs))

	assert.Equal((keeperGas+stakeGas)*bctx.Price,
		itx.(*TxAddNonceTable).ldc.balanceOf(constants.NativeToken).Uint64())
	assert.Equal((keeperGas+stakeGas)*100,
		itx.(*TxAddNonceTable).miner.balanceOf(constants.NativeToken).Uint64())
	assert.Equal([]uint64{1, 2, 3}, stakeAcc.ld.NonceTable[bs.Timestamp()+1])
	assert.Equal(uint64(2), stakeAcc.Nonce())

	// Borrow
	tf := &ld.TxTransfer{
		Nonce:  3,
		From:   &stakeid,
		To:     &approver,
		Token:  &token,
		Amount: new(big.Int).SetUint64(constants.LDC),
		Expire: bs.Timestamp() + 1,
	}
	assert.NoError(tf.SyntacticVerify())
	txData = &ld.TxData{
		Type:      ld.TypeBorrow,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      approver,
		To:        &stakeid,
		Token:     &token,
		Data:      tf.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	assert.NoError(txData.ExSignWith(util.Signer2))
	tt = txData.ToTransaction()
	tt.Timestamp = bs.Timestamp()
	tt.Gas = tt.RequiredGas(bctx.FeeConfig().ThresholdGas)
	approverGas := tt.Gas
	itx, err = NewTx(tt, true)
	assert.NoError(err)
	approverAcc := bs.MustAccount(approver)
	approverAcc.Add(constants.NativeToken, new(big.Int).SetUint64(constants.LDC))
	assert.NoError(itx.Apply(bctx, bs))

	assert.Equal((keeperGas+stakeGas+approverGas)*bctx.Price,
		itx.(*TxBorrow).ldc.balanceOf(constants.NativeToken).Uint64())
	assert.Equal((keeperGas+stakeGas+approverGas)*100,
		itx.(*TxBorrow).miner.balanceOf(constants.NativeToken).Uint64())

	assert.Equal([]uint64{1, 2}, stakeAcc.ld.NonceTable[bs.Timestamp()+1])
	assert.Equal(constants.LDC*9, stakeAcc.balanceOf(token).Uint64())
	assert.Equal(constants.LDC, approverAcc.balanceOf(token).Uint64())

	// DestroyStake
	txData = &ld.TxData{
		Type:      ld.TypeDestroyStake,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     2,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      stakeid,
		To:        &keeper,
	}
	assert.NoError(txData.SignWith(util.Signer1))
	assert.NoError(txData.SignWith(util.Signer2))
	tt = txData.ToTransaction()
	tt.Timestamp = bs.Timestamp()
	tt.Gas = tt.RequiredGas(bctx.FeeConfig().ThresholdGas)
	itx, err = NewTx(tt, true)
	assert.NoError(err)
	bs.CommitAccounts()
	assert.ErrorContains(itx.Apply(bctx, bs),
		"TxDestroyStake.Apply error: Account(0x0000000000000000000000000000002354455354).DestroyStake error: please repay all before close")
	bs.CheckoutAccounts()

	// TypeRepay
	txData = &ld.TxData{
		Type:      ld.TypeRepay,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     1,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      approver,
		To:        &stakeid,
		Token:     &token,
		Amount:    new(big.Int).SetUint64(constants.LDC),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	tt = txData.ToTransaction()
	tt.Gas = tt.RequiredGas(bctx.FeeConfig().ThresholdGas)
	approverGas += tt.Gas
	itx, err = NewTx(tt, true)
	assert.NoError(err)
	assert.NoError(itx.Apply(bctx, bs))

	assert.Equal((keeperGas+stakeGas+approverGas)*bctx.Price,
		itx.(*TxRepay).ldc.balanceOf(constants.NativeToken).Uint64())
	assert.Equal((keeperGas+stakeGas+approverGas)*100,
		itx.(*TxRepay).miner.balanceOf(constants.NativeToken).Uint64())

	assert.Equal(constants.LDC*10, stakeAcc.balanceOf(token).Uint64())
	assert.Equal(uint64(0), approverAcc.balanceOf(token).Uint64())

	txData = &ld.TxData{
		Type:      ld.TypeDestroyStake,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     2,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      stakeid,
		To:        &keeper,
	}
	assert.NoError(txData.SignWith(util.Signer1))
	assert.NoError(txData.SignWith(util.Signer2))
	tt = txData.ToTransaction()
	tt.Timestamp = bs.Timestamp()
	tt.Gas = tt.RequiredGas(bctx.FeeConfig().ThresholdGas)
	stakeGas += tt.Gas
	itx, err = NewTx(tt, true)
	assert.NoError(err)
	assert.NoError(itx.Apply(bctx, bs))

	assert.Equal((keeperGas+stakeGas+approverGas)*bctx.Price,
		itx.(*TxDestroyStake).ldc.balanceOf(constants.NativeToken).Uint64())
	assert.Equal((keeperGas+stakeGas+approverGas)*100,
		itx.(*TxDestroyStake).miner.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(bctx.FeeConfig().MinStakePledge.Uint64()+constants.LDC*2-(keeperGas+stakeGas)*(bctx.Price+100),
		keeperAcc.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(constants.LDC*0, stakeAcc.balanceOf(token).Uint64())
	assert.Equal(constants.LDC*10, keeperAcc.balanceOf(token).Uint64())

	assert.Equal(ld.AccountType(0), stakeAcc.ld.Type)
	assert.Equal(uint16(0), stakeAcc.ld.Threshold)
	assert.Equal(uint64(3), stakeAcc.ld.Nonce)
	assert.Equal(util.EthIDs{}, stakeAcc.ld.Keepers)
	assert.Equal(make(map[uint64][]uint64), stakeAcc.ld.NonceTable)
	assert.Nil(stakeAcc.ld.Approver)
	assert.Nil(stakeAcc.ld.ApproveList)
	assert.Nil(stakeAcc.ld.Stake)
	assert.Nil(stakeAcc.ld.StakeLedger)
	assert.Nil(stakeAcc.ld.Lending)
	assert.Nil(stakeAcc.ld.LendingLedger)

	assert.NoError(bs.VerifyState())
}
