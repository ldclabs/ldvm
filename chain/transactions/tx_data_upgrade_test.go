// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package transactions

import (
	"math/big"
	"testing"

	cborpatch "github.com/ldclabs/cbor-patch"
	"github.com/ldclabs/ldvm/constants"
	"github.com/ldclabs/ldvm/ld"
	"github.com/ldclabs/ldvm/ld/service"
	"github.com/ldclabs/ldvm/util"

	"github.com/stretchr/testify/assert"
)

func TestTxUpgradeData(t *testing.T) {
	assert := assert.New(t)

	// SyntacticVerify
	tx := &TxUpgradeData{}
	assert.ErrorContains(tx.SyntacticVerify(), "nil pointer")
	_, err := tx.MarshalJSON()
	assert.NoError(err)

	ctx := NewMockChainContext()
	cs := ctx.MockChainState()
	token := ld.MustNewToken("$LDC")

	owner := util.Signer1.Address()
	modelKeeper := util.Signer2.Address()

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
	assert.ErrorContains(err, "DeriveSigners error: no signature")

	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeUpgradeData,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      owner,
		Token:     &token,
	}}
	assert.NoError(ltx.SignWith(util.Signer1))
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
	assert.NoError(ltx.SignWith(util.Signer1))
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
	assert.NoError(ltx.SignWith(util.Signer1))
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
	assert.NoError(ltx.SignWith(util.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "invalid data id")

	input = &ld.TxUpdater{ID: &util.DataIDEmpty}
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeUpgradeData,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      owner,
		Data:      input.Bytes(),
	}}
	assert.NoError(ltx.SignWith(util.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "invalid data id")

	did := util.DataID{1, 2, 3, 4}
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
	assert.NoError(ltx.SignWith(util.Signer1))
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
	assert.NoError(ltx.SignWith(util.Signer1))
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
	assert.NoError(ltx.SignWith(util.Signer1))
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
	assert.NoError(ltx.SignWith(util.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "invalid model id")

	modelID := util.ModelID{1, 2, 3, 4}
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
	assert.NoError(ltx.SignWith(util.Signer1))
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
	assert.NoError(ltx.SignWith(util.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "no keepers, threshold should be nil")

	input = &ld.TxUpdater{ID: &did, ModelID: &modelID, Version: 1,
		Threshold: ld.Uint16Ptr(1),
		Keepers:   &util.EthIDs{util.Signer1.Address()}}
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeUpgradeData,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      owner,
		Data:      input.Bytes(),
	}}
	assert.NoError(ltx.SignWith(util.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "invalid threshold, should be nil")

	input = &ld.TxUpdater{ID: &did, ModelID: &modelID, Version: 1,
		Approver: &util.EthIDEmpty}
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeUpgradeData,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      owner,
		Data:      input.Bytes(),
	}}
	assert.NoError(ltx.SignWith(util.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "invalid approver, should be nil")

	input = &ld.TxUpdater{ID: &did, ModelID: &modelID,
		Version: 1, ApproveList: []ld.TxType{ld.TypeDeleteData}}
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeUpgradeData,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      owner,
		Data:      input.Bytes(),
	}}
	assert.NoError(ltx.SignWith(util.Signer1))
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
	assert.NoError(ltx.SignWith(util.Signer1))
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
	assert.NoError(ltx.SignWith(util.Signer1))
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
		Amount:    new(big.Int).SetUint64(constants.MilliLDC),
	}}
	assert.NoError(ltx.SignWith(util.Signer1))
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
	assert.NoError(ltx.SignWith(util.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err,
		"invalid to, expected 0x44171C37Ff5D7B7bb8dcad5C81f16284A229e641, got <nil>")

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
		To:        &constants.GenesisAccount,
		Data:      input.Bytes(),
	}}
	assert.NoError(ltx.SignWith(util.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err,
		"invalid to, expected 0x44171C37Ff5D7B7bb8dcad5C81f16284A229e641, got 0xFFfFfFffFFfffFFfFFfFFFFFffFFFffffFfFFFfF")

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
	assert.NoError(ltx.SignWith(util.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "nil amount")

	input = &ld.TxUpdater{ID: &did, ModelID: &modelID, Version: 1,
		Data:   []byte(`421`),
		To:     &modelKeeper,
		Amount: new(big.Int).SetUint64(constants.MilliLDC),
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
	assert.NoError(ltx.SignWith(util.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "nil amount")

	input = &ld.TxUpdater{ID: &did, ModelID: &modelID, Version: 1,
		Data:   []byte(`421`),
		To:     &modelKeeper,
		Amount: new(big.Int).SetUint64(constants.MilliLDC),
	}
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeUpgradeData,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      owner,
		To:        &modelKeeper,
		Amount:    new(big.Int).SetUint64(constants.MilliLDC),
		Data:      input.Bytes(),
	}}
	assert.NoError(ltx.SignWith(util.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	ltx.Timestamp = 10
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "data expired")

	input = &ld.TxUpdater{ID: &did, ModelID: &modelID, Version: 1,
		Data:   []byte(`421`),
		To:     &modelKeeper,
		Amount: new(big.Int).SetUint64(constants.MilliLDC),
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
		Amount:    new(big.Int).SetUint64(constants.MilliLDC),
		Data:      input.Bytes(),
	}}
	assert.NoError(ltx.SignWith(util.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	ltx.Timestamp = 10
	_, err = NewTx(ltx)
	assert.ErrorContains(err,
		"invalid exSignatures, Transaction.ExSigners error: DeriveSigners error: no signature")

	assert.NoError(ltx.ExSignWith(util.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	ltx.Timestamp = 10
	itx, err := NewTx(ltx)
	assert.NoError(err)

	cs.CommitAccounts()
	assert.ErrorContains(itx.Apply(ctx, cs),
		"insufficient NativeLDC balance, expected 3004200, got 0")
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
		ModelID:   ld.CBORModelID,
		Version:   1,
		Threshold: 1,
		Keepers:   util.EthIDs{util.Signer1.Address()},
		Payload:   util.MustMarshalCBOR(42),
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
		Keepers:   util.EthIDs{util.Signer2.Address()},
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
	assert.NoError(ltx.SignWith(util.Signer2))
	assert.NoError(ltx.SyntacticVerify())
	itx, err = NewTx(ltx)
	assert.NoError(err)

	modelAcc := cs.MustAccount(modelKeeper)
	modelAcc.Add(constants.NativeToken, new(big.Int).SetUint64(constants.LDC))
	assert.NoError(itx.Apply(ctx, cs))
	modelKeeperGas := ltx.Gas()

	modelID = util.ModelID(ltx.ShortID())
	input = &ld.TxUpdater{ID: &did, ModelID: &modelID, Version: 1,
		Data: util.MustMarshalCBOR(421),
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
	assert.NoError(ltx.SignWith(util.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	itx, err = NewTx(ltx)
	assert.NoError(err)

	cs.CommitAccounts()
	assert.ErrorContains(itx.Apply(ctx, cs),
		`apply patch error, IPLDModel("Tester").ApplyPatch error: invalid CBOR patch`)
	cs.CheckoutAccounts()

	patchDoc := cborpatch.Patch{
		{Op: "add", Path: "/v", Value: util.MustMarshalCBOR(2)},
	}
	input = &ld.TxUpdater{ID: &did, ModelID: &modelID, Version: 1,
		Data:   util.MustMarshalCBOR(patchDoc),
		To:     &modelKeeper,
		Amount: new(big.Int).SetUint64(constants.MilliLDC),
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
	assert.NoError(ltx.SignWith(util.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err,
		"TxUpgradeData.SyntacticVerify error: invalid to, expected 0x44171C37Ff5D7B7bb8dcad5C81f16284A229e641, got <nil>")

	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeUpgradeData,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      owner,
		To:        &constants.GenesisAccount,
		Amount:    new(big.Int).SetUint64(constants.NanoLDC),
		Data:      input.Bytes(),
	}}
	assert.NoError(ltx.SignWith(util.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err,
		"TxUpgradeData.SyntacticVerify error: invalid to, expected 0x44171C37Ff5D7B7bb8dcad5C81f16284A229e641, got 0xFFfFfFffFFfffFFfFFfFFFFFffFFFffffFfFFFfF")

	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeUpgradeData,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      owner,
		To:        &modelKeeper,
		Amount:    new(big.Int).SetUint64(constants.NanoLDC),
		Data:      input.Bytes(),
	}}
	assert.NoError(ltx.SignWith(util.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err,
		"TxUpgradeData.SyntacticVerify error: invalid amount, expected 1000000, got 1")

	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeUpgradeData,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      owner,
		To:        &modelKeeper,
		Amount:    new(big.Int).SetUint64(constants.MilliLDC),
		Data:      input.Bytes(),
	}}
	assert.NoError(ltx.SignWith(util.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err,
		"TxUpgradeData.SyntacticVerify error: invalid exSignatures, Transaction.ExSigners error: DeriveSigners error: no signature")

	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeUpgradeData,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      owner,
		To:        &modelKeeper,
		Amount:    new(big.Int).SetUint64(constants.MilliLDC),
		Data:      input.Bytes(),
	}}
	assert.NoError(ltx.SignWith(util.Signer1))
	assert.NoError(ltx.ExSignWith(util.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	itx, err = NewTx(ltx)
	assert.NoError(err)

	cs.CommitAccounts()
	assert.ErrorContains(itx.Apply(ctx, cs),
		`apply patch error, IPLDModel("Tester").ApplyPatch error: unexpected node "42", invalid node detected`)
	cs.CheckoutAccounts()

	di = &ld.DataInfo{
		ModelID:   ld.CBORModelID,
		Version:   1,
		Threshold: 1,
		Keepers:   util.EthIDs{util.Signer1.Address()},
		Payload:   util.MustMarshalCBOR(map[string]interface{}{"n": "LDVM"}),
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
		Amount:    new(big.Int).SetUint64(constants.MilliLDC),
		Data:      input.Bytes(),
	}}
	assert.NoError(ltx.SignWith(util.Signer1))
	assert.NoError(ltx.ExSignWith(util.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	itx, err = NewTx(ltx)
	assert.NoError(err)

	cs.CommitAccounts()
	assert.ErrorContains(itx.Apply(ctx, cs),
		`TxUpgradeData.Apply error: invalid exSignature for model keepers`)
	cs.CheckoutAccounts()

	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeUpgradeData,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      owner,
		To:        &modelKeeper,
		Amount:    new(big.Int).SetUint64(constants.MilliLDC),
		Data:      input.Bytes(),
	}}
	assert.NoError(ltx.SignWith(util.Signer1))
	assert.NoError(ltx.ExSignWith(util.Signer2))
	assert.NoError(ltx.SyntacticVerify())
	itx, err = NewTx(ltx)
	assert.NoError(err)
	assert.NoError(itx.Apply(ctx, cs))

	ownerGas := ltx.Gas()
	assert.Equal((ownerGas+modelKeeperGas)*ctx.Price,
		itx.(*TxUpgradeData).ldc.Balance().Uint64())
	assert.Equal((ownerGas+modelKeeperGas)*100,
		itx.(*TxUpgradeData).miner.Balance().Uint64())
	assert.Equal(constants.LDC-ownerGas*(ctx.Price+100)-constants.MilliLDC,
		ownerAcc.Balance().Uint64())
	assert.Equal(uint64(1), ownerAcc.Nonce())

	di2, err := cs.LoadData(di.ID)
	assert.NoError(err)
	assert.Equal(uint64(2), di2.Version)
	assert.Equal(ld.CBORModelID, di.ModelID)
	assert.Equal(modelID, di2.ModelID)
	assert.Equal(util.MustMarshalCBOR(map[string]interface{}{
		"n": "LDVM",
	}), []byte(di.Payload))
	assert.Equal(util.MustMarshalCBOR(map[string]interface{}{
		"n": "LDVM",
		"v": 2,
	}), []byte(di2.Payload))

	jsondata, err := itx.MarshalJSON()
	assert.NoError(err)
	// fmt.Println(string(jsondata))
	assert.Equal(`{"tx":{"type":"TypeUpgradeData","chainID":2357,"nonce":0,"gasTip":100,"gasFeeCap":1000,"from":"0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC","to":"0x44171C37Ff5D7B7bb8dcad5C81f16284A229e641","amount":1000000,"data":{"id":"SkB92DD9M2yeCadw22VbnxfV6b7W5YEnnLRs6fKivk6wh2Zy","mid":"BGC5DQCP5JQLhATeqAUfDy3BKELz4RsXS","version":1,"to":"0x44171C37Ff5D7B7bb8dcad5C81f16284A229e641","amount":1000000,"data":"0x81a3626f70636164646470617468622f766576616c75650299318ca7"}},"sigs":["e41c97e27ec618e4f99f4f045571a93af88923242c8a899b9bcda3f6099a4fa268994e415b9b5f495ca6344a1f51d22d86dd25a32b59e2df9fa023e302ebd10601"],"exSigs":["6607151cc7c169a940aed88a580a516ec8fda7040c780831c645b99f8b3fa6a6693da512eb82c986d63f98147a4d88be9dbd5023839fc56530848cb0b1e01ba001"],"id":"juiZ3wi7F3SSqyD4y5vzGD1xpQV4bSGMShE1rarMyJDgijGFG"}`, string(jsondata))

	assert.NoError(cs.VerifyState())
}

func TestTxUpgradeNameServiceData(t *testing.T) {
	t.Run("name service data can not upgrade", func(t *testing.T) {
		assert := assert.New(t)

		ctx := NewMockChainContext()
		cs := ctx.MockChainState()

		sender := util.Signer1.Address()
		recipient := util.Signer2.Address()

		nm, err := service.NameModel()
		assert.NoError(err)
		mi := &ld.ModelInfo{
			Name:      nm.Name(),
			Threshold: 1,
			Keepers:   util.EthIDs{util.Signer2.Address()},
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
			Keepers:   &util.EthIDs{util.Signer1.Address()},
			Data:      name.Bytes(),
			To:        &recipient,
			Expire:    100,
			Amount:    new(big.Int).SetUint64(constants.MilliLDC),
		}
		assert.NoError(input.SyntacticVerify())
		ltx := &ld.Transaction{Tx: ld.TxData{
			Type:      ld.TypeCreateData,
			ChainID:   ctx.ChainConfig().ChainID,
			Nonce:     0,
			GasTip:    100,
			GasFeeCap: ctx.Price,
			From:      sender,
			To:        &recipient,
			Amount:    new(big.Int).SetUint64(constants.MilliLDC),
			Data:      input.Bytes(),
		}}
		assert.NoError(ltx.SignWith(util.Signer1))
		assert.NoError(ltx.ExSignWith(util.Signer2))
		assert.NoError(ltx.SyntacticVerify())

		senderAcc := cs.MustAccount(sender)
		assert.NoError(senderAcc.Add(constants.NativeToken, new(big.Int).SetUint64(constants.LDC)))
		assert.NoError(cs.SaveModel(mi))

		ltx.Timestamp = 10
		itx, err := NewTx(ltx)
		assert.NoError(err)

		_, err = cs.LoadDataByName("ldc.to.")
		assert.ErrorContains(err, `"ldc.to." not found`)
		assert.NoError(itx.Apply(ctx, cs))
		di, err := cs.LoadDataByName("ldc.to.")
		assert.NoError(err)
		assert.Equal(mi.ID, di.ModelID)

		modelID := util.ModelID{1, 2, 3}
		patchDoc := cborpatch.Patch{
			{Op: "replace", Path: "/n", Value: util.MustMarshalCBOR("ldc2.to.")},
		}
		input = &ld.TxUpdater{ID: &di.ID, ModelID: &modelID, Version: 1,
			Data:   util.MustMarshalCBOR(patchDoc),
			To:     &recipient,
			Amount: new(big.Int).SetUint64(constants.MilliLDC),
		}

		ltx = &ld.Transaction{Tx: ld.TxData{
			Type:      ld.TypeUpgradeData,
			ChainID:   ctx.ChainConfig().ChainID,
			Nonce:     1,
			GasTip:    100,
			GasFeeCap: ctx.Price,
			From:      sender,
			To:        &recipient,
			Amount:    new(big.Int).SetUint64(constants.MilliLDC),
			Data:      input.Bytes(),
		}}
		assert.NoError(ltx.SignWith(util.Signer1))
		assert.NoError(ltx.ExSignWith(util.Signer2))
		assert.NoError(ltx.SyntacticVerify())
		itx, err = NewTx(ltx)
		assert.NoError(err)

		cs.CommitAccounts()
		assert.ErrorContains(itx.Apply(ctx, cs),
			`TxUpgradeData.Apply error: name service data can not upgrade`)
		cs.CheckoutAccounts()

		assert.NoError(cs.VerifyState())
	})

	t.Run("can not upgrade to name service data", func(t *testing.T) {
		assert := assert.New(t)

		ctx := NewMockChainContext()
		cs := ctx.MockChainState()

		sender := util.Signer1.Address()
		recipient := util.Signer2.Address()

		nm, err := service.NameModel()
		assert.NoError(err)
		mi := &ld.ModelInfo{
			Name:      nm.Name(),
			Threshold: 1,
			Keepers:   util.EthIDs{util.Signer2.Address()},
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
			Keepers:   &util.EthIDs{util.Signer1.Address()},
			Data:      name.Bytes(),
			Expire:    100,
			Amount:    new(big.Int).SetUint64(constants.MilliLDC),
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
		assert.NoError(ltx.SignWith(util.Signer1))
		assert.NoError(ltx.SyntacticVerify())

		senderAcc := cs.MustAccount(sender)
		assert.NoError(senderAcc.Add(constants.NativeToken, new(big.Int).SetUint64(constants.LDC)))
		assert.NoError(cs.SaveModel(mi))

		ltx.Timestamp = 10
		itx, err := NewTx(ltx)
		assert.NoError(err)

		_, err = cs.LoadDataByName("ldc.to.")
		assert.ErrorContains(err, `"ldc.to." not found`)
		assert.NoError(itx.Apply(ctx, cs))
		_, err = cs.LoadDataByName("ldc.to.")
		assert.ErrorContains(err, `"ldc.to." not found`)

		di, err := cs.LoadData(util.DataID(ltx.ID))
		assert.NoError(err)
		assert.Equal(name.Bytes(), []byte(di.Payload))

		patchDoc := cborpatch.Patch{
			{Op: "replace", Path: "/n", Value: util.MustMarshalCBOR("ldc2.to.")},
		}
		input = &ld.TxUpdater{ID: &di.ID, ModelID: &mi.ID, Version: 1,
			Data:   util.MustMarshalCBOR(patchDoc),
			To:     &recipient,
			Amount: new(big.Int).SetUint64(constants.MilliLDC),
		}

		ltx = &ld.Transaction{Tx: ld.TxData{
			Type:      ld.TypeUpgradeData,
			ChainID:   ctx.ChainConfig().ChainID,
			Nonce:     1,
			GasTip:    100,
			GasFeeCap: ctx.Price,
			From:      sender,
			To:        &recipient,
			Amount:    new(big.Int).SetUint64(constants.MilliLDC),
			Data:      input.Bytes(),
		}}
		assert.NoError(ltx.SignWith(util.Signer1))
		assert.NoError(ltx.ExSignWith(util.Signer2))
		assert.NoError(ltx.SyntacticVerify())
		itx, err = NewTx(ltx)
		assert.NoError(err)

		cs.CommitAccounts()
		assert.ErrorContains(itx.Apply(ctx, cs),
			`TxUpgradeData.Apply error: can not upgrade to name service data`)
		cs.CheckoutAccounts()

		assert.NoError(cs.VerifyState())
	})
}
