// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package chain

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/ava-labs/avalanchego/database/memdb"
	"github.com/stretchr/testify/assert"

	"github.com/ldclabs/ldvm/constants"
	"github.com/ldclabs/ldvm/ld"
	"github.com/ldclabs/ldvm/util"
)

func TestAccountCache(t *testing.T) {
	assert := assert.New(t)

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

	// SaveTo
	db := memdb.New()
	assert.NoError(acc.SaveTo(db))
	data, err := db.Get(acc.id[:])
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

	db := memdb.New()
	assert.NoError(nativeToken.SaveTo(db))
	data, err := db.Get(nativeToken.id[:])
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

	// Save
	assert.NoError(testToken.SaveTo(db))
	data, err = db.Get(testToken.id[:])
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

	// Save again
	assert.NoError(testToken.SaveTo(db))
	data, err = db.Get(testToken.id[:])
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

	// Save
	db := memdb.New()
	assert.NoError(testStake.SaveTo(db))
	data, err := db.Get(testStake.id[:])
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
	}), "Account.CheckResetStake failed: stake ledger not empty, please withdraw all except holder")
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

	// Save again
	assert.NoError(testStake.SaveTo(db))
	data, err = db.Get(testStake.id[:])
	assert.NoError(err)
	testStake2, err = ParseAccount(testStake.id, data)
	assert.NoError(err)
	assert.Equal(testStake.ld.Bytes(), testStake2.ld.Bytes())

	// Create again
	assert.NoError(testStake.CheckCreateStake(util.Signer1.Address(), pledge, cfg, scfg))
	assert.NoError(testStake.CreateStake(util.Signer1.Address(), pledge, cfg, scfg))
}

func TestTakeStakeAndWithdraw(t *testing.T) {}

func TestLending(t *testing.T) {}
