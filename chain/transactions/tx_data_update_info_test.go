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

func TestTxUpdateDataInfo(t *testing.T) {
	assert := assert.New(t)

	// SyntacticVerify
	tx := &TxUpdateDataInfo{}
	assert.ErrorContains(tx.SyntacticVerify(), "nil pointer")
	_, err := tx.MarshalJSON()
	assert.NoError(err)

	ctx := NewMockChainContext()
	cs := ctx.MockChainState()

	owner := util.Signer1.Address()
	assert.NoError(err)
	approver := util.Signer2.Address()

	ltx := &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeUpdateDataInfo,
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
		Type:      ld.TypeUpdateDataInfo,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      owner,
		To:        &approver,
	}}
	assert.NoError(ltx.SignWith(util.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "invalid to, should be nil")

	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeUpdateDataInfo,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      owner,
		Token:     &constants.NativeToken,
	}}
	assert.NoError(ltx.SignWith(util.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "invalid token, should be nil")

	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeUpdateDataInfo,
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
		Type:      ld.TypeUpdateDataInfo,
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
		Type:      ld.TypeUpdateDataInfo,
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
		Type:      ld.TypeUpdateDataInfo,
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
		Type:      ld.TypeUpdateDataInfo,
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

	input = &ld.TxUpdater{ID: &did, Version: 1}
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeUpdateDataInfo,
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
	assert.ErrorContains(err, "no thing to update")

	di := &ld.DataInfo{
		ModelID:   ld.RawModelID,
		Version:   2,
		Threshold: 1,
		Keepers:   util.EthIDs{util.Signer2.Address()},
		Payload:   []byte(`42`),
		ID:        did,
	}
	assert.NoError(di.SyntacticVerify())

	input = &ld.TxUpdater{ID: &did, Version: 1,
		Approver:    &approver,
		ApproveList: []ld.TxType{ld.TypeDeleteData},
		Threshold:   ld.Uint16Ptr(1),
		Keepers:     &util.EthIDs{util.Signer1.Address()},
	}
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeUpdateDataInfo,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      owner,
		Data:      input.Bytes(),
	}}
	assert.NoError(ltx.SignWith(util.Signer1))
	assert.NoError(ltx.SyntacticVerify())

	itx, err := NewTx(ltx)
	assert.NoError(err)

	cs.CommitAccounts()
	assert.ErrorContains(itx.Apply(ctx, cs),
		"insufficient NativeLDC balance, expected 1137400, got 0")
	cs.CheckoutAccounts()

	ownerAcc := cs.MustAccount(owner)
	ownerAcc.Add(constants.NativeToken, new(big.Int).SetUint64(constants.LDC))
	cs.CommitAccounts()
	assert.ErrorContains(itx.Apply(ctx, cs),
		"SkB92DD9M2yeCadw22VbnxfV6b7W5YEnnLRs6fKivk6wh2Zy not found")
	cs.CheckoutAccounts()

	assert.NoError(cs.SaveData(di))
	cs.CommitAccounts()
	assert.ErrorContains(itx.Apply(ctx, cs),
		"invalid version, expected 2, got 1")
	cs.CheckoutAccounts()

	input = &ld.TxUpdater{ID: &did, Version: 2,
		Approver:    &approver,
		ApproveList: []ld.TxType{ld.TypeDeleteData},
		Threshold:   ld.Uint16Ptr(1),
		Keepers:     &util.EthIDs{util.Signer1.Address()},
	}
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeUpdateDataInfo,
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
	assert.ErrorContains(itx.Apply(ctx, cs), "invalid signatures for data keepers")
	cs.CheckoutAccounts()

	input = &ld.TxUpdater{ID: &did, Version: 2,
		Approver: &approver,
		ApproveList: []ld.TxType{
			ld.TypeUpdateDataInfo,
			ld.TypeUpdateDataInfoByAuth,
			ld.TypeDeleteData},
		Threshold: ld.Uint16Ptr(1),
		Keepers:   &util.EthIDs{util.Signer1.Address()},
		SigClaims: &ld.SigClaims{
			Issuer:     util.DataID{1, 2, 3, 4},
			Subject:    di.ID,
			Audience:   di.ModelID,
			Expiration: 100,
			IssuedAt:   1,
			CWTID:      util.HashFromData(di.Payload),
		},
	}
	sig, err := util.Signer2.SignData(input.SigClaims.Bytes())
	assert.NoError(err)
	input.TypedSig = sig.Typed()

	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeUpdateDataInfo,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      owner,
		Data:      input.Bytes(),
	}}
	assert.NoError(ltx.SignWith(util.Signer1, util.Signer2))
	assert.NoError(ltx.SyntacticVerify())

	itx, err = NewTx(ltx)
	assert.NoError(err)
	assert.NoError(itx.Apply(ctx, cs))

	ownerGas := ltx.Gas()
	assert.Equal(ownerGas*ctx.Price,
		itx.(*TxUpdateDataInfo).ldc.Balance().Uint64())
	assert.Equal(ownerGas*100,
		itx.(*TxUpdateDataInfo).miner.Balance().Uint64())
	assert.Equal(constants.LDC-ownerGas*(ctx.Price+100),
		ownerAcc.Balance().Uint64())
	assert.Equal(uint64(1), ownerAcc.Nonce())

	di2, err := cs.LoadData(di.ID)
	assert.NoError(err)
	assert.Equal(di.Version+1, di2.Version)
	assert.Equal(uint16(1), di2.Threshold)
	assert.Equal(util.EthIDs{util.Signer2.Address()}, di.Keepers)
	assert.Equal(util.EthIDs{util.Signer1.Address()}, di2.Keepers)
	assert.Equal(di.Payload, di2.Payload)
	assert.Nil(di.Approver)
	assert.NotNil(di2.Approver)
	assert.Equal(util.Signer2.Address(), *di2.Approver)
	assert.Nil(di.ApproveList)
	assert.Equal(ld.TxTypes{
		ld.TypeUpdateDataInfo,
		ld.TypeUpdateDataInfoByAuth,
		ld.TypeDeleteData}, di2.ApproveList)

	jsondata, err := itx.MarshalJSON()
	assert.NoError(err)
	// fmt.Println(string(jsondata))
	assert.Equal(`{"tx":{"type":"TypeUpdateDataInfo","chainID":2357,"nonce":0,"gasTip":100,"gasFeeCap":1000,"from":"0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC","data":{"id":"SkB92DD9M2yeCadw22VbnxfV6b7W5YEnnLRs6fKivk6wh2Zy","version":2,"threshold":1,"keepers":["0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC"],"approver":"0x44171C37Ff5D7B7bb8dcad5C81f16284A229e641","approveList":["TypeUpdateDataInfo","TypeUpdateDataInfoByAuth","TypeDeleteData"],"sigClaims":{"iss":"SkB92DD9M2yeCadw22VbnxfV6b7W5YEnnLRs6fKivk6wh2Zy","sub":"SkB92DD9M2yeCadw22VbnxfV6b7W5YEnnLRs6fKivk6wh2Zy","aud":"111111111111111111116DBWJs","exp":100,"nbf":0,"iat":1,"cti":"bPfPyx1epFu5ges4t55fJYpS4HrQFr5zCPfT3mzNmL8XCDAZP"},"typedSig":"0x001a0e1ef089eb22bf43b895deec225feb3960ef5538f20bf44780e11e2d773a2329445f5de46c42fa72176401a8d73c7e134bfeb9c8e3cef124bf6896e5da695200de1abfe8"}},"sigs":["3777101d23c2dcce66521e501325fca5921e8d390f62dc9a087c85ccc3c8fe792f637431610cba647236b0cf6c2652fec66e3dc988533000aa5d530ea13edbbf00","2e3a5a2c70753bd5cf217669816fa77c06a530d5f59742939b59886f23048a840f29d12b5b2782ed36cfef70c0e507ca29822fa63692efa34e39dc9c0ffa36b300"],"id":"GF1jp69xJTKiAbhrxgzJvpzdgVmR5EjsDBr8g4mh8ftCnJsVr"}`, string(jsondata))

	input = &ld.TxUpdater{ID: &did, Version: 3,
		Approver: &util.EthIDEmpty,
	}
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeUpdateDataInfo,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     1,
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
	assert.ErrorContains(itx.Apply(ctx, cs), "invalid signature for data approver")
	cs.CheckoutAccounts()

	assert.NoError(ltx.SignWith(util.Signer1, util.Signer2))
	assert.NoError(ltx.SyntacticVerify())

	itx, err = NewTx(ltx)
	assert.NoError(err)
	assert.NoError(itx.Apply(ctx, cs))

	ownerGas += ltx.Gas()
	di2, err = cs.LoadData(di.ID)
	assert.NoError(err)
	assert.Equal(uint64(4), di2.Version)
	assert.Equal(uint16(1), di2.Threshold)
	assert.Equal(util.EthIDs{util.Signer1.Address()}, di2.Keepers)
	assert.Equal(di.Payload, di2.Payload)
	assert.Nil(di2.Approver)
	assert.Nil(di2.ApproveList)

	// clear threshold
	input = &ld.TxUpdater{ID: &did, Version: 4,
		Threshold: ld.Uint16Ptr(0),
		Keepers:   &util.EthIDs{util.Signer1.Address()},
	}
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeUpdateDataInfo,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     2,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      owner,
		Data:      input.Bytes(),
	}}
	assert.NoError(ltx.SignWith(util.Signer1))
	assert.NoError(ltx.SyntacticVerify())

	itx, err = NewTx(ltx)
	assert.NoError(err)
	assert.NoError(itx.Apply(ctx, cs))

	ownerGas += ltx.Gas()
	di2, err = cs.LoadData(di.ID)
	assert.NoError(err)
	assert.Equal(uint64(5), di2.Version)
	assert.Equal(uint16(0), di2.Threshold)
	assert.Equal(util.EthIDs{util.Signer1.Address()}, di2.Keepers)

	// clear keepers
	input = &ld.TxUpdater{ID: &did, Version: 5,
		Threshold: ld.Uint16Ptr(0),
		Keepers:   &util.EthIDs{},
	}
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeUpdateDataInfo,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     3,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      owner,
		Data:      input.Bytes(),
	}}
	assert.NoError(ltx.SignWith(util.Signer1))
	assert.NoError(ltx.SyntacticVerify())

	itx, err = NewTx(ltx)
	assert.NoError(err)
	assert.NoError(itx.Apply(ctx, cs))

	ownerGas += ltx.Gas()
	di2, err = cs.LoadData(di.ID)
	assert.NoError(err)
	assert.Equal(uint64(6), di2.Version)
	assert.Equal(uint16(0), di2.Threshold)
	assert.Equal(util.EthIDs{}, di2.Keepers)

	// can't update keepers
	input = &ld.TxUpdater{ID: &did, Version: 6,
		Threshold: ld.Uint16Ptr(0),
		Keepers:   &util.EthIDs{util.Signer1.Address()},
	}
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeUpdateDataInfo,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     4,
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
	assert.ErrorContains(itx.Apply(ctx, cs), "invalid signatures for data keepers")
	cs.CheckoutAccounts()

	// can't update data
	input = &ld.TxUpdater{ID: &did, Version: 6,
		Data: []byte(`421`),
	}
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeUpdateData,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     4,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      owner,
		Data:      input.Bytes(),
	}}
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "DeriveSigners error: no signature")

	assert.NoError(ltx.SignWith(util.Signer1))
	assert.NoError(ltx.SyntacticVerify())

	itx, err = NewTx(ltx)
	assert.NoError(err)
	cs.CommitAccounts()
	assert.ErrorContains(itx.Apply(ctx, cs), "invalid sigClaims, should not be nil")
	cs.CheckoutAccounts()

	// can't update data
	input = &ld.TxUpdater{ID: &did, Version: 6,
		Data: []byte(`421`),
		SigClaims: &ld.SigClaims{
			Issuer:     util.DataID{1, 2, 3, 4},
			Subject:    di.ID,
			Audience:   di.ModelID,
			Expiration: 100,
			IssuedAt:   1,
			CWTID:      util.HashFromData([]byte(`421`)),
		},
	}
	sig, err = util.Signer2.SignData(input.SigClaims.Bytes())
	assert.NoError(err)
	input.TypedSig = sig.Typed()

	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeUpdateData,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     4,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      owner,
		Data:      input.Bytes(),
	}}
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "DeriveSigners error: no signature")

	assert.NoError(ltx.SignWith(util.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	itx, err = NewTx(ltx)
	assert.NoError(err)
	cs.CommitAccounts()
	assert.ErrorContains(itx.Apply(ctx, cs), "invalid signatures for data keepers")
	cs.CheckoutAccounts()

	// can't delete data
	input = &ld.TxUpdater{ID: &did, Version: 6, Data: []byte(`421`)}
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeDeleteData,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     4,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      owner,
		Data:      input.Bytes(),
	}}
	assert.NoError(ltx.SignWith(util.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	itx, err = NewTx(ltx)
	assert.NoError(err)
	assert.ErrorContains(itx.Apply(ctx, cs), "invalid signatures for data keepers")

	assert.NoError(cs.VerifyState())
}
