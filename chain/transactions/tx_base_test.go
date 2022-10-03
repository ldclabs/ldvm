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

func TestTxBase(t *testing.T) {
	assert := assert.New(t)

	// SyntacticVerify
	var tx *TxBase
	assert.ErrorContains(tx.SyntacticVerify(), "nil pointer")
	_, err := tx.MarshalJSON()
	assert.NoError(err)

	ctx := NewMockChainContext()
	cs := ctx.MockChainState()

	sender := util.Signer1.Address()
	approver := util.Signer2.Address()
	senderAcc := cs.MustAccount(sender)
	senderAcc.ld.Approver = &approver
	senderAcc.ld.Nonce = 1

	ltx := &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeTransfer,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasFeeCap: 0,
		From:      util.EthIDEmpty,
	}}
	tx = &TxBase{ld: ltx}
	assert.NoError(ltx.SyntacticVerify())
	assert.ErrorContains(tx.SyntacticVerify(), "invalid from")

	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeTransfer,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasFeeCap: 0,
		From:      constants.GenesisAccount,
		To:        &constants.GenesisAccount,
	}}
	tx = &TxBase{ld: ltx}
	assert.NoError(ltx.SyntacticVerify())
	assert.ErrorContains(tx.SyntacticVerify(), "invalid to")

	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeTransfer,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasFeeCap: 0,
		From:      sender,
		To:        &constants.GenesisAccount,
		Amount:    new(big.Int).SetUint64(1000),
	}}
	tx = &TxBase{ld: ltx}
	assert.NoError(ltx.SyntacticVerify())
	assert.ErrorContains(tx.SyntacticVerify(), "DeriveSigners error: no signature")

	// Verify
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeTransfer,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasFeeCap: ctx.Price - 1,
		From:      sender,
		To:        &constants.GenesisAccount,
	}}
	tx = &TxBase{ld: ltx}
	assert.ErrorContains(tx.verify(ctx, cs), "invalid gasFeeCap")
	assert.ErrorContains(tx.Apply(ctx, cs), "invalid gasFeeCap")

	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeTransfer,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasFeeCap: ctx.Price,
		From:      sender,
		To:        &constants.GenesisAccount,
		Amount:    new(big.Int).SetUint64(1000),
	}}

	tx = &TxBase{ld: ltx}
	assert.NoError(ltx.SignWith(util.Signer2))
	assert.NoError(ltx.SyntacticVerify())
	assert.NoError(tx.SyntacticVerify())
	assert.ErrorContains(tx.verify(ctx, cs), "invalid nonce for sender")
	assert.ErrorContains(tx.Apply(ctx, cs), "invalid nonce for sender")

	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeTransfer,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     1,
		GasFeeCap: ctx.Price,
		From:      sender,
		To:        &constants.GenesisAccount,
		Amount:    new(big.Int).SetUint64(1000),
	}}

	tx = &TxBase{ld: ltx}
	assert.NoError(ltx.SignWith(util.Signer2))
	assert.NoError(ltx.SyntacticVerify())
	assert.NoError(tx.SyntacticVerify())
	assert.ErrorContains(tx.verify(ctx, cs), "invalid signatures for sender")
	assert.ErrorContains(tx.Apply(ctx, cs), "invalid signatures for sender")

	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeTransfer,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     1,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      sender,
		To:        &constants.GenesisAccount,
		Amount:    new(big.Int).SetUint64(1000),
	}}

	tx = &TxBase{ld: ltx}
	assert.NoError(ltx.SignWith(util.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	assert.NoError(tx.SyntacticVerify())
	assert.ErrorContains(tx.verify(ctx, cs), "invalid signature for approver")
	assert.ErrorContains(tx.Apply(ctx, cs), "invalid signature for approver")

	assert.NoError(ltx.SignWith(util.Signer1, util.Signer2))
	tx = &TxBase{ld: ltx}

	assert.NoError(ltx.SyntacticVerify())
	assert.NoError(tx.SyntacticVerify())
	cs.CommitAccounts()
	assert.ErrorContains(tx.verify(ctx, cs), "insufficient NativeLDC balance")
	assert.ErrorContains(tx.Apply(ctx, cs), "insufficient NativeLDC balance")
	cs.CheckoutAccounts()

	senderAcc.Add(constants.NativeToken, new(big.Int).SetUint64(constants.LDC))
	assert.NoError(tx.Apply(ctx, cs))

	senderGas := ltx.Gas()
	assert.Equal(senderGas*ctx.Price,
		tx.ldc.Balance().Uint64())
	assert.Equal(senderGas*100,
		tx.miner.Balance().Uint64())
	assert.Equal(constants.LDC-senderGas*(ctx.Price+100)-1000,
		senderAcc.Balance().Uint64())
	assert.Equal(uint64(1000), tx.to.Balance().Uint64())
	assert.Equal(uint64(2), senderAcc.Nonce())

	jsondata, err := tx.MarshalJSON()
	assert.NoError(err)
	// fmt.Println(string(jsondata))
	assert.Equal(`{"tx":{"type":"TypeTransfer","chainID":2357,"nonce":1,"gasTip":100,"gasFeeCap":1000,"from":"0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC","to":"0xFFfFfFffFFfffFFfFFfFFFFFffFFFffffFfFFFfF","amount":1000},"sigs":["217f378218dd8aed3d660e3e6635c830095922da32389f59c5349e017eb7815e78f4433882d0dffdf31e79f516cc7e294fa60a61c86484be9af6961d5516427a01","70c90b4dee8b2442d8974a568bc0640c858fcaa100d4888daf582e33be5510622e5de01281cc2bc7c4a9269caf959dbca03f7fce68032dd03121d375721c2fbb00"],"id":"m2E9KQYgowM2Koa2GofHQMCepW5HHdsDrbpTRh8nRBGiRWnE5"}`, string(jsondata))

	senderAcc.ld.Approver = nil
	token := ld.MustNewToken("$LDC")
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeTransfer,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     2,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      sender,
		To:        &constants.GenesisAccount,
		Token:     &token,
		Amount:    new(big.Int).SetUint64(1000),
	}}

	tx = &TxBase{ld: ltx}
	assert.NoError(ltx.SignWith(util.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	assert.NoError(tx.SyntacticVerify())
	cs.CommitAccounts()
	assert.ErrorContains(tx.verify(ctx, cs), "insufficient $LDC balance")
	assert.ErrorContains(tx.Apply(ctx, cs), "insufficient $LDC balance")
	cs.CheckoutAccounts()

	senderAcc.Add(token, new(big.Int).SetUint64(constants.LDC))
	assert.NoError(tx.Apply(ctx, cs))

	senderGas += ltx.Gas()
	assert.Equal(senderGas*ctx.Price,
		tx.ldc.Balance().Uint64())
	assert.Equal(senderGas*100,
		tx.miner.Balance().Uint64())
	assert.Equal(constants.LDC-senderGas*(ctx.Price+100)-1000,
		senderAcc.Balance().Uint64())
	assert.Equal(uint64(1000), tx.to.balanceOf(token).Uint64())
	assert.Equal(constants.LDC-1000, tx.from.balanceOf(token).Uint64())
	assert.Equal(uint64(3), tx.from.Nonce())

	jsondata, err = tx.MarshalJSON()
	assert.NoError(err)
	// fmt.Println(string(jsondata))
	assert.Equal(`{"tx":{"type":"TypeTransfer","chainID":2357,"nonce":2,"gasTip":100,"gasFeeCap":1000,"from":"0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC","to":"0xFFfFfFffFFfffFFfFFfFFFFFffFFFffffFfFFFfF","token":"$LDC","amount":1000},"sigs":["b861b75f52a7844ad7e8ce1b6daea144ae69f0b42fdc9ca9a97350d72a5a50d376f8948608e915f7343860b752209a8e71f2defbe127513e6928b3629dc9aa2200"],"id":"24SVjNa9K9Jio4KXfsXFTZtAGqZoMztNcaRHADW9DivjvhvzJV"}`, string(jsondata))

	assert.NoError(cs.VerifyState())
}
