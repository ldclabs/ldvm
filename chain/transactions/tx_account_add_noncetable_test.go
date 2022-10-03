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

func TestTxAddNonceTable(t *testing.T) {
	assert := assert.New(t)

	// SyntacticVerify
	tx := &TxAddNonceTable{}
	assert.ErrorContains(tx.SyntacticVerify(), "nil pointer")
	_, err := tx.MarshalJSON()
	assert.NoError(err)

	ctx := NewMockChainContext()
	cs := ctx.MockChainState()
	sender := util.Signer1.Address()

	ltx := &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeAddNonceTable,
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
		Type:      ld.TypeAddNonceTable,
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
		Type:      ld.TypeAddNonceTable,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      sender,
		Token:     &constants.NativeToken,
	}}
	assert.NoError(ltx.SignWith(util.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "invalid token, should be nil")

	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeAddNonceTable,
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
		Type:      ld.TypeAddNonceTable,
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
		Type:      ld.TypeAddNonceTable,
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

	input := []uint64{10}
	inputData, err := util.MarshalCBOR(input)
	assert.NoError(err)
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeAddNonceTable,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      sender,
		Data:      inputData,
	}}
	assert.NoError(ltx.SignWith(util.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "no nonce")

	input = make([]uint64, 1026)
	for i := range input {
		input[i] = uint64(i)
	}
	inputData, err = util.MarshalCBOR(input)
	assert.NoError(err)
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeAddNonceTable,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      sender,
		Data:      inputData,
	}}
	assert.NoError(ltx.SignWith(util.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "too many nonces, expected <= 1024, got 1025")

	input = []uint64{cs.Timestamp() - 1, 123}
	inputData, err = util.MarshalCBOR(input)
	assert.NoError(err)
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeAddNonceTable,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      sender,
		Data:      inputData,
	}}
	assert.NoError(ltx.SignWith(util.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	ltx.Timestamp = cs.Timestamp()
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "invalid expire time, expected > 1000, got 999")

	input = []uint64{3600*24*30 + 2, 123}
	inputData, err = util.MarshalCBOR(input)
	assert.NoError(err)
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeAddNonceTable,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      sender,
		Data:      inputData,
	}}
	assert.NoError(ltx.SignWith(util.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	ltx.Timestamp = 1
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "invalid expire time, expected <= 2592001, got 2592002")

	input = []uint64{cs.Timestamp() + 1, 1, 3, 7, 5}
	inputData, err = util.MarshalCBOR(input)
	assert.NoError(err)
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeAddNonceTable,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      sender,
		Data:      inputData,
	}}
	assert.NoError(ltx.SignWith(util.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	ltx.Timestamp = cs.Timestamp()
	itx, err := NewTx(ltx)
	assert.NoError(err)

	cs.CommitAccounts()
	assert.ErrorContains(itx.Apply(ctx, cs), "insufficient NativeLDC balance, expected 603900, got 0")
	cs.CheckoutAccounts()

	senderAcc := cs.MustAccount(sender)
	senderAcc.Add(constants.NativeToken, new(big.Int).SetUint64(constants.LDC))
	assert.NoError(itx.Apply(ctx, cs))

	senderGas := ltx.Gas()
	assert.Equal(senderGas*ctx.Price,
		itx.(*TxAddNonceTable).ldc.Balance().Uint64())
	assert.Equal(senderGas*100,
		itx.(*TxAddNonceTable).miner.Balance().Uint64())
	assert.Equal(constants.LDC-senderGas*(ctx.Price+100),
		senderAcc.Balance().Uint64())
	assert.Equal(uint64(1), senderAcc.Nonce())
	assert.Equal(1, len(senderAcc.ld.NonceTable))
	assert.Equal([]uint64{1, 3, 5, 7}, senderAcc.ld.NonceTable[cs.Timestamp()+1])

	jsondata, err := itx.MarshalJSON()
	assert.NoError(err)
	// fmt.Println(string(jsondata))
	assert.Equal(`{"tx":{"type":"TypeAddNonceTable","chainID":2357,"nonce":0,"gasTip":100,"gasFeeCap":1000,"from":"0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC","data":[1001,1,3,7,5]},"sigs":["ef07cf7075394c343ee99f34d2c76efaa3789ecc4b9c48f896aecd01e343f30c0d3e8c67958bf10a33979cdcf1fbcf9c3b6df7c6f7583ec795a3dace2f75b4c200"],"id":"BK1BDD5iwMuYKNYwyTgqwZuPDai3go2mmsFmUnvn5RCixEHHs"}`, string(jsondata))

	input = []uint64{cs.Timestamp() + 1, 2, 4, 1}
	inputData, err = util.MarshalCBOR(input)
	assert.NoError(err)
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeAddNonceTable,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     1,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      sender,
		Data:      inputData,
	}}
	assert.NoError(ltx.SignWith(util.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	ltx.Timestamp = cs.Timestamp()
	itx, err = NewTx(ltx)
	assert.NoError(err)

	cs.CommitAccounts()
	assert.ErrorContains(itx.Apply(ctx, cs), "nonce 1 exists at 1001")
	cs.CheckoutAccounts()

	input = []uint64{cs.Timestamp() + 1, 2, 4, 6}
	inputData, err = util.MarshalCBOR(input)
	assert.NoError(err)
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeAddNonceTable,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     1,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      sender,
		Data:      inputData,
	}}
	assert.NoError(ltx.SignWith(util.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	ltx.Timestamp = cs.Timestamp()
	itx, err = NewTx(ltx)
	assert.NoError(err)
	assert.NoError(itx.Apply(ctx, cs))

	senderGas += ltx.Gas()
	assert.Equal(senderGas*ctx.Price,
		itx.(*TxAddNonceTable).ldc.Balance().Uint64())
	assert.Equal(senderGas*100,
		itx.(*TxAddNonceTable).miner.Balance().Uint64())
	assert.Equal(uint64(2), senderAcc.Nonce())
	assert.Equal(1, len(senderAcc.ld.NonceTable))
	assert.Equal([]uint64{1, 2, 3, 4, 5, 6, 7}, senderAcc.ld.NonceTable[cs.Timestamp()+1])

	input = []uint64{cs.Timestamp() + 2, 0}
	inputData, err = util.MarshalCBOR(input)
	assert.NoError(err)
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeAddNonceTable,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     2,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      sender,
		Data:      inputData,
	}}
	assert.NoError(ltx.SignWith(util.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	ltx.Timestamp = cs.Timestamp()
	itx, err = NewTx(ltx)
	assert.NoError(err)
	assert.NoError(itx.Apply(ctx, cs))

	senderGas += ltx.Gas()
	assert.Equal(senderGas*ctx.Price,
		itx.(*TxAddNonceTable).ldc.Balance().Uint64())
	assert.Equal(senderGas*100,
		itx.(*TxAddNonceTable).miner.Balance().Uint64())
	assert.Equal(uint64(3), senderAcc.Nonce())
	assert.Equal(2, len(senderAcc.ld.NonceTable))
	assert.Equal([]uint64{1, 2, 3, 4, 5, 6, 7}, senderAcc.ld.NonceTable[cs.Timestamp()+1])
	assert.Equal([]uint64{0}, senderAcc.ld.NonceTable[cs.Timestamp()+2])

	// consume nonce table
	recipientAcc := cs.MustAccount(util.Signer2.Address())
	input2 := ld.TxTransfer{
		Nonce:  0,
		From:   &sender,
		To:     &recipientAcc.id,
		Amount: new(big.Int).SetUint64(constants.MilliLDC),
		Expire: cs.Timestamp() + 1,
	}
	assert.NoError(input2.SyntacticVerify())
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeTransferCash,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      recipientAcc.id,
		To:        &sender,
		Data:      input2.Bytes(),
	}}

	assert.NoError(ltx.SignWith(util.Signer2))
	assert.NoError(ltx.ExSignWith(util.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	ltx.Timestamp = cs.Timestamp()
	recipientAcc.Add(constants.NativeToken, new(big.Int).SetUint64(constants.LDC))
	itx, err = NewTx(ltx)
	assert.NoError(err)

	cs.CommitAccounts()
	assert.ErrorContains(itx.Apply(ctx, cs), "nonce 0 not exists at 1001")
	cs.CheckoutAccounts()

	input2 = ld.TxTransfer{
		Nonce:  0,
		From:   &sender,
		To:     &recipientAcc.id,
		Amount: new(big.Int).SetUint64(constants.MilliLDC),
		Expire: cs.Timestamp() + 2,
	}
	assert.NoError(input2.SyntacticVerify())
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeTransferCash,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      recipientAcc.id,
		To:        &sender,
		Data:      input2.Bytes(),
	}}

	assert.NoError(ltx.SignWith(util.Signer2))
	assert.NoError(ltx.ExSignWith(util.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	ltx.Timestamp = cs.Timestamp()
	itx, err = NewTx(ltx)
	assert.NoError(err)
	assert.NoError(itx.Apply(ctx, cs))

	senderGas += ltx.Gas()
	assert.Equal(senderGas*ctx.Price,
		itx.(*TxTransferCash).ldc.Balance().Uint64())
	assert.Equal(senderGas*100,
		itx.(*TxTransferCash).miner.Balance().Uint64())
	assert.Equal(1, len(senderAcc.ld.NonceTable))
	assert.Equal([]uint64{1, 2, 3, 4, 5, 6, 7}, senderAcc.ld.NonceTable[cs.Timestamp()+1])
	assert.Nil(senderAcc.ld.NonceTable[cs.Timestamp()+2], "should clean emtpy nonce table")

	assert.NoError(cs.VerifyState())
}
