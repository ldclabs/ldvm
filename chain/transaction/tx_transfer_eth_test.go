// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package transaction

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

	bctx := NewMockBCtx()
	bs := bctx.MockBS()
	token := ld.MustNewToken("$LDC")

	from := bs.MustAccount(util.Signer1.Address())
	to := bs.MustAccount(util.Signer2.Address())

	testTo := common.Address(to.id)

	txe, err := ld.NewEthTx(&types.AccessListTx{
		ChainID:  new(big.Int).SetUint64(bctx.Chain().ChainID),
		Nonce:    0,
		To:       &testTo,
		Value:    big.NewInt(1_000_000),
		Gas:      0,
		GasPrice: new(big.Int).SetUint64(bctx.Price),
	})
	assert.NoError(err)
	tt := txe.ToTransaction()
	itx, err := NewTx(tt, true)
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
	tx.ld.ChainID = bctx.Chain().ChainID
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

	itx, err = NewTx(tt, true)
	assert.NoError(err)

	bs.CommitAccounts()
	assert.ErrorContains(itx.Apply(bctx, bs),
		"insufficient NativeLDC balance, expected 2204000, got 0")
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
	// fmt.Println(string(jsondata))
	assert.Equal(`{"type":"TypeEth","chainID":2357,"nonce":0,"gasTip":0,"gasFeeCap":1000,"from":"0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC","to":"0x44171C37Ff5D7B7bb8dcad5C81f16284A229e641","amount":1000000,"data":"0x01f866820935808203e8809444171c37ff5d7b7bb8dcad5c81f16284a229e641830f424080c080a03b6f2401a270681519ba730c2f53bb87c2fd41ed7d91434ddfbfec506e6c8768a018cb0cde3ef83bdd83809f63852d7fc07de699239007a8046d48006ec5cb66ac271d15b9","signatures":["3b6f2401a270681519ba730c2f53bb87c2fd41ed7d91434ddfbfec506e6c876818cb0cde3ef83bdd83809f63852d7fc07de699239007a8046d48006ec5cb66ac00"],"id":"LahrjRUUeJfjsvWAc9GLpN2ArK5sTNN2kJnHoS8RggVE5dMzi"}`, string(jsondata))

	assert.NoError(bs.VerifyState())
}
