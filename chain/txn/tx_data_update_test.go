// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package txn

import (
	"math/big"
	"testing"

	"github.com/ldclabs/ldvm/ids"
	"github.com/ldclabs/ldvm/ld"
	"github.com/ldclabs/ldvm/ld/service"
	"github.com/ldclabs/ldvm/signer"
	"github.com/ldclabs/ldvm/unit"
	"github.com/ldclabs/ldvm/util/encoding"

	cborpatch "github.com/ldclabs/cbor-patch"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTxUpdateData(t *testing.T) {
	assert := assert.New(t)

	// SyntacticVerify
	tx := &TxUpdateData{}
	assert.ErrorContains(tx.SyntacticVerify(), "nil pointer")
	_, err := tx.MarshalJSON()
	require.NoError(t, err)

	ctx := NewMockChainContext()
	cs := ctx.MockChainState()
	token := ld.MustNewToken("$LDC")

	owner := signer.Signer1.Key().Address()
	modelKeeper := signer.Signer2.Key().Address()

	ltx := &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeUpdateData,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      owner,
	}}
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "no signatures")

	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeUpdateData,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      owner,
		Token:     token.Ptr(),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "invalid token, should be nil")

	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeUpdateData,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      owner,
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "invalid data")

	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeUpdateData,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      owner,
		Data:      []byte("你好👋"),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "extraneous data")

	input := &ld.TxUpdater{}
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeUpdateData,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      owner,
		Data:      input.Bytes(),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "invalid data id")

	input = &ld.TxUpdater{ID: ids.EmptyDataID.Ptr()}
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeUpdateData,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      owner,
		Data:      input.Bytes(),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "invalid data id")

	did := ids.DataID{1, 2, 3, 4}
	input = &ld.TxUpdater{ID: &did}
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeUpdateData,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      owner,
		Data:      input.Bytes(),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "invalid data version")

	input = &ld.TxUpdater{ID: &did, Version: 1, Threshold: ld.Uint16Ptr(1)}
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeUpdateData,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      owner,
		Data:      input.Bytes(),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "no keepers, threshold should be nil")

	input = &ld.TxUpdater{ID: &did, Version: 1, Threshold: ld.Uint16Ptr(1),
		Keepers: &signer.Keys{signer.Signer1.Key()}}
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeUpdateData,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      owner,
		Data:      input.Bytes(),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "invalid threshold, should be nil")

	input = &ld.TxUpdater{ID: &did, Version: 1, Approver: &signer.Key{}}
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeUpdateData,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      owner,
		Data:      input.Bytes(),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "invalid approver, should be nil")

	input = &ld.TxUpdater{ID: &did, Version: 1, ApproveList: &ld.TxTypes{ld.TypeDeleteData}}
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeUpdateData,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      owner,
		Data:      input.Bytes(),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "invalid approveList, should be nil")

	input = &ld.TxUpdater{ID: &did, Version: 1}
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeUpdateData,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      owner,
		Data:      input.Bytes(),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "invalid data")

	input = &ld.TxUpdater{ID: &did, Version: 1,
		Data: []byte(`421`),
	}
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeUpdateData,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      owner,
		Data:      input.Bytes(),
		To:        &modelKeeper,
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "invalid to, should be nil")

	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeUpdateData,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      owner,
		Data:      input.Bytes(),
		Amount:    new(big.Int).SetUint64(unit.MilliLDC),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.ErrorContains(ltx.SyntacticVerify(), "nil \"to\" together with amount")

	input = &ld.TxUpdater{ID: &did, Version: 1,
		Data: []byte(`421`),
		To:   &modelKeeper,
	}
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeUpdateData,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      owner,
		Data:      input.Bytes(),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err,
		"invalid to, expected 0x44171C37Ff5D7B7bb8Dcad5C81f16284A229E641, got <nil>")

	input = &ld.TxUpdater{ID: &did, Version: 1,
		Data: []byte(`421`),
		To:   &modelKeeper,
	}
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeUpdateData,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      owner,
		To:        ids.GenesisAccount.Ptr(),
		Data:      input.Bytes(),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err,
		"invalid to, expected 0x44171C37Ff5D7B7bb8Dcad5C81f16284A229E641, got 0xFFfFFFfFfffFFfFFffFFFfFfFffFFFfffFfFFFff")

	input = &ld.TxUpdater{ID: &did, Version: 1,
		Data: []byte(`421`),
		To:   &modelKeeper,
	}
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeUpdateData,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      owner,
		To:        &modelKeeper,
		Data:      input.Bytes(),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "nil amount")

	input = &ld.TxUpdater{ID: &did, Version: 1,
		Data:   []byte(`421`),
		To:     &modelKeeper,
		Amount: new(big.Int).SetUint64(unit.MilliLDC),
	}
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeUpdateData,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      owner,
		To:        &modelKeeper,
		Data:      input.Bytes(),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "nil amount")

	input = &ld.TxUpdater{ID: &did, Version: 1,
		Data:   []byte(`421`),
		To:     &modelKeeper,
		Amount: new(big.Int).SetUint64(unit.MilliLDC),
	}
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeUpdateData,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      owner,
		To:        &modelKeeper,
		Amount:    new(big.Int).SetUint64(unit.MilliLDC),
		Data:      input.Bytes(),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())

	ltx.Timestamp = 10
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "data expired")

	input = &ld.TxUpdater{ID: &did, Version: 1,
		Data:   []byte(`421`),
		To:     &modelKeeper,
		Amount: new(big.Int).SetUint64(unit.MilliLDC),
		Expire: 10,
	}
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeUpdateData,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      owner,
		To:        &modelKeeper,
		Amount:    new(big.Int).SetUint64(unit.MilliLDC),
		Data:      input.Bytes(),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())

	ltx.Timestamp = 10
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "no exSignatures")

	assert.NoError(ltx.ExSignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())

	ltx.Timestamp = 10
	itx, err := NewTx(ltx)
	require.NoError(t, err)

	cs.CommitAccounts()
	assert.ErrorContains(itx.Apply(ctx, cs),
		"insufficient NativeLDC balance, expected 2662100, got 0")
	cs.CheckoutAccounts()

	ownerAcc := cs.MustAccount(owner)
	ownerAcc.Add(ids.NativeToken, new(big.Int).SetUint64(unit.LDC*2))
	cs.CommitAccounts()
	assert.ErrorContains(itx.Apply(ctx, cs),
		"AQIDBAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAACs148t not found")
	cs.CheckoutAccounts()

	di := &ld.DataInfo{
		ModelID:   ld.RawModelID,
		Version:   2,
		Threshold: 1,
		Keepers:   signer.Keys{signer.Signer1.Key()},
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
		Keepers:   signer.Keys{signer.Signer1.Key()},
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
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeUpdateData,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      owner,
		Data:      input.Bytes(),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())

	itx, err = NewTx(ltx)
	require.NoError(t, err)
	assert.NoError(itx.Apply(ctx, cs))

	ownerGas := ltx.Gas()
	assert.Equal(ownerGas*ctx.Price,
		itx.(*TxUpdateData).ldc.Balance().Uint64())
	assert.Equal(ownerGas*100,
		itx.(*TxUpdateData).miner.Balance().Uint64())
	assert.Equal(unit.LDC-ownerGas*(ctx.Price+100),
		ownerAcc.Balance().Uint64())
	assert.Equal(unit.LDC*2-ownerGas*(ctx.Price+100),
		ownerAcc.BalanceOfAll(ids.NativeToken).Uint64())
	assert.Equal(uint64(1), ownerAcc.Nonce())

	di2, err := cs.LoadData(di.ID)
	require.NoError(t, err)
	assert.Equal(uint64(2), di2.Version)
	assert.Equal([]byte(`42`), []byte(di.Payload))
	assert.Equal([]byte(`421`), []byte(di2.Payload))
	assert.Equal(cs.PDC[di.ID], di.Bytes())

	jsondata, err := itx.MarshalJSON()
	require.NoError(t, err)
	// fmt.Println(string(jsondata))
	assert.Equal(`{"tx":{"type":"TypeUpdateData","chainID":2357,"nonce":0,"gasTip":100,"gasFeeCap":1000,"from":"0x8db97c7cECe249C2b98bdc0226cc4C2A57bF52fc","data":{"id":"AQIDBAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAACs148t","version":1,"data":421}},"sigs":["BcFojxqmTnyayqNlocYESaeCz5lGrK7Zj2XhiB25RxFB1EijQCl4wyVgjjTs0soNHxPjSup-ST0QWTgRMdDdxQGM1L5d"],"id":"XJGFQoGdpV3rjeWl_1uwaMDim_Sw2uyQiGnrvWmDk40828u1"}`, string(jsondata))

	assert.NoError(cs.VerifyState())
}

