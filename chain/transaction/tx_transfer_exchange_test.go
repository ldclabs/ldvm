// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package transaction

import (
	"math/big"
	"testing"

	"github.com/ldclabs/ldvm/constants"
	"github.com/ldclabs/ldvm/ld"
	"github.com/ldclabs/ldvm/util"
	"github.com/stretchr/testify/assert"
)

func TestTxTransferExchange(t *testing.T) {
	assert := assert.New(t)

	// SyntacticVerify
	tx := &TxTransferExchange{}
	assert.ErrorContains(tx.SyntacticVerify(), "nil pointer")
	_, err := tx.MarshalJSON()
	assert.NoError(err)

	token := ld.MustNewToken("$LDC")
	bctx := NewMockBCtx()
	bs := NewMockBS(bctx)

	from, err := bs.LoadAccount(util.Signer1.Address())
	assert.NoError(err)
	from.ld.Nonce = 1
	to, err := bs.LoadAccount(util.Signer2.Address())
	assert.NoError(err)

	txData := &ld.TxData{
		Type:      ld.TypeExchange,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     1,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      from.id,
	}
	assert.NoError(txData.SyntacticVerify())
	_, err = NewTx(txData.ToTransaction(), true)
	assert.ErrorContains(err, "DeriveSigners: no signature")

	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx(txData.ToTransaction(), true)
	assert.ErrorContains(err, "invalid to")

	txData = &ld.TxData{
		Type:      ld.TypeExchange,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     1,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      from.id,
		To:        &to.id,
	}
	assert.NoError(txData.SyntacticVerify())
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx(txData.ToTransaction(), true)
	assert.ErrorContains(err, "invalid amount")

	txData = &ld.TxData{
		Type:      ld.TypeExchange,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     1,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      from.id,
		To:        &to.id,
		Amount:    new(big.Int).SetUint64(constants.LDC),
	}
	assert.NoError(txData.SyntacticVerify())
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx(txData.ToTransaction(), true)
	assert.ErrorContains(err, "invalid data")

	txData = &ld.TxData{
		Type:      ld.TypeExchange,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     1,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      from.id,
		To:        &to.id,
		Amount:    new(big.Int).SetUint64(constants.LDC),
		Data:      []byte("abc"),
	}
	assert.NoError(txData.SyntacticVerify())
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx(txData.ToTransaction(), true)
	assert.ErrorContains(err, "cbor: cannot unmarshal")

	input := ld.TxExchanger{
		Nonce:   1,
		Sell:    token,
		Receive: constants.NativeToken,
		Quota:   new(big.Int).SetUint64(constants.LDC * 1000),
		Minimum: new(big.Int).SetUint64(constants.LDC),
		Price:   new(big.Int).SetUint64(1_000_000),
		Expire:  1,
		Payee:   to.id,
	}
	assert.NoError(input.SyntacticVerify())
	txData = &ld.TxData{
		Type:      ld.TypeExchange,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     1,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      from.id,
		To:        &to.id,
		Amount:    new(big.Int).SetUint64(1_000_000 - 1),
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SyntacticVerify())
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx(txData.ToTransaction(), true)
	assert.ErrorContains(err, "invalid amount, expected >=1000000, got 999999")

	txData = &ld.TxData{
		Type:      ld.TypeExchange,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     1,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      from.id,
		To:        &to.id,
		Amount:    new(big.Int).SetUint64(1_000_000*1000 + 1),
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SyntacticVerify())
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx(txData.ToTransaction(), true)
	assert.ErrorContains(err, "invalid amount, expected <=1000000000, got 1000000001")

	input = ld.TxExchanger{
		Nonce:     1,
		Sell:      token,
		Receive:   constants.NativeToken,
		Quota:     new(big.Int).SetUint64(constants.LDC * 1000),
		Minimum:   new(big.Int).SetUint64(constants.LDC),
		Price:     new(big.Int).SetUint64(1_000_000),
		Expire:    1,
		Payee:     to.id,
		Purchaser: &constants.GenesisAccount,
	}
	assert.NoError(input.SyntacticVerify())
	txData = &ld.TxData{
		Type:      ld.TypeExchange,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     1,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      from.id,
		To:        &to.id,
		Amount:    new(big.Int).SetUint64(1_000_000),
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SyntacticVerify())
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx(txData.ToTransaction(), true)
	assert.ErrorContains(err,
		"invalid from, expected 0xFFfFfFffFFfffFFfFFfFFFFFffFFFffffFfFFFfF, got 0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC")

	input = ld.TxExchanger{
		Nonce:   1,
		Sell:    token,
		Receive: constants.NativeToken,
		Quota:   new(big.Int).SetUint64(constants.LDC * 1000),
		Minimum: new(big.Int).SetUint64(constants.LDC),
		Price:   new(big.Int).SetUint64(1_000_000),
		Expire:  1,
		Payee:   to.id,
	}
	assert.NoError(input.SyntacticVerify())
	txData = &ld.TxData{
		Type:      ld.TypeExchange,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     1,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      from.id,
		To:        &constants.GenesisAccount,
		Amount:    new(big.Int).SetUint64(1_000_000),
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SyntacticVerify())
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx(txData.ToTransaction(), true)
	assert.ErrorContains(err,
		"invalid to, expected 0x44171C37Ff5D7B7bb8dcad5C81f16284A229e641, got 0xFFfFfFffFFfffFFfFFfFFFFFffFFFffffFfFFFfF")

	input = ld.TxExchanger{
		Nonce:   1,
		Sell:    token,
		Receive: constants.NativeToken,
		Quota:   new(big.Int).SetUint64(constants.LDC * 1000),
		Minimum: new(big.Int).SetUint64(constants.LDC),
		Price:   new(big.Int).SetUint64(1_000_000),
		Expire:  1,
		Payee:   to.id,
	}
	assert.NoError(input.SyntacticVerify())
	txData = &ld.TxData{
		Type:      ld.TypeExchange,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     1,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      from.id,
		To:        &to.id,
		Token:     &token,
		Amount:    new(big.Int).SetUint64(1_000_000),
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SyntacticVerify())
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx(txData.ToTransaction(), true)
	assert.ErrorContains(err,
		"invalid token, expected NativeLDC, got $LDC")

	input = ld.TxExchanger{
		Nonce:   1,
		Sell:    token,
		Receive: constants.NativeToken,
		Quota:   new(big.Int).SetUint64(constants.LDC * 1000),
		Minimum: new(big.Int).SetUint64(constants.LDC),
		Price:   new(big.Int).SetUint64(1_000_000),
		Expire:  1,
		Payee:   to.id,
	}
	assert.NoError(input.SyntacticVerify())
	txData = &ld.TxData{
		Type:      ld.TypeExchange,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     1,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      from.id,
		To:        &to.id,
		Amount:    new(big.Int).SetUint64(1_000_000),
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SyntacticVerify())
	assert.NoError(txData.SignWith(util.Signer1))
	tt := txData.ToTransaction()
	tt.Timestamp = 10
	_, err = NewTx(tt, true)
	assert.ErrorContains(err, "data expired")
	tt.Timestamp = 1
	_, err = NewTx(tt, true)
	assert.ErrorContains(err, "DeriveSigners: no signature")

	txData = &ld.TxData{
		Type:      ld.TypeExchange,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     1,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      from.id,
		To:        &to.id,
		Amount:    new(big.Int).SetUint64(1_000_000),
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SyntacticVerify())
	assert.NoError(txData.SignWith(util.Signer1))
	assert.NoError(txData.ExSignWith(util.Signer1))
	tt = txData.ToTransaction()
	itx, err := NewTx(tt, true)
	assert.NoError(err)
	assert.ErrorContains(itx.Verify(bctx, bs), "invalid gas, expected 227, got 0")

	txData = &ld.TxData{
		Type:      ld.TypeExchange,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     1,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      from.id,
		To:        &to.id,
		Amount:    new(big.Int).SetUint64(1_000_000),
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SyntacticVerify())
	assert.NoError(txData.SignWith(util.Signer1))
	assert.NoError(txData.ExSignWith(util.Signer1))
	tt = txData.ToTransaction()
	tt.Gas = tt.RequiredGas(bctx.FeeConfig().ThresholdGas)
	itx, err = NewTx(tt, true)
	assert.NoError(err)
	assert.ErrorContains(itx.Verify(bctx, bs), "insufficient NativeLDC balance, expected 1249700, got 0")
	from.Add(constants.NativeToken, new(big.Int).SetUint64(constants.LDC))
	assert.ErrorContains(itx.Verify(bctx, bs), "invalid signatures for seller")

	txData = &ld.TxData{
		Type:      ld.TypeExchange,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     1,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      from.id,
		To:        &to.id,
		Amount:    new(big.Int).SetUint64(1_000_000),
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SyntacticVerify())
	assert.NoError(txData.SignWith(util.Signer1))
	assert.NoError(txData.ExSignWith(util.Signer2))
	tt = txData.ToTransaction()
	tt.Gas = tt.RequiredGas(bctx.FeeConfig().ThresholdGas)
	itx, err = NewTx(tt, true)
	assert.NoError(err)
	assert.ErrorContains(itx.Verify(bctx, bs),
		"nonce 1 not exists at 1 on 0x44171C37Ff5D7B7bb8dcad5C81f16284A229e641")
	assert.NoError(to.AddNonceTable(1, []uint64{1, 2, 3}))
	assert.ErrorContains(itx.Verify(bctx, bs),
		"insufficient $LDC balance, expected 1000000000, got 0")
	to.Add(token, new(big.Int).SetUint64(constants.LDC))
	assert.NoError(itx.Verify(bctx, bs))
	assert.NoError(itx.Accept(bctx, bs))

	tx = itx.(*TxTransferExchange)
	assert.Equal(tx.ld.Gas*bctx.Price, tx.ldc.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(tx.ld.Gas*100, tx.miner.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(uint64(1_000_000), to.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(uint64(0), to.balanceOf(token).Uint64())
	assert.Equal(constants.LDC, from.balanceOf(token).Uint64())
	assert.Equal(constants.LDC-tx.ld.Gas*(bctx.Price+100)-1_000_000,
		from.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(uint64(2), tx.from.Nonce())
	assert.Equal([]uint64{2, 3}, to.ld.NonceTable[1])

	jsondata, err := itx.MarshalJSON()
	assert.NoError(err)
	// fmt.Println(string(jsondata))
	assert.Equal(`{"type":"TypeExchange","chainID":2357,"nonce":1,"gasTip":100,"gasFeeCap":1000,"from":"0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC","to":"0x44171C37Ff5D7B7bb8dcad5C81f16284A229e641","amount":1000000,"data":{"nonce":1,"sell":"$LDC","receive":"","quota":1000000000000,"minimum":1000000000,"price":1000000,"expire":1,"payee":"0x44171C37Ff5D7B7bb8dcad5C81f16284A229e641"},"signatures":["c1ad16b7420ad47b8a4f0f506840c57df9b42510a43faee9d9a150fb5a7c00155f5c5fb952693b7af3075ce69d98e14486d10bf6c0a886243c64563004b019ab01"],"exSignatures":["b29b8525280dde2e056b99b20a1e252f2a5b44320837b2b7bc0280c6fa20aa7c743cd3959b97b11c757dbdabf177bbc22ec5263ee529cddceda3a5415b89850000"],"gas":227,"id":"qEstDMoKQN1Z81gwGGY29LdPwVhZRThZq8enUbw35tTgqJiuK"}`, string(jsondata))

	assert.NoError(bs.VerifyState())
}
