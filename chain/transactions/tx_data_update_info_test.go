// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package transactions

import (
	"math/big"
	"testing"

	"github.com/ldclabs/ldvm/constants"
	"github.com/ldclabs/ldvm/ld"
	"github.com/ldclabs/ldvm/util"
	"github.com/ldclabs/ldvm/util/signer"
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
	owner := signer.Signer1.Key().Address()

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
	assert.ErrorContains(err, "no signatures")

	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeUpdateDataInfo,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      owner,
		To:        signer.Signer2.Key().Address().Ptr(),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
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
	assert.NoError(ltx.SignWith(signer.Signer1))
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
	assert.NoError(ltx.SignWith(signer.Signer1))
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
	assert.NoError(ltx.SignWith(signer.Signer1))
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
	assert.NoError(ltx.SignWith(signer.Signer1))
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
	assert.NoError(ltx.SignWith(signer.Signer1))
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
	assert.NoError(ltx.SignWith(signer.Signer1))
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
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "no thing to update")

	di := &ld.DataInfo{
		ModelID:   ld.RawModelID,
		Version:   2,
		Threshold: 1,
		Keepers:   signer.Keys{signer.Signer2.Key()},
		Payload:   []byte(`42`),
		ID:        did,
	}
	assert.NoError(di.SyntacticVerify())

	input = &ld.TxUpdater{ID: &did, Version: 1,
		Approver:    signer.Signer2.Key().Ptr(),
		ApproveList: &ld.TxTypes{ld.TypeDeleteData},
		Threshold:   ld.Uint16Ptr(1),
		Keepers:     &signer.Keys{signer.Signer1.Key()},
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
	assert.NoError(ltx.SignWith(signer.Signer1))
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
		"AQIDBAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAACs148t not found")
	cs.CheckoutAccounts()

	assert.NoError(cs.SaveData(di))
	cs.CommitAccounts()
	assert.ErrorContains(itx.Apply(ctx, cs),
		"invalid version, expected 2, got 1")
	cs.CheckoutAccounts()

	input = &ld.TxUpdater{ID: &did, Version: 2,
		Approver:    signer.Signer2.Key().Ptr(),
		ApproveList: &ld.TxTypes{ld.TypeDeleteData},
		Threshold:   ld.Uint16Ptr(1),
		Keepers:     &signer.Keys{signer.Signer1.Key()},
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
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())

	itx, err = NewTx(ltx)
	assert.NoError(err)
	cs.CommitAccounts()
	assert.ErrorContains(itx.Apply(ctx, cs), "invalid signatures for data keepers")
	cs.CheckoutAccounts()

	input = &ld.TxUpdater{ID: &did, Version: 2,
		Approver: signer.Signer2.Key().Ptr(),
		ApproveList: &ld.TxTypes{
			ld.TypeUpdateDataInfo,
			ld.TypeUpdateDataInfoByAuth,
			ld.TypeDeleteData},
		Threshold: ld.Uint16Ptr(1),
		Keepers:   &signer.Keys{signer.Signer1.Key()},
		SigClaims: &ld.SigClaims{
			Issuer:     util.DataID{1, 2, 3, 4},
			Subject:    di.ID,
			Audience:   di.ModelID,
			Expiration: 100,
			IssuedAt:   1,
			CWTID:      util.HashFromData(di.Payload),
		},
	}
	input.Sig = signer.Signer2.MustSignData(input.SigClaims.Bytes()).Ptr()

	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeUpdateDataInfo,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      owner,
		Data:      input.Bytes(),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1, signer.Signer2))
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
	assert.Equal(signer.Keys{signer.Signer2.Key()}, di.Keepers)
	assert.Equal(signer.Keys{signer.Signer1.Key()}, di2.Keepers)
	assert.Equal(di.Payload, di2.Payload)
	assert.Nil(di.Approver)
	assert.NotNil(di2.Approver)
	assert.True(signer.Signer2.Key().Equal(di2.Approver))
	assert.Nil(di.ApproveList)
	assert.Equal(ld.TxTypes{
		ld.TypeUpdateDataInfo,
		ld.TypeUpdateDataInfoByAuth,
		ld.TypeDeleteData}, di2.ApproveList)

	jsondata, err := itx.MarshalJSON()
	assert.NoError(err)
	// fmt.Println(string(jsondata))
	assert.Equal(`{"tx":{"type":"TypeUpdateDataInfo","chainID":2357,"nonce":0,"gasTip":100,"gasFeeCap":1000,"from":"0x8db97c7cECe249C2b98bdc0226cc4C2A57bF52fc","data":{"id":"AQIDBAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAACs148t","version":2,"threshold":1,"keepers":["jbl8fOziScK5i9wCJsxMKle_UvwKxwPH"],"approver":"RBccN_9de3u43K1cgfFihKIp5kE1lmGG","approveList":["TypeUpdateDataInfo","TypeUpdateDataInfoByAuth","TypeDeleteData"],"sigClaims":{"iss":"AQIDBAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAACs148t","sub":"AQIDBAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAACs148t","aud":"AAAAAAAAAAAAAAAAAAAAAAAAAADzaDye","exp":100,"nbf":0,"iat":1,"cti":"Thad30ecjNm0xF4ChBgXMMtC3zqL6JKn83m9aQ2x6vrOudt9"},"sig":"Gg4e8InrIr9DuJXe7CJf6zlg71U48gv0R4DhHi13OiMpRF9d5GxC-nIXZAGo1zx-E0v-ucjjzvEkv2iW5dppUgCF1by8"}},"sigs":["Xjq0_tGT4AnE6nKpag-FqB2eJHlBwXpWsjYhudnW-gQ1OGgFKQCTrWliZK1T2qPZVb_pTUF3aEtltc197dGdnwDDe55m","t5cwBzgvxm-2hAUV4lioaQ0APvacyHe-ARDbppivWV4UKk0nQL5r-O1-KBYpRIq6YZsO41uqe3mT0PQYU2QPXwA2Kbnj"],"id":"_SOjtobO5nph-igJ9dITxbRrjeCCX75mNmJo8gkYE2vuBKGB"}`, string(jsondata))

	input = &ld.TxUpdater{ID: &did, Version: 3,
		Approver: &signer.Key{},
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
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())

	itx, err = NewTx(ltx)
	assert.NoError(err)
	cs.CommitAccounts()
	assert.ErrorContains(itx.Apply(ctx, cs), "invalid signature for data approver")
	cs.CheckoutAccounts()

	assert.NoError(ltx.SignWith(signer.Signer1, signer.Signer2))
	assert.NoError(ltx.SyntacticVerify())

	itx, err = NewTx(ltx)
	assert.NoError(err)
	assert.NoError(itx.Apply(ctx, cs))

	ownerGas += ltx.Gas()
	di2, err = cs.LoadData(di.ID)
	assert.NoError(err)
	assert.Equal(uint64(4), di2.Version)
	assert.Equal(uint16(1), di2.Threshold)
	assert.Equal(signer.Keys{signer.Signer1.Key()}, di2.Keepers)
	assert.Equal(di.Payload, di2.Payload)
	assert.Nil(di2.Approver)
	assert.Nil(di2.ApproveList)

	// clear threshold
	input = &ld.TxUpdater{ID: &did, Version: 4,
		Threshold: ld.Uint16Ptr(0),
		Keepers:   &signer.Keys{signer.Signer1.Key()},
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
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())

	itx, err = NewTx(ltx)
	assert.NoError(err)
	assert.NoError(itx.Apply(ctx, cs))

	ownerGas += ltx.Gas()
	di2, err = cs.LoadData(di.ID)
	assert.NoError(err)
	assert.Equal(uint64(5), di2.Version)
	assert.Equal(uint16(0), di2.Threshold)
	assert.Equal(signer.Keys{signer.Signer1.Key()}, di2.Keepers)

	// clear keepers
	input = &ld.TxUpdater{ID: &did, Version: 5,
		Threshold: ld.Uint16Ptr(0),
		Keepers:   &signer.Keys{},
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
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())

	itx, err = NewTx(ltx)
	assert.NoError(err)
	assert.NoError(itx.Apply(ctx, cs))

	ownerGas += ltx.Gas()
	di2, err = cs.LoadData(di.ID)
	assert.NoError(err)
	assert.Equal(uint64(6), di2.Version)
	assert.Equal(uint16(0), di2.Threshold)
	assert.Equal(signer.Keys{}, di2.Keepers)

	// can't update keepers
	input = &ld.TxUpdater{ID: &did, Version: 6,
		Threshold: ld.Uint16Ptr(0),
		Keepers:   &signer.Keys{signer.Signer1.Key()},
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
	assert.NoError(ltx.SignWith(signer.Signer1))
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
	assert.ErrorContains(err, "no signatures")

	assert.NoError(ltx.SignWith(signer.Signer1))
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
	input.Sig = signer.Signer2.MustSignData(input.SigClaims.Bytes()).Ptr()
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
	assert.ErrorContains(err, "no signatures")

	assert.NoError(ltx.SignWith(signer.Signer1))
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
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	itx, err = NewTx(ltx)
	assert.NoError(err)
	assert.ErrorContains(itx.Apply(ctx, cs), "invalid signatures for data keepers")

	assert.NoError(cs.VerifyState())
}
