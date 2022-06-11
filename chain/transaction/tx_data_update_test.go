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

func TestTxUpdateData(t *testing.T) {
	assert := assert.New(t)

	// SyntacticVerify
	tx := &TxUpdateData{}
	assert.ErrorContains(tx.SyntacticVerify(), "nil pointer")
	_, err := tx.MarshalJSON()
	assert.NoError(err)

	bctx := NewMockBCtx()
	bs := NewMockBS(bctx)
	token := ld.MustNewToken("$LDC")

	from, err := bs.LoadAccount(util.Signer1.Address())
	assert.NoError(err)
	to, err := bs.LoadAccount(util.Signer2.Address())
	assert.NoError(err)

	txData := &ld.TxData{
		Type:      ld.TypeUpdateData,
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
		Type:      ld.TypeUpdateData,
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
		Type:      ld.TypeUpdateData,
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
		Type:      ld.TypeUpdateData,
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
		Type:      ld.TypeUpdateData,
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
		Type:      ld.TypeUpdateData,
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
		Type:      ld.TypeUpdateData,
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

	input = &ld.TxUpdater{ID: &did, Version: 1, Threshold: ld.Uint8Ptr(1)}
	txData = &ld.TxData{
		Type:      ld.TypeUpdateData,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      from.id,
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx(txData.ToTransaction(), true)
	assert.ErrorContains(err, "nil keepers together with threshold")

	input = &ld.TxUpdater{ID: &did, Version: 1, Threshold: ld.Uint8Ptr(1),
		Keepers: &util.EthIDs{util.Signer1.Address()}}
	txData = &ld.TxData{
		Type:      ld.TypeUpdateData,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      from.id,
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx(txData.ToTransaction(), true)
	assert.ErrorContains(err, "invalid threshold, should be nil")

	input = &ld.TxUpdater{ID: &did, Version: 1, Approver: &util.EthIDEmpty}
	txData = &ld.TxData{
		Type:      ld.TypeUpdateData,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      from.id,
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx(txData.ToTransaction(), true)
	assert.ErrorContains(err, "invalid approver, should be nil")

	input = &ld.TxUpdater{ID: &did, Version: 1, ApproveList: []ld.TxType{ld.TypeDeleteData}}
	txData = &ld.TxData{
		Type:      ld.TypeUpdateData,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      from.id,
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx(txData.ToTransaction(), true)
	assert.ErrorContains(err, "invalid approveList, should be nil")

	input = &ld.TxUpdater{ID: &did, Version: 1}
	txData = &ld.TxData{
		Type:      ld.TypeUpdateData,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      from.id,
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx(txData.ToTransaction(), true)
	assert.ErrorContains(err, "invalid data")

	input = &ld.TxUpdater{ID: &did, Version: 1,
		Data: []byte(`421`)}
	txData = &ld.TxData{
		Type:      ld.TypeUpdateData,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      from.id,
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx(txData.ToTransaction(), true)
	assert.ErrorContains(err, "nil kSig")

	kSig, err := util.Signer1.Sign(input.Data)
	assert.NoError(err)
	input = &ld.TxUpdater{ID: &did, Version: 1,
		Data: []byte(`421`),
		KSig: &kSig,
	}
	txData = &ld.TxData{
		Type:      ld.TypeUpdateData,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      from.id,
		Data:      input.Bytes(),
		To:        &to.id,
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx(txData.ToTransaction(), true)
	assert.ErrorContains(err, "invalid to, should be nil")

	txData = &ld.TxData{
		Type:      ld.TypeUpdateData,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      from.id,
		Data:      input.Bytes(),
		Amount:    new(big.Int).SetUint64(constants.MilliLDC),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx(txData.ToTransaction(), true)
	assert.ErrorContains(err, "nil to together with amount")

	input = &ld.TxUpdater{ID: &did, Version: 1,
		Data: []byte(`421`),
		KSig: &kSig,
		To:   &to.id,
	}
	txData = &ld.TxData{
		Type:      ld.TypeUpdateData,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      from.id,
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx(txData.ToTransaction(), true)
	assert.ErrorContains(err,
		"invalid to, expected 0x44171C37Ff5D7B7bb8dcad5C81f16284A229e641, got <nil>")

	input = &ld.TxUpdater{ID: &did, Version: 1,
		Data: []byte(`421`),
		KSig: &kSig,
		To:   &to.id,
	}
	txData = &ld.TxData{
		Type:      ld.TypeUpdateData,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      from.id,
		To:        &constants.GenesisAccount,
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx(txData.ToTransaction(), true)
	assert.ErrorContains(err,
		"invalid to, expected 0x44171C37Ff5D7B7bb8dcad5C81f16284A229e641, got 0xFFfFfFffFFfffFFfFFfFFFFFffFFFffffFfFFFfF")

	input = &ld.TxUpdater{ID: &did, Version: 1,
		Data: []byte(`421`),
		KSig: &kSig,
		To:   &to.id,
	}
	txData = &ld.TxData{
		Type:      ld.TypeUpdateData,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      from.id,
		To:        &to.id,
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx(txData.ToTransaction(), true)
	assert.ErrorContains(err, "nil amount")

	input = &ld.TxUpdater{ID: &did, Version: 1,
		Data:   []byte(`421`),
		KSig:   &kSig,
		To:     &to.id,
		Amount: new(big.Int).SetUint64(constants.MilliLDC),
	}
	txData = &ld.TxData{
		Type:      ld.TypeUpdateData,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      from.id,
		To:        &to.id,
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx(txData.ToTransaction(), true)
	assert.ErrorContains(err, "nil amount")

	input = &ld.TxUpdater{ID: &did, Version: 1,
		Data:   []byte(`421`),
		KSig:   &kSig,
		To:     &to.id,
		Amount: new(big.Int).SetUint64(constants.MilliLDC),
	}
	txData = &ld.TxData{
		Type:      ld.TypeUpdateData,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      from.id,
		To:        &to.id,
		Amount:    new(big.Int).SetUint64(constants.MilliLDC),
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx(txData.ToTransaction(), true)
	assert.ErrorContains(err, "nil mSig")

	input = &ld.TxUpdater{ID: &did, Version: 1,
		Data:   []byte(`421`),
		KSig:   &kSig,
		To:     &to.id,
		Amount: new(big.Int).SetUint64(constants.MilliLDC),
		MSig:   &kSig,
	}
	txData = &ld.TxData{
		Type:      ld.TypeUpdateData,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      from.id,
		To:        &to.id,
		Amount:    new(big.Int).SetUint64(constants.MilliLDC),
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	tt := txData.ToTransaction()
	tt.Timestamp = 10
	_, err = NewTx(tt, true)
	assert.ErrorContains(err, "data expired")

	input = &ld.TxUpdater{ID: &did, Version: 1,
		Data:   []byte(`421`),
		KSig:   &kSig,
		To:     &to.id,
		Amount: new(big.Int).SetUint64(constants.MilliLDC),
		MSig:   &kSig,
		Expire: 10,
	}
	txData = &ld.TxData{
		Type:      ld.TypeUpdateData,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      from.id,
		To:        &to.id,
		Amount:    new(big.Int).SetUint64(constants.MilliLDC),
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	tt = txData.ToTransaction()
	tt.Timestamp = 10
	_, err = NewTx(tt, true)
	assert.ErrorContains(err, "invalid exSignatures: DeriveSigners: no signature")

	assert.NoError(txData.ExSignWith(util.Signer1))
	tt = txData.ToTransaction()
	tt.Timestamp = 10
	itx, err := NewTx(tt, true)
	assert.NoError(err)
	assert.ErrorContains(itx.Verify(bctx, bs),
		"invalid gas, expected 390, got 0")

	tt.Gas = tt.RequiredGas(bctx.FeeConfig().ThresholdGas)
	itx, err = NewTx(tt, true)
	assert.NoError(err)
	assert.ErrorContains(itx.Verify(bctx, bs),
		"insufficient NativeLDC balance, expected 1429000, got 0")

	from.Add(constants.NativeToken, new(big.Int).SetUint64(constants.LDC))
	assert.ErrorContains(itx.Verify(bctx, bs),
		"LD6L5yRJL2iYi9PbrhRru6uKfEAzDGHwUJ not found")

	dm := &ld.DataMeta{
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
	bs.SaveData(dm.ID, dm)
	assert.ErrorContains(itx.Verify(bctx, bs),
		"invalid version, expected 2, got 1")

	dm = &ld.DataMeta{
		ModelID:   constants.RawModelID,
		Version:   1,
		Threshold: 1,
		Keepers:   util.EthIDs{util.Signer1.Address()},
		Data:      []byte(`42`),
		ID:        did,
	}
	kSig, err = util.Signer1.Sign(dm.Data)
	assert.NoError(err)
	dm.KSig = kSig
	assert.NoError(dm.SyntacticVerify())
	bs.SaveData(dm.ID, dm)
	assert.ErrorContains(itx.Verify(bctx, bs),
		"invalid to, should be nil")

	input = &ld.TxUpdater{ID: &did, Version: 1,
		Data: []byte(`421`),
	}
	kSig, err = util.Signer2.Sign(input.Data)
	assert.NoError(err)
	input.KSig = &kSig
	txData = &ld.TxData{
		Type:      ld.TypeUpdateData,
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
	assert.ErrorContains(itx.Verify(bctx, bs),
		"invalid data signature for data keepers, invalid signer")

	input = &ld.TxUpdater{ID: &did, Version: 1,
		Data: []byte(`421`),
	}
	kSig, err = util.Signer1.Sign(input.Data)
	assert.NoError(err)
	input.KSig = &kSig
	txData = &ld.TxData{
		Type:      ld.TypeUpdateData,
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
	assert.NoError(itx.Verify(bctx, bs))
	assert.NoError(itx.Accept(bctx, bs))

	tx = itx.(*TxUpdateData)
	assert.Equal(tx.ld.Gas*bctx.Price, tx.ldc.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(tx.ld.Gas*100, tx.miner.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(constants.LDC-tx.ld.Gas*(bctx.Price+100),
		from.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(uint64(1), from.Nonce())

	dm2, err := bs.LoadData(dm.ID)
	assert.NoError(err)
	assert.Equal(uint64(2), dm2.Version)
	assert.NotEqual(kSig, dm.KSig)
	assert.Equal(kSig, dm2.KSig)
	assert.Equal([]byte(`42`), []byte(dm.Data))
	assert.Equal([]byte(`421`), []byte(dm2.Data))
	assert.Equal(bs.PDC[dm.ID], dm.Bytes())

	jsondata, err := itx.MarshalJSON()
	assert.NoError(err)
	// fmt.Println(string(jsondata))
	assert.Equal(`{"type":"TypeUpdateData","chainID":2357,"nonce":0,"gasTip":100,"gasFeeCap":1000,"from":"0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC","data":{"id":"LD6L5yRJL2iYi9PbrhRru6uKfEAzDGHwUJ","version":1,"kSig":"a7ea3017929b5c73b2a3d83f01514c1cbd311b799e2baa25aedb32bb6d8843eb51b30264f29ea32f37676e75de95664c1a64a3ed719259409395d8b09e14fe1100","data":421},"signatures":["2af27438a4d317195e2237dcddf6da4e94371195a3c7ed58934b69318a0d5d607c73692fd341ca8f56d3b7477fddaf957875b19ab1acba3c2ebba25db579bad301"],"gas":255,"id":"2aabVr5bHFzNHzSjC4VSxNN1MeZZUxf3NzJKucBReRxP8r73hF"}`, string(jsondata))

	assert.NoError(bs.VerifyState())
}

func TestTxUpdateCBORData(t *testing.T) {
	assert := assert.New(t)

	// SyntacticVerify
	tx := &TxUpdateData{}
	assert.ErrorContains(tx.SyntacticVerify(), "nil pointer")
	_, err := tx.MarshalJSON()
	assert.NoError(err)

	bctx := NewMockBCtx()
	bs := NewMockBS(bctx)

	from, err := bs.LoadAccount(util.Signer1.Address())
	assert.NoError(err)
	assert.NoError(from.Add(constants.NativeToken, new(big.Int).SetUint64(constants.LDC)))

	type cborData struct {
		Name   string `cbor:"na"`
		Nonces []int  `cbor:"no"`
	}

	data, err := ld.EncMode.Marshal(&cborData{Name: "test", Nonces: []int{1, 2, 3}})
	assert.NoError(err)

	dm := &ld.DataMeta{
		ModelID:   constants.CBORModelID,
		Version:   2,
		Threshold: 1,
		Keepers:   util.EthIDs{util.Signer1.Address()},
		Data:      data,
		ID:        util.DataID{1, 2, 3, 4},
	}
	kSig, err := util.Signer1.Sign(dm.Data)
	assert.NoError(err)
	dm.KSig = kSig
	assert.NoError(dm.SyntacticVerify())
	bs.SaveData(dm.ID, dm)

	input := &ld.TxUpdater{ID: &dm.ID, Version: 2,
		Data: dm.Data[2:],
	}
	kSig, err = util.Signer1.Sign(input.Data)
	assert.NoError(err)
	input.KSig = &kSig
	txData := &ld.TxData{
		Type:      ld.TypeUpdateData,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      from.id,
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	tt := txData.ToTransaction()
	tt.Gas = tt.RequiredGas(bctx.FeeConfig().ThresholdGas)
	itx, err := NewTx(tt, true)
	assert.NoError(err)
	assert.ErrorContains(itx.Verify(bctx, bs),
		"invalid CBOR encoding data")

	data, err = ld.EncMode.Marshal(&cborData{Name: "test", Nonces: []int{1, 2, 3, 4}})
	assert.NoError(err)
	input = &ld.TxUpdater{ID: &dm.ID, Version: 2,
		Data: data,
	}
	input.KSig = &kSig
	txData = &ld.TxData{
		Type:      ld.TypeUpdateData,
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
	assert.ErrorContains(itx.Verify(bctx, bs),
		"invalid data signature for data keepers, invalid signer")

	// TODO CBOR Patch
	input = &ld.TxUpdater{ID: &dm.ID, Version: 2,
		Data: data,
	}
	kSig, err = util.Signer1.Sign(input.Data)
	assert.NoError(err)
	input.KSig = &kSig
	txData = &ld.TxData{
		Type:      ld.TypeUpdateData,
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
	assert.NoError(itx.Verify(bctx, bs))
	assert.NoError(itx.Accept(bctx, bs))

	tx = itx.(*TxUpdateData)
	assert.Equal(tx.ld.Gas*bctx.Price, tx.ldc.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(tx.ld.Gas*100, tx.miner.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(constants.LDC-tx.ld.Gas*(bctx.Price+100),
		from.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(uint64(1), from.Nonce())

	dm2, err := bs.LoadData(dm.ID)
	assert.NoError(err)
	assert.Equal(uint64(3), dm2.Version)
	assert.NotEqual(kSig, dm.KSig)
	assert.Equal(kSig, dm2.KSig)
	assert.NotEqual(data, []byte(dm.Data))
	assert.Equal(data, []byte(dm2.Data))
	assert.Equal(bs.PDC[dm.ID], dm.Bytes())

	jsondata, err := itx.MarshalJSON()
	assert.NoError(err)
	// fmt.Println(string(jsondata))
	assert.Equal(`{"type":"TypeUpdateData","chainID":2357,"nonce":0,"gasTip":100,"gasFeeCap":1000,"from":"0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC","data":{"id":"LD6L5yRJL2iYi9PbrhRru6uKfEAzDGHwUJ","version":2,"kSig":"565023ebbf2f8a00bf65056c6e4aebd63ec58c66d410b6d4ae62b6859e451b6d00de0bd83535adcaa1bb8d81bd06d0727b3cf2d3a712f0576be588151f02403600","data":"0xa2626e616474657374626e6f8401020304b7efd5cd"},"signatures":["0621bde177bc2bb5d0a53d2cf8f71d18e908d30bd542cfa464c7a1a46e2667cf06c1c2e7ecce4fbdd7efa1969dad9d45c64fc27374b85538a628bc4cea96632500"],"gas":269,"id":"pdJkLeuZSK6xVqyghyV9PTmv3cyY2W4KjxKAzTCLXSxDsm7u9"}`, string(jsondata))

	assert.NoError(bs.VerifyState())
}

func TestTxUpdateJSONData(t *testing.T) {
	assert := assert.New(t)

	// SyntacticVerify
	tx := &TxUpdateData{}
	assert.ErrorContains(tx.SyntacticVerify(), "nil pointer")
	_, err := tx.MarshalJSON()
	assert.NoError(err)

	bctx := NewMockBCtx()
	bs := NewMockBS(bctx)

	from, err := bs.LoadAccount(util.Signer1.Address())
	assert.NoError(err)
	assert.NoError(from.Add(constants.NativeToken, new(big.Int).SetUint64(constants.LDC)))

	data := []byte(`{"name":"test","nonces":[1,2,3]}`)
	dm := &ld.DataMeta{
		ModelID:   constants.JSONModelID,
		Version:   2,
		Threshold: 1,
		Keepers:   util.EthIDs{util.Signer1.Address()},
		Data:      data,
		ID:        util.DataID{1, 2, 3, 4},
	}
	kSig, err := util.Signer1.Sign(dm.Data)
	assert.NoError(err)
	dm.KSig = kSig
	assert.NoError(dm.SyntacticVerify())
	bs.SaveData(dm.ID, dm)

	input := &ld.TxUpdater{ID: &dm.ID, Version: 2,
		Data: []byte(`{}`),
	}
	kSig, err = util.Signer1.Sign(input.Data)
	assert.NoError(err)
	input.KSig = &kSig
	txData := &ld.TxData{
		Type:      ld.TypeUpdateData,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      from.id,
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	tt := txData.ToTransaction()
	tt.Gas = tt.RequiredGas(bctx.FeeConfig().ThresholdGas)
	itx, err := NewTx(tt, true)
	assert.NoError(err)
	assert.ErrorContains(itx.Verify(bctx, bs),
		"invalid JSON patch")

	input = &ld.TxUpdater{ID: &dm.ID, Version: 2,
		Data: []byte(`[{"op": "replace", "path": "/name", "value": "Tester"}]`),
	}
	kSig, err = util.Signer1.Sign(input.Data)
	assert.NoError(err)
	input.KSig = &kSig
	txData = &ld.TxData{
		Type:      ld.TypeUpdateData,
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
	assert.ErrorContains(itx.Verify(bctx, bs),
		"invalid data signature for data keepers, invalid signer")

	input = &ld.TxUpdater{ID: &dm.ID, Version: 2,
		Data: []byte(`[{"op": "replace", "path": "/name", "value": "Tester"}]`),
	}
	kSig, err = util.Signer1.Sign([]byte(`{"name":"Tester","nonces":[1,2,3]}`))
	assert.NoError(err)
	input.KSig = &kSig
	txData = &ld.TxData{
		Type:      ld.TypeUpdateData,
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
	assert.NoError(itx.Verify(bctx, bs))
	assert.NoError(itx.Accept(bctx, bs))

	tx = itx.(*TxUpdateData)
	assert.Equal(tx.ld.Gas*bctx.Price, tx.ldc.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(tx.ld.Gas*100, tx.miner.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(constants.LDC-tx.ld.Gas*(bctx.Price+100),
		from.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(uint64(1), from.Nonce())

	dm2, err := bs.LoadData(dm.ID)
	assert.NoError(err)
	assert.Equal(uint64(3), dm2.Version)
	assert.NotEqual(kSig, dm.KSig)
	assert.Equal(kSig, dm2.KSig)
	assert.Equal([]byte(`{"name":"test","nonces":[1,2,3]}`), []byte(dm.Data))
	assert.Equal([]byte(`{"name":"Tester","nonces":[1,2,3]}`), []byte(dm2.Data))
	assert.Equal(bs.PDC[dm.ID], dm.Bytes())

	jsondata, err := itx.MarshalJSON()
	assert.NoError(err)
	// fmt.Println(string(jsondata))
	assert.Equal(`{"type":"TypeUpdateData","chainID":2357,"nonce":0,"gasTip":100,"gasFeeCap":1000,"from":"0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC","data":{"id":"LD6L5yRJL2iYi9PbrhRru6uKfEAzDGHwUJ","version":2,"kSig":"5ea1b80028d7b80a1545ae788183bc7be028ac2fc9d0e2859203378b53e678106bdfc3fa4960277d48ed6025e2b461b10f16a6c6034a4b290aaa696facfedbc501","data":[{"op":"replace","path":"/name","value":"Tester"}]},"signatures":["8a6a5d649497eb1aec1e50026b179474a9f2f42055b3a3437aab983ac85317c15220f0423a3328b0823c88812e8fc442061ab18efc2ed5ccdbf3aa39154b02cb00"],"gas":308,"id":"GPLaz3tZacJDwa9AEYQYZyanwjb5xW23CCD5wA6fVdRzGqcX6"}`, string(jsondata))

	assert.NoError(bs.VerifyState())
}

func TestTxUpdateDataWithoutModelKeepers(t *testing.T) {
	t.Skip("IPLDModel.ApplyPatch")
}
