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
	bs := NewMockBS(bctx)

	from, err := bs.LoadAccount(util.Signer1.Address())
	assert.NoError(err)
	from.ld.Nonce = 1

	txData := &ld.TxData{
		Type:      ld.TypeTransfer,
		ChainID:   2357,
		Nonce:     1,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      from.id,
	}

	_, err = NewTx(txData.ToTransaction(), true)
	assert.ErrorContains(err, "invalid to")

	txData.To = &constants.GenesisAccount
	_, err = NewTx(txData.ToTransaction(), true)
	assert.ErrorContains(err, "invalid amount")

	txData.Amount = new(big.Int).SetUint64(1000)
	_, err = NewTx(txData.ToTransaction(), true)
	assert.NoError(err)

	// Verify
	assert.NoError(txData.SyntacticVerify())
	tt := txData.ToTransaction()
	tt.Gas = tt.RequiredGas(bctx.FeeConfig().ThresholdGas)
	itx, err := NewTx(tt, true)
	assert.ErrorContains(itx.Verify(bctx, bs), "DeriveSigners: no signature")

	assert.NoError(txData.SignWith(util.Signer1))
	tt = txData.ToTransaction()
	tt.Gas = tt.RequiredGas(bctx.FeeConfig().ThresholdGas)
	itx, err = NewTx(tt, true)
	assert.ErrorContains(itx.Verify(bctx, bs), "insufficient NativeLDC balance")

	tx = itx.(*TxTransfer)
	from.Add(constants.NativeToken, new(big.Int).SetUint64(constants.LDC))
	assert.NoError(itx.Verify(bctx, bs))
	assert.NoError(itx.Accept(bctx, bs))
	ldcbalance := tx.ldc.balanceOf(constants.NativeToken).Uint64()
	assert.Equal(tx.ld.Gas*bctx.Price, ldcbalance)
	minerbalance := tx.miner.balanceOf(constants.NativeToken).Uint64()
	assert.Equal(tx.ld.Gas*100, minerbalance)
	assert.Equal(uint64(1000), tx.to.balanceOf(constants.NativeToken).Uint64())
	accbalance := tx.from.balanceOf(constants.NativeToken).Uint64()
	assert.Equal(constants.LDC-tx.ld.Gas*(bctx.Price+100)-1000, accbalance)
	assert.Equal(uint64(2), tx.from.Nonce())

	jsondata, err := itx.MarshalJSON()
	assert.NoError(err)
	assert.Equal(`{"type":3,"chainID":2357,"nonce":1,"gasTip":100,"gasFeeCap":1000,"from":"0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC","to":"0xFFfFfFffFFfffFFfFFfFFFFFffFFFffffFfFFFfF","amount":1000,"signatures":["5d919bf1bd400e2756bc469e62fa9224370bbfb3fb675006adcc20c03295b7644e4ae94ae8cd8caeb66efe5645e04cd1cc53133590aed0cb5cebcbb6545dde8c00"],"gas":119,"name":"TransferTx","id":"2chgEyDeQWiofeTywV9UZDZ2hQAj5TJRetmKJ1RSvje3Xnn2Mz"}`, string(jsondata))

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
	tt = txData.ToTransaction()
	tt.Gas = tt.RequiredGas(bctx.FeeConfig().ThresholdGas)
	itx, err = NewTx(tt, true)
	assert.ErrorContains(itx.Verify(bctx, bs), "insufficient $LDC balance")

	tx = itx.(*TxTransfer)
	from.Add(token, new(big.Int).SetUint64(constants.LDC))
	assert.NoError(itx.Verify(bctx, bs))
	assert.NoError(itx.Accept(bctx, bs))
	assert.Equal(tx.ld.Gas*bctx.Price, tx.ldc.balanceOf(constants.NativeToken).Uint64()-ldcbalance)
	assert.Equal(tx.ld.Gas*100, tx.miner.balanceOf(constants.NativeToken).Uint64()-minerbalance)
	assert.Equal(uint64(1000), tx.to.balanceOf(token).Uint64())
	assert.Equal(constants.LDC-1000, tx.from.balanceOf(token).Uint64())
	assert.Equal(accbalance-tx.ld.Gas*(bctx.Price+100),
		tx.from.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(uint64(3), tx.from.Nonce())

	jsondata, err = itx.MarshalJSON()
	assert.NoError(err)
	assert.Equal(`{"type":3,"chainID":2357,"nonce":2,"gasTip":100,"gasFeeCap":1000,"from":"0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC","to":"0xFFfFfFffFFfffFFfFFfFFFFFffFFFffffFfFFFfF","token":"$LDC","amount":1000,"signatures":["ef35b1ae4e1cad99704f0884f45221b9609fdfbd5c0df939c78faaecec2dbf616071b3793a5378803821915f65bb1fe172edeaecd0d7ce4ff2db14cf15427e7100"],"gas":143,"name":"TransferTx","id":"jjGWDHh44LQfhUjVy8GNcKAQpGBUPNVzktbdGxU358CVgeGrh"}`, string(jsondata))
}

func TestTxTransferGenesis(t *testing.T) {
	assert := assert.New(t)

	bctx := NewMockBCtx()
	bs := NewMockBS(bctx)

	from, err := bs.LoadAccount(constants.LDCAccount)
	from.Add(constants.NativeToken, bctx.Chain().MaxTotalSupply)
	assert.NoError(err)

	txData := &ld.TxData{
		Type:    ld.TypeTransfer,
		ChainID: bctx.Chain().ChainID,
		From:    from.id,
		To:      &constants.GenesisAccount,
		Amount:  bctx.Chain().MaxTotalSupply,
	}

	itx, err := NewGenesisTx(txData.ToTransaction())
	assert.NoError(err)

	tx := itx.(*TxTransfer)
	assert.NoError(tx.VerifyGenesis(bctx, bs))
	assert.NoError(itx.Accept(bctx, bs))
	assert.Equal(uint64(0), tx.ldc.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(uint64(0), tx.miner.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(bctx.Chain().MaxTotalSupply.Uint64(), tx.to.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(uint64(0), tx.from.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(uint64(1), tx.from.Nonce())

	jsondata, err := itx.MarshalJSON()
	assert.NoError(err)
	assert.Equal(`{"type":3,"chainID":2357,"nonce":0,"gasTip":0,"gasFeeCap":0,"from":"0x0000000000000000000000000000000000000000","to":"0xFFfFfFffFFfffFFfFFfFFFFFffFFFffffFfFFFfF","amount":1000000000000000000,"gas":0,"name":"TransferTx","id":"2ctSeFRMfP214MC7QGHoM9orSRPQ5asyyYRUebsH32o5d5W87D"}`, string(jsondata))
}
