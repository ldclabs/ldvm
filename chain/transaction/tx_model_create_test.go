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
	assert.ErrorContains(err, "DeriveSigners error: no signature")

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

	input := &ld.ModelInfo{}
	assert.ErrorContains(input.SyntacticVerify(), "ModelInfo.SyntacticVerify error: invalid name")
	ipldm, err := service.ProfileModel()
	assert.NoError(err)
	input = &ld.ModelInfo{
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
	assert.ErrorContains(itx.Verify(bctx, bs), "invalid gas, expected 952, got 0")

	tt.Gas = tt.RequiredGas(bctx.FeeConfig().ThresholdGas)
	itx, err = NewTx(tt, true)
	assert.NoError(err)
	assert.ErrorContains(itx.Verify(bctx, bs),
		"insufficient NativeLDC balance, expected 1047200, got 0")
	from.Add(constants.NativeToken, new(big.Int).SetUint64(constants.LDC))
	assert.NoError(itx.Verify(bctx, bs))
	assert.NoError(itx.Accept(bctx, bs))

	tx = itx.(*TxCreateModel)
	assert.Equal(tx.ld.Gas*bctx.Price, tx.ldc.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(tx.ld.Gas*100, tx.miner.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(constants.LDC-tx.ld.Gas*(bctx.Price+100),
		from.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(uint64(1), tx.from.Nonce())

	mi, err := bs.LoadModel(tx.input.ID)
	assert.NoError(err)
	assert.Equal(input.Name, mi.Name)
	assert.Equal(input.Data, mi.Data)

	jsondata, err := itx.MarshalJSON()
	assert.NoError(err)
	// fmt.Println(string(jsondata))
	assert.Equal(`{"type":"TypeCreateModel","chainID":2357,"nonce":0,"gasTip":100,"gasFeeCap":1000,"from":"0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC","data":{"name":"ProfileService","threshold":1,"keepers":["0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC"],"data":"0x0a097479706520494432302062797465730a09747970652050726f66696c655365727669636520737472756374207b0a09097479706520202020202020496e74202020202020202020202020202872656e616d6520227422290a09096e616d6520202020202020537472696e67202020202020202020202872656e616d6520226e22290a0909696d616765202020202020537472696e67202020202020202020202872656e616d6520226922290a090975726c2020202020202020537472696e67202020202020202020202872656e616d6520227522290a0909666f6c6c6f7773202020205b494432305d202020202020202020202872656e616d652022667322290a09096d656d62657273202020206f7074696f6e616c205b494432305d202872656e616d6520226d7322290a0909657874656e73696f6e73205b416e795d20202020202020202020202872656e616d652022657822290a097d0a1d0d5a60","id":"LM9FuFurGH4WzjnhVd3FVp9JB4g38xAuNQi"},"signatures":["075346d8a5b2a0a50fa739a9c0676664cadac52498e7f59d8b61701e8d0f7be61e7a9d6ddd18b042877d85aa7577e3c91db67891731fb68fa9c96b0eed0e1b3901"],"gas":952,"id":"gtYSo91ozbBBW8YTQwaFLsUJJq16ExVSt6QiVcDKFy5CaZmzs"}`, string(jsondata))

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
	mi := &ld.ModelInfo{
		Name:      nm.Name(),
		Threshold: *ld.Uint16Ptr(1),
		Keepers:   util.EthIDs{util.Signer1.Address()},
		Data:      nm.Schema(),
	}
	assert.NoError(mi.SyntacticVerify())

	tt := &ld.Transaction{
		Type:    ld.TypeCreateModel,
		ChainID: bctx.Chain().ChainID,
		From:    from.id,
		Data:    ld.MustMarshal(mi),
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

	mi2, err := bs.LoadModel(tx.input.ID)
	assert.NoError(err)
	assert.Equal(uint16(1), mi2.Threshold)
	assert.Equal(util.EthIDs{from.id}, mi2.Keepers)
	assert.Nil(mi2.Approver)
	assert.Equal(mi.Name, mi2.Name)
	assert.Equal(mi.Data, mi2.Data)

	jsondata, err := itx.MarshalJSON()
	assert.NoError(err)
	// fmt.Println(string(jsondata))
	assert.Equal(`{"type":"TypeCreateModel","chainID":2357,"nonce":0,"gasTip":0,"gasFeeCap":0,"from":"0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC","data":{"name":"NameService","threshold":1,"keepers":["0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC"],"data":"0x0a097479706520494432302062797465730a0974797065204e616d655365727669636520737472756374207b0a09096e616d6520202020537472696e6720202020202020202872656e616d6520226e22290a09096c696e6b656420206f7074696f6e616c2049443230202872656e616d6520226c22290a09097265636f726473205b537472696e675d2020202020202872656e616d652022727322290a097d0ad7077bb1","id":"LMG7314Qg687h2DRg3WHszkeM9wHsGJVEDM"},"gas":0,"id":"2Fy56G1oPMf5xW1MTCDHtc4DqX2M6td8TjG2d41yWsaZwBzkFy"}`, string(jsondata))

	pm, err := service.ProfileModel()
	assert.NoError(err)
	mi = &ld.ModelInfo{
		Name:      pm.Name(),
		Threshold: *ld.Uint16Ptr(1),
		Keepers:   util.EthIDs{util.Signer1.Address()},
		Data:      pm.Schema(),
	}
	assert.NoError(mi.SyntacticVerify())

	tt = &ld.Transaction{
		Type:    ld.TypeCreateModel,
		ChainID: bctx.Chain().ChainID,
		Nonce:   1,
		From:    from.id,
		Data:    ld.MustMarshal(mi),
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

	mi2, err = bs.LoadModel(tx.input.ID)
	assert.NoError(err)
	assert.Equal(uint16(1), mi2.Threshold)
	assert.Equal(util.EthIDs{from.id}, mi2.Keepers)
	assert.Nil(mi2.Approver)
	assert.Equal(mi.Name, mi2.Name)
	assert.Equal(mi.Data, mi2.Data)

	jsondata, err = itx.MarshalJSON()
	assert.NoError(err)
	// fmt.Println(string(jsondata))
	assert.Equal(`{"type":"TypeCreateModel","chainID":2357,"nonce":1,"gasTip":0,"gasFeeCap":0,"from":"0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC","data":{"name":"ProfileService","threshold":1,"keepers":["0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC"],"data":"0x0a097479706520494432302062797465730a09747970652050726f66696c655365727669636520737472756374207b0a09097479706520202020202020496e74202020202020202020202020202872656e616d6520227422290a09096e616d6520202020202020537472696e67202020202020202020202872656e616d6520226e22290a0909696d616765202020202020537472696e67202020202020202020202872656e616d6520226922290a090975726c2020202020202020537472696e67202020202020202020202872656e616d6520227522290a0909666f6c6c6f7773202020205b494432305d202020202020202020202872656e616d652022667322290a09096d656d62657273202020206f7074696f6e616c205b494432305d202872656e616d6520226d7322290a0909657874656e73696f6e73205b416e795d20202020202020202020202872656e616d652022657822290a097d0a1d0d5a60","id":"LM3j1qoRABHpE1R518UrWPgV8KFkLvd9gEc"},"gas":0,"id":"EAMob5Uwf3DgyZiu2KHwED9DR2HXGMH8Sf7wo2ymJWMQf5vXh"}`, string(jsondata))

	assert.NoError(bs.VerifyState())
}
