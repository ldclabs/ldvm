// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package txn

import (
	"math/big"
	"testing"

	cborpatch "github.com/ldclabs/cbor-patch"
	"github.com/ldclabs/ldvm/ids"
	"github.com/ldclabs/ldvm/ld"
	"github.com/ldclabs/ldvm/ld/service"
	"github.com/ldclabs/ldvm/signer"
	"github.com/ldclabs/ldvm/unit"
	"github.com/ldclabs/ldvm/util/encoding"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTxUpgradeData(t *testing.T) {
	assert := assert.New(t)

	// SyntacticVerify
	tx := &TxUpgradeData{}
	assert.ErrorContains(tx.SyntacticVerify(), "nil pointer")
	_, err := tx.MarshalJSON()
	require.NoError(t, err)

	ctx := NewMockChainContext()
	cs := ctx.MockChainState()
	token := ld.MustNewToken("$LDC")

	owner := signer.Signer1.Key().Address()
	modelKeeper := signer.Signer2.Key().Address()

	ltx := &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeUpgradeData,
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
		Type:      ld.TypeUpgradeData,
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
		Type:      ld.TypeUpgradeData,
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
		Type:      ld.TypeUpgradeData,
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
	assert.ErrorContains(err, "cbor: unexpected following extraneous data")

	input := &ld.TxUpdater{}
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeUpgradeData,
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
		Type:      ld.TypeUpgradeData,
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
		Type:      ld.TypeUpgradeData,
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
	assert.ErrorContains(err, "invalid model id")

	input = &ld.TxUpdater{
		ID:      &did,
		ModelID: &ld.RawModelID,
	}
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeUpgradeData,
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
	assert.ErrorContains(err, "invalid model id")

	input = &ld.TxUpdater{
		ID:      &did,
		ModelID: &ld.CBORModelID,
	}
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeUpgradeData,
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
	assert.ErrorContains(err, "invalid model id")

	input = &ld.TxUpdater{
		ID:      &did,
		ModelID: &ld.JSONModelID,
	}
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeUpgradeData,
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
	assert.ErrorContains(err, "invalid model id")

	modelID := ids.ModelID{1, 2, 3, 4}
	input = &ld.TxUpdater{ID: &did, ModelID: &modelID}
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeUpgradeData,
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

	input = &ld.TxUpdater{ID: &did, ModelID: &modelID,
		Version: 1, Threshold: ld.Uint16Ptr(1)}
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeUpgradeData,
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

	input = &ld.TxUpdater{ID: &did, ModelID: &modelID, Version: 1,
		Threshold: ld.Uint16Ptr(1),
		Keepers:   &signer.Keys{signer.Signer1.Key()}}
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeUpgradeData,
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

	input = &ld.TxUpdater{ID: &did, ModelID: &modelID, Version: 1,
		Approver: &signer.Key{}}
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeUpgradeData,
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

	input = &ld.TxUpdater{ID: &did, ModelID: &modelID,
		Version: 1, ApproveList: &ld.TxTypes{ld.TypeDeleteData}}
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeUpgradeData,
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

	input = &ld.TxUpdater{ID: &did, ModelID: &modelID, Version: 1}
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeUpgradeData,
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

	input = &ld.TxUpdater{ID: &did, ModelID: &modelID, Version: 1,
		Data: []byte(`421`),
	}
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeUpgradeData,
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
		Type:      ld.TypeUpgradeData,
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

	input = &ld.TxUpdater{ID: &did, ModelID: &modelID, Version: 1,
		Data: []byte(`421`),
		To:   &modelKeeper,
	}
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeUpgradeData,
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

	input = &ld.TxUpdater{ID: &did, ModelID: &modelID, Version: 1,
		Data: []byte(`421`),
		To:   &modelKeeper,
	}
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeUpgradeData,
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

	input = &ld.TxUpdater{ID: &did, ModelID: &modelID, Version: 1,
		Data: []byte(`421`),
		To:   &modelKeeper,
	}
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeUpgradeData,
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

	input = &ld.TxUpdater{ID: &did, ModelID: &modelID, Version: 1,
		Data:   []byte(`421`),
		To:     &modelKeeper,
		Amount: new(big.Int).SetUint64(unit.MilliLDC),
	}
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeUpgradeData,
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

	input = &ld.TxUpdater{ID: &did, ModelID: &modelID, Version: 1,
		Data:   []byte(`421`),
		To:     &modelKeeper,
		Amount: new(big.Int).SetUint64(unit.MilliLDC),
	}
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeUpgradeData,
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

	input = &ld.TxUpdater{ID: &did, ModelID: &modelID, Version: 1,
		Data:   []byte(`421`),
		To:     &modelKeeper,
		Amount: new(big.Int).SetUint64(unit.MilliLDC),
		Expire: 10,
	}
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeUpgradeData,
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
		"insufficient NativeLDC balance, expected 3004200, got 0")
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
		ModelID:   ld.CBORModelID,
		Version:   1,
		Threshold: 1,
		Keepers:   signer.Keys{signer.Signer1.Key()},
		Payload:   encoding.MustMarshalCBOR(42),
		ID:        did,
	}
	assert.NoError(di.SyntacticVerify())
	cs.SaveData(di)
	cs.CommitAccounts()
	assert.ErrorContains(itx.Apply(ctx, cs),
		"load model error")
	cs.CheckoutAccounts()

	modelInput := &ld.ModelInfo{
		Name:      "Tester",
		Threshold: 1,
		Keepers:   signer.Keys{signer.Signer2.Key()},
		Schema: `
		type Tester struct {
			name    String (rename "n")
			version Int    (rename "v")
		}
		`,
	}

	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeCreateModel,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      modelKeeper,
		Data:      modelInput.Bytes(),
	}}
	assert.NoError(ltx.SignWith(signer.Signer2))
	assert.NoError(ltx.SyntacticVerify())
	itx, err = NewTx(ltx)
	require.NoError(t, err)

	modelAcc := cs.MustAccount(modelKeeper)
	modelAcc.Add(ids.NativeToken, new(big.Int).SetUint64(unit.LDC))
	assert.NoError(itx.Apply(ctx, cs))
	modelKeeperGas := ltx.Gas()

	modelID = ids.ModelIDFromHash(ltx.ID)
	input = &ld.TxUpdater{ID: &did, ModelID: &modelID, Version: 1,
		Data: encoding.MustMarshalCBOR(421),
	}
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeUpgradeData,
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

	cs.CommitAccounts()
	assert.ErrorContains(itx.Apply(ctx, cs),
		`apply patch error, ld.IPLDModel("Tester").ApplyPatch`)
	cs.CheckoutAccounts()

	patchDoc := cborpatch.Patch{
		{Op: "add", Path: "/v", Value: encoding.MustMarshalCBOR(2)},
	}
	input = &ld.TxUpdater{ID: &did, ModelID: &modelID, Version: 1,
		Data:   encoding.MustMarshalCBOR(patchDoc),
		To:     &modelKeeper,
		Amount: new(big.Int).SetUint64(unit.MilliLDC),
	}
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeUpgradeData,
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
		"TxUpgradeData.SyntacticVerify: invalid to, expected 0x44171C37Ff5D7B7bb8Dcad5C81f16284A229E641, got <nil>")

	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeUpgradeData,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      owner,
		To:        ids.GenesisAccount.Ptr(),
		Amount:    new(big.Int).SetUint64(unit.NanoLDC),
		Data:      input.Bytes(),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err,
		"TxUpgradeData.SyntacticVerify: invalid to, expected 0x44171C37Ff5D7B7bb8Dcad5C81f16284A229E641, got 0xFFfFFFfFfffFFfFFffFFFfFfFffFFFfffFfFFFff")

	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeUpgradeData,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      owner,
		To:        &modelKeeper,
		Amount:    new(big.Int).SetUint64(unit.NanoLDC),
		Data:      input.Bytes(),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err,
		"TxUpgradeData.SyntacticVerify: invalid amount, expected 1000000, got 1")

	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeUpgradeData,
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
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "no exSignatures")

	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeUpgradeData,
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
	assert.NoError(ltx.ExSignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	itx, err = NewTx(ltx)
	require.NoError(t, err)

	cs.CommitAccounts()
	assert.ErrorContains(itx.Apply(ctx, cs),
		`invalid node detected`)
	cs.CheckoutAccounts()

	di = &ld.DataInfo{
		ModelID:   ld.CBORModelID,
		Version:   1,
		Threshold: 1,
		Keepers:   signer.Keys{signer.Signer1.Key()},
		Payload:   encoding.MustMarshalCBOR(map[string]interface{}{"n": "LDVM"}),
		ID:        did,
	}
	assert.NoError(di.SyntacticVerify())
	cs.SaveData(di)

	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeUpgradeData,
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
	assert.NoError(ltx.ExSignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	itx, err = NewTx(ltx)
	require.NoError(t, err)

	cs.CommitAccounts()
	assert.ErrorContains(itx.Apply(ctx, cs),
		`TxUpgradeData.Apply: invalid exSignature for model keepers`)
	cs.CheckoutAccounts()

	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeUpgradeData,
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
	assert.NoError(ltx.ExSignWith(signer.Signer2))
	assert.NoError(ltx.SyntacticVerify())
	itx, err = NewTx(ltx)
	require.NoError(t, err)
	assert.NoError(itx.Apply(ctx, cs))

	ownerGas := ltx.Gas()
	assert.Equal((ownerGas+modelKeeperGas)*ctx.Price,
		itx.(*TxUpgradeData).ldc.Balance().Uint64())
	assert.Equal((ownerGas+modelKeeperGas)*100,
		itx.(*TxUpgradeData).miner.Balance().Uint64())
	assert.Equal(unit.LDC-ownerGas*(ctx.Price+100)-unit.MilliLDC,
		ownerAcc.Balance().Uint64())
	assert.Equal(unit.LDC*2-ownerGas*(ctx.Price+100)-unit.MilliLDC,
		ownerAcc.BalanceOfAll(ids.NativeToken).Uint64())
	assert.Equal(uint64(1), ownerAcc.Nonce())

	di2, err := cs.LoadData(di.ID)
	require.NoError(t, err)
	assert.Equal(uint64(2), di2.Version)
	assert.Equal(ld.CBORModelID, di.ModelID)
	assert.Equal(modelID, di2.ModelID)
	assert.Equal(encoding.MustMarshalCBOR(map[string]interface{}{
		"n": "LDVM",
	}), []byte(di.Payload))
	assert.Equal(encoding.MustMarshalCBOR(map[string]interface{}{
		"n": "LDVM",
		"v": 2,
	}), []byte(di2.Payload))

	jsondata, err := itx.MarshalJSON()
	require.NoError(t, err)
	// fmt.Println(string(jsondata))
	assert.Equal(`{"tx":{"type":"TypeUpgradeData","chainID":2357,"nonce":0,"gasTip":100,"gasFeeCap":1000,"from":"0x8db97c7cECe249C2b98bdc0226cc4C2A57bF52fc","to":"0x44171C37Ff5D7B7bb8Dcad5C81f16284A229E641","amount":1000000,"data":{"id":"AQIDBAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAACs148t","mid":"ePJIiccufhcsCSZxBGUt_rw_RArEjblf","version":1,"to":"0x44171C37Ff5D7B7bb8Dcad5C81f16284A229E641","amount":1000000,"data":"gaNib3BjYWRkZHBhdGhiL3ZldmFsdWUCNeW97w"}},"sigs":["zfl447E-aNXCXjzf9d31gdzr5F5jA1eOX2PN668ieBENojR4dLiMGnZNx8KYDs-hhQu96qtX8bIg90f3AF3_JwEhWqX4"],"exSigs":["Hru-bi1UK7t7Mci9nKLs0vTv7-m2c3WYy0rcbzXTcNEwCXplMZGeQ4XVfssiTq6QH2ykq9h-Fm-hNNUlDKtStAFEBRnD"],"id":"h-wQf-Ck4OvCdRdGFU04ZHuQEt3fD45IpXITXT2DAXzRXVE0"}`, string(jsondata))

	assert.NoError(cs.VerifyState())
}

func TestTxUpgradeNameServiceData(t *testing.T) {
	t.Run("name service data can not upgrade", func(t *testing.T) {
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

		modelID := ids.ModelID{1, 2, 3}
		patchDoc := cborpatch.Patch{
			{Op: "replace", Path: "/n", Value: encoding.MustMarshalCBOR("ldc2.to.")},
		}
		input = &ld.TxUpdater{ID: &di.ID, ModelID: &modelID, Version: 1,
			Data:   encoding.MustMarshalCBOR(patchDoc),
			To:     recipient.Ptr(),
			Amount: new(big.Int).SetUint64(unit.MilliLDC),
		}

		ltx = &ld.Transaction{Tx: ld.TxData{
			Type:      ld.TypeUpgradeData,
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
			`TxUpgradeData.Apply: name service data can not upgrade`)
		cs.CheckoutAccounts()

		assert.NoError(cs.VerifyState())
	})

	t.Run("can not upgrade to name service data", func(t *testing.T) {
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
			ModelID:   &ld.CBORModelID,
			Version:   1,
			Threshold: ld.Uint16Ptr(1),
			Keepers:   &signer.Keys{signer.Signer1.Key()},
			Data:      name.Bytes(),
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
			Data:      input.Bytes(),
		}}
		assert.NoError(ltx.SignWith(signer.Signer1))
		assert.NoError(ltx.SyntacticVerify())

		senderAcc := cs.MustAccount(sender)
		assert.NoError(senderAcc.Add(ids.NativeToken, new(big.Int).SetUint64(unit.LDC)))
		assert.NoError(cs.SaveModel(mi))

		ltx.Timestamp = 10
		itx, err := NewTx(ltx)
		require.NoError(t, err)

		_, err = cs.LoadDataByName("ldc.to.")
		assert.ErrorContains(err, `"ldc.to." not found`)
		assert.NoError(itx.Apply(ctx, cs))
		_, err = cs.LoadDataByName("ldc.to.")
		assert.ErrorContains(err, `"ldc.to." not found`)

		di, err := cs.LoadData(ids.DataID(ltx.ID))
		require.NoError(t, err)
		assert.Equal(name.Bytes(), []byte(di.Payload))

		patchDoc := cborpatch.Patch{
			{Op: "replace", Path: "/n", Value: encoding.MustMarshalCBOR("ldc2.to.")},
		}
		input = &ld.TxUpdater{ID: &di.ID, ModelID: &mi.ID, Version: 1,
			Data:   encoding.MustMarshalCBOR(patchDoc),
			To:     recipient.Ptr(),
			Amount: new(big.Int).SetUint64(unit.MilliLDC),
		}

		ltx = &ld.Transaction{Tx: ld.TxData{
			Type:      ld.TypeUpgradeData,
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
			`TxUpgradeData.Apply: can not upgrade to name service data`)
		cs.CheckoutAccounts()

		assert.NoError(cs.VerifyState())
	})
}
