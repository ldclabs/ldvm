// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package transactions

import (
	"math/big"
	"testing"

	"github.com/ldclabs/ldvm/constants"
	"github.com/ldclabs/ldvm/ld"
	"github.com/ldclabs/ldvm/ld/service"
	"github.com/ldclabs/ldvm/util"
	"github.com/stretchr/testify/assert"
)

func TestTxUpdateModelInfo(t *testing.T) {
	assert := assert.New(t)

	// SyntacticVerify
	tx := &TxUpdateModelInfo{}
	assert.ErrorContains(tx.SyntacticVerify(), "nil pointer")
	_, err := tx.MarshalJSON()
	assert.NoError(err)

	ctx := NewMockChainContext()
	cs := ctx.MockChainState()
	token := ld.MustNewToken("$LDC")

	owner := util.Signer1.Address()
	assert.NoError(err)
	approver := util.Signer2.Address()

	txData := &ld.TxData{
		Type:      ld.TypeUpdateModelInfo,
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
		Type:      ld.TypeUpdateModelInfo,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      owner,
		To:        &constants.GenesisAccount,
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx2(txData.ToTransaction())
	assert.ErrorContains(err, "invalid to, should be nil")

	txData = &ld.TxData{
		Type:      ld.TypeUpdateModelInfo,
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
		Type:      ld.TypeUpdateModelInfo,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      owner,
		Amount:    big.NewInt(1),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx2(txData.ToTransaction())
	assert.ErrorContains(err, "nil to together with amount")

	txData = &ld.TxData{
		Type:      ld.TypeUpdateModelInfo,
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
		Type:      ld.TypeUpdateModelInfo,
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

	input := ld.TxUpdater{}
	assert.NoError(input.SyntacticVerify())
	txData = &ld.TxData{
		Type:      ld.TypeUpdateModelInfo,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      owner,
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx2(txData.ToTransaction())
	assert.ErrorContains(err, "invalid mid")

	input = ld.TxUpdater{ModelID: &util.ModelIDEmpty}
	assert.NoError(input.SyntacticVerify())
	txData = &ld.TxData{
		Type:      ld.TypeUpdateModelInfo,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      owner,
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx2(txData.ToTransaction())
	assert.ErrorContains(err, "invalid mid")

	mid := util.ModelID{'1', '2', '3', '4', '5', '6'}
	input = ld.TxUpdater{ModelID: &mid, Keepers: &util.EthIDs{}}
	assert.ErrorContains(input.SyntacticVerify(), "nil threshold")
	input = ld.TxUpdater{ModelID: &mid, Threshold: ld.Uint16Ptr(0)}
	assert.ErrorContains(input.SyntacticVerify(), "nil keepers")
	input = ld.TxUpdater{ModelID: &mid, Threshold: ld.Uint16Ptr(1), Keepers: &util.EthIDs{}}
	assert.ErrorContains(input.SyntacticVerify(), "invalid threshold, expected <= 0, got 1")

	mid = util.ModelID{'1', '2', '3', '4', '5', '6'}
	input = ld.TxUpdater{ModelID: &mid}
	assert.NoError(input.SyntacticVerify())
	txData = &ld.TxData{
		Type:      ld.TypeUpdateModelInfo,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      owner,
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx2(txData.ToTransaction())
	assert.ErrorContains(err, "nothing to update")

	mid = util.ModelID{'1', '2', '3', '4', '5', '6'}
	input = ld.TxUpdater{
		ModelID:  &mid,
		Approver: &approver,
	}
	assert.NoError(input.SyntacticVerify())
	txData = &ld.TxData{
		Type:      ld.TypeUpdateModelInfo,
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
	assert.ErrorContains(itx.Apply(ctx, cs),
		"insufficient NativeLDC balance, expected 1316700, got 0")
	cs.CheckoutAccounts()
	ownerAcc := cs.MustAccount(owner)
	ownerAcc.Add(constants.NativeToken, new(big.Int).SetUint64(constants.LDC))
	cs.CommitAccounts()
	assert.ErrorContains(itx.Apply(ctx, cs),
		"5V8FMkzy77ibQauKnRxM6aGSLG4AaYTdB not found")
	cs.CheckoutAccounts()

	ipldm, err := service.ProfileModel()
	assert.NoError(err)
	mi := &ld.ModelInfo{
		Name:      ipldm.Name(),
		Threshold: 1,
		Keepers:   util.EthIDs{util.Signer1.Address()},
		Data:      ipldm.Schema(),
		ID:        mid,
	}
	assert.NoError(mi.SyntacticVerify())
	assert.NoError(cs.SaveModel(mi))
	assert.NoError(itx.Apply(ctx, cs))

	assert.Equal(tt.Gas()*ctx.Price,
		itx.(*TxUpdateModelInfo).ldc.Balance().Uint64())
	assert.Equal(tt.Gas()*100,
		itx.(*TxUpdateModelInfo).miner.Balance().Uint64())
	assert.Equal(constants.LDC-tt.Gas()*(ctx.Price+100),
		ownerAcc.Balance().Uint64())
	assert.Equal(uint64(1), ownerAcc.Nonce())

	mi, err = cs.LoadModel(mid)
	assert.NoError(err)
	assert.NotNil(mi.Approver)
	assert.Equal(approver, *mi.Approver)

	jsondata, err := itx.MarshalJSON()
	assert.NoError(err)
	// fmt.Println(string(jsondata))
	assert.Equal(`{"type":"TypeUpdateModelInfo","chainID":2357,"nonce":0,"gasTip":100,"gasFeeCap":1000,"from":"0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC","data":{"mid":"5V8FMkzy77ibQauKnRxM6aGSLG4AaYTdB","approver":"0x44171C37Ff5D7B7bb8dcad5C81f16284A229e641"},"signatures":["454b957046a4413fe6b9c7cc06f9d6b2ce77b0a0a57b236b66d966917e8e2abb6f763e9ecf9255b12e0e9a7c13f82c5004733df0ac9d38266198d465b0145fa100"],"id":"uMUvxUX46YdPKvuUpqVtEBTFpxQeF5fP4x9Jb9zePaNHEVRh1"}`, string(jsondata))

	assert.NoError(cs.VerifyState())

	// approver sign and clear approver
	input = ld.TxUpdater{
		ModelID:   &mid,
		Approver:  &util.EthIDEmpty,
		Threshold: ld.Uint16Ptr(1),
		Keepers:   &util.EthIDs{util.Signer1.Address(), util.Signer2.Address()},
	}
	assert.NoError(input.SyntacticVerify())
	txData = &ld.TxData{
		Type:      ld.TypeUpdateModelInfo,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     1,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      owner,
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	tt = txData.ToTransaction()
	itx, err = NewTx2(tt)
	assert.NoError(err)
	cs.CommitAccounts()
	assert.ErrorContains(itx.Apply(ctx, cs), "invalid signature for approver")
	cs.CheckoutAccounts()

	txData = &ld.TxData{
		Type:      ld.TypeUpdateModelInfo,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     1,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      owner,
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	assert.NoError(txData.SignWith(util.Signer2))
	tt = txData.ToTransaction()
	itx, err = NewTx2(tt)
	assert.NoError(err)
	assert.NoError(itx.Apply(ctx, cs))

	mi, err = cs.LoadModel(mid)
	assert.NoError(err)
	assert.Nil(mi.Approver)
	assert.Equal(util.EthIDs{util.Signer1.Address(), util.Signer2.Address()}, mi.Keepers)

	// check SatisfySigningPlus
	input = ld.TxUpdater{
		ModelID:   &mid,
		Threshold: ld.Uint16Ptr(0),
		Keepers:   &util.EthIDs{util.Signer2.Address()},
	}
	assert.NoError(input.SyntacticVerify())
	txData = &ld.TxData{
		Type:      ld.TypeUpdateModelInfo,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     2,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      owner,
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	tt = txData.ToTransaction()
	itx, err = NewTx2(tt)
	assert.NoError(err)

	cs.CommitAccounts()
	assert.ErrorContains(itx.Apply(ctx, cs), "invalid signatures for keepers")
	cs.CheckoutAccounts()

	txData = &ld.TxData{
		Type:      ld.TypeUpdateModelInfo,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     2,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      owner,
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	assert.NoError(txData.SignWith(util.Signer2))
	tt = txData.ToTransaction()
	itx, err = NewTx2(tt)
	assert.NoError(err)
	assert.NoError(itx.Apply(ctx, cs))

	mi, err = cs.LoadModel(mid)
	assert.NoError(err)
	assert.Nil(mi.Approver)
	assert.Equal(uint16(0), mi.Threshold)
	assert.Equal(util.EthIDs{util.Signer2.Address()}, mi.Keepers)

	assert.NoError(cs.VerifyState())
}
