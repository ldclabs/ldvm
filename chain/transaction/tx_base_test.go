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

func TestTxBase(t *testing.T) {
	assert := assert.New(t)

	// SyntacticVerify
	var tx *TxBase
	assert.ErrorContains(tx.SyntacticVerify(), "nil pointer")
	_, err := tx.MarshalJSON()
	assert.NoError(err)

	bctx := NewMockBCtx()
	bs := NewMockBS(bctx)

	from, err := bs.LoadAccount(util.Signer1.Address())
	assert.NoError(err)
	approver := util.Signer2.Address()
	from.ld.Approver = &approver
	from.ld.Nonce = 1

	tx = &TxBase{ld: (&ld.TxData{
		Type:      ld.TypeTransfer,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     0,
		GasFeeCap: 0,
		From:      util.EthIDEmpty,
	}).ToTransaction()}
	assert.ErrorContains(tx.SyntacticVerify(), "invalid from")

	tx = &TxBase{ld: (&ld.TxData{
		Type:      ld.TypeTransfer,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     0,
		GasFeeCap: 0,
		From:      constants.GenesisAccount,
		To:        &constants.GenesisAccount,
	}).ToTransaction()}
	assert.ErrorContains(tx.SyntacticVerify(), "invalid to")

	tx = &TxBase{ld: (&ld.TxData{
		Type:      ld.TypeTransfer,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     0,
		GasFeeCap: 0,
		From:      from.id,
		To:        &constants.GenesisAccount,
		Amount:    new(big.Int).SetUint64(1000),
	}).ToTransaction()}
	assert.NoError(tx.ld.SyntacticVerify())
	assert.ErrorContains(tx.SyntacticVerify(), "DeriveSigners: no signature")

	// Verify
	tx = &TxBase{ld: (&ld.TxData{
		Type:      ld.TypeTransfer,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     0,
		GasFeeCap: bctx.Price - 1,
		From:      from.id,
		To:        &constants.GenesisAccount,
	}).ToTransaction()}
	assert.ErrorContains(tx.Verify(bctx, bs), "invalid gasFeeCap")

	tx = &TxBase{ld: (&ld.TxData{
		Type:      ld.TypeTransfer,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     0,
		GasFeeCap: bctx.Price,
		From:      from.id,
		To:        &constants.GenesisAccount,
	}).ToTransaction()}
	assert.ErrorContains(tx.Verify(bctx, bs), "invalid gas")

	txData := &ld.TxData{
		Type:      ld.TypeTransfer,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     0,
		GasFeeCap: bctx.Price,
		From:      from.id,
		To:        &constants.GenesisAccount,
		Amount:    new(big.Int).SetUint64(1000),
	}
	assert.NoError(txData.SyntacticVerify())
	assert.NoError(txData.SignWith(util.Signer2))
	tx = &TxBase{ld: txData.ToTransaction()}
	tx.ld.Gas = tx.ld.RequiredGas(bctx.FeeConfig().ThresholdGas)
	assert.NoError(tx.ld.SyntacticVerify())
	assert.NoError(tx.SyntacticVerify())
	assert.ErrorContains(tx.Verify(bctx, bs), "invalid nonce for sender")

	txData = &ld.TxData{
		Type:      ld.TypeTransfer,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     1,
		GasFeeCap: bctx.Price,
		From:      from.id,
		To:        &constants.GenesisAccount,
		Amount:    new(big.Int).SetUint64(1000),
	}
	assert.NoError(txData.SyntacticVerify())
	assert.NoError(txData.SignWith(util.Signer2))
	tx = &TxBase{ld: txData.ToTransaction()}
	tx.ld.Gas = tx.ld.RequiredGas(bctx.FeeConfig().ThresholdGas)
	assert.NoError(tx.ld.SyntacticVerify())
	assert.NoError(tx.SyntacticVerify())
	assert.ErrorContains(tx.Verify(bctx, bs), "invalid signatures for sender")

	txData = &ld.TxData{
		Type:      ld.TypeTransfer,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     1,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      from.id,
		To:        &constants.GenesisAccount,
		Amount:    new(big.Int).SetUint64(1000),
	}
	assert.NoError(txData.SyntacticVerify())
	assert.NoError(txData.SignWith(util.Signer1))
	tx = &TxBase{ld: txData.ToTransaction()}
	tx.ld.Gas = tx.ld.RequiredGas(bctx.FeeConfig().ThresholdGas)
	assert.NoError(tx.ld.SyntacticVerify())
	assert.NoError(tx.SyntacticVerify())
	assert.ErrorContains(tx.Verify(bctx, bs), "invalid signature for approver")

	assert.NoError(txData.SignWith(util.Signer2))
	tx = &TxBase{ld: txData.ToTransaction()}
	tx.ld.Gas = tx.ld.RequiredGas(bctx.FeeConfig().ThresholdGas)
	assert.NoError(tx.ld.SyntacticVerify())
	assert.NoError(tx.SyntacticVerify())
	assert.ErrorContains(tx.Verify(bctx, bs), "insufficient NativeLDC balance")

	from.Add(constants.NativeToken, new(big.Int).SetUint64(constants.LDC))
	assert.NoError(tx.Verify(bctx, bs))
	assert.NoError(tx.Accept(bctx, bs))
	ldcbalance := tx.ldc.balanceOf(constants.NativeToken).Uint64()
	assert.Equal(tx.ld.Gas*bctx.Price, ldcbalance)
	minerbalance := tx.miner.balanceOf(constants.NativeToken).Uint64()
	assert.Equal(tx.ld.Gas*100, minerbalance)
	assert.Equal(uint64(1000), tx.to.balanceOf(constants.NativeToken).Uint64())
	accbalance := tx.from.balanceOf(constants.NativeToken).Uint64()
	assert.Equal(constants.LDC-tx.ld.Gas*(bctx.Price+100)-1000, accbalance)
	assert.Equal(uint64(2), tx.from.Nonce())

	jsondata, err := tx.MarshalJSON()
	assert.NoError(err)
	assert.Equal(`{"type":2,"chainID":2357,"nonce":1,"gasTip":100,"gasFeeCap":1000,"from":"0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC","to":"0xFFfFfFffFFfffFFfFFfFFFFFffFFFffffFfFFFfF","amount":1000,"signatures":["217f378218dd8aed3d660e3e6635c830095922da32389f59c5349e017eb7815e78f4433882d0dffdf31e79f516cc7e294fa60a61c86484be9af6961d5516427a01","70c90b4dee8b2442d8974a568bc0640c858fcaa100d4888daf582e33be5510622e5de01281cc2bc7c4a9269caf959dbca03f7fce68032dd03121d375721c2fbb00"],"gas":119,"name":"TransferTx","id":"KtpU3iErfEz63uBEhoWPLk816UhNF3kjUj1dV3Zi6rfBPqewg"}`, string(jsondata))

	from.ld.Approver = nil
	token := ld.MustNewToken("$LDC")
	txData = &ld.TxData{
		Type:      ld.TypeTransfer,
		ChainID:   bctx.Chain().ChainID,
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
	tx = &TxBase{ld: txData.ToTransaction()}
	tx.ld.Gas = tx.ld.RequiredGas(bctx.FeeConfig().ThresholdGas)
	assert.NoError(tx.ld.SyntacticVerify())
	assert.NoError(tx.SyntacticVerify())
	assert.ErrorContains(tx.Verify(bctx, bs), "insufficient $LDC balance")

	from.Add(token, new(big.Int).SetUint64(constants.LDC))
	assert.NoError(tx.Verify(bctx, bs))
	assert.NoError(tx.Accept(bctx, bs))
	assert.Equal(tx.ld.Gas*bctx.Price, tx.ldc.balanceOf(constants.NativeToken).Uint64()-ldcbalance)
	assert.Equal(tx.ld.Gas*100, tx.miner.balanceOf(constants.NativeToken).Uint64()-minerbalance)
	assert.Equal(uint64(1000), tx.to.balanceOf(token).Uint64())
	assert.Equal(constants.LDC-1000, tx.from.balanceOf(token).Uint64())
	assert.Equal(accbalance-tx.ld.Gas*(bctx.Price+100),
		tx.from.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(uint64(3), tx.from.Nonce())

	jsondata, err = tx.MarshalJSON()
	assert.NoError(err)
	assert.Equal(`{"type":2,"chainID":2357,"nonce":2,"gasTip":100,"gasFeeCap":1000,"from":"0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC","to":"0xFFfFfFffFFfffFFfFFfFFFFFffFFFffffFfFFFfF","token":"$LDC","amount":1000,"signatures":["b861b75f52a7844ad7e8ce1b6daea144ae69f0b42fdc9ca9a97350d72a5a50d376f8948608e915f7343860b752209a8e71f2defbe127513e6928b3629dc9aa2200"],"gas":143,"name":"TransferTx","id":"sJV9ndy4B654Nmt6YVKsybB3DSe9GRFNvqTL1cab2s4cG34rm"}`, string(jsondata))

	assert.NoError(bs.VerifyState())
}
