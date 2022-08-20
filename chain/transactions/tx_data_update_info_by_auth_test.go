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

	txData := &ld.TxData{
		Type:      ld.TypeUpdateDataInfoByAuth,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      buyer,
	}
	assert.NoError(txData.SyntacticVerify())
	_, err = NewTx2(txData.ToTransaction())
	assert.ErrorContains(err, "DeriveSigners error: no signature")

	txData = &ld.TxData{
		Type:      ld.TypeUpdateDataInfoByAuth,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      buyer,
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx2(txData.ToTransaction())
	assert.ErrorContains(err, "nil to")

	txData = &ld.TxData{
		Type:      ld.TypeUpdateDataInfoByAuth,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      buyer,
		To:        &owner,
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx2(txData.ToTransaction())
	assert.ErrorContains(err, "invalid data")

	txData = &ld.TxData{
		Type:      ld.TypeUpdateDataInfoByAuth,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      buyer,
		To:        &owner,
		Data:      []byte("你好👋"),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx2(txData.ToTransaction())
	assert.ErrorContains(err, "cbor: unexpected following extraneous data")

	input := &ld.TxUpdater{}
	txData = &ld.TxData{
		Type:      ld.TypeUpdateDataInfoByAuth,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      buyer,
		To:        &owner,
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx2(txData.ToTransaction())
	assert.ErrorContains(err, "invalid data id")

	input = &ld.TxUpdater{ID: &util.DataIDEmpty}
	txData = &ld.TxData{
		Type:      ld.TypeUpdateDataInfoByAuth,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      buyer,
		To:        &owner,
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx2(txData.ToTransaction())
	assert.ErrorContains(err, "invalid data id")

	did := util.DataID{1, 2, 3, 4}
	input = &ld.TxUpdater{ID: &did}
	txData = &ld.TxData{
		Type:      ld.TypeUpdateDataInfoByAuth,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      buyer,
		To:        &owner,
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx2(txData.ToTransaction())
	assert.ErrorContains(err, "invalid data version")

	input = &ld.TxUpdater{ID: &did, Version: 1,
		Keepers:   &util.EthIDs{util.Signer1.Address()},
		Threshold: ld.Uint16Ptr(1)}
	txData = &ld.TxData{
		Type:      ld.TypeUpdateDataInfoByAuth,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      buyer,
		To:        &owner,
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx2(txData.ToTransaction())
	assert.ErrorContains(err, "invalid keepers, should be nil")

	input = &ld.TxUpdater{
		ID: &did, Version: 1,
		Sig: &util.Signature{1, 2, 3},
		SigClaims: &ld.SigClaims{
			Issuer:     util.DataID{1, 2, 3, 4},
			Subject:    util.DataID{5, 6, 7, 8},
			Audience:   ld.RawModelID,
			Expiration: 100,
			IssuedAt:   1,
			CWTID:      util.Hash{9, 10, 11, 12},
		},
	}
	txData = &ld.TxData{
		Type:      ld.TypeUpdateDataInfoByAuth,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      buyer,
		To:        &owner,
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx2(txData.ToTransaction())
	assert.ErrorContains(err, "invalid sigClaims, should be nil")

	input = &ld.TxUpdater{ID: &did, Version: 1, Approver: &constants.GenesisAccount}
	txData = &ld.TxData{
		Type:      ld.TypeUpdateDataInfoByAuth,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      buyer,
		To:        &owner,
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx2(txData.ToTransaction())
	assert.ErrorContains(err, "invalid approver, should be nil")

	input = &ld.TxUpdater{ID: &did, Version: 1,
		ApproveList: []ld.TxType{ld.TypeUpdateDataInfoByAuth}}
	txData = &ld.TxData{
		Type:      ld.TypeUpdateDataInfoByAuth,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      buyer,
		To:        &owner,
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx2(txData.ToTransaction())
	assert.ErrorContains(err, "invalid approveList, should be nil")

	input = &ld.TxUpdater{ID: &did, Version: 1}
	txData = &ld.TxData{
		Type:      ld.TypeUpdateDataInfoByAuth,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      buyer,
		To:        &owner,
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx2(txData.ToTransaction())
	assert.ErrorContains(err, "nil to")

	input = &ld.TxUpdater{ID: &did, Version: 1, To: &owner}
	txData = &ld.TxData{
		Type:      ld.TypeUpdateDataInfoByAuth,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      buyer,
		To:        &constants.GenesisAccount,
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx2(txData.ToTransaction())
	assert.ErrorContains(err,
		"invalid to, expected 0x44171C37Ff5D7B7bb8dcad5C81f16284A229e641, got 0xFFfFfFffFFfffFFfFFfFFFFFffFFFffffFfFFFfF")

	input = &ld.TxUpdater{ID: &did, Version: 1, To: &owner}
	txData = &ld.TxData{
		Type:      ld.TypeUpdateDataInfoByAuth,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      buyer,
		To:        &owner,
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx2(txData.ToTransaction())
	assert.ErrorContains(err, "nil amount")

	input = &ld.TxUpdater{ID: &did, Version: 1, To: &owner,
		Amount: new(big.Int).SetUint64(constants.MilliLDC)}
	txData = &ld.TxData{
		Type:      ld.TypeUpdateDataInfoByAuth,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      buyer,
		To:        &owner,
		Amount:    new(big.Int).SetUint64(1),
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx2(txData.ToTransaction())
	assert.ErrorContains(err, "invalid amount, expected 1000000, got 1")

	txData = &ld.TxData{
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
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx2(txData.ToTransaction())
	assert.ErrorContains(err, "invalid token, expected NativeToken, got $LDC")

	input = &ld.TxUpdater{ID: &did, Version: 1, To: &owner,
		Amount: new(big.Int).SetUint64(constants.MilliLDC), Token: &token}

	txData = &ld.TxData{
		Type:      ld.TypeUpdateDataInfoByAuth,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      buyer,
		To:        &owner,
		Amount:    new(big.Int).SetUint64(constants.MilliLDC),
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx2(txData.ToTransaction())
	assert.ErrorContains(err, "invalid token, expected $LDC, got NativeLDC")

	txData = &ld.TxData{
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
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx2(txData.ToTransaction())
	assert.ErrorContains(err,
		"invalid exSignatures, Transaction.ExSigners error: DeriveSigners error: no signature")

	assert.NoError(txData.ExSignWith(util.Signer1))
	tt := txData.ToTransaction()
	itx, err := NewTx2(tt)
	assert.NoError(err)

	cs.CommitAccounts()
	assert.ErrorContains(itx.Apply(ctx, cs),
		"insufficient NativeLDC balance, expected 1984400, got 0")
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
		"LD6L5yRJL2iYi9PbrhRru6uKfEAzDGHwUJ not found")
	cs.CheckoutAccounts()

	di := &ld.DataInfo{
		ModelID:   ld.RawModelID,
		Version:   2,
		Threshold: 1,
		Keepers:   util.EthIDs{util.Signer2.Address()},
		Data:      []byte(`42`),
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
	txData = &ld.TxData{
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
	}
	assert.NoError(txData.SignWith(util.Signer1))
	assert.NoError(txData.ExSignWith(util.Signer1))
	tt = txData.ToTransaction()
	itx, err = NewTx2(tt)
	assert.NoError(err)
	cs.CommitAccounts()
	assert.ErrorContains(itx.Apply(ctx, cs), "invalid exSignatures for data keepers")
	cs.CheckoutAccounts()

	txData = &ld.TxData{
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
	}
	assert.NoError(txData.SignWith(util.Signer1))
	assert.NoError(txData.ExSignWith(util.Signer2))
	tt = txData.ToTransaction()
	itx, err = NewTx2(tt)
	assert.NoError(err)
	cs.CommitAccounts()
	assert.ErrorContains(itx.Apply(ctx, cs), "invalid signature for data approver")
	cs.CheckoutAccounts()

	txData = &ld.TxData{
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
	}
	assert.NoError(txData.SignWith(util.Signer1))
	assert.NoError(txData.ExSignWith(util.Signer1))
	assert.NoError(txData.ExSignWith(util.Signer2))
	tt = txData.ToTransaction()
	itx, err = NewTx2(tt)
	assert.NoError(err)
	assert.NoError(itx.Apply(ctx, cs))

	buyerGas := tt.Gas()
	assert.Equal(buyerGas*ctx.Price,
		itx.(*TxUpdateDataInfoByAuth).ldc.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(buyerGas*100,
		itx.(*TxUpdateDataInfoByAuth).miner.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(constants.LDC-buyerGas*(ctx.Price+100),
		buyerAcc.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(constants.LDC-constants.MilliLDC, buyerAcc.balanceOf(token).Uint64())
	assert.Equal(constants.MilliLDC,
		itx.(*TxUpdateDataInfoByAuth).to.balanceOf(token).Uint64())
	assert.Equal(uint64(1), buyerAcc.Nonce())

	di2, err := cs.LoadData(di.ID)
	assert.NoError(err)
	assert.Equal(di.Version+1, di2.Version)
	assert.Equal(uint16(1), di2.Threshold)
	assert.Equal(util.EthIDs{buyer}, di2.Keepers)
	assert.Equal(di.Data, di2.Data)
	assert.Nil(di2.Sig)
	assert.Nil(di2.Approver)
	assert.Nil(di2.ApproveList)

	jsondata, err := itx.MarshalJSON()
	assert.NoError(err)
	// fmt.Println(string(jsondata))
	assert.Equal(`{"type":"TypeUpdateDataInfoByAuth","chainID":2357,"nonce":0,"gasTip":100,"gasFeeCap":1000,"from":"0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC","to":"0x44171C37Ff5D7B7bb8dcad5C81f16284A229e641","token":"$LDC","amount":1000000,"data":{"id":"LD6L5yRJL2iYi9PbrhRru6uKfEAzDGHwUJ","version":2,"token":"$LDC","to":"0x44171C37Ff5D7B7bb8dcad5C81f16284A229e641","amount":1000000},"signatures":["0c245068bae0fd48c5080a5b22ede3e241eabda09cd1b995945e6685cd40b3886b1d661e0e3ba959c3d74a1cad580ceaa6b2a35d87448eebc38479fc252e7a7e00"],"exSignatures":["6a899b77c48dded7b87e374f111368bf56c49e3b7fd1d8329147721bb393b5d810e45732973137ad372bdebf28e5e0168c883957d63577b34f537d7527b4457e01","ce51288efa3bfa119530759b0d1a19ef6c0c20323686b1bd5c2bdd0c09cc3cb574ffda96e0be8bdcd1e24d8a7d31cb737cb376e5d637bb0dd6a4b9d3e253f8c000"],"id":"2TWQe2HNMuJGkfW5ANaAp1yHTZiWQLZKFqKScM1SWx3J6XDetP"}`, string(jsondata))

	assert.NoError(cs.VerifyState())
}
