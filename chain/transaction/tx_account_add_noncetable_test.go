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

func TestTxAddAccountNonceTable(t *testing.T) {
	assert := assert.New(t)

	// SyntacticVerify
	tx := &TxAddAccountNonceTable{}
	assert.ErrorContains(tx.SyntacticVerify(), "nil pointer")
	_, err := tx.MarshalJSON()
	assert.NoError(err)

	bctx := NewMockBCtx()
	bs := NewMockBS(bctx)

	from, err := bs.LoadAccount(util.Signer1.Address())
	assert.NoError(err)

	txData := &ld.TxData{
		Type:      ld.TypeAddNonceTable,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      from.id,
	}
	assert.NoError(txData.SyntacticVerify())
	_, err = NewTx(txData.ToTransaction(), true)
	assert.ErrorContains(err, "DeriveSigners: no signature")

	txData = &ld.TxData{
		Type:      ld.TypeAddNonceTable,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      from.id,
		To:        &constants.GenesisAccount,
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx(txData.ToTransaction(), true)
	assert.ErrorContains(err, "invalid to, should be nil")

	txData = &ld.TxData{
		Type:      ld.TypeAddNonceTable,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      from.id,
		Token:     &constants.NativeToken,
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx(txData.ToTransaction(), true)
	assert.ErrorContains(err, "invalid token, should be nil")

	txData = &ld.TxData{
		Type:      ld.TypeAddNonceTable,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      from.id,
		Amount:    big.NewInt(1),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx(txData.ToTransaction(), true)
	assert.ErrorContains(err, "nil to together with amount")

	txData = &ld.TxData{
		Type:      ld.TypeAddNonceTable,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      from.id,
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx(txData.ToTransaction(), true)
	assert.ErrorContains(err, "invalid data")

	txData = &ld.TxData{
		Type:      ld.TypeAddNonceTable,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      from.id,
		Data:      []byte("你好👋"),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx(txData.ToTransaction(), true)
	assert.ErrorContains(err, "cbor: cannot unmarshal")

	input := []uint64{10}
	inputData, err := ld.MarshalCBOR(input)
	assert.NoError(err)
	txData = &ld.TxData{
		Type:      ld.TypeAddNonceTable,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      from.id,
		Data:      inputData,
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx(txData.ToTransaction(), true)
	assert.ErrorContains(err, "no nonce")

	input = make([]uint64, 1026)
	for i := range input {
		input[i] = uint64(i)
	}
	inputData, err = ld.MarshalCBOR(input)
	assert.NoError(err)
	txData = &ld.TxData{
		Type:      ld.TypeAddNonceTable,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      from.id,
		Data:      inputData,
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx(txData.ToTransaction(), true)
	assert.ErrorContains(err, "too many nonces, expected <= 1024, got 1025")

	input = []uint64{bs.Timestamp() - 1, 123}
	inputData, err = ld.MarshalCBOR(input)
	assert.NoError(err)
	txData = &ld.TxData{
		Type:      ld.TypeAddNonceTable,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      from.id,
		Data:      inputData,
	}
	assert.NoError(txData.SignWith(util.Signer1))
	tt := txData.ToTransaction()
	tt.Timestamp = bs.Timestamp()
	_, err = NewTx(tt, true)
	assert.ErrorContains(err, "invalid expire time, expected > 1000, got 999")

	input = []uint64{3600*24*30 + 2, 123}
	inputData, err = ld.MarshalCBOR(input)
	assert.NoError(err)
	txData = &ld.TxData{
		Type:      ld.TypeAddNonceTable,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      from.id,
		Data:      inputData,
	}
	assert.NoError(txData.SignWith(util.Signer1))
	tt = txData.ToTransaction()
	tt.Timestamp = 1
	_, err = NewTx(tt, true)
	assert.ErrorContains(err, "invalid expire time, expected <= 2592001, got 2592002")

	input = []uint64{bs.Timestamp() + 1, 1, 3, 7, 5}
	inputData, err = ld.MarshalCBOR(input)
	assert.NoError(err)
	txData = &ld.TxData{
		Type:      ld.TypeAddNonceTable,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      from.id,
		Data:      inputData,
	}
	assert.NoError(txData.SignWith(util.Signer1))
	tt = txData.ToTransaction()
	tt.Timestamp = bs.Timestamp()
	itx, err := NewTx(tt, true)
	assert.NoError(err)
	assert.ErrorContains(itx.Verify(bctx, bs), "invalid gas, expected 101, got 0")

	tt.Gas = tt.RequiredGas(bctx.FeeConfig().ThresholdGas)
	itx, err = NewTx(tt, true)
	assert.NoError(err)
	assert.ErrorContains(itx.Verify(bctx, bs), "insufficient NativeLDC balance, expected 111100, got 0")

	from.Add(constants.NativeToken, new(big.Int).SetUint64(constants.LDC))
	assert.NoError(itx.Verify(bctx, bs))
	assert.NoError(itx.Accept(bctx, bs))

	tx = itx.(*TxAddAccountNonceTable)
	assert.Equal(tx.ld.Gas*bctx.Price, tx.ldc.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(tx.ld.Gas*100, tx.miner.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(constants.LDC-tx.ld.Gas*(bctx.Price+100),
		from.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(uint64(1), from.Nonce())
	assert.Equal(1, len(from.ld.NonceTable))
	assert.Equal([]uint64{1, 3, 5, 7}, tx.from.ld.NonceTable[bs.Timestamp()+1])

	jsondata, err := itx.MarshalJSON()
	assert.NoError(err)
	// fmt.Println(string(jsondata))
	assert.Equal(`{"type":"TypeAddNonceTable","chainID":2357,"nonce":0,"gasTip":100,"gasFeeCap":1000,"from":"0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC","data":[1001,1,3,7,5],"signatures":["ef07cf7075394c343ee99f34d2c76efaa3789ecc4b9c48f896aecd01e343f30c0d3e8c67958bf10a33979cdcf1fbcf9c3b6df7c6f7583ec795a3dace2f75b4c200"],"gas":101,"id":"svYgQEJj8X7cydgowuZ3Dj4pZftQFA4fzVreLeVUbUbhGxkwS"}`, string(jsondata))

	input = []uint64{bs.Timestamp() + 1, 2, 4, 1}
	inputData, err = ld.MarshalCBOR(input)
	assert.NoError(err)
	txData = &ld.TxData{
		Type:      ld.TypeAddNonceTable,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     1,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      from.id,
		Data:      inputData,
	}
	assert.NoError(txData.SignWith(util.Signer1))
	tt = txData.ToTransaction()
	tt.Timestamp = bs.Timestamp()
	tt.Gas = tt.RequiredGas(bctx.FeeConfig().ThresholdGas)
	itx, err = NewTx(tt, true)
	assert.NoError(err)
	assert.ErrorContains(itx.Verify(bctx, bs), "nonce 1 exists at 1001")

	input = []uint64{bs.Timestamp() + 1, 2, 4, 6}
	inputData, err = ld.MarshalCBOR(input)
	assert.NoError(err)
	txData = &ld.TxData{
		Type:      ld.TypeAddNonceTable,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     1,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      from.id,
		Data:      inputData,
	}
	assert.NoError(txData.SignWith(util.Signer1))
	tt = txData.ToTransaction()
	tt.Timestamp = bs.Timestamp()
	tt.Gas = tt.RequiredGas(bctx.FeeConfig().ThresholdGas)
	itx, err = NewTx(tt, true)
	assert.NoError(err)
	assert.NoError(itx.Verify(bctx, bs))
	assert.NoError(itx.Accept(bctx, bs))

	assert.Equal(uint64(2), from.Nonce())
	assert.Equal(1, len(from.ld.NonceTable))
	assert.Equal([]uint64{1, 2, 3, 4, 5, 6, 7}, tx.from.ld.NonceTable[bs.Timestamp()+1])

	input = []uint64{bs.Timestamp() + 2, 0}
	inputData, err = ld.MarshalCBOR(input)
	assert.NoError(err)
	txData = &ld.TxData{
		Type:      ld.TypeAddNonceTable,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     2,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      from.id,
		Data:      inputData,
	}
	assert.NoError(txData.SignWith(util.Signer1))
	tt = txData.ToTransaction()
	tt.Timestamp = bs.Timestamp()
	tt.Gas = tt.RequiredGas(bctx.FeeConfig().ThresholdGas)
	itx, err = NewTx(tt, true)
	assert.NoError(err)
	assert.NoError(itx.Verify(bctx, bs))
	assert.NoError(itx.Accept(bctx, bs))

	assert.Equal(uint64(3), from.Nonce())
	assert.Equal(2, len(from.ld.NonceTable))
	assert.Equal([]uint64{1, 2, 3, 4, 5, 6, 7}, tx.from.ld.NonceTable[bs.Timestamp()+1])
	assert.Equal([]uint64{0}, tx.from.ld.NonceTable[bs.Timestamp()+2])

	// consume nonce table
	to, err := bs.LoadAccount(util.Signer2.Address())
	assert.NoError(err)
	input2 := ld.TxTransfer{
		Nonce:  0,
		From:   &from.id,
		To:     &to.id,
		Amount: new(big.Int).SetUint64(constants.MilliLDC),
		Expire: bs.Timestamp() + 1,
	}
	assert.NoError(input2.SyntacticVerify())
	txData = &ld.TxData{
		Type:      ld.TypeTransferCash,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      to.id,
		To:        &from.id,
		Data:      input2.Bytes(),
	}
	assert.NoError(txData.SyntacticVerify())
	assert.NoError(txData.SignWith(util.Signer2))
	assert.NoError(txData.ExSignWith(util.Signer1))
	tt = txData.ToTransaction()
	tt.Timestamp = bs.Timestamp()
	tt.Gas = tt.RequiredGas(bctx.FeeConfig().ThresholdGas)
	to.Add(constants.NativeToken, new(big.Int).SetUint64(constants.LDC))
	itx, err = NewTx(tt, true)
	assert.NoError(err)
	assert.ErrorContains(itx.Verify(bctx, bs), "nonce 0 not exists at 1001")

	input2 = ld.TxTransfer{
		Nonce:  0,
		From:   &from.id,
		To:     &to.id,
		Amount: new(big.Int).SetUint64(constants.MilliLDC),
		Expire: bs.Timestamp() + 2,
	}
	assert.NoError(input2.SyntacticVerify())
	txData = &ld.TxData{
		Type:      ld.TypeTransferCash,
		ChainID:   bctx.Chain().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      to.id,
		To:        &from.id,
		Data:      input2.Bytes(),
	}
	assert.NoError(txData.SyntacticVerify())
	assert.NoError(txData.SignWith(util.Signer2))
	assert.NoError(txData.ExSignWith(util.Signer1))
	tt = txData.ToTransaction()
	tt.Timestamp = bs.Timestamp()
	tt.Gas = tt.RequiredGas(bctx.FeeConfig().ThresholdGas)
	itx, err = NewTx(tt, true)
	assert.NoError(err)
	assert.NoError(itx.Verify(bctx, bs))
	assert.NoError(itx.Accept(bctx, bs))

	assert.Equal(1, len(from.ld.NonceTable))
	assert.Equal([]uint64{1, 2, 3, 4, 5, 6, 7}, tx.from.ld.NonceTable[bs.Timestamp()+1])
	assert.Nil(tx.from.ld.NonceTable[bs.Timestamp()+2], "should clean emtpy nonce table")

	assert.NoError(bs.VerifyState())
}
