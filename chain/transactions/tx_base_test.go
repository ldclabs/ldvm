// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package transactions

import (
	"math/big"
	"testing"

	"github.com/ldclabs/ldvm/constants"
	"github.com/ldclabs/ldvm/ld"
	"github.com/ldclabs/ldvm/util"
	"github.com/ldclabs/ldvm/util/signer"
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

	sender := signer.Signer1.Key().Address()
	approver := signer.Signer2.Key()
	senderAcc := cs.MustAccount(sender)
	senderAcc.ld.Approver = approver
	senderAcc.ld.Nonce = 1

	ltx := &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeTransfer,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasFeeCap: 0,
		From:      util.AddressEmpty,
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
	assert.ErrorContains(tx.SyntacticVerify(), "no signatures")

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
	assert.NoError(ltx.SignWith(signer.Signer2))
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
	assert.NoError(ltx.SignWith(signer.Signer2))
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
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	assert.NoError(tx.SyntacticVerify())
	assert.ErrorContains(tx.verify(ctx, cs), "invalid signature for approver")
	assert.ErrorContains(tx.Apply(ctx, cs), "invalid signature for approver")

	assert.NoError(ltx.SignWith(signer.Signer1, signer.Signer2))
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
	assert.Equal(`{"tx":{"type":"TypeTransfer","chainID":2357,"nonce":1,"gasTip":100,"gasFeeCap":1000,"from":"0x8db97c7cECe249C2b98bdc0226cc4C2A57bF52fc","to":"0xFFfFFFfFfffFFfFFffFFFfFfFffFFFfffFfFFFff","amount":1000},"sigs":["IX83ghjdiu09Zg4-ZjXIMAlZItoyOJ9ZxTSeAX63gV549EM4gtDf_fMeefUWzH4pT6YKYchkhL6a9pYdVRZCegGBPL0O","cMkLTe6LJELYl0pWi8BkDIWPyqEA1IiNr1guM75VEGIuXeASgcwrx8SpJpyvlZ28oD9_zmgDLdAxIdN1chwvuwBGm81i"],"id":"Y_SVR0ZOS38KfuTP6Eqrhe_JvXNusA4tTlyjwWNqIFGcaSHX"}`, string(jsondata))

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
	assert.NoError(ltx.SignWith(signer.Signer1))
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
	assert.Equal(`{"tx":{"type":"TypeTransfer","chainID":2357,"nonce":2,"gasTip":100,"gasFeeCap":1000,"from":"0x8db97c7cECe249C2b98bdc0226cc4C2A57bF52fc","to":"0xFFfFFFfFfffFFfFFffFFFfFfFffFFFfffFfFFFff","token":"$LDC","amount":1000},"sigs":["uGG3X1KnhErX6M4bba6hRK5p8LQv3JypqXNQ1ypaUNN2-JSGCOkV9zQ4YLdSIJqOcfLe--EnUT5pKLNincmqIgBsXXMx"],"id":"i4GN43OdfBiEOseH1fP8u0sg7LoUQ4TF2yLgIteaOKxhHz1v"}`, string(jsondata))

	assert.NoError(cs.VerifyState())
}
