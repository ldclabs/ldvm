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

func TestTokenAccount(t *testing.T) {
	assert := assert.New(t)

	acc := NewAccount(util.Signer1.Address())
	acc.Init(big.NewInt(0), 0, 0)
	amount := big.NewInt(1_000_000)
	cfg := &ld.TxAccounter{
		Threshold: ld.Uint16Ptr(1),
		Keepers:   &util.EthIDs{util.Signer1.Address(), util.Signer2.Address()},
		Amount:    amount,
	}
	assert.ErrorContains(acc.CreateToken(cfg),
		"Account(0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC).CreateToken error: invalid token 0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC")
	assert.ErrorContains(acc.DestroyToken(acc),
		"Account(0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC).DestroyToken error: invalid token account 0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC")

	// create NativeToken
	nativeToken := NewAccount(constants.LDCAccount)
	assert.NoError(nativeToken.CreateToken(&ld.TxAccounter{
		Amount: amount,
	}))
	assert.Equal(true, nativeToken.valid(ld.TokenAccount))
	assert.Equal(amount.Uint64(), nativeToken.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(amount.Uint64(), nativeToken.balanceOfAll(constants.NativeToken).Uint64())
	assert.Equal(amount.Uint64(), nativeToken.ld.MaxTotalSupply.Uint64())
	assert.Equal(uint16(0), nativeToken.Threshold())
	assert.Equal(util.EthIDs{}, nativeToken.Keepers())
	assert.False(nativeToken.SatisfySigning(util.EthIDs{}), "no controller")
	assert.False(nativeToken.SatisfySigning(util.EthIDs{util.Signer1.Address()}), "no controller")
	assert.False(nativeToken.SatisfySigningPlus(util.EthIDs{}), "no controller")
	assert.False(nativeToken.SatisfySigningPlus(util.EthIDs{util.Signer1.Address()}), "no controller")
	assert.ErrorContains(nativeToken.DestroyToken(acc), "invalid token")

	nativeToken.Sub(constants.NativeToken, big.NewInt(1000))
	acc.Add(constants.NativeToken, big.NewInt(1000))
	assert.Equal(uint64(999000), nativeToken.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(uint64(999000), nativeToken.balanceOfAll(constants.NativeToken).Uint64())
	assert.Equal(uint64(1000), acc.balanceOf(constants.NativeToken).Uint64())

	// CheckAsFrom
	for _, ty := range ld.AllTxTypes {
		switch {
		case ld.TokenFromTxTypes.Has(ty):
			assert.NoError(nativeToken.CheckAsFrom(ty))
		default:
			assert.Error(nativeToken.CheckAsFrom(ty))
		}
	}
	// CheckAsTo
	for _, ty := range ld.AllTxTypes {
		switch {
		case ld.TokenToTxTypes.Has(ty):
			assert.NoError(nativeToken.CheckAsTo(ty))
		default:
			assert.Error(nativeToken.CheckAsTo(ty))
		}
	}

	// Marshal
	data, _, err := nativeToken.Marshal()
	assert.NoError(err)
	acc2, err := ParseAccount(nativeToken.id, data)
	assert.NoError(err)
	assert.Equal(nativeToken.ld.Bytes(), acc2.ld.Bytes())

	token := ld.MustNewToken("$TEST")
	testToken := NewAccount(util.EthID(token))
	testToken.Init(big.NewInt(100), 0, 0)
	assert.NoError(testToken.CreateToken(cfg))
	assert.Equal(false, testToken.valid(ld.TokenAccount))
	testToken.Add(constants.NativeToken, big.NewInt(100))
	assert.Equal(true, testToken.valid(ld.TokenAccount))

	assert.Equal(uint64(0), testToken.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(uint64(100), testToken.balanceOfAll(constants.NativeToken).Uint64())
	assert.Equal(amount.Uint64(), testToken.balanceOf(token).Uint64())
	assert.Equal(amount.Uint64(), testToken.balanceOfAll(token).Uint64())
	assert.Equal(amount.Uint64(), testToken.ld.MaxTotalSupply.Uint64())
	assert.Equal(uint16(1), testToken.Threshold())
	assert.Equal(util.EthIDs{util.Signer1.Address(), util.Signer2.Address()}, testToken.Keepers())
	assert.False(testToken.SatisfySigning(util.EthIDs{}))
	assert.True(testToken.SatisfySigning(util.EthIDs{util.Signer1.Address()}))
	assert.True(testToken.SatisfySigning(util.EthIDs{util.Signer2.Address()}))
	assert.False(testToken.SatisfySigningPlus(util.EthIDs{}))
	assert.False(testToken.SatisfySigningPlus(util.EthIDs{util.Signer1.Address()}))
	assert.True(testToken.SatisfySigningPlus(util.EthIDs{util.Signer1.Address(), util.Signer2.Address()}))

	testToken.Sub(token, big.NewInt(1000))
	acc.Add(token, big.NewInt(1000))
	assert.Equal(uint64(999000), testToken.balanceOf(token).Uint64())
	assert.Equal(uint64(999000), testToken.balanceOfAll(token).Uint64())
	assert.Equal(uint64(1000), acc.balanceOf(token).Uint64())
	testToken.Sub(token, big.NewInt(1000))
	acc.Add(token, big.NewInt(1000))
	assert.Equal(uint64(998000), testToken.balanceOf(token).Uint64())
	assert.Equal(uint64(2000), acc.balanceOf(token).Uint64())

	// Marshal
	data, _, err = testToken.Marshal()
	assert.NoError(err)
	acc2, err = ParseAccount(testToken.id, data)
	assert.NoError(err)
	assert.Equal(testToken.ld.Bytes(), acc2.ld.Bytes())

	// Lending
	lcfg := &ld.LendingConfig{
		Token:           token,
		DailyInterest:   10,
		OverdueInterest: 10,
		MinAmount:       big.NewInt(1000),
		MaxAmount:       big.NewInt(1_000_000),
	}
	assert.NoError(testToken.InitLedger(nil))
	assert.NoError(testToken.OpenLending(lcfg))
	assert.NotNil(testToken.ld.Lending)
	assert.NotNil(testToken.ledger)

	// Destroy
	assert.ErrorContains(testToken.DestroyToken(acc), "some token in the use")
	assert.NoError(testToken.Borrow(token, acc.id, big.NewInt(1000), 0))
	assert.ErrorContains(testToken.DestroyToken(acc),
		"Account(0x0000000000000000000000000000002454455354).DestroyToken error: some token in the use, maxTotalSupply expected 1000000, got 998000")
	actual, err := testToken.Repay(token, acc.id, big.NewInt(1000))
	assert.NoError(err)
	assert.Equal(uint64(1000), actual.Uint64())

	assert.ErrorContains(testToken.DestroyToken(acc), "some token in the use")
	testToken.Add(token, big.NewInt(2000))
	assert.NoError(testToken.DestroyToken(acc))
	assert.Equal(uint64(0), testToken.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(uint64(0), testToken.balanceOfAll(constants.NativeToken).Uint64())
	assert.Equal(uint64(0), testToken.balanceOf(token).Uint64())
	assert.Equal(uint64(0), testToken.balanceOfAll(token).Uint64())
	assert.Equal(uint16(0), testToken.Threshold())
	assert.Equal(util.EthIDs{}, testToken.Keepers())
	assert.Equal(0, len(testToken.ld.Tokens))
	assert.Nil(testToken.ld.MaxTotalSupply)
	assert.Nil(testToken.ld.Lending)
	assert.Equal(0, len(testToken.ledger.Lending))
	assert.Equal(uint64(1100), acc.balanceOf(constants.NativeToken).Uint64())

	// Destroy again
	assert.ErrorContains(testToken.DestroyToken(acc), "invalid token account")

	// Marshal again
	data, _, err = testToken.Marshal()
	assert.NoError(err)
	acc2, err = ParseAccount(testToken.id, data)
	assert.NoError(err)
	assert.Equal(testToken.ld.Bytes(), acc2.ld.Bytes())

	// Create again
	assert.NoError(testToken.CreateToken(cfg))
	assert.Equal(false, testToken.valid(ld.TokenAccount))
	testToken.Add(constants.NativeToken, big.NewInt(100))
	assert.Equal(true, testToken.valid(ld.TokenAccount))
}
