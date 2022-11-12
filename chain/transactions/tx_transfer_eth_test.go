// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package transactions

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ldclabs/ldvm/constants"
	"github.com/ldclabs/ldvm/ld"
	"github.com/ldclabs/ldvm/util/signer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTxEth(t *testing.T) {
	assert := assert.New(t)

	// SyntacticVerify
	tx := &TxEth{}
	assert.ErrorContains(tx.SyntacticVerify(), "nil pointer")
	_, err := tx.MarshalJSON()
	require.NoError(t, err)

	ctx := NewMockChainContext()
	cs := ctx.MockChainState()
	token := ld.MustNewToken("$LDC")

	from := cs.MustAccount(signer.Signer1.Key().Address())
	to := cs.MustAccount(signer.Signer2.Key().Address())

	testTo := common.Address(to.ID())

	txe, err := ld.NewEthTx(&types.AccessListTx{
		ChainID:  new(big.Int).SetUint64(ctx.ChainConfig().ChainID),
		Nonce:    0,
		To:       &testTo,
		Value:    ld.ToEthBalance(big.NewInt(1_000_000)),
		Gas:      0,
		GasPrice: ld.ToEthBalance(new(big.Int).SetUint64(ctx.Price)),
	})
	require.NoError(t, err)
	ltx := txe.ToTransaction()
	assert.NoError(ltx.SyntacticVerify())
	itx, err := NewTx(ltx)
	require.NoError(t, err)
	tx = itx.(*TxEth)

	tx.ld.Tx.To = nil
	assert.ErrorContains(itx.SyntacticVerify(), "invalid to")
	tx.ld.Tx.To = to.ID().Ptr()
	tx.ld.Tx.Amount = nil
	assert.ErrorContains(itx.SyntacticVerify(), "invalid amount")
	tx.ld.Tx.Amount = big.NewInt(1_000_000)
	data := tx.ld.Tx.Data
	tx.ld.Tx.Data = []byte{}
	assert.ErrorContains(itx.SyntacticVerify(), "invalid data")
	tx.ld.Tx.Data = data
	tx.ld.Tx.ChainID = 1
	assert.ErrorContains(itx.SyntacticVerify(), "invalid chainID")
	tx.ld.Tx.ChainID = ctx.ChainConfig().ChainID
	tx.ld.Tx.Nonce = 1
	assert.ErrorContains(itx.SyntacticVerify(), "invalid nonce")
	tx.ld.Tx.Nonce = 0
	tx.ld.Tx.GasTip = 1
	assert.ErrorContains(itx.SyntacticVerify(), "invalid gasTip")
	tx.ld.Tx.GasTip = 0
	tx.ld.Tx.GasFeeCap = ctx.Price - 1
	assert.ErrorContains(itx.SyntacticVerify(), "invalid gasFeeCap")
	tx.ld.Tx.GasFeeCap = ctx.Price
	tx.ld.Tx.From = constants.GenesisAccount
	assert.ErrorContains(itx.SyntacticVerify(), "invalid from")
	tx.ld.Tx.From = from.ID()
	tx.ld.Tx.To = constants.GenesisAccount.Ptr()
	assert.ErrorContains(itx.SyntacticVerify(), "invalid to")
	tx.ld.Tx.To = to.ID().Ptr()
	tx.ld.Tx.Token = &token
	assert.ErrorContains(itx.SyntacticVerify(), "invalid token")
	tx.ld.Tx.Token = nil
	tx.ld.Tx.Amount = big.NewInt(1_000_000 - 1)
	assert.ErrorContains(itx.SyntacticVerify(), "invalid amount")
	tx.ld.Tx.Amount = big.NewInt(1_000_000)
	sigs := tx.ld.Signatures
	tx.ld.Signatures = nil
	assert.ErrorContains(itx.SyntacticVerify(), "no signatures")
	tx.ld.Signatures = signer.Sigs{{}}
	assert.ErrorContains(itx.SyntacticVerify(), "invalid signatures")
	tx.ld.Signatures = sigs
	tx.ld.ExSignatures = signer.Sigs{}
	assert.ErrorContains(itx.SyntacticVerify(), "invalid exSignatures")
	tx.ld.ExSignatures = nil
	assert.NoError(itx.SyntacticVerify())
	itx, err = NewTx(ltx)
	require.NoError(t, err)

	cs.CommitAccounts()
	assert.ErrorContains(itx.Apply(ctx, cs),
		"insufficient NativeLDC balance, expected 2268000, got 0")
	cs.CheckoutAccounts()

	from.Add(constants.NativeToken, new(big.Int).SetUint64(constants.LDC*2))
	assert.NoError(itx.Apply(ctx, cs))

	assert.Equal(ltx.Gas()*ctx.Price,
		itx.(*TxEth).ldc.Balance().Uint64())
	assert.Equal(uint64(0),
		itx.(*TxEth).miner.Balance().Uint64())
	assert.Equal(uint64(0), to.Balance().Uint64())
	assert.Equal(uint64(1_000_000), to.BalanceOfAll(constants.NativeToken).Uint64())
	assert.Equal(constants.LDC-ltx.Gas()*(ctx.Price)-1_000_000,
		from.Balance().Uint64())
	assert.Equal(uint64(1), from.Nonce())

	jsondata, err := itx.MarshalJSON()
	require.NoError(t, err)
	// fmt.Println(string(jsondata))
	assert.Equal(`{"tx":{"type":"TypeEth","chainID":2357,"nonce":0,"gasTip":0,"gasFeeCap":1000,"from":"0x8db97c7cECe249C2b98bdc0226cc4C2A57bF52fc","to":"0x44171C37Ff5D7B7bb8Dcad5C81f16284A229E641","amount":1000000,"data":"Afhtggk1gIXo1KUQAICURBccN_9de3u43K1cgfFihKIp5kGHA41-pMaAAIDAgKAdoEgM7s4-fcjMiP6WKyney6v3xkseSstOMxf-iVOg06AboSU-K8y7fsL0NZER8sKbPIxp5u1W_majwQtli6fvxNKn68g"},"sigs":["HaBIDO7OPn3IzIj-lisp3sur98ZLHkrLTjMX_olToNMboSU-K8y7fsL0NZER8sKbPIxp5u1W_majwQtli6fvxAD1gFQV"],"id":"8_IJeoNODJbFU-y0c3MZhjUcTqTL8yUttza6-S8dRQ1S4v3t"}`, string(jsondata))

	assert.NoError(cs.VerifyState())
}
