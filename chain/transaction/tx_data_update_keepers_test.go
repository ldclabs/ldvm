// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package transaction

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/ldclabs/ldvm/constants"
	"github.com/ldclabs/ldvm/ld"
	"github.com/ldclabs/ldvm/util"
	"github.com/stretchr/testify/assert"
)

func TestTxUpdateDataKeepers(t *testing.T) {
	assert := assert.New(t)

	// SyntacticVerify
	tx := &TxUpdateDataKeepers{}
	assert.ErrorContains(tx.SyntacticVerify(), "nil pointer")
	_, err := tx.MarshalJSON()
	assert.NoError(err)

	bctx := NewMockBCtx()
	bs := NewMockBS(bctx)

	from, err := bs.LoadAccount(util.Signer1.Address())
	assert.NoError(err)
	approver := util.Signer2.Address()

	txData := &ld.TxData{
		Type:      ld.TypeUpdateDataKeepers,
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
		Type:      ld.TypeUpdateDataKeepers,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      from.id,
		To:        &approver,
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx(txData.ToTransaction(), true)
	assert.ErrorContains(err, "invalid to, should be nil")

	txData = &ld.TxData{
		Type:      ld.TypeUpdateDataKeepers,
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
		Type:      ld.TypeUpdateDataKeepers,
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
		Type:      ld.TypeUpdateDataKeepers,
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
		Type:      ld.TypeUpdateDataKeepers,
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
		Type:      ld.TypeUpdateDataKeepers,
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
		Type:      ld.TypeUpdateDataKeepers,
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
		Type:      ld.TypeUpdateDataKeepers,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      from.id,
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx(txData.ToTransaction(), true)
	assert.ErrorContains(err, "no thing to update")

	dm := &ld.DataMeta{
		ModelID:   constants.RawModelID,
		Version:   2,
		Threshold: 1,
		Keepers:   util.EthIDs{util.Signer2.Address()},
		Data:      []byte(`42`),
		ID:        did,
	}
	kSig, err := util.Signer2.Sign(dm.Data)
	assert.NoError(err)
	dm.KSig = kSig
	assert.NoError(dm.SyntacticVerify())

	input = &ld.TxUpdater{ID: &did, Version: 1,
		Approver:    &approver,
		ApproveList: []ld.TxType{ld.TypeDeleteData},
		Threshold:   ld.Uint8Ptr(1),
		Keepers:     &util.EthIDs{util.Signer1.Address()},
		KSig:        &kSig,
	}
	txData = &ld.TxData{
		Type:      ld.TypeUpdateDataKeepers,
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
	assert.ErrorContains(itx.Verify(bctx, bs), "invalid gas, expected 308, got 0")

	tt.Gas = tt.RequiredGas(bctx.FeeConfig().ThresholdGas)
	itx, err = NewTx(tt, true)
	assert.NoError(err)
	assert.ErrorContains(itx.Verify(bctx, bs),
		"insufficient NativeLDC balance, expected 338800, got 0")

	from.Add(constants.NativeToken, new(big.Int).SetUint64(constants.LDC))
	assert.ErrorContains(itx.Verify(bctx, bs),
		"LD6L5yRJL2iYi9PbrhRru6uKfEAzDGHwUJ not found")

	assert.NoError(bs.SaveData(dm.ID, dm))
	assert.ErrorContains(itx.Verify(bctx, bs),
		"invalid version, expected 2, got 1")

	input = &ld.TxUpdater{ID: &did, Version: 2,
		Approver:    &approver,
		ApproveList: []ld.TxType{ld.TypeDeleteData},
		Threshold:   ld.Uint8Ptr(1),
		Keepers:     &util.EthIDs{util.Signer1.Address()},
		KSig:        &kSig,
	}
	txData = &ld.TxData{
		Type:      ld.TypeUpdateDataKeepers,
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

	assert.NoError(txData.SignWith(util.Signer2))
	tt = txData.ToTransaction()
	tt.Gas = tt.RequiredGas(bctx.FeeConfig().ThresholdGas)
	itx, err = NewTx(tt, true)
	assert.NoError(err)
	assert.ErrorContains(itx.Verify(bctx, bs), "invalid kSig")

	kSig, err = util.Signer1.Sign(dm.Data)
	assert.NoError(err)
	input = &ld.TxUpdater{ID: &did, Version: 2,
		Approver: &approver,
		ApproveList: []ld.TxType{
			ld.TypeUpdateDataKeepers,
			ld.TypeUpdateDataKeepersByAuth,
			ld.TypeDeleteData},
		Threshold: ld.Uint8Ptr(1),
		Keepers:   &util.EthIDs{util.Signer1.Address()},
		KSig:      &kSig,
	}
	txData = &ld.TxData{
		Type:      ld.TypeUpdateDataKeepers,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     0,
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

	tx = itx.(*TxUpdateDataKeepers)
	assert.Equal(tx.ld.Gas*bctx.Price, tx.ldc.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(tx.ld.Gas*100, tx.miner.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(constants.LDC-tx.ld.Gas*(bctx.Price+100),
		from.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(uint64(1), from.Nonce())

	dm2, err := bs.LoadData(dm.ID)
	assert.NoError(err)
	assert.Equal(dm.Version+1, dm2.Version)
	assert.NotEqual(kSig, dm.KSig)
	assert.Equal(kSig, dm2.KSig)
	assert.Equal(uint8(1), dm2.Threshold)
	assert.Equal(util.EthIDs{util.Signer2.Address()}, dm.Keepers)
	assert.Equal(util.EthIDs{util.Signer1.Address()}, dm2.Keepers)
	assert.Equal(dm.Data, dm2.Data)
	assert.Nil(dm.Approver)
	assert.NotNil(dm2.Approver)
	assert.Equal(util.Signer2.Address(), *dm2.Approver)
	assert.Nil(dm.ApproveList)
	assert.Equal([]ld.TxType{
		ld.TypeUpdateDataKeepers,
		ld.TypeUpdateDataKeepersByAuth,
		ld.TypeDeleteData}, dm2.ApproveList)

	jsondata, err := itx.MarshalJSON()
	assert.NoError(err)
	fmt.Println(string(jsondata))
	assert.Equal(`{"type":21,"chainID":2357,"nonce":0,"gasTip":100,"gasFeeCap":1000,"from":"0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC","data":{"id":"LD6L5yRJL2iYi9PbrhRru6uKfEAzDGHwUJ","version":2,"threshold":1,"keepers":["0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC"],"approver":"0x44171C37Ff5D7B7bb8dcad5C81f16284A229e641","approveList":"FRYX","kSig":"505a3dfb3372ef790ba8237ab40a53f8e626b56b3778f9edcb67436ea1ac9fd65a7a10f80921aa34809a056c18f8cd9f905367c65b30734e137428554e71735001"},"signatures":["b42c3d5bd42cf996afb105bd31e74cfea70e5e69c70d56978da4edbf7691e500529fbeb9eec4adb016c1d3e10be01b709d18b450e418f3564109ad8824c997ff00","406f90c21685e7e26039a45d8f236b98d2cf70e431b706166808a2ed63c6705f58436f9c2a46d513935bb6ceabe067d99cf8791aded747677966c5963d597e5701"],"gas":310,"name":"UpdateDataKeepersTx","id":"278NGpRg8mA7gM7BKGUYKSbVzVWYLupboK4WLodjFVQHQo9QAL"}`, string(jsondata))

	input = &ld.TxUpdater{ID: &did, Version: 3,
		Approver: &util.EthIDEmpty,
	}
	txData = &ld.TxData{
		Type:      ld.TypeUpdateDataKeepers,
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
	assert.ErrorContains(itx.Verify(bctx, bs), "invalid signature for data approver")

	assert.NoError(txData.SignWith(util.Signer2))
	tt = txData.ToTransaction()
	tt.Gas = tt.RequiredGas(bctx.FeeConfig().ThresholdGas)
	itx, err = NewTx(tt, true)
	assert.NoError(err)
	assert.NoError(itx.Verify(bctx, bs))
	assert.NoError(itx.Accept(bctx, bs))

	dm2, err = bs.LoadData(dm.ID)
	assert.NoError(err)
	assert.Equal(uint64(4), dm2.Version)
	assert.Equal(kSig, dm2.KSig)
	assert.Equal(uint8(1), dm2.Threshold)
	assert.Equal(util.EthIDs{util.Signer1.Address()}, dm2.Keepers)
	assert.Equal(dm.Data, dm2.Data)
	assert.Nil(dm2.Approver)
	assert.Nil(dm2.ApproveList)

	assert.NoError(bs.VerifyState())
}
