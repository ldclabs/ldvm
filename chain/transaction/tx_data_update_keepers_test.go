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
	assert.ErrorContains(err, "DeriveSigners error: no signature")

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

	di := &ld.DataInfo{
		ModelID:   constants.RawModelID,
		Version:   2,
		Threshold: 1,
		Keepers:   util.EthIDs{util.Signer2.Address()},
		Data:      []byte(`42`),
		ID:        did,
	}
	kSig, err := util.Signer2.Sign(di.Data)
	assert.NoError(err)
	di.KSig = kSig
	assert.NoError(di.SyntacticVerify())

	input = &ld.TxUpdater{ID: &did, Version: 1,
		Approver:    &approver,
		ApproveList: []ld.TxType{ld.TypeDeleteData},
		Threshold:   ld.Uint16Ptr(1),
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

	assert.NoError(bs.SaveData(di.ID, di))
	assert.ErrorContains(itx.Verify(bctx, bs),
		"invalid version, expected 2, got 1")

	input = &ld.TxUpdater{ID: &did, Version: 2,
		Approver:    &approver,
		ApproveList: []ld.TxType{ld.TypeDeleteData},
		Threshold:   ld.Uint16Ptr(1),
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

	kSig, err = util.Signer1.Sign(di.Data)
	assert.NoError(err)
	input = &ld.TxUpdater{ID: &did, Version: 2,
		Approver: &approver,
		ApproveList: []ld.TxType{
			ld.TypeUpdateDataKeepers,
			ld.TypeUpdateDataKeepersByAuth,
			ld.TypeDeleteData},
		Threshold: ld.Uint16Ptr(1),
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

	di2, err := bs.LoadData(di.ID)
	assert.NoError(err)
	assert.Equal(di.Version+1, di2.Version)
	assert.NotEqual(kSig, di.KSig)
	assert.Equal(kSig, di2.KSig)
	assert.Equal(uint16(1), di2.Threshold)
	assert.Equal(util.EthIDs{util.Signer2.Address()}, di.Keepers)
	assert.Equal(util.EthIDs{util.Signer1.Address()}, di2.Keepers)
	assert.Equal(di.Data, di2.Data)
	assert.Nil(di.Approver)
	assert.NotNil(di2.Approver)
	assert.Equal(util.Signer2.Address(), *di2.Approver)
	assert.Nil(di.ApproveList)
	assert.Equal(ld.TxTypes{
		ld.TypeUpdateDataKeepers,
		ld.TypeUpdateDataKeepersByAuth,
		ld.TypeDeleteData}, di2.ApproveList)

	jsondata, err := itx.MarshalJSON()
	assert.NoError(err)
	// fmt.Println(string(jsondata))
	assert.Equal(`{"type":"TypeUpdateDataKeepers","chainID":2357,"nonce":0,"gasTip":100,"gasFeeCap":1000,"from":"0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC","data":{"id":"LD6L5yRJL2iYi9PbrhRru6uKfEAzDGHwUJ","version":2,"threshold":1,"keepers":["0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC"],"approver":"0x44171C37Ff5D7B7bb8dcad5C81f16284A229e641","approveList":["TypeUpdateDataKeepers","TypeUpdateDataKeepersByAuth","TypeDeleteData"],"kSig":"505a3dfb3372ef790ba8237ab40a53f8e626b56b3778f9edcb67436ea1ac9fd65a7a10f80921aa34809a056c18f8cd9f905367c65b30734e137428554e71735001"},"signatures":["2062c8e0867f4dd2cbac4137ab7ced3d57deafed9393904c3b3d0a6ed631f942101466787868cffb55262bd5e8c4f460d646b2e645ffd660eb8b8da3a0ab7c7a00","6bc7e9fcf0c89d4a754373842293f2996665f208470b69e6176627a1bb8099bc1825ff22cb171b89c06346039da4a4ba47a90642ce6699c57065ddb4f40afedc00"],"gas":310,"id":"D7b91sptH22A7J3mmhpFhNwrkFxfdhcr3jgP4nFCpLeXyGgkP"}`, string(jsondata))

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

	di2, err = bs.LoadData(di.ID)
	assert.NoError(err)
	assert.Equal(uint64(4), di2.Version)
	assert.Equal(kSig, di2.KSig)
	assert.Equal(uint16(1), di2.Threshold)
	assert.Equal(util.EthIDs{util.Signer1.Address()}, di2.Keepers)
	assert.Equal(di.Data, di2.Data)
	assert.Nil(di2.Approver)
	assert.Nil(di2.ApproveList)

	// clear threshold
	input = &ld.TxUpdater{ID: &did, Version: 4,
		Threshold: ld.Uint16Ptr(0),
		Keepers:   &util.EthIDs{util.Signer1.Address()},
	}
	txData = &ld.TxData{
		Type:      ld.TypeUpdateDataKeepers,
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
	assert.NoError(itx.Verify(bctx, bs))
	assert.NoError(itx.Accept(bctx, bs))

	di2, err = bs.LoadData(di.ID)
	assert.NoError(err)
	assert.Equal(uint64(5), di2.Version)
	assert.Equal(kSig, di2.KSig)
	assert.Equal(uint16(0), di2.Threshold)
	assert.Equal(util.EthIDs{util.Signer1.Address()}, di2.Keepers)

	// clear keepers
	input = &ld.TxUpdater{ID: &did, Version: 5,
		Threshold: ld.Uint16Ptr(0),
		Keepers:   &util.EthIDs{},
	}
	txData = &ld.TxData{
		Type:      ld.TypeUpdateDataKeepers,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     3,
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

	di2, err = bs.LoadData(di.ID)
	assert.NoError(err)
	assert.Equal(uint64(6), di2.Version)
	assert.Equal(kSig, di2.KSig)
	assert.Equal(uint16(0), di2.Threshold)
	assert.Equal(util.EthIDs{}, di2.Keepers)

	// can't update keepers
	input = &ld.TxUpdater{ID: &did, Version: 6,
		Threshold: ld.Uint16Ptr(0),
		Keepers:   &util.EthIDs{util.Signer1.Address()},
	}
	txData = &ld.TxData{
		Type:      ld.TypeUpdateDataKeepers,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     4,
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

	// can't update data
	input = &ld.TxUpdater{ID: &did, Version: 6,
		Data: []byte(`421`),
	}
	kSig, err = util.Signer1.Sign([]byte(`421`))
	assert.NoError(err)
	input.KSig = &kSig
	txData = &ld.TxData{
		Type:      ld.TypeUpdateData,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     4,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      from.id,
		Data:      input.Bytes(),
	}
	tt = txData.ToTransaction()
	tt.Gas = tt.RequiredGas(bctx.FeeConfig().ThresholdGas)
	itx, err = NewTx(tt, true)
	assert.ErrorContains(err, "DeriveSigners error: no signature")

	assert.NoError(txData.SignWith(util.Signer1))
	tt = txData.ToTransaction()
	tt.Gas = tt.RequiredGas(bctx.FeeConfig().ThresholdGas)
	itx, err = NewTx(tt, true)
	assert.NoError(err)
	assert.ErrorContains(itx.Verify(bctx, bs), "invalid signatures for data keepers")

	// can't delete data
	input = &ld.TxUpdater{ID: &did, Version: 6, Data: []byte(`421`)}
	txData = &ld.TxData{
		Type:      ld.TypeDeleteData,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     4,
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

	assert.NoError(bs.VerifyState())
}
