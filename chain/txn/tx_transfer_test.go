// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package txn

import (
	"math/big"
	"testing"

	"github.com/ldclabs/ldvm/ids"
	"github.com/ldclabs/ldvm/ld"
	"github.com/ldclabs/ldvm/signer"
	"github.com/ldclabs/ldvm/unit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTxTransfer(t *testing.T) {
	assert := assert.New(t)

	// SyntacticVerify
	tx := &TxTransfer{}
	assert.ErrorContains(tx.SyntacticVerify(), "nil pointer")
	_, err := tx.MarshalJSON()
	require.NoError(t, err)

	ctx := NewMockChainContext()
	cs := ctx.MockChainState()

	from := cs.MustAccount(signer.Signer1.Key().Address())
	from.LD().Nonce = 1

	ltx := &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeTransfer,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     1,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      from.ID(),
	}}
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "no signatures")

	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "invalid to")

	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeTransfer,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     1,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      from.ID(),
		To:        ids.GenesisAccount.Ptr(),
	}}

	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "invalid amount")

	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeTransfer,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     1,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      from.ID(),
		To:        ids.GenesisAccount.Ptr(),
		Amount:    new(big.Int).SetUint64(1000),
	}}

	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	itx, err := NewTx(ltx)
	require.NoError(t, err)
	cs.CommitAccounts()
	assert.ErrorContains(itx.Apply(ctx, cs), "insufficient NativeLDC balance")
	cs.CheckoutAccounts()

	from.Add(ids.NativeToken, new(big.Int).SetUint64(unit.LDC*2))
	assert.NoError(itx.Apply(ctx, cs))

	fromGas := ltx.Gas()
	assert.Equal(fromGas*ctx.Price,
		itx.(*TxTransfer).ldc.Balance().Uint64())
	assert.Equal(fromGas*100,
		itx.(*TxTransfer).miner.Balance().Uint64())
	assert.Equal(uint64(1000),
		itx.(*TxTransfer).to.Balance().Uint64())
	assert.Equal(unit.LDC-fromGas*(ctx.Price+100)-1000,
		from.Balance().Uint64())
	assert.Equal(uint64(2), from.Nonce())

	jsondata, err := itx.MarshalJSON()
	require.NoError(t, err)
	// fmt.Println(string(jsondata))
	assert.Equal(`{"tx":{"type":"TypeTransfer","chainID":2357,"nonce":1,"gasTip":100,"gasFeeCap":1000,"from":"0x8db97c7cECe249C2b98bdc0226cc4C2A57bF52fc","to":"0xFFfFFFfFfffFFfFFffFFFfFfFffFFFfffFfFFFff","amount":1000},"sigs":["IX83ghjdiu09Zg4-ZjXIMAlZItoyOJ9ZxTSeAX63gV549EM4gtDf_fMeefUWzH4pT6YKYchkhL6a9pYdVRZCegGBPL0O"],"id":"3-5PqLJQpKdAtM7MDHuIuH5gopri-5McUcBu9b-stOW9Uimp"}`, string(jsondata))

	token := ld.MustNewToken("$LDC")
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeTransfer,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     2,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      from.ID(),
		To:        ids.GenesisAccount.Ptr(),
		Token:     token.Ptr(),
		Amount:    new(big.Int).SetUint64(1000),
	}}

	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	itx, err = NewTx(ltx)
	require.NoError(t, err)
	cs.CommitAccounts()
	assert.ErrorContains(itx.Apply(ctx, cs), "insufficient $LDC balance")
	cs.CheckoutAccounts()

	from.Add(token, new(big.Int).SetUint64(unit.LDC))
	assert.NoError(itx.Apply(ctx, cs))

	fromGas += ltx.Gas()
	assert.Equal(fromGas*ctx.Price,
		itx.(*TxTransfer).ldc.Balance().Uint64())
	assert.Equal(fromGas*100,
		itx.(*TxTransfer).miner.Balance().Uint64())
	assert.Equal(uint64(1000), itx.(*TxTransfer).to.BalanceOf(token).Uint64())
	assert.Equal(unit.LDC-1000, from.BalanceOf(token).Uint64())
	assert.Equal(unit.LDC-fromGas*(ctx.Price+100)-1000,
		from.Balance().Uint64())
	assert.Equal(uint64(3), from.Nonce())

	jsondata, err = itx.MarshalJSON()
	require.NoError(t, err)
	// fmt.Println(string(jsondata))
	assert.Equal(`{"tx":{"type":"TypeTransfer","chainID":2357,"nonce":2,"gasTip":100,"gasFeeCap":1000,"from":"0x8db97c7cECe249C2b98bdc0226cc4C2A57bF52fc","to":"0xFFfFFFfFfffFFfFFffFFFfFfFffFFFfffFfFFFff","token":"$LDC","amount":1000},"sigs":["uGG3X1KnhErX6M4bba6hRK5p8LQv3JypqXNQ1ypaUNN2-JSGCOkV9zQ4YLdSIJqOcfLe--EnUT5pKLNincmqIgBsXXMx"],"id":"i4GN43OdfBiEOseH1fP8u0sg7LoUQ4TF2yLgIteaOKxhHz1v"}`, string(jsondata))

	// support 0 amount
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeTransfer,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     3,
		GasTip:    0,
		GasFeeCap: ctx.Price,
		From:      from.ID(),
		To:        ids.GenesisAccount.Ptr(),
		Amount:    new(big.Int).SetUint64(0),
		Data:      []byte(`"some message"`),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	itx, err = NewTx(ltx)
	require.NoError(t, err)
	assert.NoError(itx.Apply(ctx, cs), "should support 0 amount")

	assert.NoError(cs.VerifyState())
}

