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

	from := cs.MustAccount(util.Signer1.Address())
	from.ld.Nonce = 1
	to := cs.MustAccount(util.Signer2.Address())

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
	assert.ErrorContains(err, "DeriveSigners error: no signature")

	assert.NoError(ltx.SignWith(util.Signer1))
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

	assert.NoError(ltx.SignWith(util.Signer1))
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

	assert.NoError(ltx.SignWith(util.Signer1))
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

	assert.NoError(ltx.SignWith(util.Signer1))
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

	assert.NoError(ltx.SignWith(util.Signer1))
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

	assert.NoError(ltx.SignWith(util.Signer1))
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

	assert.NoError(ltx.SignWith(util.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err,
		"invalid from, expected 0xFFfFfFffFFfffFFfFFfFFFFFffFFFffffFfFFFfF, got 0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC")

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

	assert.NoError(ltx.SignWith(util.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err,
		"invalid to, expected 0x44171C37Ff5D7B7bb8dcad5C81f16284A229e641, got 0xFFfFfFffFFfffFFfFFfFFFFFffFFFffffFfFFFfF")

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

	assert.NoError(ltx.SignWith(util.Signer1))
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

	assert.NoError(ltx.SignWith(util.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	ltx.Timestamp = cs.Timestamp() + 1
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "data expired")
	ltx.Timestamp = 1
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "DeriveSigners error: no signature")

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

	assert.NoError(ltx.SignWith(util.Signer1))
	assert.NoError(ltx.ExSignWith(util.Signer1))
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

	assert.NoError(ltx.SignWith(util.Signer1))
	assert.NoError(ltx.ExSignWith(util.Signer2))
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
	assert.Equal(`{"tx":{"type":"TypeExchange","chainID":2357,"nonce":1,"gasTip":100,"gasFeeCap":1000,"from":"0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC","to":"0x44171C37Ff5D7B7bb8dcad5C81f16284A229e641","amount":1000000,"data":{"nonce":1,"sell":"$LDC","receive":"","quota":1000000000000,"minimum":1000000000,"price":1000000,"expire":1000,"payee":"0x44171C37Ff5D7B7bb8dcad5C81f16284A229e641"}},"sigs":["44a43280624de00dd1847a6fe933ff21cf11032659aa9ef83d3326fa674dea9d116d623520c7e6e6d7af569c3bf887ada476e757cc13e506348df7b835c60c5601"],"exSigs":["af5416b1b07d2b0392ed0fe43ab56fee32839d8cba3b1bc0c619e338aa960b1c4b6c7dff6339d645d6169cfe85893664c5697f3cf45ccdef0a130f65b69aeb6f00"],"id":"RgbBMKUxxZvNVEP4WR4aADGuUC4rwPdqXRsg1nAjEemkroUKw"}`, string(jsondata))

	assert.NoError(cs.VerifyState())
}
