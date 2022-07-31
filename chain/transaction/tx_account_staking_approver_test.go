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

func TestTxUpdateStakeApprover(t *testing.T) {
	assert := assert.New(t)

	// SyntacticVerify
	tx := &TxUpdateStakeApprover{}
	assert.ErrorContains(tx.SyntacticVerify(), "nil pointer")
	_, err := tx.MarshalJSON()
	assert.NoError(err)

	bctx := NewMockBCtx()
	bs := bctx.MockBS()
	stake := ld.MustNewStake("#TEST")
	stakeid := stake.EthID()
	token := ld.MustNewToken("$TEST")

	sender := util.Signer1.Address()
	approver := util.Signer2.Address()

	txData := &ld.TxData{
		Type:      ld.TypeUpdateStakeApprover,
		ChainID:   bctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      sender,
	}
	assert.NoError(txData.SyntacticVerify())
	_, err = NewTx2(txData.ToTransaction())
	assert.ErrorContains(err, "DeriveSigners error: no signature")

	txData = &ld.TxData{
		Type:      ld.TypeUpdateStakeApprover,
		ChainID:   bctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      sender,
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx2(txData.ToTransaction())
	assert.ErrorContains(err, "nil to as stake account")

	txData = &ld.TxData{
		Type:      ld.TypeUpdateStakeApprover,
		ChainID:   bctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      sender,
		To:        &stakeid,
		Amount:    new(big.Int).SetUint64(100),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx2(txData.ToTransaction())
	assert.ErrorContains(err, "invalid amount, should be nil")

	txData = &ld.TxData{
		Type:      ld.TypeUpdateStakeApprover,
		ChainID:   bctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      sender,
		To:        &stakeid,
		Token:     &token,
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx2(txData.ToTransaction())
	assert.ErrorContains(err, "invalid token, should be nil")

	txData = &ld.TxData{
		Type:      ld.TypeUpdateStakeApprover,
		ChainID:   bctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      sender,
		To:        &stakeid,
		Data:      []byte{},
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx2(txData.ToTransaction())
	assert.ErrorContains(err, "TxData.SyntacticVerify error: empty data")

	txData = &ld.TxData{
		Type:      ld.TypeUpdateStakeApprover,
		ChainID:   bctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      sender,
		To:        &constants.GenesisAccount,
		Data:      []byte("你好👋"),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx2(txData.ToTransaction())
	assert.ErrorContains(err, "invalid stake account 0xFFfFfFffFFfffFFfFFfFFFFFffFFFffffFfFFFfF")

	txData = &ld.TxData{
		Type:      ld.TypeUpdateStakeApprover,
		ChainID:   bctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      sender,
		To:        &stakeid,
		Data:      []byte("你好👋"),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx2(txData.ToTransaction())
	assert.ErrorContains(err, "cbor: cannot unmarshal")

	input := &ld.TxAccounter{}
	txData = &ld.TxData{
		Type:      ld.TypeUpdateStakeApprover,
		ChainID:   bctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      sender,
		To:        &stakeid,
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx2(txData.ToTransaction())
	assert.ErrorContains(err, "nil approver")

	input = &ld.TxAccounter{Approver: &approver, ApproveList: ld.AccountTxTypes}
	txData = &ld.TxData{
		Type:      ld.TypeUpdateStakeApprover,
		ChainID:   bctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      sender,
		To:        &stakeid,
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx2(txData.ToTransaction())
	assert.ErrorContains(err, "invalid approveList, should be nil")

	input = &ld.TxAccounter{Approver: &approver}
	txData = &ld.TxData{
		Type:      ld.TypeUpdateStakeApprover,
		ChainID:   bctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      sender,
		To:        &stakeid,
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	tt := txData.ToTransaction()
	itx, err := NewTx2(tt)
	assert.NoError(err)

	bs.CommitAccounts()
	assert.ErrorContains(itx.Apply(bctx, bs),
		"insufficient NativeLDC balance, expected 1316700, got 0")
	bs.CheckoutAccounts()

	senderAcc := bs.MustAccount(sender)
	senderAcc.Add(constants.NativeToken, new(big.Int).SetUint64(constants.LDC*1001))

	bs.CommitAccounts()
	assert.ErrorContains(itx.Apply(bctx, bs),
		"Account(0x0000000000000000000000000000002354455354).UpdateStakeApprover error: invalid stake account")
	bs.CheckoutAccounts()

	scfg := &ld.StakeConfig{
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
	txData = &ld.TxData{
		Type:      ld.TypeCreateStake,
		ChainID:   bctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      sender,
		To:        &stakeid,
		Amount:    new(big.Int).SetUint64(constants.LDC * 1000),
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	tt = txData.ToTransaction()
	tt.Timestamp = bs.Timestamp()
	itx, err = NewTx2(tt)
	assert.NoError(err)
	assert.NoError(itx.Apply(bctx, bs))

	stakeAcc := bs.MustAccount(stakeid)

	senderGas := tt.Gas()
	tx2 := itx.(*TxCreateStake)
	assert.Equal(senderGas*bctx.Price,
		tx2.ldc.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(senderGas*100,
		tx2.miner.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(constants.LDC*0, stakeAcc.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(constants.LDC*1000, stakeAcc.balanceOfAll(constants.NativeToken).Uint64())
	assert.Equal(constants.LDC-senderGas*(bctx.Price+100),
		senderAcc.balanceOf(constants.NativeToken).Uint64())

	assert.Nil(stakeAcc.ld.Approver)
	assert.Equal(ld.StakeAccount, stakeAcc.ld.Type)
	assert.Nil(stakeAcc.ld.MaxTotalSupply)
	assert.NotNil(stakeAcc.ld.Stake)
	assert.NotNil(stakeAcc.ledger)
	assert.Nil(stakeAcc.ledger.Stake[sender.AsKey()])

	input = &ld.TxAccounter{Approver: &approver}
	txData = &ld.TxData{
		Type:      ld.TypeUpdateStakeApprover,
		ChainID:   bctx.ChainConfig().ChainID,
		Nonce:     1,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      sender,
		To:        &stakeid,
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	tt = txData.ToTransaction()
	itx, err = NewTx2(tt)
	assert.NoError(err)

	bs.CommitAccounts()
	assert.ErrorContains(itx.Apply(bctx, bs),
		"Account(0x0000000000000000000000000000002354455354).UpdateStakeApprover error: 0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC has no stake ledger to update")
	bs.CheckoutAccounts()

	stakeAcc.ledger.Stake[sender.AsKey()] = &ld.StakeEntry{Amount: new(big.Int).SetUint64(constants.LDC)}
	assert.NoError(itx.Apply(bctx, bs))

	senderGas += tt.Gas()
	assert.Equal(senderGas*bctx.Price,
		itx.(*TxUpdateStakeApprover).ldc.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(senderGas*100,
		itx.(*TxUpdateStakeApprover).miner.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(constants.LDC*0, stakeAcc.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(constants.LDC*1000, stakeAcc.balanceOfAll(constants.NativeToken).Uint64())
	assert.Equal(constants.LDC-senderGas*(bctx.Price+100),
		senderAcc.balanceOf(constants.NativeToken).Uint64())

	assert.NotNil(stakeAcc.ledger.Stake[sender.AsKey()])
	assert.Equal(approver, *stakeAcc.ledger.Stake[sender.AsKey()].Approver)

	jsondata, err := itx.MarshalJSON()
	assert.NoError(err)
	// fmt.Println(string(jsondata))
	assert.Equal(`{"type":"TypeUpdateStakeApprover","chainID":2357,"nonce":1,"gasTip":100,"gasFeeCap":1000,"from":"0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC","to":"0x0000000000000000000000000000002354455354","data":{"approver":"0x44171C37Ff5D7B7bb8dcad5C81f16284A229e641"},"signatures":["a605b8679f7413b8e1b8a1f637e3a0febf1856edebca5b037d67f4d83554014162d9979a664eacfef2be97a5bcabe1b73b16265eb09faf4ba619cc5f1fa4e1bf00"],"id":"HrZYtiYDqeQxGN8jpz8HgPi7rZgGw71UHHLnRKqzRz8DvS2V5"}`, string(jsondata))

	// clear Approver but need approver signing
	input = &ld.TxAccounter{Approver: &util.EthIDEmpty}
	txData = &ld.TxData{
		Type:      ld.TypeUpdateStakeApprover,
		ChainID:   bctx.ChainConfig().ChainID,
		Nonce:     2,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      sender,
		To:        &stakeid,
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	tt = txData.ToTransaction()
	itx, err = NewTx2(tt)
	assert.NoError(err)

	bs.CommitAccounts()
	assert.ErrorContains(itx.Apply(bctx, bs),
		"Account(0x0000000000000000000000000000002354455354).UpdateStakeApprover error: 0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC need approver signing")
	bs.CheckoutAccounts()

	assert.NoError(txData.SignWith(util.Signer2))
	tt = txData.ToTransaction()
	itx, err = NewTx2(tt)
	assert.NoError(err)
	assert.NoError(itx.Apply(bctx, bs))

	senderGas += tt.Gas()
	assert.Equal(senderGas*bctx.Price,
		itx.(*TxUpdateStakeApprover).ldc.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(senderGas*100,
		itx.(*TxUpdateStakeApprover).miner.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(constants.LDC*0, stakeAcc.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(constants.LDC*1000, stakeAcc.balanceOfAll(constants.NativeToken).Uint64())
	assert.Equal(constants.LDC-senderGas*(bctx.Price+100),
		senderAcc.balanceOf(constants.NativeToken).Uint64())

	assert.NotNil(stakeAcc.ledger.Stake[sender.AsKey()])
	assert.Nil(stakeAcc.ledger.Stake[sender.AsKey()].Approver)

	jsondata, err = itx.MarshalJSON()
	assert.NoError(err)
	// fmt.Println(string(jsondata))
	assert.Equal(`{"type":"TypeUpdateStakeApprover","chainID":2357,"nonce":2,"gasTip":100,"gasFeeCap":1000,"from":"0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC","to":"0x0000000000000000000000000000002354455354","data":{"approver":"0x0000000000000000000000000000000000000000"},"signatures":["b4f9e8762a9290f7c36d86a7f187086017b397ad21497bc2ec3137be8eede9616bedc5fc0bc79eda0cbfa532bf484247b2850c018ec1dde0acb8e43c07b3bbfb01","070ddb6c1e5ced0031343f0e9940de31fc34b973252d445638717c3c7ec909a87a94215c0dc6640a414b759433219b0dca972baf168d0b37a78510418337d04801"],"id":"2C7YHue3h8TKn9trj1bdKYeSaAPCV7CCotC7yaguBLSuHV1G5u"}`, string(jsondata))

	assert.NoError(bs.VerifyState())
}
