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

func TestTxTransferPay(t *testing.T) {
	assert := assert.New(t)

	// SyntacticVerify
	tx := &TxTransferPay{}
	assert.ErrorContains(tx.SyntacticVerify(), "nil pointer")
	_, err := tx.MarshalJSON()
	assert.NoError(err)

	token := ld.MustNewToken("$LDC")
	bctx := NewMockBCtx()
	bs := NewMockBS(bctx)

	from, err := bs.LoadAccount(util.Signer1.Address())
	assert.NoError(err)
	from.ld.Nonce = 1
	singer2 := util.Signer2.Address()
	to, err := bs.LoadAccount(constants.GenesisAccount)
	assert.NoError(err)
	assert.NoError(to.UpdateKeepers(ld.Uint16Ptr(1), &util.EthIDs{singer2}, nil, nil))

	txData := &ld.TxData{
		Type:      ld.TypeTransferPay,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     1,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      from.id,
	}
	assert.NoError(txData.SyntacticVerify())
	_, err = NewTx(txData.ToTransaction(), true)
	assert.ErrorContains(err, "DeriveSigners error: no signature")

	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx(txData.ToTransaction(), true)
	assert.ErrorContains(err, "invalid to")

	txData = &ld.TxData{
		Type:      ld.TypeTransferPay,
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
		Type:      ld.TypeTransferPay,
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
		Type:      ld.TypeTransferPay,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     1,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      from.id,
		To:        &to.id,
		Amount:    new(big.Int).SetUint64(constants.LDC),
		Data:      []byte("0"),
	}
	assert.NoError(txData.SyntacticVerify())
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx(txData.ToTransaction(), true)
	assert.ErrorContains(err, "cbor: cannot unmarshal")

	input := ld.TxTransfer{
		From: &singer2,
	}
	assert.NoError(input.SyntacticVerify())
	txData = &ld.TxData{
		Type:      ld.TypeTransferPay,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     1,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      from.id,
		To:        &to.id,
		Amount:    new(big.Int).SetUint64(constants.LDC),
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SyntacticVerify())
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx(txData.ToTransaction(), true)
	assert.ErrorContains(err,
		"invalid sender, expected 0x44171C37Ff5D7B7bb8dcad5C81f16284A229e641, got 0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC")

	input = ld.TxTransfer{
		From: &from.id,
	}
	assert.NoError(input.SyntacticVerify())
	txData = &ld.TxData{
		Type:      ld.TypeTransferPay,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     1,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      from.id,
		To:        &to.id,
		Amount:    new(big.Int).SetUint64(constants.LDC),
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SyntacticVerify())
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx(txData.ToTransaction(), true)
	assert.ErrorContains(err, "nil recipient")

	input = ld.TxTransfer{
		To: &constants.GenesisAccount,
	}
	assert.NoError(input.SyntacticVerify())
	txData = &ld.TxData{
		Type:      ld.TypeTransferPay,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     1,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      from.id,
		To:        &singer2,
		Amount:    new(big.Int).SetUint64(constants.LDC),
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SyntacticVerify())
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx(txData.ToTransaction(), true)
	assert.ErrorContains(err,
		"invalid recipient, expected 0xFFfFfFffFFfffFFfFFfFFFFFffFFFffffFfFFFfF, got 0x44171C37Ff5D7B7bb8dcad5C81f16284A229e641")

	input = ld.TxTransfer{
		To: &constants.GenesisAccount,
	}
	assert.NoError(input.SyntacticVerify())
	txData = &ld.TxData{
		Type:      ld.TypeTransferPay,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     1,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      from.id,
		To:        &to.id,
		Token:     &token,
		Amount:    new(big.Int).SetUint64(constants.LDC),
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SyntacticVerify())
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx(txData.ToTransaction(), true)
	assert.ErrorContains(err, "invalid token, expected NativeLDC, got $LDC")

	input = ld.TxTransfer{
		To:    &constants.GenesisAccount,
		Token: &token,
	}
	assert.NoError(input.SyntacticVerify())
	txData = &ld.TxData{
		Type:      ld.TypeTransferPay,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     1,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      from.id,
		To:        &to.id,
		Amount:    new(big.Int).SetUint64(constants.LDC),
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SyntacticVerify())
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx(txData.ToTransaction(), true)
	assert.ErrorContains(err, "invalid token, expected $LDC, got NativeLDC")

	input = ld.TxTransfer{
		To:    &constants.GenesisAccount,
		Token: &token,
	}
	assert.NoError(input.SyntacticVerify())
	txData = &ld.TxData{
		Type:      ld.TypeTransferPay,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     1,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      from.id,
		To:        &to.id,
		Token:     &token,
		Amount:    new(big.Int).SetUint64(constants.LDC),
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SyntacticVerify())
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx(txData.ToTransaction(), true)
	assert.ErrorContains(err, "nil amount")

	input = ld.TxTransfer{
		To:     &constants.GenesisAccount,
		Token:  &token,
		Amount: new(big.Int).SetUint64(constants.LDC),
	}
	assert.NoError(input.SyntacticVerify())
	txData = &ld.TxData{
		Type:      ld.TypeTransferPay,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     1,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      from.id,
		To:        &to.id,
		Token:     &token,
		Amount:    new(big.Int).SetUint64(constants.MilliLDC),
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SyntacticVerify())
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx(txData.ToTransaction(), true)
	assert.ErrorContains(err, "invalid amount, expected 1000000000, got 1000000")

	input = ld.TxTransfer{
		To:     &constants.GenesisAccount,
		Token:  &token,
		Amount: new(big.Int).SetUint64(constants.LDC),
		Expire: 10,
	}
	assert.NoError(input.SyntacticVerify())
	txData = &ld.TxData{
		Type:      ld.TypeTransferPay,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     1,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      from.id,
		To:        &to.id,
		Token:     &token,
		Amount:    new(big.Int).SetUint64(constants.LDC),
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SyntacticVerify())
	assert.NoError(txData.SignWith(util.Signer1))
	tt := txData.ToTransaction()
	tt.Timestamp = bs.Timestamp()
	_, err = NewTx(tt, true)
	assert.ErrorContains(err, "data expired")

	input = ld.TxTransfer{
		To:     &constants.GenesisAccount,
		Token:  &token,
		Amount: new(big.Int).SetUint64(constants.LDC),
	}
	assert.NoError(input.SyntacticVerify())
	txData = &ld.TxData{
		Type:      ld.TypeTransferPay,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     1,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      from.id,
		To:        &to.id,
		Token:     &token,
		Amount:    new(big.Int).SetUint64(constants.LDC),
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SyntacticVerify())
	assert.NoError(txData.SignWith(util.Signer1))
	tt = txData.ToTransaction()
	tt.Timestamp = bs.Timestamp()
	_, err = NewTx(tt, true)
	assert.ErrorContains(err, "DeriveSigners error: no signature")

	input = ld.TxTransfer{
		To:     &constants.GenesisAccount,
		Token:  &token,
		Amount: new(big.Int).SetUint64(constants.LDC),
	}
	assert.NoError(input.SyntacticVerify())
	txData = &ld.TxData{
		Type:      ld.TypeTransferPay,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     1,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      from.id,
		To:        &to.id,
		Token:     &token,
		Amount:    new(big.Int).SetUint64(constants.LDC),
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SyntacticVerify())
	assert.NoError(txData.SignWith(util.Signer1))
	assert.NoError(txData.ExSignWith(util.Signer1))
	tt = txData.ToTransaction()
	tt.Timestamp = bs.Timestamp()
	itx, err := NewTx(tt, true)
	assert.NoError(err)
	assert.ErrorContains(itx.Verify(bctx, bs),
		"TxBase.Verify error: invalid gas, expected 206, got 0")

	input = ld.TxTransfer{
		To:     &constants.GenesisAccount,
		Token:  &token,
		Amount: new(big.Int).SetUint64(constants.LDC),
	}
	assert.NoError(input.SyntacticVerify())
	txData = &ld.TxData{
		Type:      ld.TypeTransferPay,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     1,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      from.id,
		To:        &to.id,
		Token:     &token,
		Amount:    new(big.Int).SetUint64(constants.LDC),
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SyntacticVerify())
	assert.NoError(txData.SignWith(util.Signer1))
	assert.NoError(txData.ExSignWith(util.Signer1))
	tt = txData.ToTransaction()
	tt.Gas = tt.RequiredGas(bctx.FeeConfig().ThresholdGas)
	itx, err = NewTx(tt, true)
	assert.NoError(err)
	assert.ErrorContains(itx.Verify(bctx, bs),
		"insufficient NativeLDC balance, expected 226600, got 0")
	from.Add(constants.NativeToken, new(big.Int).SetUint64(constants.LDC))
	assert.ErrorContains(itx.Verify(bctx, bs),
		"insufficient $LDC balance, expected 1000000000, got 0")
	from.Add(token, new(big.Int).SetUint64(constants.LDC))
	assert.ErrorContains(itx.Verify(bctx, bs),
		"invalid exSignatures for recipient")

	txData = &ld.TxData{
		Type:      ld.TypeTransferPay,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     1,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      from.id,
		To:        &to.id,
		Token:     &token,
		Amount:    new(big.Int).SetUint64(constants.LDC),
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SyntacticVerify())
	assert.NoError(txData.SignWith(util.Signer1))
	assert.NoError(txData.ExSignWith(util.Signer2))
	tt = txData.ToTransaction()
	tt.Gas = tt.RequiredGas(bctx.FeeConfig().ThresholdGas)
	itx, err = NewTx(tt, true)
	assert.NoError(err)
	tx = itx.(*TxTransferPay)
	assert.NoError(itx.Verify(bctx, bs))
	assert.NoError(itx.Accept(bctx, bs))
	assert.Equal(tx.ld.Gas*bctx.Price, tx.ldc.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(tx.ld.Gas*100, tx.miner.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(constants.LDC, to.balanceOf(token).Uint64())
	assert.Equal(uint64(0), from.balanceOf(token).Uint64())
	assert.Equal(constants.LDC-tx.ld.Gas*(bctx.Price+100),
		from.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(uint64(2), tx.from.Nonce())

	jsondata, err := itx.MarshalJSON()
	assert.NoError(err)
	// fmt.Println(string(jsondata))
	assert.Equal(`{"type":"TypeTransferPay","chainID":2357,"nonce":1,"gasTip":100,"gasFeeCap":1000,"from":"0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC","to":"0xFFfFfFffFFfffFFfFFfFFFFFffFFFffffFfFFFfF","token":"$LDC","amount":1000000000,"data":{"to":"0xFFfFfFffFFfffFFfFFfFFFFFffFFFffffFfFFFfF","token":"$LDC","amount":1000000000},"signatures":["a1fdfe1053216c95e4dabbf0f6bf0ba602672d35097ce907d7ac156f3474f6853f99c5f7507039b3b8d4bf9d54b8b24a93ca10208b5808dcb91441035bef249f00"],"exSignatures":["c5613bb2ac47e7d5a8be0f58584791158838ef5dbfd11031b41c0576560e9af373175c38018c19d7b1d4f5d5ebabc3d82cc61d6a1e914727141ccb2cdc9a7dfb00"],"gas":206,"id":"2dsFu9TNZWVLW7pmfBVwVRCmCd9hbRXiFpr6X33KekRP3tFZCE"}`, string(jsondata))

	// should support 0 amount
	input = ld.TxTransfer{
		From:   &from.id,
		To:     &constants.GenesisAccount,
		Amount: new(big.Int).SetUint64(0),
		Data:   []byte(`"some data`),
	}
	assert.NoError(input.SyntacticVerify())
	txData = &ld.TxData{
		Type:      ld.TypeTransferPay,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     2,
		GasFeeCap: bctx.Price,
		From:      from.id,
		To:        &to.id,
		Amount:    new(big.Int).SetUint64(0),
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SyntacticVerify())
	assert.NoError(txData.SignWith(util.Signer1))
	assert.NoError(txData.ExSignWith(util.Signer2))
	tt = txData.ToTransaction()
	tt.Gas = tt.RequiredGas(bctx.FeeConfig().ThresholdGas)
	itx, err = NewTx(tt, true)
	assert.NoError(err)
	assert.NoError(itx.Verify(bctx, bs))
	assert.NoError(itx.Accept(bctx, bs), "should support 0 amount")

	assert.NoError(bs.VerifyState())
}
