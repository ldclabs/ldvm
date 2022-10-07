// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package transactions

import (
	"math/big"
	"testing"

	"github.com/ldclabs/ldvm/constants"
	"github.com/ldclabs/ldvm/ld"
	"github.com/ldclabs/ldvm/util/signer"
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

	sender := signer.Signer1.Key().Address()
	approver := signer.Signer2.Key()

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
	assert.ErrorContains(err, "no signatures")

	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeUpdateAccountInfo,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      sender,
		To:        &constants.GenesisAccount,
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
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
	assert.NoError(ltx.SignWith(signer.Signer1))
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
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.ErrorContains(ltx.SyntacticVerify(), "nil \"to\" together with amount")

	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeUpdateAccountInfo,
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
		Type:      ld.TypeUpdateAccountInfo,
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
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "no keepers nor approver")

	input = ld.TxAccounter{Threshold: ld.Uint16Ptr(0)}
	assert.ErrorContains(input.SyntacticVerify(), "nil keepers together with threshold")
	input = ld.TxAccounter{Keepers: &signer.Keys{}}
	assert.ErrorContains(input.SyntacticVerify(), "nil threshold together with keepers")
	input = ld.TxAccounter{
		Threshold:   ld.Uint16Ptr(1),
		Keepers:     &signer.Keys{signer.Signer1.Key()},
		Approver:    &approver,
		ApproveList: &ld.AccountTxTypes,
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
	assert.NoError(ltx.SignWith(signer.Signer1))
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
	assert.Equal(signer.Keys{}, senderAcc.Keepers())
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
	assert.Equal(signer.Keys{signer.Signer1.Key()}, senderAcc.Keepers())
	assert.Equal(approver, senderAcc.ld.Approver)
	assert.Equal(ld.AccountTxTypes, senderAcc.ld.ApproveList)

	jsondata, err := itx.MarshalJSON()
	assert.NoError(err)
	// fmt.Println(string(jsondata))
	assert.Equal(`{"tx":{"type":"TypeUpdateAccountInfo","chainID":2357,"nonce":0,"gasTip":100,"gasFeeCap":1000,"from":"0x8db97c7cECe249C2b98bdc0226cc4C2A57bF52fc","data":{"threshold":1,"keepers":["jbl8fOziScK5i9wCJsxMKle_UvwKxwPH"],"approver":"RBccN_9de3u43K1cgfFihKIp5kE1lmGG","approveList":["TypeAddNonceTable","TypeUpdateAccountInfo","TypeCreateToken","TypeDestroyToken","TypeCreateStake","TypeResetStake","TypeDestroyStake","TypeTakeStake","TypeWithdrawStake","TypeUpdateStakeApprover","TypeOpenLending","TypeCloseLending","TypeBorrow","TypeRepay"]}},"sigs":["8XFo0t3PUW4mO9J60r1AC4m4SCBT7Udgq6eClT2_LksF0f-W17wKHAcmgpNzV5IWYC5jGsJfjCNSs8xrlHIxVABb1Egx"],"id":"_9nauLRM0XLKzAumQThx_afPZ9yBbttrgLAc28u9uJDeOEkU"}`, string(jsondata))

	// update ApproveList
	input = ld.TxAccounter{
		ApproveList: &ld.TransferTxTypes,
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
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	itx, err = NewTx(ltx)
	assert.NoError(err)
	cs.CommitAccounts()
	assert.ErrorContains(itx.Apply(ctx, cs),
		"invalid signature for approver")
	cs.CheckoutAccounts()

	assert.NoError(ltx.SignWith(signer.Signer1, signer.Signer2))
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
	assert.Equal(signer.Keys{signer.Signer1.Key()}, senderAcc.Keepers())
	assert.Equal(approver, senderAcc.ld.Approver)
	assert.Equal(ld.TransferTxTypes, senderAcc.ld.ApproveList)

	// clear Approver
	input = ld.TxAccounter{
		Approver: &signer.Key{},
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
	assert.NoError(ltx.SignWith(signer.Signer1))
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
	assert.Equal(signer.Keys{signer.Signer1.Key()}, senderAcc.Keepers())
	assert.Nil(senderAcc.ld.Approver)
	assert.Nil(senderAcc.ld.ApproveList)

	// update Keepers
	input = ld.TxAccounter{
		Threshold: ld.Uint16Ptr(1),
		Keepers:   &signer.Keys{signer.Signer1.Key(), signer.Signer2.Key()},
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
	assert.NoError(ltx.SignWith(signer.Signer1))
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
	assert.Equal(signer.Keys{signer.Signer1.Key(), signer.Signer2.Key()}, senderAcc.Keepers())

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
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	itx, err = NewTx(ltx)
	assert.NoError(err)
	cs.CommitAccounts()
	assert.ErrorContains(itx.Apply(ctx, cs),
		"invalid signatures for keepers")
	cs.CheckoutAccounts()

	// check duplicate signatures
	assert.NoError(ltx.SignWith(signer.Signer1, signer.Signer1))
	assert.ErrorContains(ltx.SyntacticVerify(), "duplicate sig AGUY9GbJGrC737U5m994ZNqFdGghyPrSHP0ne6ehnJBY55CC6RnqjqkJDBZ0aIOt_b5coQDvMAb7YWidTHqMCwHzBtbS")

	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeUpdateAccountInfo,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     4,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      sender,
		Data:      input.Bytes(),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1, signer.Signer2))
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
	assert.Equal(signer.Keys{signer.Signer1.Key(), signer.Signer2.Key()}, senderAcc.Keepers())
	assert.Equal(approver, senderAcc.ld.Approver)
	assert.Nil(senderAcc.ld.ApproveList)

	// clear keepers should failed
	input = ld.TxAccounter{
		Threshold: ld.Uint16Ptr(0),
		Keepers:   &signer.Keys{},
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
	assert.NoError(ltx.SignWith(signer.Signer1, signer.Signer2))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "invalid threshold, expected >= 1")

	assert.NoError(cs.VerifyState())
}

func TestTxUpdateAccountInfoGenesis(t *testing.T) {
	assert := assert.New(t)

	ctx := NewMockChainContext()
	cs := ctx.MockChainState()
	sender := signer.Signer1.Key().Address()

	input := ld.TxAccounter{
		Threshold: ld.Uint16Ptr(1),
		Keepers:   &signer.Keys{signer.Signer1.Key()},
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
	assert.Equal(signer.Keys{signer.Signer1.Key()}, senderAcc.Keepers())

	jsondata, err := itx.MarshalJSON()
	assert.NoError(err)
	// fmt.Println(string(jsondata))
	assert.Equal(`{"tx":{"type":"TypeUpdateAccountInfo","chainID":2357,"nonce":0,"gasTip":0,"gasFeeCap":0,"from":"0x8db97c7cECe249C2b98bdc0226cc4C2A57bF52fc","data":{"threshold":1,"keepers":["jbl8fOziScK5i9wCJsxMKle_UvwKxwPH"]}},"id":"vEdXnrHRGJLl_9UI0g0BVaSSdNbGwdH25RY1LYUjcwuXOv6g"}`, string(jsondata))

	assert.NoError(cs.VerifyState())
}
