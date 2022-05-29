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

	tx = &TxBase{ld: (&ld.TxData{
		Type:      ld.TypeTransfer,
		ChainID:   2357,
		Nonce:     0,
		GasFeeCap: 0,
		From:      util.EthIDEmpty,
	}).ToTransaction()}
	assert.ErrorContains(tx.SyntacticVerify(), "invalid from")

	tx = &TxBase{ld: (&ld.TxData{
		Type:      ld.TypeTransfer,
		ChainID:   2357,
		Nonce:     0,
		GasFeeCap: 0,
		From:      constants.GenesisAccount,
		To:        &constants.GenesisAccount,
	}).ToTransaction()}
	assert.ErrorContains(tx.SyntacticVerify(), "invalid to")

	// Verify
	bctx := NewMockBCtx()
	bs := NewMockBS(bctx)

	from, err := bs.LoadAccount(util.Signer1.Address())
	assert.NoError(err)
	approver := util.Signer2.Address()
	from.ld.Approver = &approver
	from.ld.Nonce = 1

	tx = &TxBase{ld: (&ld.TxData{
		Type:      ld.TypeTransfer,
		ChainID:   2357,
		Nonce:     0,
		GasFeeCap: bctx.Price - 1,
		From:      from.id,
		To:        &constants.GenesisAccount,
	}).ToTransaction()}
	assert.ErrorContains(tx.Verify(bctx, bs), "invalid gasFeeCap")

	tx = &TxBase{ld: (&ld.TxData{
		Type:      ld.TypeTransfer,
		ChainID:   2357,
		Nonce:     0,
		GasFeeCap: bctx.Price,
		From:      from.id,
		To:        &constants.GenesisAccount,
	}).ToTransaction()}
	assert.ErrorContains(tx.Verify(bctx, bs), "invalid gas")

	txData := &ld.TxData{
		Type:      ld.TypeTransfer,
		ChainID:   2357,
		Nonce:     0,
		GasFeeCap: bctx.Price,
		From:      from.id,
		To:        &constants.GenesisAccount,
		Amount:    new(big.Int).SetUint64(1000),
	}
	tx = &TxBase{ld: txData.ToTransaction()}
	tx.ld.Gas = tx.ld.RequiredGas(bctx.FeeConfig().ThresholdGas)
	assert.NoError(tx.ld.SyntacticVerify())
	assert.NoError(tx.SyntacticVerify())
	assert.ErrorContains(tx.Verify(bctx, bs), "DeriveSigners: no signature")

	assert.NoError(txData.SyntacticVerify())
	assert.NoError(txData.SignWith(util.Signer2))
	tx = &TxBase{ld: txData.ToTransaction()}
	tx.ld.Gas = tx.ld.RequiredGas(bctx.FeeConfig().ThresholdGas)
	assert.NoError(tx.ld.SyntacticVerify())
	assert.NoError(tx.SyntacticVerify())
	assert.ErrorContains(tx.Verify(bctx, bs), "invalid nonce for sender")

	txData = &ld.TxData{
		Type:      ld.TypeTransfer,
		ChainID:   2357,
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
		ChainID:   2357,
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
	assert.Equal(`{"type":3,"chainID":2357,"nonce":1,"gasTip":100,"gasFeeCap":1000,"from":"0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC","to":"0xFFfFfFffFFfffFFfFFfFFFFFffFFFffffFfFFFfF","amount":1000,"signatures":["5d919bf1bd400e2756bc469e62fa9224370bbfb3fb675006adcc20c03295b7644e4ae94ae8cd8caeb66efe5645e04cd1cc53133590aed0cb5cebcbb6545dde8c00","150af39e0304769ae06f83c5952a1c16278c185baec0fd1336461859a16793e913029b75142a3a93647662ce918cd8fed25205e7d1866b49e1e99276ea9701ab01"],"gas":119,"name":"TransferTx","id":"jnMsit3Abuex72x6QeVdQE43TRFAsVXUnt7iHe4vbjQ8HVWL1"}`, string(jsondata))

	from.ld.Approver = nil
	token := ld.MustNewToken("$LDC")
	txData = &ld.TxData{
		Type:      ld.TypeTransfer,
		ChainID:   2357,
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
	assert.Equal(`{"type":3,"chainID":2357,"nonce":2,"gasTip":100,"gasFeeCap":1000,"from":"0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC","to":"0xFFfFfFffFFfffFFfFFfFFFFFffFFFffffFfFFFfF","token":"$LDC","amount":1000,"signatures":["ef35b1ae4e1cad99704f0884f45221b9609fdfbd5c0df939c78faaecec2dbf616071b3793a5378803821915f65bb1fe172edeaecd0d7ce4ff2db14cf15427e7100"],"gas":143,"name":"TransferTx","id":"jjGWDHh44LQfhUjVy8GNcKAQpGBUPNVzktbdGxU358CVgeGrh"}`, string(jsondata))
}
