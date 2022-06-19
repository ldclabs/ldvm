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

func TestTxCloseLending(t *testing.T) {
	assert := assert.New(t)

	// SyntacticVerify
	tx := &TxCloseLending{}
	assert.ErrorContains(tx.SyntacticVerify(), "nil pointer")
	_, err := tx.MarshalJSON()
	assert.NoError(err)

	bctx := NewMockBCtx()
	bs := NewMockBS(bctx)
	token := ld.MustNewToken("$LDC")

	from, err := bs.LoadAccount(util.Signer1.Address())
	assert.NoError(err)

	txData := &ld.TxData{
		Type:      ld.TypeCloseLending,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      from.id,
	}
	assert.NoError(txData.SyntacticVerify())
	_, err = NewTx(txData.ToTransaction(), true)
	assert.ErrorContains(err, "DeriveSigners error: no signature")

	txData = &ld.TxData{
		Type:      ld.TypeCloseLending,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      from.id,
		To:        &constants.GenesisAccount,
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx(txData.ToTransaction(), true)
	assert.ErrorContains(err, "invalid to, should be nil")

	txData = &ld.TxData{
		Type:      ld.TypeCloseLending,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      from.id,
		Token:     &constants.NativeToken,
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx(txData.ToTransaction(), true)
	assert.ErrorContains(err, "invalid token, should be nil")

	txData = &ld.TxData{
		Type:      ld.TypeCloseLending,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      from.id,
		Amount:    big.NewInt(1),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx(txData.ToTransaction(), true)
	assert.ErrorContains(err, "nil to together with amount")

	txData = &ld.TxData{
		Type:      ld.TypeCloseLending,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      from.id,
	}
	assert.NoError(txData.SignWith(util.Signer1))
	tt := txData.ToTransaction()
	tt.Timestamp = bs.Timestamp()
	itx, err := NewTx(tt, true)
	assert.NoError(err)
	assert.ErrorContains(itx.Verify(bctx, bs), "invalid gas, expected 1137, got 0")

	tt.Gas = tt.RequiredGas(bctx.FeeConfig().ThresholdGas)
	itx, err = NewTx(tt, true)
	assert.NoError(err)
	assert.ErrorContains(itx.Verify(bctx, bs),
		"insufficient NativeLDC balance, expected 1250700, got 0")

	from.Add(constants.NativeToken, new(big.Int).SetUint64(constants.LDC))
	assert.ErrorContains(itx.Verify(bctx, bs),
		"Account(0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC).CheckCloseLending error: invalid lending")

	input := &ld.LendingConfig{
		Token:           token,
		DailyInterest:   100,
		OverdueInterest: 10,
		MinAmount:       new(big.Int).SetUint64(constants.LDC),
		MaxAmount:       new(big.Int).SetUint64(constants.LDC * 10),
	}
	assert.NoError(input.SyntacticVerify())
	txData = &ld.TxData{
		Type:      ld.TypeOpenLending,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      from.id,
		Data:      ld.MustMarshal(input),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	tt = txData.ToTransaction()
	tt.Timestamp = bs.Timestamp()
	tt.Gas = tt.RequiredGas(bctx.FeeConfig().ThresholdGas)
	gas1 := tt.Gas
	itx, err = NewTx(tt, true)
	assert.NoError(err)
	assert.NoError(itx.Verify(bctx, bs))
	assert.NoError(itx.Accept(bctx, bs))

	assert.Equal(uint64(1), from.Nonce())
	assert.NotNil(from.ld.Lending)
	assert.Equal(token, from.ld.Lending.Token)
	assert.Equal(make(map[util.EthID]*ld.LendingEntry), from.ld.LendingLedger)

	txData = &ld.TxData{
		Type:      ld.TypeCloseLending,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     1,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      from.id,
	}
	assert.NoError(txData.SignWith(util.Signer1))
	tt = txData.ToTransaction()
	tt.Timestamp = bs.Timestamp()
	tt.Gas = tt.RequiredGas(bctx.FeeConfig().ThresholdGas)
	itx, err = NewTx(tt, true)
	assert.NoError(err)
	assert.NoError(itx.Verify(bctx, bs))
	assert.NoError(itx.Accept(bctx, bs))

	tx = itx.(*TxCloseLending)
	assert.Equal((gas1+tx.ld.Gas)*bctx.Price, tx.ldc.balanceOf(constants.NativeToken).Uint64())
	assert.Equal((gas1+tx.ld.Gas)*100, tx.miner.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(constants.LDC-(gas1+tx.ld.Gas)*(bctx.Price+100),
		from.balanceOf(constants.NativeToken).Uint64())
	assert.Nil(from.ld.Lending)
	assert.Nil(from.ld.LendingLedger)

	jsondata, err := itx.MarshalJSON()
	assert.NoError(err)
	// fmt.Println(string(jsondata))
	assert.Equal(`{"type":"TypeCloseLending","chainID":2357,"nonce":1,"gasTip":100,"gasFeeCap":1000,"from":"0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC","signatures":["ceb99eaf8a7825a8c9d26eeeea7fe3672eede4e0a4d0c18cb50638c00fca73831d50efafc6533b29258b3938811d5fe02a15688911f9958b121c7e14f6e01e8500"],"gas":1137,"id":"242SMs7S5niL4YYNsaNrQgASCKh4Brcnep9sjhxv7toj8R1PNB"}`, string(jsondata))

	txData = &ld.TxData{
		Type:      ld.TypeCloseLending,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     2,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      from.id,
	}
	assert.NoError(txData.SignWith(util.Signer1))
	tt = txData.ToTransaction()
	tt.Timestamp = bs.Timestamp()
	tt.Gas = tt.RequiredGas(bctx.FeeConfig().ThresholdGas)
	itx, err = NewTx(tt, true)
	assert.NoError(err)
	assert.ErrorContains(itx.Verify(bctx, bs),
		"Account(0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC).CheckCloseLending error: invalid lending")

	assert.NoError(bs.VerifyState())
}
