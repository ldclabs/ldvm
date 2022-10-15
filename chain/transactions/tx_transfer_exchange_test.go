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

func TestTxExchange(t *testing.T) {
	assert := assert.New(t)

	// SyntacticVerify
	tx := &TxExchange{}
	assert.ErrorContains(tx.SyntacticVerify(), "nil pointer")
	_, err := tx.MarshalJSON()
	assert.NoError(err)

	token := ld.MustNewToken("$LDC")
	ctx := NewMockChainContext()
	cs := ctx.MockChainState()

	from := cs.MustAccount(signer.Signer1.Key().Address())
	from.ld.Nonce = 1
	to := cs.MustAccount(signer.Signer2.Key().Address())

	ltx := &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeExchange,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     1,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      from.id,
	}}

	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "no signatures")

	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "invalid to")

	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeExchange,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     1,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      from.id,
		To:        &to.id,
	}}

	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "invalid amount")

	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeExchange,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     1,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      from.id,
		To:        &to.id,
		Amount:    new(big.Int).SetUint64(constants.LDC),
	}}

	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "invalid data")

	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeExchange,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     1,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      from.id,
		To:        &to.id,
		Amount:    new(big.Int).SetUint64(constants.LDC),
		Data:      []byte("abc"),
	}}

	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "no exSignatures")

	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeExchange,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     1,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      from.id,
		To:        &to.id,
		Amount:    new(big.Int).SetUint64(constants.LDC),
		Data:      []byte("abc"),
	}}

	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.ExSignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "cbor: unexpected following extraneous data")

	input := ld.TxExchanger{
		Nonce:   1,
		Sell:    token,
		Receive: constants.NativeToken,
		Quota:   new(big.Int).SetUint64(constants.LDC * 1000),
		Minimum: new(big.Int).SetUint64(constants.LDC),
		Price:   new(big.Int).SetUint64(1_000_000),
		Expire:  1,
		Payee:   to.id,
	}

	assert.NoError(input.SyntacticVerify())
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeExchange,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     1,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      from.id,
		To:        &to.id,
		Amount:    new(big.Int).SetUint64(1_000_000 - 1),
		Data:      input.Bytes(),
	}}

	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.ExSignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "invalid amount, expected >=1000000, got 999999")

	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeExchange,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     1,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      from.id,
		To:        &to.id,
		Amount:    new(big.Int).SetUint64(1_000_000*1000 + 1),
		Data:      input.Bytes(),
	}}

	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.ExSignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "invalid amount, expected <=1000000000, got 1000000001")

	input = ld.TxExchanger{
		Nonce:     1,
		Sell:      token,
		Receive:   constants.NativeToken,
		Quota:     new(big.Int).SetUint64(constants.LDC * 1000),
		Minimum:   new(big.Int).SetUint64(constants.LDC),
		Price:     new(big.Int).SetUint64(1_000_000),
		Expire:    1,
		Payee:     to.id,
		Purchaser: &constants.GenesisAccount,
	}
	assert.NoError(input.SyntacticVerify())
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeExchange,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     1,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      from.id,
		To:        &to.id,
		Amount:    new(big.Int).SetUint64(1_000_000),
		Data:      input.Bytes(),
	}}

	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.ExSignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err,
		"invalid from, expected 0xFFfFFFfFfffFFfFFffFFFfFfFffFFFfffFfFFFff, got 0x8db97c7cECe249C2b98bdc0226cc4C2A57bF52fc")

	input = ld.TxExchanger{
		Nonce:   1,
		Sell:    token,
		Receive: constants.NativeToken,
		Quota:   new(big.Int).SetUint64(constants.LDC * 1000),
		Minimum: new(big.Int).SetUint64(constants.LDC),
		Price:   new(big.Int).SetUint64(1_000_000),
		Expire:  1,
		Payee:   to.id,
	}
	assert.NoError(input.SyntacticVerify())
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeExchange,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     1,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      from.id,
		To:        &constants.GenesisAccount,
		Amount:    new(big.Int).SetUint64(1_000_000),
		Data:      input.Bytes(),
	}}

	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.ExSignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err,
		"invalid to, expected 0x44171C37Ff5D7B7bb8Dcad5C81f16284A229E641, got 0xFFfFFFfFfffFFfFFffFFFfFfFffFFFfffFfFFFff")

	input = ld.TxExchanger{
		Nonce:   1,
		Sell:    token,
		Receive: constants.NativeToken,
		Quota:   new(big.Int).SetUint64(constants.LDC * 1000),
		Minimum: new(big.Int).SetUint64(constants.LDC),
		Price:   new(big.Int).SetUint64(1_000_000),
		Expire:  1,
		Payee:   to.id,
	}
	assert.NoError(input.SyntacticVerify())
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeExchange,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     1,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      from.id,
		To:        &to.id,
		Token:     &token,
		Amount:    new(big.Int).SetUint64(1_000_000),
		Data:      input.Bytes(),
	}}

	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.ExSignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err,
		"invalid token, expected NativeLDC, got $LDC")

	input = ld.TxExchanger{
		Nonce:   1,
		Sell:    token,
		Receive: constants.NativeToken,
		Quota:   new(big.Int).SetUint64(constants.LDC * 1000),
		Minimum: new(big.Int).SetUint64(constants.LDC),
		Price:   new(big.Int).SetUint64(1_000_000),
		Expire:  cs.Timestamp(),
		Payee:   to.id,
	}
	assert.NoError(input.SyntacticVerify())
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeExchange,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     1,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      from.id,
		To:        &to.id,
		Amount:    new(big.Int).SetUint64(1_000_000),
		Data:      input.Bytes(),
	}}

	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.ExSignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	ltx.Timestamp = cs.Timestamp() + 1
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "data expired")
	ltx.Timestamp = 1
	_, err = NewTx(ltx)
	assert.NoError(err)

	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeExchange,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     1,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      from.id,
		To:        &to.id,
		Amount:    new(big.Int).SetUint64(1_000_000),
		Data:      input.Bytes(),
	}}

	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.ExSignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	itx, err := NewTx(ltx)
	assert.NoError(err)

	cs.CommitAccounts()
	assert.ErrorContains(itx.Apply(ctx, cs),
		"insufficient NativeLDC balance, expected 2823800, got 0")
	cs.CheckoutAccounts()

	from.Add(constants.NativeToken, new(big.Int).SetUint64(constants.LDC))
	cs.CommitAccounts()
	assert.ErrorContains(itx.Apply(ctx, cs), "invalid signatures for seller")
	cs.CheckoutAccounts()

	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeExchange,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     1,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      from.id,
		To:        &to.id,
		Amount:    new(big.Int).SetUint64(1_000_000),
		Data:      input.Bytes(),
	}}

	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.ExSignWith(signer.Signer2))
	assert.NoError(ltx.SyntacticVerify())
	ltx.Timestamp = cs.Timestamp()
	itx, err = NewTx(ltx)
	assert.NoError(err)
	cs.CommitAccounts()
	assert.ErrorContains(itx.Apply(ctx, cs),
		"nonce 1 not exists at 1")
	cs.CheckoutAccounts()
	assert.NoError(to.AddNonceTable(cs.Timestamp(), []uint64{1, 2, 3}))
	cs.CommitAccounts()
	assert.ErrorContains(itx.Apply(ctx, cs),
		"insufficient $LDC balance, expected 1000000000, got 0")
	cs.CheckoutAccounts()
	to.Add(token, new(big.Int).SetUint64(constants.LDC))
	assert.NoError(itx.Apply(ctx, cs))

	assert.Equal(ltx.Gas()*ctx.Price,
		itx.(*TxExchange).ldc.Balance().Uint64())
	assert.Equal(ltx.Gas()*100,
		itx.(*TxExchange).miner.Balance().Uint64())
	assert.Equal(uint64(1_000_000), to.Balance().Uint64())
	assert.Equal(uint64(0), to.balanceOf(token).Uint64())
	assert.Equal(constants.LDC, from.balanceOf(token).Uint64())
	assert.Equal(constants.LDC-ltx.Gas()*(ctx.Price+100)-1_000_000,
		from.Balance().Uint64())
	assert.Equal(uint64(2), from.Nonce())
	assert.Equal([]uint64{2, 3}, to.ld.NonceTable[cs.Timestamp()])

	jsondata, err := itx.MarshalJSON()
	assert.NoError(err)
	// fmt.Println(string(jsondata))
	assert.Equal(`{"tx":{"type":"TypeExchange","chainID":2357,"nonce":1,"gasTip":100,"gasFeeCap":1000,"from":"0x8db97c7cECe249C2b98bdc0226cc4C2A57bF52fc","to":"0x44171C37Ff5D7B7bb8Dcad5C81f16284A229E641","amount":1000000,"data":{"nonce":1,"sell":"$LDC","receive":"","quota":1000000000000,"minimum":1000000000,"price":1000000,"expire":1000,"payee":"0x44171C37Ff5D7B7bb8Dcad5C81f16284A229E641"}},"sigs":["s5FjLK_TknR_zxjX1O3tcZFs9sCjRnn-xlHTtD8NdGtn-ZMWvtEhC6Lrn639a4tUuDD0ikmG8gg3Y7s9su1T-wFaaCov"],"exSigs":["r1QWsbB9KwOS7Q_kOrVv7jKDnYy6OxvAxhnjOKqWCxxLbH3_YznWRdYWnP6FiTZkxWl_PPRcze8KEw9ltprrbwBdOUm0"],"id":"D9-qUGRL7VXf06Bi212flAX6Vk4AZScB5SBOFC25m78bUxuP"}`, string(jsondata))

	assert.NoError(cs.VerifyState())
}
