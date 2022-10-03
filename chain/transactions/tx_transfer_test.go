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

func TestTxTransfer(t *testing.T) {
	assert := assert.New(t)

	// SyntacticVerify
	tx := &TxTransfer{}
	assert.ErrorContains(tx.SyntacticVerify(), "nil pointer")
	_, err := tx.MarshalJSON()
	assert.NoError(err)

	ctx := NewMockChainContext()
	cs := ctx.MockChainState()

	from := cs.MustAccount(util.Signer1.Address())
	from.ld.Nonce = 1

	ltx := &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeTransfer,
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
		Type:      ld.TypeTransfer,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     1,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      from.id,
		To:        &constants.GenesisAccount,
	}}

	assert.NoError(ltx.SignWith(util.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "invalid amount")

	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeTransfer,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     1,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      from.id,
		To:        &constants.GenesisAccount,
		Amount:    new(big.Int).SetUint64(1000),
	}}

	assert.NoError(ltx.SignWith(util.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	itx, err := NewTx(ltx)
	assert.NoError(err)
	cs.CommitAccounts()
	assert.ErrorContains(itx.Apply(ctx, cs), "insufficient NativeLDC balance")
	cs.CheckoutAccounts()

	from.Add(constants.NativeToken, new(big.Int).SetUint64(constants.LDC))
	assert.NoError(itx.Apply(ctx, cs))

	fromGas := ltx.Gas()
	assert.Equal(fromGas*ctx.Price,
		itx.(*TxTransfer).ldc.Balance().Uint64())
	assert.Equal(fromGas*100,
		itx.(*TxTransfer).miner.Balance().Uint64())
	assert.Equal(uint64(1000),
		itx.(*TxTransfer).to.Balance().Uint64())
	assert.Equal(constants.LDC-fromGas*(ctx.Price+100)-1000,
		from.Balance().Uint64())
	assert.Equal(uint64(2), from.Nonce())

	jsondata, err := itx.MarshalJSON()
	assert.NoError(err)
	// fmt.Println(string(jsondata))
	assert.Equal(`{"tx":{"type":"TypeTransfer","chainID":2357,"nonce":1,"gasTip":100,"gasFeeCap":1000,"from":"0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC","to":"0xFFfFfFffFFfffFFfFFfFFFFFffFFFffffFfFFFfF","amount":1000},"sigs":["217f378218dd8aed3d660e3e6635c830095922da32389f59c5349e017eb7815e78f4433882d0dffdf31e79f516cc7e294fa60a61c86484be9af6961d5516427a01"],"id":"2hd1qLmkBJMyMdAoonML7Anp7gq9dRehunS3trLkxKdYgeA9Cv"}`, string(jsondata))

	token := ld.MustNewToken("$LDC")
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeTransfer,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     2,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      from.id,
		To:        &constants.GenesisAccount,
		Token:     &token,
		Amount:    new(big.Int).SetUint64(1000),
	}}

	assert.NoError(ltx.SignWith(util.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	itx, err = NewTx(ltx)
	assert.NoError(err)
	cs.CommitAccounts()
	assert.ErrorContains(itx.Apply(ctx, cs), "insufficient $LDC balance")
	cs.CheckoutAccounts()

	from.Add(token, new(big.Int).SetUint64(constants.LDC))
	assert.NoError(itx.Apply(ctx, cs))

	fromGas += ltx.Gas()
	assert.Equal(fromGas*ctx.Price,
		itx.(*TxTransfer).ldc.Balance().Uint64())
	assert.Equal(fromGas*100,
		itx.(*TxTransfer).miner.Balance().Uint64())
	assert.Equal(uint64(1000), itx.(*TxTransfer).to.balanceOf(token).Uint64())
	assert.Equal(constants.LDC-1000, from.balanceOf(token).Uint64())
	assert.Equal(constants.LDC-fromGas*(ctx.Price+100)-1000,
		from.Balance().Uint64())
	assert.Equal(uint64(3), from.Nonce())

	jsondata, err = itx.MarshalJSON()
	assert.NoError(err)
	// fmt.Println(string(jsondata))
	assert.Equal(`{"tx":{"type":"TypeTransfer","chainID":2357,"nonce":2,"gasTip":100,"gasFeeCap":1000,"from":"0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC","to":"0xFFfFfFffFFfffFFfFFfFFFFFffFFFffffFfFFFfF","token":"$LDC","amount":1000},"sigs":["b861b75f52a7844ad7e8ce1b6daea144ae69f0b42fdc9ca9a97350d72a5a50d376f8948608e915f7343860b752209a8e71f2defbe127513e6928b3629dc9aa2200"],"id":"24SVjNa9K9Jio4KXfsXFTZtAGqZoMztNcaRHADW9DivjvhvzJV"}`, string(jsondata))

	// support 0 amount
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeTransfer,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     3,
		GasTip:    0,
		GasFeeCap: ctx.Price,
		From:      from.id,
		To:        &constants.GenesisAccount,
		Amount:    new(big.Int).SetUint64(0),
		Data:      []byte(`"some message"`),
	}}
	assert.NoError(ltx.SignWith(util.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	itx, err = NewTx(ltx)
	assert.NoError(err)
	assert.NoError(itx.Apply(ctx, cs), "should support 0 amount")

	assert.NoError(cs.VerifyState())
}

func TestTxTransferGenesis(t *testing.T) {
	assert := assert.New(t)

	ctx := NewMockChainContext()
	cs := ctx.MockChainState()

	from := cs.MustAccount(constants.LDCAccount)
	from.Add(constants.NativeToken, ctx.ChainConfig().MaxTotalSupply)

	ltx := &ld.Transaction{Tx: ld.TxData{
		Type:    ld.TypeTransfer,
		ChainID: ctx.ChainConfig().ChainID,
		From:    from.id,
		To:      &constants.GenesisAccount,
		Amount:  ctx.ChainConfig().MaxTotalSupply,
	}}

	itx, err := NewGenesisTx(ltx)
	assert.NoError(err)

	assert.NoError(itx.(*TxTransfer).ApplyGenesis(ctx, cs))

	assert.Equal(uint64(0),
		itx.(*TxTransfer).ldc.Balance().Uint64())
	assert.Equal(uint64(0),
		itx.(*TxTransfer).miner.Balance().Uint64())
	assert.Equal(ctx.ChainConfig().MaxTotalSupply.Uint64(),
		itx.(*TxTransfer).to.Balance().Uint64())
	assert.Equal(uint64(0),
		itx.(*TxTransfer).from.Balance().Uint64())
	assert.Equal(uint64(1),
		itx.(*TxTransfer).from.Nonce())

	jsondata, err := itx.MarshalJSON()
	assert.NoError(err)
	// fmt.Println(string(jsondata))
	assert.Equal(`{"tx":{"type":"TypeTransfer","chainID":2357,"nonce":0,"gasTip":0,"gasFeeCap":0,"from":"0x0000000000000000000000000000000000000000","to":"0xFFfFfFffFFfffFFfFFfFFFFFffFFFffffFfFFFfF","amount":1000000000000000000},"id":"cbUdLXs4Mjww56sMZaYSWMNbwmaYMvYDbrhs2AqrV4r3sZx7P"}`, string(jsondata))

	assert.NoError(cs.VerifyState())
}

func TestTxTransferFromNoKeeperAccount(t *testing.T) {
	assert := assert.New(t)

	ctx := NewMockChainContext()
	cs := ctx.MockChainState()

	signer1 := util.NewSigner()
	from := cs.MustAccount(signer1.Address())
	from.Add(constants.NativeToken, new(big.Int).SetUint64(constants.LDC))
	to := cs.MustAccount(util.EthID{1, 2, 3})

	ltx := &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeTransfer,
		ChainID:   ctx.ChainConfig().ChainID,
		GasFeeCap: ctx.Price,
		From:      from.ID(),
		To:        &to.id,
		Amount:    new(big.Int).SetUint64(constants.MilliLDC * 500),
	}}

	assert.NoError(ltx.SignWith(signer1))
	assert.NoError(ltx.SyntacticVerify())
	itx, err := NewTx(ltx)
	assert.NoError(err)
	assert.NoError(itx.Apply(ctx, cs))

	fromGas := ltx.Gas()
	assert.Equal(fromGas*ctx.Price,
		itx.(*TxTransfer).ldc.Balance().Uint64())
	assert.Equal(fromGas*0,
		itx.(*TxTransfer).miner.Balance().Uint64())
	assert.Equal(constants.MilliLDC*500, to.Balance().Uint64())
	assert.Equal(constants.LDC-fromGas*(ctx.Price+0)-constants.MilliLDC*500,
		from.Balance().Uint64())
	assert.Equal(uint64(1), from.Nonce())
	assert.True(from.IsEmpty())
	assert.True(to.IsEmpty())

	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeTransfer,
		ChainID:   ctx.ChainConfig().ChainID,
		GasFeeCap: ctx.Price,
		From:      to.ID(),
		To:        &from.id,
		Amount:    new(big.Int).SetUint64(constants.MilliLDC * 100),
	}}

	assert.NoError(ltx.SignWith(signer1))
	assert.NoError(ltx.SyntacticVerify())
	itx, err = NewTx(ltx)
	assert.NoError(err)
	assert.ErrorContains(itx.Apply(ctx, cs), "TxBase.Apply error: invalid signatures for sender")

	assert.NoError(cs.VerifyState())
}
