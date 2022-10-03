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
	assert.ErrorContains(err, "DeriveSigners error: no signature")

	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeCreateData,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      sender,
		Token:     &token,
	}}
	assert.NoError(ltx.SignWith(util.Signer1))
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
	assert.NoError(ltx.SignWith(util.Signer1))
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
	assert.NoError(ltx.SignWith(util.Signer1))
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
	assert.NoError(ltx.SignWith(util.Signer1))
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
	assert.NoError(ltx.SignWith(util.Signer1))
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
	assert.NoError(ltx.SignWith(util.Signer1))
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
	assert.NoError(ltx.SignWith(util.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "no keepers, threshold should be nil")

	input = &ld.TxUpdater{
		ModelID:   &ld.RawModelID,
		Version:   1,
		Threshold: ld.Uint16Ptr(0),
		Keepers:   &util.EthIDs{},
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
	assert.NoError(ltx.SignWith(util.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "empty keepers")

	input = &ld.TxUpdater{
		ModelID:   &ld.RawModelID,
		Version:   1,
		Threshold: ld.Uint16Ptr(0),
		Keepers:   &util.EthIDs{util.Signer1.Address()},
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
	assert.NoError(ltx.SignWith(util.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "empty data")

	input = &ld.TxUpdater{
		ModelID:   &ld.RawModelID,
		Version:   1,
		Threshold: ld.Uint16Ptr(0),
		Keepers:   &util.EthIDs{util.Signer1.Address()},
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
	assert.NoError(ltx.SignWith(util.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
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
	assert.NoError(ltx.SignWith(util.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "invalid to, should be nil")

	input = &ld.TxUpdater{
		ModelID:   &ld.RawModelID,
		Version:   1,
		Threshold: ld.Uint16Ptr(0),
		Keepers:   &util.EthIDs{util.Signer1.Address()},
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
	assert.NoError(ltx.SignWith(util.Signer1))
	assert.ErrorContains(ltx.SyntacticVerify(), "nil \"to\" together with amount")

	input = &ld.TxUpdater{
		ModelID:   &ld.RawModelID,
		Version:   1,
		Threshold: ld.Uint16Ptr(0),
		Keepers:   &util.EthIDs{util.Signer1.Address()},
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
	assert.NoError(ltx.SignWith(util.Signer1))
	assert.NoError(ltx.ExSignWith(util.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "invalid exSignatures, should be nil")

	input = &ld.TxUpdater{
		ModelID:   &ld.RawModelID,
		Version:   1,
		Threshold: ld.Uint16Ptr(0),
		Keepers:   &util.EthIDs{util.Signer1.Address()},
		Data:      []byte(`42`),
		TypedSig:  util.Signature{1, 2, 3}.Typed(),
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
	assert.NoError(ltx.SignWith(util.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "invalid sigClaims, should be nil")

	input = &ld.TxUpdater{
		ModelID:   &ld.RawModelID,
		Version:   1,
		Threshold: ld.Uint16Ptr(0),
		Keepers:   &util.EthIDs{util.Signer1.Address()},
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
	assert.NoError(ltx.SignWith(util.Signer1))
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
	assert.Equal(util.EthIDs{sender}, di.Keepers)
	assert.Nil(di.Approver)
	assert.Nil(di.ApproveList)
	assert.Equal([]byte(`42`), []byte(di.Payload))

	jsondata, err := itx.MarshalJSON()
	assert.NoError(err)
	// fmt.Println(string(jsondata))
	assert.Equal(`{"tx":{"type":"TypeCreateData","chainID":2357,"nonce":0,"gasTip":100,"gasFeeCap":1000,"from":"0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC","data":{"mid":"111111111111111111116DBWJs","version":1,"threshold":0,"keepers":["0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC"],"data":42}},"sigs":["b6d67cf0d07327bae70469845c65c7a0edb50f5495b36dd9d74a6ceff8fccbab71f1c3a3384759b0f25d4c2b86c4caea70f4bed89add3c1dccc65a85f4592f5901"],"id":"21R8A5wmzAEefxWYg23vz5jW34tJv6C5amNUV7ZHgQ6sKpqnxr"}`, string(jsondata))

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
		Keepers:   &util.EthIDs{util.Signer1.Address()},
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
	assert.NoError(ltx.SignWith(util.Signer1))
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
	assert.Equal(util.EthIDs{sender}, di.Keepers)
	assert.Nil(di.Approver)
	assert.Nil(di.ApproveList)
	assert.Equal(data, []byte(di.Payload))

	jsondata, err := itx.MarshalJSON()
	assert.NoError(err)
	// fmt.Println(string(jsondata))
	assert.Equal(`{"tx":{"type":"TypeCreateData","chainID":2357,"nonce":0,"gasTip":100,"gasFeeCap":1000,"from":"0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC","data":{"mid":"1111111111111111111Ax1asG","version":1,"threshold":0,"keepers":["0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC"],"data":"0xa2626e616474657374626e6f830102031e0946b2"}},"sigs":["0973e7f973a6d332bc6f43391b48fbd51eb01bf3f0e1606e702b06e58c507fa519904657c9d634a96d6fcadd3028aab5ab7cb5e6c7502c9911df9f4a3afc569201"],"id":"9AU8c755eCawyiYYaqsZNxJwFkSJ2HHmTRAvuCiHB3qUvWydZ"}`, string(jsondata))

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
		Keepers:   &util.EthIDs{util.Signer1.Address()},
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
	assert.NoError(ltx.SignWith(util.Signer1))
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
	assert.Equal(util.EthIDs{sender}, di.Keepers)
	assert.Nil(di.Approver)
	assert.Nil(di.ApproveList)
	assert.Equal(data, []byte(di.Payload))

	jsondata, err := itx.MarshalJSON()
	assert.NoError(err)
	// fmt.Println(string(jsondata))
	assert.Equal(`{"tx":{"type":"TypeCreateData","chainID":2357,"nonce":0,"gasTip":100,"gasFeeCap":1000,"from":"0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC","data":{"mid":"1111111111111111111L17Xp3","version":1,"threshold":0,"keepers":["0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC"],"data":{"na":"test","no":[1,2,3]}}},"sigs":["8a56cbb69abcf6fcbe55a15d68cf30bc67c86f6b90e2db5f8410ef4ef800fb8b2d042562b7596bb3aa1dd27fba9596b620f354db78a24b270e79e12dc6cbd56901"],"id":"27MXPREjeRK3PwJfe9vJ7HjKXVCWsskBzun5S4LivKXKx59brt"}`, string(jsondata))

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
		Extensions: service.Extensions{},
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
	itx, err := NewTx(ltx)
	assert.NoError(err)

	cs.CommitAccounts()
	assert.ErrorContains(itx.Apply(ctx, cs),
		"insufficient NativeLDC balance, expected 1250700, got 0")
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
	ltx = &ld.Transaction{Tx: ld.TxData{
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
	assert.Equal(util.EthIDs{sender}, di.Keepers)
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
	assert.Equal(`{"tx":{"type":"TypeCreateData","chainID":2357,"nonce":0,"gasTip":100,"gasFeeCap":1000,"from":"0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC","data":{"mid":"6L5yRNNMubYqZoZRtmk1ykJMmZppNwb1","version":1,"threshold":1,"keepers":["0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC"],"data":"0xa7616460616960616e6674657374657261740161756062657380626673800f668b8f"}},"sigs":["b458c06845f6801cc7dba3a408f317f6ecb309300bae69294f5795db860828c168bd7e357ba66c871ec3f28f76bbd73f5192103432ff3a97b61bb96fb3280c1800"],"id":"2UiySW8Ha1AeKkojivQR1zQDTQ41Fi6tTGHd8KyLchbSwgFKia"}`, string(jsondata))

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
		Keepers:   &util.EthIDs{util.Signer1.Address()},
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
	assert.NoError(ltx.SignWith(util.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	itx, err := NewTx(ltx)
	assert.NoError(err)

	senderAcc := cs.MustAccount(sender)
	senderAcc.Add(constants.NativeToken, new(big.Int).SetUint64(constants.LDC))
	cs.CommitAccounts()
	assert.ErrorContains(itx.Apply(ctx, cs), "6L5yB2u4uKaHNHEMc4ygsv9c58ZNDTE4 not found")
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
	ltx = &ld.Transaction{Tx: ld.TxData{
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
	_, err = NewTx(ltx)
	assert.ErrorContains(err,
		"invalid to, expected 0x44171C37Ff5D7B7bb8dcad5C81f16284A229e641, got <nil>")

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
	assert.NoError(ltx.SignWith(util.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err,
		"invalid to, expected 0x44171C37Ff5D7B7bb8dcad5C81f16284A229e641, got 0xFFfFfFffFFfffFFfFFfFFFFFffFFFffffFfFFFfF")

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
	assert.NoError(ltx.SignWith(util.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	ltx.Timestamp = 10
	_, err = NewTx(ltx)
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
	assert.NoError(ltx.SignWith(util.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	ltx.Timestamp = 10
	_, err = NewTx(ltx)
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
	assert.NoError(ltx.SignWith(util.Signer1))
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
	assert.NoError(ltx.SignWith(util.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	ltx.Timestamp = 10
	_, err = NewTx(ltx)
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
	assert.NoError(ltx.SignWith(util.Signer1))
	assert.NoError(ltx.ExSignWith(util.Signer1))
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
	assert.NoError(ltx.SignWith(util.Signer1))
	assert.NoError(ltx.ExSignWith(util.Signer2))
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
	assert.Equal(util.EthIDs{sender}, di.Keepers)
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
	assert.Equal(`{"tx":{"type":"TypeCreateData","chainID":2357,"nonce":0,"gasTip":100,"gasFeeCap":1000,"from":"0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC","to":"0x44171C37Ff5D7B7bb8dcad5C81f16284A229e641","amount":0,"data":{"mid":"6L5yB2u4uKaHNHEMc4ygsv9c58ZNDTE4","version":1,"threshold":1,"keepers":["0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC"],"to":"0x44171C37Ff5D7B7bb8dcad5C81f16284A229e641","amount":0,"expire":100,"data":"0xa7616460616960616e634c444361740161756062657380626673802e82fe4d"}},"sigs":["3be83660f8903004c4910843ad716e472aab5920b0dc72083b21809b65c057673877ea2a9acf5fac5cc489fd23b40554d49551fd4100b9c7509ce36d4d803bb401"],"exSigs":["5ecdaa105760371558a82ba8090ad84d37a1b1f4afb1234b252b4408704c814f6c1227062bbcc8b967e863dcfe64ff357f288e92e0f95e4f275c81513f83ecce01"],"id":"2KXrsJYq4bofU2TPdwLKmZ98LvY9BWgBB6JHYxCLKX25EdFr7m"}`, string(jsondata))

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
		Keepers:   &util.EthIDs{util.Signer1.Address()},
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
	assert.Equal(util.EthIDs{sender}, di.Keepers)
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
	assert.Equal(`{"tx":{"type":"TypeCreateData","chainID":2357,"nonce":0,"gasTip":100,"gasFeeCap":1000,"from":"0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC","to":"0x44171C37Ff5D7B7bb8dcad5C81f16284A229e641","amount":1000000,"data":{"mid":"G61B2NvDV1bG57M1skG1ZAbu5g1uVBQHX","version":1,"threshold":1,"keepers":["0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC"],"to":"0x44171C37Ff5D7B7bb8dcad5C81f16284A229e641","amount":1000000,"expire":100,"data":"0xa3616e676c64632e746f2e62657381a261746454657374627073a16464657363646465736362727381756c64632e746f2e20494e20412031302e302e302e31e5c01479"}},"sigs":["8928c45b73d9d7afdcd42799fad24f508a2cd7aa5be01a715828b909eb4a1a31281fb325443964d59f5b9aaec1d2f7c4f8c4b947331caeba0337ac11dbed7aa300"],"exSigs":["18ec645e4353142ea3c4759c1f65f0584ac3fcb3176dc87fdbd43d37d89293e41e96b92c7160a861906886f6e7322ab3208761180df6c4332d045e49f5816e7700"],"id":"2SJhyeUfVx9azeRnKkBxuhpN3WANM6fg3ZJVD4HE4injyYHhqA"}`, string(jsondata))

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
		Keepers:   &util.EthIDs{util.Signer1.Address()},
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
	assert.NoError(ltx.SignWith(util.Signer1))
	assert.NoError(ltx.ExSignWith(util.Signer2))
	assert.NoError(ltx.SyntacticVerify())
	ltx.Timestamp = 10
	itx, err = NewTx(ltx)
	assert.NoError(err)
	cs.CommitAccounts()
	assert.ErrorContains(itx.Apply(ctx, cs),
		`TxCreateData.Apply error: name "ldc.to." conflict`)
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
		Keepers:   &util.EthIDs{util.Signer1.Address()},
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
	assert.NoError(ltx.SignWith(util.Signer1))
	assert.NoError(ltx.ExSignWith(util.Signer2))
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
	assert.Equal(util.EthIDs{sender}, di.Keepers)
	assert.Nil(di.Approver)
	assert.Nil(di.ApproveList)
	assert.True(jsonpatch.Equal(cfg, di.Payload))

	jsondata, err := itx.MarshalJSON()
	assert.NoError(err)
	// fmt.Println(string(jsondata))
	assert.Equal(`{"tx":{"type":"TypeCreateData","chainID":2357,"nonce":0,"gasTip":0,"gasFeeCap":0,"from":"0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC","data":{"mid":"1111111111111111111L17Xp3","version":1,"threshold":1,"keepers":["0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC"],"data":{"startHeight":0,"thresholdGas":1000,"minGasPrice":10000,"maxGasPrice":100000,"maxTxGas":42000000,"maxBlockTxsSize":4200000,"gasRebateRate":1000,"minTokenPledge":10000000000000,"minStakePledge":1000000000000}}},"id":"29LeE5wUNmyHLgpgEbhG5fwi3KWprSEy7tfeJS8ATfkkuYohf4"}`, string(jsondata))

	assert.NoError(cs.VerifyState())
}
