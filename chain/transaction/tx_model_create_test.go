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

func TestTxCreateModel(t *testing.T) {
	assert := assert.New(t)

	// SyntacticVerify
	tx := &TxCreateModel{}
	assert.ErrorContains(tx.SyntacticVerify(), "nil pointer")
	_, err := tx.MarshalJSON()
	assert.NoError(err)

	bctx := NewMockBCtx()
	bs := NewMockBS(bctx)
	token := ld.MustNewToken("$LDC")

	from, err := bs.LoadAccount(util.Signer1.Address())
	assert.NoError(err)

	txData := &ld.TxData{
		Type:      ld.TypeCreateModel,
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
		Type:      ld.TypeCreateModel,
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
		Type:      ld.TypeCreateModel,
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
		Type:      ld.TypeCreateModel,
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
		Type:      ld.TypeCreateModel,
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
		Type:      ld.TypeCreateModel,
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

	input := &ld.ModelMeta{}
	assert.ErrorContains(input.SyntacticVerify(), "ModelMeta.SyntacticVerify failed: invalid name")
	ipldm, err := service.ProfileModel()
	assert.NoError(err)
	input = &ld.ModelMeta{
		Name:      ipldm.Name(),
		Threshold: 1,
		Keepers:   util.EthIDs{util.Signer1.Address()},
		Data:      ipldm.Schema(),
	}
	txData = &ld.TxData{
		Type:      ld.TypeCreateModel,
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
	assert.ErrorContains(itx.Verify(bctx, bs), "invalid gas, expected 994, got 0")

	tt.Gas = tt.RequiredGas(bctx.FeeConfig().ThresholdGas)
	itx, err = NewTx(tt, true)
	assert.NoError(err)
	assert.ErrorContains(itx.Verify(bctx, bs),
		"insufficient NativeLDC balance, expected 1093400, got 0")
	from.Add(constants.NativeToken, new(big.Int).SetUint64(constants.LDC))
	assert.NoError(itx.Verify(bctx, bs))
	assert.NoError(itx.Accept(bctx, bs))

	tx = itx.(*TxCreateModel)
	assert.Equal(tx.ld.Gas*bctx.Price, tx.ldc.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(tx.ld.Gas*100, tx.miner.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(constants.LDC-tx.ld.Gas*(bctx.Price+100),
		from.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(uint64(1), tx.from.Nonce())

	mm, err := bs.LoadModel(tx.input.ID)
	assert.NoError(err)
	assert.Equal(input.Name, mm.Name)
	assert.Equal(input.Data, mm.Data)

	jsondata, err := itx.MarshalJSON()
	assert.NoError(err)
	// fmt.Println(string(jsondata))
	assert.Equal(`{"type":"TypeCreateModel","chainID":2357,"nonce":0,"gasTip":100,"gasFeeCap":1000,"from":"0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC","data":{"name":"ProfileService","threshold":1,"keepers":["0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC"],"data":"0x0a097479706520494432302062797465730a09747970652050726f66696c655365727669636520737472756374207b0a09097479706520202020202020496e74202020202020202020202020202872656e616d6520227422290a09096e616d6520202020202020537472696e67202020202020202020202872656e616d6520226e22290a0909696d616765202020202020537472696e67202020202020202020202872656e616d6520226922290a090975726c2020202020202020537472696e67202020202020202020202872656e616d6520227522290a09096b796320202020202020206f7074696f6e616c20494432302020202872656e616d6520226b22290a0909666f6c6c6f7773202020205b494432305d202020202020202020202872656e616d652022667322290a09096d656d62657273202020206f7074696f6e616c205b494432305d202872656e616d6520226d7322290a0909657874656e73696f6e73205b416e795d20202020202020202020202872656e616d652022657822290a097d0ad67eed64","id":"LM8NjhGKzwhrZUrnihmnyWiHwEXZFPZRif5"},"signatures":["1b7ea1de77fbeb2c28a24a05eab7b6d72420746f08218a7a48f64838bc6800731ce43c33d51349b33af0b8d5c3d3fcacf129d3c013941684c262426bbe500f0f01"],"gas":994,"id":"cdNKrh2Mz3hrnQUvjiUtnKCbzvx2MZUsh8BmqMkvwQqZUZkcc"}`, string(jsondata))

	assert.NoError(bs.VerifyState())
}

func TestTxCreateModelGenesis(t *testing.T) {
	assert := assert.New(t)

	bctx := NewMockBCtx()
	bs := NewMockBS(bctx)

	from, err := bs.LoadAccount(util.Signer1.Address())
	assert.NoError(err)

	nm, err := service.NameModel()
	assert.NoError(err)
	mm := &ld.ModelMeta{
		Name:      nm.Name(),
		Threshold: *ld.Uint8Ptr(1),
		Keepers:   util.EthIDs{util.Signer1.Address()},
		Data:      nm.Schema(),
	}
	assert.NoError(mm.SyntacticVerify())

	tt := &ld.Transaction{
		Type:    ld.TypeCreateModel,
		ChainID: bctx.Chain().ChainID,
		From:    from.id,
		Data:    ld.MustMarshal(mm),
	}
	assert.NoError(tt.SyntacticVerify())

	itx, err := NewGenesisTx(tt)
	assert.NoError(err)
	assert.NoError(itx.(GenesisTx).VerifyGenesis(bctx, bs))
	assert.NoError(itx.Accept(bctx, bs))

	tx := itx.(*TxCreateModel)
	assert.Equal(uint64(0), tx.ldc.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(uint64(0), tx.miner.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(uint64(0), from.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(uint64(1), from.Nonce())

	mm2, err := bs.LoadModel(tx.input.ID)
	assert.NoError(err)
	assert.Equal(uint8(1), mm2.Threshold)
	assert.Equal(util.EthIDs{from.id}, mm2.Keepers)
	assert.Nil(mm2.Approver)
	assert.Equal(mm.Name, mm2.Name)
	assert.Equal(mm.Data, mm2.Data)

	jsondata, err := itx.MarshalJSON()
	assert.NoError(err)
	// fmt.Println(string(jsondata))
	assert.Equal(`{"type":"TypeCreateModel","chainID":2357,"nonce":0,"gasTip":0,"gasFeeCap":0,"from":"0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC","data":{"name":"NameService","threshold":1,"keepers":["0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC"],"data":"0x0a097479706520494432302062797465730a0974797065204e616d655365727669636520737472756374207b0a09096e616d6520202020537472696e6720202020202020202872656e616d6520226e22290a09096c696e6b656420206f7074696f6e616c2049443230202872656e616d6520226c22290a09097265636f726473205b537472696e675d2020202020202872656e616d652022727322290a097d0ad7077bb1","id":"LMG7314Qg687h2DRg3WHszkeM9wHsGJVEDM"},"gas":0,"id":"2Fy56G1oPMf5xW1MTCDHtc4DqX2M6td8TjG2d41yWsaZwBzkFy"}`, string(jsondata))

	pm, err := service.ProfileModel()
	assert.NoError(err)
	mm = &ld.ModelMeta{
		Name:      pm.Name(),
		Threshold: *ld.Uint8Ptr(1),
		Keepers:   util.EthIDs{util.Signer1.Address()},
		Data:      pm.Schema(),
	}
	assert.NoError(mm.SyntacticVerify())

	tt = &ld.Transaction{
		Type:    ld.TypeCreateModel,
		ChainID: bctx.Chain().ChainID,
		Nonce:   1,
		From:    from.id,
		Data:    ld.MustMarshal(mm),
	}
	assert.NoError(tt.SyntacticVerify())

	itx, err = NewGenesisTx(tt)
	assert.NoError(err)
	assert.NoError(itx.(GenesisTx).VerifyGenesis(bctx, bs))
	assert.NoError(itx.Accept(bctx, bs))

	tx = itx.(*TxCreateModel)
	assert.Equal(uint64(0), tx.ldc.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(uint64(0), tx.miner.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(uint64(0), from.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(uint64(2), from.Nonce())

	mm2, err = bs.LoadModel(tx.input.ID)
	assert.NoError(err)
	assert.Equal(uint8(1), mm2.Threshold)
	assert.Equal(util.EthIDs{from.id}, mm2.Keepers)
	assert.Nil(mm2.Approver)
	assert.Equal(mm.Name, mm2.Name)
	assert.Equal(mm.Data, mm2.Data)

	jsondata, err = itx.MarshalJSON()
	assert.NoError(err)
	// fmt.Println(string(jsondata))
	assert.Equal(`{"type":"TypeCreateModel","chainID":2357,"nonce":1,"gasTip":0,"gasFeeCap":0,"from":"0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC","data":{"name":"ProfileService","threshold":1,"keepers":["0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC"],"data":"0x0a097479706520494432302062797465730a09747970652050726f66696c655365727669636520737472756374207b0a09097479706520202020202020496e74202020202020202020202020202872656e616d6520227422290a09096e616d6520202020202020537472696e67202020202020202020202872656e616d6520226e22290a0909696d616765202020202020537472696e67202020202020202020202872656e616d6520226922290a090975726c2020202020202020537472696e67202020202020202020202872656e616d6520227522290a09096b796320202020202020206f7074696f6e616c20494432302020202872656e616d6520226b22290a0909666f6c6c6f7773202020205b494432305d202020202020202020202872656e616d652022667322290a09096d656d62657273202020206f7074696f6e616c205b494432305d202872656e616d6520226d7322290a0909657874656e73696f6e73205b416e795d20202020202020202020202872656e616d652022657822290a097d0ad67eed64","id":"LMvKxf23G7SmT3sJreTJzDVoqA44UeoddP"},"gas":0,"id":"5ScuQntPgZY9vDiJorKhW1WmZ14GRw3VCVV9eGgdbbX9wX591"}`, string(jsondata))

	assert.NoError(bs.VerifyState())
}
