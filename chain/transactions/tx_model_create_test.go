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

func TestTxCreateModel(t *testing.T) {
	assert := assert.New(t)

	// SyntacticVerify
	tx := &TxCreateModel{}
	assert.ErrorContains(tx.SyntacticVerify(), "nil pointer")
	_, err := tx.MarshalJSON()
	assert.NoError(err)

	ctx := NewMockChainContext()
	cs := ctx.MockChainState()
	token := ld.MustNewToken("$LDC")

	owner := util.Signer1.Address()

	txData := &ld.TxData{
		Type:      ld.TypeCreateModel,
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
		Type:      ld.TypeCreateModel,
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
		Type:      ld.TypeCreateModel,
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
		Type:      ld.TypeCreateModel,
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
		Type:      ld.TypeCreateModel,
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
		Type:      ld.TypeCreateModel,
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

	input := &ld.ModelInfo{}
	assert.ErrorContains(input.SyntacticVerify(), "ModelInfo.SyntacticVerify error: invalid name")
	ipldm, err := service.ProfileModel()
	assert.NoError(err)
	input = &ld.ModelInfo{
		Name:      ipldm.Name(),
		Threshold: 1,
		Keepers:   util.EthIDs{util.Signer1.Address()},
		Schema:    ipldm.Schema(),
	}
	txData = &ld.TxData{
		Type:      ld.TypeCreateModel,
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
		"insufficient NativeLDC balance, expected 4079900, got 0")
	cs.CheckoutAccounts()

	ownerAcc := cs.MustAccount(owner)
	ownerAcc.Add(constants.NativeToken, new(big.Int).SetUint64(constants.LDC))
	assert.NoError(itx.Apply(ctx, cs))

	fromGas := tt.Gas()
	assert.Equal(fromGas*ctx.Price,
		itx.(*TxCreateModel).ldc.Balance().Uint64())
	assert.Equal(fromGas*100,
		itx.(*TxCreateModel).miner.Balance().Uint64())
	assert.Equal(constants.LDC-fromGas*(ctx.Price+100),
		ownerAcc.Balance().Uint64())
	assert.Equal(uint64(1), ownerAcc.Nonce())

	mi, err := cs.LoadModel(itx.(*TxCreateModel).input.ID)
	assert.NoError(err)
	assert.Equal(input.Name, mi.Name)
	assert.Equal(input.Schema, mi.Schema)
	assert.Equal(input.Threshold, mi.Threshold)
	assert.Equal(input.Keepers, mi.Keepers)

	jsondata, err := itx.MarshalJSON()
	assert.NoError(err)
	// fmt.Println(string(jsondata))
	assert.Equal(`{"type":"TypeCreateModel","chainID":2357,"nonce":0,"gasTip":100,"gasFeeCap":1000,"from":"0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC","data":{"name":"ProfileService","threshold":1,"keepers":["0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC"],"schema":"type ID20 bytes\n\ttype ProfileService struct {\n\t\ttype        Int             (rename \"t\")\n\t\tname        String          (rename \"n\")\n\t\tdescription String          (rename \"d\")\n\t\timage       String          (rename \"i\")\n\t\turl         String          (rename \"u\")\n\t\tfollows     [ID20]          (rename \"fs\")\n\t\tmembers     optional [ID20] (rename \"ms\")\n\t\textensions  [Any]           (rename \"es\")\n\t}","id":"FWLyyrjop29R2urXcaaaE46DTVdr9eube"},"signatures":["c9cea46c2c7fd90f19527247e058725bfae8239d2fd58fa8c0df638ca8f98fdf2e2110a9ed4fbb875727e7289edf1260c438a4b167a2b52c64fd4514fe0ad8e601"],"id":"2D5V9ctBiKfRmhSkncyBMpWuG9CU6tExmr85M5tHKteh1jANzE"}`, string(jsondata))

	modelAcc := cs.MustAccount(util.EthID(mi.ID))
	assert.Equal(input.Threshold, modelAcc.Threshold())
	assert.Equal(input.Keepers, modelAcc.Keepers())

	// transfer token to model account
	txData = &ld.TxData{
		Type:      ld.TypeTransfer,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     1,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      owner,
		To:        &modelAcc.id,
		Amount:    new(big.Int).SetUint64(constants.MilliLDC * 500),
	}
	assert.NoError(txData.SyntacticVerify())
	assert.NoError(txData.SignWith(util.Signer1))

	tt = txData.ToTransaction()
	itx, err = NewTx2(tt)
	assert.NoError(err)
	assert.NoError(itx.Apply(ctx, cs))

	fromGas += tt.Gas()
	assert.Equal(fromGas*ctx.Price,
		itx.(*TxTransfer).ldc.Balance().Uint64())
	assert.Equal(fromGas*100,
		itx.(*TxTransfer).miner.Balance().Uint64())
	assert.Equal(constants.MilliLDC*500, modelAcc.Balance().Uint64())
	assert.Equal(constants.LDC-fromGas*(ctx.Price+100)-constants.MilliLDC*500,
		ownerAcc.Balance().Uint64())
	assert.Equal(uint64(2), ownerAcc.Nonce())
	assert.True(ownerAcc.IsEmpty())
	assert.False(modelAcc.IsEmpty())

	// transfer token from model account
	txData = &ld.TxData{
		Type:      ld.TypeTransfer,
		ChainID:   ctx.ChainConfig().ChainID,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      modelAcc.ID(),
		To:        &ownerAcc.id,
		Amount:    new(big.Int).SetUint64(constants.MilliLDC * 100),
	}
	assert.NoError(txData.SyntacticVerify())
	assert.NoError(txData.SignWith(util.Signer1))

	tt = txData.ToTransaction()
	itx, err = NewTx2(tt)
	assert.NoError(err)
	assert.NoError(itx.Apply(ctx, cs))

	modelGas := tt.Gas()
	assert.Equal((fromGas+modelGas)*ctx.Price,
		itx.(*TxTransfer).ldc.Balance().Uint64())
	assert.Equal((fromGas+modelGas)*100,
		itx.(*TxTransfer).miner.Balance().Uint64())
	assert.Equal(constants.MilliLDC*400-modelGas*(ctx.Price+100), modelAcc.Balance().Uint64())
	assert.Equal(constants.LDC-fromGas*(ctx.Price+100)-constants.MilliLDC*400,
		ownerAcc.Balance().Uint64())
	assert.Equal(uint64(1), modelAcc.Nonce())

	assert.NoError(cs.VerifyState())
}

func TestTxCreateModelGenesis(t *testing.T) {
	assert := assert.New(t)

	ctx := NewMockChainContext()
	cs := ctx.MockChainState()
	owner := util.Signer1.Address()

	nm, err := service.NameModel()
	assert.NoError(err)
	mi := &ld.ModelInfo{
		Name:      nm.Name(),
		Threshold: 1,
		Keepers:   util.EthIDs{util.Signer1.Address()},
		Schema:    nm.Schema(),
	}
	assert.NoError(mi.SyntacticVerify())

	tt := &ld.Transaction{
		Type:    ld.TypeCreateModel,
		ChainID: ctx.ChainConfig().ChainID,
		From:    owner,
		Data:    ld.MustMarshal(mi),
	}
	assert.NoError(tt.SyntacticVerify())

	itx, err := NewGenesisTx(tt)
	assert.NoError(err)
	assert.NoError(itx.(GenesisTx).ApplyGenesis(ctx, cs))

	assert.Equal(uint64(0), itx.(*TxCreateModel).ldc.Balance().Uint64())
	assert.Equal(uint64(0), itx.(*TxCreateModel).miner.Balance().Uint64())
	assert.Equal(uint64(0), itx.(*TxCreateModel).from.Balance().Uint64())
	assert.Equal(uint64(1), itx.(*TxCreateModel).from.Nonce())

	mi2, err := cs.LoadModel(itx.(*TxCreateModel).input.ID)
	assert.NoError(err)
	assert.Equal(uint16(1), mi2.Threshold)
	assert.Equal(util.EthIDs{owner}, mi2.Keepers)
	assert.Nil(mi2.Approver)
	assert.Equal(mi.Name, mi2.Name)
	assert.Equal(mi.Schema, mi2.Schema)

	jsondata, err := itx.MarshalJSON()
	assert.NoError(err)
	// fmt.Println(string(jsondata))
	assert.Equal(`{"type":"TypeCreateModel","chainID":2357,"nonce":0,"gasTip":0,"gasFeeCap":0,"from":"0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC","data":{"name":"NameService","threshold":1,"keepers":["0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC"],"schema":"type ID20 bytes\n\ttype NameService struct {\n\t\tname       String        (rename \"n\")\n\t\tlinked     optional ID20 (rename \"l\")\n\t\trecords    [String]      (rename \"rs\")\n\t\textensions [Any]         (rename \"es\")\n\t}","id":"7MJdfcQqjZiMogaPToZX1yJFvQvaYZ6t4"},"id":"XgFmsDsxNX6ZL2VcggxtNSb74ovVBqT68FEXKTa4JY4as5L9p"}`, string(jsondata))

	pm, err := service.ProfileModel()
	assert.NoError(err)
	mi = &ld.ModelInfo{
		Name:      pm.Name(),
		Threshold: 1,
		Keepers:   util.EthIDs{util.Signer1.Address()},
		Schema:    pm.Schema(),
	}
	assert.NoError(mi.SyntacticVerify())

	tt = &ld.Transaction{
		Type:    ld.TypeCreateModel,
		ChainID: ctx.ChainConfig().ChainID,
		Nonce:   1,
		From:    owner,
		Data:    ld.MustMarshal(mi),
	}
	assert.NoError(tt.SyntacticVerify())

	itx, err = NewGenesisTx(tt)
	assert.NoError(err)
	assert.NoError(itx.(GenesisTx).ApplyGenesis(ctx, cs))

	assert.Equal(uint64(0), itx.(*TxCreateModel).ldc.Balance().Uint64())
	assert.Equal(uint64(0), itx.(*TxCreateModel).miner.Balance().Uint64())
	assert.Equal(uint64(0), itx.(*TxCreateModel).from.Balance().Uint64())
	assert.Equal(uint64(2), itx.(*TxCreateModel).from.Nonce())

	mi2, err = cs.LoadModel(itx.(*TxCreateModel).input.ID)
	assert.NoError(err)
	assert.Equal(uint16(1), mi2.Threshold)
	assert.Equal(util.EthIDs{owner}, mi2.Keepers)
	assert.Nil(mi2.Approver)
	assert.Equal(mi.Name, mi2.Name)
	assert.Equal(mi.Schema, mi2.Schema)

	jsondata, err = itx.MarshalJSON()
	assert.NoError(err)
	// fmt.Println(string(jsondata))
	assert.Equal(`{"type":"TypeCreateModel","chainID":2357,"nonce":1,"gasTip":0,"gasFeeCap":0,"from":"0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC","data":{"name":"ProfileService","threshold":1,"keepers":["0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC"],"schema":"type ID20 bytes\n\ttype ProfileService struct {\n\t\ttype        Int             (rename \"t\")\n\t\tname        String          (rename \"n\")\n\t\tdescription String          (rename \"d\")\n\t\timage       String          (rename \"i\")\n\t\turl         String          (rename \"u\")\n\t\tfollows     [ID20]          (rename \"fs\")\n\t\tmembers     optional [ID20] (rename \"ms\")\n\t\textensions  [Any]           (rename \"es\")\n\t}","id":"KRKV4ykzkaWBf5f1hJ7SfMQd1kbsmwNFP"},"id":"2Xyz5As3Gp5EFPyuRuek9nHEBASeSZ6Gz7KuLhLwuAvq9JixMg"}`, string(jsondata))

	assert.NoError(cs.VerifyState())
}
