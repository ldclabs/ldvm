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

func TestTxUpdateDataInfoByAuth(t *testing.T) {
	assert := assert.New(t)

	// SyntacticVerify
	tx := &TxUpdateDataInfoByAuth{}
	assert.ErrorContains(tx.SyntacticVerify(), "nil pointer")
	_, err := tx.MarshalJSON()
	assert.NoError(err)

	ctx := NewMockChainContext()
	cs := ctx.MockChainState()
	token := ld.MustNewToken("$LDC")

	buyer := util.Signer1.Address()
	owner := util.Signer2.Address()

	ltx := &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeUpdateDataInfoByAuth,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      buyer,
	}}
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "DeriveSigners error: no signature")

	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeUpdateDataInfoByAuth,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      buyer,
	}}
	assert.NoError(ltx.SignWith(util.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "nil to")

	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeUpdateDataInfoByAuth,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      buyer,
		To:        &owner,
	}}
	assert.NoError(ltx.SignWith(util.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "invalid data")

	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeUpdateDataInfoByAuth,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      buyer,
		To:        &owner,
		Data:      []byte("你好👋"),
	}}
	assert.NoError(ltx.SignWith(util.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "cbor: unexpected following extraneous data")

	input := &ld.TxUpdater{}
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeUpdateDataInfoByAuth,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      buyer,
		To:        &owner,
		Data:      input.Bytes(),
	}}
	assert.NoError(ltx.SignWith(util.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "invalid data id")

	input = &ld.TxUpdater{ID: &util.DataIDEmpty}
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeUpdateDataInfoByAuth,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      buyer,
		To:        &owner,
		Data:      input.Bytes(),
	}}
	assert.NoError(ltx.SignWith(util.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "invalid data id")

	did := util.DataID{1, 2, 3, 4}
	input = &ld.TxUpdater{ID: &did}
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeUpdateDataInfoByAuth,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      buyer,
		To:        &owner,
		Data:      input.Bytes(),
	}}
	assert.NoError(ltx.SignWith(util.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "invalid data version")

	input = &ld.TxUpdater{ID: &did, Version: 1,
		Keepers:   &util.EthIDs{util.Signer1.Address()},
		Threshold: ld.Uint16Ptr(1)}
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeUpdateDataInfoByAuth,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      buyer,
		To:        &owner,
		Data:      input.Bytes(),
	}}
	assert.NoError(ltx.SignWith(util.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "invalid keepers, should be nil")

	input = &ld.TxUpdater{
		ID: &did, Version: 1,
		TypedSig: util.Signature{1, 2, 3}.Typed(),
		SigClaims: &ld.SigClaims{
			Issuer:     util.DataID{1, 2, 3, 4},
			Subject:    util.DataID{5, 6, 7, 8},
			Audience:   ld.RawModelID,
			Expiration: 100,
			IssuedAt:   1,
			CWTID:      util.Hash{9, 10, 11, 12},
		},
	}
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeUpdateDataInfoByAuth,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      buyer,
		To:        &owner,
		Data:      input.Bytes(),
	}}
	assert.NoError(ltx.SignWith(util.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "invalid sigClaims, should be nil")

	input = &ld.TxUpdater{ID: &did, Version: 1, Approver: &constants.GenesisAccount}
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeUpdateDataInfoByAuth,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      buyer,
		To:        &owner,
		Data:      input.Bytes(),
	}}
	assert.NoError(ltx.SignWith(util.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "invalid approver, should be nil")

	input = &ld.TxUpdater{ID: &did, Version: 1,
		ApproveList: []ld.TxType{ld.TypeUpdateDataInfoByAuth}}
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeUpdateDataInfoByAuth,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      buyer,
		To:        &owner,
		Data:      input.Bytes(),
	}}
	assert.NoError(ltx.SignWith(util.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "invalid approveList, should be nil")

	input = &ld.TxUpdater{ID: &did, Version: 1}
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeUpdateDataInfoByAuth,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      buyer,
		To:        &owner,
		Data:      input.Bytes(),
	}}
	assert.NoError(ltx.SignWith(util.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "nil to")

	input = &ld.TxUpdater{ID: &did, Version: 1, To: &owner}
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeUpdateDataInfoByAuth,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      buyer,
		To:        &constants.GenesisAccount,
		Data:      input.Bytes(),
	}}
	assert.NoError(ltx.SignWith(util.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err,
		"invalid to, expected 0x44171C37Ff5D7B7bb8dcad5C81f16284A229e641, got 0xFFfFfFffFFfffFFfFFfFFFFFffFFFffffFfFFFfF")

	input = &ld.TxUpdater{ID: &did, Version: 1, To: &owner}
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeUpdateDataInfoByAuth,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      buyer,
		To:        &owner,
		Data:      input.Bytes(),
	}}
	assert.NoError(ltx.SignWith(util.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "nil amount")

	input = &ld.TxUpdater{ID: &did, Version: 1, To: &owner,
		Amount: new(big.Int).SetUint64(constants.MilliLDC)}
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeUpdateDataInfoByAuth,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      buyer,
		To:        &owner,
		Amount:    new(big.Int).SetUint64(1),
		Data:      input.Bytes(),
	}}
	assert.NoError(ltx.SignWith(util.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "invalid amount, expected 1000000, got 1")

	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeUpdateDataInfoByAuth,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      buyer,
		To:        &owner,
		Token:     &token,
		Amount:    new(big.Int).SetUint64(constants.MilliLDC),
		Data:      input.Bytes(),
	}}
	assert.NoError(ltx.SignWith(util.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "invalid token, expected NativeToken, got $LDC")

	input = &ld.TxUpdater{ID: &did, Version: 1, To: &owner,
		Amount: new(big.Int).SetUint64(constants.MilliLDC), Token: &token}

	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeUpdateDataInfoByAuth,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      buyer,
		To:        &owner,
		Amount:    new(big.Int).SetUint64(constants.MilliLDC),
		Data:      input.Bytes(),
	}}
	assert.NoError(ltx.SignWith(util.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "invalid token, expected $LDC, got NativeLDC")

	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeUpdateDataInfoByAuth,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      buyer,
		To:        &owner,
		Token:     &token,
		Amount:    new(big.Int).SetUint64(constants.MilliLDC),
		Data:      input.Bytes(),
	}}
	assert.NoError(ltx.SignWith(util.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err,
		"invalid exSignatures, Transaction.ExSigners error: DeriveSigners error: no signature")

	assert.NoError(ltx.ExSignWith(util.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	itx, err := NewTx(ltx)
	assert.NoError(err)

	cs.CommitAccounts()
	assert.ErrorContains(itx.Apply(ctx, cs),
		"insufficient NativeLDC balance, expected 2099900, got 0")
	cs.CheckoutAccounts()

	buyerAcc := cs.MustAccount(buyer)
	buyerAcc.Add(constants.NativeToken, new(big.Int).SetUint64(constants.LDC))
	cs.CommitAccounts()
	assert.ErrorContains(itx.Apply(ctx, cs),
		"insufficient $LDC balance, expected 1000000, got 0")
	cs.CheckoutAccounts()
	buyerAcc.Add(token, new(big.Int).SetUint64(constants.LDC))
	cs.CommitAccounts()
	assert.ErrorContains(itx.Apply(ctx, cs),
		"SkB92DD9M2yeCadw22VbnxfV6b7W5YEnnLRs6fKivk6wh2Zy not found")
	cs.CheckoutAccounts()

	di := &ld.DataInfo{
		ModelID:   ld.RawModelID,
		Version:   2,
		Threshold: 1,
		Keepers:   util.EthIDs{util.Signer2.Address()},
		Payload:   []byte(`42`),
		Approver:  &buyer,
		ID:        did,
	}
	assert.NoError(di.SyntacticVerify())
	assert.NoError(cs.SaveData(di))
	cs.CommitAccounts()
	assert.ErrorContains(itx.Apply(ctx, cs), "invalid version, expected 2, got 1")
	cs.CheckoutAccounts()

	input = &ld.TxUpdater{ID: &did, Version: 2, To: &owner,
		Amount: new(big.Int).SetUint64(constants.MilliLDC), Token: &token}
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeUpdateDataInfoByAuth,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      buyer,
		To:        &owner,
		Token:     &token,
		Amount:    new(big.Int).SetUint64(constants.MilliLDC),
		Data:      input.Bytes(),
	}}
	assert.NoError(ltx.SignWith(util.Signer1))
	assert.NoError(ltx.ExSignWith(util.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	itx, err = NewTx(ltx)
	assert.NoError(err)
	cs.CommitAccounts()
	assert.ErrorContains(itx.Apply(ctx, cs), "invalid exSignatures for data keepers")
	cs.CheckoutAccounts()

	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeUpdateDataInfoByAuth,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      buyer,
		To:        &owner,
		Token:     &token,
		Amount:    new(big.Int).SetUint64(constants.MilliLDC),
		Data:      input.Bytes(),
	}}
	assert.NoError(ltx.SignWith(util.Signer1))
	assert.NoError(ltx.ExSignWith(util.Signer2))
	assert.NoError(ltx.SyntacticVerify())
	itx, err = NewTx(ltx)
	assert.NoError(err)
	cs.CommitAccounts()
	assert.ErrorContains(itx.Apply(ctx, cs), "invalid signature for data approver")
	cs.CheckoutAccounts()

	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeUpdateDataInfoByAuth,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      buyer,
		To:        &owner,
		Token:     &token,
		Amount:    new(big.Int).SetUint64(constants.MilliLDC),
		Data:      input.Bytes(),
	}}
	assert.NoError(ltx.SignWith(util.Signer1))
	assert.NoError(ltx.ExSignWith(util.Signer1, util.Signer2))
	assert.NoError(ltx.SyntacticVerify())

	itx, err = NewTx(ltx)
	assert.NoError(err)
	assert.NoError(itx.Apply(ctx, cs))

	buyerGas := ltx.Gas()
	assert.Equal(buyerGas*ctx.Price,
		itx.(*TxUpdateDataInfoByAuth).ldc.Balance().Uint64())
	assert.Equal(buyerGas*100,
		itx.(*TxUpdateDataInfoByAuth).miner.Balance().Uint64())
	assert.Equal(constants.LDC-buyerGas*(ctx.Price+100),
		buyerAcc.Balance().Uint64())
	assert.Equal(constants.LDC-constants.MilliLDC, buyerAcc.balanceOf(token).Uint64())
	assert.Equal(constants.MilliLDC,
		itx.(*TxUpdateDataInfoByAuth).to.balanceOf(token).Uint64())
	assert.Equal(uint64(1), buyerAcc.Nonce())

	di2, err := cs.LoadData(di.ID)
	assert.NoError(err)
	assert.Equal(di.Version+1, di2.Version)
	assert.Equal(uint16(1), di2.Threshold)
	assert.Equal(util.EthIDs{buyer}, di2.Keepers)
	assert.Equal(di.Payload, di2.Payload)
	assert.Nil(di2.TypedSig)
	assert.Nil(di2.Approver)
	assert.Nil(di2.ApproveList)

	jsondata, err := itx.MarshalJSON()
	assert.NoError(err)
	// fmt.Println(string(jsondata))
	assert.Equal(`{"tx":{"type":"TypeUpdateDataInfoByAuth","chainID":2357,"nonce":0,"gasTip":100,"gasFeeCap":1000,"from":"0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC","to":"0x44171C37Ff5D7B7bb8dcad5C81f16284A229e641","token":"$LDC","amount":1000000,"data":{"id":"SkB92DD9M2yeCadw22VbnxfV6b7W5YEnnLRs6fKivk6wh2Zy","version":2,"token":"$LDC","to":"0x44171C37Ff5D7B7bb8dcad5C81f16284A229e641","amount":1000000}},"sigs":["8c8c1b663eba2435e1ae8882516ed3738a8b2c5b1733667c43d65379d448827b43461acad4d21a7db062ae90b4315066e40b0e0c16ddfa69920e722f137d301700"],"exSigs":["1b207b020f679fec178e6430960f58626eb55f56bb6e056351f35b3db34e9cb773e9a6d720174e2c6e81738d11c8d32b5d2d7bdf08f2e5c3f5988251800eaf5100","e89d959e0add6c27a44fb48f5030bddd603dbd3298c6dbeb2815692c2067c6f25c9020664ebe1d94100d0a6587e0ec4359489313fc88b6e4ad144cff61f58a3800"],"id":"UvvyfWCWKyioR5t7LYGh92jC7w12NWDBWtyhXwiwnKt3GEhFA"}`, string(jsondata))

	assert.NoError(cs.VerifyState())
}
