// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package transactions

import (
	"encoding/json"
	"math/big"
	"testing"

	jsonpatch "github.com/ldclabs/json-patch"
	"github.com/ldclabs/ldvm/constants"
	"github.com/ldclabs/ldvm/ld"
	"github.com/ldclabs/ldvm/ld/service"
	"github.com/ldclabs/ldvm/util"
	"github.com/ldclabs/ldvm/util/signer"
	"github.com/stretchr/testify/assert"
)

func TestTxCreateData(t *testing.T) {
	assert := assert.New(t)

	// SyntacticVerify
	tx := &TxCreateData{}
	assert.ErrorContains(tx.SyntacticVerify(), "nil pointer")
	_, err := tx.MarshalJSON()
	assert.NoError(err)

	ctx := NewMockChainContext()
	cs := ctx.MockChainState()
	token := ld.MustNewToken("$LDC")
	sender := signer.Signer1.Key().Address()

	ltx := &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeCreateData,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      sender,
	}}
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "no signatures")

	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeCreateData,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      sender,
		Token:     &token,
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "invalid token, should be nil")

	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeCreateData,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      sender,
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "invalid data")

	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeCreateData,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      sender,
		Data:      []byte("你好👋"),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "cbor: unexpected following extraneous data")

	input := &ld.TxUpdater{}
	ltx = &ld.Transaction{Tx: ld.TxData{
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
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "nil mid")

	input = &ld.TxUpdater{
		ModelID: &ld.RawModelID,
	}
	ltx = &ld.Transaction{Tx: ld.TxData{
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
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "invalid version, expected 1, got 0")

	input = &ld.TxUpdater{
		ModelID: &ld.RawModelID,
		Version: 1,
	}
	ltx = &ld.Transaction{Tx: ld.TxData{
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
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "nil threshold")

	input = &ld.TxUpdater{
		ModelID:   &ld.RawModelID,
		Version:   1,
		Threshold: ld.Uint16Ptr(0),
	}
	ltx = &ld.Transaction{Tx: ld.TxData{
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
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "no keepers, threshold should be nil")

	input = &ld.TxUpdater{
		ModelID:   &ld.RawModelID,
		Version:   1,
		Threshold: ld.Uint16Ptr(0),
		Keepers:   &signer.Keys{},
	}
	ltx = &ld.Transaction{Tx: ld.TxData{
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
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "empty keepers")

	input = &ld.TxUpdater{
		ModelID:   &ld.RawModelID,
		Version:   1,
		Threshold: ld.Uint16Ptr(0),
		Keepers:   &signer.Keys{signer.Signer1.Key()},
	}
	ltx = &ld.Transaction{Tx: ld.TxData{
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
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "empty data")

	input = &ld.TxUpdater{
		ModelID:   &ld.RawModelID,
		Version:   1,
		Threshold: ld.Uint16Ptr(0),
		Keepers:   &signer.Keys{signer.Signer1.Key()},
		Data:      []byte(`{}`),
	}

	assert.NoError(input.SyntacticVerify())
	ltx = &ld.Transaction{Tx: ld.TxData{
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
	_, err = NewTx(ltx)
	assert.NoError(err)

	// RawModel
	input = &ld.TxUpdater{
		ModelID:   &ld.RawModelID,
		Version:   1,
		Threshold: ld.Uint16Ptr(0),
		Keepers:   &signer.Keys{signer.Signer1.Key()},
		Data:      []byte(`42`),
	}
	assert.NoError(input.SyntacticVerify())
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeCreateData,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      sender,
		To:        &constants.GenesisAccount,
		Data:      input.Bytes(),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "invalid to, should be nil")

	input = &ld.TxUpdater{
		ModelID:   &ld.RawModelID,
		Version:   1,
		Threshold: ld.Uint16Ptr(0),
		Keepers:   &signer.Keys{signer.Signer1.Key()},
		Data:      []byte(`42`),
	}
	assert.NoError(input.SyntacticVerify())
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeCreateData,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      sender,
		Amount:    big.NewInt(1),
		Data:      input.Bytes(),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.ErrorContains(ltx.SyntacticVerify(), "nil \"to\" together with amount")

	input = &ld.TxUpdater{
		ModelID:   &ld.RawModelID,
		Version:   1,
		Threshold: ld.Uint16Ptr(0),
		Keepers:   &signer.Keys{signer.Signer1.Key()},
		Data:      []byte(`42`),
	}
	assert.NoError(input.SyntacticVerify())
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeCreateData,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      sender,
		Data:      input.Bytes(),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.ExSignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "invalid exSignatures, should be nil")

	sig := make(signer.Sig, 65)
	input = &ld.TxUpdater{
		ModelID:   &ld.RawModelID,
		Version:   1,
		Threshold: ld.Uint16Ptr(0),
		Keepers:   &signer.Keys{signer.Signer1.Key()},
		Data:      []byte(`42`),
		Sig:       &sig,
		SigClaims: &ld.SigClaims{
			Issuer:     util.DataID{1, 2, 3, 4},
			Subject:    util.DataID{5, 6, 7, 8},
			Audience:   ld.RawModelID,
			Expiration: 100,
			IssuedAt:   1,
			CWTID:      util.Hash{9, 10, 11, 12},
		},
	}
	assert.NoError(input.SyntacticVerify())
	ltx = &ld.Transaction{Tx: ld.TxData{
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
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "invalid sigClaims, should be nil")

	input = &ld.TxUpdater{
		ModelID:   &ld.RawModelID,
		Version:   1,
		Threshold: ld.Uint16Ptr(0),
		Keepers:   &signer.Keys{signer.Signer1.Key()},
		Data:      []byte(`42`),
	}
	assert.NoError(input.SyntacticVerify())
	ltx = &ld.Transaction{Tx: ld.TxData{
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
	itx, err := NewTx(ltx)
	assert.NoError(err)

	cs.CommitAccounts()
	assert.ErrorContains(itx.Apply(ctx, cs),
		"insufficient NativeLDC balance, expected 1084600, got 0")
	cs.CheckoutAccounts()

	senderAcc := cs.MustAccount(sender)
	senderAcc.Add(constants.NativeToken, new(big.Int).SetUint64(constants.LDC))
	assert.NoError(itx.Apply(ctx, cs))

	senderGas := ltx.Gas()
	assert.Equal(senderGas*ctx.Price,
		itx.(*TxCreateData).ldc.Balance().Uint64())
	assert.Equal(senderGas*100,
		itx.(*TxCreateData).miner.Balance().Uint64())
	assert.Equal(constants.LDC-senderGas*(ctx.Price+100),
		senderAcc.Balance().Uint64())
	assert.Equal(uint64(1), senderAcc.Nonce())

	di, err := cs.LoadData(itx.(*TxCreateData).di.ID)
	assert.NoError(err)
	assert.Equal(ld.RawModelID, di.ModelID)
	assert.Equal(uint64(1), di.Version)
	assert.Equal(uint16(0), di.Threshold)
	assert.Equal(signer.Keys{signer.Signer1.Key()}, di.Keepers)
	assert.Nil(di.Approver)
	assert.Nil(di.ApproveList)
	assert.Equal([]byte(`42`), []byte(di.Payload))

	jsondata, err := itx.MarshalJSON()
	assert.NoError(err)
	// fmt.Println(string(jsondata))
	assert.Equal(`{"tx":{"type":"TypeCreateData","chainID":2357,"nonce":0,"gasTip":100,"gasFeeCap":1000,"from":"0x8db97c7cECe249C2b98bdc0226cc4C2A57bF52fc","data":{"mid":"AAAAAAAAAAAAAAAAAAAAAAAAAADzaDye","version":1,"threshold":0,"keepers":["jbl8fOziScK5i9wCJsxMKle_UvwKxwPH"],"data":42}},"sigs":["ttZ88NBzJ7rnBGmEXGXHoO21D1SVs23Z10ps7_j8y6tx8cOjOEdZsPJdTCuGxMrqcPS-2JrdPB3MxlqF9FkvWQGgs59e"],"id":"hKP3Ze5zrRLoHxZVnNmzhSZnkYN5VfIRtrHWIJYCbPFWLc4F"}`, string(jsondata))

	assert.NoError(cs.VerifyState())
}

func TestTxCreateCBORData(t *testing.T) {
	assert := assert.New(t)

	ctx := NewMockChainContext()
	cs := ctx.MockChainState()

	sender := signer.Signer1.Key().Address()

	type cborData struct {
		Name   string `cbor:"na"`
		Nonces []int  `cbor:"no"`
	}

	data, err := util.MarshalCBOR(&cborData{Name: "test", Nonces: []int{1, 2, 3}})
	assert.NoError(err)
	invalidData := data[:len(data)-3]

	// CBORModel
	input := &ld.TxUpdater{
		ModelID:   &ld.CBORModelID,
		Version:   1,
		Threshold: ld.Uint16Ptr(0),
		Keepers:   &signer.Keys{signer.Signer1.Key()},
		Data:      invalidData,
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
	itx, err := NewTx(ltx)
	assert.NoError(err)

	cs.CommitAccounts()
	assert.ErrorContains(itx.Apply(ctx, cs),
		"insufficient NativeLDC balance, expected 1149500, got 0")
	cs.CheckoutAccounts()

	senderAcc := cs.MustAccount(sender)
	senderAcc.Add(constants.NativeToken, new(big.Int).SetUint64(constants.LDC))
	cs.CommitAccounts()
	assert.ErrorContains(itx.Apply(ctx, cs), "invalid CBOR encoding data")
	cs.CheckoutAccounts()

	input = &ld.TxUpdater{
		ModelID:   &ld.CBORModelID,
		Version:   1,
		Threshold: ld.Uint16Ptr(0),
		Keepers:   &signer.Keys{signer.Signer1.Key()},
		Data:      data,
	}
	assert.NoError(input.SyntacticVerify())
	ltx = &ld.Transaction{Tx: ld.TxData{
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
	itx, err = NewTx(ltx)
	assert.NoError(err)
	assert.NoError(itx.Apply(ctx, cs))

	senderGas := ltx.Gas()
	assert.Equal(senderGas*ctx.Price,
		itx.(*TxCreateData).ldc.Balance().Uint64())
	assert.Equal(senderGas*100,
		itx.(*TxCreateData).miner.Balance().Uint64())
	assert.Equal(constants.LDC-senderGas*(ctx.Price+100),
		senderAcc.Balance().Uint64())
	assert.Equal(uint64(1), senderAcc.Nonce())

	di, err := cs.LoadData(itx.(*TxCreateData).di.ID)
	assert.NoError(err)
	assert.Equal(ld.CBORModelID, di.ModelID)
	assert.Equal(uint64(1), di.Version)
	assert.Equal(uint16(0), di.Threshold)
	assert.Equal(signer.Keys{signer.Signer1.Key()}, di.Keepers)
	assert.Nil(di.Approver)
	assert.Nil(di.ApproveList)
	assert.Equal(data, []byte(di.Payload))

	jsondata, err := itx.MarshalJSON()
	assert.NoError(err)
	// fmt.Println(string(jsondata))
	assert.Equal(`{"tx":{"type":"TypeCreateData","chainID":2357,"nonce":0,"gasTip":100,"gasFeeCap":1000,"from":"0x8db97c7cECe249C2b98bdc0226cc4C2A57bF52fc","data":{"mid":"AAAAAAAAAAAAAAAAAAAAAAAAAAGIYKah","version":1,"threshold":0,"keepers":["jbl8fOziScK5i9wCJsxMKle_UvwKxwPH"],"data":"omJuYWR0ZXN0Ym5vgwECA_B3TG8"}},"sigs":["CXPn-XOm0zK8b0M5G0j71R6wG_Pw4WBucCsG5YxQf6UZkEZXydY0qW1vyt0wKKq1q3y15sdQLJkR359KOvxWkgGW9Ibr"],"id":"EokfA74wRa86_ZQxxI-ieGgP7trT6QbEkEU0lX-ZmKYI29Gr"}`, string(jsondata))

	assert.NoError(cs.VerifyState())
}

func TestTxCreateJSONData(t *testing.T) {
	assert := assert.New(t)

	ctx := NewMockChainContext()
	cs := ctx.MockChainState()

	sender := signer.Signer1.Key().Address()

	type jsonData struct {
		Name   string `json:"na"`
		Nonces []int  `json:"no"`
	}

	data, err := json.Marshal(&jsonData{Name: "test", Nonces: []int{1, 2, 3}})
	assert.NoError(err)
	invalidData := data[:len(data)-3]

	// JSONModel
	input := &ld.TxUpdater{
		ModelID:   &ld.JSONModelID,
		Version:   1,
		Threshold: ld.Uint16Ptr(0),
		Keepers:   &signer.Keys{signer.Signer1.Key()},
		Data:      invalidData,
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
	itx, err := NewTx(ltx)
	assert.NoError(err)

	cs.CommitAccounts()
	assert.ErrorContains(itx.Apply(ctx, cs),
		"insufficient NativeLDC balance, expected 1208900, got 0")
	cs.CheckoutAccounts()

	senderAcc := cs.MustAccount(sender)
	senderAcc.Add(constants.NativeToken, new(big.Int).SetUint64(constants.LDC))
	cs.CommitAccounts()
	assert.ErrorContains(itx.Apply(ctx, cs), "invalid JSON encoding data")
	cs.CheckoutAccounts()

	input = &ld.TxUpdater{
		ModelID:   &ld.JSONModelID,
		Version:   1,
		Threshold: ld.Uint16Ptr(0),
		Keepers:   &signer.Keys{signer.Signer1.Key()},
		Data:      data,
	}
	assert.NoError(input.SyntacticVerify())
	ltx = &ld.Transaction{Tx: ld.TxData{
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
	itx, err = NewTx(ltx)
	assert.NoError(err)
	assert.NoError(itx.Apply(ctx, cs))

	senderGas := ltx.Gas()
	assert.Equal(senderGas*ctx.Price,
		itx.(*TxCreateData).ldc.Balance().Uint64())
	assert.Equal(senderGas*100,
		itx.(*TxCreateData).miner.Balance().Uint64())
	assert.Equal(constants.LDC-senderGas*(ctx.Price+100),
		senderAcc.Balance().Uint64())
	assert.Equal(uint64(1), senderAcc.Nonce())

	di, err := cs.LoadData(itx.(*TxCreateData).di.ID)
	assert.NoError(err)
	assert.Equal(ld.JSONModelID, di.ModelID)
	assert.Equal(uint64(1), di.Version)
	assert.Equal(uint16(0), di.Threshold)
	assert.Equal(signer.Keys{signer.Signer1.Key()}, di.Keepers)
	assert.Nil(di.Approver)
	assert.Nil(di.ApproveList)
	assert.Equal(data, []byte(di.Payload))

	jsondata, err := itx.MarshalJSON()
	assert.NoError(err)
	// fmt.Println(string(jsondata))
	assert.Equal(`{"tx":{"type":"TypeCreateData","chainID":2357,"nonce":0,"gasTip":100,"gasFeeCap":1000,"from":"0x8db97c7cECe249C2b98bdc0226cc4C2A57bF52fc","data":{"mid":"AAAAAAAAAAAAAAAAAAAAAAAAAALZFhrw","version":1,"threshold":0,"keepers":["jbl8fOziScK5i9wCJsxMKle_UvwKxwPH"],"data":{"na":"test","no":[1,2,3]}}},"sigs":["ilbLtpq89vy-VaFdaM8wvGfIb2uQ4ttfhBDvTvgA-4stBCVit1lrs6od0n-6lZa2IPNU23iiSycOeeEtxsvVaQHc0U3B"],"id":"kh-RbEn8dkNCaFnv2aht4C9YMdqZHxrhSeRIDdV9DiGFCeLh"}`, string(jsondata))

	assert.NoError(cs.VerifyState())
}

func TestTxCreateModelDataWithoutKeepers(t *testing.T) {
	assert := assert.New(t)

	ctx := NewMockChainContext()
	cs := ctx.MockChainState()

	sender := signer.Signer1.Key().Address()

	pm, err := service.ProfileModel()
	assert.NoError(err)
	ps := &ld.ModelInfo{
		Name:      pm.Name(),
		Threshold: 0,
		Keepers:   signer.Keys{signer.Signer2.Key()},
		Schema:    pm.Schema(),
		ID:        util.ModelID{1, 2, 3, 4, 5},
	}

	p := &service.Profile{
		Type:       1,
		Name:       "tester",
		Follows:    util.DataIDs{},
		Extensions: service.Extensions{},
	}
	assert.NoError(p.SyntacticVerify())

	data := p.Bytes()
	input := &ld.TxUpdater{
		ModelID:   &ps.ID,
		Version:   1,
		Threshold: ld.Uint16Ptr(0),
		Keepers:   &signer.Keys{signer.Signer1.Key()},
		Data:      data[1:],
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
	itx, err := NewTx(ltx)
	assert.NoError(err)

	cs.CommitAccounts()
	assert.ErrorContains(itx.Apply(ctx, cs),
		"insufficient NativeLDC balance, expected 1250700, got 0")
	cs.CheckoutAccounts()

	senderAcc := cs.MustAccount(sender)
	senderAcc.Add(constants.NativeToken, new(big.Int).SetUint64(constants.LDC))
	cs.CommitAccounts()
	assert.ErrorContains(itx.Apply(ctx, cs), "AQIDBAUAAAAAAAAAAAAAAAAAAADELdAH not found")
	cs.CheckoutAccounts()

	assert.NoError(cs.SaveModel(ps))
	cs.CommitAccounts()
	assert.ErrorContains(itx.Apply(ctx, cs),
		`TxCreateData.Apply: ld.IPLDModel("ProfileService").Valid: cbor`)
	cs.CheckoutAccounts()

	input = &ld.TxUpdater{
		ModelID:   &ps.ID,
		Version:   1,
		Threshold: ld.Uint16Ptr(1),
		Keepers:   &signer.Keys{signer.Signer1.Key()},
		Data:      data,
	}
	assert.NoError(input.SyntacticVerify())
	ltx = &ld.Transaction{Tx: ld.TxData{
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

	itx, err = NewTx(ltx)
	assert.NoError(err)
	assert.NoError(itx.Apply(ctx, cs))

	senderGas := ltx.Gas()
	assert.Equal(senderGas*ctx.Price,
		itx.(*TxCreateData).ldc.Balance().Uint64())
	assert.Equal(senderGas*100,
		itx.(*TxCreateData).miner.Balance().Uint64())
	assert.Equal(constants.LDC-senderGas*(ctx.Price+100),
		senderAcc.Balance().Uint64())
	assert.Equal(uint64(1), senderAcc.Nonce())

	di, err := cs.LoadData(itx.(*TxCreateData).di.ID)
	assert.NoError(err)
	assert.Equal(ps.ID, di.ModelID)
	assert.Equal(uint64(1), di.Version)
	assert.Equal(uint16(1), di.Threshold)
	assert.Equal(signer.Keys{signer.Signer1.Key()}, di.Keepers)
	assert.Nil(di.Approver)
	assert.Nil(di.ApproveList)
	assert.Equal(data, []byte(di.Payload))

	p2 := &service.Profile{}
	assert.NoError(p2.Unmarshal(di.Payload))
	assert.NoError(p2.SyntacticVerify())
	assert.Equal(p.Type, p2.Type)
	assert.Equal(p.Name, p2.Name)
	assert.Equal(p.Follows, p2.Follows)
	assert.Equal(p.Extensions, p2.Extensions)
	assert.Equal(p.Bytes(), p2.Bytes())

	jsondata, err := itx.MarshalJSON()
	assert.NoError(err)
	// fmt.Println(string(jsondata))
	assert.Equal(`{"tx":{"type":"TypeCreateData","chainID":2357,"nonce":0,"gasTip":100,"gasFeeCap":1000,"from":"0x8db97c7cECe249C2b98bdc0226cc4C2A57bF52fc","data":{"mid":"AQIDBAUAAAAAAAAAAAAAAAAAAADELdAH","version":1,"threshold":1,"keepers":["jbl8fOziScK5i9wCJsxMKle_UvwKxwPH"],"data":"p2FkYGFpYGFuZnRlc3RlcmF0AWF1YGJlc4BiZnOA_mpfEg"}},"sigs":["tFjAaEX2gBzH26OkCPMX9uyzCTALrmkpT1eV24YIKMFovX41e6Zshx7D8o92u9c_UZIQNDL_Ope2G7lvsygMGABoALzU"],"id":"wqVmFPM_CfUdWb1brcvmarDK5BA0-Ed_altqVsi9yX8K_pyi"}`, string(jsondata))

	assert.NoError(cs.VerifyState())
}

func TestTxCreateModelDataWithKeepers(t *testing.T) {
	assert := assert.New(t)

	ctx := NewMockChainContext()
	cs := ctx.MockChainState()

	sender := signer.Signer1.Key().Address()
	recipient := signer.Signer2.Key().Address()

	pm, err := service.ProfileModel()
	assert.NoError(err)
	mi := &ld.ModelInfo{
		Name:      pm.Name(),
		Threshold: 1,
		Keepers:   signer.Keys{signer.Signer2.Key()},
		Schema:    pm.Schema(),
		ID:        util.ModelID{1, 2, 3},
	}

	pf := &service.Profile{
		Type:       1,
		Name:       "LDC",
		Follows:    util.DataIDs{},
		Extensions: service.Extensions{},
	}
	assert.NoError(pf.SyntacticVerify())
	data := pf.Bytes()
	input := &ld.TxUpdater{
		ModelID:   &mi.ID,
		Version:   1,
		Threshold: ld.Uint16Ptr(0),
		Keepers:   &signer.Keys{signer.Signer1.Key()},
		Data:      data,
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
	itx, err := NewTx(ltx)
	assert.NoError(err)

	senderAcc := cs.MustAccount(sender)
	senderAcc.Add(constants.NativeToken, new(big.Int).SetUint64(constants.LDC))
	cs.CommitAccounts()
	assert.ErrorContains(itx.Apply(ctx, cs), "AQIDAAAAAAAAAAAAAAAAAAAAAABuT_CC not found")
	cs.CheckoutAccounts()
	assert.NoError(cs.SaveModel(mi))
	cs.CommitAccounts()
	assert.ErrorContains(itx.Apply(ctx, cs), `TxCreateData.Apply: nil to`)
	cs.CheckoutAccounts()

	input = &ld.TxUpdater{
		ModelID:   &mi.ID,
		Version:   1,
		Threshold: ld.Uint16Ptr(0),
		Keepers:   &signer.Keys{signer.Signer1.Key()},
		Data:      data,
		To:        &recipient,
	}
	assert.NoError(input.SyntacticVerify())
	ltx = &ld.Transaction{Tx: ld.TxData{
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
	_, err = NewTx(ltx)
	assert.ErrorContains(err,
		"invalid to, expected 0x44171C37Ff5D7B7bb8Dcad5C81f16284A229E641, got <nil>")

	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeCreateData,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      sender,
		To:        &constants.GenesisAccount,
		Data:      input.Bytes(),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err,
		"invalid to, expected 0x44171C37Ff5D7B7bb8Dcad5C81f16284A229E641, got 0xFFfFFFfFfffFFfFFffFFFfFfFffFFFfffFfFFFff")

	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeCreateData,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      sender,
		To:        &recipient,
		Data:      input.Bytes(),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	ltx.Timestamp = 10
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "data expired")

	input = &ld.TxUpdater{
		ModelID:   &mi.ID,
		Version:   1,
		Threshold: ld.Uint16Ptr(0),
		Keepers:   &signer.Keys{signer.Signer1.Key()},
		Data:      data,
		To:        &recipient,
		Expire:    100,
	}
	assert.NoError(input.SyntacticVerify())
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeCreateData,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      sender,
		To:        &recipient,
		Data:      input.Bytes(),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	ltx.Timestamp = 10
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "nil amount")

	input = &ld.TxUpdater{
		ModelID:   &mi.ID,
		Version:   1,
		Threshold: ld.Uint16Ptr(0),
		Keepers:   &signer.Keys{signer.Signer1.Key()},
		Data:      data,
		To:        &recipient,
		Expire:    100,
		Amount:    new(big.Int).SetUint64(0),
	}
	assert.NoError(input.SyntacticVerify())
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeCreateData,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      sender,
		To:        &recipient,
		Amount:    new(big.Int).SetUint64(1),
		Data:      input.Bytes(),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	ltx.Timestamp = 10
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "invalid amount, expected 0, got 1")

	assert.NoError(input.SyntacticVerify())
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeCreateData,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      sender,
		To:        &recipient,
		Amount:    new(big.Int).SetUint64(0),
		Data:      input.Bytes(),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	ltx.Timestamp = 10
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "no exSignatures")

	input = &ld.TxUpdater{
		ModelID:   &mi.ID,
		Version:   1,
		Threshold: ld.Uint16Ptr(1),
		Keepers:   &signer.Keys{signer.Signer1.Key()},
		Data:      data,
		To:        &recipient,
		Expire:    100,
		Amount:    new(big.Int).SetUint64(0),
	}
	assert.NoError(input.SyntacticVerify())
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeCreateData,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      sender,
		To:        &recipient,
		Amount:    new(big.Int).SetUint64(0),
		Data:      input.Bytes(),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.ExSignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	ltx.Timestamp = 10
	itx, err = NewTx(ltx)
	assert.NoError(err)
	cs.CommitAccounts()
	assert.ErrorContains(itx.Apply(ctx, cs),
		"invalid exSignatures for model keepers")
	cs.CheckoutAccounts()

	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeCreateData,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      sender,
		To:        &recipient,
		Amount:    new(big.Int).SetUint64(0),
		Data:      input.Bytes(),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.ExSignWith(signer.Signer2))
	assert.NoError(ltx.SyntacticVerify())
	ltx.Timestamp = 10
	itx, err = NewTx(ltx)
	assert.NoError(err)
	assert.NoError(itx.Apply(ctx, cs))

	senderGas := ltx.Gas()
	assert.Equal(senderGas*ctx.Price,
		itx.(*TxCreateData).ldc.Balance().Uint64())
	assert.Equal(senderGas*100,
		itx.(*TxCreateData).miner.Balance().Uint64())
	assert.Equal(constants.LDC-senderGas*(ctx.Price+100),
		senderAcc.Balance().Uint64())
	assert.Equal(uint64(1), senderAcc.Nonce())

	di, err := cs.LoadData(itx.(*TxCreateData).di.ID)
	assert.NoError(err)
	assert.Equal(mi.ID, di.ModelID)
	assert.Equal(uint64(1), di.Version)
	assert.Equal(uint16(1), di.Threshold)
	assert.Equal(signer.Keys{signer.Signer1.Key()}, di.Keepers)
	assert.Nil(di.Approver)
	assert.Nil(di.ApproveList)
	assert.Equal(data, []byte(di.Payload))

	p2 := &service.Profile{}
	assert.NoError(p2.Unmarshal(di.Payload))
	assert.NoError(p2.SyntacticVerify())
	assert.Equal(pf.Type, p2.Type)
	assert.Equal(pf.Name, p2.Name)
	assert.Equal(pf.Follows, p2.Follows)
	assert.Equal(pf.Extensions, p2.Extensions)
	assert.Equal(pf.Bytes(), p2.Bytes())

	jsondata, err := itx.MarshalJSON()
	assert.NoError(err)
	// fmt.Println(string(jsondata))
	assert.Equal(`{"tx":{"type":"TypeCreateData","chainID":2357,"nonce":0,"gasTip":100,"gasFeeCap":1000,"from":"0x8db97c7cECe249C2b98bdc0226cc4C2A57bF52fc","to":"0x44171C37Ff5D7B7bb8Dcad5C81f16284A229E641","amount":0,"data":{"mid":"AQIDAAAAAAAAAAAAAAAAAAAAAABuT_CC","version":1,"threshold":1,"keepers":["jbl8fOziScK5i9wCJsxMKle_UvwKxwPH"],"to":"0x44171C37Ff5D7B7bb8Dcad5C81f16284A229E641","amount":0,"expire":100,"data":"p2FkYGFpYGFuY0xEQ2F0AWF1YGJlc4BiZnOABj3NMQ"}},"sigs":["O-g2YPiQMATEkQhDrXFuRyqrWSCw3HIIOyGAm2XAV2c4d-oqms9frFzEif0jtAVU1JVR_UEAucdQnONtTYA7tAEvC_nZ"],"exSigs":["Xs2qEFdgNxVYqCuoCQrYTTehsfSvsSNLJStECHBMgU9sEicGK7zIuWfoY9z-ZP81fyiOkuD5Xk8nXIFRP4PszgEUU4Um"],"id":"rcaBsIO7PtL7AZJQiAQBjUZb7J59aXPJ4Fpe-ys3QnU1frQw"}`, string(jsondata))

	assert.NoError(cs.VerifyState())
}

func TestTxCreateNameModelData(t *testing.T) {
	assert := assert.New(t)

	ctx := NewMockChainContext()
	cs := ctx.MockChainState()

	sender := signer.Signer1.Key().Address()
	recipient := signer.Signer2.Key().Address()

	nm, err := service.NameModel()
	assert.NoError(err)
	mi := &ld.ModelInfo{
		Name:      nm.Name(),
		Threshold: 1,
		Keepers:   signer.Keys{signer.Signer2.Key()},
		Schema:    nm.Schema(),
		ID:        ctx.ChainConfig().NameServiceID,
	}

	name := &service.Name{
		Name:    "ldc.to.",
		Records: []string{"ldc.to. IN A 10.0.0.1"},
		Extensions: service.Extensions{{
			Title: "Test",
			Properties: map[string]interface{}{
				"desc": "desc",
			},
		}},
	}
	assert.NoError(name.SyntacticVerify())
	data := name.Bytes()

	input := &ld.TxUpdater{
		ModelID:   &mi.ID,
		Version:   1,
		Threshold: ld.Uint16Ptr(1),
		Keepers:   &signer.Keys{signer.Signer1.Key()},
		Data:      data,
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
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.ExSignWith(signer.Signer2))
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

	senderGas := ltx.Gas()
	assert.Equal(senderGas*ctx.Price,
		itx.(*TxCreateData).ldc.Balance().Uint64())
	assert.Equal(senderGas*100,
		itx.(*TxCreateData).miner.Balance().Uint64())
	assert.Equal(constants.MilliLDC, itx.(*TxCreateData).to.Balance().Uint64())
	assert.Equal(constants.LDC-senderGas*(ctx.Price+100)-constants.MilliLDC,
		senderAcc.Balance().Uint64())
	assert.Equal(uint64(1), senderAcc.Nonce())

	di, err := cs.LoadDataByName("ldc.to.")
	assert.NoError(err)
	assert.Equal(mi.ID, di.ModelID)
	assert.Equal(uint64(1), di.Version)
	assert.Equal(uint16(1), di.Threshold)
	assert.Equal(signer.Keys{signer.Signer1.Key()}, di.Keepers)
	assert.Nil(di.Approver)
	assert.Nil(di.ApproveList)
	assert.Equal(data, []byte(di.Payload))

	n2 := &service.Name{}
	assert.NoError(n2.Unmarshal(di.Payload))
	assert.NoError(n2.SyntacticVerify())
	assert.Equal(n2.Name, name.Name)
	assert.Equal(n2.Records, name.Records)
	assert.Equal(n2.Bytes(), name.Bytes())
	assert.Equal(1, len(n2.Extensions))
	assert.Equal("Test", n2.Extensions[0].Title)
	assert.Equal(1, len(n2.Extensions[0].Properties))
	assert.Equal("desc", n2.Extensions[0].Properties["desc"].(string))

	jsondata, err := itx.MarshalJSON()
	assert.NoError(err)
	// fmt.Println(string(jsondata))
	assert.Equal(`{"tx":{"type":"TypeCreateData","chainID":2357,"nonce":0,"gasTip":100,"gasFeeCap":1000,"from":"0x8db97c7cECe249C2b98bdc0226cc4C2A57bF52fc","to":"0x44171C37Ff5D7B7bb8Dcad5C81f16284A229E641","amount":1000000,"data":{"mid":"b8onI5zOwqPZO9jxMBBgZWnnCUzd-187","version":1,"threshold":1,"keepers":["jbl8fOziScK5i9wCJsxMKle_UvwKxwPH"],"to":"0x44171C37Ff5D7B7bb8Dcad5C81f16284A229E641","amount":1000000,"expire":100,"data":"o2FuZ2xkYy50by5iZXOBomF0ZFRlc3RicHOhZGRlc2NkZGVzY2Jyc4F1bGRjLnRvLiBJTiBBIDEwLjAuMC4xJFhefQ"}},"sigs":["Agi0n3ypfN8pX-EuIbhKfDmH-NSksy8IyPwtKiy4xqZsCD6F06sOe_H-SGgd7xqWoy-ZkQl0LZnMA6SBDnKLmgBgH3sg"],"exSigs":["NkF-wtCdSzXees4-Xw_xxmzNRl9hinEoFGHZ0KZCW5MLfAf-YdE6RCTdUflBg2ss5ncv_Sba3zD818ihV1Tj8QH31jau"],"id":"h8Ac6wx8oCUWzySj6Ze0RqsAFYSM3FFefWdbIwYn50EhNMbm"}`, string(jsondata))

	assert.NoError(cs.VerifyState())

	name2 := &service.Name{
		Name:       "ldc.to.",
		Records:    []string{"ldc.to. IN A 10.0.0.2"},
		Extensions: service.Extensions{},
	}
	assert.NoError(name2.SyntacticVerify())
	data = name2.Bytes()

	input = &ld.TxUpdater{
		ModelID:   &mi.ID,
		Version:   1,
		Threshold: ld.Uint16Ptr(1),
		Keepers:   &signer.Keys{signer.Signer1.Key()},
		Data:      data,
		To:        &recipient,
		Expire:    100,
		Amount:    new(big.Int).SetUint64(constants.MilliLDC),
	}
	assert.NoError(input.SyntacticVerify())
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeCreateData,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     1,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      sender,
		To:        &recipient,
		Amount:    new(big.Int).SetUint64(constants.MilliLDC),
		Data:      input.Bytes(),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.ExSignWith(signer.Signer2))
	assert.NoError(ltx.SyntacticVerify())
	ltx.Timestamp = 10
	itx, err = NewTx(ltx)
	assert.NoError(err)
	cs.CommitAccounts()
	assert.ErrorContains(itx.Apply(ctx, cs),
		`TxCreateData.Apply: name "ldc.to." is conflict`)
	cs.CheckoutAccounts()

	name2 = &service.Name{
		Name:       "api.ldc.to.",
		Records:    []string{},
		Extensions: service.Extensions{},
	}
	assert.NoError(name2.SyntacticVerify())
	data = name2.Bytes()

	input = &ld.TxUpdater{
		ModelID:   &mi.ID,
		Version:   1,
		Threshold: ld.Uint16Ptr(1),
		Keepers:   &signer.Keys{signer.Signer1.Key()},
		Data:      data,
		To:        &recipient,
		Expire:    100,
		Amount:    new(big.Int).SetUint64(constants.MilliLDC),
	}

	assert.NoError(input.SyntacticVerify())
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeCreateData,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     1,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      sender,
		To:        &recipient,
		Amount:    new(big.Int).SetUint64(constants.MilliLDC),
		Data:      input.Bytes(),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.ExSignWith(signer.Signer2))
	assert.NoError(ltx.SyntacticVerify())
	ltx.Timestamp = 10
	itx, err = NewTx(ltx)
	assert.NoError(err)
	assert.NoError(itx.Apply(ctx, cs))
}

func TestTxCreateDataGenesis(t *testing.T) {
	assert := assert.New(t)

	ctx := NewMockChainContext()
	cs := ctx.MockChainState()
	sender := signer.Signer1.Key().Address()

	cfg, err := json.Marshal(ctx.ChainConfig().FeeConfig)
	assert.NoError(err)

	cfgData := &ld.TxUpdater{
		ModelID:   &ld.JSONModelID,
		Version:   1,
		Threshold: ld.Uint16Ptr(1),
		Keepers:   &signer.Keys{signer.Signer1.Key()},
		Data:      cfg,
	}
	assert.NoError(cfgData.SyntacticVerify())
	ltx := &ld.Transaction{Tx: ld.TxData{
		Type:    ld.TypeCreateData,
		ChainID: ctx.ChainConfig().ChainID,
		From:    sender,
		Data:    ld.MustMarshal(cfgData),
	}}
	assert.NoError(ltx.SyntacticVerify())
	itx, err := NewGenesisTx(ltx)
	assert.NoError(err)
	assert.NoError(itx.(GenesisTx).ApplyGenesis(ctx, cs))

	assert.Equal(uint64(0), itx.(*TxCreateData).ldc.Balance().Uint64())
	assert.Equal(uint64(0), itx.(*TxCreateData).miner.Balance().Uint64())
	assert.Equal(uint64(0), itx.(*TxCreateData).from.Balance().Uint64())
	assert.Equal(uint64(1), itx.(*TxCreateData).from.Nonce())

	di, err := cs.LoadData(itx.(*TxCreateData).di.ID)
	assert.NoError(err)
	assert.Equal(ld.JSONModelID, di.ModelID)
	assert.Equal(uint64(1), di.Version)
	assert.Equal(uint16(1), di.Threshold)
	assert.Equal(signer.Keys{signer.Signer1.Key()}, di.Keepers)
	assert.Nil(di.Approver)
	assert.Nil(di.ApproveList)
	assert.True(jsonpatch.Equal(cfg, di.Payload))

	jsondata, err := itx.MarshalJSON()
	assert.NoError(err)
	// fmt.Println(string(jsondata))
	assert.Equal(`{"tx":{"type":"TypeCreateData","chainID":2357,"nonce":0,"gasTip":0,"gasFeeCap":0,"from":"0x8db97c7cECe249C2b98bdc0226cc4C2A57bF52fc","data":{"mid":"AAAAAAAAAAAAAAAAAAAAAAAAAALZFhrw","version":1,"threshold":1,"keepers":["jbl8fOziScK5i9wCJsxMKle_UvwKxwPH"],"data":{"startHeight":0,"thresholdGas":1000,"minGasPrice":10000,"maxGasPrice":100000,"maxTxGas":42000000,"maxBlockTxsSize":4200000,"gasRebateRate":1000,"minTokenPledge":10000000000000,"minStakePledge":1000000000000}}},"id":"lqFJUOJetbRIx7A6LGXP5ON2u0mhzpCoH4nXNEzBEodzLpRM"}`, string(jsondata))

	assert.NoError(cs.VerifyState())
}
