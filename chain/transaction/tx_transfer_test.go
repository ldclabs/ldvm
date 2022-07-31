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

func TestTxTransfer(t *testing.T) {
	assert := assert.New(t)

	// SyntacticVerify
	tx := &TxTransfer{}
	assert.ErrorContains(tx.SyntacticVerify(), "nil pointer")
	_, err := tx.MarshalJSON()
	assert.NoError(err)

	bctx := NewMockBCtx()
	bs := bctx.MockBS()

	from := bs.MustAccount(util.Signer1.Address())
	from.ld.Nonce = 1

	txData := &ld.TxData{
		Type:      ld.TypeTransfer,
		ChainID:   bctx.ChainConfig().ChainID,
		Nonce:     1,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      from.id,
	}
	assert.NoError(txData.SyntacticVerify())
	_, err = NewTx2(txData.ToTransaction())
	assert.ErrorContains(err, "DeriveSigners error: no signature")

	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx2(txData.ToTransaction())
	assert.ErrorContains(err, "invalid to")

	txData = &ld.TxData{
		Type:      ld.TypeTransfer,
		ChainID:   bctx.ChainConfig().ChainID,
		Nonce:     1,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      from.id,
		To:        &constants.GenesisAccount,
	}
	assert.NoError(txData.SyntacticVerify())
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx2(txData.ToTransaction())
	assert.ErrorContains(err, "invalid amount")

	txData = &ld.TxData{
		Type:      ld.TypeTransfer,
		ChainID:   bctx.ChainConfig().ChainID,
		Nonce:     1,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      from.id,
		To:        &constants.GenesisAccount,
		Amount:    new(big.Int).SetUint64(1000),
	}

	assert.NoError(txData.SyntacticVerify())
	assert.NoError(txData.SignWith(util.Signer1))
	assert.NoError(txData.SyntacticVerify())
	tt := txData.ToTransaction()
	itx, err := NewTx2(tt)
	bs.CommitAccounts()
	assert.ErrorContains(itx.Apply(bctx, bs), "insufficient NativeLDC balance")
	bs.CheckoutAccounts()

	from.Add(constants.NativeToken, new(big.Int).SetUint64(constants.LDC))
	assert.NoError(itx.Apply(bctx, bs))

	fromGas := tt.Gas()
	assert.Equal(fromGas*bctx.Price,
		itx.(*TxTransfer).ldc.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(fromGas*100,
		itx.(*TxTransfer).miner.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(uint64(1000),
		itx.(*TxTransfer).to.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(constants.LDC-fromGas*(bctx.Price+100)-1000,
		from.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(uint64(2), from.Nonce())

	jsondata, err := itx.MarshalJSON()
	assert.NoError(err)
	// fmt.Println(string(jsondata))
	assert.Equal(`{"type":"TypeTransfer","chainID":2357,"nonce":1,"gasTip":100,"gasFeeCap":1000,"from":"0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC","to":"0xFFfFfFffFFfffFFfFFfFFFFFffFFFffffFfFFFfF","amount":1000,"signatures":["217f378218dd8aed3d660e3e6635c830095922da32389f59c5349e017eb7815e78f4433882d0dffdf31e79f516cc7e294fa60a61c86484be9af6961d5516427a01"],"id":"Hhp7vYTnhNemXC7N6w9N9DiUV8vuvPzc6TWDbPa2cpy4gHYwo"}`, string(jsondata))

	token := ld.MustNewToken("$LDC")
	txData = &ld.TxData{
		Type:      ld.TypeTransfer,
		ChainID:   bctx.ChainConfig().ChainID,
		Nonce:     2,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      from.id,
		To:        &constants.GenesisAccount,
		Token:     &token,
		Amount:    new(big.Int).SetUint64(1000),
	}
	assert.NoError(txData.SyntacticVerify())
	assert.NoError(txData.SignWith(util.Signer1))
	tt = txData.ToTransaction()
	itx, err = NewTx2(tt)
	bs.CommitAccounts()
	assert.ErrorContains(itx.Apply(bctx, bs), "insufficient $LDC balance")
	bs.CheckoutAccounts()

	from.Add(token, new(big.Int).SetUint64(constants.LDC))
	assert.NoError(itx.Apply(bctx, bs))

	fromGas += tt.Gas()
	assert.Equal(fromGas*bctx.Price,
		itx.(*TxTransfer).ldc.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(fromGas*100,
		itx.(*TxTransfer).miner.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(uint64(1000), itx.(*TxTransfer).to.balanceOf(token).Uint64())
	assert.Equal(constants.LDC-1000, from.balanceOf(token).Uint64())
	assert.Equal(constants.LDC-fromGas*(bctx.Price+100)-1000,
		from.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(uint64(3), from.Nonce())

	jsondata, err = itx.MarshalJSON()
	assert.NoError(err)
	// fmt.Println(string(jsondata))
	assert.Equal(`{"type":"TypeTransfer","chainID":2357,"nonce":2,"gasTip":100,"gasFeeCap":1000,"from":"0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC","to":"0xFFfFfFffFFfffFFfFFfFFFFFffFFFffffFfFFFfF","token":"$LDC","amount":1000,"signatures":["b861b75f52a7844ad7e8ce1b6daea144ae69f0b42fdc9ca9a97350d72a5a50d376f8948608e915f7343860b752209a8e71f2defbe127513e6928b3629dc9aa2200"],"id":"sJV9ndy4B654Nmt6YVKsybB3DSe9GRFNvqTL1cab2s4cG34rm"}`, string(jsondata))

	// support 0 amount
	txData = &ld.TxData{
		Type:      ld.TypeTransfer,
		ChainID:   bctx.ChainConfig().ChainID,
		Nonce:     3,
		GasTip:    0,
		GasFeeCap: bctx.Price,
		From:      from.id,
		To:        &constants.GenesisAccount,
		Amount:    new(big.Int).SetUint64(0),
		Data:      []byte(`"some message"`),
	}
	assert.NoError(txData.SyntacticVerify())
	assert.NoError(txData.SignWith(util.Signer1))
	tt = txData.ToTransaction()
	itx, err = NewTx2(tt)
	assert.NoError(itx.Apply(bctx, bs), "should support 0 amount")

	assert.NoError(bs.VerifyState())
}

func TestTxTransferGenesis(t *testing.T) {
	assert := assert.New(t)

	bctx := NewMockBCtx()
	bs := bctx.MockBS()

	from := bs.MustAccount(constants.LDCAccount)
	from.Add(constants.NativeToken, bctx.ChainConfig().MaxTotalSupply)

	txData := &ld.TxData{
		Type:    ld.TypeTransfer,
		ChainID: bctx.ChainConfig().ChainID,
		From:    from.id,
		To:      &constants.GenesisAccount,
		Amount:  bctx.ChainConfig().MaxTotalSupply,
	}

	itx, err := NewGenesisTx(txData.ToTransaction())
	assert.NoError(err)

	assert.NoError(itx.(*TxTransfer).ApplyGenesis(bctx, bs))

	assert.Equal(uint64(0),
		itx.(*TxTransfer).ldc.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(uint64(0),
		itx.(*TxTransfer).miner.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(bctx.ChainConfig().MaxTotalSupply.Uint64(),
		itx.(*TxTransfer).to.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(uint64(0),
		itx.(*TxTransfer).from.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(uint64(1),
		itx.(*TxTransfer).from.Nonce())

	jsondata, err := itx.MarshalJSON()
	assert.NoError(err)
	// fmt.Println(string(jsondata))
	assert.Equal(`{"type":"TypeTransfer","chainID":2357,"nonce":0,"gasTip":0,"gasFeeCap":0,"from":"0x0000000000000000000000000000000000000000","to":"0xFFfFfFffFFfffFFfFFfFFFFFffFFFffffFfFFFfF","amount":1000000000000000000,"id":"rCfqgA8NHYcjHxvkTURAmsmMrCZDps31H3iMDLqHkgDWGiA73"}`, string(jsondata))

	assert.NoError(bs.VerifyState())
}
