// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package transactions

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

	ctx := NewMockChainContext()
	cs := ctx.MockChainState()
	token := ld.MustNewToken("$LDC")

	owner := util.Signer1.Address()
	modelKeeper := util.Signer2.Address()

	txData := &ld.TxData{
		Type:      ld.TypeUpdateData,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      owner,
	}
	assert.NoError(txData.SyntacticVerify())
	_, err = NewTx2(txData.ToTransaction())
	assert.ErrorContains(err, "DeriveSigners error: no signature")

	txData = &ld.TxData{
		Type:      ld.TypeUpdateData,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      owner,
		Token:     &token,
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx2(txData.ToTransaction())
	assert.ErrorContains(err, "invalid token, should be nil")

	txData = &ld.TxData{
		Type:      ld.TypeUpdateData,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      owner,
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx2(txData.ToTransaction())
	assert.ErrorContains(err, "invalid data")

	txData = &ld.TxData{
		Type:      ld.TypeUpdateData,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      owner,
		Data:      []byte("你好👋"),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx2(txData.ToTransaction())
	assert.ErrorContains(err, "cbor: unexpected following extraneous data")

	input := &ld.TxUpdater{}
	txData = &ld.TxData{
		Type:      ld.TypeUpdateData,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      owner,
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx2(txData.ToTransaction())
	assert.ErrorContains(err, "invalid data id")

	input = &ld.TxUpdater{ID: &util.DataIDEmpty}
	txData = &ld.TxData{
		Type:      ld.TypeUpdateData,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      owner,
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx2(txData.ToTransaction())
	assert.ErrorContains(err, "invalid data id")

	did := util.DataID{1, 2, 3, 4}
	input = &ld.TxUpdater{ID: &did}
	txData = &ld.TxData{
		Type:      ld.TypeUpdateData,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      owner,
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx2(txData.ToTransaction())
	assert.ErrorContains(err, "invalid data version")

	input = &ld.TxUpdater{ID: &did, Version: 1, Threshold: ld.Uint16Ptr(1)}
	txData = &ld.TxData{
		Type:      ld.TypeUpdateData,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      owner,
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx2(txData.ToTransaction())
	assert.ErrorContains(err, "no keepers, threshold should be nil")

	input = &ld.TxUpdater{ID: &did, Version: 1, Threshold: ld.Uint16Ptr(1),
		Keepers: &util.EthIDs{util.Signer1.Address()}}
	txData = &ld.TxData{
		Type:      ld.TypeUpdateData,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      owner,
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx2(txData.ToTransaction())
	assert.ErrorContains(err, "invalid threshold, should be nil")

	input = &ld.TxUpdater{ID: &did, Version: 1, Approver: &util.EthIDEmpty}
	txData = &ld.TxData{
		Type:      ld.TypeUpdateData,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      owner,
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx2(txData.ToTransaction())
	assert.ErrorContains(err, "invalid approver, should be nil")

	input = &ld.TxUpdater{ID: &did, Version: 1, ApproveList: []ld.TxType{ld.TypeDeleteData}}
	txData = &ld.TxData{
		Type:      ld.TypeUpdateData,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      owner,
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx2(txData.ToTransaction())
	assert.ErrorContains(err, "invalid approveList, should be nil")

	input = &ld.TxUpdater{ID: &did, Version: 1}
	txData = &ld.TxData{
		Type:      ld.TypeUpdateData,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      owner,
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx2(txData.ToTransaction())
	assert.ErrorContains(err, "invalid data")

	input = &ld.TxUpdater{ID: &did, Version: 1,
		Data: []byte(`421`),
	}
	txData = &ld.TxData{
		Type:      ld.TypeUpdateData,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      owner,
		Data:      input.Bytes(),
		To:        &modelKeeper,
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx2(txData.ToTransaction())
	assert.ErrorContains(err, "invalid to, should be nil")

	txData = &ld.TxData{
		Type:      ld.TypeUpdateData,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      owner,
		Data:      input.Bytes(),
		Amount:    new(big.Int).SetUint64(constants.MilliLDC),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx2(txData.ToTransaction())
	assert.ErrorContains(err, "nil to together with amount")

	input = &ld.TxUpdater{ID: &did, Version: 1,
		Data: []byte(`421`),
		To:   &modelKeeper,
	}
	txData = &ld.TxData{
		Type:      ld.TypeUpdateData,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      owner,
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx2(txData.ToTransaction())
	assert.ErrorContains(err,
		"invalid to, expected 0x44171C37Ff5D7B7bb8dcad5C81f16284A229e641, got <nil>")

	input = &ld.TxUpdater{ID: &did, Version: 1,
		Data: []byte(`421`),
		To:   &modelKeeper,
	}
	txData = &ld.TxData{
		Type:      ld.TypeUpdateData,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      owner,
		To:        &constants.GenesisAccount,
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx2(txData.ToTransaction())
	assert.ErrorContains(err,
		"invalid to, expected 0x44171C37Ff5D7B7bb8dcad5C81f16284A229e641, got 0xFFfFfFffFFfffFFfFFfFFFFFffFFFffffFfFFFfF")

	input = &ld.TxUpdater{ID: &did, Version: 1,
		Data: []byte(`421`),
		To:   &modelKeeper,
	}
	txData = &ld.TxData{
		Type:      ld.TypeUpdateData,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      owner,
		To:        &modelKeeper,
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx2(txData.ToTransaction())
	assert.ErrorContains(err, "nil amount")

	input = &ld.TxUpdater{ID: &did, Version: 1,
		Data:   []byte(`421`),
		To:     &modelKeeper,
		Amount: new(big.Int).SetUint64(constants.MilliLDC),
	}
	txData = &ld.TxData{
		Type:      ld.TypeUpdateData,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      owner,
		To:        &modelKeeper,
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx2(txData.ToTransaction())
	assert.ErrorContains(err, "nil amount")

	input = &ld.TxUpdater{ID: &did, Version: 1,
		Data:   []byte(`421`),
		To:     &modelKeeper,
		Amount: new(big.Int).SetUint64(constants.MilliLDC),
	}
	txData = &ld.TxData{
		Type:      ld.TypeUpdateData,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      owner,
		To:        &modelKeeper,
		Amount:    new(big.Int).SetUint64(constants.MilliLDC),
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	tt := txData.ToTransaction()
	tt.Timestamp = 10
	_, err = NewTx2(tt)
	assert.ErrorContains(err, "data expired")

	input = &ld.TxUpdater{ID: &did, Version: 1,
		Data:   []byte(`421`),
		To:     &modelKeeper,
		Amount: new(big.Int).SetUint64(constants.MilliLDC),
		Expire: 10,
	}
	txData = &ld.TxData{
		Type:      ld.TypeUpdateData,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      owner,
		To:        &modelKeeper,
		Amount:    new(big.Int).SetUint64(constants.MilliLDC),
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	tt = txData.ToTransaction()
	tt.Timestamp = 10
	_, err = NewTx2(tt)
	assert.ErrorContains(err,
		"invalid exSignatures, Transaction.ExSigners error: DeriveSigners error: no signature")

	assert.NoError(txData.ExSignWith(util.Signer1))
	tt = txData.ToTransaction()
	tt.Timestamp = 10
	itx, err := NewTx2(tt)
	assert.NoError(err)

	cs.CommitAccounts()
	assert.ErrorContains(itx.Apply(ctx, cs),
		"insufficient NativeLDC balance, expected 2635700, got 0")
	cs.CheckoutAccounts()

	ownerAcc := cs.MustAccount(owner)
	ownerAcc.Add(constants.NativeToken, new(big.Int).SetUint64(constants.LDC))
	cs.CommitAccounts()
	assert.ErrorContains(itx.Apply(ctx, cs),
		"SkB92DD9M2yeCadw22VbnxfV6b7W5YEnnLRs6fKivk6wh2Zy not found")
	cs.CheckoutAccounts()

	di := &ld.DataInfo{
		ModelID:   ld.RawModelID,
		Version:   2,
		Threshold: 1,
		Keepers:   util.EthIDs{util.Signer1.Address()},
		Payload:   []byte(`42`),
		ID:        did,
	}
	assert.NoError(di.SyntacticVerify())
	cs.SaveData(di)
	cs.CommitAccounts()
	assert.ErrorContains(itx.Apply(ctx, cs),
		"invalid version, expected 2, got 1")
	cs.CheckoutAccounts()

	di = &ld.DataInfo{
		ModelID:   ld.RawModelID,
		Version:   1,
		Threshold: 1,
		Keepers:   util.EthIDs{util.Signer1.Address()},
		Payload:   []byte(`42`),
		ID:        did,
	}
	assert.NoError(di.SyntacticVerify())
	cs.SaveData(di)
	cs.CommitAccounts()
	assert.ErrorContains(itx.Apply(ctx, cs),
		"invalid to, should be nil")
	cs.CheckoutAccounts()

	input = &ld.TxUpdater{ID: &did, Version: 1,
		Data: []byte(`421`),
	}
	txData = &ld.TxData{
		Type:      ld.TypeUpdateData,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      owner,
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	tt = txData.ToTransaction()
	itx, err = NewTx2(tt)
	assert.NoError(err)
	assert.NoError(itx.Apply(ctx, cs))

	ownerGas := tt.Gas()
	assert.Equal(ownerGas*ctx.Price,
		itx.(*TxUpdateData).ldc.Balance().Uint64())
	assert.Equal(ownerGas*100,
		itx.(*TxUpdateData).miner.Balance().Uint64())
	assert.Equal(constants.LDC-ownerGas*(ctx.Price+100),
		ownerAcc.Balance().Uint64())
	assert.Equal(uint64(1), ownerAcc.Nonce())

	di2, err := cs.LoadData(di.ID)
	assert.NoError(err)
	assert.Equal(uint64(2), di2.Version)
	assert.Equal([]byte(`42`), []byte(di.Payload))
	assert.Equal([]byte(`421`), []byte(di2.Payload))
	assert.Equal(cs.PDC[di.ID], di.Bytes())

	jsondata, err := itx.MarshalJSON()
	assert.NoError(err)
	// fmt.Println(string(jsondata))
	assert.Equal(`{"type":"TypeUpdateData","chainID":2357,"nonce":0,"gasTip":100,"gasFeeCap":1000,"from":"0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC","data":{"id":"SkB92DD9M2yeCadw22VbnxfV6b7W5YEnnLRs6fKivk6wh2Zy","version":1,"data":421},"signatures":["05c1688f1aa64e7c9acaa365a1c60449a782cf9946acaed98f65e1881db9471141d448a3402978c325608e34ecd2ca0d1f13e34aea7e493d1059381131d0ddc501"],"id":"zJQCouuJJLGUVMEP8EdyFXwALbeHYE1nzMKWBTtRUJgko5MeC"}`, string(jsondata))

	assert.NoError(cs.VerifyState())
}

func TestTxUpdateCBORData(t *testing.T) {
	assert := assert.New(t)

	// SyntacticVerify
	tx := &TxUpdateData{}
	assert.ErrorContains(tx.SyntacticVerify(), "nil pointer")
	_, err := tx.MarshalJSON()
	assert.NoError(err)

	ctx := NewMockChainContext()
	cs := ctx.MockChainState()

	owner := util.Signer1.Address()
	ownerAcc := cs.MustAccount(owner)
	assert.NoError(ownerAcc.Add(constants.NativeToken, new(big.Int).SetUint64(constants.LDC)))

	type cborData struct {
		Name   string `cbor:"na"`
		Nonces []int  `cbor:"no"`
	}

	data, err := util.MarshalCBOR(&cborData{Name: "test", Nonces: []int{1, 2, 3}})
	assert.NoError(err)

	di := &ld.DataInfo{
		ModelID:   ld.CBORModelID,
		Version:   2,
		Threshold: 1,
		Keepers:   util.EthIDs{util.Signer1.Address()},
		Payload:   data,
		ID:        util.DataID{1, 2, 3, 4},
	}
	assert.NoError(di.SyntacticVerify())
	cs.SaveData(di)

	input := &ld.TxUpdater{ID: &di.ID, Version: 2,
		Data: di.Payload[2:],
	}
	txData := &ld.TxData{
		Type:      ld.TypeUpdateData,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      owner,
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	tt := txData.ToTransaction()
	itx, err := NewTx2(tt)
	assert.NoError(err)
	cs.CommitAccounts()
	assert.ErrorContains(itx.Apply(ctx, cs), "invalid CBOR patch")
	cs.CheckoutAccounts()

	patch := cborpatch.Patch{
		{Op: "add", Path: "/no/-", Value: util.MustMarshalCBOR(4)},
	}
	patchdata := util.MustMarshalCBOR(patch)

	input = &ld.TxUpdater{ID: &di.ID, Version: 2,
		Data: patchdata,
	}
	newData, err := patch.Apply(di.Payload)
	assert.NoError(err)
	txData = &ld.TxData{
		Type:      ld.TypeUpdateData,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      owner,
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	tt = txData.ToTransaction()
	itx, err = NewTx2(tt)
	assert.NoError(err)
	assert.NoError(itx.Apply(ctx, cs))

	ownerGas := tt.Gas()
	assert.Equal(ownerGas*ctx.Price,
		itx.(*TxUpdateData).ldc.Balance().Uint64())
	assert.Equal(ownerGas*100,
		itx.(*TxUpdateData).miner.Balance().Uint64())
	assert.Equal(constants.LDC-ownerGas*(ctx.Price+100),
		itx.(*TxUpdateData).from.Balance().Uint64())
	assert.Equal(uint64(1), itx.(*TxUpdateData).from.Nonce())

	di2, err := cs.LoadData(di.ID)
	assert.NoError(err)
	assert.Equal(uint64(3), di2.Version)
	assert.NotEqual(newData, []byte(di.Payload))
	assert.Equal(newData, []byte(di2.Payload))
	assert.Equal(cs.PDC[di.ID], di.Bytes())

	var nc cborData
	assert.NoError(util.UnmarshalCBOR(di2.Payload, &nc))
	assert.Equal([]int{1, 2, 3, 4}, nc.Nonces)

	jsondata, err := itx.MarshalJSON()
	assert.NoError(err)
	// fmt.Println(string(jsondata))
	assert.Equal(`{"type":"TypeUpdateData","chainID":2357,"nonce":0,"gasTip":100,"gasFeeCap":1000,"from":"0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC","data":{"id":"SkB92DD9M2yeCadw22VbnxfV6b7W5YEnnLRs6fKivk6wh2Zy","version":2,"data":"0x81a3626f70636164646470617468652f6e6f2f2d6576616c7565040f9dc5ca"},"signatures":["7f10eae5b8aed7fa2610e9a26300b9c6a735a551cf5f4b36be8304a9604eaa58773dcb0893ecdb440fd91e0e39ff1c0c3eea36ff4cc91b5630d1f3e1239938fb00"],"id":"2tfc2PMdDFrzt4TLwwB3A7BtazMZfS4huM3FVPtmtGD6uVcvNL"}`, string(jsondata))

	assert.NoError(cs.VerifyState())
}

func TestTxUpdateJSONData(t *testing.T) {
	assert := assert.New(t)

	// SyntacticVerify
	tx := &TxUpdateData{}
	assert.ErrorContains(tx.SyntacticVerify(), "nil pointer")
	_, err := tx.MarshalJSON()
	assert.NoError(err)

	ctx := NewMockChainContext()
	cs := ctx.MockChainState()

	owner := util.Signer1.Address()
	ownerAcc := cs.MustAccount(owner)
	assert.NoError(ownerAcc.Add(constants.NativeToken, new(big.Int).SetUint64(constants.LDC)))

	data := []byte(`{"name":"test","nonces":[1,2,3]}`)
	di := &ld.DataInfo{
		ModelID:   ld.JSONModelID,
		Version:   2,
		Threshold: 1,
		Keepers:   util.EthIDs{util.Signer1.Address()},
		Payload:   data,
		ID:        util.DataID{1, 2, 3, 4},
	}
	assert.NoError(di.SyntacticVerify())
	cs.SaveData(di)

	input := &ld.TxUpdater{ID: &di.ID, Version: 2,
		Data: []byte(`{}`),
	}
	txData := &ld.TxData{
		Type:      ld.TypeUpdateData,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      owner,
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	tt := txData.ToTransaction()
	itx, err := NewTx2(tt)
	assert.NoError(err)
	cs.CommitAccounts()
	assert.ErrorContains(itx.Apply(ctx, cs), "invalid JSON patch")
	cs.CheckoutAccounts()

	input = &ld.TxUpdater{ID: &di.ID, Version: 2,
		Data: []byte(`[{"op": "replace", "path": "/name", "value": "Tester"}]`),
		SigClaims: &ld.SigClaims{
			Issuer:     util.DataID{1, 2, 3, 4},
			Subject:    di.ID,
			Audience:   di.ModelID,
			Expiration: 100,
			IssuedAt:   1,
			CWTID:      util.HashFromData([]byte(`{"name":"Tester","nonces":[1,2,3]}`)),
		},
	}
	sig, err := util.Signer2.Sign(input.SigClaims.Bytes())
	assert.NoError(err)
	input.TypedSig = sig.Typed()

	txData = &ld.TxData{
		Type:      ld.TypeUpdateData,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      owner,
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	tt = txData.ToTransaction()
	itx, err = NewTx2(tt)
	assert.NoError(err)
	assert.NoError(itx.Apply(ctx, cs))

	ownerGas := tt.Gas()
	assert.Equal(ownerGas*ctx.Price,
		itx.(*TxUpdateData).ldc.Balance().Uint64())
	assert.Equal(ownerGas*100,
		itx.(*TxUpdateData).miner.Balance().Uint64())
	assert.Equal(constants.LDC-ownerGas*(ctx.Price+100),
		ownerAcc.Balance().Uint64())
	assert.Equal(uint64(1), ownerAcc.Nonce())

	di2, err := cs.LoadData(di.ID)
	assert.NoError(err)
	assert.Equal(uint64(3), di2.Version)
	assert.Equal([]byte(`{"name":"test","nonces":[1,2,3]}`), []byte(di.Payload))
	assert.Equal([]byte(`{"name":"Tester","nonces":[1,2,3]}`), []byte(di2.Payload))
	assert.Equal(cs.PDC[di.ID], di.Bytes())

	assert.NoError(di.ValidSigClaims())
	signer, err := di.Signer()
	assert.ErrorContains(err, "invalid typed signature length, expected 66, got 0")
	assert.Equal(util.EthIDEmpty, signer)

	assert.NoError(di2.ValidSigClaims())
	signer, err = di2.Signer()
	assert.NoError(err)
	assert.Equal(util.Signer2.Address(), signer)

	jsondata, err := itx.MarshalJSON()
	assert.NoError(err)
	// fmt.Println(string(jsondata))
	assert.Equal(`{"type":"TypeUpdateData","chainID":2357,"nonce":0,"gasTip":100,"gasFeeCap":1000,"from":"0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC","data":{"id":"SkB92DD9M2yeCadw22VbnxfV6b7W5YEnnLRs6fKivk6wh2Zy","version":2,"sigClaims":{"iss":"SkB92DD9M2yeCadw22VbnxfV6b7W5YEnnLRs6fKivk6wh2Zy","sub":"SkB92DD9M2yeCadw22VbnxfV6b7W5YEnnLRs6fKivk6wh2Zy","aud":"1111111111111111111L17Xp3","exp":100,"nbf":0,"iat":1,"cti":"Do6oUfVMCkZKbUZ4FgNmjFQfNVjBYWUDJNuq1GVpgcuBFmibS"},"typedSig":"0x00e64cc098eaa126aeee90a85936baf2e5e7ea4ae2b6eb9f5fa537939097c08ff7286b246640c92a4473ac30ce8fb02e175134abf5c16415c279e901743ccf519a00a71dcdb3","data":[{"op":"replace","path":"/name","value":"Tester"}]},"signatures":["790f55433ee2f2e7938bf3edcac127ab035778f3f7aa8bfe13a8578479324fa130060b4bebf88c461f294f335559e12f6938c6f3189ff8c531f5d5026784d9ca01"],"id":"b6vRYF1y6vyYLetyLRDyejKoHHczAiXmUaS4aXQuEVdi4v4eK"}`, string(jsondata))

	assert.NoError(cs.VerifyState())
}
