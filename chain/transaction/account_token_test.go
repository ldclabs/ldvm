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
	assert.ErrorContains(acc.CheckCreateToken(cfg),
		"Account(0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC).CheckCreateToken error: invalid token 0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC")
	assert.ErrorContains(acc.CreateToken(cfg),
		"Account(0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC).CreateToken error: invalid token 0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC")
	assert.ErrorContains(acc.CheckDestroyToken(acc),
		"Account(0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC).CheckDestroyToken error: invalid token account 0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC")
	assert.ErrorContains(acc.DestroyToken(acc),
		"Account(0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC).DestroyToken error: invalid token account 0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC")

	// create NativeToken
	nativeToken := NewAccount(constants.LDCAccount)
	assert.NoError(nativeToken.CheckCreateToken(&ld.TxAccounter{
		Amount: amount,
	}))
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
	assert.ErrorContains(nativeToken.CheckDestroyToken(acc), "invalid token")
	assert.ErrorContains(nativeToken.CheckDestroyToken(acc), "invalid token")

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
	data, err := nativeToken.Marshal()
	assert.NoError(err)
	acc2, err := ParseAccount(nativeToken.id, data)
	assert.NoError(err)
	assert.Equal(nativeToken.ld.Bytes(), acc2.ld.Bytes())

	token := ld.MustNewToken("$TEST")
	testToken := NewAccount(util.EthID(token))
	testToken.Init(big.NewInt(100), 0, 0)
	assert.NoError(testToken.CheckCreateToken(cfg))
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
	data, err = testToken.Marshal()
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
	assert.NoError(testToken.CheckOpenLending(lcfg))
	assert.NoError(testToken.OpenLending(lcfg))
	assert.NoError(testToken.CheckCloseLending())
	assert.NotNil(testToken.ld.Lending)
	assert.NotNil(testToken.ld.LendingLedger)

	// Destroy
	assert.ErrorContains(testToken.CheckDestroyToken(acc), "some token in the use")
	assert.NoError(testToken.Borrow(token, acc.id, big.NewInt(1000), 0))
	assert.ErrorContains(testToken.CheckDestroyToken(acc),
		"Account(0x0000000000000000000000000000002454455354).CheckDestroyToken error: please repay all before close")
	assert.ErrorContains(testToken.DestroyToken(acc),
		"Account(0x0000000000000000000000000000002454455354).DestroyToken error: please repay all before close")
	actual, err := testToken.Repay(token, acc.id, big.NewInt(1000))
	assert.NoError(err)
	assert.Equal(uint64(1000), actual.Uint64())

	assert.ErrorContains(testToken.CheckDestroyToken(acc), "some token in the use")
	assert.ErrorContains(testToken.DestroyToken(acc), "some token in the use")
	testToken.Add(token, big.NewInt(2000))
	assert.NoError(testToken.CheckDestroyToken(acc))
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
	assert.Nil(testToken.ld.LendingLedger)
	assert.Equal(uint64(1100), acc.balanceOf(constants.NativeToken).Uint64())

	// Destroy again
	assert.ErrorContains(testToken.CheckDestroyToken(acc), "invalid token account")
	assert.ErrorContains(testToken.DestroyToken(acc), "invalid token account")

	// Marshal again
	data, err = testToken.Marshal()
	assert.NoError(err)
	acc2, err = ParseAccount(testToken.id, data)
	assert.NoError(err)
	assert.Equal(testToken.ld.Bytes(), acc2.ld.Bytes())

	// Create again
	assert.NoError(testToken.CheckCreateToken(cfg))
	assert.NoError(testToken.CreateToken(cfg))
	assert.Equal(false, testToken.valid(ld.TokenAccount))
	testToken.Add(constants.NativeToken, big.NewInt(100))
	assert.Equal(true, testToken.valid(ld.TokenAccount))
}
