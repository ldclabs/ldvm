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

func TestTxTransferCash(t *testing.T) {
	assert := assert.New(t)

	// SyntacticVerify
	tx := &TxTransferCash{}
	assert.ErrorContains(tx.SyntacticVerify(), "nil pointer")
	_, err := tx.MarshalJSON()
	require.NoError(t, err)

	token := ld.MustNewToken("$LDC")
	ctx := NewMockChainContext()
	cs := ctx.MockChainState()

	from := cs.MustAccount(signer.Signer1.Key().Address())
	from.LD().Nonce = 2
	to := cs.MustAccount(signer.Signer2.Key().Address())

	ltx := &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeTransferCash,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     2,
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
		Type:      ld.TypeTransferCash,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     2,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      from.ID(),
		To:        to.ID().Ptr(),
		Amount:    new(big.Int).SetUint64(1),
	}}

	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "invalid amount, should be nil")

	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeTransferCash,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     2,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      from.ID(),
		To:        to.ID().Ptr(),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "invalid data")

	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeTransferCash,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     2,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      from.ID(),
		To:        to.ID().Ptr(),
		Data:      []byte("ABC"),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "no exSignatures")

	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeTransferCash,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     2,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      from.ID(),
		To:        to.ID().Ptr(),
		Data:      []byte("ABC"),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.ExSignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "extraneous data")

	input := ld.TxTransfer{
		To:     from.ID().Ptr(),
		Token:  token.Ptr(),
		Amount: new(big.Int).SetUint64(unit.LDC),
	}
	assert.NoError(input.SyntacticVerify())
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeTransferCash,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     2,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      from.ID(),
		To:        to.ID().Ptr(),
		Data:      input.Bytes(),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.ExSignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "nil issuer")

	input = ld.TxTransfer{
		From:   ids.GenesisAccount.Ptr(),
		To:     from.ID().Ptr(),
		Token:  token.Ptr(),
		Amount: new(big.Int).SetUint64(unit.LDC),
	}
	assert.NoError(input.SyntacticVerify())
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeTransferCash,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     2,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      from.ID(),
		To:        to.ID().Ptr(),
		Data:      input.Bytes(),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.ExSignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err,
		"invalid issuer, expected 0xFFfFFFfFfffFFfFFffFFFfFfFffFFFfffFfFFFff, got 0x44171C37Ff5D7B7bb8Dcad5C81f16284A229E641")

	input = ld.TxTransfer{
		From:   to.ID().Ptr(),
		Token:  token.Ptr(),
		Amount: new(big.Int).SetUint64(unit.LDC),
	}
	assert.NoError(input.SyntacticVerify())
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeTransferCash,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     2,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      from.ID(),
		To:        to.ID().Ptr(),
		Data:      input.Bytes(),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.ExSignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "nil recipient")

	input = ld.TxTransfer{
		From:   to.ID().Ptr(),
		To:     ids.GenesisAccount.Ptr(),
		Token:  token.Ptr(),
		Amount: new(big.Int).SetUint64(unit.LDC),
	}
	assert.NoError(input.SyntacticVerify())
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeTransferCash,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     2,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      from.ID(),
		To:        to.ID().Ptr(),
		Data:      input.Bytes(),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.ExSignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err,
		"invalid recipient, expected 0xFFfFFFfFfffFFfFFffFFFfFfFffFFFfffFfFFFff, got 0x8db97c7cECe249C2b98bdc0226cc4C2A57bF52fc")

	input = ld.TxTransfer{
		From:   to.ID().Ptr(),
		To:     from.ID().Ptr(),
		Amount: new(big.Int).SetUint64(unit.LDC),
	}
	assert.NoError(input.SyntacticVerify())
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeTransferCash,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     2,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      from.ID(),
		To:        to.ID().Ptr(),
		Token:     token.Ptr(),
		Data:      input.Bytes(),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.ExSignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "invalid token, expected NativeLDC, got $LDC")

	input = ld.TxTransfer{
		From:   to.ID().Ptr(),
		To:     from.ID().Ptr(),
		Token:  token.Ptr(),
		Amount: new(big.Int).SetUint64(unit.LDC),
	}
	assert.NoError(input.SyntacticVerify())
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeTransferCash,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     2,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      from.ID(),
		To:        to.ID().Ptr(),
		Data:      input.Bytes(),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.ExSignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "invalid token, expected $LDC, got NativeLDC")

	input = ld.TxTransfer{
		From: to.ID().Ptr(),
		To:   from.ID().Ptr(),
	}
	assert.NoError(input.SyntacticVerify())
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeTransferCash,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     2,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      from.ID(),
		To:        to.ID().Ptr(),
		Data:      input.Bytes(),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.ExSignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "invalid amount, expected >= 1")

	input = ld.TxTransfer{
		From:   to.ID().Ptr(),
		To:     from.ID().Ptr(),
		Amount: new(big.Int).SetUint64(unit.LDC),
	}
	assert.NoError(input.SyntacticVerify())
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeTransferCash,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     2,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      from.ID(),
		To:        to.ID().Ptr(),
		Data:      input.Bytes(),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.ExSignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	ltx.Timestamp = 10
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "data expired")

	input = ld.TxTransfer{
		From:   to.ID().Ptr(),
		To:     from.ID().Ptr(),
		Amount: new(big.Int).SetUint64(unit.LDC),
		Expire: cs.Timestamp() + 1,
	}
	assert.NoError(input.SyntacticVerify())
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeTransferCash,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     2,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      from.ID(),
		To:        to.ID().Ptr(),
		Data:      input.Bytes(),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.ExSignWith(signer.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	ltx.Timestamp = 10
	itx, err := NewTx(ltx)
	require.NoError(t, err)
	cs.CommitAccounts()
	assert.ErrorContains(itx.Apply(ctx, cs),
		"insufficient NativeLDC balance, expected 1491600, got 0")
	cs.CheckoutAccounts()
	from.Add(ids.NativeToken, new(big.Int).SetUint64(unit.LDC*2))
	cs.CommitAccounts()
	assert.ErrorContains(itx.Apply(ctx, cs),
		"invalid signature for issuer")
	cs.CheckoutAccounts()

	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeTransferCash,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     2,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      from.ID(),
		To:        to.ID().Ptr(),
		Data:      input.Bytes(),
	}}
	assert.NoError(ltx.SignWith(signer.Signer1))
	assert.NoError(ltx.ExSignWith(signer.Signer2))
	assert.NoError(ltx.SyntacticVerify())
	ltx.Timestamp = cs.Timestamp()
	itx, err = NewTx(ltx)
	require.NoError(t, err)

	cs.CommitAccounts()
	assert.ErrorContains(itx.Apply(ctx, cs),
		"nonce 0 not exists at 1001")
	cs.CheckoutAccounts()
	assert.NoError(to.UpdateNonceTable(cs.Timestamp()+1, []uint64{2, 1, 0}))
	cs.CommitAccounts()
	assert.ErrorContains(itx.Apply(ctx, cs),
		"insufficient transferable NativeLDC balance, expected 1000000000, got 0")
	cs.CheckoutAccounts()
	to.Add(ids.NativeToken, new(big.Int).SetUint64(unit.LDC*2))
	assert.NoError(itx.Apply(ctx, cs))

	tx = itx.(*TxTransferCash)
	assert.Equal(ltx.Gas()*ctx.Price, tx.ldc.Balance().Uint64())
	assert.Equal(ltx.Gas()*100, tx.miner.Balance().Uint64())
	assert.Equal(uint64(0), to.Balance().Uint64())
	assert.Equal(unit.LDC*2-ltx.Gas()*(ctx.Price+100),
		from.Balance().Uint64())
	assert.Equal(uint64(3), tx.from.Nonce())
	assert.Equal([]uint64{1, 2}, to.LD().NonceTable[cs.Timestamp()+1])

	jsondata, err := itx.MarshalJSON()
	require.NoError(t, err)
	// fmt.Println(string(jsondata))
	assert.Equal(`{"tx":{"type":"TypeTransferCash","chainID":2357,"nonce":2,"gasTip":100,"gasFeeCap":1000,"from":"0x8db97c7cECe249C2b98bdc0226cc4C2A57bF52fc","to":"0x44171C37Ff5D7B7bb8Dcad5C81f16284A229E641","data":{"from":"0x44171C37Ff5D7B7bb8Dcad5C81f16284A229E641","to":"0x8db97c7cECe249C2b98bdc0226cc4C2A57bF52fc","amount":1000000000,"expire":1001}},"sigs":["PikI2rs3hdiy_2lnFoStuaqqyN0-ulnHV8_t1FZC57dqSsI5GgS1Po0gfQqCOkvsCud2MoyVs9m4TbI-IuFHOgG2E2qo"],"exSigs":["Jt_wdqDpYzI5Ib7815vM-olw6isgg5UeMhmSz2in4VA5cRCNdNbd0pqD-XYAK0gDzwvhRxLLj6b-Kf0LGwOKSwBOmpIz"],"id":"um4FGzP6ffE4XZZAnvUcoD6C76kxqTaOnsRST1o0_Ljon6D2"}`, string(jsondata))

	assert.NoError(cs.VerifyState())
}
