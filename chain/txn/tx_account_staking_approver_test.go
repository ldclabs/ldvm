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

func TestTxUpdateStakeApprover(t *testing.T) {
	assert := assert.New(t)

	// SyntacticVerify
	tx := &TxUpdateStakeApprover{}
	assert.ErrorContains(tx.SyntacticVerify(), "nil pointer")
	_, err := tx.MarshalJSON()
	require.NoError(t, err)

	ctx := NewMockChainContext()
	cs := ctx.MockChainState()
	stake := ld.MustNewStake("#TEST")
	stakeid := stake.Address()
	token := ld.MustNewToken("$TEST")

	sender := signer.Signer1.Key().Address()
	approver := signer.Signer2.Key()

	ltx := &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeUpdateStakeApprover,
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
		Type:      ld.TypeUpdateStakeApprover,
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
		Type:      ld.TypeUpdateStakeApprover,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      sender,
		To:        &stakeid,
		Amount:    new(big.Int).SetUint64(100),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "invalid amount, should be nil")

	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeUpdateStakeApprover,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      sender,
		To:        &stakeid,
		Token:     token.Ptr(),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "invalid token, should be nil")

	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeUpdateStakeApprover,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      sender,
		To:        &stakeid,
		Data:      []byte{},
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.ErrorContains(ltx.SyntacticVerify(), "TxData.SyntacticVerify: empty data")

	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeUpdateStakeApprover,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      sender,
		To:        ids.GenesisAccount.Ptr(),
		Data:      []byte("你好👋"),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "invalid stake account 0xFFfFFFfFfffFFfFFffFFFfFfFffFFFfffFfFFFff")

	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeUpdateStakeApprover,
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

	input := &ld.TxAccounter{}
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeUpdateStakeApprover,
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
	assert.ErrorContains(err, "nil approver")

	input = &ld.TxAccounter{Approver: &approver, ApproveList: &ld.AccountTxTypes}
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeUpdateStakeApprover,
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
	assert.ErrorContains(err, "invalid approveList, should be nil")

	input = &ld.TxAccounter{Approver: &approver}
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeUpdateStakeApprover,
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
	require.NoError(t, err)

	cs.CommitAccounts()
	assert.ErrorContains(itx.Apply(ctx, cs),
		"insufficient NativeLDC balance, expected 1009800, got 0")
	cs.CheckoutAccounts()

	senderAcc := cs.MustAccount(sender)
	senderAcc.Add(ids.NativeToken, new(big.Int).SetUint64(unit.LDC*1002))

	cs.CommitAccounts()
	assert.ErrorContains(itx.Apply(ctx, cs),
		"invalid stake account")
	cs.CheckoutAccounts()

	scfg := &ld.StakeConfig{
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
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      sender,
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

	stakeAcc := cs.MustAccount(stakeid)

	senderGas := ltx.Gas()
	tx2 := itx.(*TxCreateStake)

	assert.Equal(senderGas*ctx.Price, tx2.ldc.Balance().Uint64())
	assert.Equal(senderGas*100, tx2.miner.Balance().Uint64())
	assert.Equal(unit.LDC*0, stakeAcc.Balance().Uint64())
	assert.Equal(unit.LDC*1000, stakeAcc.BalanceOfAll(ids.NativeToken).Uint64())
	assert.Equal(unit.LDC-senderGas*(ctx.Price+100), senderAcc.Balance().Uint64())

	assert.Nil(stakeAcc.LD().Approver)
	assert.Equal(ld.StakeAccount, stakeAcc.LD().Type)
	assert.Nil(stakeAcc.LD().MaxTotalSupply)
	require.NotNil(t, stakeAcc.LD().Stake)
	require.NotNil(t, stakeAcc.Ledger())
	assert.Nil(stakeAcc.Ledger().Stake[sender.AsKey()])

	input = &ld.TxAccounter{Approver: &approver}
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeUpdateStakeApprover,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     1,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      sender,
		To:        &stakeid,
		Data:      input.Bytes(),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	itx, err = NewTx(ltx)
	require.NoError(t, err)

	cs.CommitAccounts()
	assert.ErrorContains(itx.Apply(ctx, cs),
		"0x8db97c7cECe249C2b98bdc0226cc4C2A57bF52fc has no stake ledger to update")
	cs.CheckoutAccounts()

	stakeAcc.Ledger().Stake[sender.AsKey()] = &ld.StakeEntry{Amount: new(big.Int).SetUint64(unit.LDC)}
	assert.NoError(itx.Apply(ctx, cs))

	senderGas += ltx.Gas()
	assert.Equal(senderGas*ctx.Price,
		itx.(*TxUpdateStakeApprover).ldc.Balance().Uint64())
	assert.Equal(senderGas*100,
		itx.(*TxUpdateStakeApprover).miner.Balance().Uint64())
	assert.Equal(unit.LDC*0, stakeAcc.Balance().Uint64())
	assert.Equal(unit.LDC*1000, stakeAcc.BalanceOfAll(ids.NativeToken).Uint64())
	assert.Equal(unit.LDC-senderGas*(ctx.Price+100),
		senderAcc.Balance().Uint64())

	require.NotNil(t, stakeAcc.Ledger().Stake[sender.AsKey()])
	assert.Equal(approver, *stakeAcc.Ledger().Stake[sender.AsKey()].Approver)

	jsondata, err := itx.MarshalJSON()
	require.NoError(t, err)
	// fmt.Println(string(jsondata))
	assert.Equal(`{"tx":{"type":"TypeUpdateStakeApprover","chainID":2357,"nonce":1,"gasTip":100,"gasFeeCap":1000,"from":"0x8db97c7cECe249C2b98bdc0226cc4C2A57bF52fc","to":"0x0000000000000000000000000000002354455354","data":{"approver":"RBccN_9de3u43K1cgfFihKIp5kE1lmGG"}},"sigs":["pgW4Z590E7jhuKH2N-Og_r8YVu3rylsDfWf02DVUAUFi2ZeaZk6s_vK-l6W8q-G3OxYmXrCfr0umGcxfH6ThvwDeg2W1"],"id":"ieVICnmp0R08Pl5mgJD0dNYeC4Xh64YI_IER0ZNxSgJLFQBN"}`, string(jsondata))

	// clear Approver but need approver signing
	input = &ld.TxAccounter{Approver: &signer.Key{}}
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeUpdateStakeApprover,
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
	itx, err = NewTx(ltx)
	require.NoError(t, err)

	cs.CommitAccounts()
	assert.ErrorContains(itx.Apply(ctx, cs),
		"0x8db97c7cECe249C2b98bdc0226cc4C2A57bF52fc need approver signing")
	cs.CheckoutAccounts()

	assert.NoError(ltx.SignWith(signer.Signer1, signer.Signer2))
	assert.NoError(ltx.SyntacticVerify())
	itx, err = NewTx(ltx)
	require.NoError(t, err)
	assert.NoError(itx.Apply(ctx, cs))

	senderGas += ltx.Gas()
	assert.Equal(senderGas*ctx.Price,
		itx.(*TxUpdateStakeApprover).ldc.Balance().Uint64())
	assert.Equal(senderGas*100,
		itx.(*TxUpdateStakeApprover).miner.Balance().Uint64())
	assert.Equal(unit.LDC*0, stakeAcc.Balance().Uint64())
	assert.Equal(unit.LDC*1000, stakeAcc.BalanceOfAll(ids.NativeToken).Uint64())
	assert.Equal(unit.LDC-senderGas*(ctx.Price+100),
		senderAcc.Balance().Uint64())

	require.NotNil(t, stakeAcc.Ledger().Stake[sender.AsKey()])
	assert.Nil(stakeAcc.Ledger().Stake[sender.AsKey()].Approver)

	jsondata, err = itx.MarshalJSON()
	require.NoError(t, err)
	// fmt.Println(string(jsondata))
	assert.Equal(`{"tx":{"type":"TypeUpdateStakeApprover","chainID":2357,"nonce":2,"gasTip":100,"gasFeeCap":1000,"from":"0x8db97c7cECe249C2b98bdc0226cc4C2A57bF52fc","to":"0x0000000000000000000000000000002354455354","data":{"approver":"p__G-A"}},"sigs":["OER5mnDmbqgBFckuLwU2nU_GeETVpWlvGQN7yxBw4DQpDoCNGBOm2-6ERLMJzqaSqOsKnNazHYSrU6GKfyGoEQF1874F","dQBSBvPAn3uOCjsSsQDJmXNcvWmHvBtaqzf7WmR9Ggo19R0NFqbMtRwYATl-QuaP4EfZ2toec_N9u30JuQ1msABZ2ZH1"],"id":"q6jDVcY7UKEkThILjGqSBAbmGfJyVDq90e2VS4XWjG2XzoZu"}`, string(jsondata))

	assert.NoError(cs.VerifyState())
}
