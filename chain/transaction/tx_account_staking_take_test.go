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

func TestTxTakeStake(t *testing.T) {
	assert := assert.New(t)

	// SyntacticVerify
	tx := &TxTakeStake{}
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
		Type:      ld.TypeTakeStake,
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
		Type:      ld.TypeTakeStake,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      sender,
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx(txData.ToTransaction(), true)
	assert.ErrorContains(err, "nil to as stake account")

	txData = &ld.TxData{
		Type:      ld.TypeTakeStake,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      sender,
		To:        &stakeid,
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx(txData.ToTransaction(), true)
	assert.ErrorContains(err, "nil amount")

	txData = &ld.TxData{
		Type:      ld.TypeTakeStake,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      sender,
		To:        &stakeid,
		Amount:    new(big.Int).SetUint64(constants.LDC),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx(txData.ToTransaction(), true)
	assert.ErrorContains(err, "invalid data")

	txData = &ld.TxData{
		Type:      ld.TypeTakeStake,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      sender,
		To:        &constants.GenesisAccount,
		Amount:    new(big.Int).SetUint64(constants.LDC),
		Data:      []byte("你好👋"),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx(txData.ToTransaction(), true)
	assert.ErrorContains(err, "invalid stake account 0xFFfFfFffFFfffFFfFFfFFFFFffFFFffffFfFFFfF")

	txData = &ld.TxData{
		Type:      ld.TypeTakeStake,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      sender,
		To:        &stakeid,
		Amount:    new(big.Int).SetUint64(constants.LDC),
		Data:      []byte("你好👋"),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx(txData.ToTransaction(), true)
	assert.ErrorContains(err, "cbor: cannot unmarshal")

	input := &ld.TxTransfer{Nonce: 1}
	txData = &ld.TxData{
		Type:      ld.TypeTakeStake,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      sender,
		To:        &stakeid,
		Amount:    new(big.Int).SetUint64(constants.LDC),
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx(txData.ToTransaction(), true)
	assert.ErrorContains(err, "invalid nonce, expected 1, got 0")

	input = &ld.TxTransfer{
		Nonce: 0,
	}
	txData = &ld.TxData{
		Type:      ld.TypeTakeStake,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      sender,
		To:        &stakeid,
		Amount:    new(big.Int).SetUint64(constants.LDC),
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx(txData.ToTransaction(), true)
	assert.ErrorContains(err, "nil from")

	input = &ld.TxTransfer{
		Nonce: 0,
		From:  &constants.GenesisAccount,
	}
	txData = &ld.TxData{
		Type:      ld.TypeTakeStake,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      sender,
		To:        &stakeid,
		Amount:    new(big.Int).SetUint64(constants.LDC),
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx(txData.ToTransaction(), true)
	assert.ErrorContains(err,
		"invalid from, expected 0xFFfFfFffFFfffFFfFFfFFFFFffFFFffffFfFFFfF, got 0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC")

	input = &ld.TxTransfer{
		Nonce: 0,
		From:  &sender,
	}
	txData = &ld.TxData{
		Type:      ld.TypeTakeStake,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      sender,
		To:        &stakeid,
		Amount:    new(big.Int).SetUint64(constants.LDC),
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx(txData.ToTransaction(), true)
	assert.ErrorContains(err, "nil to")

	input = &ld.TxTransfer{
		Nonce: 0,
		From:  &sender,
		To:    &constants.GenesisAccount,
	}
	txData = &ld.TxData{
		Type:      ld.TypeTakeStake,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      sender,
		To:        &stakeid,
		Amount:    new(big.Int).SetUint64(constants.LDC),
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx(txData.ToTransaction(), true)
	assert.ErrorContains(err,
		"invalid to, expected 0xFFfFfFffFFfffFFfFFfFFFFFffFFFffffFfFFFfF, got 0x0000000000000000000000000000002354455354")

	input = &ld.TxTransfer{
		Nonce: 0,
		From:  &sender,
		To:    &stakeid,
	}
	txData = &ld.TxData{
		Type:      ld.TypeTakeStake,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      sender,
		To:        &stakeid,
		Token:     &token,
		Amount:    new(big.Int).SetUint64(constants.LDC),
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx(txData.ToTransaction(), true)
	assert.ErrorContains(err,
		"invalid token, expected NativeLDC, got $TEST")

	input = &ld.TxTransfer{
		Nonce: 0,
		From:  &sender,
		To:    &stakeid,
		Token: &token,
	}
	txData = &ld.TxData{
		Type:      ld.TypeTakeStake,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      sender,
		To:        &stakeid,
		Amount:    new(big.Int).SetUint64(constants.LDC),
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx(txData.ToTransaction(), true)
	assert.ErrorContains(err,
		"invalid token, expected $TEST, got NativeLDC")

	input = &ld.TxTransfer{
		Nonce: 0,
		From:  &sender,
		To:    &stakeid,
	}
	txData = &ld.TxData{
		Type:      ld.TypeTakeStake,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      sender,
		To:        &stakeid,
		Amount:    new(big.Int).SetUint64(constants.LDC),
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx(txData.ToTransaction(), true)
	assert.ErrorContains(err, "nil amount")

	input = &ld.TxTransfer{
		Nonce:  0,
		From:   &sender,
		To:     &stakeid,
		Amount: new(big.Int).SetUint64(constants.LDC * 10),
	}
	txData = &ld.TxData{
		Type:      ld.TypeTakeStake,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      sender,
		To:        &stakeid,
		Amount:    new(big.Int).SetUint64(constants.LDC),
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx(txData.ToTransaction(), true)
	assert.ErrorContains(err, "invalid amount, expected 10000000000, got 1000000000")

	input = &ld.TxTransfer{
		Nonce:  0,
		From:   &sender,
		To:     &stakeid,
		Amount: new(big.Int).SetUint64(constants.LDC * 10),
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
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	tt := txData.ToTransaction()
	tt.Timestamp = bs.Timestamp()
	_, err = NewTx(tt, true)
	assert.ErrorContains(err, "data expired, expected >= 1000, got 0")

	input = &ld.TxTransfer{
		Nonce:  0,
		From:   &sender,
		To:     &stakeid,
		Amount: new(big.Int).SetUint64(constants.LDC * 10),
		Expire: bs.Timestamp(),
		Data:   util.MustMarshalCBOR("a"),
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
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	tt = txData.ToTransaction()
	tt.Timestamp = bs.Timestamp()
	_, err = NewTx(tt, true)
	assert.ErrorContains(err,
		"invalid lockTime, cbor: cannot unmarshal UTF-8 text string into Go value of type uint64")

	input = &ld.TxTransfer{
		Nonce:  0,
		From:   &sender,
		To:     &stakeid,
		Amount: new(big.Int).SetUint64(constants.LDC * 10),
		Expire: bs.Timestamp(),
		Data:   util.MustMarshalCBOR(bs.Timestamp() + 1),
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
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	tt = txData.ToTransaction()
	tt.Timestamp = bs.Timestamp()
	_, err = NewTx(tt, true)
	assert.ErrorContains(err,
		"invalid exSignatures, Transaction.ExSigners error: DeriveSigners error: no signature")

	assert.NoError(txData.ExSignWith(util.Signer2))
	tt = txData.ToTransaction()
	tt.Timestamp = bs.Timestamp()
	itx, err := NewTx(tt, true)
	assert.NoError(err)

	bs.CommitAccounts()
	assert.ErrorContains(itx.Apply(bctx, bs),
		"invalid gas, expected 654, got 0")
	bs.CheckoutAccounts()

	tt.Gas = tt.RequiredGas(bctx.FeeConfig().ThresholdGas)
	itx, err = NewTx(tt, true)
	assert.NoError(err)
	bs.CommitAccounts()
	assert.ErrorContains(itx.Apply(bctx, bs),
		"Account(0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC).CheckBalance error: insufficient NativeLDC balance, expected 10000719400, got 0")
	bs.CheckoutAccounts()

	senderAcc := bs.MustAccount(sender)
	senderAcc.Add(constants.NativeToken, new(big.Int).SetUint64(constants.LDC*11))
	bs.CommitAccounts()
	assert.ErrorContains(itx.Apply(bctx, bs),
		"Account(0x0000000000000000000000000000002354455354).TakeStake error: invalid stake account")
	bs.CheckoutAccounts()

	scfg := &ld.StakeConfig{
		LockTime:    0,
		WithdrawFee: 100_000,
		MinAmount:   new(big.Int).SetUint64(constants.LDC * 10),
		MaxAmount:   new(big.Int).SetUint64(constants.LDC * 100),
	}
	sinput := &ld.TxAccounter{
		Threshold: ld.Uint16Ptr(1),
		Keepers:   &util.EthIDs{util.Signer2.Address()},
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

	stakeAcc, err := bs.LoadAccount(stakeid)

	assert.Equal(keeperGas*bctx.Price,
		itx.(*TxCreateStake).ldc.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(keeperGas*100,
		itx.(*TxCreateStake).miner.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(constants.LDC*0, stakeAcc.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(bctx.FeeConfig().MinStakePledge.Uint64(),
		stakeAcc.balanceOfAll(constants.NativeToken).Uint64())
	assert.Equal(constants.LDC-keeperGas*(bctx.Price+100),
		keeperAcc.balanceOf(constants.NativeToken).Uint64())

	assert.Nil(stakeAcc.ld.Approver)
	assert.Equal(ld.StakeAccount, stakeAcc.ld.Type)
	assert.Nil(stakeAcc.ld.MaxTotalSupply)
	assert.NotNil(stakeAcc.ld.Stake)
	assert.NotNil(stakeAcc.ld.StakeLedger)
	assert.Nil(stakeAcc.ld.StakeLedger[sender])
	keeperEntry := stakeAcc.ld.StakeLedger[keeper]
	assert.NotNil(keeperEntry)
	assert.Equal(bctx.FeeConfig().MinStakePledge.Uint64(), keeperEntry.Amount.Uint64())
	assert.Equal(uint64(0), keeperEntry.LockTime)
	assert.Nil(keeperEntry.Approver)

	input = &ld.TxTransfer{
		Nonce:  0,
		From:   &sender,
		To:     &stakeid,
		Amount: new(big.Int).SetUint64(constants.LDC * 10),
		Expire: bs.Timestamp(),
		Data:   util.MustMarshalCBOR(bs.Timestamp() + 1),
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
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	assert.NoError(txData.ExSignWith(util.Signer2))
	tt = txData.ToTransaction()
	tt.Gas = tt.RequiredGas(bctx.FeeConfig().ThresholdGas)
	senderGas := tt.Gas
	tt.Timestamp = bs.Timestamp()
	itx, err = NewTx(tt, true)
	assert.NoError(err)
	assert.NoError(itx.Apply(bctx, bs))

	assert.Equal((keeperGas+senderGas)*bctx.Price,
		itx.(*TxTakeStake).ldc.balanceOf(constants.NativeToken).Uint64())
	assert.Equal((keeperGas+senderGas)*100,
		itx.(*TxTakeStake).miner.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(constants.LDC*10, stakeAcc.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(bctx.FeeConfig().MinStakePledge.Uint64()+constants.LDC*10,
		stakeAcc.balanceOfAll(constants.NativeToken).Uint64())
	assert.Equal(constants.LDC-senderGas*(bctx.Price+100),
		senderAcc.balanceOf(constants.NativeToken).Uint64())
	senderEntry := stakeAcc.ld.StakeLedger[sender]
	assert.NotNil(senderEntry)
	assert.Equal(constants.LDC*10, senderEntry.Amount.Uint64())
	assert.Equal(bs.Timestamp()+1, senderEntry.LockTime)
	assert.Nil(senderEntry.Approver)
	keeperEntry = stakeAcc.ld.StakeLedger[keeper]
	assert.Equal(bctx.FeeConfig().MinStakePledge.Uint64(), keeperEntry.Amount.Uint64())

	jsondata, err := itx.MarshalJSON()
	assert.NoError(err)
	// fmt.Println(string(jsondata))
	assert.Equal(`{"type":"TypeTakeStake","chainID":2357,"nonce":0,"gasTip":100,"gasFeeCap":1000,"from":"0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC","to":"0x0000000000000000000000000000002354455354","amount":10000000000,"data":{"from":"0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC","to":"0x0000000000000000000000000000002354455354","amount":10000000000,"expire":1000,"data":"0x1903e91a5af090"},"signatures":["230f5220839b3cf7f92fe6ea65c0c8cfdbeaa992f519ea583adbfff51725eb03721f5d6cdff64aafe7e1fada8391c8e017bf4ada63dc0bf0cf5954b45e64e63b00"],"exSignatures":["54b5fa755a0bd4e82c9f561f4a7493a647d1b114f4b48c62a4b95a5e82bb16dc65b5179a81109c14180b5c457b5fae91d1126ae935bf903ec1c03b68eb8b048300"],"gas":654,"id":"2Lohph5mLZZabmMo32G6uaHfoRDkjDsLzZpweprLvZRMvxVE6z"}`, string(jsondata))
}