func TestTxTransferGenesis(t *testing.T) {
	assert := assert.New(t)

	ctx := NewMockChainContext()
	cs := ctx.MockChainState()

	from := cs.MustAccount(ids.LDCAccount)
	from.Add(ids.NativeToken, ctx.ChainConfig().MaxTotalSupply)

	ltx := &ld.Transaction{Tx: ld.TxData{
		Type:    ld.TypeTransfer,
		ChainID: ctx.ChainConfig().ChainID,
		From:    from.ID(),
		To:      ids.GenesisAccount.Ptr(),
		Amount:  ctx.ChainConfig().MaxTotalSupply,
	}}

	itx, err := NewGenesisTx(ltx)
	require.NoError(t, err)

	assert.NoError(itx.(*TxTransfer).ApplyGenesis(ctx, cs))

	assert.Equal(uint64(0), itx.(*TxTransfer).ldc.Balance().Uint64())
	assert.Equal(uint64(0), itx.(*TxTransfer).miner.Balance().Uint64())
	assert.Equal(ctx.ChainConfig().MaxTotalSupply.Uint64(), itx.(*TxTransfer).to.Balance().Uint64())
	assert.Equal(uint64(0), itx.(*TxTransfer).from.Balance().Uint64())
	assert.Equal(uint64(1), itx.(*TxTransfer).from.Nonce())

	jsondata, err := itx.MarshalJSON()
	require.NoError(t, err)
	// fmt.Println(string(jsondata))
	assert.Equal(`{"tx":{"type":"TypeTransfer","chainID":2357,"nonce":0,"gasTip":0,"gasFeeCap":0,"from":"0x0000000000000000000000000000000000000000","to":"0xFFfFFFfFfffFFfFFffFFFfFfFffFFFfffFfFFFff","amount":1000000000000000000},"id":"UNJMvAMIzdXEqMoJF7eVHAg1l42bkP3ceNXdTR7vs8qbbhhI"}`, string(jsondata))

	assert.NoError(cs.VerifyState())
}

func TestTxTransferFromNoKeeperAccount(t *testing.T) {
	assert := assert.New(t)

	ctx := NewMockChainContext()
	cs := ctx.MockChainState()

	signer1 := signer.NewSigner()
	from := cs.MustAccount(signer1.Key().Address())
	from.Add(ids.NativeToken, new(big.Int).SetUint64(unit.LDC*2))
	to := cs.MustAccount(ids.Address{1, 2, 3})

	ltx := &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeTransfer,
		ChainID:   ctx.ChainConfig().ChainID,
		GasFeeCap: ctx.Price,
		From:      from.ID(),
		To:        to.ID().Ptr(),
		Amount:    new(big.Int).SetUint64(unit.MilliLDC * 500),
	}}

	assert.NoError(ltx.SignWith(signer1))
	assert.NoError(ltx.SyntacticVerify())
	itx, err := NewTx(ltx)
	require.NoError(t, err)
	assert.NoError(itx.Apply(ctx, cs))

	fromGas := ltx.Gas()
	assert.Equal(fromGas*ctx.Price, itx.(*TxTransfer).ldc.Balance().Uint64())
	assert.Equal(fromGas*0, itx.(*TxTransfer).miner.Balance().Uint64())
	assert.Equal(uint64(0), to.Balance().Uint64())
	assert.Equal(unit.MilliLDC*500, to.BalanceOfAll(ids.NativeToken).Uint64())
	assert.Equal(unit.LDC-fromGas*(ctx.Price+0)-unit.MilliLDC*500,
		from.Balance().Uint64())
	assert.Equal(uint64(1), from.Nonce())
	assert.True(from.IsEmpty())
	assert.True(to.IsEmpty())

	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeTransfer,
		ChainID:   ctx.ChainConfig().ChainID,
		GasFeeCap: ctx.Price,
		From:      to.ID(),
		To:        from.ID().Ptr(),
		Amount:    new(big.Int).SetUint64(unit.MilliLDC * 100),
	}}

	assert.NoError(ltx.SignWith(signer1))
	assert.NoError(ltx.SyntacticVerify())
	itx, err = NewTx(ltx)
	require.NoError(t, err)
	assert.ErrorContains(itx.Apply(ctx, cs), "TxBase.Apply: invalid signatures for sender")

	assert.NoError(cs.VerifyState())
}
