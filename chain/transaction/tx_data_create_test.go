// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package transaction

import (
	"encoding/json"
	"math/big"
	"testing"

	jsonpatch "github.com/evanphx/json-patch/v5"
	"github.com/ldclabs/ldvm/constants"
	"github.com/ldclabs/ldvm/ld"
	"github.com/ldclabs/ldvm/ld/service"
	"github.com/ldclabs/ldvm/util"
	"github.com/stretchr/testify/assert"
)

func TestTxCreateData(t *testing.T) {
	assert := assert.New(t)

	// SyntacticVerify
	tx := &TxCreateData{}
	assert.ErrorContains(tx.SyntacticVerify(), "nil pointer")
	_, err := tx.MarshalJSON()
	assert.NoError(err)

	bctx := NewMockBCtx()
	bs := NewMockBS(bctx)
	token := ld.MustNewToken("$LDC")

	from, err := bs.LoadAccount(util.Signer1.Address())
	assert.NoError(err)

	txData := &ld.TxData{
		Type:      ld.TypeCreateData,
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
		Type:      ld.TypeCreateData,
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
		Type:      ld.TypeCreateData,
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
		Type:      ld.TypeCreateData,
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
		Type:      ld.TypeCreateData,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      from.id,
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx(txData.ToTransaction(), true)
	assert.ErrorContains(err, "nil mid")

	input = &ld.TxUpdater{
		ModelID: &constants.RawModelID,
	}
	txData = &ld.TxData{
		Type:      ld.TypeCreateData,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      from.id,
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx(txData.ToTransaction(), true)
	assert.ErrorContains(err, "invalid version, expected 1, got 0")

	input = &ld.TxUpdater{
		ModelID: &constants.RawModelID,
		Version: 1,
	}
	txData = &ld.TxData{
		Type:      ld.TypeCreateData,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      from.id,
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx(txData.ToTransaction(), true)
	assert.ErrorContains(err, "nil threshold")

	input = &ld.TxUpdater{
		ModelID:   &constants.RawModelID,
		Version:   1,
		Threshold: ld.Uint16Ptr(0),
	}
	txData = &ld.TxData{
		Type:      ld.TypeCreateData,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      from.id,
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx(txData.ToTransaction(), true)
	assert.ErrorContains(err, "nil keepers")

	input = &ld.TxUpdater{
		ModelID:   &constants.RawModelID,
		Version:   1,
		Threshold: ld.Uint16Ptr(0),
		Keepers:   &util.EthIDs{},
	}
	txData = &ld.TxData{
		Type:      ld.TypeCreateData,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      from.id,
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx(txData.ToTransaction(), true)
	assert.ErrorContains(err, "empty keepers")

	input = &ld.TxUpdater{
		ModelID:   &constants.RawModelID,
		Version:   1,
		Threshold: ld.Uint16Ptr(0),
		Keepers:   &util.EthIDs{util.Signer1.Address()},
	}
	txData = &ld.TxData{
		Type:      ld.TypeCreateData,
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

	input = &ld.TxUpdater{
		ModelID:   &constants.RawModelID,
		Version:   1,
		Threshold: ld.Uint16Ptr(0),
		Keepers:   &util.EthIDs{util.Signer1.Address()},
		Data:      []byte(`{}`),
	}
	txData = &ld.TxData{
		Type:      ld.TypeCreateData,
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

	input.KSig = &util.Signature{1, 2, 3}
	assert.NoError(input.SyntacticVerify())
	txData = &ld.TxData{
		Type:      ld.TypeCreateData,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      from.id,
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx(txData.ToTransaction(), true)
	assert.ErrorContains(err, "invalid kSig: DeriveSigner: recovery failed")

	kSig, err := util.Signer2.Sign(input.Data)
	assert.NoError(err)
	input.KSig = &kSig
	assert.NoError(input.SyntacticVerify())
	txData = &ld.TxData{
		Type:      ld.TypeCreateData,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      from.id,
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx(txData.ToTransaction(), true)
	assert.ErrorContains(err, "invalid kSig for keepers")

	// RawModel
	input = &ld.TxUpdater{
		ModelID:   &constants.RawModelID,
		Version:   1,
		Threshold: ld.Uint16Ptr(0),
		Keepers:   &util.EthIDs{util.Signer1.Address()},
		Data:      []byte(`42`),
	}
	kSig, err = util.Signer1.Sign(input.Data)
	assert.NoError(err)
	input.KSig = &kSig
	assert.NoError(input.SyntacticVerify())
	txData = &ld.TxData{
		Type:      ld.TypeCreateData,
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
	assert.ErrorContains(err, "invalid to, should be nil")

	input = &ld.TxUpdater{
		ModelID:   &constants.RawModelID,
		Version:   1,
		Threshold: ld.Uint16Ptr(0),
		Keepers:   &util.EthIDs{util.Signer1.Address()},
		Data:      []byte(`42`),
	}
	kSig, err = util.Signer1.Sign(input.Data)
	assert.NoError(err)
	input.KSig = &kSig
	assert.NoError(input.SyntacticVerify())
	txData = &ld.TxData{
		Type:      ld.TypeCreateData,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      from.id,
		Amount:    big.NewInt(1),
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx(txData.ToTransaction(), true)
	assert.ErrorContains(err, "nil to together with amount")

	input = &ld.TxUpdater{
		ModelID:   &constants.RawModelID,
		Version:   1,
		Threshold: ld.Uint16Ptr(0),
		Keepers:   &util.EthIDs{util.Signer1.Address()},
		Data:      []byte(`42`),
	}
	kSig, err = util.Signer1.Sign(input.Data)
	assert.NoError(err)
	input.KSig = &kSig
	assert.NoError(input.SyntacticVerify())
	txData = &ld.TxData{
		Type:      ld.TypeCreateData,
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
	assert.ErrorContains(itx.Verify(bctx, bs),
		"TxBase.Verify failed: invalid gas, expected 284, got 0")

	tt.Gas = tt.RequiredGas(bctx.FeeConfig().ThresholdGas)
	itx, err = NewTx(tt, true)
	assert.NoError(err)
	assert.ErrorContains(itx.Verify(bctx, bs),
		"insufficient NativeLDC balance, expected 312400, got 0")
	from.Add(constants.NativeToken, new(big.Int).SetUint64(constants.LDC))
	assert.NoError(itx.Verify(bctx, bs))
	assert.NoError(itx.Accept(bctx, bs))

	tx = itx.(*TxCreateData)
	assert.Equal(tx.ld.Gas*bctx.Price, tx.ldc.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(tx.ld.Gas*100, tx.miner.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(constants.LDC-tx.ld.Gas*(bctx.Price+100),
		from.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(uint64(1), tx.from.Nonce())

	dm, err := bs.LoadData(tx.dm.ID)
	assert.NoError(err)
	assert.Equal(constants.RawModelID, dm.ModelID)
	assert.Equal(uint64(1), dm.Version)
	assert.Equal(uint16(0), dm.Threshold)
	assert.Equal(util.EthIDs{from.id}, dm.Keepers)
	assert.Nil(dm.Approver)
	assert.Nil(dm.ApproveList)
	assert.Equal([]byte(`42`), []byte(dm.Data))
	assert.Equal(kSig, dm.KSig)
	assert.Nil(dm.MSig)

	jsondata, err := itx.MarshalJSON()
	assert.NoError(err)
	// fmt.Println(string(jsondata))
	assert.Equal(`{"type":"TypeCreateData","chainID":2357,"nonce":0,"gasTip":100,"gasFeeCap":1000,"from":"0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC","data":{"mid":"LM111111111111111111116DBWJs","version":1,"threshold":0,"keepers":["0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC"],"kSig":"505a3dfb3372ef790ba8237ab40a53f8e626b56b3778f9edcb67436ea1ac9fd65a7a10f80921aa34809a056c18f8cd9f905367c65b30734e137428554e71735001","data":42},"signatures":["63458545b18a568b067712a8b725030b9a3e2df2196a0d0bbee2c8f8b808239967fc611d8251f3b9f4552d22e9b9f1b7fbb2066a38160b2d1762d8875a3cf30701"],"gas":284,"id":"zmZum3cGtTFKYqE4pKtmckqB5YT6FB1upTaruj6d1WLbNYNmG"}`, string(jsondata))

	assert.NoError(bs.VerifyState())
}

func TestTxCreateCBORData(t *testing.T) {
	assert := assert.New(t)

	bctx := NewMockBCtx()
	bs := NewMockBS(bctx)

	from, err := bs.LoadAccount(util.Signer1.Address())
	assert.NoError(err)

	type cborData struct {
		Name   string `cbor:"na"`
		Nonces []int  `cbor:"no"`
	}

	data, err := ld.EncMode.Marshal(&cborData{Name: "test", Nonces: []int{1, 2, 3}})
	assert.NoError(err)
	invalidData := data[:len(data)-3]

	// CBORModel
	input := &ld.TxUpdater{
		ModelID:   &constants.CBORModelID,
		Version:   1,
		Threshold: ld.Uint16Ptr(0),
		Keepers:   &util.EthIDs{util.Signer1.Address()},
		Data:      invalidData,
	}
	kSig, err := util.Signer1.Sign(input.Data)
	assert.NoError(err)
	input.KSig = &kSig
	assert.NoError(input.SyntacticVerify())
	txData := &ld.TxData{
		Type:      ld.TypeCreateData,
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
	assert.ErrorContains(itx.Verify(bctx, bs),
		"TxBase.Verify failed: invalid gas, expected 295, got 0")

	tt.Gas = tt.RequiredGas(bctx.FeeConfig().ThresholdGas)
	itx, err = NewTx(tt, true)
	assert.NoError(err)
	assert.ErrorContains(itx.Verify(bctx, bs),
		"insufficient NativeLDC balance, expected 324500, got 0")
	from.Add(constants.NativeToken, new(big.Int).SetUint64(constants.LDC))
	assert.ErrorContains(itx.Verify(bctx, bs), "invalid CBOR encoding data")

	input = &ld.TxUpdater{
		ModelID:   &constants.CBORModelID,
		Version:   1,
		Threshold: ld.Uint16Ptr(0),
		Keepers:   &util.EthIDs{util.Signer1.Address()},
		Data:      data,
	}
	kSig, err = util.Signer1.Sign(input.Data)
	assert.NoError(err)
	input.KSig = &kSig
	assert.NoError(input.SyntacticVerify())
	txData = &ld.TxData{
		Type:      ld.TypeCreateData,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      from.id,
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	tt = txData.ToTransaction()
	itx, err = NewTx(tt, true)
	assert.NoError(err)
	assert.ErrorContains(itx.Verify(bctx, bs),
		"TxBase.Verify failed: invalid gas, expected 298, got 0")

	tt.Gas = tt.RequiredGas(bctx.FeeConfig().ThresholdGas)
	itx, err = NewTx(tt, true)
	assert.NoError(err)
	assert.NoError(itx.Verify(bctx, bs))
	assert.NoError(itx.Accept(bctx, bs))

	tx := itx.(*TxCreateData)
	assert.Equal(tx.ld.Gas*bctx.Price, tx.ldc.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(tx.ld.Gas*100, tx.miner.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(constants.LDC-tx.ld.Gas*(bctx.Price+100),
		from.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(uint64(1), tx.from.Nonce())

	dm, err := bs.LoadData(tx.dm.ID)
	assert.NoError(err)
	assert.Equal(constants.CBORModelID, dm.ModelID)
	assert.Equal(uint64(1), dm.Version)
	assert.Equal(uint16(0), dm.Threshold)
	assert.Equal(util.EthIDs{from.id}, dm.Keepers)
	assert.Nil(dm.Approver)
	assert.Nil(dm.ApproveList)
	assert.Equal(data, []byte(dm.Data))
	assert.Equal(kSig, dm.KSig)
	assert.Nil(dm.MSig)

	jsondata, err := itx.MarshalJSON()
	assert.NoError(err)
	// fmt.Println(string(jsondata))
	assert.Equal(`{"type":"TypeCreateData","chainID":2357,"nonce":0,"gasTip":100,"gasFeeCap":1000,"from":"0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC","data":{"mid":"LM1111111111111111111Ax1asG","version":1,"threshold":0,"keepers":["0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC"],"kSig":"d83fefbe1d0306ddf62d499a14b7c70904f4f5cbe501893768f32807bad26d1a3ecee945baecabf48cded8b09c807911c531a69e8cff9b2bcf83bec35822ad9901","data":"0xa2626e616474657374626e6f830102031e0946b2"},"signatures":["db109746c95a79880dfafbab8c2c147efe0dec62eebbb849c2ed80a2151229ba172702002a85dcd8bc08e7b7fb10496ec61f779ad69d394b57e1e8df0533457c01"],"gas":298,"id":"2VSgiuLzT8BxtxWhfEF1wAUQUXvHrVXto7Rmiqn4zMoBTUCodr"}`, string(jsondata))

	assert.NoError(bs.VerifyState())
}

func TestTxCreateJSONData(t *testing.T) {
	assert := assert.New(t)

	bctx := NewMockBCtx()
	bs := NewMockBS(bctx)

	from, err := bs.LoadAccount(util.Signer1.Address())
	assert.NoError(err)

	type jsonData struct {
		Name   string `json:"na"`
		Nonces []int  `json:"no"`
	}

	data, err := json.Marshal(&jsonData{Name: "test", Nonces: []int{1, 2, 3}})
	assert.NoError(err)
	invalidData := data[:len(data)-3]

	// JSONModel
	input := &ld.TxUpdater{
		ModelID:   &constants.JSONModelID,
		Version:   1,
		Threshold: ld.Uint16Ptr(0),
		Keepers:   &util.EthIDs{util.Signer1.Address()},
		Data:      invalidData,
	}
	kSig, err := util.Signer1.Sign(input.Data)
	assert.NoError(err)
	input.KSig = &kSig
	assert.NoError(input.SyntacticVerify())
	txData := &ld.TxData{
		Type:      ld.TypeCreateData,
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
	assert.ErrorContains(itx.Verify(bctx, bs),
		"TxBase.Verify failed: invalid gas, expected 305, got 0")

	tt.Gas = tt.RequiredGas(bctx.FeeConfig().ThresholdGas)
	itx, err = NewTx(tt, true)
	assert.NoError(err)
	assert.ErrorContains(itx.Verify(bctx, bs),
		"insufficient NativeLDC balance, expected 335500, got 0")
	from.Add(constants.NativeToken, new(big.Int).SetUint64(constants.LDC))
	assert.ErrorContains(itx.Verify(bctx, bs), "invalid JSON encoding data")

	input = &ld.TxUpdater{
		ModelID:   &constants.JSONModelID,
		Version:   1,
		Threshold: ld.Uint16Ptr(0),
		Keepers:   &util.EthIDs{util.Signer1.Address()},
		Data:      data,
	}
	kSig, err = util.Signer1.Sign(input.Data)
	assert.NoError(err)
	input.KSig = &kSig
	assert.NoError(input.SyntacticVerify())
	txData = &ld.TxData{
		Type:      ld.TypeCreateData,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      from.id,
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	tt = txData.ToTransaction()
	itx, err = NewTx(tt, true)
	assert.NoError(err)
	assert.ErrorContains(itx.Verify(bctx, bs),
		"TxBase.Verify failed: invalid gas, expected 309, got 0")

	tt.Gas = tt.RequiredGas(bctx.FeeConfig().ThresholdGas)
	itx, err = NewTx(tt, true)
	assert.NoError(err)
	assert.NoError(itx.Verify(bctx, bs))
	assert.NoError(itx.Accept(bctx, bs))

	tx := itx.(*TxCreateData)
	assert.Equal(tx.ld.Gas*bctx.Price, tx.ldc.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(tx.ld.Gas*100, tx.miner.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(constants.LDC-tx.ld.Gas*(bctx.Price+100),
		from.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(uint64(1), tx.from.Nonce())

	dm, err := bs.LoadData(tx.dm.ID)
	assert.NoError(err)
	assert.Equal(constants.JSONModelID, dm.ModelID)
	assert.Equal(uint64(1), dm.Version)
	assert.Equal(uint16(0), dm.Threshold)
	assert.Equal(util.EthIDs{from.id}, dm.Keepers)
	assert.Nil(dm.Approver)
	assert.Nil(dm.ApproveList)
	assert.Equal(data, []byte(dm.Data))
	assert.Equal(kSig, dm.KSig)
	assert.Nil(dm.MSig)

	jsondata, err := itx.MarshalJSON()
	assert.NoError(err)
	// fmt.Println(string(jsondata))
	assert.Equal(`{"type":"TypeCreateData","chainID":2357,"nonce":0,"gasTip":100,"gasFeeCap":1000,"from":"0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC","data":{"mid":"LM1111111111111111111L17Xp3","version":1,"threshold":0,"keepers":["0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC"],"kSig":"0340d58ceafd4fc5b7f9b296d24bf0d410804b0aa6bf5406661d9bbcbc0a58870901775c985ec519104d133b303e24bd2fc3807a6a2113bfdc2df9c1b92935ba01","data":{"na":"test","no":[1,2,3]}},"signatures":["08119318c6a7923526d393771ce505ea175d39247f253960ff27cb0cd93315003284c68a9a55ead28a1a24919f970dd73664510a3da0544548aadc29d90feba700"],"gas":309,"id":"2mPaawYLYpZEB62sFQH914QBCrAyGjNxLEDPyhbwcFUSeDNQ3G"}`, string(jsondata))

	assert.NoError(bs.VerifyState())
}

func TestTxCreateModelDataWithoutKeepers(t *testing.T) {
	assert := assert.New(t)

	bctx := NewMockBCtx()
	bs := NewMockBS(bctx)

	from, err := bs.LoadAccount(util.Signer1.Address())
	assert.NoError(err)

	pm, err := service.ProfileModel()
	assert.NoError(err)
	ps := &ld.ModelMeta{
		Name:      pm.Name(),
		Threshold: 0,
		Keepers:   util.EthIDs{util.Signer2.Address()},
		Data:      pm.Schema(),
		ID:        util.ModelID{1, 2, 3, 4, 5},
	}

	p := &service.Profile{
		Type:       1,
		Name:       "tester",
		Follows:    []util.DataID{},
		Extensions: []*service.Extension{},
	}
	assert.NoError(p.SyntacticVerify())

	data := p.Bytes()
	input := &ld.TxUpdater{
		ModelID:   &ps.ID,
		Version:   1,
		Threshold: ld.Uint16Ptr(0),
		Keepers:   &util.EthIDs{util.Signer1.Address()},
		Data:      data[1:],
	}
	kSig, err := util.Signer1.Sign(input.Data)
	assert.NoError(err)
	input.KSig = &kSig
	assert.NoError(input.SyntacticVerify())
	txData := &ld.TxData{
		Type:      ld.TypeCreateData,
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
	assert.ErrorContains(itx.Verify(bctx, bs),
		"TxBase.Verify failed: invalid gas, expected 309, got 0")

	tt.Gas = tt.RequiredGas(bctx.FeeConfig().ThresholdGas)
	itx, err = NewTx(tt, true)
	assert.NoError(err)
	assert.ErrorContains(itx.Verify(bctx, bs),
		"insufficient NativeLDC balance, expected 339900, got 0")
	from.Add(constants.NativeToken, new(big.Int).SetUint64(constants.LDC))
	assert.ErrorContains(itx.Verify(bctx, bs), "LM6L5yRNNMubYqZoZRtmk1ykJMmZppNwb1 not found")
	assert.NoError(bs.SaveModel(ps.ID, ps))
	assert.ErrorContains(itx.Verify(bctx, bs), `IPLDModel "ProfileService" error`)

	input = &ld.TxUpdater{
		ModelID:   &ps.ID,
		Version:   1,
		Threshold: ld.Uint16Ptr(1),
		Keepers:   &util.EthIDs{util.Signer1.Address()},
		Data:      data,
	}
	kSig, err = util.Signer1.Sign(input.Data)
	assert.NoError(err)
	input.KSig = &kSig
	assert.NoError(input.SyntacticVerify())
	txData = &ld.TxData{
		Type:      ld.TypeCreateData,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      from.id,
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	tt = txData.ToTransaction()
	itx, err = NewTx(tt, true)
	assert.NoError(err)
	tt.Gas = tt.RequiredGas(bctx.FeeConfig().ThresholdGas)
	itx, err = NewTx(tt, true)
	assert.NoError(err)
	assert.NoError(itx.Verify(bctx, bs))
	assert.NoError(itx.Accept(bctx, bs))

	tx := itx.(*TxCreateData)
	assert.Equal(tx.ld.Gas*bctx.Price, tx.ldc.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(tx.ld.Gas*100, tx.miner.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(constants.LDC-tx.ld.Gas*(bctx.Price+100),
		from.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(uint64(1), tx.from.Nonce())

	dm, err := bs.LoadData(tx.dm.ID)
	assert.NoError(err)
	assert.Equal(ps.ID, dm.ModelID)
	assert.Equal(uint64(1), dm.Version)
	assert.Equal(uint16(1), dm.Threshold)
	assert.Equal(util.EthIDs{from.id}, dm.Keepers)
	assert.Nil(dm.Approver)
	assert.Nil(dm.ApproveList)
	assert.Equal(data, []byte(dm.Data))
	assert.Equal(kSig, dm.KSig)
	assert.Nil(dm.MSig)

	p2 := &service.Profile{}
	assert.NoError(p2.Unmarshal(dm.Data))
	assert.NoError(p2.SyntacticVerify())
	assert.Equal(p.Type, p2.Type)
	assert.Equal(p.Name, p2.Name)
	assert.Equal(p.Follows, p2.Follows)
	assert.Equal(p.Extensions, p2.Extensions)
	assert.Equal(p.Bytes(), p2.Bytes())

	jsondata, err := itx.MarshalJSON()
	assert.NoError(err)
	// fmt.Println(string(jsondata))
	assert.Equal(`{"type":"TypeCreateData","chainID":2357,"nonce":0,"gasTip":100,"gasFeeCap":1000,"from":"0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC","data":{"mid":"LM6L5yRNNMubYqZoZRtmk1ykJMmZppNwb1","version":1,"threshold":1,"keepers":["0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC"],"kSig":"96f2dceeb001b0f604ad5fe1e2fd5bdca907152486afd88e89c88aa93677f6761dbc9798a939467b88df1b0528ba9471f87305aad7eec1a9bf839e79fbcc0d7901","data":"0xa6616960616e66746573746572617401617560626578806266738041034f47"},"signatures":["3b05de218cf3ad4d50721650314fa2734a9050e0edb205f780dca81bcaa962fb2459c4ae8be30839a62b1bfea121419442b2da8ccc103ff4e023a457e7d1314301"],"gas":310,"id":"2VaM9qTsEVnzNaeCTtNWzXEWeuyDL1D83UxfTdhJmNKG4d2qv4"}`, string(jsondata))

	assert.NoError(bs.VerifyState())
}

func TestTxCreateModelDataWithKeepers(t *testing.T) {
	assert := assert.New(t)

	bctx := NewMockBCtx()
	bs := NewMockBS(bctx)

	from, err := bs.LoadAccount(util.Signer1.Address())
	assert.NoError(err)
	to, err := bs.LoadAccount(util.Signer2.Address())
	assert.NoError(err)

	pm, err := service.ProfileModel()
	assert.NoError(err)
	ps := &ld.ModelMeta{
		Name:      pm.Name(),
		Threshold: 1,
		Keepers:   util.EthIDs{util.Signer2.Address()},
		Data:      pm.Schema(),
		ID:        bctx.Chain().ProfileServiceID,
	}

	p := &service.Profile{
		Type:       1,
		Name:       "LDC",
		Follows:    []util.DataID{},
		Extensions: []*service.Extension{},
	}
	assert.NoError(p.SyntacticVerify())
	data := p.Bytes()
	input := &ld.TxUpdater{
		ModelID:   &ps.ID,
		Version:   1,
		Threshold: ld.Uint16Ptr(0),
		Keepers:   &util.EthIDs{util.Signer1.Address()},
		Data:      data,
	}
	kSig, err := util.Signer1.Sign(input.Data)
	assert.NoError(err)
	input.KSig = &kSig
	assert.NoError(input.SyntacticVerify())
	txData := &ld.TxData{
		Type:      ld.TypeCreateData,
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
	from.Add(constants.NativeToken, new(big.Int).SetUint64(constants.LDC))
	assert.ErrorContains(itx.Verify(bctx, bs), "LMDWuG2ggqziTRsZRvVwCf5W9Vr6j1QqWNt not found")
	assert.NoError(bs.SaveModel(ps.ID, ps))
	assert.ErrorContains(itx.Verify(bctx, bs), `invalid mSig for model keepers`)

	input = &ld.TxUpdater{
		ModelID:   &ps.ID,
		Version:   1,
		Threshold: ld.Uint16Ptr(0),
		Keepers:   &util.EthIDs{util.Signer1.Address()},
		Data:      data,
		To:        &to.id,
	}
	kSig, err = util.Signer1.Sign(input.Data)
	assert.NoError(err)
	input.KSig = &kSig
	assert.NoError(input.SyntacticVerify())
	txData = &ld.TxData{
		Type:      ld.TypeCreateData,
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
	assert.ErrorContains(err,
		"invalid to, expected 0x44171C37Ff5D7B7bb8dcad5C81f16284A229e641, got <nil>")

	txData = &ld.TxData{
		Type:      ld.TypeCreateData,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      from.id,
		To:        &constants.GenesisAccount,
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	tt = txData.ToTransaction()
	tt.Gas = tt.RequiredGas(bctx.FeeConfig().ThresholdGas)
	itx, err = NewTx(tt, true)
	assert.ErrorContains(err,
		"invalid to, expected 0x44171C37Ff5D7B7bb8dcad5C81f16284A229e641, got 0xFFfFfFffFFfffFFfFFfFFFFFffFFFffffFfFFFfF")

	txData = &ld.TxData{
		Type:      ld.TypeCreateData,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      from.id,
		To:        &to.id,
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	tt = txData.ToTransaction()
	tt.Timestamp = 10
	tt.Gas = tt.RequiredGas(bctx.FeeConfig().ThresholdGas)
	itx, err = NewTx(tt, true)
	assert.ErrorContains(err, "data expired")

	input = &ld.TxUpdater{
		ModelID:   &ps.ID,
		Version:   1,
		Threshold: ld.Uint16Ptr(0),
		Keepers:   &util.EthIDs{util.Signer1.Address()},
		Data:      data,
		To:        &to.id,
		Expire:    100,
	}
	kSig, err = util.Signer1.Sign(input.Data)
	assert.NoError(err)
	input.KSig = &kSig
	assert.NoError(input.SyntacticVerify())
	txData = &ld.TxData{
		Type:      ld.TypeCreateData,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      from.id,
		To:        &to.id,
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	tt = txData.ToTransaction()
	tt.Timestamp = 10
	tt.Gas = tt.RequiredGas(bctx.FeeConfig().ThresholdGas)
	itx, err = NewTx(tt, true)
	assert.ErrorContains(err, "nil mSig")

	mSig, err := util.Signer1.Sign(input.Data)
	assert.NoError(err)
	input.MSig = &mSig
	assert.NoError(input.SyntacticVerify())
	txData = &ld.TxData{
		Type:      ld.TypeCreateData,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      from.id,
		To:        &to.id,
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	tt = txData.ToTransaction()
	tt.Timestamp = 10
	tt.Gas = tt.RequiredGas(bctx.FeeConfig().ThresholdGas)
	itx, err = NewTx(tt, true)
	assert.ErrorContains(err, "nil amount")

	input = &ld.TxUpdater{
		ModelID:   &ps.ID,
		Version:   1,
		Threshold: ld.Uint16Ptr(0),
		Keepers:   &util.EthIDs{util.Signer1.Address()},
		Data:      data,
		To:        &to.id,
		Expire:    100,
		Amount:    new(big.Int).SetUint64(0),
	}
	kSig, err = util.Signer1.Sign(input.Data)
	assert.NoError(err)
	input.KSig = &kSig
	mSig, err = util.Signer1.Sign(input.Data)
	assert.NoError(err)
	input.MSig = &mSig
	assert.NoError(input.SyntacticVerify())
	txData = &ld.TxData{
		Type:      ld.TypeCreateData,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      from.id,
		To:        &to.id,
		Amount:    new(big.Int).SetUint64(1),
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	tt = txData.ToTransaction()
	tt.Timestamp = 10
	tt.Gas = tt.RequiredGas(bctx.FeeConfig().ThresholdGas)
	itx, err = NewTx(tt, true)
	assert.ErrorContains(err, "invalid amount, expected 0, got 1")

	input.MSig = &util.Signature{1, 2, 3}
	assert.NoError(input.SyntacticVerify())
	txData = &ld.TxData{
		Type:      ld.TypeCreateData,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      from.id,
		To:        &to.id,
		Amount:    new(big.Int).SetUint64(0),
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	tt = txData.ToTransaction()
	tt.Timestamp = 10
	tt.Gas = tt.RequiredGas(bctx.FeeConfig().ThresholdGas)
	itx, err = NewTx(tt, true)
	assert.ErrorContains(err, "invalid mSig: DeriveSigner: recovery failed")

	input.MSig = &mSig
	assert.NoError(input.SyntacticVerify())
	txData = &ld.TxData{
		Type:      ld.TypeCreateData,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      from.id,
		To:        &to.id,
		Amount:    new(big.Int).SetUint64(0),
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	tt = txData.ToTransaction()
	tt.Timestamp = 10
	tt.Gas = tt.RequiredGas(bctx.FeeConfig().ThresholdGas)
	itx, err = NewTx(tt, true)
	assert.ErrorContains(err, "invalid exSignatures: DeriveSigners: no signature")

	txData = &ld.TxData{
		Type:      ld.TypeCreateData,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      from.id,
		To:        &to.id,
		Amount:    new(big.Int).SetUint64(0),
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	assert.NoError(txData.ExSignWith(util.Signer1))
	tt = txData.ToTransaction()
	tt.Timestamp = 10
	tt.Gas = tt.RequiredGas(bctx.FeeConfig().ThresholdGas)
	itx, err = NewTx(tt, true)
	assert.NoError(err)
	assert.ErrorContains(itx.Verify(bctx, bs),
		"invalid mSig for model keepers")

	input = &ld.TxUpdater{
		ModelID:   &ps.ID,
		Version:   1,
		Threshold: ld.Uint16Ptr(1),
		Keepers:   &util.EthIDs{util.Signer1.Address()},
		Data:      data,
		To:        &to.id,
		Expire:    100,
		Amount:    new(big.Int).SetUint64(0),
	}
	kSig, err = util.Signer1.Sign(input.Data)
	assert.NoError(err)
	input.KSig = &kSig
	mSig, err = util.Signer2.Sign(input.Data)
	assert.NoError(err)
	input.MSig = &mSig
	assert.NoError(input.SyntacticVerify())
	txData = &ld.TxData{
		Type:      ld.TypeCreateData,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      from.id,
		To:        &to.id,
		Amount:    new(big.Int).SetUint64(0),
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	assert.NoError(txData.ExSignWith(util.Signer1))
	tt = txData.ToTransaction()
	tt.Timestamp = 10
	tt.Gas = tt.RequiredGas(bctx.FeeConfig().ThresholdGas)
	itx, err = NewTx(tt, true)
	assert.NoError(err)
	assert.ErrorContains(itx.Verify(bctx, bs),
		"invalid exSignatures for model keepers")

	txData = &ld.TxData{
		Type:      ld.TypeCreateData,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      from.id,
		To:        &to.id,
		Amount:    new(big.Int).SetUint64(0),
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	assert.NoError(txData.ExSignWith(util.Signer2))
	tt = txData.ToTransaction()
	tt.Timestamp = 10
	tt.Gas = tt.RequiredGas(bctx.FeeConfig().ThresholdGas)
	itx, err = NewTx(tt, true)
	assert.NoError(err)
	assert.NoError(itx.Verify(bctx, bs))
	assert.NoError(itx.Accept(bctx, bs))

	tx := itx.(*TxCreateData)
	assert.Equal(tx.ld.Gas*bctx.Price, tx.ldc.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(tx.ld.Gas*100, tx.miner.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(constants.LDC-tx.ld.Gas*(bctx.Price+100),
		from.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(uint64(1), tx.from.Nonce())

	dm, err := bs.LoadData(tx.dm.ID)
	assert.NoError(err)
	assert.Equal(ps.ID, dm.ModelID)
	assert.Equal(uint64(1), dm.Version)
	assert.Equal(uint16(1), dm.Threshold)
	assert.Equal(util.EthIDs{from.id}, dm.Keepers)
	assert.Nil(dm.Approver)
	assert.Nil(dm.ApproveList)
	assert.Equal(data, []byte(dm.Data))
	assert.Equal(kSig, dm.KSig)
	assert.Equal(mSig, *dm.MSig)

	p2 := &service.Profile{}
	assert.NoError(p2.Unmarshal(dm.Data))
	assert.NoError(p2.SyntacticVerify())
	assert.Equal(p.Type, p2.Type)
	assert.Equal(p.Name, p2.Name)
	assert.Equal(p.Follows, p2.Follows)
	assert.Equal(p.Extensions, p2.Extensions)
	assert.Equal(p.Bytes(), p2.Bytes())

	jsondata, err := itx.MarshalJSON()
	assert.NoError(err)
	// fmt.Println(string(jsondata))
	assert.Equal(`{"type":"TypeCreateData","chainID":2357,"nonce":0,"gasTip":100,"gasFeeCap":1000,"from":"0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC","to":"0x44171C37Ff5D7B7bb8dcad5C81f16284A229e641","amount":0,"data":{"mid":"LMDWuG2ggqziTRsZRvVwCf5W9Vr6j1QqWNt","version":1,"threshold":1,"keepers":["0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC"],"to":"0x44171C37Ff5D7B7bb8dcad5C81f16284A229e641","amount":0,"kSig":"075b5688d926286f26c6d04e4be059fbf6e4d9f493f028579881652b2b03cb5a3a459aead16b779bcc6636d3469f8cee83a54b855bd94d165d88d93638c1415a00","mSig":"f8ad98293d17b364b613d51b9e9abc5aa7e28eddae7c9dd04cad52c81eb9b5471ae870b2f9267293cf9015a362096e2087417eea9ae1305673187a71fbba6d6101","expire":100,"data":"0xa6616960616e634c444361740161756062657880626673807f1775c3"},"signatures":["6e5018fa44c5eaff3a2ea73bbf535ffae3a35ac0c32705f80dc4ff6889c956e32f7205fb66daa0a4f5250e8490220d2d59600bf61358fb2aa234dc211670252500"],"exSignatures":["a15d6b62ad7b94f8244648bbdc704ff23be86bd74e69648fb2fc47964d3c3a1d47d7ed9beedd6e9fa0401ffd7c61bad2e89280277e2cbd38c627d7a278b5ad1300"],"gas":438,"id":"wfcZU9b8s11idXxkq5oY1kFDPTU7A3K5kjMs2Quvave4ohGvF"}`, string(jsondata))

	assert.NoError(bs.VerifyState())
}

func TestTxCreateNameModelData(t *testing.T) {
	assert := assert.New(t)

	bctx := NewMockBCtx()
	bs := NewMockBS(bctx)

	from, err := bs.LoadAccount(util.Signer1.Address())
	assert.NoError(err)
	to, err := bs.LoadAccount(util.Signer2.Address())
	assert.NoError(err)

	nm, err := service.NameModel()
	assert.NoError(err)
	mm := &ld.ModelMeta{
		Name:      nm.Name(),
		Threshold: 1,
		Keepers:   util.EthIDs{util.Signer2.Address()},
		Data:      nm.Schema(),
		ID:        bctx.Chain().NameServiceID,
	}

	name := &service.Name{
		Name:    "ldc.to.",
		Records: []string{"ldc.to. IN A 10.0.0.1"},
	}
	assert.NoError(name.SyntacticVerify())
	data := name.Bytes()

	input := &ld.TxUpdater{
		ModelID:   &mm.ID,
		Version:   1,
		Threshold: ld.Uint16Ptr(1),
		Keepers:   &util.EthIDs{util.Signer1.Address()},
		Data:      data,
		To:        &to.id,
		Expire:    100,
		Amount:    new(big.Int).SetUint64(constants.MilliLDC),
	}
	kSig, err := util.Signer1.Sign(input.Data)
	assert.NoError(err)
	input.KSig = &kSig
	mSig, err := util.Signer2.Sign(input.Data)
	assert.NoError(err)
	input.MSig = &mSig
	assert.NoError(input.SyntacticVerify())
	txData := &ld.TxData{
		Type:      ld.TypeCreateData,
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
	assert.NoError(txData.ExSignWith(util.Signer2))
	assert.NoError(from.Add(constants.NativeToken, new(big.Int).SetUint64(constants.LDC)))
	assert.NoError(bs.SaveModel(mm.ID, mm))

	tt := txData.ToTransaction()
	tt.Timestamp = 10
	tt.Gas = tt.RequiredGas(bctx.FeeConfig().ThresholdGas)
	itx, err := NewTx(tt, true)
	assert.NoError(err)
	assert.NoError(itx.Verify(bctx, bs))

	id, err := bs.ResolveNameID("ldc.to.")
	assert.ErrorContains(err, `"ldc.to." not found`)
	assert.NoError(itx.Accept(bctx, bs))

	tx := itx.(*TxCreateData)
	assert.Equal(tx.ld.Gas*bctx.Price, tx.ldc.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(tx.ld.Gas*100, tx.miner.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(constants.MilliLDC, to.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(constants.LDC-tx.ld.Gas*(bctx.Price+100)-constants.MilliLDC,
		from.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(uint64(1), tx.from.Nonce())

	id, err = bs.ResolveNameID("ldc.to.")
	assert.NoError(err)
	assert.Equal(id, tx.dm.ID)

	dm, err := bs.ResolveName("ldc.to.")
	assert.NoError(err)
	assert.Equal(mm.ID, dm.ModelID)
	assert.Equal(uint64(1), dm.Version)
	assert.Equal(uint16(1), dm.Threshold)
	assert.Equal(util.EthIDs{from.id}, dm.Keepers)
	assert.Nil(dm.Approver)
	assert.Nil(dm.ApproveList)
	assert.Equal(data, []byte(dm.Data))
	assert.Equal(kSig, dm.KSig)
	assert.Equal(mSig, *dm.MSig)

	n2 := &service.Name{}
	assert.NoError(n2.Unmarshal(dm.Data))
	assert.NoError(n2.SyntacticVerify())
	assert.Equal(n2.Name, name.Name)
	assert.Equal(n2.Records, name.Records)
	assert.Equal(n2.Bytes(), name.Bytes())

	jsondata, err := itx.MarshalJSON()
	assert.NoError(err)
	// fmt.Println(string(jsondata))
	assert.Equal(`{"type":"TypeCreateData","chainID":2357,"nonce":0,"gasTip":100,"gasFeeCap":1000,"from":"0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC","to":"0x44171C37Ff5D7B7bb8dcad5C81f16284A229e641","amount":1000000,"data":{"mid":"LM4rB4RoU8Xa2FAJRVAER8bcprHcpAYFRBs","version":1,"threshold":1,"keepers":["0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC"],"to":"0x44171C37Ff5D7B7bb8dcad5C81f16284A229e641","amount":1000000,"kSig":"5b23ad8462e39ba1e9921d2afb09ecf2c056bc62805a9259860cda8f53c64d422bf2f1d744d5c2979af407b22fb5726410db433f23c0ab644d9c45325442b9c901","mSig":"e286779f92e10b160ca27af117eb935dd31e78bea636921290062c66b5ae1e5a341ae0b80c2716f4b324225e9a4549064a77df943a442fd5b7d49480c722926a01","expire":100,"data":"0xa2616e676c64632e746f2e62727381756c64632e746f2e20494e20412031302e302e302e31851ac289"},"signatures":["fa58aaecd2e3b6dd1e845aa639ae38e28a115466ee39441340021c6ddfc22de57facb02f0a9046f192bf7aef1d6cc44a207fa0302c8becfa1bb88aaacf856fc700"],"exSignatures":["c279c42e5e61b1c5863706f0993f9f65c8c97ea0b6251a61758ca531bffed28c00bbda6f249af41d130f2781594261890881a0e67216a9226b202ba2f77c25fa01"],"gas":457,"id":"2gxKFC6JnY6SFiQJpHb18DK2WqM1C5QbZNaanhTBugdSuvgvJg"}`, string(jsondata))

	assert.NoError(bs.VerifyState())

	name2 := &service.Name{
		Name:    "ldc.to.",
		Records: []string{"ldc.to. IN A 10.0.0.2"},
	}
	assert.NoError(name2.SyntacticVerify())
	data = name2.Bytes()

	input = &ld.TxUpdater{
		ModelID:   &mm.ID,
		Version:   1,
		Threshold: ld.Uint16Ptr(1),
		Keepers:   &util.EthIDs{util.Signer1.Address()},
		Data:      data,
		To:        &to.id,
		Expire:    100,
		Amount:    new(big.Int).SetUint64(constants.MilliLDC),
	}
	kSig, err = util.Signer1.Sign(input.Data)
	assert.NoError(err)
	input.KSig = &kSig
	mSig, err = util.Signer2.Sign(input.Data)
	assert.NoError(err)
	input.MSig = &mSig
	assert.NoError(input.SyntacticVerify())
	txData = &ld.TxData{
		Type:      ld.TypeCreateData,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     1,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      from.id,
		To:        &to.id,
		Amount:    new(big.Int).SetUint64(constants.MilliLDC),
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	assert.NoError(txData.ExSignWith(util.Signer2))

	tt = txData.ToTransaction()
	tt.Timestamp = 10
	tt.Gas = tt.RequiredGas(bctx.FeeConfig().ThresholdGas)
	itx, err = NewTx(tt, true)
	assert.NoError(err)
	assert.ErrorContains(itx.Verify(bctx, bs),
		`TxCreateData.Verify failed: name "ldc.to." conflict`)

	name2 = &service.Name{
		Name:    "api.ldc.to.",
		Records: []string{},
	}
	assert.NoError(name2.SyntacticVerify())
	data = name2.Bytes()

	input = &ld.TxUpdater{
		ModelID:   &mm.ID,
		Version:   1,
		Threshold: ld.Uint16Ptr(1),
		Keepers:   &util.EthIDs{util.Signer1.Address()},
		Data:      data,
		To:        &to.id,
		Expire:    100,
		Amount:    new(big.Int).SetUint64(constants.MilliLDC),
	}
	kSig, err = util.Signer1.Sign(input.Data)
	assert.NoError(err)
	input.KSig = &kSig
	mSig, err = util.Signer2.Sign(input.Data)
	assert.NoError(err)
	input.MSig = &mSig
	assert.NoError(input.SyntacticVerify())
	txData = &ld.TxData{
		Type:      ld.TypeCreateData,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     1,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      from.id,
		To:        &to.id,
		Amount:    new(big.Int).SetUint64(constants.MilliLDC),
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	assert.NoError(txData.ExSignWith(util.Signer2))

	tt = txData.ToTransaction()
	tt.Timestamp = 10
	tt.Gas = tt.RequiredGas(bctx.FeeConfig().ThresholdGas)
	itx, err = NewTx(tt, true)
	assert.NoError(err)
	assert.NoError(itx.Verify(bctx, bs))
}

func TestTxCreateDataGenesis(t *testing.T) {
	assert := assert.New(t)

	bctx := NewMockBCtx()
	bs := NewMockBS(bctx)

	from, err := bs.LoadAccount(util.Signer1.Address())
	assert.NoError(err)

	cfg, err := json.Marshal(bctx.Chain().FeeConfig)
	assert.NoError(err)

	cfgData := &ld.TxUpdater{
		ModelID:   &constants.JSONModelID,
		Version:   1,
		Threshold: ld.Uint16Ptr(1),
		Keepers:   &util.EthIDs{util.Signer1.Address()},
		Data:      cfg,
	}
	assert.NoError(cfgData.SyntacticVerify())
	tt := &ld.Transaction{
		Type:    ld.TypeCreateData,
		ChainID: bctx.Chain().ChainID,
		From:    from.id,
		Data:    ld.MustMarshal(cfgData),
	}
	assert.NoError(tt.SyntacticVerify())

	itx, err := NewGenesisTx(tt)
	assert.NoError(err)
	assert.NoError(itx.(GenesisTx).VerifyGenesis(bctx, bs))
	assert.NoError(itx.Accept(bctx, bs))

	tx := itx.(*TxCreateData)
	assert.Equal(uint64(0), tx.ldc.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(uint64(0), tx.miner.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(uint64(0), from.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(uint64(1), from.Nonce())

	dm, err := bs.LoadData(tx.dm.ID)
	assert.NoError(err)
	assert.Equal(constants.JSONModelID, dm.ModelID)
	assert.Equal(uint64(1), dm.Version)
	assert.Equal(uint16(1), dm.Threshold)
	assert.Equal(util.EthIDs{from.id}, dm.Keepers)
	assert.Nil(dm.Approver)
	assert.Nil(dm.ApproveList)
	assert.True(jsonpatch.Equal(cfg, dm.Data))

	jsondata, err := itx.MarshalJSON()
	assert.NoError(err)
	// fmt.Println(string(jsondata))
	assert.Equal(`{"type":"TypeCreateData","chainID":2357,"nonce":0,"gasTip":0,"gasFeeCap":0,"from":"0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC","data":{"mid":"LM1111111111111111111L17Xp3","version":1,"threshold":1,"keepers":["0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC"],"data":{"startHeight":0,"thresholdGas":1000,"minGasPrice":10000,"maxGasPrice":100000,"maxTxGas":42000000,"maxBlockTxsSize":4200000,"gasRebateRate":1000,"minTokenPledge":10000000000000,"minStakePledge":1000000000000}},"gas":0,"id":"7o7fYNFS27SGZF8uEXC8PExi5mgVW4p4Hj7gdCg1wEqCUB1qk"}`, string(jsondata))

	assert.NoError(bs.VerifyState())
}
