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

func TestTxUpdateAccountInfo(t *testing.T) {
	assert := assert.New(t)

	// SyntacticVerify
	tx := &TxUpdateAccountInfo{}
	assert.ErrorContains(tx.SyntacticVerify(), "nil pointer")
	_, err := tx.MarshalJSON()
	assert.NoError(err)

	ctx := NewMockChainContext()
	cs := ctx.MockChainState()
	token := ld.MustNewToken("$LDC")

	sender := util.Signer1.Address()
	approver := util.Signer2.Address()

	ltx := &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeUpdateAccountInfo,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      sender,
	}}
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "DeriveSigners error: no signature")

	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeUpdateAccountInfo,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      sender,
		To:        &constants.GenesisAccount,
	}}
	assert.NoError(ltx.SignWith(util.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "invalid to, should be nil")

	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeUpdateAccountInfo,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      sender,
		Token:     &token,
	}}
	assert.NoError(ltx.SignWith(util.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "invalid token, should be nil")

	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeUpdateAccountInfo,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      sender,
		Amount:    big.NewInt(1),
	}}
	assert.NoError(ltx.SignWith(util.Signer1))
	assert.ErrorContains(ltx.SyntacticVerify(), "nil \"to\" together with amount")

	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeUpdateAccountInfo,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      sender,
	}}
	assert.NoError(ltx.SignWith(util.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "invalid data")

	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeUpdateAccountInfo,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      sender,
		Data:      []byte("你好👋"),
	}}
	assert.NoError(ltx.SignWith(util.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "cbor: unexpected following extraneous data")

	input := ld.TxAccounter{}
	assert.NoError(input.SyntacticVerify())
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeUpdateAccountInfo,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      sender,
		Data:      input.Bytes(),
	}}
	assert.NoError(ltx.SignWith(util.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "no keepers nor approver")

	input = ld.TxAccounter{Threshold: ld.Uint16Ptr(0)}
	assert.ErrorContains(input.SyntacticVerify(), "nil keepers together with threshold")
	input = ld.TxAccounter{Keepers: &util.EthIDs{}}
	assert.ErrorContains(input.SyntacticVerify(), "nil threshold together with keepers")
	input = ld.TxAccounter{
		Threshold:   ld.Uint16Ptr(1),
		Keepers:     &util.EthIDs{util.Signer1.Address()},
		Approver:    &approver,
		ApproveList: ld.AccountTxTypes,
	}
	assert.NoError(input.SyntacticVerify())
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeUpdateAccountInfo,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      sender,
		Data:      input.Bytes(),
	}}
	assert.NoError(ltx.SignWith(util.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	itx, err := NewTx(ltx)
	assert.NoError(err)

	cs.CommitAccounts()
	assert.ErrorContains(itx.Apply(ctx, cs),
		"insufficient NativeLDC balance, expected 1059300, got 0")
	cs.CheckoutAccounts()

	senderAcc := cs.MustAccount(sender)
	senderAcc.Add(constants.NativeToken, new(big.Int).SetUint64(constants.LDC))

	assert.Equal(uint16(0), senderAcc.Threshold())
	assert.Equal(util.EthIDs{}, senderAcc.Keepers())
	assert.Nil(senderAcc.ld.Approver)
	assert.Nil(senderAcc.ld.ApproveList)
	assert.NoError(itx.Apply(ctx, cs))

	senderGas := ltx.Gas()
	assert.Equal(senderGas*ctx.Price,
		itx.(*TxUpdateAccountInfo).ldc.Balance().Uint64())
	assert.Equal(senderGas*100,
		itx.(*TxUpdateAccountInfo).miner.Balance().Uint64())
	assert.Equal(constants.LDC-senderGas*(ctx.Price+100),
		senderAcc.Balance().Uint64())
	assert.Equal(uint64(1), senderAcc.Nonce())

	assert.Equal(uint16(1), senderAcc.Threshold())
	assert.Equal(util.EthIDs{util.Signer1.Address()}, senderAcc.Keepers())
	assert.Equal(approver, *senderAcc.ld.Approver)
	assert.Equal(ld.AccountTxTypes, senderAcc.ld.ApproveList)

	jsondata, err := itx.MarshalJSON()
	assert.NoError(err)
	// fmt.Println(string(jsondata))
	assert.Equal(`{"tx":{"type":"TypeUpdateAccountInfo","chainID":2357,"nonce":0,"gasTip":100,"gasFeeCap":1000,"from":"0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC","data":{"threshold":1,"keepers":["0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC"],"approver":"0x44171C37Ff5D7B7bb8dcad5C81f16284A229e641","approveList":["TypeAddNonceTable","TypeUpdateAccountInfo","TypeCreateToken","TypeDestroyToken","TypeCreateStake","TypeResetStake","TypeDestroyStake","TypeTakeStake","TypeWithdrawStake","TypeUpdateStakeApprover","TypeOpenLending","TypeCloseLending","TypeBorrow","TypeRepay"]}},"sigs":["f17168d2ddcf516e263bd27ad2bd400b89b8482053ed4760aba782953dbf2e4b05d1ff96d7bc0a1c0726829373579216602e631ac25f8c2352b3cc6b9472315400"],"id":"2wgNUWFzkRANetUdGMaVM3wmtbHrtrBLQJSpX9i6jjbBWkRFym"}`, string(jsondata))

	// update ApproveList
	input = ld.TxAccounter{
		ApproveList: ld.TransferTxTypes,
	}
	assert.NoError(input.SyntacticVerify())
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeUpdateAccountInfo,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     1,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      sender,
		Data:      input.Bytes(),
	}}
	assert.NoError(ltx.SignWith(util.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	itx, err = NewTx(ltx)
	assert.NoError(err)
	cs.CommitAccounts()
	assert.ErrorContains(itx.Apply(ctx, cs),
		"invalid signature for approver")
	cs.CheckoutAccounts()

	assert.NoError(ltx.SignWith(util.Signer1, util.Signer2))
	assert.NoError(ltx.SyntacticVerify())
	itx, err = NewTx(ltx)
	assert.NoError(err)
	assert.NoError(itx.Apply(ctx, cs))

	senderGas += ltx.Gas()
	assert.Equal(senderGas*ctx.Price,
		itx.(*TxUpdateAccountInfo).ldc.Balance().Uint64())
	assert.Equal(senderGas*100,
		itx.(*TxUpdateAccountInfo).miner.Balance().Uint64())
	assert.Equal(constants.LDC-senderGas*(ctx.Price+100),
		senderAcc.Balance().Uint64())
	assert.Equal(uint16(1), senderAcc.Threshold())
	assert.Equal(util.EthIDs{util.Signer1.Address()}, senderAcc.Keepers())
	assert.Equal(approver, *senderAcc.ld.Approver)
	assert.Equal(ld.TransferTxTypes, senderAcc.ld.ApproveList)

	// clear Approver
	input = ld.TxAccounter{
		Approver: &util.EthIDEmpty,
	}
	assert.NoError(input.SyntacticVerify())
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeUpdateAccountInfo,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     2,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      sender,
		Data:      input.Bytes(),
	}}
	assert.NoError(ltx.SignWith(util.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	itx, err = NewTx(ltx)
	assert.NoError(err)
	assert.NoError(itx.Apply(ctx, cs))

	senderGas += ltx.Gas()
	assert.Equal(senderGas*ctx.Price,
		itx.(*TxUpdateAccountInfo).ldc.Balance().Uint64())
	assert.Equal(senderGas*100,
		itx.(*TxUpdateAccountInfo).miner.Balance().Uint64())
	assert.Equal(constants.LDC-senderGas*(ctx.Price+100),
		senderAcc.Balance().Uint64())
	assert.Equal(uint16(1), senderAcc.Threshold())
	assert.Equal(util.EthIDs{util.Signer1.Address()}, senderAcc.Keepers())
	assert.Nil(senderAcc.ld.Approver)
	assert.Nil(senderAcc.ld.ApproveList)

	// update Keepers
	input = ld.TxAccounter{
		Threshold: ld.Uint16Ptr(1),
		Keepers:   &util.EthIDs{util.Signer1.Address(), util.Signer2.Address()},
	}
	assert.NoError(input.SyntacticVerify())
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeUpdateAccountInfo,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     3,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      sender,
		Data:      input.Bytes(),
	}}
	assert.NoError(ltx.SignWith(util.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	itx, err = NewTx(ltx)
	assert.NoError(err)
	assert.NoError(itx.Apply(ctx, cs))

	senderGas += ltx.Gas()
	assert.Equal(senderGas*ctx.Price,
		itx.(*TxUpdateAccountInfo).ldc.Balance().Uint64())
	assert.Equal(senderGas*100,
		itx.(*TxUpdateAccountInfo).miner.Balance().Uint64())
	assert.Equal(constants.LDC-senderGas*(ctx.Price+100),
		senderAcc.Balance().Uint64())
	assert.Equal(uint16(1), senderAcc.Threshold())
	assert.Equal(util.EthIDs{util.Signer1.Address(), util.Signer2.Address()}, senderAcc.Keepers())

	// add Approver again
	input = ld.TxAccounter{
		Approver: &approver,
	}
	assert.NoError(input.SyntacticVerify())
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeUpdateAccountInfo,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     4,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      sender,
		Data:      input.Bytes(),
	}}
	assert.NoError(ltx.SignWith(util.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	itx, err = NewTx(ltx)
	assert.NoError(err)
	cs.CommitAccounts()
	assert.ErrorContains(itx.Apply(ctx, cs),
		"invalid signatures for keepers")
	cs.CheckoutAccounts()

	// check duplicate signatures
	assert.NoError(ltx.SignWith(util.Signer1, util.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err,
		"DeriveSigners error: duplicate address 0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC")

	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeUpdateAccountInfo,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     4,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      sender,
		Data:      input.Bytes(),
	}}
	assert.NoError(ltx.SignWith(util.Signer1, util.Signer2))
	assert.NoError(ltx.SyntacticVerify())
	itx, err = NewTx(ltx)
	assert.NoError(err)
	assert.NoError(itx.Apply(ctx, cs))

	senderGas += ltx.Gas()
	assert.Equal(senderGas*ctx.Price,
		itx.(*TxUpdateAccountInfo).ldc.Balance().Uint64())
	assert.Equal(senderGas*100,
		itx.(*TxUpdateAccountInfo).miner.Balance().Uint64())
	assert.Equal(constants.LDC-senderGas*(ctx.Price+100),
		senderAcc.Balance().Uint64())
	assert.Equal(uint16(1), senderAcc.Threshold())
	assert.Equal(util.EthIDs{util.Signer1.Address(), util.Signer2.Address()}, senderAcc.Keepers())
	assert.Equal(approver, *senderAcc.ld.Approver)
	assert.Nil(senderAcc.ld.ApproveList)

	// clear keepers should failed
	input = ld.TxAccounter{
		Threshold: ld.Uint16Ptr(0),
		Keepers:   &util.EthIDs{},
	}
	assert.NoError(input.SyntacticVerify())
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeUpdateAccountInfo,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     5,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      sender,
		Data:      input.Bytes(),
	}}
	assert.NoError(ltx.SignWith(util.Signer1, util.Signer2))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "invalid threshold, expected >= 1")

	assert.NoError(cs.VerifyState())
}

