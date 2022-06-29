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

func TestTxUpdateAccountKeepers(t *testing.T) {
	assert := assert.New(t)

	// SyntacticVerify
	tx := &TxUpdateAccountKeepers{}
	assert.ErrorContains(tx.SyntacticVerify(), "nil pointer")
	_, err := tx.MarshalJSON()
	assert.NoError(err)

	bctx := NewMockBCtx()
	bs := bctx.MockBS()
	token := ld.MustNewToken("$LDC")

	sender := util.Signer1.Address()
	approver := util.Signer2.Address()

	txData := &ld.TxData{
		Type:      ld.TypeUpdateAccountKeepers,
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
		Type:      ld.TypeUpdateAccountKeepers,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      sender,
		To:        &constants.GenesisAccount,
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx(txData.ToTransaction(), true)
	assert.ErrorContains(err, "invalid to, should be nil")

	txData = &ld.TxData{
		Type:      ld.TypeUpdateAccountKeepers,
		ChainID:   bctx.Chain().ChainID,
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
		Type:      ld.TypeUpdateAccountKeepers,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      sender,
		Amount:    big.NewInt(1),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx(txData.ToTransaction(), true)
	assert.ErrorContains(err, "nil to together with amount")

	txData = &ld.TxData{
		Type:      ld.TypeUpdateAccountKeepers,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      sender,
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx(txData.ToTransaction(), true)
	assert.ErrorContains(err, "invalid data")

	txData = &ld.TxData{
		Type:      ld.TypeUpdateAccountKeepers,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      sender,
		Data:      []byte("你好👋"),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx(txData.ToTransaction(), true)
	assert.ErrorContains(err, "cbor: cannot unmarshal")

	input := ld.TxAccounter{}
	assert.NoError(input.SyntacticVerify())
	txData = &ld.TxData{
		Type:      ld.TypeUpdateAccountKeepers,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      sender,
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx(txData.ToTransaction(), true)
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
	txData = &ld.TxData{
		Type:      ld.TypeUpdateAccountKeepers,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      sender,
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	tt := txData.ToTransaction()
	itx, err := NewTx(tt, true)
	assert.NoError(err)

	bs.CommitAccounts()
	assert.ErrorContains(itx.Apply(bctx, bs),
		"insufficient NativeLDC balance, expected 2088900, got 0")
	bs.CheckoutAccounts()

	senderAcc := bs.MustAccount(sender)
	senderAcc.Add(constants.NativeToken, new(big.Int).SetUint64(constants.LDC))

	assert.Equal(uint16(0), senderAcc.Threshold())
	assert.Equal(util.EthIDs{}, senderAcc.Keepers())
	assert.Nil(senderAcc.ld.Approver)
	assert.Nil(senderAcc.ld.ApproveList)
	assert.NoError(itx.Apply(bctx, bs))

	senderGas := tt.Gas()
	assert.Equal(senderGas*bctx.Price,
		itx.(*TxUpdateAccountKeepers).ldc.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(senderGas*100,
		itx.(*TxUpdateAccountKeepers).miner.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(constants.LDC-senderGas*(bctx.Price+100),
		senderAcc.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(uint64(1), senderAcc.Nonce())

	assert.Equal(uint16(1), senderAcc.Threshold())
	assert.Equal(util.EthIDs{util.Signer1.Address()}, senderAcc.Keepers())
	assert.Equal(approver, *senderAcc.ld.Approver)
	assert.Equal(ld.AccountTxTypes, senderAcc.ld.ApproveList)

	jsondata, err := itx.MarshalJSON()
	assert.NoError(err)
	// fmt.Println(string(jsondata))
	assert.Equal(`{"type":"TypeUpdateAccountKeepers","chainID":2357,"nonce":0,"gasTip":100,"gasFeeCap":1000,"from":"0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC","data":{"threshold":1,"keepers":["0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC"],"approver":"0x44171C37Ff5D7B7bb8dcad5C81f16284A229e641","approveList":["TypeAddNonceTable","TypeUpdateAccountKeepers","TypeCreateToken","TypeDestroyToken","TypeCreateStake","TypeResetStake","TypeDestroyStake","TypeTakeStake","TypeWithdrawStake","TypeUpdateStakeApprover","TypeOpenLending","TypeCloseLending","TypeBorrow","TypeRepay"]},"signatures":["f17168d2ddcf516e263bd27ad2bd400b89b8482053ed4760aba782953dbf2e4b05d1ff96d7bc0a1c0726829373579216602e631ac25f8c2352b3cc6b9472315400"],"id":"KuMRgdoifs9ytwAv7EWwgmwrHrQpmoFEMu9YF9XaF8eMRZjMj"}`, string(jsondata))

	// update ApproveList
	input = ld.TxAccounter{
		ApproveList: ld.TransferTxTypes,
	}
	assert.NoError(input.SyntacticVerify())
	txData = &ld.TxData{
		Type:      ld.TypeUpdateAccountKeepers,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     1,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      sender,
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	tt = txData.ToTransaction()
	itx, err = NewTx(tt, true)
	assert.NoError(err)
	bs.CommitAccounts()
	assert.ErrorContains(itx.Apply(bctx, bs),
		"invalid signature for approver")
	bs.CheckoutAccounts()

	assert.NoError(txData.SignWith(util.Signer2))
	tt = txData.ToTransaction()
	itx, err = NewTx(tt, true)
	assert.NoError(err)
	assert.NoError(itx.Apply(bctx, bs))

	senderGas += tt.Gas()
	assert.Equal(senderGas*bctx.Price,
		itx.(*TxUpdateAccountKeepers).ldc.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(senderGas*100,
		itx.(*TxUpdateAccountKeepers).miner.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(constants.LDC-senderGas*(bctx.Price+100),
		senderAcc.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(uint16(1), senderAcc.Threshold())
	assert.Equal(util.EthIDs{util.Signer1.Address()}, senderAcc.Keepers())
	assert.Equal(approver, *senderAcc.ld.Approver)
	assert.Equal(ld.TransferTxTypes, senderAcc.ld.ApproveList)

	// clear Approver
	input = ld.TxAccounter{
		Approver: &util.EthIDEmpty,
	}
	assert.NoError(input.SyntacticVerify())
	txData = &ld.TxData{
		Type:      ld.TypeUpdateAccountKeepers,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     2,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      sender,
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	tt = txData.ToTransaction()
	itx, err = NewTx(tt, true)
	assert.NoError(err)
	assert.NoError(itx.Apply(bctx, bs))

	senderGas += tt.Gas()
	assert.Equal(senderGas*bctx.Price,
		itx.(*TxUpdateAccountKeepers).ldc.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(senderGas*100,
		itx.(*TxUpdateAccountKeepers).miner.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(constants.LDC-senderGas*(bctx.Price+100),
		senderAcc.balanceOf(constants.NativeToken).Uint64())
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
	txData = &ld.TxData{
		Type:      ld.TypeUpdateAccountKeepers,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     3,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      sender,
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	tt = txData.ToTransaction()
	itx, err = NewTx(tt, true)
	assert.NoError(err)
	assert.NoError(itx.Apply(bctx, bs))

	senderGas += tt.Gas()
	assert.Equal(senderGas*bctx.Price,
		itx.(*TxUpdateAccountKeepers).ldc.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(senderGas*100,
		itx.(*TxUpdateAccountKeepers).miner.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(constants.LDC-senderGas*(bctx.Price+100),
		senderAcc.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(uint16(1), senderAcc.Threshold())
	assert.Equal(util.EthIDs{util.Signer1.Address(), util.Signer2.Address()}, senderAcc.Keepers())

	// add Approver again
	input = ld.TxAccounter{
		Approver: &approver,
	}
	assert.NoError(input.SyntacticVerify())
	txData = &ld.TxData{
		Type:      ld.TypeUpdateAccountKeepers,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     4,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      sender,
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	tt = txData.ToTransaction()
	itx, err = NewTx(tt, true)
	assert.NoError(err)
	bs.CommitAccounts()
	assert.ErrorContains(itx.Apply(bctx, bs),
		"invalid signatures for keepers")
	bs.CheckoutAccounts()

	// check duplicate signatures
	assert.NoError(txData.SignWith(util.Signer1))
	tt = txData.ToTransaction()
	itx, err = NewTx(tt, true)
	assert.ErrorContains(err,
		"DeriveSigners error: duplicate address 0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC")

	txData = &ld.TxData{
		Type:      ld.TypeUpdateAccountKeepers,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     4,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      sender,
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	assert.NoError(txData.SignWith(util.Signer2))
	tt = txData.ToTransaction()
	itx, err = NewTx(tt, true)
	assert.NoError(err)
	assert.NoError(itx.Apply(bctx, bs))

	senderGas += tt.Gas()
	assert.Equal(senderGas*bctx.Price,
		itx.(*TxUpdateAccountKeepers).ldc.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(senderGas*100,
		itx.(*TxUpdateAccountKeepers).miner.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(constants.LDC-senderGas*(bctx.Price+100),
		senderAcc.balanceOf(constants.NativeToken).Uint64())
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
	txData = &ld.TxData{
		Type:      ld.TypeUpdateAccountKeepers,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     5,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      sender,
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	assert.NoError(txData.SignWith(util.Signer2))
	tt = txData.ToTransaction()
	itx, err = NewTx(tt, true)
	assert.ErrorContains(err, "invalid threshold, expected >= 1")

	assert.NoError(bs.VerifyState())
}

func TestTxUpdateAccountKeepersGenesis(t *testing.T) {
	assert := assert.New(t)

	bctx := NewMockBCtx()
	bs := bctx.MockBS()
	sender := util.Signer1.Address()

	input := ld.TxAccounter{
		Threshold: ld.Uint16Ptr(1),
		Keepers:   &util.EthIDs{util.Signer1.Address()},
	}
	assert.NoError(input.SyntacticVerify())
	txData := &ld.TxData{
		Type:    ld.TypeUpdateAccountKeepers,
		ChainID: bctx.Chain().ChainID,
		From:    sender,
		Data:    input.Bytes(),
	}
	itx, err := NewGenesisTx(txData.ToTransaction())
	assert.NoError(err)
	assert.NoError(itx.(*TxUpdateAccountKeepers).ApplyGenesis(bctx, bs))

	senderAcc := bs.MustAccount(sender)
	assert.Equal(uint64(0), itx.(*TxUpdateAccountKeepers).ldc.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(uint64(0), itx.(*TxUpdateAccountKeepers).miner.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(uint64(0), senderAcc.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(uint64(1), senderAcc.Nonce())

	assert.Equal(uint16(1), senderAcc.Threshold())
	assert.Equal(util.EthIDs{util.Signer1.Address()}, senderAcc.Keepers())

	jsondata, err := itx.MarshalJSON()
	assert.NoError(err)
	// fmt.Println(string(jsondata))
	assert.Equal(`{"type":"TypeUpdateAccountKeepers","chainID":2357,"nonce":0,"gasTip":0,"gasFeeCap":0,"from":"0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC","data":{"threshold":1,"keepers":["0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC"]},"id":"yK8rnb4pH5r5vJGPWicdsR4yiL6AhK9gJLnWhh2tLSQDSfNFz"}`, string(jsondata))

	assert.NoError(bs.VerifyState())
}