func TestTxUpdateCBORData(t *testing.T) {
	assert := assert.New(t)

	// SyntacticVerify
	tx := &TxUpdateData{}
	assert.ErrorContains(tx.SyntacticVerify(), "nil pointer")
	_, err := tx.MarshalJSON()
	require.NoError(t, err)

	ctx := NewMockChainContext()
	cs := ctx.MockChainState()

	owner := signer.Signer1.Key().Address()
	ownerAcc := cs.MustAccount(owner)
	assert.NoError(ownerAcc.Add(ids.NativeToken, new(big.Int).SetUint64(unit.LDC)))

	type cborData struct {
		Name   string `cbor:"na"`
		Nonces []int  `cbor:"no"`
	}

	data, err := encoding.MarshalCBOR(&cborData{Name: "test", Nonces: []int{1, 2, 3}})
	require.NoError(t, err)

	di := &ld.DataInfo{
		ModelID:   ld.CBORModelID,
		Version:   2,
		Threshold: 1,
		Keepers:   signer.Keys{signer.Signer1.Key()},
		Payload:   data,
		ID:        ids.DataID{1, 2, 3, 4},
	}
	assert.NoError(di.SyntacticVerify())
	cs.SaveData(di)

	input := &ld.TxUpdater{ID: &di.ID, Version: 2,
		Data: di.Payload[2:],
	}
	ltx := &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeUpdateData,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      owner,
		Data:      input.Bytes(),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())

	itx, err := NewTx(ltx)
	require.NoError(t, err)
	cs.CommitAccounts()
	assert.ErrorContains(itx.Apply(ctx, cs), "invalid CBOR patch")
	cs.CheckoutAccounts()

	patch := cborpatch.Patch{{
		Op:    cborpatch.OpAdd,
		Path:  cborpatch.PathMustFrom("no", "-"),
		Value: encoding.MustMarshalCBOR(4),
	}}
	patchdata := encoding.MustMarshalCBOR(patch)

	input = &ld.TxUpdater{ID: &di.ID, Version: 2,
		Data: patchdata,
	}
	newData, err := patch.Apply(di.Payload)
	require.NoError(t, err)
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeUpdateData,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      owner,
		Data:      input.Bytes(),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())

	itx, err = NewTx(ltx)
	require.NoError(t, err)
	assert.NoError(itx.Apply(ctx, cs))

	ownerGas := ltx.Gas()
	assert.Equal(ownerGas*ctx.Price,
		itx.(*TxUpdateData).ldc.Balance().Uint64())
	assert.Equal(ownerGas*100,
		itx.(*TxUpdateData).miner.Balance().Uint64())
	assert.Equal(unit.LDC-ownerGas*(ctx.Price+100),
		itx.(*TxUpdateData).from.BalanceOfAll(ids.NativeToken).Uint64())
	assert.Equal(uint64(1), itx.(*TxUpdateData).from.Nonce())

	di2, err := cs.LoadData(di.ID)
	require.NoError(t, err)
	assert.Equal(uint64(3), di2.Version)
	assert.NotEqual(newData, []byte(di.Payload))
	assert.Equal(newData, []byte(di2.Payload))
	assert.Equal(cs.PDC[di.ID], di.Bytes())

	var nc cborData
	assert.NoError(encoding.UnmarshalCBOR(di2.Payload, &nc))
	assert.Equal("test", nc.Name)
	assert.Equal([]int{1, 2, 3, 4}, nc.Nonces)

	jsondata, err := itx.MarshalJSON()
	require.NoError(t, err)
	// fmt.Println(string(jsondata))
	assert.Equal(`{"tx":{"type":"TypeUpdateData","chainID":2357,"nonce":0,"gasTip":100,"gasFeeCap":1000,"from":"0x8db97c7cECe249C2b98bdc0226cc4C2A57bF52fc","data":{"id":"AQIDBAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAACs148t","version":2,"data":"gaMBAQOCYm5vYS0EBIXv9uI"}},"sigs":["wgoyFE9DndlKYwR13dYCSjgCgINtLr8zstlPyVwl9BgyWYGncVjSNeVUV2IYfhVOZBYKd9IDyxIWb_mF4aFhKQAMD6e5"],"id":"CNd8SqAKQzs2e04wFMMvZs6qgvO6kJqUBqiRFRK-x3lPap8-"}`, string(jsondata))

	assert.NoError(cs.VerifyState())
}

