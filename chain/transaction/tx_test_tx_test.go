// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package transaction

import (
	"math/big"
	"testing"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ldclabs/ldvm/constants"
	"github.com/ldclabs/ldvm/ld"
	"github.com/ldclabs/ldvm/util"
	"github.com/stretchr/testify/assert"
)

func TestTxTest(t *testing.T) {
	assert := assert.New(t)

	// SyntacticVerify
	tx := &TxTest{}
	assert.ErrorContains(tx.SyntacticVerify(), "nil pointer")
	_, err := tx.MarshalJSON()
	assert.NoError(err)

	bctx := NewMockBCtx()
	bs := bctx.MockBS()
	sender := util.Signer1.Address()
	to := util.Signer2.Address()

	txData := &ld.TxData{
		Type:      ld.TypeTest,
		ChainID:   bctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      sender,
	}
	assert.NoError(txData.SyntacticVerify())
	_, err = NewTx2(txData.ToTransaction())
	assert.ErrorContains(err, "DeriveSigners error: no signature")

	txData = &ld.TxData{
		Type:      ld.TypeTest,
		ChainID:   bctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      sender,
		To:        &constants.GenesisAccount,
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx2(txData.ToTransaction())
	assert.ErrorContains(err, "invalid to, should be nil")

	txData = &ld.TxData{
		Type:      ld.TypeTest,
		ChainID:   bctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      sender,
		Token:     &constants.NativeToken,
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx2(txData.ToTransaction())
	assert.ErrorContains(err, "invalid token, should be nil")

	txData = &ld.TxData{
		Type:      ld.TypeTest,
		ChainID:   bctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      sender,
		Amount:    big.NewInt(1),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx2(txData.ToTransaction())
	assert.ErrorContains(err, "nil to together with amount")

	txData = &ld.TxData{
		Type:      ld.TypeTest,
		ChainID:   bctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      sender,
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx2(txData.ToTransaction())
	assert.ErrorContains(err, "invalid data")

	txData = &ld.TxData{
		Type:      ld.TypeTest,
		ChainID:   bctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      sender,
		Data:      []byte("你好👋"),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx2(txData.ToTransaction())
	assert.ErrorContains(err, "cbor: unexpected following extraneous data")

	input := ld.TxTester{
		ObjectType: ld.AddressObject,
		ObjectID:   ids.ShortID(to),
	}
	txData = &ld.TxData{
		Type:      ld.TypeTest,
		ChainID:   bctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      sender,
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	_, err = NewTx2(txData.ToTransaction())
	assert.ErrorContains(err,
		"TxTest.SyntacticVerify error: TxTester.SyntacticVerify error: empty tests")

	input = ld.TxTester{
		ObjectType: ld.AddressObject,
		ObjectID:   ids.ShortID(sender),
		Tests: ld.TestOps{
			{Path: "/b", Value: util.MustMarshalCBOR(new(big.Int).SetUint64(constants.LDC))},
		},
	}
	assert.NoError(input.SyntacticVerify())
	txData = &ld.TxData{
		Type:      ld.TypeTest,
		ChainID:   bctx.ChainConfig().ChainID,
		Nonce:     0,
		GasTip:    100,
		GasFeeCap: bctx.Price,
		From:      sender,
		Data:      input.Bytes(),
	}
	assert.NoError(txData.SignWith(util.Signer1))
	tt := txData.ToTransaction()
	tt.Timestamp = bs.Timestamp()
	itx, err := NewTx2(tt)
	assert.NoError(err)

	bs.CommitAccounts()
	assert.ErrorContains(itx.Apply(bctx, bs),
		"insufficient NativeLDC balance, expected 738100, got 0")
	bs.CheckoutAccounts()

	senderAcc := bs.MustAccount(sender)
	senderAcc.Add(constants.NativeToken, new(big.Int).SetUint64(constants.LDC))
	assert.NoError(itx.Apply(bctx, bs))

	senderGas := tt.Gas()
	assert.Equal(senderGas*bctx.Price,
		itx.(*TxTest).ldc.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(senderGas*100,
		itx.(*TxTest).miner.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(constants.LDC-tt.Gas()*(bctx.Price+100),
		senderAcc.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(uint64(1), senderAcc.Nonce())

	jsondata, err := itx.MarshalJSON()
	assert.NoError(err)
	// fmt.Println(string(jsondata))
	assert.Equal(`{"type":"TypeTest","chainID":2357,"nonce":0,"gasTip":100,"gasFeeCap":1000,"from":"0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC","data":{"objectType":"Address","objectId":"0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC","tests":[{"path":"/b","value":"0xc2443b9aca00dfb73dae"}]},"signatures":["455c06419f62d87ce2fbb91feb11441c7926590a079613b105bf48ebf2fb286137ca45185d5ec6bc754288d6894244942660db94ad1118849b9d65ebddefa4d701"],"id":"UDSNmc4yoNvgugjusj5Z1hnujt5UT2Za6TWUhgbM1PMAFmcAL"}`, string(jsondata))
}
