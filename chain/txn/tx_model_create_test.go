// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package txn

import (
	"math/big"
	"testing"

	"github.com/ldclabs/ldvm/ids"
	"github.com/ldclabs/ldvm/ld"
	"github.com/ldclabs/ldvm/ld/service"
	"github.com/ldclabs/ldvm/signer"
	"github.com/ldclabs/ldvm/unit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTxCreateModel(t *testing.T) {
	assert := assert.New(t)

	// SyntacticVerify
	tx := &TxCreateModel{}
	assert.ErrorContains(tx.SyntacticVerify(), "nil pointer")
	_, err := tx.MarshalJSON()
	require.NoError(t, err)

	ctx := NewMockChainContext()
	cs := ctx.MockChainState()
	token := ld.MustNewToken("$LDC")
	owner := signer.Signer1.Key().Address()

	ltx := &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeCreateModel,
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
		Type:      ld.TypeCreateModel,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      owner,
		To:        ids.GenesisAccount.Ptr(),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "invalid to, should be nil")

	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeCreateModel,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      owner,
		Token:     token.Ptr(),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "invalid token, should be nil")

	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeCreateModel,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      owner,
		Amount:    big.NewInt(1),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.ErrorContains(ltx.SyntacticVerify(), "nil \"to\" together with amount")

	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeCreateModel,
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
		Type:      ld.TypeCreateModel,
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

	input := &ld.ModelInfo{}
	assert.ErrorContains(input.SyntacticVerify(), "ModelInfo.SyntacticVerify: invalid name")
	ipldm, err := service.ProfileModel()
	require.NoError(t, err)
	input = &ld.ModelInfo{
		Name:      ipldm.Name(),
		Threshold: 1,
		Keepers:   signer.Keys{signer.Signer1.Key()},
		Schema:    ipldm.Schema(),
	}
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeCreateModel,
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
	require.NoError(t, err)

	cs.CommitAccounts()
	assert.ErrorContains(itx.Apply(ctx, cs),
		"insufficient NativeLDC balance, expected 4143700, got 0")
	cs.CheckoutAccounts()

	ownerAcc := cs.MustAccount(owner)
	ownerAcc.Add(ids.NativeToken, new(big.Int).SetUint64(unit.LDC*3))
	assert.NoError(itx.Apply(ctx, cs))

	fromGas := ltx.Gas()
	assert.Equal(fromGas*ctx.Price,
		itx.(*TxCreateModel).ldc.Balance().Uint64())
	assert.Equal(fromGas*100,
		itx.(*TxCreateModel).miner.Balance().Uint64())
	assert.Equal(unit.LDC*3-fromGas*(ctx.Price+100),
		ownerAcc.BalanceOfAll(ids.NativeToken).Uint64())
	assert.Equal(uint64(1), ownerAcc.Nonce())

	mi, err := cs.LoadModel(itx.(*TxCreateModel).input.ID)
	require.NoError(t, err)
	assert.Equal(input.Name, mi.Name)
	assert.Equal(input.Schema, mi.Schema)
	assert.Equal(input.Threshold, mi.Threshold)
	assert.Equal(input.Keepers, mi.Keepers)

	jsondata, err := itx.MarshalJSON()
	require.NoError(t, err)
	// fmt.Println(string(jsondata))
	assert.Equal(`{"tx":{"type":"TypeCreateModel","chainID":2357,"nonce":0,"gasTip":100,"gasFeeCap":1000,"from":"0x8db97c7cECe249C2b98bdc0226cc4C2A57bF52fc","data":{"name":"ProfileService","threshold":1,"keepers":["jbl8fOziScK5i9wCJsxMKle_UvwKxwPH"],"schema":"type ID20 bytes\n\ttype ProfileService struct {\n\t\ttype        Int             (rename \"t\")\n\t\tname        String          (rename \"n\")\n\t\tdescription String          (rename \"d\")\n\t\timage       String          (rename \"i\")\n\t\turl         String          (rename \"u\")\n\t\tfollows     [ID20]          (rename \"fs\")\n\t\tmembers     optional [ID20] (rename \"ms\")\n\t\textensions  [Any]           (rename \"es\")\n\t}","id":"HY5mXzOlgOcwc-xstUV7EJPalAQIeL2b"}},"sigs":["pOREi-jwXzVnccK3LB67sQ_ZIVAVV8uuN_EzAUCrAsBdUYdk7tFPwZEkR7tMppp3RHVAPjam5wRVK_mUhP7v7AG0NV3V"],"id":"HY5mXzOlgOcwc-xstUV7EJPalAQGOgsRjOXNdaxNAuza5mb2"}`, string(jsondata))

	modelAcc := cs.MustAccount(ids.Address(mi.ID))
	assert.Equal(input.Threshold, modelAcc.Threshold())
	assert.Equal(input.Keepers, modelAcc.Keepers())

	// transfer token to model account
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeTransfer,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     1,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      owner,
		To:        modelAcc.ID().Ptr(),
		Amount:    new(big.Int).SetUint64(unit.MilliLDC * 1500),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())

	itx, err = NewTx(ltx)
	require.NoError(t, err)
	assert.NoError(itx.Apply(ctx, cs))

	fromGas += ltx.Gas()
	assert.Equal(fromGas*ctx.Price,
		itx.(*TxTransfer).ldc.Balance().Uint64())
	assert.Equal(fromGas*100,
		itx.(*TxTransfer).miner.Balance().Uint64())
	assert.Equal(unit.MilliLDC*1500, modelAcc.BalanceOfAll(ids.NativeToken).Uint64())
	assert.Equal(unit.MilliLDC*500, modelAcc.Balance().Uint64())
	assert.Equal(unit.LDC*3-fromGas*(ctx.Price+100)-unit.MilliLDC*1500,
		ownerAcc.BalanceOfAll(ids.NativeToken).Uint64())
	assert.Equal(unit.LDC*2-fromGas*(ctx.Price+100)-unit.MilliLDC*1500,
		ownerAcc.Balance().Uint64())
	assert.Equal(uint64(2), ownerAcc.Nonce())
	assert.True(ownerAcc.IsEmpty())
	assert.False(modelAcc.IsEmpty())

	// transfer token from model account
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeTransfer,
		ChainID:   ctx.ChainConfig().ChainID,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      modelAcc.ID(),
		To:        ownerAcc.ID().Ptr(),
		Amount:    new(big.Int).SetUint64(unit.MilliLDC * 100),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())

	itx, err = NewTx(ltx)
	require.NoError(t, err)
	assert.NoError(itx.Apply(ctx, cs))

	modelGas := ltx.Gas()
	assert.Equal((fromGas+modelGas)*ctx.Price,
		itx.(*TxTransfer).ldc.Balance().Uint64())
	assert.Equal((fromGas+modelGas)*100,
		itx.(*TxTransfer).miner.Balance().Uint64())
	assert.Equal(unit.MilliLDC*400-modelGas*(ctx.Price+100), modelAcc.Balance().Uint64())
	assert.Equal(unit.LDC-fromGas*(ctx.Price+100)-unit.MilliLDC*400,
		ownerAcc.Balance().Uint64())
	assert.Equal(uint64(1), modelAcc.Nonce())

	assert.NoError(cs.VerifyState())
}

