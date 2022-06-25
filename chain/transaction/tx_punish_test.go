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

func TestTxPunish(t *testing.T) {
	assert := assert.New(t)

	// SyntacticVerify
	tx := &TxPunish{}
	assert.ErrorContains(tx.SyntacticVerify(), "nil pointer")
	_, err := tx.MarshalJSON()
	assert.NoError(err)

	bctx := NewMockBCtx()
	bs := bctx.MockBS()

	from := bs.MustAccount(constants.GenesisAccount)
	assert.NoError(err)
	singer1 := util.Signer1.Address()
	assert.NoError(from.UpdateKeepers(ld.Uint16Ptr(1), &util.EthIDs{singer1}, nil, nil))

	to, err := bs.LoadAccount(util.Signer2.Address())
	assert.NoError(err)

	txData := &ld.TxData{
		Type:      ld.TypePunish,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      to.id,
	}
	assert.NoError(txData.SyntacticVerify())
	_, err = NewTx(txData.ToTransaction(), true)
	assert.ErrorContains(err, "DeriveSigners error: no signature")

	txData = &ld.TxData{
		Type:      ld.TypePunish,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      to.id,
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx(txData.ToTransaction(), true)
	assert.ErrorContains(err,
		"invalid from, expected GenesisAccount, got 0x44171C37Ff5D7B7bb8dcad5C81f16284A229e641")

	txData = &ld.TxData{
		Type:      ld.TypePunish,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      from.id,
		To:        &to.id,
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx(txData.ToTransaction(), true)
	assert.ErrorContains(err, "invalid to, should be nil")

	txData = &ld.TxData{
		Type:      ld.TypePunish,
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
		Type:      ld.TypePunish,
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
		Type:      ld.TypePunish,
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
		Type:      ld.TypePunish,
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
		Type:      ld.TypePunish,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      from.id,
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx(txData.ToTransaction(), true)
	assert.ErrorContains(err, "nil data id")

	input = ld.TxUpdater{ID: &util.DataIDEmpty}
	assert.NoError(input.SyntacticVerify())
	txData = &ld.TxData{
		Type:      ld.TypePunish,
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

	did := util.DataID{'a', 'b', 'c', 'd', 'e', 'f'}
	input = ld.TxUpdater{ID: &did, Data: []byte(`"Illegal content"`)}
	assert.NoError(input.SyntacticVerify())
	txData = &ld.TxData{
		Type:      ld.TypePunish,
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
	bs.CommitAccounts()
	assert.ErrorContains(itx.Apply(bctx, bs),
		"TxPunish.Apply error: invalid gas, expected 138, got 0")
	bs.CheckoutAccounts()

	tt.Gas = tt.RequiredGas(bctx.FeeConfig().ThresholdGas)
	itx, err = NewTx(tt, true)
	assert.NoError(err)
	bs.CommitAccounts()
	assert.ErrorContains(itx.Apply(bctx, bs),
		"insufficient NativeLDC balance, expected 151800, got 0")
	bs.CheckoutAccounts()

	from.Add(constants.NativeToken, new(big.Int).SetUint64(constants.LDC))
	bs.CommitAccounts()
	assert.ErrorContains(itx.Apply(bctx, bs),
		"LD9svQk6dYkcjZ33L4mZdXJArdPt5vQS7r8 not found")
	bs.CheckoutAccounts()

	di := &ld.DataInfo{
		Version:   1,
		Threshold: 1,
		Keepers:   util.EthIDs{util.Signer2.Address()},
		Data:      []byte(`"test...."`),
	}
	di.KSig, err = util.Signer2.Sign(di.Data)
	assert.NoError(err)
	assert.NoError(di.SyntacticVerify())
	assert.NoError(bs.SaveData(did, di))
	assert.NoError(bs.SavePrevData(did, di))
	assert.NoError(itx.Apply(bctx, bs))

	assert.Equal(tt.Gas*bctx.Price,
		itx.(*TxPunish).ldc.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(tt.Gas*100,
		itx.(*TxPunish).miner.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(constants.LDC-tt.Gas*(bctx.Price+100),
		from.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(uint64(1), from.Nonce())

	di, err = bs.LoadData(did)
	assert.NoError(err)
	assert.Equal(uint64(0), di.Version)
	assert.Equal(util.SignatureEmpty, di.KSig)
	assert.Nil(di.MSig)
	assert.Equal(input.Data, di.Data)

	jsondata, err := itx.MarshalJSON()
	assert.NoError(err)
	// fmt.Println(string(jsondata))
	assert.Equal(`{"type":"TypePunish","chainID":2357,"nonce":0,"gasTip":100,"gasFeeCap":1000,"from":"0xFFfFfFffFFfffFFfFFfFFFFFffFFFffffFfFFFfF","data":{"id":"LD9svQk6dYkcjZ33L4mZdXJArdPt5vQS7r8","data":"Illegal content"},"signatures":["c9f430b760115127737634c92ba2fb544134b00542703d1731239ddefa5ea28d06bc883916b3c8d08802c4109d7fadcaa7feeeb0900ac25b84f927424b8120c701"],"gas":138,"id":"2CfBQuuhuptvM81Hf9v4zZMmhaiQCurzh9Sgzu9YsJDSxF8nYy"}`, string(jsondata))

	assert.NoError(bs.VerifyState())
}
