// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package transaction

import (
	"math/big"
	"testing"

	"github.com/ldclabs/ldvm/constants"
	"github.com/ldclabs/ldvm/ld"
	"github.com/ldclabs/ldvm/ld/service"
	"github.com/ldclabs/ldvm/util"
	"github.com/stretchr/testify/assert"
)

func TestTxUpdateModelKeepers(t *testing.T) {
	assert := assert.New(t)

	// SyntacticVerify
	tx := &TxUpdateModelKeepers{}
	assert.ErrorContains(tx.SyntacticVerify(), "nil pointer")
	_, err := tx.MarshalJSON()
	assert.NoError(err)

	bctx := NewMockBCtx()
	bs := NewMockBS(bctx)
	token := ld.MustNewToken("$LDC")

	from, err := bs.LoadAccount(util.Signer1.Address())
	assert.NoError(err)
	approver := util.Signer2.Address()

	txData := &ld.TxData{
		Type:      ld.TypeUpdateModelKeepers,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      from.id,
	}
	assert.NoError(txData.SyntacticVerify())
	_, err = NewTx(txData.ToTransaction(), true)
	assert.ErrorContains(err, "DeriveSigners: no signature")

	txData = &ld.TxData{
		Type:      ld.TypeUpdateModelKeepers,
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
		Type:      ld.TypeUpdateModelKeepers,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      from.id,
		Token:     &token,
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx(txData.ToTransaction(), true)
	assert.ErrorContains(err, "invalid token, should be nil")

	txData = &ld.TxData{
		Type:      ld.TypeUpdateModelKeepers,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      from.id,
		Amount:    big.NewInt(1),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx(txData.ToTransaction(), true)
	assert.ErrorContains(err, "TxData.SyntacticVerify failed: invalid to")

	txData = &ld.TxData{
		Type:      ld.TypeUpdateModelKeepers,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      from.id,
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx(txData.ToTransaction(), true)
	assert.ErrorContains(err, "invalid data")

	txData = &ld.TxData{
		Type:      ld.TypeUpdateModelKeepers,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      from.id,
		Data:      []byte("你好👋"),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx(txData.ToTransaction(), true)
	assert.ErrorContains(err, "cbor: cannot unmarshal")

	input := ld.TxUpdater{}
	assert.NoError(input.SyntacticVerify())
	txData = &ld.TxData{
		Type:      ld.TypeUpdateModelKeepers,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      from.id,
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx(txData.ToTransaction(), true)
	assert.ErrorContains(err, "invalid mid")

	input = ld.TxUpdater{ModelID: &util.ModelIDEmpty}
	assert.NoError(input.SyntacticVerify())
	txData = &ld.TxData{
		Type:      ld.TypeUpdateModelKeepers,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      from.id,
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx(txData.ToTransaction(), true)
	assert.ErrorContains(err, "invalid mid")

	mid := util.ModelID{'1', '2', '3', '4', '5', '6'}
	input = ld.TxUpdater{ModelID: &mid, Keepers: &util.EthIDs{}}
	assert.ErrorContains(input.SyntacticVerify(), "nil threshold")
	input = ld.TxUpdater{ModelID: &mid, Threshold: ld.Uint8Ptr(0)}
	assert.ErrorContains(input.SyntacticVerify(), "nil keepers")
	input = ld.TxUpdater{ModelID: &mid, Threshold: ld.Uint8Ptr(1), Keepers: &util.EthIDs{}}
	assert.ErrorContains(input.SyntacticVerify(), "invalid threshold, expected <= 0, got 1")

	mid = util.ModelID{'1', '2', '3', '4', '5', '6'}
	input = ld.TxUpdater{ModelID: &mid}
	assert.NoError(input.SyntacticVerify())
	txData = &ld.TxData{
		Type:      ld.TypeUpdateModelKeepers,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      from.id,
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx(txData.ToTransaction(), true)
	assert.ErrorContains(err, "nothing to update")

	mid = util.ModelID{'1', '2', '3', '4', '5', '6'}
	input = ld.TxUpdater{
		ModelID:  &mid,
		Approver: &approver,
	}
	assert.NoError(input.SyntacticVerify())
	txData = &ld.TxData{
		Type:      ld.TypeUpdateModelKeepers,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      from.id,
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	tt := txData.ToTransaction()
	itx, err := NewTx(tt, true)
	assert.NoError(err)
	assert.ErrorContains(itx.Verify(bctx, bs), "invalid gas, expected 601, got 0")

	tt.Gas = tt.RequiredGas(bctx.FeeConfig().ThresholdGas)
	itx, err = NewTx(tt, true)
	assert.NoError(err)
	assert.ErrorContains(itx.Verify(bctx, bs),
		"0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC has an insufficient NativeLDC balance, expected 661100, got 0")
	from.Add(constants.NativeToken, new(big.Int).SetUint64(constants.LDC))
	assert.ErrorContains(itx.Verify(bctx, bs),
		"LM5V8FMkzy77ibQauKnRxM6aGSLG4AaYTdB not found")

	ipldm, err := service.ProfileModel()
	assert.NoError(err)
	mm := &ld.ModelMeta{
		Name:      ipldm.Name(),
		Threshold: 1,
		Keepers:   util.EthIDs{util.Signer1.Address()},
		Data:      ipldm.Schema(),
	}
	assert.NoError(mm.SyntacticVerify())
	assert.NoError(bs.SaveModel(mid, mm))
	assert.NoError(itx.Verify(bctx, bs))
	assert.NoError(itx.Accept(bctx, bs))

	tx = itx.(*TxUpdateModelKeepers)
	assert.Equal(tx.ld.Gas*bctx.Price, tx.ldc.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(tx.ld.Gas*100, tx.miner.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(constants.LDC-tx.ld.Gas*(bctx.Price+100),
		from.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(uint64(1), tx.from.Nonce())

	mm, err = bs.LoadModel(mid)
	assert.NoError(err)
	assert.NotNil(mm.Approver)
	assert.Equal(approver, *mm.Approver)

	jsondata, err := itx.MarshalJSON()
	assert.NoError(err)
	// fmt.Println(string(jsondata))
	assert.Equal(`{"type":18,"chainID":2357,"nonce":0,"gasTip":100,"gasFeeCap":1000,"from":"0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC","data":{"mid":"LM5V8FMkzy77ibQauKnRxM6aGSLG4AaYTdB","approver":"0x44171C37Ff5D7B7bb8dcad5C81f16284A229e641"},"signatures":["454b957046a4413fe6b9c7cc06f9d6b2ce77b0a0a57b236b66d966917e8e2abb6f763e9ecf9255b12e0e9a7c13f82c5004733df0ac9d38266198d465b0145fa100"],"gas":601,"name":"UpdateModelKeepersTx","id":"uMUvxUX46YdPKvuUpqVtEBTFpxQeF5fP4x9Jb9zePaNHEVRh1"}`, string(jsondata))

	assert.NoError(bs.VerifyState())

	// approver sign and clear approver
	input = ld.TxUpdater{
		ModelID:   &mid,
		Approver:  &util.EthIDEmpty,
		Threshold: ld.Uint8Ptr(1),
		Keepers:   &util.EthIDs{util.Signer1.Address(), util.Signer2.Address()},
	}
	assert.NoError(input.SyntacticVerify())
	txData = &ld.TxData{
		Type:      ld.TypeUpdateModelKeepers,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     1,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      from.id,
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	tt = txData.ToTransaction()
	tt.Gas = tt.RequiredGas(bctx.FeeConfig().ThresholdGas)
	itx, err = NewTx(tt, true)
	assert.NoError(err)
	assert.ErrorContains(itx.Verify(bctx, bs), "invalid signature for approver")

	txData = &ld.TxData{
		Type:      ld.TypeUpdateModelKeepers,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     1,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      from.id,
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	assert.NoError(txData.SignWith(util.Signer2))
	tt = txData.ToTransaction()
	tt.Gas = tt.RequiredGas(bctx.FeeConfig().ThresholdGas)
	itx, err = NewTx(tt, true)
	assert.NoError(err)
	assert.NoError(itx.Verify(bctx, bs))
	assert.NoError(itx.Accept(bctx, bs))

	mm, err = bs.LoadModel(mid)
	assert.NoError(err)
	assert.Nil(mm.Approver)
	assert.Equal(util.EthIDs{util.Signer1.Address(), util.Signer2.Address()}, mm.Keepers)

	// check SatisfySigningPlus
	input = ld.TxUpdater{
		ModelID:   &mid,
		Threshold: ld.Uint8Ptr(0),
		Keepers:   &util.EthIDs{util.Signer2.Address()},
	}
	assert.NoError(input.SyntacticVerify())
	txData = &ld.TxData{
		Type:      ld.TypeUpdateModelKeepers,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     2,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      from.id,
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	tt = txData.ToTransaction()
	tt.Gas = tt.RequiredGas(bctx.FeeConfig().ThresholdGas)
	itx, err = NewTx(tt, true)
	assert.NoError(err)
	assert.ErrorContains(itx.Verify(bctx, bs), "invalid signature for keepers")

	txData = &ld.TxData{
		Type:      ld.TypeUpdateModelKeepers,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     2,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      from.id,
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	assert.NoError(txData.SignWith(util.Signer2))
	tt = txData.ToTransaction()
	tt.Gas = tt.RequiredGas(bctx.FeeConfig().ThresholdGas)
	itx, err = NewTx(tt, true)
	assert.NoError(err)
	assert.NoError(itx.Verify(bctx, bs))
	assert.NoError(itx.Accept(bctx, bs))

	mm, err = bs.LoadModel(mid)
	assert.NoError(err)
	assert.Nil(mm.Approver)
	assert.Equal(uint8(0), mm.Threshold)
	assert.Equal(util.EthIDs{util.Signer2.Address()}, mm.Keepers)

	assert.NoError(bs.VerifyState())
}