func TestTxUpdateJSONData(t *testing.T) {
	assert := assert.New(t)

	// SyntacticVerify
	tx := &TxUpdateData{}
	assert.ErrorContains(tx.SyntacticVerify(), "nil pointer")
	_, err := tx.MarshalJSON()
	require.NoError(t, err)

	ctx := NewMockChainContext()
	cs := ctx.MockChainState()

	owner := signer.Signer1.Key().Address()
	ownerAcc := cs.MustAccount(owner)
	assert.NoError(ownerAcc.Add(ids.NativeToken, new(big.Int).SetUint64(unit.LDC)))

	data := []byte(`{"name":"test","nonces":[1,2,3]}`)
	di := &ld.DataInfo{
		ModelID:   ld.JSONModelID,
		Version:   2,
		Threshold: 1,
		Keepers:   signer.Keys{signer.Signer1.Key()},
		Payload:   data,
		ID:        ids.DataID{1, 2, 3, 4},
	}
	assert.NoError(di.SyntacticVerify())
	cs.SaveData(di)

	input := &ld.TxUpdater{ID: &di.ID, Version: 2,
		Data: []byte(`{}`),
	}
	ltx := &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeUpdateData,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      owner,
		Data:      input.Bytes(),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())

	itx, err := NewTx(ltx)
	require.NoError(t, err)
	cs.CommitAccounts()
	assert.ErrorContains(itx.Apply(ctx, cs), "invalid JSON patch")
	cs.CheckoutAccounts()

	input = &ld.TxUpdater{ID: &di.ID, Version: 2,
		Data: []byte(`[{"op": "replace", "path": "/name", "value": "Tester"}]`),
		SigClaims: &ld.SigClaims{
			Issuer:     ids.DataID{1, 2, 3, 4},
			Subject:    di.ID,
			Audience:   di.ModelID,
			Expiration: 100,
			IssuedAt:   1,
			CWTID:      ids.ID32FromData([]byte(`{"name":"Tester","nonces":[1,2,3]}`)),
		},
	}
	input.Sig = signer.Signer2.MustSignData(input.SigClaims.Bytes()).Ptr()

	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeUpdateData,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      owner,
		Data:      input.Bytes(),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	itx, err = NewTx(ltx)
	require.NoError(t, err)
	assert.NoError(itx.Apply(ctx, cs))

	ownerGas := ltx.Gas()
	assert.Equal(ownerGas*ctx.Price,
		itx.(*TxUpdateData).ldc.Balance().Uint64())
	assert.Equal(ownerGas*100,
		itx.(*TxUpdateData).miner.Balance().Uint64())
	assert.Equal(unit.LDC-ownerGas*(ctx.Price+100),
		ownerAcc.BalanceOfAll(ids.NativeToken).Uint64())
	assert.Equal(uint64(1), ownerAcc.Nonce())

	di2, err := cs.LoadData(di.ID)
	require.NoError(t, err)
	assert.Equal(uint64(3), di2.Version)
	assert.Equal([]byte(`{"name":"test","nonces":[1,2,3]}`), []byte(di.Payload))
	assert.Equal([]byte(`{"name":"Tester","nonces":[1,2,3]}`), []byte(di2.Payload))
	assert.Equal(cs.PDC[di.ID], di.Bytes())

	assert.NoError(di.ValidSigClaims())
	assert.NoError(di2.ValidSigClaims())

	jsondata, err := itx.MarshalJSON()
	require.NoError(t, err)
	// fmt.Println(string(jsondata))
	assert.Equal(`{"tx":{"type":"TypeUpdateData","chainID":2357,"nonce":0,"gasTip":100,"gasFeeCap":1000,"from":"0x8db97c7cECe249C2b98bdc0226cc4C2A57bF52fc","data":{"id":"AQIDBAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAACs148t","version":2,"sigClaims":{"iss":"AQIDBAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAACs148t","sub":"AQIDBAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAACs148t","aud":"AAAAAAAAAAAAAAAAAAAAAAAAAALZFhrw","exp":100,"nbf":0,"iat":1,"cti":"HQ1ebnZXNJoRQij56bSc5UqffxMktgk7X0sc4fsuAo3laabe"},"sig":"5kzAmOqhJq7ukKhZNrry5efqSuK2659fpTeTkJfAj_coayRmQMkqRHOsMM6PsC4XUTSr9cFkFcJ56QF0PM9RmgAIdq69","data":[{"op":"replace","path":"/name","value":"Tester"}]}},"sigs":["WFThCMEtoY-jGj-foQPFlmnWmwwcKxLmPD-DO_9fNqltJMTcc_Nx5aTAUIDg2GF58t5FdPhDLKA9RzjbEp-lMgGcRJWF"],"id":"kXLbTFWpwFVlCEOCjcNBREwx0pwJDhF6MM-lgHOsUp3HVkRD"}`, string(jsondata))

	assert.NoError(cs.VerifyState())
}

