// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package transaction

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ldclabs/ldvm/constants"
	"github.com/ldclabs/ldvm/ld"
	"github.com/ldclabs/ldvm/util"
	"github.com/stretchr/testify/assert"
)

func TestTxEth(t *testing.T) {
	assert := assert.New(t)

	// SyntacticVerify
	tx := &TxEth{}
	assert.ErrorContains(tx.SyntacticVerify(), "nil pointer")
	_, err := tx.MarshalJSON()
	assert.NoError(err)

	bctx := NewMockBCtx()
	bs := bctx.MockBS()
	token := ld.MustNewToken("$LDC")

	from := bs.MustAccount(util.Signer1.Address())
	to := bs.MustAccount(util.Signer2.Address())

	testTo := common.Address(to.id)

	txe, err := ld.NewEthTx(&types.AccessListTx{
		ChainID:  new(big.Int).SetUint64(bctx.ChainConfig().ChainID),
		Nonce:    0,
		To:       &testTo,
		Value:    ld.ToEthBalance(big.NewInt(1_000_000)),
		Gas:      0,
		GasPrice: ld.ToEthBalance(new(big.Int).SetUint64(bctx.Price)),
	})
	assert.NoError(err)
	tt := txe.ToTransaction()
	itx, err := NewTx2(tt)
	assert.NoError(err)
	tx = itx.(*TxEth)

	tx.ld.To = nil
	assert.ErrorContains(itx.SyntacticVerify(), "invalid to")
	tx.ld.To = &to.id
	tx.ld.Amount = nil
	assert.ErrorContains(itx.SyntacticVerify(), "invalid amount")
	tx.ld.Amount = big.NewInt(1_000_000)
	data := tx.ld.Data
	tx.ld.Data = []byte{}
	assert.ErrorContains(itx.SyntacticVerify(), "invalid data")
	tx.ld.Data = data
	tx.ld.ChainID = 1
	assert.ErrorContains(itx.SyntacticVerify(), "invalid chainID")
	tx.ld.ChainID = bctx.ChainConfig().ChainID
	tx.ld.Nonce = 1
	assert.ErrorContains(itx.SyntacticVerify(), "invalid nonce")
	tx.ld.Nonce = 0
	tx.ld.GasTip = 1
	assert.ErrorContains(itx.SyntacticVerify(), "invalid gasTip")
	tx.ld.GasTip = 0
	tx.ld.GasFeeCap = bctx.Price - 1
	assert.ErrorContains(itx.SyntacticVerify(), "invalid gasFeeCap")
	tx.ld.GasFeeCap = bctx.Price
	tx.ld.From = constants.GenesisAccount
	assert.ErrorContains(itx.SyntacticVerify(), "invalid from")
	tx.ld.From = from.id
	tx.ld.To = &constants.GenesisAccount
	assert.ErrorContains(itx.SyntacticVerify(), "invalid to")
	tx.ld.To = &to.id
	tx.ld.Token = &token
	assert.ErrorContains(itx.SyntacticVerify(), "invalid token")
	tx.ld.Token = nil
	tx.ld.Amount = big.NewInt(1_000_000 - 1)
	assert.ErrorContains(itx.SyntacticVerify(), "invalid amount")
	tx.ld.Amount = big.NewInt(1_000_000)
	sigs := tx.ld.Signatures
	tx.ld.Signatures = nil
	assert.ErrorContains(itx.SyntacticVerify(), "invalid signatures")
	tx.ld.Signatures = []util.Signature{{}}
	assert.ErrorContains(itx.SyntacticVerify(), "invalid signatures")
	tx.ld.Signatures = sigs
	tx.ld.ExSignatures = []util.Signature{}
	assert.ErrorContains(itx.SyntacticVerify(), "invalid exSignatures")
	tx.ld.ExSignatures = nil
	assert.NoError(itx.SyntacticVerify())

	itx, err = NewTx2(tt)
	assert.NoError(err)

	bs.CommitAccounts()
	assert.ErrorContains(itx.Apply(bctx, bs),
		"insufficient NativeLDC balance, expected 2245000, got 0")
	bs.CheckoutAccounts()

	from.Add(constants.NativeToken, new(big.Int).SetUint64(constants.LDC))
	assert.NoError(itx.Apply(bctx, bs))

	assert.Equal(tt.Gas()*bctx.Price,
		itx.(*TxEth).ldc.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(uint64(0),
		itx.(*TxEth).miner.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(uint64(1_000_000), to.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(constants.LDC-tt.Gas()*(bctx.Price)-1_000_000,
		from.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(uint64(1), from.Nonce())

	jsondata, err := itx.MarshalJSON()
	assert.NoError(err)
	fmt.Println(string(jsondata))
	assert.Equal(`{"type":"TypeEth","chainID":2357,"nonce":0,"gasTip":0,"gasFeeCap":1000,"from":"0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC","to":"0x44171C37Ff5D7B7bb8dcad5C81f16284A229e641","amount":1000000,"data":"0x01f86d8209358085e8d4a51000809444171c37ff5d7b7bb8dcad5c81f16284a229e64187038d7ea4c6800080c080a01da0480ceece3e7dc8cc88fe962b29decbabf7c64b1e4acb4e3317fe8953a0d3a01ba1253e2bccbb7ec2f4359111f2c29b3c8c69e6ed56fe66a3c10b658ba7efc4c3037b00","signatures":["1da0480ceece3e7dc8cc88fe962b29decbabf7c64b1e4acb4e3317fe8953a0d31ba1253e2bccbb7ec2f4359111f2c29b3c8c69e6ed56fe66a3c10b658ba7efc400"],"id":"2LkK2JNjNXavmRcPLnq8YD11CH5yYhsDCuLiG5r9CPx6r6qs28"}`, string(jsondata))

	assert.NoError(bs.VerifyState())
}
