// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package transaction

import (
	"math/big"
	"testing"

	"github.com/ldclabs/ldvm/constants"
	"github.com/ldclabs/ldvm/ld"
	"github.com/ldclabs/ldvm/util"

	cborpatch "github.com/ldclabs/cbor-patch"
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
	bs := bctx.MockBS()
	token := ld.MustNewToken("$LDC")

	owner := util.Signer1.Address()
	modelKeeper := util.Signer2.Address()

	txData := &ld.TxData{
		Type:      ld.TypeUpdateData,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      owner,
	}
	assert.NoError(txData.SyntacticVerify())
	_, err = NewTx(txData.ToTransaction(), true)
	assert.ErrorContains(err, "DeriveSigners error: no signature")

	txData = &ld.TxData{
		Type:      ld.TypeUpdateData,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      owner,
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
		From:      owner,
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
		From:      owner,
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
		From:      owner,
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
		From:      owner,
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
		From:      owner,
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx(txData.ToTransaction(), true)
	assert.ErrorContains(err, "invalid data version")

	input = &ld.TxUpdater{ID: &did, Version: 1, Threshold: ld.Uint16Ptr(1)}
	txData = &ld.TxData{
		Type:      ld.TypeUpdateData,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      owner,
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx(txData.ToTransaction(), true)
	assert.ErrorContains(err, "nil keepers together with threshold")

	input = &ld.TxUpdater{ID: &did, Version: 1, Threshold: ld.Uint16Ptr(1),
		Keepers: &util.EthIDs{util.Signer1.Address()}}
	txData = &ld.TxData{
		Type:      ld.TypeUpdateData,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      owner,
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
		From:      owner,
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
		From:      owner,
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
		From:      owner,
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
		From:      owner,
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
		From:      owner,
		Data:      input.Bytes(),
		To:        &modelKeeper,
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
		From:      owner,
		Data:      input.Bytes(),
		Amount:    new(big.Int).SetUint64(constants.MilliLDC),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx(txData.ToTransaction(), true)
	assert.ErrorContains(err, "nil to together with amount")

	input = &ld.TxUpdater{ID: &did, Version: 1,
		Data: []byte(`421`),
		KSig: &kSig,
		To:   &modelKeeper,
	}
	txData = &ld.TxData{
		Type:      ld.TypeUpdateData,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      owner,
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx(txData.ToTransaction(), true)
	assert.ErrorContains(err,
		"invalid to, expected 0x44171C37Ff5D7B7bb8dcad5C81f16284A229e641, got <nil>")

	input = &ld.TxUpdater{ID: &did, Version: 1,
		Data: []byte(`421`),
		KSig: &kSig,
		To:   &modelKeeper,
	}
	txData = &ld.TxData{
		Type:      ld.TypeUpdateData,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      owner,
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
		To:   &modelKeeper,
	}
	txData = &ld.TxData{
		Type:      ld.TypeUpdateData,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      owner,
		To:        &modelKeeper,
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx(txData.ToTransaction(), true)
	assert.ErrorContains(err, "nil amount")

	input = &ld.TxUpdater{ID: &did, Version: 1,
		Data:   []byte(`421`),
		KSig:   &kSig,
		To:     &modelKeeper,
		Amount: new(big.Int).SetUint64(constants.MilliLDC),
	}
	txData = &ld.TxData{
		Type:      ld.TypeUpdateData,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      owner,
		To:        &modelKeeper,
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx(txData.ToTransaction(), true)
	assert.ErrorContains(err, "nil amount")

	input = &ld.TxUpdater{ID: &did, Version: 1,
		Data:   []byte(`421`),
		KSig:   &kSig,
		To:     &modelKeeper,
		Amount: new(big.Int).SetUint64(constants.MilliLDC),
	}
	txData = &ld.TxData{
		Type:      ld.TypeUpdateData,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      owner,
		To:        &modelKeeper,
		Amount:    new(big.Int).SetUint64(constants.MilliLDC),
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx(txData.ToTransaction(), true)
	assert.ErrorContains(err, "nil mSig")

	input = &ld.TxUpdater{ID: &did, Version: 1,
		Data:   []byte(`421`),
		KSig:   &kSig,
		To:     &modelKeeper,
		Amount: new(big.Int).SetUint64(constants.MilliLDC),
		MSig:   &kSig,
	}
	txData = &ld.TxData{
		Type:      ld.TypeUpdateData,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      owner,
		To:        &modelKeeper,
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
		To:     &modelKeeper,
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
		From:      owner,
		To:        &modelKeeper,
		Amount:    new(big.Int).SetUint64(constants.MilliLDC),
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	tt = txData.ToTransaction()
	tt.Timestamp = 10
	_, err = NewTx(tt, true)
	assert.ErrorContains(err,
		"invalid exSignatures, Transaction.ExSigners error: DeriveSigners error: no signature")

	assert.NoError(txData.ExSignWith(util.Signer1))
	tt = txData.ToTransaction()
	tt.Timestamp = 10
	itx, err := NewTx(tt, true)
	assert.NoError(err)
	bs.CommitAccounts()
	assert.ErrorContains(itx.Apply(bctx, bs),
		"invalid gas, expected 390, got 0")
	bs.CheckoutAccounts()

	tt.Gas = tt.RequiredGas(bctx.FeeConfig().ThresholdGas)
	itx, err = NewTx(tt, true)
	assert.NoError(err)
	bs.CommitAccounts()
	assert.ErrorContains(itx.Apply(bctx, bs),
		"insufficient NativeLDC balance, expected 1429000, got 0")
	bs.CheckoutAccounts()

	ownerAcc := bs.MustAccount(owner)
	ownerAcc.Add(constants.NativeToken, new(big.Int).SetUint64(constants.LDC))
	bs.CommitAccounts()
	assert.ErrorContains(itx.Apply(bctx, bs),
		"LD6L5yRJL2iYi9PbrhRru6uKfEAzDGHwUJ not found")
	bs.CheckoutAccounts()

	di := &ld.DataInfo{
		ModelID:   constants.RawModelID,
		Version:   2,
		Threshold: 1,
		Keepers:   util.EthIDs{util.Signer1.Address()},
		Data:      []byte(`42`),
		ID:        did,
	}
	kSig, err = util.Signer1.Sign(di.Data)
	assert.NoError(err)
	di.KSig = kSig
	assert.NoError(di.SyntacticVerify())
	bs.SaveData(di.ID, di)
	bs.CommitAccounts()
	assert.ErrorContains(itx.Apply(bctx, bs),
		"invalid version, expected 2, got 1")
	bs.CheckoutAccounts()

	di = &ld.DataInfo{
		ModelID:   constants.RawModelID,
		Version:   1,
		Threshold: 1,
		Keepers:   util.EthIDs{util.Signer1.Address()},
		Data:      []byte(`42`),
		ID:        did,
	}
	kSig, err = util.Signer1.Sign(di.Data)
	assert.NoError(err)
	di.KSig = kSig
	assert.NoError(di.SyntacticVerify())
	bs.SaveData(di.ID, di)
	bs.CommitAccounts()
	assert.ErrorContains(itx.Apply(bctx, bs),
		"invalid to, should be nil")
	bs.CheckoutAccounts()

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
		From:      owner,
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	tt = txData.ToTransaction()
	tt.Gas = tt.RequiredGas(bctx.FeeConfig().ThresholdGas)
	itx, err = NewTx(tt, true)
	assert.NoError(err)
	bs.CommitAccounts()
	assert.ErrorContains(itx.Apply(bctx, bs),
		"invalid data signature for data keepers, invalid signature")
	bs.CheckoutAccounts()

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
		From:      owner,
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	tt = txData.ToTransaction()
	tt.Gas = tt.RequiredGas(bctx.FeeConfig().ThresholdGas)
	ownerGas := tt.Gas
	itx, err = NewTx(tt, true)
	assert.NoError(err)
	assert.NoError(itx.Apply(bctx, bs))

	assert.Equal(ownerGas*bctx.Price,
		itx.(*TxUpdateData).ldc.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(ownerGas*100,
		itx.(*TxUpdateData).miner.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(constants.LDC-ownerGas*(bctx.Price+100),
		ownerAcc.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(uint64(1), ownerAcc.Nonce())

	di2, err := bs.LoadData(di.ID)
	assert.NoError(err)
	assert.Equal(uint64(2), di2.Version)
	assert.NotEqual(kSig, di.KSig)
	assert.Equal(kSig, di2.KSig)
	assert.Equal([]byte(`42`), []byte(di.Data))
	assert.Equal([]byte(`421`), []byte(di2.Data))
	assert.Equal(bs.PDC[di.ID], di.Bytes())

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
	bs := bctx.MockBS()

	owner := util.Signer1.Address()
	ownerAcc := bs.MustAccount(owner)
	assert.NoError(ownerAcc.Add(constants.NativeToken, new(big.Int).SetUint64(constants.LDC)))

	type cborData struct {
		Name   string `cbor:"na"`
		Nonces []int  `cbor:"no"`
	}

	data, err := util.MarshalCBOR(&cborData{Name: "test", Nonces: []int{1, 2, 3}})
	assert.NoError(err)

	di := &ld.DataInfo{
		ModelID:   constants.CBORModelID,
		Version:   2,
		Threshold: 1,
		Keepers:   util.EthIDs{util.Signer1.Address()},
		Data:      data,
		ID:        util.DataID{1, 2, 3, 4},
	}
	kSig, err := util.Signer1.Sign(di.Data)
	assert.NoError(err)
	di.KSig = kSig
	assert.NoError(di.SyntacticVerify())
	bs.SaveData(di.ID, di)

	input := &ld.TxUpdater{ID: &di.ID, Version: 2,
		Data: di.Data[2:],
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
		From:      owner,
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	tt := txData.ToTransaction()
	tt.Gas = tt.RequiredGas(bctx.FeeConfig().ThresholdGas)
	itx, err := NewTx(tt, true)
	assert.NoError(err)
	bs.CommitAccounts()
	assert.ErrorContains(itx.Apply(bctx, bs), "invalid CBOR patch")
	bs.CheckoutAccounts()

	patch := cborpatch.Patch{
		{Op: "add", Path: "/no/-", Value: util.MustMarshalCBOR(4)},
	}
	patchdata := util.MustMarshalCBOR(patch)
	input = &ld.TxUpdater{ID: &di.ID, Version: 2,
		Data: patchdata,
	}
	input.KSig = &kSig
	txData = &ld.TxData{
		Type:      ld.TypeUpdateData,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      owner,
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	tt = txData.ToTransaction()
	tt.Gas = tt.RequiredGas(bctx.FeeConfig().ThresholdGas)
	itx, err = NewTx(tt, true)
	assert.NoError(err)
	bs.CommitAccounts()
	assert.ErrorContains(itx.Apply(bctx, bs),
		"invalid data signature for data keepers, invalid signature")
	bs.CheckoutAccounts()

	input = &ld.TxUpdater{ID: &di.ID, Version: 2,
		Data: patchdata,
	}
	newData, err := patch.Apply(di.Data)
	assert.NoError(err)
	kSig, err = util.Signer1.Sign(newData)
	assert.NoError(err)
	input.KSig = &kSig
	txData = &ld.TxData{
		Type:      ld.TypeUpdateData,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      owner,
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	tt = txData.ToTransaction()
	tt.Gas = tt.RequiredGas(bctx.FeeConfig().ThresholdGas)
	ownerGas := tt.Gas
	itx, err = NewTx(tt, true)
	assert.NoError(err)
	assert.NoError(itx.Apply(bctx, bs))

	tx = itx.(*TxUpdateData)
	assert.Equal(ownerGas*bctx.Price,
		itx.(*TxUpdateData).ldc.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(ownerGas*100,
		itx.(*TxUpdateData).miner.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(constants.LDC-ownerGas*(bctx.Price+100),
		itx.(*TxUpdateData).from.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(uint64(1), itx.(*TxUpdateData).from.Nonce())

	di2, err := bs.LoadData(di.ID)
	assert.NoError(err)
	assert.Equal(uint64(3), di2.Version)
	assert.NotEqual(kSig, di.KSig)
	assert.Equal(kSig, di2.KSig)
	assert.NotEqual(newData, []byte(di.Data))
	assert.Equal(newData, []byte(di2.Data))
	assert.Equal(bs.PDC[di.ID], di.Bytes())

	var nc cborData
	assert.NoError(util.UnmarshalCBOR(di2.Data, &nc))
	assert.Equal([]int{1, 2, 3, 4}, nc.Nonces)

	jsondata, err := itx.MarshalJSON()
	assert.NoError(err)
	// fmt.Println(string(jsondata))
	assert.Equal(`{"type":"TypeUpdateData","chainID":2357,"nonce":0,"gasTip":100,"gasFeeCap":1000,"from":"0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC","data":{"id":"LD6L5yRJL2iYi9PbrhRru6uKfEAzDGHwUJ","version":2,"kSig":"565023ebbf2f8a00bf65056c6e4aebd63ec58c66d410b6d4ae62b6859e451b6d00de0bd83535adcaa1bb8d81bd06d0727b3cf2d3a712f0576be588151f02403600","data":"0x81a3626f70636164646470617468652f6e6f2f2d6576616c7565040f9dc5ca"},"signatures":["1a0306bf9768771d9806458fea0162940e398630b27f4eaf209d0bcd276d0a3e50aeeac718ec43da7b02ac453185d7c447167d64bb8430588f1bfed6bb25b62001"],"gas":280,"id":"LdoZA49wq9kbkxQGd1XCxcmLSp7mS7XfuMuvoJ243pVgYNY2X"}`, string(jsondata))

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
	bs := bctx.MockBS()

	owner := util.Signer1.Address()
	ownerAcc := bs.MustAccount(owner)
	assert.NoError(ownerAcc.Add(constants.NativeToken, new(big.Int).SetUint64(constants.LDC)))

	data := []byte(`{"name":"test","nonces":[1,2,3]}`)
	di := &ld.DataInfo{
		ModelID:   constants.JSONModelID,
		Version:   2,
		Threshold: 1,
		Keepers:   util.EthIDs{util.Signer1.Address()},
		Data:      data,
		ID:        util.DataID{1, 2, 3, 4},
	}
	kSig, err := util.Signer1.Sign(di.Data)
	assert.NoError(err)
	di.KSig = kSig
	assert.NoError(di.SyntacticVerify())
	bs.SaveData(di.ID, di)

	input := &ld.TxUpdater{ID: &di.ID, Version: 2,
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
		From:      owner,
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	tt := txData.ToTransaction()
	tt.Gas = tt.RequiredGas(bctx.FeeConfig().ThresholdGas)
	itx, err := NewTx(tt, true)
	assert.NoError(err)
	bs.CommitAccounts()
	assert.ErrorContains(itx.Apply(bctx, bs), "invalid JSON patch")
	bs.CheckoutAccounts()

	input = &ld.TxUpdater{ID: &di.ID, Version: 2,
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
		From:      owner,
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	tt = txData.ToTransaction()
	tt.Gas = tt.RequiredGas(bctx.FeeConfig().ThresholdGas)
	itx, err = NewTx(tt, true)
	assert.NoError(err)
	bs.CommitAccounts()
	assert.ErrorContains(itx.Apply(bctx, bs),
		"invalid data signature for data keepers, invalid signature")
	bs.CheckoutAccounts()

	input = &ld.TxUpdater{ID: &di.ID, Version: 2,
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
		From:      owner,
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	tt = txData.ToTransaction()
	tt.Gas = tt.RequiredGas(bctx.FeeConfig().ThresholdGas)
	ownerGas := tt.Gas
	itx, err = NewTx(tt, true)
	assert.NoError(err)
	assert.NoError(itx.Apply(bctx, bs))

	assert.Equal(ownerGas*bctx.Price,
		itx.(*TxUpdateData).ldc.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(ownerGas*100,
		itx.(*TxUpdateData).miner.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(constants.LDC-ownerGas*(bctx.Price+100),
		ownerAcc.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(uint64(1), ownerAcc.Nonce())

	di2, err := bs.LoadData(di.ID)
	assert.NoError(err)
	assert.Equal(uint64(3), di2.Version)
	assert.NotEqual(kSig, di.KSig)
	assert.Equal(kSig, di2.KSig)
	assert.Equal([]byte(`{"name":"test","nonces":[1,2,3]}`), []byte(di.Data))
	assert.Equal([]byte(`{"name":"Tester","nonces":[1,2,3]}`), []byte(di2.Data))
	assert.Equal(bs.PDC[di.ID], di.Bytes())

	jsondata, err := itx.MarshalJSON()
	assert.NoError(err)
	// fmt.Println(string(jsondata))
	assert.Equal(`{"type":"TypeUpdateData","chainID":2357,"nonce":0,"gasTip":100,"gasFeeCap":1000,"from":"0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC","data":{"id":"LD6L5yRJL2iYi9PbrhRru6uKfEAzDGHwUJ","version":2,"kSig":"5ea1b80028d7b80a1545ae788183bc7be028ac2fc9d0e2859203378b53e678106bdfc3fa4960277d48ed6025e2b461b10f16a6c6034a4b290aaa696facfedbc501","data":[{"op":"replace","path":"/name","value":"Tester"}]},"signatures":["8a6a5d649497eb1aec1e50026b179474a9f2f42055b3a3437aab983ac85317c15220f0423a3328b0823c88812e8fc442061ab18efc2ed5ccdbf3aa39154b02cb00"],"gas":308,"id":"GPLaz3tZacJDwa9AEYQYZyanwjb5xW23CCD5wA6fVdRzGqcX6"}`, string(jsondata))

	assert.NoError(bs.VerifyState())
}

func TestTxUpdateDataWithoutModelKeepers(t *testing.T) {
	t.Skip("IPLDModel.ApplyPatch")
}
