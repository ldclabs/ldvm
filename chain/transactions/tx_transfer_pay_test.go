// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package transactions

import (
	"math/big"
	"testing"

	"github.com/ldclabs/ldvm/constants"
	"github.com/ldclabs/ldvm/ld"
	"github.com/ldclabs/ldvm/util"
	"github.com/stretchr/testify/assert"
)

func TestTxTransferPay(t *testing.T) {
	assert := assert.New(t)

	// SyntacticVerify
	tx := &TxTransferPay{}
	assert.ErrorContains(tx.SyntacticVerify(), "nil pointer")
	_, err := tx.MarshalJSON()
	assert.NoError(err)

	token := ld.MustNewToken("$LDC")
	ctx := NewMockChainContext()
	cs := ctx.MockChainState()

	from := cs.MustAccount(util.Signer1.Address())
	from.ld.Nonce = 1
	singer2 := util.Signer2.Address()
	to := cs.MustAccount(constants.GenesisAccount)
	assert.NoError(to.UpdateKeepers(ld.Uint16Ptr(1), &util.EthIDs{singer2}, nil, nil))

	ltx := &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeTransferPay,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     1,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      from.id,
	}}

	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "DeriveSigners error: no signature")

	assert.NoError(ltx.SignWith(util.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "invalid to")

	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeTransferPay,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     1,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      from.id,
		To:        &to.id,
	}}

	assert.NoError(ltx.SignWith(util.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "invalid amount")

	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeTransferPay,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     1,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      from.id,
		To:        &to.id,
		Amount:    new(big.Int).SetUint64(constants.LDC),
	}}

	assert.NoError(ltx.SignWith(util.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "invalid data")

	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeTransferPay,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     1,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      from.id,
		To:        &to.id,
		Amount:    new(big.Int).SetUint64(constants.LDC),
		Data:      []byte("0"),
	}}

	assert.NoError(ltx.SignWith(util.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "cbor: cannot unmarshal")

	input := ld.TxTransfer{
		From: &singer2,
	}
	assert.NoError(input.SyntacticVerify())
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeTransferPay,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     1,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      from.id,
		To:        &to.id,
		Amount:    new(big.Int).SetUint64(constants.LDC),
		Data:      input.Bytes(),
	}}

	assert.NoError(ltx.SignWith(util.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err,
		"invalid sender, expected 0x44171C37Ff5D7B7bb8dcad5C81f16284A229e641, got 0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC")

	input = ld.TxTransfer{
		From: &from.id,
	}
	assert.NoError(input.SyntacticVerify())
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeTransferPay,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     1,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      from.id,
		To:        &to.id,
		Amount:    new(big.Int).SetUint64(constants.LDC),
		Data:      input.Bytes(),
	}}

	assert.NoError(ltx.SignWith(util.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "nil recipient")

	input = ld.TxTransfer{
		To: &constants.GenesisAccount,
	}
	assert.NoError(input.SyntacticVerify())
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeTransferPay,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     1,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      from.id,
		To:        &singer2,
		Amount:    new(big.Int).SetUint64(constants.LDC),
		Data:      input.Bytes(),
	}}

	assert.NoError(ltx.SignWith(util.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err,
		"invalid recipient, expected 0xFFfFfFffFFfffFFfFFfFFFFFffFFFffffFfFFFfF, got 0x44171C37Ff5D7B7bb8dcad5C81f16284A229e641")

	input = ld.TxTransfer{
		To: &constants.GenesisAccount,
	}
	assert.NoError(input.SyntacticVerify())
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeTransferPay,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     1,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      from.id,
		To:        &to.id,
		Token:     &token,
		Amount:    new(big.Int).SetUint64(constants.LDC),
		Data:      input.Bytes(),
	}}

	assert.NoError(ltx.SignWith(util.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "invalid token, expected NativeLDC, got $LDC")

	input = ld.TxTransfer{
		To:    &constants.GenesisAccount,
		Token: &token,
	}
	assert.NoError(input.SyntacticVerify())
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeTransferPay,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     1,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      from.id,
		To:        &to.id,
		Amount:    new(big.Int).SetUint64(constants.LDC),
		Data:      input.Bytes(),
	}}

	assert.NoError(ltx.SignWith(util.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "invalid token, expected $LDC, got NativeLDC")

	input = ld.TxTransfer{
		To:    &constants.GenesisAccount,
		Token: &token,
	}
	assert.NoError(input.SyntacticVerify())
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeTransferPay,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     1,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      from.id,
		To:        &to.id,
		Token:     &token,
		Amount:    new(big.Int).SetUint64(constants.LDC),
		Data:      input.Bytes(),
	}}

	assert.NoError(ltx.SignWith(util.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "nil amount")

	input = ld.TxTransfer{
		To:     &constants.GenesisAccount,
		Token:  &token,
		Amount: new(big.Int).SetUint64(constants.LDC),
	}
	assert.NoError(input.SyntacticVerify())
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeTransferPay,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     1,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      from.id,
		To:        &to.id,
		Token:     &token,
		Amount:    new(big.Int).SetUint64(constants.MilliLDC),
		Data:      input.Bytes(),
	}}

	assert.NoError(ltx.SignWith(util.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "invalid amount, expected 1000000000, got 1000000")

	input = ld.TxTransfer{
		To:     &constants.GenesisAccount,
		Token:  &token,
		Amount: new(big.Int).SetUint64(constants.LDC),
		Expire: 10,
	}
	assert.NoError(input.SyntacticVerify())
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeTransferPay,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     1,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      from.id,
		To:        &to.id,
		Token:     &token,
		Amount:    new(big.Int).SetUint64(constants.LDC),
		Data:      input.Bytes(),
	}}

	assert.NoError(ltx.SignWith(util.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	ltx.Timestamp = cs.Timestamp()
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "data expired")

	input = ld.TxTransfer{
		To:     &constants.GenesisAccount,
		Token:  &token,
		Amount: new(big.Int).SetUint64(constants.LDC),
	}
	assert.NoError(input.SyntacticVerify())
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeTransferPay,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     1,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      from.id,
		To:        &to.id,
		Token:     &token,
		Amount:    new(big.Int).SetUint64(constants.LDC),
		Data:      input.Bytes(),
	}}

	assert.NoError(ltx.SignWith(util.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	ltx.Timestamp = cs.Timestamp()
	_, err = NewTx(ltx)
	assert.ErrorContains(err, "DeriveSigners error: no signature")

	input = ld.TxTransfer{
		To:     &constants.GenesisAccount,
		Token:  &token,
		Amount: new(big.Int).SetUint64(constants.LDC),
	}
	assert.NoError(input.SyntacticVerify())
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeTransferPay,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     1,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      from.id,
		To:        &to.id,
		Token:     &token,
		Amount:    new(big.Int).SetUint64(constants.LDC),
		Data:      input.Bytes(),
	}}

	assert.NoError(ltx.SignWith(util.Signer1))
	assert.NoError(ltx.ExSignWith(util.Signer1))
	assert.NoError(ltx.SyntacticVerify())
	itx, err := NewTx(ltx)
	assert.NoError(err)
	cs.CommitAccounts()
	assert.ErrorContains(itx.Apply(ctx, cs),
		"insufficient NativeLDC balance, expected 1668700, got 0")
	cs.CheckoutAccounts()
	from.Add(constants.NativeToken, new(big.Int).SetUint64(constants.LDC))

	cs.CommitAccounts()
	assert.ErrorContains(itx.Apply(ctx, cs),
		"insufficient $LDC balance, expected 1000000000, got 0")
	cs.CheckoutAccounts()
	from.Add(token, new(big.Int).SetUint64(constants.LDC))
	cs.CommitAccounts()
	assert.ErrorContains(itx.Apply(ctx, cs),
		"invalid exSignatures for recipient")
	cs.CheckoutAccounts()

	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeTransferPay,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     1,
		GasTip:    100,
		GasFeeCap: ctx.Price,
		From:      from.id,
		To:        &to.id,
		Token:     &token,
		Amount:    new(big.Int).SetUint64(constants.LDC),
		Data:      input.Bytes(),
	}}

	assert.NoError(ltx.SignWith(util.Signer1))
	assert.NoError(ltx.ExSignWith(util.Signer2))
	assert.NoError(ltx.SyntacticVerify())
	itx, err = NewTx(ltx)
	assert.NoError(err)
	assert.NoError(itx.Apply(ctx, cs))

	assert.Equal(ltx.Gas()*ctx.Price,
		itx.(*TxTransferPay).ldc.Balance().Uint64())
	assert.Equal(ltx.Gas()*100,
		itx.(*TxTransferPay).miner.Balance().Uint64())
	assert.Equal(constants.LDC, to.balanceOf(token).Uint64())
	assert.Equal(uint64(0), from.balanceOf(token).Uint64())
	assert.Equal(constants.LDC-ltx.Gas()*(ctx.Price+100),
		from.Balance().Uint64())
	assert.Equal(uint64(2), from.Nonce())

	jsondata, err := itx.MarshalJSON()
	assert.NoError(err)
	// fmt.Println(string(jsondata))
	assert.Equal(`{"tx":{"type":"TypeTransferPay","chainID":2357,"nonce":1,"gasTip":100,"gasFeeCap":1000,"from":"0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC","to":"0xFFfFfFffFFfffFFfFFfFFFFFffFFFffffFfFFFfF","token":"$LDC","amount":1000000000,"data":{"to":"0xFFfFfFffFFfffFFfFFfFFFFFffFFFffffFfFFFfF","token":"$LDC","amount":1000000000}},"sigs":["a1fdfe1053216c95e4dabbf0f6bf0ba602672d35097ce907d7ac156f3474f6853f99c5f7507039b3b8d4bf9d54b8b24a93ca10208b5808dcb91441035bef249f00"],"exSigs":["c5613bb2ac47e7d5a8be0f58584791158838ef5dbfd11031b41c0576560e9af373175c38018c19d7b1d4f5d5ebabc3d82cc61d6a1e914727141ccb2cdc9a7dfb00"],"id":"2tc9qr8gWAuxefaB5T343zN8N6MCKCCPpv2c1mLCSYKq1rhYaq"}`, string(jsondata))

	// should support 0 amount
	input = ld.TxTransfer{
		From:   &from.id,
		To:     &constants.GenesisAccount,
		Amount: new(big.Int).SetUint64(0),
		Data:   []byte(`"some data`),
	}
	assert.NoError(input.SyntacticVerify())
	ltx = &ld.Transaction{Tx: ld.TxData{
		Type:      ld.TypeTransferPay,
		ChainID:   ctx.ChainConfig().ChainID,
		Nonce:     2,
		GasFeeCap: ctx.Price,
		From:      from.id,
		To:        &to.id,
		Amount:    new(big.Int).SetUint64(0),
		Data:      input.Bytes(),
	}}

	assert.NoError(ltx.SignWith(util.Signer1))
	assert.NoError(ltx.ExSignWith(util.Signer2))
	assert.NoError(ltx.SyntacticVerify())
	itx, err = NewTx(ltx)
	assert.NoError(err)
	assert.NoError(itx.Apply(ctx, cs), "should support 0 amount")

	assert.NoError(cs.VerifyState())
}
