// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package chain

import (
	"fmt"
	"math"
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ldclabs/ldvm/constants"
	"github.com/ldclabs/ldvm/ld"
	"github.com/ldclabs/ldvm/util"
)

func TestAccountCache(t *testing.T) {
	assert := assert.New(t)
	t.Skip("it maybe failed on github CI")

	ac := getAccountCache()
	ptr := fmt.Sprintf("%p", ac)
	putAccountCache(ac)

	assert.Equal(ptr, fmt.Sprintf("%p", getAccountCache()), "should reuse one from pool")
	assert.NotEqual(ptr, fmt.Sprintf("%p", getAccountCache()), "should create one when pool empty")
}

func TestNativeAccount(t *testing.T) {
	assert := assert.New(t)

	token := ld.MustNewToken("TEST")
	acc := NewAccount(util.Signer1.Address())
	acc.Init(big.NewInt(0), 1, 1)

	assert.Equal(ld.NativeAccount, acc.Type())
	assert.Equal(true, acc.isEmpty())
	assert.Equal(true, acc.Valid(ld.NativeAccount))
	assert.Equal(false, acc.Valid(ld.TokenAccount))
	assert.Equal(false, acc.Valid(ld.StakeAccount))
	assert.Equal(uint64(0), acc.Nonce())

	assert.Equal(uint64(0), acc.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(uint64(0), acc.balanceOf(token).Uint64())
	assert.Equal(uint64(0), acc.balanceOfAll(constants.NativeToken).Uint64())
	assert.Equal(uint64(0), acc.balanceOfAll(token).Uint64())

	assert.NoError(acc.checkBalance(constants.NativeToken, big.NewInt(0)))
	assert.NoError(acc.checkBalance(token, big.NewInt(0)))
	assert.ErrorContains(acc.checkBalance(constants.NativeToken, nil), "invalid amount <nil>")
	assert.ErrorContains(acc.checkBalance(token, nil), "invalid amount <nil>")
	assert.ErrorContains(acc.checkBalance(constants.NativeToken, big.NewInt(-1)), "invalid amount -1")
	assert.ErrorContains(acc.checkBalance(token, big.NewInt(-1)), "invalid amount -1")
	assert.ErrorContains(acc.checkBalance(constants.NativeToken, big.NewInt(1)), "insufficient NativeLDC balance")
	assert.ErrorContains(acc.checkBalance(token, big.NewInt(1)), "insufficient TEST balance")

	var ty ld.TxType
	for ; ty <= ld.TypeDeleteData; ty++ {
		assert.NoError(acc.CheckAsFrom(ty))
		assert.NoError(acc.CheckAsTo(ty))
	}

	// UpdateKeepers, SatisfySigning, SatisfySigningPlus
	assert.Equal(uint8(0), acc.Threshold())
	assert.Equal(util.EthIDs{}, acc.Keepers())
	assert.True(acc.SatisfySigning(util.EthIDs{util.Signer1.Address()}))
	assert.True(acc.SatisfySigningPlus(util.EthIDs{util.Signer1.Address()}))
	assert.False(NewAccount(constants.LDCAccount).SatisfySigning(util.EthIDs{constants.LDCAccount}))
	assert.False(NewAccount(constants.LDCAccount).SatisfySigningPlus(util.EthIDs{constants.LDCAccount}))

	assert.NoError(acc.UpdateKeepers(1, util.EthIDs{util.Signer1.Address(), util.Signer2.Address()}, nil, nil))
	assert.True(acc.SatisfySigning(util.EthIDs{util.Signer1.Address()}))
	assert.True(acc.SatisfySigning(util.EthIDs{util.Signer2.Address()}))
	assert.True(acc.SatisfySigning(util.EthIDs{util.Signer1.Address(), util.Signer2.Address()}))
	assert.False(acc.SatisfySigningPlus(util.EthIDs{util.Signer1.Address()}))
	assert.False(acc.SatisfySigningPlus(util.EthIDs{util.Signer2.Address()}))
	assert.True(acc.SatisfySigningPlus(util.EthIDs{util.Signer1.Address(), util.Signer2.Address()}))

	// Add
	assert.ErrorContains(acc.Add(constants.NativeToken, nil), "invalid amount <nil>")
	assert.ErrorContains(acc.Add(token, nil), "invalid amount <nil>")
	assert.ErrorContains(acc.Add(constants.NativeToken, big.NewInt(-1)), "invalid amount -1")
	assert.ErrorContains(acc.Add(token, big.NewInt(-1)), "invalid amount -1")
	assert.NoError(acc.Add(constants.NativeToken, big.NewInt(100)))
	assert.NoError(acc.Add(token, big.NewInt(100)))
	assert.Equal(uint64(100), acc.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(uint64(100), acc.balanceOf(token).Uint64())
	assert.Equal(uint64(100), acc.balanceOfAll(constants.NativeToken).Uint64())
	assert.Equal(uint64(100), acc.balanceOfAll(token).Uint64())
	assert.NoError(acc.Add(constants.NativeToken, big.NewInt(0)))
	assert.NoError(acc.Add(token, big.NewInt(0)))
	assert.Equal(uint64(100), acc.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(uint64(100), acc.balanceOf(token).Uint64())
	assert.Equal(uint64(100), acc.balanceOfAll(constants.NativeToken).Uint64())
	assert.Equal(uint64(100), acc.balanceOfAll(token).Uint64())

	// Sub
	assert.ErrorContains(acc.Sub(constants.NativeToken, nil), "invalid amount <nil>")
	assert.ErrorContains(acc.Sub(token, nil), "invalid amount <nil>")
	assert.ErrorContains(acc.Sub(constants.NativeToken, big.NewInt(-1)), "invalid amount -1")
	assert.ErrorContains(acc.Sub(token, big.NewInt(-1)), "invalid amount -1")
	assert.NoError(acc.Sub(constants.NativeToken, big.NewInt(10)))
	assert.NoError(acc.Sub(token, big.NewInt(10)))
	assert.Equal(uint64(90), acc.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(uint64(90), acc.balanceOf(token).Uint64())
	assert.Equal(uint64(90), acc.balanceOfAll(constants.NativeToken).Uint64())
	assert.Equal(uint64(90), acc.balanceOfAll(token).Uint64())
	assert.NoError(acc.Sub(constants.NativeToken, big.NewInt(0)))
	assert.NoError(acc.Sub(token, big.NewInt(0)))
	assert.Equal(uint64(90), acc.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(uint64(90), acc.balanceOf(token).Uint64())
	assert.Equal(uint64(90), acc.balanceOfAll(constants.NativeToken).Uint64())
	assert.Equal(uint64(90), acc.balanceOfAll(token).Uint64())
	assert.ErrorContains(acc.Sub(constants.NativeToken, big.NewInt(100)),
		"0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC has an insufficient NativeLDC balance, expected 100, got 90")
	assert.ErrorContains(acc.Sub(token, big.NewInt(100)),
		"0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC has an insufficient TEST balance, expected 100, got 90")

	// SubByNonce
	assert.ErrorContains(acc.SubByNonce(token, 1, big.NewInt(10)),
		"invalid nonce for 0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC, expected 0, got 1")
	assert.NoError(acc.SubByNonce(constants.NativeToken, 0, big.NewInt(10)))
	assert.NoError(acc.SubByNonce(token, 1, big.NewInt(10)))
	assert.Equal(uint64(80), acc.balanceOfAll(constants.NativeToken).Uint64())
	assert.Equal(uint64(80), acc.balanceOfAll(token).Uint64())
	assert.ErrorContains(acc.SubByNonce(constants.NativeToken, 2, big.NewInt(100)),
		"0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC has an insufficient NativeLDC balance, expected 100, got 80")
	assert.ErrorContains(acc.SubByNonce(token, 2, big.NewInt(100)),
		"0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC has an insufficient TEST balance, expected 100, got 80")

	// NonceTable
	assert.ErrorContains(acc.CheckSubByNonceTable(constants.NativeToken, 12345, 1000, big.NewInt(10)),
		"Account.SubByNonceTable failed: nonce 1000 not exists at 12345 on 0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC")
	assert.ErrorContains(acc.SubByNonceTable(token, 12345, 1000, big.NewInt(10)),
		"Account.SubByNonceTable failed: nonce 1000 not exists at 12345 on 0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC")

	assert.NoError(acc.CheckNonceTable(12345, []uint64{1, 2, 3, 4, 0}))
	assert.NoError(acc.AddNonceTable(12345, []uint64{1, 2, 3, 4, 0}))
	assert.ErrorContains(acc.CheckNonceTable(12345, []uint64{0, 10}),
		"Account.CheckNonceTable failed: nonce 0 exists at 12345 on 0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC")
	assert.ErrorContains(acc.AddNonceTable(12345, []uint64{2, 10}),
		"Account.CheckNonceTable failed: nonce 2 exists at 12345 on 0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC")
	assert.NoError(acc.CheckSubByNonceTable(constants.NativeToken, 12345, 0, big.NewInt(10)))
	assert.NoError(acc.SubByNonceTable(constants.NativeToken, 12345, 0, big.NewInt(10)))

	assert.ErrorContains(acc.CheckSubByNonceTable(token, 12345, 0, big.NewInt(10)),
		"Account.SubByNonceTable failed: nonce 0 not exists at 12345 on 0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC")
	assert.ErrorContains(acc.SubByNonceTable(token, 12345, 0, big.NewInt(10)),
		"Account.SubByNonceTable failed: nonce 0 not exists at 12345 on 0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC")
	assert.ErrorContains(acc.CheckSubByNonceTable(token, 123456, 2, big.NewInt(10)),
		"Account.SubByNonceTable failed: nonce 2 not exists at 123456 on 0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC")
	assert.ErrorContains(acc.SubByNonceTable(token, 123456, 2, big.NewInt(10)),
		"Account.SubByNonceTable failed: nonce 2 not exists at 123456 on 0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC")
	assert.NoError(acc.CheckSubByNonceTable(token, 12345, 2, big.NewInt(10)))
	assert.NoError(acc.SubByNonceTable(token, 12345, 2, big.NewInt(10)))
	assert.Equal(uint64(70), acc.balanceOfAll(constants.NativeToken).Uint64())
	assert.Equal(uint64(70), acc.balanceOfAll(token).Uint64())

	assert.NoError(acc.CheckNonceTable(12345, []uint64{0}))
	assert.NoError(acc.AddNonceTable(12345, []uint64{0}))
	assert.Equal([]uint64{0, 1, 3, 4}, acc.ld.NonceTable[12345])
	assert.NoError(acc.SubByNonceTable(constants.NativeToken, 12345, 1, big.NewInt(10)))
	assert.NoError(acc.SubByNonceTable(constants.NativeToken, 12345, 3, big.NewInt(10)))
	assert.NoError(acc.SubByNonceTable(token, 12345, 0, big.NewInt(10)))
	assert.NoError(acc.SubByNonceTable(token, 12345, 4, big.NewInt(10)))
	assert.Equal(uint64(50), acc.balanceOfAll(constants.NativeToken).Uint64())
	assert.Equal(uint64(50), acc.balanceOfAll(token).Uint64())
	assert.Equal(0, len(acc.ld.NonceTable))

	for i := uint64(0); i < 64; i++ {
		assert.NoError(acc.CheckNonceTable(i, []uint64{i}))
		assert.NoError(acc.AddNonceTable(i, []uint64{i}))
	}
	assert.ErrorContains(acc.CheckNonceTable(100, []uint64{100}),
		"Account.CheckNonceTable failed: 0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC has too many NonceTable groups, expected <= 64")
	assert.ErrorContains(acc.AddNonceTable(100, []uint64{100}),
		"Account.CheckNonceTable failed: 0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC has too many NonceTable groups, expected <= 64")

	// Marshal
	data, err := acc.Marshal()
	assert.NoError(err)
	acc2, err := ParseAccount(acc.id, data)
	assert.NoError(err)
	assert.Equal(acc.ld.Bytes(), acc2.ld.Bytes())

	// Lending
	cfg := &ld.LendingConfig{
		DailyInterest:   10,
		OverdueInterest: 10,
		MinAmount:       big.NewInt(1000),
		MaxAmount:       big.NewInt(1_000_000),
	}
	assert.NoError(acc.CheckOpenLending(cfg))
	assert.NoError(acc.OpenLending(cfg))
	assert.NoError(acc.CheckCloseLending())
	assert.NoError(acc.CloseLending())
}

func TestTokenAccount(t *testing.T) {
	assert := assert.New(t)

	acc := NewAccount(util.Signer1.Address())
	acc.Init(big.NewInt(0), 0, 0)
	amount := big.NewInt(1000)
	cfg := &ld.TxAccounter{
		Threshold: 1,
		Keepers:   util.EthIDs{util.Signer1.Address(), util.Signer2.Address()},
		Amount:    amount,
	}
	assert.ErrorContains(acc.CheckCreateToken(cfg),
		"Account.CheckCreateToken failed: invalid token 0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC")
	assert.ErrorContains(acc.CreateToken(cfg),
		"Account.CheckCreateToken failed: invalid token 0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC")
	assert.ErrorContains(acc.CheckDestroyToken(acc),
		"Account.CheckDestroyToken failed: invalid token account 0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC")
	assert.ErrorContains(acc.DestroyToken(acc),
		"Account.CheckDestroyToken failed: invalid token account 0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC")

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
	assert.Equal(uint8(0), nativeToken.Threshold())
	assert.Equal(util.EthIDs{}, nativeToken.Keepers())
	assert.False(nativeToken.SatisfySigning(util.EthIDs{}), "no controller")
	assert.False(nativeToken.SatisfySigning(util.EthIDs{util.Signer1.Address()}), "no controller")
	assert.False(nativeToken.SatisfySigningPlus(util.EthIDs{}), "no controller")
	assert.False(nativeToken.SatisfySigningPlus(util.EthIDs{util.Signer1.Address()}), "no controller")
	assert.ErrorContains(nativeToken.CheckDestroyToken(acc),
		"invalid token")
	assert.ErrorContains(nativeToken.CheckDestroyToken(acc),
		"invalid token")

	nativeToken.Sub(constants.NativeToken, big.NewInt(100))
	acc.Add(constants.NativeToken, big.NewInt(100))
	assert.Equal(big.NewInt(900).Uint64(), nativeToken.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(big.NewInt(900).Uint64(), nativeToken.balanceOfAll(constants.NativeToken).Uint64())
	assert.Equal(big.NewInt(100).Uint64(), acc.balanceOf(constants.NativeToken).Uint64())

	// Marshal
	data, err := nativeToken.Marshal()
	assert.NoError(err)
	acc2, err := ParseAccount(nativeToken.id, data)
	assert.NoError(err)
	assert.Equal(nativeToken.ld.Bytes(), acc2.ld.Bytes())

	token := ld.MustNewToken("TEST")
	testToken := NewAccount(util.EthID(token))
	testToken.Init(big.NewInt(100), 0, 0)
	assert.NoError(testToken.CheckCreateToken(cfg))
	assert.NoError(testToken.CreateToken(cfg))
	assert.Equal(false, testToken.valid(ld.TokenAccount))
	testToken.Add(constants.NativeToken, big.NewInt(100))
	assert.Equal(true, testToken.valid(ld.TokenAccount))

	assert.Equal(big.NewInt(0).Uint64(), testToken.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(big.NewInt(100).Uint64(), testToken.balanceOfAll(constants.NativeToken).Uint64())
	assert.Equal(amount.Uint64(), testToken.balanceOf(token).Uint64())
	assert.Equal(amount.Uint64(), testToken.balanceOfAll(token).Uint64())
	assert.Equal(amount.Uint64(), testToken.ld.MaxTotalSupply.Uint64())
	assert.Equal(uint8(1), testToken.Threshold())
	assert.Equal(util.EthIDs{util.Signer1.Address(), util.Signer2.Address()}, testToken.Keepers())
	assert.False(testToken.SatisfySigning(util.EthIDs{}))
	assert.True(testToken.SatisfySigning(util.EthIDs{util.Signer1.Address()}))
	assert.True(testToken.SatisfySigning(util.EthIDs{util.Signer2.Address()}))
	assert.False(testToken.SatisfySigningPlus(util.EthIDs{}))
	assert.False(testToken.SatisfySigningPlus(util.EthIDs{util.Signer1.Address()}))
	assert.True(testToken.SatisfySigningPlus(util.EthIDs{util.Signer1.Address(), util.Signer2.Address()}))

	testToken.Sub(token, big.NewInt(100))
	acc.Add(token, big.NewInt(100))
	assert.Equal(big.NewInt(900).Uint64(), testToken.balanceOf(token).Uint64())
	assert.Equal(big.NewInt(900).Uint64(), testToken.balanceOfAll(token).Uint64())
	assert.Equal(big.NewInt(100).Uint64(), acc.balanceOf(token).Uint64())
	testToken.Sub(token, big.NewInt(100))
	acc.Add(token, big.NewInt(100))
	assert.Equal(big.NewInt(800).Uint64(), testToken.balanceOf(token).Uint64())
	assert.Equal(big.NewInt(200).Uint64(), acc.balanceOf(token).Uint64())

	// Marshal
	data, err = testToken.Marshal()
	assert.NoError(err)
	acc2, err = ParseAccount(testToken.id, data)
	assert.NoError(err)
	assert.Equal(testToken.ld.Bytes(), acc2.ld.Bytes())

	// Lending
	lcfg := &ld.LendingConfig{
		DailyInterest:   10,
		OverdueInterest: 10,
		MinAmount:       big.NewInt(1000),
		MaxAmount:       big.NewInt(1_000_000),
	}
	assert.NoError(testToken.CheckOpenLending(lcfg))
	assert.NoError(testToken.OpenLending(lcfg))
	assert.NoError(testToken.CheckCloseLending())
	assert.NoError(testToken.CloseLending())

	// Destroy
	assert.ErrorContains(testToken.CheckDestroyToken(acc), "some token in the use")
	assert.ErrorContains(testToken.DestroyToken(acc), "some token in the use")
	testToken.Add(token, big.NewInt(100))
	testToken.Add(token, big.NewInt(100))
	assert.NoError(testToken.CheckDestroyToken(acc))
	assert.NoError(testToken.DestroyToken(acc))
	assert.Equal(uint64(0), testToken.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(uint64(0), testToken.balanceOfAll(constants.NativeToken).Uint64())
	assert.Equal(uint64(0), testToken.balanceOf(token).Uint64())
	assert.Equal(uint64(0), testToken.balanceOfAll(token).Uint64())
	assert.Equal(uint8(0), testToken.Threshold())
	assert.Equal(util.EthIDs{}, testToken.Keepers())
	assert.Equal(0, len(testToken.ld.Tokens))
	assert.Nil(testToken.ld.MaxTotalSupply)
	assert.Equal(big.NewInt(200).Uint64(), acc.balanceOf(constants.NativeToken).Uint64())

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

func TestStakeAccount(t *testing.T) {
	assert := assert.New(t)

	acc := NewAccount(util.Signer1.Address())
	acc.Init(big.NewInt(0), 0, 0)
	acc2 := NewAccount(util.Signer2.Address())
	acc2.Init(big.NewInt(0), 0, 0)
	pledge := big.NewInt(1000)
	cfg := &ld.TxAccounter{
		Threshold: 1,
		Keepers:   util.EthIDs{util.Signer1.Address(), util.Signer2.Address()},
	}
	scfg := &ld.StakeConfig{
		LockTime:    2,
		WithdrawFee: 100_000,
		MinAmount:   big.NewInt(100),
		MaxAmount:   big.NewInt(1000),
	}
	assert.ErrorContains(acc.CheckCreateStake(util.Signer1.Address(), pledge, cfg, scfg),
		"Account.CheckCreateStake failed: invalid stake account 0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC")
	assert.ErrorContains(acc.CreateStake(util.Signer1.Address(), pledge, cfg, scfg),
		"Account.CheckCreateStake failed: invalid stake account 0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC")
	assert.ErrorContains(acc.CheckResetStake(scfg),
		"ccount.CheckResetStake failed: invalid stake account 0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC")
	assert.ErrorContains(acc.ResetStake(scfg),
		"ccount.CheckResetStake failed: invalid stake account 0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC")
	assert.ErrorContains(acc.CheckResetStake(scfg),
		"ccount.CheckResetStake failed: invalid stake account 0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC")
	assert.ErrorContains(acc.ResetStake(scfg),
		"ccount.CheckResetStake failed: invalid stake account 0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC")

	stake := ld.MustNewStake("@TEST")
	testStake := NewAccount(util.EthID(stake))
	testStake.Init(big.NewInt(100), 1, 1)
	assert.NoError(testStake.CheckCreateStake(util.Signer1.Address(), pledge, cfg, scfg))
	assert.NoError(testStake.CreateStake(util.Signer1.Address(), pledge, cfg, scfg))
	assert.Equal(false, testStake.valid(ld.StakeAccount))
	testStake.Add(constants.NativeToken, big.NewInt(1000))
	assert.Equal(true, testStake.valid(ld.StakeAccount))

	assert.Equal(big.NewInt(900).Uint64(), testStake.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(big.NewInt(1000).Uint64(), testStake.balanceOfAll(constants.NativeToken).Uint64())
	assert.Nil(testStake.ld.MaxTotalSupply)
	assert.NotNil(testStake.ld.StakeLedger)
	assert.Equal(uint8(1), testStake.Threshold())
	assert.Equal(util.EthIDs{util.Signer1.Address(), util.Signer2.Address()}, testStake.Keepers())
	assert.False(testStake.SatisfySigning(util.EthIDs{}))
	assert.True(testStake.SatisfySigning(util.EthIDs{util.Signer1.Address()}))
	assert.True(testStake.SatisfySigning(util.EthIDs{util.Signer2.Address()}))
	assert.False(testStake.SatisfySigningPlus(util.EthIDs{}))
	assert.False(testStake.SatisfySigningPlus(util.EthIDs{util.Signer1.Address()}))
	assert.True(testStake.SatisfySigningPlus(util.EthIDs{util.Signer1.Address(), util.Signer2.Address()}))

	assert.Equal(1, len(testStake.ld.StakeLedger))
	assert.Equal(pledge.Uint64(), testStake.ld.StakeLedger[util.Signer1.Address()].Amount.Uint64())
	assert.NoError(testStake.TakeStake(constants.NativeToken, util.Signer2.Address(), big.NewInt(1000), 0))
	testStake.Add(constants.NativeToken, big.NewInt(1000))
	assert.Equal(2, len(testStake.ld.StakeLedger))
	assert.Equal(uint64(1000), testStake.ld.StakeLedger[util.Signer2.Address()].Amount.Uint64())
	assert.Equal(big.NewInt(1900).Uint64(), testStake.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(big.NewInt(2000).Uint64(), testStake.balanceOfAll(constants.NativeToken).Uint64())

	// Marshal
	data, err := testStake.Marshal()
	assert.NoError(err)
	testStake2, err := ParseAccount(testStake.id, data)
	assert.NoError(err)
	assert.Equal(testStake.ld.Bytes(), testStake2.ld.Bytes())

	// Lending
	lcfg := &ld.LendingConfig{
		DailyInterest:   10,
		OverdueInterest: 10,
		MinAmount:       big.NewInt(1000),
		MaxAmount:       big.NewInt(1_000_000),
	}
	assert.NoError(testStake.CheckOpenLending(lcfg))
	assert.NoError(testStake.OpenLending(lcfg))
	assert.NoError(testStake.CheckCloseLending())
	assert.NoError(testStake.CloseLending())

	// Reset & Destroy
	token := ld.MustNewToken("TEST")
	assert.ErrorContains(testStake.CheckResetStake(&ld.StakeConfig{
		Type:        1,
		WithdrawFee: 10_000,
		MinAmount:   big.NewInt(1000),
		MaxAmount:   big.NewInt(10000),
	}), "Account.CheckResetStake failed: can't change stake type")
	assert.ErrorContains(testStake.CheckResetStake(&ld.StakeConfig{
		Token:       token,
		WithdrawFee: 10_000,
		MinAmount:   big.NewInt(1000),
		MaxAmount:   big.NewInt(10000),
	}), "Account.CheckResetStake failed: can't change stake token")
	assert.ErrorContains(testStake.CheckResetStake(&ld.StakeConfig{
		WithdrawFee: 10_000,
		MinAmount:   big.NewInt(1000),
		MaxAmount:   big.NewInt(10000),
	}), "Account.CheckResetStake failed: stake in lock, please retry after lockTime")
	assert.ErrorContains(testStake.CheckDestroyStake(acc),
		"Account.CheckDestroyStake failed: stake in lock, please retry after lockTime")
	testStake.ld.Timestamp = 10
	assert.ErrorContains(testStake.CheckResetStake(&ld.StakeConfig{
		WithdrawFee: 10_000,
		MinAmount:   big.NewInt(1000),
		MaxAmount:   big.NewInt(10000),
	}), "Account.CheckResetStake failed: stake holders should not more than 1")
	assert.ErrorContains(testStake.CheckDestroyStake(acc),
		"Account.CheckDestroyStake failed: stake ledger not empty, please withdraw all except recipient")
	delete(testStake.ld.StakeLedger, util.Signer2.Address())
	assert.NoError(testStake.CheckResetStake(&ld.StakeConfig{
		WithdrawFee: 10_000,
		MinAmount:   big.NewInt(1000),
		MaxAmount:   big.NewInt(10000),
	}))
	assert.Equal(uint64(100_000), testStake.ld.Stake.WithdrawFee)
	assert.NoError(testStake.ResetStake(&ld.StakeConfig{
		WithdrawFee: 10_000,
		MinAmount:   big.NewInt(1000),
		MaxAmount:   big.NewInt(10000),
	}))
	assert.Equal(uint64(10_000), testStake.ld.Stake.WithdrawFee)
	assert.Equal(uint64(1000), testStake.ld.Stake.MinAmount.Uint64())
	assert.Equal(uint64(10000), testStake.ld.Stake.MaxAmount.Uint64())

	// Destroy
	assert.ErrorContains(testStake.CheckDestroyStake(acc2),
		"Account.CheckDestroyStake failed: recipient not exists")
	assert.NoError(testStake.CheckDestroyStake(acc))
	assert.NotNil(testStake.ld.Stake)
	assert.NoError(testStake.DestroyStake(acc))
	assert.Equal(uint64(0), testStake.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(uint64(0), testStake.balanceOfAll(constants.NativeToken).Uint64())
	assert.Equal(uint8(0), testStake.Threshold())
	assert.Equal(util.EthIDs{}, testStake.Keepers())
	assert.Nil(testStake.ld.Stake)
	assert.Nil(testStake.ld.StakeLedger)
	assert.Equal(0, len(testStake.ld.Tokens))
	assert.Equal(big.NewInt(2000).Uint64(), acc.balanceOf(constants.NativeToken).Uint64())

	// Destroy again
	assert.ErrorContains(testStake.CheckDestroyStake(acc),
		"Account.CheckDestroyStake failed: invalid stake account 0x0000000000000000000000000000004054455354")
	assert.ErrorContains(testStake.CheckResetStake(&ld.StakeConfig{
		WithdrawFee: 10_000,
		MinAmount:   big.NewInt(1000),
		MaxAmount:   big.NewInt(10000),
	}), "Account.CheckResetStake failed: invalid stake account 0x0000000000000000000000000000004054455354")

	// Marshal again
	data, err = testStake.Marshal()
	assert.NoError(err)
	testStake2, err = ParseAccount(testStake.id, data)
	assert.NoError(err)
	assert.Equal(testStake.ld.Bytes(), testStake2.ld.Bytes())

	// Create again
	assert.NoError(testStake.CheckCreateStake(util.Signer1.Address(), pledge, cfg, scfg))
	assert.NoError(testStake.CreateStake(util.Signer1.Address(), pledge, cfg, scfg))
}

func TestTakeStakeAndWithdraw(t *testing.T) {
	assert := assert.New(t)

	addr0 := util.NewSigner().Address()
	addr1 := util.NewSigner().Address()
	addr2 := util.NewSigner().Address()
	addr3 := util.NewSigner().Address()
	approver := util.NewSigner().Address()
	sk := NewAccount(util.Signer1.Address()).Init(big.NewInt(0), 10, 100)
	acc0 := NewAccount(addr0).Init(big.NewInt(0), 10, 100)

	ldc := new(big.Int).SetUint64(constants.LDC)
	ldcf := float64(constants.LDC)
	pledge := new(big.Int).SetUint64(constants.LDC * 10)
	token := ld.MustNewToken("LDC")
	withdrawFee := uint64(100_000)
	sa := NewAccount(util.EthID(ld.MustNewStake("@LDC"))).Init(pledge, 1, 1)
	assert.NoError(sa.CreateStake(util.Signer1.Address(), pledge, &ld.TxAccounter{
		Threshold: 1,
		Keepers:   util.EthIDs{util.Signer1.Address(), util.Signer2.Address()},
	}, &ld.StakeConfig{
		LockTime:    10,
		WithdrawFee: withdrawFee,
		MinAmount:   new(big.Int).SetUint64(constants.LDC),
		MaxAmount:   pledge,
	}))
	sa.Add(constants.NativeToken, pledge)
	assert.True(sa.valid(ld.StakeAccount))
	assert.Equal(uint64(0), sa.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(constants.LDC*10, sa.balanceOfAll(constants.NativeToken).Uint64())
	assert.Equal(constants.LDC*10, sa.ld.StakeLedger[util.Signer1.Address()].Amount.Uint64())

	// Invalid TakeStake args
	assert.ErrorContains(sk.CheckTakeStake(constants.NativeToken, addr1, big.NewInt(1000), 0),
		"Account.CheckTakeStake failed: invalid stake account")
	assert.ErrorContains(sa.CheckTakeStake(token, addr2, big.NewInt(1000), 0),
		"Account.CheckTakeStake failed: invalid token, expected NativeLDC, got LDC")
	assert.ErrorContains(sa.CheckTakeStake(constants.NativeToken, addr3, big.NewInt(1000), 0),
		"Account.CheckTakeStake failed: invalid amount, expected >= 1000000000, got 1000")

	sa.Add(constants.NativeToken, ldc)
	assert.Equal(constants.LDC, sa.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(constants.LDC*11, sa.balanceOfAll(constants.NativeToken).Uint64())
	assert.Equal(constants.LDC*10, sa.ld.StakeLedger[util.Signer1.Address()].Amount.Uint64())

	assert.NoError(sa.CheckTakeStake(constants.NativeToken, addr0, ldc, 0))
	assert.NoError(sa.TakeStake(constants.NativeToken, addr0, ldc, 0))
	sa.Add(constants.NativeToken, ldc)
	assert.Equal(constants.LDC*2, sa.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(constants.LDC*12, sa.balanceOfAll(constants.NativeToken).Uint64())
	assert.Equal(constants.LDC*11, sa.ld.StakeLedger[util.Signer1.Address()].Amount.Uint64())
	assert.Equal(constants.LDC, sa.ld.StakeLedger[addr0].Amount.Uint64())

	sa.Add(constants.NativeToken, ldc)
	assert.Equal(constants.LDC*3, sa.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(constants.LDC*13, sa.balanceOfAll(constants.NativeToken).Uint64())
	assert.Equal(constants.LDC*11, sa.ld.StakeLedger[util.Signer1.Address()].Amount.Uint64())
	assert.Equal(constants.LDC, sa.ld.StakeLedger[addr0].Amount.Uint64())

	assert.NoError(sa.CheckTakeStake(constants.NativeToken, addr1, pledge, 0))
	assert.NoError(sa.TakeStake(constants.NativeToken, addr1, pledge, 0))
	sa.Add(constants.NativeToken, pledge)
	assert.Equal(constants.LDC*13, sa.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(constants.LDC*23, sa.balanceOfAll(constants.NativeToken).Uint64())
	assert.Equal(uint64(ldcf*(11+float64(11)/12)), sa.ld.StakeLedger[util.Signer1.Address()].Amount.Uint64())
	assert.Equal(uint64(ldcf*(1+float64(1)/12)), sa.ld.StakeLedger[addr0].Amount.Uint64())
	assert.Equal(constants.LDC*10, sa.ld.StakeLedger[addr1].Amount.Uint64())

	assert.ErrorContains(sa.CheckTakeStake(constants.NativeToken, addr1, ldc, 0),
		"Account.CheckTakeStake failed: invalid total amount, expected <= 10000000000, got 11000000000")
	// No Bonus
	assert.NoError(sa.CheckTakeStake(constants.NativeToken, addr2, ldc, 11))
	assert.NoError(sa.TakeStake(constants.NativeToken, addr2, ldc, 11))
	sa.Add(constants.NativeToken, ldc)
	assert.Equal(constants.LDC*14, sa.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(constants.LDC*24, sa.balanceOfAll(constants.NativeToken).Uint64())
	assert.Equal(uint64(ldcf*(11+float64(11)/12)), sa.ld.StakeLedger[util.Signer1.Address()].Amount.Uint64())
	assert.Equal(uint64(ldcf*(1+float64(1)/12)), sa.ld.StakeLedger[addr0].Amount.Uint64())
	assert.Equal(constants.LDC*10, sa.ld.StakeLedger[addr1].Amount.Uint64())
	assert.Equal(constants.LDC, sa.ld.StakeLedger[addr2].Amount.Uint64())

	// Marshal
	data, err := sa.Marshal()
	assert.NoError(err)
	sa2, err := ParseAccount(sa.id, data)
	assert.NoError(err)
	assert.Equal(sa.ld.Bytes(), sa2.ld.Bytes())

	// check WithdrawStake
	assert.ErrorContains(
		sk.CheckWithdrawStake(constants.NativeToken, addr1, util.EthIDs{}, big.NewInt(1000)),
		"Account.CheckWithdrawStake failed: invalid stake account")
	assert.ErrorContains(
		sa.CheckWithdrawStake(token, addr0, util.EthIDs{}, big.NewInt(1000)),
		"Account.CheckWithdrawStake failed: invalid token, expected NativeLDC, got LDC")
	assert.ErrorContains(
		sa.CheckWithdrawStake(constants.NativeToken, addr0, util.EthIDs{}, big.NewInt(1000)),
		"Account.CheckWithdrawStake failed: stake in lock, please retry after lockTime")
	sa.ld.Timestamp = 10
	assert.NoError(
		sa.CheckWithdrawStake(constants.NativeToken, addr0, util.EthIDs{}, big.NewInt(1000)))
	assert.ErrorContains(
		sa.CheckWithdrawStake(constants.NativeToken, addr3, util.EthIDs{}, big.NewInt(1000)),
		"has no stake to withdraw")
	assert.ErrorContains(
		sa.CheckWithdrawStake(constants.NativeToken, addr2, util.EthIDs{}, big.NewInt(1000)),
		"Account.CheckWithdrawStake failed: stake in lock, please retry after lockTime")
	sa.ld.Timestamp = 11
	assert.NoError(
		sa.CheckWithdrawStake(constants.NativeToken, addr2, util.EthIDs{}, big.NewInt(1000)))

	// check UpdateStakeApprover
	assert.ErrorContains(
		sk.CheckUpdateStakeApprover(addr1, approver, util.EthIDs{}),
		"Account.CheckUpdateStakeApprover failed: invalid stake account")
	assert.ErrorContains(
		sa.CheckUpdateStakeApprover(addr3, approver, util.EthIDs{}),
		"has no stake ledger to update")
	assert.NoError(
		sa.CheckUpdateStakeApprover(addr0, approver, util.EthIDs{}))
	assert.Nil(sa.ld.StakeLedger[addr0].Approver)
	assert.NoError(sa.UpdateStakeApprover(addr0, approver, util.EthIDs{}))
	assert.NotNil(sa.ld.StakeLedger[addr0].Approver)
	assert.Equal(approver, *sa.ld.StakeLedger[addr0].Approver)
	assert.ErrorContains(
		sa.CheckUpdateStakeApprover(addr0, util.EthIDEmpty, util.EthIDs{}),
		"need approver signing")
	assert.ErrorContains(
		sa.CheckUpdateStakeApprover(addr0, util.EthIDEmpty, util.EthIDs{addr0}),
		"need approver signing")
	assert.NoError(sa.CheckUpdateStakeApprover(addr0, util.EthIDEmpty, util.EthIDs{approver}))
	assert.NoError(sa.UpdateStakeApprover(addr0, util.EthIDEmpty, util.EthIDs{approver}))
	assert.Nil(sa.ld.StakeLedger[addr0].Approver)

	// continue to check WithdrawStake
	assert.NoError(sa.CheckUpdateStakeApprover(addr2, approver, util.EthIDs{}))
	assert.NoError(sa.UpdateStakeApprover(addr2, approver, util.EthIDs{}))
	assert.ErrorContains(
		sa.CheckWithdrawStake(constants.NativeToken, addr2, util.EthIDs{}, big.NewInt(1000)),
		"need approver signing")
	assert.ErrorContains(
		sa.CheckWithdrawStake(constants.NativeToken, addr2, util.EthIDs{addr2}, big.NewInt(1000)),
		"need approver signing")
	assert.NoError(
		sa.CheckWithdrawStake(constants.NativeToken, addr2, util.EthIDs{approver}, big.NewInt(1000)))
	assert.ErrorContains(
		sa.CheckWithdrawStake(constants.NativeToken, addr2, util.EthIDs{approver},
			new(big.Int).SetUint64(constants.LDC+1)),
		"has an insufficient stake to withdraw, expected 1000000001, got 1000000000")
	assert.NoError(
		sa.CheckWithdrawStake(constants.NativeToken, addr2, util.EthIDs{approver}, ldc))
	am, err := sa.WithdrawStake(constants.NativeToken, addr2, util.EthIDs{approver}, ldc)
	assert.NoError(err)
	sa.Sub(constants.NativeToken, am)
	assert.Equal(constants.LDC-uint64(ldcf*float64(withdrawFee)/1_000_000), am.Uint64(), "withdraw fee")
	assert.NotNil(sa.ld.StakeLedger[addr2])
	assert.Equal(uint64(0), sa.ld.StakeLedger[addr2].Amount.Uint64())

	total := uint64(1088043477)
	am, err = sa.WithdrawStake(constants.NativeToken, addr0, util.EthIDs{}, new(big.Int).SetUint64(total))
	assert.NoError(err)
	sa.Sub(constants.NativeToken, am)
	assert.Equal(total-uint64(float64(total*withdrawFee)/1_000_000), am.Uint64(), "withdraw fee")
	assert.Nil(sa.ld.StakeLedger[addr0])

	total = sa.GetStakeAmount(constants.NativeToken, addr1).Uint64()
	am, err = sa.WithdrawStake(constants.NativeToken, addr1, util.EthIDs{}, new(big.Int).SetUint64(total))
	assert.NoError(err)
	sa.Sub(constants.NativeToken, am)
	assert.Equal(total-uint64(float64(total*withdrawFee)/1_000_000), am.Uint64(), "withdraw fee")
	assert.Nil(sa.ld.StakeLedger[addr1])
	assert.Equal(2, len(sa.ld.StakeLedger))

	total = sa.GetStakeAmount(constants.NativeToken, util.Signer1.Address()).Uint64()
	am, err = sa.WithdrawStake(constants.NativeToken, util.Signer1.Address(),
		util.EthIDs{}, new(big.Int).SetUint64(total))
	assert.ErrorContains(err,
		"Account.CheckWithdrawStake failed: @LDC has an insufficient balance for withdraw")

	// Marshal again
	data, err = sa.Marshal()
	assert.NoError(err)
	sa2, err = ParseAccount(sa.id, data)
	assert.NoError(err)
	assert.Equal(sa.ld.Bytes(), sa2.ld.Bytes())

	// Reset & Destroy
	assert.NoError(sa.CheckResetStake(&ld.StakeConfig{
		WithdrawFee: withdrawFee / 10,
		MinAmount:   new(big.Int).SetUint64(constants.LDC),
		MaxAmount:   pledge,
	}))
	assert.NoError(sa.ResetStake(&ld.StakeConfig{
		WithdrawFee: withdrawFee / 10,
		MinAmount:   new(big.Int).SetUint64(constants.LDC),
		MaxAmount:   pledge,
	}))

	ba := sk.balanceOf(constants.NativeToken).Uint64()
	assert.NoError(sa.CheckDestroyStake(sk))
	assert.NoError(sa.DestroyStake(sk))
	assert.Equal(uint64(0), sa.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(uint64(0), sa.balanceOfAll(constants.NativeToken).Uint64())
	assert.Equal(uint8(0), sa.Threshold())
	assert.Equal(util.EthIDs{}, sa.Keepers())
	assert.Nil(sa.ld.Stake)
	assert.Nil(sa.ld.StakeLedger)
	assert.Equal(0, len(sa.ld.Tokens))
	assert.Equal(total, sk.balanceOf(constants.NativeToken).Uint64()-ba)

	// Marshal again
	data, err = sa.Marshal()
	assert.NoError(err)
	sa2, err = ParseAccount(sa.id, data)
	assert.NoError(err)
	assert.Equal(sa.ld.Bytes(), sa2.ld.Bytes())

	// Create again
	assert.NoError(sa.CreateStake(util.Signer1.Address(), pledge, &ld.TxAccounter{
		Threshold: 1,
		Keepers:   util.EthIDs{util.Signer1.Address(), util.Signer2.Address()},
	}, &ld.StakeConfig{
		Token:       token,
		WithdrawFee: withdrawFee,
		MinAmount:   new(big.Int).SetUint64(constants.LDC),
		MaxAmount:   pledge,
	}))
	assert.False(sa.valid(ld.StakeAccount))
	sa.Add(constants.NativeToken, pledge)
	assert.True(sa.valid(ld.StakeAccount))
	assert.Equal(uint64(0), sa.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(constants.LDC*10, sa.balanceOfAll(constants.NativeToken).Uint64())
	assert.Equal(uint64(0), sa.balanceOf(token).Uint64())
	assert.Equal(uint64(0), sa.balanceOfAll(token).Uint64())
	assert.Equal(0, len(sa.ld.StakeLedger))

	sa.Add(token, ldc)
	assert.Equal(constants.LDC, sa.balanceOf(token).Uint64())
	assert.Equal(constants.LDC, sa.balanceOfAll(token).Uint64())

	assert.Error(sa.CheckTakeStake(constants.NativeToken, addr0, ldc, 0))
	assert.NoError(sa.TakeStake(token, addr0, ldc, 0))
	sa.Add(token, ldc)
	assert.Equal(constants.LDC*2, sa.balanceOf(token).Uint64())
	assert.Equal(constants.LDC*2, sa.balanceOfAll(token).Uint64())
	assert.Equal(1, len(sa.ld.StakeLedger))
	assert.Equal(constants.LDC, sa.ld.StakeLedger[addr0].Amount.Uint64())
	assert.Equal(constants.LDC*2, sa.GetStakeAmount(token, addr0).Uint64())

	assert.NoError(sa.TakeStake(token, addr1, pledge, 0))
	sa.Add(token, pledge)
	assert.Equal(constants.LDC*12, sa.balanceOf(token).Uint64())
	assert.Equal(constants.LDC*12, sa.balanceOfAll(token).Uint64())
	assert.Equal(constants.LDC*2, sa.ld.StakeLedger[addr0].Amount.Uint64())
	assert.Equal(constants.LDC*10, sa.ld.StakeLedger[addr1].Amount.Uint64())

	assert.ErrorContains(sa.CheckTakeStake(token, addr1, ldc, 0),
		"Account.CheckTakeStake failed: invalid total amount, expected <= 10000000000, got 11000000000")

	// sa take a stake in sc
	sc := NewAccount(util.EthID(ld.MustNewStake("@HODLING"))).Init(pledge, 1, 1)
	assert.NoError(sc.CreateStake(util.Signer2.Address(), pledge, &ld.TxAccounter{
		Threshold: 1,
		Keepers:   util.EthIDs{util.Signer2.Address()},
	}, &ld.StakeConfig{
		Token:       token,
		WithdrawFee: withdrawFee,
		MinAmount:   new(big.Int).SetUint64(constants.LDC),
		MaxAmount:   new(big.Int).SetUint64(constants.LDC * 100),
	}))
	sc.Add(constants.NativeToken, pledge)
	assert.True(sc.valid(ld.StakeAccount))
	all := sa.balanceOfAll(token)
	assert.NoError(sc.TakeStake(token, sa.id, all, 0))
	sc.Add(token, all)
	sa.Sub(token, all)
	assert.Equal(constants.LDC*12, sc.balanceOf(token).Uint64())
	assert.Equal(constants.LDC*12, sc.balanceOfAll(token).Uint64())
	assert.Equal(constants.LDC*12, sc.ld.StakeLedger[sa.id].Amount.Uint64())
	assert.Equal(constants.LDC*0, sa.balanceOf(token).Uint64())
	assert.Equal(constants.LDC*0, sa.balanceOfAll(token).Uint64())

	assert.NoError(sa.TakeStake(token, addr2, ldc, 11))
	sa.Add(token, ldc)
	assert.Equal(constants.LDC, sa.balanceOf(token).Uint64())
	assert.Equal(constants.LDC, sa.balanceOfAll(token).Uint64())
	assert.Equal(constants.LDC*2, sa.ld.StakeLedger[addr0].Amount.Uint64())
	assert.Equal(constants.LDC*10, sa.ld.StakeLedger[addr1].Amount.Uint64())
	assert.Equal(constants.LDC, sa.ld.StakeLedger[addr2].Amount.Uint64())

	assert.ErrorContains(sa.CheckWithdrawStake(token, addr1,
		util.EthIDs{}, pledge),
		"Account.CheckWithdrawStake failed: @LDC has an insufficient balance for withdraw")

	// Marshal again
	data, err = sa.Marshal()
	assert.NoError(err)
	sa2, err = ParseAccount(sa.id, data)
	assert.NoError(err)
	assert.Equal(sa.ld.Bytes(), sa2.ld.Bytes())

	data, err = sc.Marshal()
	assert.NoError(err)
	sc2, err := ParseAccount(sc.id, data)
	assert.NoError(err)
	assert.Equal(sc.ld.Bytes(), sc2.ld.Bytes())

	am, err = sa.WithdrawStake(token, addr2, util.EthIDs{}, ldc)
	assert.NoError(err)
	sa.Sub(token, am)
	fee := uint64(float64(constants.LDC*withdrawFee) / 1_000_000)
	assert.Equal(constants.LDC-fee, am.Uint64(), "withdraw fee")
	assert.Nil(sa.ld.StakeLedger[addr2])
	assert.Equal(2, len(sa.ld.StakeLedger))
	assert.Equal(fee, sa.balanceOfAll(token).Uint64())

	// Destroy sc
	sc.Add(token, pledge)
	assert.NoError(sc.DestroyStake(sa))
	assert.Equal(uint64(0), sc.balanceOf(token).Uint64())
	assert.Equal(uint64(0), sc.balanceOfAll(token).Uint64())
	assert.Equal(uint8(0), sc.Threshold())
	assert.Equal(util.EthIDs{}, sc.Keepers())
	assert.Nil(sc.ld.Stake)
	assert.Nil(sc.ld.StakeLedger)
	assert.Equal(1, len(sc.ld.Tokens))
	assert.Equal(constants.LDC*22+fee, sa.balanceOf(token).Uint64())
	// Marshal again
	data, err = sc.Marshal()
	assert.NoError(err)
	sc2, err = ParseAccount(sc.id, data)
	assert.NoError(err)
	assert.Equal(sc.ld.Bytes(), sc2.ld.Bytes())

	// Destroy sa
	total = sa.GetStakeAmount(token, addr1).Uint64()
	am, err = sa.WithdrawStake(token, addr1, util.EthIDs{}, new(big.Int).SetUint64(total))
	assert.NoError(err)
	sa.Sub(token, am)
	assert.Equal(total-uint64(float64(total*withdrawFee)/1_000_000), am.Uint64(), "withdraw fee")
	assert.Equal(1, am.Cmp(pledge))
	assert.Equal(1, len(sa.ld.StakeLedger))
	assert.Nil(sa.ld.StakeLedger[addr1])

	assert.Equal(uint64(0), acc0.balanceOf(token).Uint64())
	assert.NoError(sa.DestroyStake(acc0))
	assert.Equal(uint64(0), sa.balanceOf(token).Uint64())
	assert.Equal(uint64(0), sa.balanceOfAll(token).Uint64())
	assert.Equal(uint8(0), sa.Threshold())
	assert.Equal(util.EthIDs{}, sa.Keepers())
	assert.Nil(sa.ld.Stake)
	assert.Nil(sa.ld.StakeLedger)
	assert.Equal(1, len(sa.ld.Tokens))
	assert.Equal(1, acc0.balanceOf(token).Cmp(ldc))
	data, err = sa.Marshal()
	assert.NoError(err)
	sa2, err = ParseAccount(sa.id, data)
	assert.NoError(err)
	assert.Equal(sa.ld.Bytes(), sa2.ld.Bytes())
}

func TestLending(t *testing.T) {
	assert := assert.New(t)

	addr0 := util.NewSigner().Address()
	na := NewAccount(util.Signer1.Address()).Init(big.NewInt(0), 10, 100)

	// Lending
	ldc := new(big.Int).SetUint64(constants.LDC)
	token := ld.MustNewToken("LDC")
	lcfg := &ld.LendingConfig{
		DailyInterest:   10_000,
		OverdueInterest: 10_000,
		MinAmount:       new(big.Int).SetUint64(constants.LDC),
		MaxAmount:       new(big.Int).SetUint64(constants.LDC * 10),
	}
	assert.ErrorContains(na.CheckCloseLending(),
		"Account.CheckCloseLending failed: invalid lending on 0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC")
	assert.ErrorContains(na.CloseLending(),
		"Account.CheckCloseLending failed: invalid lending on 0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC")
	assert.ErrorContains(na.CheckBorrow(constants.NativeToken, addr0, ldc, 0),
		"Account.CheckBorrow failed: invalid lending on 0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC")
	assert.ErrorContains(na.Borrow(constants.NativeToken, addr0, ldc, 0),
		"Account.CheckBorrow failed: invalid lending on 0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC")
	assert.ErrorContains(na.CheckRepay(constants.NativeToken, addr0, ldc),
		"Account.CheckRepay failed: invalid lending on 0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC")
	assert.NoError(na.CheckOpenLending(lcfg))
	assert.NoError(na.OpenLending(lcfg))
	assert.ErrorContains(na.CheckOpenLending(lcfg),
		"Account.CheckOpenLending failed: lending exists on 0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC")
	assert.ErrorContains(na.OpenLending(lcfg),
		"Account.CheckOpenLending failed: lending exists on 0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC")

	assert.ErrorContains(na.CheckBorrow(token, addr0, ldc, 0),
		"Account.CheckBorrow failed: invalid token, expected NativeLDC, got LDC")
	assert.ErrorContains(na.Borrow(token, addr0, ldc, 0),
		"Account.CheckBorrow failed: invalid token, expected NativeLDC, got LDC")
	assert.ErrorContains(na.CheckRepay(token, addr0, ldc),
		"Account.CheckRepay failed: invalid token, expected NativeLDC, got LDC")
	assert.ErrorContains(na.CheckRepay(constants.NativeToken, addr0, ldc),
		"Account.CheckRepay failed: don't need to repay")

	assert.ErrorContains(na.CheckBorrow(constants.NativeToken, addr0, ldc, 100),
		"Account.CheckBorrow failed: invalid dueTime, expected > 100, got 100")
	assert.ErrorContains(na.CheckBorrow(constants.NativeToken, addr0,
		new(big.Int).SetUint64(constants.LDC-1), 0),
		"Account.CheckBorrow failed: invalid amount, expected >= 1000000000, got 999999999")
	assert.ErrorContains(na.CheckBorrow(constants.NativeToken, addr0, ldc, 0),
		"Account.CheckBorrow failed: 0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC has an insufficient NativeLDC balance, expected 1000000000, got 0")

	na.Add(constants.NativeToken, new(big.Int).SetUint64(constants.LDC*10))
	assert.NoError(na.CheckBorrow(constants.NativeToken, addr0, ldc, daysecs+100))
	assert.Nil(na.ld.LendingLedger[addr0])
	assert.NoError(na.Borrow(constants.NativeToken, addr0, ldc, daysecs+100))
	assert.NotNil(na.ld.LendingLedger[addr0])
	assert.Equal(constants.LDC, na.ld.LendingLedger[addr0].Amount.Uint64())
	assert.Equal(uint64(100), na.ld.LendingLedger[addr0].UpdateAt)
	assert.Equal(uint64(daysecs+100), na.ld.LendingLedger[addr0].DueTime)

	assert.ErrorContains(na.CheckBorrow(constants.NativeToken, addr0,
		new(big.Int).SetUint64(constants.LDC*10), 0),
		"Account.CheckBorrow failed: invalid amount, expected <= 10000000000, got 11000000000")
	na.ld.Timestamp = uint64(daysecs + 100)
	assert.NoError(na.CheckBorrow(constants.NativeToken, addr0, ldc, daysecs*2+100))
	assert.NoError(na.Borrow(constants.NativeToken, addr0, ldc, daysecs*2+100))
	total := constants.LDC*2 + uint64(float64(constants.LDC*10_000/1_000_000))
	assert.Equal(total, na.ld.LendingLedger[addr0].Amount.Uint64(), "should has interest")
	assert.Equal(uint64(daysecs+100), na.ld.LendingLedger[addr0].UpdateAt)
	assert.Equal(uint64(daysecs*2+100), na.ld.LendingLedger[addr0].DueTime)

	na.ld.Timestamp = uint64(daysecs*3 + 100)
	assert.NoError(na.Borrow(constants.NativeToken, addr0, ldc, 0))
	total += uint64(float64(total * 10_000 / 1_000_000))            // DailyInterest
	total += uint64(float64(total * (10_000 + 10_000) / 1_000_000)) // DailyInterest and OverdueInterest
	total += constants.LDC                                          // new borrow
	assert.Equal(total, na.ld.LendingLedger[addr0].Amount.Uint64(), "should has interest")
	assert.Equal(uint64(daysecs*3+100), na.ld.LendingLedger[addr0].UpdateAt)
	assert.Equal(uint64(0), na.ld.LendingLedger[addr0].DueTime)

	assert.ErrorContains(na.CheckCloseLending(),
		"Account.CheckCloseLending failed: please repay all before close")
	assert.ErrorContains(na.CloseLending(),
		"Account.CheckCloseLending failed: please repay all before close")

	// Marshal
	data, err := na.Marshal()
	assert.NoError(err)
	na2, err := ParseAccount(na.id, data)
	assert.NoError(err)
	assert.Equal(na.ld.Bytes(), na2.ld.Bytes())

	// Repay
	assert.NoError(na.CheckRepay(constants.NativeToken, addr0, ldc))
	am, err := na.Repay(constants.NativeToken, addr0, ldc)
	assert.NoError(err)
	assert.Equal(constants.LDC, am.Uint64())
	total -= constants.LDC
	assert.Equal(total, na.ld.LendingLedger[addr0].Amount.Uint64())
	na.ld.Timestamp = uint64(daysecs*4 + 100)
	total += uint64(float64(total * 10_000 / 1_000_000)) // DailyInterest
	assert.NoError(na.CheckRepay(constants.NativeToken, addr0, new(big.Int).SetUint64(total+1)))
	am, err = na.Repay(constants.NativeToken, addr0, new(big.Int).SetUint64(total+1))
	assert.NoError(err)
	assert.Equal(total, am.Uint64())
	assert.Equal(0, len(na.ld.LendingLedger))
	assert.NotNil(na.ld.LendingLedger)

	assert.ErrorContains(na.CheckRepay(constants.NativeToken, addr0, new(big.Int).SetUint64(total+1)),
		"Account.CheckRepay failed: don't need to repay")

	// Close and Marshal again
	data, err = na.Marshal()
	assert.NoError(err)
	na2, err = ParseAccount(na.id, data)
	assert.NoError(err)
	assert.Equal(na.ld.Bytes(), na2.ld.Bytes())

	assert.NoError(na.CheckCloseLending())
	assert.NoError(na.CloseLending())
	assert.Error(na.CheckCloseLending())
	data, err = na.Marshal()
	assert.NoError(err)
	na2, err = ParseAccount(na.id, data)
	assert.NoError(err)
	assert.Equal(na.ld.Bytes(), na2.ld.Bytes())

	// OpenLending again
	assert.NoError(na.OpenLending(&ld.LendingConfig{
		Token:           token,
		DailyInterest:   10_000,
		OverdueInterest: 10_000,
		MinAmount:       new(big.Int).SetUint64(constants.LDC),
		MaxAmount:       new(big.Int).SetUint64(constants.LDC * 10),
	}))

	assert.ErrorContains(na.CheckBorrow(constants.NativeToken, addr0, ldc, 0),
		"Account.CheckBorrow failed: invalid token, expected LDC, got NativeLDC")
	assert.ErrorContains(na.CheckBorrow(token, addr0,
		new(big.Int).SetUint64(constants.LDC-1), 0),
		"Account.CheckBorrow failed: invalid amount, expected >= 1000000000, got 999999999")
	assert.ErrorContains(na.CheckBorrow(token, addr0, ldc, 0),
		"Account.CheckBorrow failed: 0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC has an insufficient LDC balance, expected 1000000000, got 0")

	na.ld.Timestamp = uint64(daysecs * 5)
	na.Add(token, new(big.Int).SetUint64(constants.LDC*10))
	assert.NoError(na.CheckBorrow(token, addr0, ldc, 0))
	assert.Nil(na.ld.LendingLedger[addr0])
	assert.NoError(na.Borrow(token, addr0, ldc, 0))
	assert.NotNil(na.ld.LendingLedger[addr0])
	assert.Equal(constants.LDC, na.ld.LendingLedger[addr0].Amount.Uint64())
	assert.Equal(uint64(daysecs*5), na.ld.LendingLedger[addr0].UpdateAt)
	assert.Equal(uint64(0), na.ld.LendingLedger[addr0].DueTime)

	// Save again
	data, err = na.Marshal()
	assert.NoError(err)
	na2, err = ParseAccount(na.id, data)
	assert.NoError(err)
	assert.Equal(na.ld.Bytes(), na2.ld.Bytes())

	// Repay
	na.ld.Timestamp = uint64(daysecs * 6)
	assert.Error(na.CheckRepay(constants.NativeToken, addr0, ldc))
	assert.NoError(na.CheckRepay(token, addr0, ldc))
	am, err = na.Repay(token, addr0, ldc)
	assert.NoError(err)
	assert.Equal(constants.LDC, am.Uint64())
	total = constants.LDC
	total = uint64(float64(total * 10_000 / 1_000_000)) // DailyInterest
	assert.Equal(total, na.ld.LendingLedger[addr0].Amount.Uint64())
	assert.Equal(1, len(na.ld.LendingLedger))

	am, err = na.Repay(token, addr0, ldc)
	assert.NoError(err)
	assert.Equal(total, am.Uint64())
	assert.Equal(0, len(na.ld.LendingLedger))
	assert.NotNil(na.ld.LendingLedger)

	data, err = na.Marshal()
	assert.NoError(err)
	na2, err = ParseAccount(na.id, data)
	assert.NoError(err)
	assert.Equal(na.ld.Bytes(), na2.ld.Bytes())

	// calcBorrowTotal
	na.ld.Timestamp = uint64(0)
	assert.NoError(na.Borrow(token, addr0, ldc, uint64(daysecs*10)))
	entry := na.ld.LendingLedger[addr0]
	total = constants.LDC
	assert.Equal(uint64(0), na.calcBorrowTotal(util.Signer2.Address()).Uint64())
	assert.Equal(total, na.calcBorrowTotal(addr0).Uint64())

	na.ld.Timestamp = uint64(daysecs * 1)
	total += uint64(float64(total * 10_000 / 1_000_000)) // DailyInterest * 1 day
	assert.Equal(total, na.calcBorrowTotal(addr0).Uint64())
	na.ld.Timestamp = uint64(daysecs * 2)
	total += uint64(float64(total * 10_000 / 1_000_000)) // DailyInterest * 2 day
	assert.Equal(total, na.calcBorrowTotal(addr0).Uint64())

	na.ld.Timestamp = uint64(daysecs * 5.5)
	rate := math.Pow(1+float64(10_000)/1_000_000, float64(5.5))
	total = uint64(float64(constants.LDC) * rate) // DailyInterest * 5.5 day
	assert.Equal(total, na.calcBorrowTotal(addr0).Uint64())
	entry.UpdateAt = na.ld.Timestamp
	entry.Amount.SetUint64(total)

	na.ld.Timestamp = uint64(daysecs * 12)
	rate = math.Pow(1+float64(10_000)/1_000_000, float64(10))
	fa := new(big.Float).SetUint64(constants.LDC)
	fa.Mul(fa, big.NewFloat(rate))
	rate = math.Pow(1+float64(10_000*2)/1_000_000, float64(2))
	fa.Mul(fa, big.NewFloat(rate))
	total, _ = fa.Uint64()
	assert.Equal(total, na.calcBorrowTotal(addr0).Uint64())
	entry.UpdateAt = na.ld.Timestamp
	entry.Amount.SetUint64(total)

	na.ld.Timestamp = uint64(daysecs * 99.8)
	rate = math.Pow(1+float64(10_000)/1_000_000, float64(10))
	fa = new(big.Float).SetUint64(constants.LDC)
	fa.Mul(fa, big.NewFloat(rate))
	rate = math.Pow(1+float64(10_000*2)/1_000_000, float64(89.8))
	fa.Mul(fa, big.NewFloat(rate))
	am, _ = fa.Int(nil)
	am.Sub(am, na.calcBorrowTotal(addr0))
	am.Abs(am)
	assert.True(am.Uint64() <= 2)
}
