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

func TestTxDeleteData(t *testing.T) {
	assert := assert.New(t)

	// SyntacticVerify
	tx := &TxDeleteData{}
	assert.ErrorContains(tx.SyntacticVerify(), "nil pointer")
	_, err := tx.MarshalJSON()
	assert.NoError(err)

	bctx := NewMockBCtx()
	bs := NewMockBS(bctx)

	from, err := bs.LoadAccount(util.Signer1.Address())
	assert.NoError(err)
	approver := util.Signer2.Address()

	txData := &ld.TxData{
		Type:      ld.TypeDeleteData,
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
		Type:      ld.TypeDeleteData,
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
		Type:      ld.TypeDeleteData,
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
		Type:      ld.TypeDeleteData,
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
		Type:      ld.TypeDeleteData,
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
		Type:      ld.TypeDeleteData,
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

	input := &ld.TxUpdater{}
	txData = &ld.TxData{
		Type:      ld.TypeDeleteData,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      from.id,
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx(txData.ToTransaction(), true)
	assert.ErrorContains(err, "invalid data id")

	input = &ld.TxUpdater{ID: &util.DataIDEmpty}
	txData = &ld.TxData{
		Type:      ld.TypeDeleteData,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      from.id,
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx(txData.ToTransaction(), true)
	assert.ErrorContains(err, "invalid data id")

	did := util.DataID{1, 2, 3, 4}
	input = &ld.TxUpdater{ID: &did}
	txData = &ld.TxData{
		Type:      ld.TypeDeleteData,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      from.id,
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx(txData.ToTransaction(), true)
	assert.ErrorContains(err, "invalid data version")

	input = &ld.TxUpdater{ID: &did, Version: 1}
	txData = &ld.TxData{
		Type:      ld.TypeDeleteData,
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
	assert.ErrorContains(itx.Verify(bctx, bs), "invalid gas, expected 279, got 0")

	tt.Gas = tt.RequiredGas(bctx.FeeConfig().ThresholdGas)
	from.Add(constants.NativeToken, new(big.Int).SetUint64(constants.LDC))
	itx, err = NewTx(tt, true)
	assert.NoError(err)
	assert.ErrorContains(itx.Verify(bctx, bs), "LD6L5yRJL2iYi9PbrhRru6uKfEAzDGHwUJ not found")

	dm := &ld.DataInfo{
		ModelID:   constants.RawModelID,
		Version:   2,
		Threshold: 1,
		Keepers:   util.EthIDs{util.Signer2.Address()},
		Approver:  &approver,
		Data:      []byte(`42`),
		ID:        did,
	}
	kSig, err := util.Signer2.Sign(dm.Data)
	assert.NoError(err)
	dm.KSig = kSig
	assert.NoError(dm.SyntacticVerify())
	assert.NoError(bs.SaveData(dm.ID, dm))
	assert.ErrorContains(itx.Verify(bctx, bs), "invalid version, expected 2, got 1")

	input = &ld.TxUpdater{ID: &did, Version: 2}
	txData = &ld.TxData{
		Type:      ld.TypeDeleteData,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     0,
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
	assert.ErrorContains(itx.Verify(bctx, bs), "invalid signatures for data keepers")

	dm = &ld.DataInfo{
		ModelID:   constants.RawModelID,
		Version:   2,
		Threshold: 1,
		Keepers:   util.EthIDs{util.Signer1.Address()},
		Approver:  &approver,
		Data:      []byte(`42`),
		ID:        did,
	}
	kSig, err = util.Signer1.Sign(dm.Data)
	assert.NoError(err)
	dm.KSig = kSig
	assert.NoError(dm.SyntacticVerify())
	assert.NoError(bs.SaveData(dm.ID, dm))
	assert.ErrorContains(itx.Verify(bctx, bs), "invalid signature for data approver")

	dm = &ld.DataInfo{
		ModelID:   constants.RawModelID,
		Version:   2,
		Threshold: 1,
		Keepers:   util.EthIDs{util.Signer1.Address()},
		Data:      []byte(`42`),
		ID:        did,
	}
	kSig, err = util.Signer1.Sign(dm.Data)
	assert.NoError(err)
	dm.KSig = kSig
	assert.NoError(dm.SyntacticVerify())
	assert.NoError(bs.SaveData(dm.ID, dm))
	assert.NoError(itx.Verify(bctx, bs))
	assert.NoError(itx.Accept(bctx, bs))

	tx = itx.(*TxDeleteData)
	assert.Equal(tx.ld.Gas*bctx.Price, tx.ldc.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(tx.ld.Gas*100, tx.miner.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(constants.LDC-tx.ld.Gas*(bctx.Price+100),
		from.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(uint64(1), tx.from.Nonce())

	dm2, err := bs.LoadData(dm.ID)
	assert.NoError(err)
	assert.Equal(uint64(0), dm2.Version)
	assert.Equal(util.SignatureEmpty, dm2.KSig)

	jsondata, err := itx.MarshalJSON()
	assert.NoError(err)
	// fmt.Println(string(jsondata))
	assert.Equal(`{"type":"TypeDeleteData","chainID":2357,"nonce":0,"gasTip":100,"gasFeeCap":1000,"from":"0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC","data":{"id":"LD6L5yRJL2iYi9PbrhRru6uKfEAzDGHwUJ","version":2},"signatures":["412d5a180fee8b76b36f0811584338dd0d084d8840e52b29988bc1fd2c00d37d01b3df69af6a524cebbe5f015f9c200c79d50d58d5ab6bf7fb7ce2b27f94c07300"],"gas":279,"id":"3xLkdg8i7DbhUR8u4aAr1BKEuYjDYiJ4HT3QYH1rwyK6acPA2"}`, string(jsondata))

	input = &ld.TxUpdater{ID: &did, Version: 2}
	txData = &ld.TxData{
		Type:      ld.TypeDeleteData,
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
	assert.ErrorContains(itx.Verify(bctx, bs), "invalid version, expected 0, got 2")

	dm = &ld.DataInfo{
		ModelID:   constants.RawModelID,
		Version:   2,
		Threshold: 1,
		Keepers:   util.EthIDs{util.Signer1.Address()},
		Data:      []byte(`42`),
		ID:        did,
	}
	kSig, err = util.Signer1.Sign(dm.Data)
	assert.NoError(err)
	dm.KSig = kSig
	assert.NoError(dm.SyntacticVerify())
	assert.NoError(bs.SaveData(dm.ID, dm))

	input = &ld.TxUpdater{ID: &did, Version: 2, Data: []byte(`421`)}
	txData = &ld.TxData{
		Type:      ld.TypeDeleteData,
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
	assert.NoError(itx.Verify(bctx, bs))
	assert.NoError(itx.Accept(bctx, bs))

	dm2, err = bs.LoadData(dm.ID)
	assert.NoError(err)
	assert.Equal(uint64(0), dm2.Version)
	assert.Equal(util.SignatureEmpty, dm2.KSig)
	assert.Equal([]byte(`421`), []byte(dm2.Data))

	assert.NoError(bs.VerifyState())
}
