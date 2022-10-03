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

func TestTxCreateStake(t *testing.T) {
	assert := assert.New(t)

	// SyntacticVerify
	tx := &TxCreateStake{}
	assert.ErrorContains(tx.SyntacticVerify(), "nil pointer")
	_, err := tx.MarshalJSON()
	assert.NoError(err)

	ctx := NewMockChainContext()
	cs := ctx.MockChainState()
	stake := ld.MustNewStake("#TEST")
	stakeid := util.EthID(stake)
	token := ld.MustNewToken("$TEST")

	from, err := cs.LoadAccount(util.Signer1.Address())
	assert.NoError(err)
	approver := util.Signer2.Address()

	ltx := &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeCreateStake,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      from.id,
	}}
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "DeriveSigners error: no signature")

	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeCreateStake,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      from.id,
	}}
	assert.NoError(ltx.SignWith(util.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "nil to as stake account")

	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeCreateStake,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      from.id,
		To:        &stakeid,
	}}
	assert.NoError(ltx.SignWith(util.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "nil amount")

	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeCreateStake,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      from.id,
		To:        &stakeid,
		Amount:    new(big.Int).SetUint64(100),
		Token:     &token,
	}}
	assert.NoError(ltx.SignWith(util.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "invalid token, should be nil")

	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeCreateStake,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      from.id,
		To:        &stakeid,
		Amount:    new(big.Int).SetUint64(100),
	}}
	assert.NoError(ltx.SignWith(util.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "invalid data")

	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeCreateStake,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      from.id,
		To:        &stakeid,
		Amount:    new(big.Int).SetUint64(100),
		Data:      []byte("你好👋"),
	}}
	assert.NoError(ltx.SignWith(util.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "cbor: unexpected following extraneous data")

	input := &ld.TxAccounter{}
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeCreateStake,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      from.id,
		To:        &approver,
		Amount:    new(big.Int).SetUint64(100),
		Data:      input.Bytes(),
	}}
	assert.NoError(ltx.SignWith(util.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "invalid stake account 0x44171C37Ff5D7B7bb8dcad5C81f16284A229e641")

	input = &ld.TxAccounter{}
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeCreateStake,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      from.id,
		To:        &stakeid,
		Amount:    new(big.Int).SetUint64(100),
		Data:      input.Bytes(),
	}}
	assert.NoError(ltx.SignWith(util.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "invalid threshold, expected >= 1")

	input = &ld.TxAccounter{
		Threshold: ld.Uint16Ptr(0),
		Keepers:   &util.EthIDs{util.Signer1.Address()},
	}
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeCreateStake,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      from.id,
		To:        &stakeid,
		Amount:    new(big.Int).SetUint64(100),
		Data:      input.Bytes(),
	}}
	assert.NoError(ltx.SignWith(util.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "invalid threshold, expected >= 1")

	input = &ld.TxAccounter{
		Threshold: ld.Uint16Ptr(1),
		Keepers:   &util.EthIDs{util.Signer1.Address()},
		Amount:    new(big.Int).SetUint64(constants.LDC * 10),
	}
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeCreateStake,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      from.id,
		To:        &stakeid,
		Amount:    new(big.Int).SetUint64(100),
		Data:      input.Bytes(),
	}}
	assert.NoError(ltx.SignWith(util.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "invalid amount, should be nil")

	input = &ld.TxAccounter{
		Threshold: ld.Uint16Ptr(1),
		Keepers:   &util.EthIDs{util.Signer1.Address()},
		Approver:  &util.EthIDEmpty,
	}
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeCreateStake,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      from.id,
		To:        &stakeid,
		Amount:    new(big.Int).SetUint64(100),
		Data:      input.Bytes(),
	}}
	assert.NoError(ltx.SignWith(util.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err,
		"invalid approver, expected not 0x0000000000000000000000000000000000000000")

	input = &ld.TxAccounter{
		Threshold: ld.Uint16Ptr(1),
		Keepers:   &util.EthIDs{util.Signer1.Address()},
		Approver:  &approver,
	}
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeCreateStake,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      from.id,
		To:        &stakeid,
		Amount:    new(big.Int).SetUint64(100),
		Data:      input.Bytes(),
	}}
	assert.NoError(ltx.SignWith(util.Signer1))
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
		Keepers:   &util.EthIDs{util.Signer1.Address()},
		Approver:  &approver,
		Data:      ld.MustMarshal(scfg),
	}
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeCreateStake,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      from.id,
		To:        &stakeid,
		Amount:    new(big.Int).SetUint64(100),
		Data:      input.Bytes(),
	}}
	assert.NoError(ltx.SignWith(util.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	ltx.Timestamp = cs.Timestamp()
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "invalid lockTime, expected 0 or >= 1000")

	scfg = &ld.StakeConfig{
		LockTime:    cs.Timestamp() + 1000,
		WithdrawFee: 100_000,
		MinAmount:   new(big.Int).SetUint64(constants.LDC),
		MaxAmount:   new(big.Int).SetUint64(constants.LDC),
	}
	input = &ld.TxAccounter{
		Threshold: ld.Uint16Ptr(1),
		Keepers:   &util.EthIDs{util.Signer1.Address()},
		Approver:  &approver,
		Data:      ld.MustMarshal(scfg),
	}
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeCreateStake,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      from.id,
		To:        &stakeid,
		Amount:    new(big.Int).SetUint64(100),
		Data:      input.Bytes(),
	}}
	assert.NoError(ltx.SignWith(util.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	ltx.Timestamp = cs.Timestamp()
	itx, err := NewTx(ltx)
	assert.NoError(err)

	cs.CommitAccounts()
	assert.ErrorContains(itx.Apply(ctx, cs), "insufficient NativeLDC balance, expected 2378300, got 0")
	cs.CheckoutAccounts()

	from.Add(constants.NativeToken, new(big.Int).SetUint64(constants.LDC))
	cs.CommitAccounts()
	assert.ErrorContains(itx.Apply(ctx, cs), "invalid amount, expected >= 1000000000000, got 100")
	cs.CheckoutAccounts()

	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeCreateStake,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      from.id,
		To:        &stakeid,
		Amount:    new(big.Int).SetUint64(constants.LDC * 1001),
		Data:      input.Bytes(),
	}}
	assert.NoError(ltx.SignWith(util.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	ltx.Timestamp = cs.Timestamp()
	itx, err = NewTx(ltx)
	assert.NoError(err)
	cs.CommitAccounts()
	assert.ErrorContains(itx.Apply(ctx, cs),
		"insufficient NativeLDC balance, expected 1001002403500, got 1000000000")
	cs.CheckoutAccounts()

	from.Add(constants.NativeToken, new(big.Int).SetUint64(constants.LDC*1001))
	assert.NoError(itx.Apply(ctx, cs))

	stakeAcc, err := cs.LoadAccount(stakeid)
	assert.NoError(err)
	ldc, err := cs.LoadAccount(constants.LDCAccount)
	assert.NoError(err)
	miner, err := cs.LoadMiner(ctx.Miner())
	assert.NoError(err)

	assert.Equal(ltx.Gas()*ctx.Price,
		ldc.Balance().Uint64())
	assert.Equal(ltx.Gas()*100,
		miner.Balance().Uint64())
	assert.Equal(constants.LDC, stakeAcc.Balance().Uint64())
	assert.Equal(constants.LDC*1001, stakeAcc.balanceOfAll(constants.NativeToken).Uint64())
	assert.Equal(constants.LDC-ltx.Gas()*(ctx.Price+100),
		from.Balance().Uint64())

	assert.Equal(uint64(0), stakeAcc.Nonce())
	assert.Equal(uint16(1), stakeAcc.Threshold())
	assert.Equal(util.EthIDs{util.Signer1.Address()}, stakeAcc.Keepers())
	assert.Equal(approver, *stakeAcc.ld.Approver)
	assert.Equal(ld.StakeAccount, stakeAcc.ld.Type)
	assert.Nil(stakeAcc.ld.MaxTotalSupply)
	assert.NotNil(stakeAcc.ld.Stake)
	assert.NotNil(stakeAcc.ledger)
	assert.NotNil(stakeAcc.ledger.Stake[from.id.AsKey()])
	assert.Equal(constants.LDC*1000, stakeAcc.ledger.Stake[from.id.AsKey()].Amount.Uint64())

	jsondata, err := itx.MarshalJSON()
	assert.NoError(err)
	// fmt.Println(string(jsondata))
	assert.Equal(`{"tx":{"type":"TypeCreateStake","chainID":2357,"nonce":0,"gasTip":100,"gasFeeCap":1000,"from":"0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC","to":"0x0000000000000000000000000000002354455354","amount":1001000000000,"data":{"threshold":1,"keepers":["0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC"],"approver":"0x44171C37Ff5D7B7bb8dcad5C81f16284A229e641","data":"0x86540000000000000000000000000000000000000000001907d01a000186a0c2443b9aca00c2443b9aca004cc0e681"}},"sigs":["cd310d21b3eeec3e8e05d0215c5899c354981049dbeaf3607821c58c6eb47b5059f56e810200c51a8095866d3caa26a2171db6e4aaffd6f9eee3686b4def337d00"],"id":"UVvExSbSormjR21DTbB66cphbPeRCqqeb3G6rfXMdwNdaKVec"}`, string(jsondata))

	// create again
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeCreateStake,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     1,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      from.id,
		To:        &stakeid,
		Amount:    new(big.Int).SetUint64(constants.LDC * 1001),
		Data:      input.Bytes(),
	}}
	assert.NoError(ltx.SignWith(util.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	ltx.Timestamp = cs.Timestamp()
	itx, err = NewTx(ltx)
	assert.NoError(err)
	from.Add(constants.NativeToken, new(big.Int).SetUint64(constants.LDC*1001))

	cs.CommitAccounts()
	assert.ErrorContains(itx.Apply(ctx, cs),
		"Account(0x0000000000000000000000000000002354455354).CreateStake error: stake account #TEST exists")
	cs.CheckoutAccounts()

	// destroy and create again
	ctx.timestamp += 1001
	stakeAcc.ld.Timestamp = cs.Timestamp()
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeDestroyStake,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      stakeid,
		To:        &from.id,
	}}
	assert.NoError(ltx.SignWith(util.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	ltx.Timestamp = cs.Timestamp()
	itx, err = NewTx(ltx)
	assert.NoError(err)

	cs.CommitAccounts()
	assert.ErrorContains(itx.Apply(ctx, cs),
		"TxDestroyStake.Apply error: invalid signature for approver")
	cs.CheckoutAccounts()

	assert.NoError(ltx.SignWith(util.Signer1, util.Signer2))
	assert.NoError(ltx.SyntacticVerify())
	ltx.Timestamp = cs.Timestamp()
	itx, err = NewTx(ltx)
	assert.NoError(err)
	assert.NoError(itx.Apply(ctx, cs))

	assert.Equal(uint64(1), stakeAcc.Nonce())
	assert.Equal(uint16(0), stakeAcc.Threshold())
	assert.Equal(util.EthIDs{}, stakeAcc.Keepers())
	assert.Nil(stakeAcc.ld.Approver)
	assert.Equal(ld.NativeAccount, stakeAcc.ld.Type)
	assert.Nil(stakeAcc.ld.MaxTotalSupply)
	assert.Nil(stakeAcc.ld.Stake)
	assert.Equal(0, len(stakeAcc.ledger.Stake))
	assert.Equal(uint64(0), stakeAcc.balanceOfAll(constants.NativeToken).Uint64())

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
		Keepers:   &util.EthIDs{util.Signer1.Address()},
		Data:      ld.MustMarshal(scfg),
	}
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeCreateStake,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     1,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      from.id,
		To:        &stakeid,
		Amount:    new(big.Int).SetUint64(constants.LDC * 1000),
		Data:      input.Bytes(),
	}}
	assert.NoError(ltx.SignWith(util.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	ltx.Timestamp = cs.Timestamp()
	itx, err = NewTx(ltx)
	assert.NoError(err)
	assert.NoError(itx.Apply(ctx, cs))

	assert.Equal(constants.LDC*0, stakeAcc.Balance().Uint64())
	assert.Equal(constants.LDC*1000, stakeAcc.balanceOfAll(constants.NativeToken).Uint64())

	assert.Equal(uint64(1), stakeAcc.Nonce())
	assert.Equal(uint16(1), stakeAcc.Threshold())
	assert.Equal(util.EthIDs{util.Signer1.Address()}, stakeAcc.Keepers())
	assert.Nil(stakeAcc.ld.Approver)
	assert.Equal(ld.StakeAccount, stakeAcc.ld.Type)
	assert.Nil(stakeAcc.ld.MaxTotalSupply)
	assert.NotNil(stakeAcc.ld.Stake)
	assert.NotNil(stakeAcc.ledger.Stake)
	assert.Nil(stakeAcc.ledger.Stake[from.id.AsKey()])
	assert.NotNil(stakeAcc.ld.Tokens[token.AsKey()])
	assert.Equal(uint64(0), stakeAcc.ld.Tokens[token.AsKey()].Uint64())

	assert.NoError(cs.VerifyState())
}