func TestTxUpdateAccountInfoGenesis(t *testing.T) {
	assert := assert.New(t)

	ctx := NewMockChainContext()
	cs := ctx.MockChainState()
	sender := util.Signer1.Address()

	input := ld.TxAccounter{
		Threshold: ld.Uint16Ptr(1),
		Keepers:   &util.EthIDs{util.Signer1.Address()},
	}
	assert.NoError(input.SyntacticVerify())
	ltx := &ld.Transaction{Tx: ld.TxData{
		Type:    ld.TypeUpdateAccountInfo,
		ChainID: ctx.ChainConfig().ChainID,
		From:    sender,
		Data:    input.Bytes(),
	}}
	assert.NoError(ltx.SyntacticVerify())
	itx, err := NewGenesisTx(ltx)
	assert.NoError(err)
	assert.NoError(itx.(*TxUpdateAccountInfo).ApplyGenesis(ctx, cs))

	senderAcc := cs.MustAccount(sender)
	assert.Equal(uint64(0), itx.(*TxUpdateAccountInfo).ldc.Balance().Uint64())
	assert.Equal(uint64(0), itx.(*TxUpdateAccountInfo).miner.Balance().Uint64())
	assert.Equal(uint64(0), senderAcc.Balance().Uint64())
	assert.Equal(uint64(1), senderAcc.Nonce())

	assert.Equal(uint16(1), senderAcc.Threshold())
	assert.Equal(util.EthIDs{util.Signer1.Address()}, senderAcc.Keepers())

	jsondata, err := itx.MarshalJSON()
	assert.NoError(err)
	// fmt.Println(string(jsondata))
	assert.Equal(`{"tx":{"type":"TypeUpdateAccountInfo","chainID":2357,"nonce":0,"gasTip":0,"gasFeeCap":0,"from":"0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC","data":{"threshold":1,"keepers":["0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC"]}},"id":"2RvKvLzo52bcJ11AYJZLQiA2YzWTXck5LzZmHa8uUYC4UEGnB4"}`, string(jsondata))

	assert.NoError(cs.VerifyState())
}