func TestTxUpdateNameServiceData(t *testing.T) {
	assert := assert.New(t)

	ctx := NewMockChainContext()
	cs := ctx.MockChainState()

	sender := signer.Signer1.Key().Address()
	recipient := signer.Signer2.Key().Address()

	nm, err := service.NameModel()
	require.NoError(t, err)
	mi := &ld.ModelInfo{
		Name:      nm.Name(),
		Threshold: 1,
		Keepers:   signer.Keys{signer.Signer2.Key()},
		Schema:    nm.Schema(),
		ID:        ctx.ChainConfig().NameServiceID,
	}

	name := &service.Name{
		Name:       "ldc.to.",
		Records:    []string{"ldc.to. IN A 10.0.0.1"},
		Extensions: service.Extensions{},
	}
	assert.NoError(name.SyntacticVerify())

	input := &ld.TxUpdater{
		ModelID:   &mi.ID,
		Version:   1,
		Threshold: ld.Uint16Ptr(1),
		Keepers:   &signer.Keys{signer.Signer1.Key()},
		Data:      name.Bytes(),
		To:        recipient.Ptr(),
		Expire:    100,
		Amount:    new(big.Int).SetUint64(unit.MilliLDC),
	}
	assert.NoError(input.SyntacticVerify())
	ltx := &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeCreateData,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      sender,
		To:        recipient.Ptr(),
		Amount:    new(big.Int).SetUint64(unit.MilliLDC),
		Data:      input.Bytes(),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.ExSignWith(signer.Signer2))
	assert.NoError(ltx.SyntacticVerify())

	senderAcc := cs.MustAccount(sender)
	assert.NoError(senderAcc.Add(ids.NativeToken, new(big.Int).SetUint64(unit.LDC*2)))
	assert.NoError(cs.SaveModel(mi))

	ltx.Timestamp = 10
	itx, err := NewTx(ltx)
	require.NoError(t, err)

	_, err = cs.LoadDataByName("ldc.to.")
	assert.ErrorContains(err, `"ldc.to." not found`)
	assert.NoError(itx.Apply(ctx, cs))
	di, err := cs.LoadDataByName("ldc.to.")
	require.NoError(t, err)
	assert.Equal(mi.ID, di.ModelID)

	patchDoc := cborpatch.Patch{
		{Op: cborpatch.OpReplace, Path: cborpatch.PathMustFrom("n"), Value: encoding.MustMarshalCBOR("ld.to.")},
	}
	input = &ld.TxUpdater{ID: &di.ID, Version: 1,
		Data:   encoding.MustMarshalCBOR(patchDoc),
		To:     recipient.Ptr(),
		Amount: new(big.Int).SetUint64(unit.MilliLDC),
	}

	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeUpdateData,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     1,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      sender,
		To:        recipient.Ptr(),
		Amount:    new(big.Int).SetUint64(unit.MilliLDC),
		Data:      input.Bytes(),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.ExSignWith(signer.Signer2))
	assert.NoError(ltx.SyntacticVerify())
	itx, err = NewTx(ltx)
	require.NoError(t, err)

	cs.CommitAccounts()
	assert.ErrorContains(itx.Apply(ctx, cs),
		`can't update name, expected "ldc.to.", got "ld.to."`)
	cs.CheckoutAccounts()

	assert.NoError(cs.VerifyState())
}
