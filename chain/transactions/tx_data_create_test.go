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
	sender := util.Signer1.Address()

	txData := &ld.TxData{
		Type:      ld.TypeCreateData,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      sender,
	}
	assert.NoError(txData.SyntacticVerify())
	_, err = NewTx2(txData.ToTransaction())
	assert.ErrorContains(err, "DeriveSigners error: no signature")

	txData = &ld.TxData{
		Type:      ld.TypeCreateData,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      sender,
		Token:     &token,
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx2(txData.ToTransaction())
	assert.ErrorContains(err, "invalid token, should be nil")

	txData = &ld.TxData{
		Type:      ld.TypeCreateData,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      sender,
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx2(txData.ToTransaction())
	assert.ErrorContains(err, "invalid data")

	txData = &ld.TxData{
		Type:      ld.TypeCreateData,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      sender,
		Data:      []byte("你好👋"),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx2(txData.ToTransaction())
	assert.ErrorContains(err, "cbor: unexpected following extraneous data")

	input := &ld.TxUpdater{}
	txData = &ld.TxData{
		Type:      ld.TypeCreateData,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      sender,
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx2(txData.ToTransaction())
	assert.ErrorContains(err, "nil mid")

	input = &ld.TxUpdater{
		ModelID: &ld.RawModelID,
	}
	txData = &ld.TxData{
		Type:      ld.TypeCreateData,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      sender,
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx2(txData.ToTransaction())
	assert.ErrorContains(err, "invalid version, expected 1, got 0")

	input = &ld.TxUpdater{
		ModelID: &ld.RawModelID,
		Version: 1,
	}
	txData = &ld.TxData{
		Type:      ld.TypeCreateData,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      sender,
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx2(txData.ToTransaction())
	assert.ErrorContains(err, "nil threshold")

	input = &ld.TxUpdater{
		ModelID:   &ld.RawModelID,
		Version:   1,
		Threshold: ld.Uint16Ptr(0),
	}
	txData = &ld.TxData{
		Type:      ld.TypeCreateData,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      sender,
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx2(txData.ToTransaction())
	assert.ErrorContains(err, "nil keepers")

	input = &ld.TxUpdater{
		ModelID:   &ld.RawModelID,
		Version:   1,
		Threshold: ld.Uint16Ptr(0),
		Keepers:   &util.EthIDs{},
	}
	txData = &ld.TxData{
		Type:      ld.TypeCreateData,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      sender,
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx2(txData.ToTransaction())
	assert.ErrorContains(err, "empty keepers")

	input = &ld.TxUpdater{
		ModelID:   &ld.RawModelID,
		Version:   1,
		Threshold: ld.Uint16Ptr(0),
		Keepers:   &util.EthIDs{util.Signer1.Address()},
	}
	txData = &ld.TxData{
		Type:      ld.TypeCreateData,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      sender,
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx2(txData.ToTransaction())
	assert.ErrorContains(err, "empty data")

	input = &ld.TxUpdater{
		ModelID:   &ld.RawModelID,
		Version:   1,
		Threshold: ld.Uint16Ptr(0),
		Keepers:   &util.EthIDs{util.Signer1.Address()},
		Data:      []byte(`{}`),
	}

	assert.NoError(input.SyntacticVerify())
	txData = &ld.TxData{
		Type:      ld.TypeCreateData,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      sender,
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx2(txData.ToTransaction())
	assert.NoError(err)

	// RawModel
	input = &ld.TxUpdater{
		ModelID:   &ld.RawModelID,
		Version:   1,
		Threshold: ld.Uint16Ptr(0),
		Keepers:   &util.EthIDs{util.Signer1.Address()},
		Data:      []byte(`42`),
	}
	assert.NoError(input.SyntacticVerify())
	txData = &ld.TxData{
		Type:      ld.TypeCreateData,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      sender,
		To:        &constants.GenesisAccount,
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx2(txData.ToTransaction())
	assert.ErrorContains(err, "invalid to, should be nil")

	input = &ld.TxUpdater{
		ModelID:   &ld.RawModelID,
		Version:   1,
		Threshold: ld.Uint16Ptr(0),
		Keepers:   &util.EthIDs{util.Signer1.Address()},
		Data:      []byte(`42`),
	}
	assert.NoError(input.SyntacticVerify())
	txData = &ld.TxData{
		Type:      ld.TypeCreateData,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      sender,
		Amount:    big.NewInt(1),
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx2(txData.ToTransaction())
	assert.ErrorContains(err, "nil to together with amount")

	input = &ld.TxUpdater{
		ModelID:   &ld.RawModelID,
		Version:   1,
		Threshold: ld.Uint16Ptr(0),
		Keepers:   &util.EthIDs{util.Signer1.Address()},
		Data:      []byte(`42`),
	}
	assert.NoError(input.SyntacticVerify())
	txData = &ld.TxData{
		Type:      ld.TypeCreateData,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      sender,
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	assert.NoError(txData.ExSignWith(util.Signer1))
	_, err = NewTx2(txData.ToTransaction())
	assert.ErrorContains(err, "invalid exSignatures, should be nil")

	input = &ld.TxUpdater{
		ModelID:   &ld.RawModelID,
		Version:   1,
		Threshold: ld.Uint16Ptr(0),
		Keepers:   &util.EthIDs{util.Signer1.Address()},
		Data:      []byte(`42`),
		Sig:       &util.Signature{1, 2, 3},
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
	txData = &ld.TxData{
		Type:      ld.TypeCreateData,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      sender,
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx2(txData.ToTransaction())
	assert.ErrorContains(err, "invalid sigClaims, should be nil")

	input = &ld.TxUpdater{
		ModelID:   &ld.RawModelID,
		Version:   1,
		Threshold: ld.Uint16Ptr(0),
		Keepers:   &util.EthIDs{util.Signer1.Address()},
		Data:      []byte(`42`),
	}
	assert.NoError(input.SyntacticVerify())
	txData = &ld.TxData{
		Type:      ld.TypeCreateData,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      sender,
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	tt := txData.ToTransaction()
	itx, err := NewTx2(tt)
	assert.NoError(err)

	cs.CommitAccounts()
	assert.ErrorContains(itx.Apply(ctx, cs),
		"insufficient NativeLDC balance, expected 1061500, got 0")
	cs.CheckoutAccounts()

	senderAcc := cs.MustAccount(sender)
	senderAcc.Add(constants.NativeToken, new(big.Int).SetUint64(constants.LDC))
	assert.NoError(itx.Apply(ctx, cs))

	senderGas := tt.Gas()
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
	assert.Equal(util.EthIDs{sender}, di.Keepers)
	assert.Nil(di.Approver)
	assert.Nil(di.ApproveList)
	assert.Equal([]byte(`42`), []byte(di.Data))

	jsondata, err := itx.MarshalJSON()
	assert.NoError(err)
	// fmt.Println(string(jsondata))
	assert.Equal(`{"type":"TypeCreateData","chainID":2357,"nonce":0,"gasTip":100,"gasFeeCap":1000,"from":"0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC","data":{"mid":"111111111111111111116DBWJs","version":1,"threshold":0,"keepers":["0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC"],"data":42},"signatures":["b6d67cf0d07327bae70469845c65c7a0edb50f5495b36dd9d74a6ceff8fccbab71f1c3a3384759b0f25d4c2b86c4caea70f4bed89add3c1dccc65a85f4592f5901"],"id":"2PpJt2u36UAPEHnGfimSqKdp7cUoJBdJr5BY8GQhjuGs32zVp"}`, string(jsondata))

	assert.NoError(cs.VerifyState())
}

func TestTxCreateCBORData(t *testing.T) {
	assert := assert.New(t)

	ctx := NewMockChainContext()
	cs := ctx.MockChainState()

	sender := util.Signer1.Address()

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
		Keepers:   &util.EthIDs{util.Signer1.Address()},
		Data:      invalidData,
	}
	assert.NoError(input.SyntacticVerify())
	txData := &ld.TxData{
		Type:      ld.TypeCreateData,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      sender,
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	tt := txData.ToTransaction()
	itx, err := NewTx2(tt)
	assert.NoError(err)

	cs.CommitAccounts()
	assert.ErrorContains(itx.Apply(ctx, cs),
		"insufficient NativeLDC balance, expected 1125300, got 0")
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
		Keepers:   &util.EthIDs{util.Signer1.Address()},
		Data:      data,
	}
	assert.NoError(input.SyntacticVerify())
	txData = &ld.TxData{
		Type:      ld.TypeCreateData,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      sender,
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	tt = txData.ToTransaction()
	itx, err = NewTx2(tt)
	assert.NoError(err)
	assert.NoError(itx.Apply(ctx, cs))

	senderGas := tt.Gas()
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
	assert.Equal(util.EthIDs{sender}, di.Keepers)
	assert.Nil(di.Approver)
	assert.Nil(di.ApproveList)
	assert.Equal(data, []byte(di.Data))

	jsondata, err := itx.MarshalJSON()
	assert.NoError(err)
	// fmt.Println(string(jsondata))
	assert.Equal(`{"type":"TypeCreateData","chainID":2357,"nonce":0,"gasTip":100,"gasFeeCap":1000,"from":"0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC","data":{"mid":"1111111111111111111Ax1asG","version":1,"threshold":0,"keepers":["0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC"],"data":"0xa2626e616474657374626e6f830102031e0946b2"},"signatures":["0973e7f973a6d332bc6f43391b48fbd51eb01bf3f0e1606e702b06e58c507fa519904657c9d634a96d6fcadd3028aab5ab7cb5e6c7502c9911df9f4a3afc569201"],"id":"2FpeGMLZMgzjHD1WtmUkKY53SsZ5ZASPQkrdqcmqkLHXk2Dj8R"}`, string(jsondata))

	assert.NoError(cs.VerifyState())
}

func TestTxCreateJSONData(t *testing.T) {
	assert := assert.New(t)

	ctx := NewMockChainContext()
	cs := ctx.MockChainState()

	sender := util.Signer1.Address()

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
		Keepers:   &util.EthIDs{util.Signer1.Address()},
		Data:      invalidData,
	}
	assert.NoError(input.SyntacticVerify())
	txData := &ld.TxData{
		Type:      ld.TypeCreateData,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      sender,
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	tt := txData.ToTransaction()
	itx, err := NewTx2(tt)
	assert.NoError(err)

	cs.CommitAccounts()
	assert.ErrorContains(itx.Apply(ctx, cs),
		"insufficient NativeLDC balance, expected 1184700, got 0")
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
		Keepers:   &util.EthIDs{util.Signer1.Address()},
		Data:      data,
	}
	assert.NoError(input.SyntacticVerify())
	txData = &ld.TxData{
		Type:      ld.TypeCreateData,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      sender,
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	tt = txData.ToTransaction()
	itx, err = NewTx2(tt)
	assert.NoError(err)
	assert.NoError(itx.Apply(ctx, cs))

	senderGas := tt.Gas()
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
	assert.Equal(util.EthIDs{sender}, di.Keepers)
	assert.Nil(di.Approver)
	assert.Nil(di.ApproveList)
	assert.Equal(data, []byte(di.Data))

	jsondata, err := itx.MarshalJSON()
	assert.NoError(err)
	// fmt.Println(string(jsondata))
	assert.Equal(`{"type":"TypeCreateData","chainID":2357,"nonce":0,"gasTip":100,"gasFeeCap":1000,"from":"0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC","data":{"mid":"1111111111111111111L17Xp3","version":1,"threshold":0,"keepers":["0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC"],"data":{"na":"test","no":[1,2,3]}},"signatures":["8a56cbb69abcf6fcbe55a15d68cf30bc67c86f6b90e2db5f8410ef4ef800fb8b2d042562b7596bb3aa1dd27fba9596b620f354db78a24b270e79e12dc6cbd56901"],"id":"bcscpL2jzUQZSeVJVRuhzHB6US81zpUX7U8SDMxDMeCyxWRGC"}`, string(jsondata))

	assert.NoError(cs.VerifyState())
}

func TestTxCreateModelDataWithoutKeepers(t *testing.T) {
	assert := assert.New(t)

	ctx := NewMockChainContext()
	cs := ctx.MockChainState()

	sender := util.Signer1.Address()

	pm, err := service.ProfileModel()
	assert.NoError(err)
	ps := &ld.ModelInfo{
		Name:      pm.Name(),
		Threshold: 0,
		Keepers:   util.EthIDs{util.Signer2.Address()},
		Schema:    pm.Schema(),
		ID:        util.ModelID{1, 2, 3, 4, 5},
	}

	p := &service.Profile{
		Type:       1,
		Name:       "tester",
		Follows:    util.DataIDs{},
		Extensions: []*service.Extension{},
	}
	assert.NoError(p.SyntacticVerify())

	data := p.Bytes()
	input := &ld.TxUpdater{
		ModelID:   &ps.ID,
		Version:   1,
		Threshold: ld.Uint16Ptr(0),
		Keepers:   &util.EthIDs{util.Signer1.Address()},
		Data:      data[1:],
	}
	assert.NoError(input.SyntacticVerify())
	txData := &ld.TxData{
		Type:      ld.TypeCreateData,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      sender,
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	tt := txData.ToTransaction()
	itx, err := NewTx2(tt)
	assert.NoError(err)

	cs.CommitAccounts()
	assert.ErrorContains(itx.Apply(ctx, cs),
		"insufficient NativeLDC balance, expected 1226500, got 0")
	cs.CheckoutAccounts()

	senderAcc := cs.MustAccount(sender)
	senderAcc.Add(constants.NativeToken, new(big.Int).SetUint64(constants.LDC))
	cs.CommitAccounts()
	assert.ErrorContains(itx.Apply(ctx, cs), "6L5yRNNMubYqZoZRtmk1ykJMmZppNwb1 not found")
	cs.CheckoutAccounts()

	assert.NoError(cs.SaveModel(ps))
	cs.CommitAccounts()
	assert.ErrorContains(itx.Apply(ctx, cs),
		`TxCreateData.Apply error: IPLDModel("ProfileService").Valid error: cbor`)
	cs.CheckoutAccounts()

	input = &ld.TxUpdater{
		ModelID:   &ps.ID,
		Version:   1,
		Threshold: ld.Uint16Ptr(1),
		Keepers:   &util.EthIDs{util.Signer1.Address()},
		Data:      data,
	}
	assert.NoError(input.SyntacticVerify())
	txData = &ld.TxData{
		Type:      ld.TypeCreateData,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      sender,
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	tt = txData.ToTransaction()
	_, err = NewTx2(tt)
	assert.NoError(err)
	itx, err = NewTx2(tt)
	assert.NoError(err)
	assert.NoError(itx.Apply(ctx, cs))

	senderGas := tt.Gas()
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
	assert.Equal(util.EthIDs{sender}, di.Keepers)
	assert.Nil(di.Approver)
	assert.Nil(di.ApproveList)
	assert.Equal(data, []byte(di.Data))

	p2 := &service.Profile{}
	assert.NoError(p2.Unmarshal(di.Data))
	assert.NoError(p2.SyntacticVerify())
	assert.Equal(p.Type, p2.Type)
	assert.Equal(p.Name, p2.Name)
	assert.Equal(p.Follows, p2.Follows)
	assert.Equal(p.Extensions, p2.Extensions)
	assert.Equal(p.Bytes(), p2.Bytes())

	jsondata, err := itx.MarshalJSON()
	assert.NoError(err)
	// fmt.Println(string(jsondata))
	assert.Equal(`{"type":"TypeCreateData","chainID":2357,"nonce":0,"gasTip":100,"gasFeeCap":1000,"from":"0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC","data":{"mid":"6L5yRNNMubYqZoZRtmk1ykJMmZppNwb1","version":1,"threshold":1,"keepers":["0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC"],"data":"0xa7616460616960616e6674657374657261740161756062657880626673807e921573"},"signatures":["9edd071f71feaaa25454991debd6d7b66713ad0fad75a909fe81549db99419687eac7f9a1dd75d657bcc718e9beeabe057aa919d30ce24e0ab165a740e21c65600"],"id":"2aVVPBL5DHKBhqy2p4DTHh7pDEfm7o7MMsm4ZCJFREuSPHXk1T"}`, string(jsondata))

	assert.NoError(cs.VerifyState())
}

func TestTxCreateModelDataWithKeepers(t *testing.T) {
	assert := assert.New(t)

	ctx := NewMockChainContext()
	cs := ctx.MockChainState()

	sender := util.Signer1.Address()
	recipient := util.Signer2.Address()

	pm, err := service.ProfileModel()
	assert.NoError(err)
	mi := &ld.ModelInfo{
		Name:      pm.Name(),
		Threshold: 1,
		Keepers:   util.EthIDs{util.Signer2.Address()},
		Schema:    pm.Schema(),
		ID:        ctx.ChainConfig().ProfileServiceID,
	}

	pf := &service.Profile{
		Type:       1,
		Name:       "LDC",
		Follows:    util.DataIDs{},
		Extensions: []*service.Extension{},
	}
	assert.NoError(pf.SyntacticVerify())
	data := pf.Bytes()
	input := &ld.TxUpdater{
		ModelID:   &mi.ID,
		Version:   1,
		Threshold: ld.Uint16Ptr(0),
		Keepers:   &util.EthIDs{util.Signer1.Address()},
		Data:      data,
	}
	assert.NoError(input.SyntacticVerify())
	txData := &ld.TxData{
		Type:      ld.TypeCreateData,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      sender,
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	tt := txData.ToTransaction()
	itx, err := NewTx2(tt)
	assert.NoError(err)

	senderAcc := cs.MustAccount(sender)
	senderAcc.Add(constants.NativeToken, new(big.Int).SetUint64(constants.LDC))
	cs.CommitAccounts()
	assert.ErrorContains(itx.Apply(ctx, cs), "2FGxmZwYAuebXdEukTpj84EVKFVMK5fHu not found")
	cs.CheckoutAccounts()
	assert.NoError(cs.SaveModel(mi))
	cs.CommitAccounts()
	assert.ErrorContains(itx.Apply(ctx, cs), `TxCreateData.Apply error: nil to`)
	cs.CheckoutAccounts()

	input = &ld.TxUpdater{
		ModelID:   &mi.ID,
		Version:   1,
		Threshold: ld.Uint16Ptr(0),
		Keepers:   &util.EthIDs{util.Signer1.Address()},
		Data:      data,
		To:        &recipient,
	}
	assert.NoError(input.SyntacticVerify())
	txData = &ld.TxData{
		Type:      ld.TypeCreateData,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      sender,
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	tt = txData.ToTransaction()
	_, err = NewTx2(tt)
	assert.ErrorContains(err,
		"invalid to, expected 0x44171C37Ff5D7B7bb8dcad5C81f16284A229e641, got <nil>")

	txData = &ld.TxData{
		Type:      ld.TypeCreateData,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      sender,
		To:        &constants.GenesisAccount,
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	tt = txData.ToTransaction()
	_, err = NewTx2(tt)
	assert.ErrorContains(err,
		"invalid to, expected 0x44171C37Ff5D7B7bb8dcad5C81f16284A229e641, got 0xFFfFfFffFFfffFFfFFfFFFFFffFFFffffFfFFFfF")

	txData = &ld.TxData{
		Type:      ld.TypeCreateData,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      sender,
		To:        &recipient,
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	tt = txData.ToTransaction()
	tt.Timestamp = 10
	_, err = NewTx2(tt)
	assert.ErrorContains(err, "data expired")

	input = &ld.TxUpdater{
		ModelID:   &mi.ID,
		Version:   1,
		Threshold: ld.Uint16Ptr(0),
		Keepers:   &util.EthIDs{util.Signer1.Address()},
		Data:      data,
		To:        &recipient,
		Expire:    100,
	}
	assert.NoError(input.SyntacticVerify())
	txData = &ld.TxData{
		Type:      ld.TypeCreateData,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      sender,
		To:        &recipient,
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	tt = txData.ToTransaction()
	tt.Timestamp = 10
	_, err = NewTx2(tt)
	assert.ErrorContains(err, "nil amount")

	input = &ld.TxUpdater{
		ModelID:   &mi.ID,
		Version:   1,
		Threshold: ld.Uint16Ptr(0),
		Keepers:   &util.EthIDs{util.Signer1.Address()},
		Data:      data,
		To:        &recipient,
		Expire:    100,
		Amount:    new(big.Int).SetUint64(0),
	}
	assert.NoError(input.SyntacticVerify())
	txData = &ld.TxData{
		Type:      ld.TypeCreateData,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      sender,
		To:        &recipient,
		Amount:    new(big.Int).SetUint64(1),
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	tt = txData.ToTransaction()
	tt.Timestamp = 10
	_, err = NewTx2(tt)
	assert.ErrorContains(err, "invalid amount, expected 0, got 1")

	assert.NoError(input.SyntacticVerify())
	txData = &ld.TxData{
		Type:      ld.TypeCreateData,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      sender,
		To:        &recipient,
		Amount:    new(big.Int).SetUint64(0),
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	tt = txData.ToTransaction()
	tt.Timestamp = 10
	_, err = NewTx2(tt)
	assert.ErrorContains(err,
		"invalid exSignatures, Transaction.ExSigners error: DeriveSigners error: no signature")

	input = &ld.TxUpdater{
		ModelID:   &mi.ID,
		Version:   1,
		Threshold: ld.Uint16Ptr(1),
		Keepers:   &util.EthIDs{util.Signer1.Address()},
		Data:      data,
		To:        &recipient,
		Expire:    100,
		Amount:    new(big.Int).SetUint64(0),
	}
	assert.NoError(input.SyntacticVerify())
	txData = &ld.TxData{
		Type:      ld.TypeCreateData,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      sender,
		To:        &recipient,
		Amount:    new(big.Int).SetUint64(0),
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	assert.NoError(txData.ExSignWith(util.Signer1))
	tt = txData.ToTransaction()
	tt.Timestamp = 10
	itx, err = NewTx2(tt)
	assert.NoError(err)
	cs.CommitAccounts()
	assert.ErrorContains(itx.Apply(ctx, cs),
		"invalid exSignatures for model keepers")
	cs.CheckoutAccounts()

	txData = &ld.TxData{
		Type:      ld.TypeCreateData,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      sender,
		To:        &recipient,
		Amount:    new(big.Int).SetUint64(0),
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	assert.NoError(txData.ExSignWith(util.Signer2))
	tt = txData.ToTransaction()
	tt.Timestamp = 10
	itx, err = NewTx2(tt)
	assert.NoError(err)
	assert.NoError(itx.Apply(ctx, cs))

	senderGas := tt.Gas()
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
	assert.Equal(util.EthIDs{sender}, di.Keepers)
	assert.Nil(di.Approver)
	assert.Nil(di.ApproveList)
	assert.Equal(data, []byte(di.Data))

	p2 := &service.Profile{}
	assert.NoError(p2.Unmarshal(di.Data))
	assert.NoError(p2.SyntacticVerify())
	assert.Equal(pf.Type, p2.Type)
	assert.Equal(pf.Name, p2.Name)
	assert.Equal(pf.Follows, p2.Follows)
	assert.Equal(pf.Extensions, p2.Extensions)
	assert.Equal(pf.Bytes(), p2.Bytes())

	jsondata, err := itx.MarshalJSON()
	assert.NoError(err)
	// fmt.Println(string(jsondata))
	assert.Equal(`{"type":"TypeCreateData","chainID":2357,"nonce":0,"gasTip":100,"gasFeeCap":1000,"from":"0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC","to":"0x44171C37Ff5D7B7bb8dcad5C81f16284A229e641","amount":0,"data":{"mid":"2FGxmZwYAuebXdEukTpj84EVKFVMK5fHu","version":1,"threshold":1,"keepers":["0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC"],"to":"0x44171C37Ff5D7B7bb8dcad5C81f16284A229e641","amount":0,"expire":100,"data":"0xa7616460616960616e634c444361740161756062657880626673804aa7ee41"},"signatures":["fbd88343c51942fa1ec0c2c5539a1937622fa0d913028ee2bb03e8d4575da3052e1c3f3ac960051a7af49873fdd27ab30c4f247b758f6dd2f15e56c666a90ca801"],"exSignatures":["9dc099eca03d85ed7058f9cf7a1b5e8bdfae60542c973527dab69958da0f1da8472efc71b1fd9d1be6b3386cce589c370632024b429d549ee0476474422397d500"],"id":"2oGuwdYVzBqd4LZqdUHJwBVbV1ERjbwdv5KnHK8c8QRbEbGvoc"}`, string(jsondata))

	assert.NoError(cs.VerifyState())
}

func TestTxCreateNameModelData(t *testing.T) {
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
		Name:    "ldc.to.",
		Records: []string{"ldc.to. IN A 10.0.0.1"},
	}
	assert.NoError(name.SyntacticVerify())
	data := name.Bytes()

	input := &ld.TxUpdater{
		ModelID:   &mi.ID,
		Version:   1,
		Threshold: ld.Uint16Ptr(1),
		Keepers:   &util.EthIDs{util.Signer1.Address()},
		Data:      data,
		To:        &recipient,
		Expire:    100,
		Amount:    new(big.Int).SetUint64(constants.MilliLDC),
	}
	assert.NoError(input.SyntacticVerify())
	txData := &ld.TxData{
		Type:      ld.TypeCreateData,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      sender,
		To:        &recipient,
		Amount:    new(big.Int).SetUint64(constants.MilliLDC),
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	assert.NoError(txData.ExSignWith(util.Signer2))

	senderAcc := cs.MustAccount(sender)
	assert.NoError(senderAcc.Add(constants.NativeToken, new(big.Int).SetUint64(constants.LDC)))
	assert.NoError(cs.SaveModel(mi))

	tt := txData.ToTransaction()
	tt.Timestamp = 10
	itx, err := NewTx2(tt)
	assert.NoError(err)

	_, err = cs.LoadDataByName("ldc.to.")
	assert.ErrorContains(err, `"ldc.to." not found`)
	assert.NoError(itx.Apply(ctx, cs))

	senderGas := tt.Gas()
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
	assert.Equal(util.EthIDs{sender}, di.Keepers)
	assert.Nil(di.Approver)
	assert.Nil(di.ApproveList)
	assert.Equal(data, []byte(di.Data))

	n2 := &service.Name{}
	assert.NoError(n2.Unmarshal(di.Data))
	assert.NoError(n2.SyntacticVerify())
	assert.Equal(n2.Name, name.Name)
	assert.Equal(n2.Records, name.Records)
	assert.Equal(n2.Bytes(), name.Bytes())

	jsondata, err := itx.MarshalJSON()
	assert.NoError(err)
	// fmt.Println(string(jsondata))
	assert.Equal(`{"type":"TypeCreateData","chainID":2357,"nonce":0,"gasTip":100,"gasFeeCap":1000,"from":"0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC","to":"0x44171C37Ff5D7B7bb8dcad5C81f16284A229e641","amount":1000000,"data":{"mid":"CMpsSKUM1dfWVyYMpRHuQfHHYTUWcV6PQ","version":1,"threshold":1,"keepers":["0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC"],"to":"0x44171C37Ff5D7B7bb8dcad5C81f16284A229e641","amount":1000000,"expire":100,"data":"0xa2616e676c64632e746f2e62727381756c64632e746f2e20494e20412031302e302e302e31851ac289"},"signatures":["94ed805397ddfedb106a5a3f91e9434d573c238ebdf1535f7c54d7a344133c161cd8d737d0a3a811d5aec4e7ad53d3507d829bd6be2114fc2586604fdb18e3c200"],"exSignatures":["33dfd5a016ce6a9aa1de4e1995fab94b48178003a46d8707203aeed7f57bf2b962643f05e95b0f681e877c9ab6b1cc3df43712c3012a080daf1184588cab6f9801"],"id":"U9M2MvTEjmwCqKfsG9w9YJ8BaouiyJuVDhaY4rXsauduTy2gn"}`, string(jsondata))

	assert.NoError(cs.VerifyState())

	name2 := &service.Name{
		Name:    "ldc.to.",
		Records: []string{"ldc.to. IN A 10.0.0.2"},
	}
	assert.NoError(name2.SyntacticVerify())
	data = name2.Bytes()

	input = &ld.TxUpdater{
		ModelID:   &mi.ID,
		Version:   1,
		Threshold: ld.Uint16Ptr(1),
		Keepers:   &util.EthIDs{util.Signer1.Address()},
		Data:      data,
		To:        &recipient,
		Expire:    100,
		Amount:    new(big.Int).SetUint64(constants.MilliLDC),
	}
	assert.NoError(input.SyntacticVerify())
	txData = &ld.TxData{
		Type:      ld.TypeCreateData,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     1,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      sender,
		To:        &recipient,
		Amount:    new(big.Int).SetUint64(constants.MilliLDC),
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	assert.NoError(txData.ExSignWith(util.Signer2))

	tt = txData.ToTransaction()
	tt.Timestamp = 10
	itx, err = NewTx2(tt)
	assert.NoError(err)
	cs.CommitAccounts()
	assert.ErrorContains(itx.Apply(ctx, cs),
		`TxCreateData.Apply error: name "ldc.to." conflict`)
	cs.CheckoutAccounts()

	name2 = &service.Name{
		Name:    "api.ldc.to.",
		Records: []string{},
	}
	assert.NoError(name2.SyntacticVerify())
	data = name2.Bytes()

	input = &ld.TxUpdater{
		ModelID:   &mi.ID,
		Version:   1,
		Threshold: ld.Uint16Ptr(1),
		Keepers:   &util.EthIDs{util.Signer1.Address()},
		Data:      data,
		To:        &recipient,
		Expire:    100,
		Amount:    new(big.Int).SetUint64(constants.MilliLDC),
	}

	assert.NoError(input.SyntacticVerify())
	txData = &ld.TxData{
		Type:      ld.TypeCreateData,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     1,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      sender,
		To:        &recipient,
		Amount:    new(big.Int).SetUint64(constants.MilliLDC),
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	assert.NoError(txData.ExSignWith(util.Signer2))

	tt = txData.ToTransaction()
	tt.Timestamp = 10
	itx, err = NewTx2(tt)
	assert.NoError(err)
	assert.NoError(itx.Apply(ctx, cs))
}

func TestTxCreateDataGenesis(t *testing.T) {
	assert := assert.New(t)

	ctx := NewMockChainContext()
	cs := ctx.MockChainState()

	sender := util.Signer1.Address()

	cfg, err := json.Marshal(ctx.ChainConfig().FeeConfig)
	assert.NoError(err)

	cfgData := &ld.TxUpdater{
		ModelID:   &ld.JSONModelID,
		Version:   1,
		Threshold: ld.Uint16Ptr(1),
		Keepers:   &util.EthIDs{util.Signer1.Address()},
		Data:      cfg,
	}
	assert.NoError(cfgData.SyntacticVerify())
	tt := &ld.Transaction{
		Type:    ld.TypeCreateData,
		ChainID: ctx.ChainConfig().ChainID,
		From:    sender,
		Data:    ld.MustMarshal(cfgData),
	}
	assert.NoError(tt.SyntacticVerify())

	itx, err := NewGenesisTx(tt)
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
	assert.Equal(util.EthIDs{sender}, di.Keepers)
	assert.Nil(di.Approver)
	assert.Nil(di.ApproveList)
	assert.True(jsonpatch.Equal(cfg, di.Data))

	jsondata, err := itx.MarshalJSON()
	assert.NoError(err)
	// fmt.Println(string(jsondata))
	assert.Equal(`{"type":"TypeCreateData","chainID":2357,"nonce":0,"gasTip":0,"gasFeeCap":0,"from":"0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC","data":{"mid":"1111111111111111111L17Xp3","version":1,"threshold":1,"keepers":["0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC"],"data":{"startHeight":0,"thresholdGas":1000,"minGasPrice":10000,"maxGasPrice":100000,"maxTxGas":42000000,"maxBlockTxsSize":4200000,"gasRebateRate":1000,"minTokenPledge":10000000000000,"minStakePledge":1000000000000}},"id":"7o7fYNFS27SGZF8uEXC8PExi5mgVW4p4Hj7gdCg1wEqCUB1qk"}`, string(jsondata))

	assert.NoError(cs.VerifyState())
}
