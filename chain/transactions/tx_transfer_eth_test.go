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

	ctx := NewMockChainContext()
	cs := ctx.MockChainState()
	token := ld.MustNewToken("$LDC")

	from := cs.MustAccount(util.Signer1.Address())
	to := cs.MustAccount(util.Signer2.Address())

	testTo := common.Address(to.id)

	txe, err := ld.NewEthTx(&types.AccessListTx{
		ChainID:  new(big.Int).SetUint64(ctx.ChainConfig().ChainID),
		Nonce:    0,
		To:       &testTo,
		Value:    ld.ToEthBalance(big.NewInt(1_000_000)),
		Gas:      0,
		GasPrice: ld.ToEthBalance(new(big.Int).SetUint64(ctx.Price)),
	})
	assert.NoError(err)
	ltx := txe.ToTransaction()
	assert.NoError(ltx.SyntacticVerify())
	itx, err := NewTx(ltx)
	assert.NoError(err)
	tx = itx.(*TxEth)

	tx.ld.Tx.To = nil
	assert.ErrorContains(itx.SyntacticVerify(), "invalid to")
	tx.ld.Tx.To = &to.id
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
	tx.ld.Tx.From = from.id
	tx.ld.Tx.To = &constants.GenesisAccount
	assert.ErrorContains(itx.SyntacticVerify(), "invalid to")
	tx.ld.Tx.To = &to.id
	tx.ld.Tx.Token = &token
	assert.ErrorContains(itx.SyntacticVerify(), "invalid token")
	tx.ld.Tx.Token = nil
	tx.ld.Tx.Amount = big.NewInt(1_000_000 - 1)
	assert.ErrorContains(itx.SyntacticVerify(), "invalid amount")
	tx.ld.Tx.Amount = big.NewInt(1_000_000)
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
	itx, err = NewTx(ltx)
	assert.NoError(err)

	cs.CommitAccounts()
	assert.ErrorContains(itx.Apply(ctx, cs),
		"insufficient NativeLDC balance, expected 2268000, got 0")
	cs.CheckoutAccounts()

	from.Add(constants.NativeToken, new(big.Int).SetUint64(constants.LDC))
	assert.NoError(itx.Apply(ctx, cs))

	assert.Equal(ltx.Gas()*ctx.Price,
		itx.(*TxEth).ldc.Balance().Uint64())
	assert.Equal(uint64(0),
		itx.(*TxEth).miner.Balance().Uint64())
	assert.Equal(uint64(1_000_000), to.Balance().Uint64())
	assert.Equal(constants.LDC-ltx.Gas()*(ctx.Price)-1_000_000,
		from.Balance().Uint64())
	assert.Equal(uint64(1), from.Nonce())

	jsondata, err := itx.MarshalJSON()
	assert.NoError(err)
	// fmt.Println(string(jsondata))
	assert.Equal(`{"tx":{"type":"TypeEth","chainID":2357,"nonce":0,"gasTip":0,"gasFeeCap":1000,"from":"0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC","to":"0x44171C37Ff5D7B7bb8dcad5C81f16284A229e641","amount":1000000,"data":"0x01f86d8209358085e8d4a51000809444171c37ff5d7b7bb8dcad5c81f16284a229e64187038d7ea4c6800080c080a01da0480ceece3e7dc8cc88fe962b29decbabf7c64b1e4acb4e3317fe8953a0d3a01ba1253e2bccbb7ec2f4359111f2c29b3c8c69e6ed56fe66a3c10b658ba7efc4c3037b00"},"sigs":["1da0480ceece3e7dc8cc88fe962b29decbabf7c64b1e4acb4e3317fe8953a0d31ba1253e2bccbb7ec2f4359111f2c29b3c8c69e6ed56fe66a3c10b658ba7efc400"],"id":"2rSG3GB6X3cySkF2ZXuYLRBTWeoH81J9fgRwLqQpdxqugYszuE"}`, string(jsondata))

	assert.NoError(cs.VerifyState())
}
