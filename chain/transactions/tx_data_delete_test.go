// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package transactions

import (
	"math/big"
	"testing"

	"github.com/ldclabs/ldvm/constants"
	"github.com/ldclabs/ldvm/ld"
	"github.com/ldclabs/ldvm/util"
	"github.com/stretchr/testify/assert"
)

func TestTxDeleteData(t *testing.T) {
	assert := assert.New(t)

	// SyntacticVerify
	tx := &TxDeleteData{}
	assert.ErrorContains(tx.SyntacticVerify(), "nil pointer")
	_, err := tx.MarshalJSON()
	assert.NoError(err)

	ctx := NewMockChainContext()
	cs := ctx.MockChainState()

	sender := util.Signer1.Address()
	approver := util.Signer2.Address()

	ltx := &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeDeleteData,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      sender,
	}}
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "DeriveSigners error: no signature")

	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeDeleteData,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      sender,
		To:        &constants.GenesisAccount,
	}}
	assert.NoError(ltx.SignWith(util.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "invalid to, should be nil")

	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeDeleteData,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      sender,
		Token:     &constants.NativeToken,
	}}
	assert.NoError(ltx.SignWith(util.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "invalid token, should be nil")

	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeDeleteData,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      sender,
		Amount:    big.NewInt(1),
	}}
	assert.NoError(ltx.SignWith(util.Signer1))
	assert.ErrorContains(ltx.SyntacticVerify(), "nil \"to\" together with amount")

	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeDeleteData,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      sender,
	}}
	assert.NoError(ltx.SignWith(util.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "invalid data")

	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeDeleteData,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      sender,
		Data:      []byte("你好👋"),
	}}
	assert.NoError(ltx.SignWith(util.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "cbor: unexpected following extraneous data")

	input := &ld.TxUpdater{}
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeDeleteData,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      sender,
		Data:      input.Bytes(),
	}}
	assert.NoError(ltx.SignWith(util.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "invalid data id")

	input = &ld.TxUpdater{ID: &util.DataIDEmpty}
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeDeleteData,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      sender,
		Data:      input.Bytes(),
	}}
	assert.NoError(ltx.SignWith(util.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "invalid data id")

	did := util.DataID{1, 2, 3, 4}
	input = &ld.TxUpdater{ID: &did}
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeDeleteData,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      sender,
		Data:      input.Bytes(),
	}}
	assert.NoError(ltx.SignWith(util.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "invalid data version")

	input = &ld.TxUpdater{ID: &did, Version: 1}
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeDeleteData,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      sender,
		Data:      input.Bytes(),
	}}
	assert.NoError(ltx.SignWith(util.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	itx, err := NewTx(ltx)
	assert.NoError(err)

	senderAcc := cs.MustAccount(sender)
	senderAcc.Add(constants.NativeToken, new(big.Int).SetUint64(constants.LDC))
	cs.CommitAccounts()
	assert.ErrorContains(itx.Apply(ctx, cs), "SkB92DD9M2yeCadw22VbnxfV6b7W5YEnnLRs6fKivk6wh2Zy not found")
	cs.CheckoutAccounts()

	di := &ld.DataInfo{
		ModelID:   ld.RawModelID,
		Version:   2,
		Threshold: 1,
		Keepers:   util.EthIDs{util.Signer2.Address()},
		Approver:  &approver,
		Payload:   []byte(`42`),
		ID:        did,
	}
	assert.NoError(di.SyntacticVerify())
	assert.NoError(cs.SaveData(di))
	cs.CommitAccounts()
	assert.ErrorContains(itx.Apply(ctx, cs), "invalid version, expected 2, got 1")
	cs.CheckoutAccounts()

	input = &ld.TxUpdater{ID: &did, Version: 2}
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeDeleteData,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      sender,
		Data:      input.Bytes(),
	}}
	assert.NoError(ltx.SignWith(util.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	itx, err = NewTx(ltx)
	assert.NoError(err)
	cs.CommitAccounts()
	assert.ErrorContains(itx.Apply(ctx, cs), "invalid signatures for data keepers")
	cs.CheckoutAccounts()

	di = &ld.DataInfo{
		ModelID:   ld.RawModelID,
		Version:   2,
		Threshold: 1,
		Keepers:   util.EthIDs{util.Signer1.Address()},
		Approver:  &approver,
		Payload:   []byte(`42`),
		ID:        did,
	}
	assert.NoError(di.SyntacticVerify())
	assert.NoError(cs.SaveData(di))

	cs.CommitAccounts()
	assert.ErrorContains(itx.Apply(ctx, cs), "invalid signature for data approver")
	cs.CheckoutAccounts()

	di = &ld.DataInfo{
		ModelID:   ld.RawModelID,
		Version:   2,
		Threshold: 1,
		Keepers:   util.EthIDs{util.Signer1.Address()},
		Payload:   []byte(`42`),
		ID:        did,
	}
	assert.NoError(di.SyntacticVerify())
	assert.NoError(cs.SaveData(di))
	assert.NoError(itx.Apply(ctx, cs))

	senderGas := ltx.Gas()
	assert.Equal(senderGas*ctx.Price,
		itx.(*TxDeleteData).ldc.Balance().Uint64())
	assert.Equal(senderGas*100,
		itx.(*TxDeleteData).miner.Balance().Uint64())
	assert.Equal(constants.LDC-senderGas*(ctx.Price+100),
		senderAcc.Balance().Uint64())
	assert.Equal(uint64(1), senderAcc.Nonce())

	di2, err := cs.LoadData(di.ID)
	assert.NoError(err)
	assert.Equal(uint64(0), di2.Version)

	jsondata, err := itx.MarshalJSON()
	assert.NoError(err)
	// fmt.Println(string(jsondata))
	assert.Equal(`{"tx":{"type":"TypeDeleteData","chainID":2357,"nonce":0,"gasTip":100,"gasFeeCap":1000,"from":"0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC","data":{"id":"SkB92DD9M2yeCadw22VbnxfV6b7W5YEnnLRs6fKivk6wh2Zy","version":2}},"sigs":["4efd89df9fd7468f998d4c594069ea19b9a03615c5337cb3f41411d88377b7bf16c334450b5b2df4ca8bb94f98542e4df7d94ea923d67d444cd44a8d850fdd0601"],"id":"2cfPGtkyWPf5ubHz5Wo3zTCY2NL5EZyT2fdgxuEymSs4T8b4Um"}`, string(jsondata))

	input = &ld.TxUpdater{ID: &did, Version: 2}
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeDeleteData,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     1,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      sender,
		Data:      input.Bytes(),
	}}
	assert.NoError(ltx.SignWith(util.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	itx, err = NewTx(ltx)
	assert.NoError(err)
	cs.CommitAccounts()
	assert.ErrorContains(itx.Apply(ctx, cs), "invalid version, expected 0, got 2")
	cs.CheckoutAccounts()

	di = &ld.DataInfo{
		ModelID:   ld.RawModelID,
		Version:   2,
		Threshold: 1,
		Keepers:   util.EthIDs{util.Signer1.Address()},
		Payload:   []byte(`42`),
		ID:        did,
	}
	assert.NoError(di.SyntacticVerify())
	assert.NoError(cs.SaveData(di))

	input = &ld.TxUpdater{ID: &did, Version: 2, Data: []byte(`421`)}
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeDeleteData,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     1,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      sender,
		Data:      input.Bytes(),
	}}
	assert.NoError(ltx.SignWith(util.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	itx, err = NewTx(ltx)
	assert.NoError(err)
	assert.NoError(itx.Apply(ctx, cs))

	di2, err = cs.LoadData(di.ID)
	assert.NoError(err)
	assert.Equal(uint64(0), di2.Version)
	assert.Equal([]byte(`421`), []byte(di2.Payload))

	assert.NoError(cs.VerifyState())
}