func TestTxCreateModelGenesis(t *testing.T) {
	assert := assert.New(t)

	ctx := NewMockChainContext()
	cs := ctx.MockChainState()
	owner := signer.Signer1.Key().Address()

	nm, err := service.NameModel()
	require.NoError(t, err)
	mi := &ld.ModelInfo{
		Name:      nm.Name(),
		Threshold: 1,
		Keepers:   signer.Keys{signer.Signer1.Key()},
		Schema:    nm.Schema(),
	}
	assert.NoError(mi.SyntacticVerify())

	ltx := &ld.Transaction{Tx: ld.TxData{
		Type:    ld.TypeCreateModel,
		ChainID: ctx.ChainConfig().ChainID,
		From:    owner,
		Data:    ld.MustMarshal(mi),
	}}
	assert.NoError(ltx.SyntacticVerify())

	itx, err := NewGenesisTx(ltx)
	require.NoError(t, err)
	assert.NoError(itx.(GenesisTx).ApplyGenesis(ctx, cs))

	assert.Equal(uint64(0), itx.(*TxCreateModel).ldc.Balance().Uint64())
	assert.Equal(uint64(0), itx.(*TxCreateModel).miner.Balance().Uint64())
	assert.Equal(uint64(0), itx.(*TxCreateModel).from.Balance().Uint64())
	assert.Equal(uint64(1), itx.(*TxCreateModel).from.Nonce())

	mi2, err := cs.LoadModel(itx.(*TxCreateModel).input.ID)
	require.NoError(t, err)
	assert.Equal(uint16(1), mi2.Threshold)
	assert.Equal(signer.Keys{signer.Signer1.Key()}, mi2.Keepers)
	assert.Nil(mi2.Approver)
	assert.Equal(mi.Name, mi2.Name)
	assert.Equal(mi.Schema, mi2.Schema)

	jsondata, err := itx.MarshalJSON()
	require.NoError(t, err)
	// fmt.Println(string(jsondata))
	assert.Equal(`{"tx":{"type":"TypeCreateModel","chainID":2357,"nonce":0,"gasTip":0,"gasFeeCap":0,"from":"0x8db97c7cECe249C2b98bdc0226cc4C2A57bF52fc","data":{"name":"NameService","threshold":1,"keepers":["jbl8fOziScK5i9wCJsxMKle_UvwKxwPH"],"schema":"type ID20 bytes\n\ttype NameService struct {\n\t\tname       String        (rename \"n\")\n\t\tlinked     optional ID20 (rename \"l\")\n\t\trecords    [String]      (rename \"rs\")\n\t\textensions [Any]         (rename \"es\")\n\t}","id":"keQIKtE491kODEshAG0tI9EVd1So6OM7"}},"id":"keQIKtE491kODEshAG0tI9EVd1RD71cR4h73M0xgUOm0D9tM"}`, string(jsondata))

	pm, err := service.ProfileModel()
	require.NoError(t, err)
	mi = &ld.ModelInfo{
		Name:      pm.Name(),
		Threshold: 1,
		Keepers:   signer.Keys{signer.Signer1.Key()},
		Schema:    pm.Schema(),
	}
	assert.NoError(mi.SyntacticVerify())

	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:    ld.TypeCreateModel,
		ChainID: ctx.ChainConfig().ChainID,
		Nonce:   1,
		From:    owner,
		Data:    ld.MustMarshal(mi),
	}}
	assert.NoError(ltx.SyntacticVerify())

	itx, err = NewGenesisTx(ltx)
	require.NoError(t, err)
	assert.NoError(itx.(GenesisTx).ApplyGenesis(ctx, cs))

	assert.Equal(uint64(0), itx.(*TxCreateModel).ldc.Balance().Uint64())
	assert.Equal(uint64(0), itx.(*TxCreateModel).miner.Balance().Uint64())
	assert.Equal(uint64(0), itx.(*TxCreateModel).from.Balance().Uint64())
	assert.Equal(uint64(2), itx.(*TxCreateModel).from.Nonce())

	mi2, err = cs.LoadModel(itx.(*TxCreateModel).input.ID)
	require.NoError(t, err)
	assert.Equal(uint16(1), mi2.Threshold)
	assert.Equal(signer.Keys{signer.Signer1.Key()}, mi2.Keepers)
	assert.Nil(mi2.Approver)
	assert.Equal(mi.Name, mi2.Name)
	assert.Equal(mi.Schema, mi2.Schema)

	jsondata, err = itx.MarshalJSON()
	require.NoError(t, err)
	// fmt.Println(string(jsondata))
	assert.Equal(`{"tx":{"type":"TypeCreateModel","chainID":2357,"nonce":1,"gasTip":0,"gasFeeCap":0,"from":"0x8db97c7cECe249C2b98bdc0226cc4C2A57bF52fc","data":{"name":"ProfileService","threshold":1,"keepers":["jbl8fOziScK5i9wCJsxMKle_UvwKxwPH"],"schema":"type ID20 bytes\n\ttype ProfileService struct {\n\t\ttype        Int             (rename \"t\")\n\t\tname        String          (rename \"n\")\n\t\tdescription String          (rename \"d\")\n\t\timage       String          (rename \"i\")\n\t\turl         String          (rename \"u\")\n\t\tfollows     [ID20]          (rename \"fs\")\n\t\tmembers     optional [ID20] (rename \"ms\")\n\t\textensions  [Any]           (rename \"es\")\n\t}","id":"mO2W1IlEMsieYIWUAwnK6OHgSm9J3pLH"}},"id":"mO2W1IlEMsieYIWUAwnK6OHgSm-A_0zgrD-KKWK_bpIj8V-F"}`, string(jsondata))

	assert.NoError(cs.VerifyState())
}
