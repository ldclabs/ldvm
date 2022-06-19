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

func TestStakeAccount(t *testing.T) {
	assert := assert.New(t)

	acc := NewAccount(util.Signer1.Address())
	acc.Init(big.NewInt(0), 0, 0)
	acc2 := NewAccount(util.Signer2.Address())
	acc2.Init(big.NewInt(0), 0, 0)
	pledge := big.NewInt(1000)
	cfg := &ld.TxAccounter{
		Threshold: ld.Uint16Ptr(1),
		Keepers:   &util.EthIDs{util.Signer1.Address(), util.Signer2.Address()},
	}
	scfg := &ld.StakeConfig{
		LockTime:    2,
		WithdrawFee: 100_000,
		MinAmount:   big.NewInt(100),
		MaxAmount:   big.NewInt(1000),
	}
	assert.ErrorContains(acc.CheckCreateStake(util.Signer1.Address(), pledge, cfg, scfg),
		"Account(0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC).CheckCreateStake error: invalid stake account")
	assert.ErrorContains(acc.CreateStake(util.Signer1.Address(), pledge, cfg, scfg),
		"Account(0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC).CreateStake error: invalid stake account")
	assert.ErrorContains(acc.CheckResetStake(scfg),
		"Account(0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC).CheckResetStake error: invalid stake account")
	assert.ErrorContains(acc.ResetStake(scfg),
		"Account(0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC).ResetStake error: invalid stake account")
	assert.ErrorContains(acc.CheckResetStake(scfg),
		"Account(0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC).CheckResetStake error: invalid stake account")
	assert.ErrorContains(acc.ResetStake(scfg),
		"Account(0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC).ResetStake error: invalid stake account")

	stake := ld.MustNewStake("#TEST")
	testStake := NewAccount(util.EthID(stake))
	testStake.Init(big.NewInt(100), 1, 1)
	assert.NoError(testStake.CheckCreateStake(util.Signer1.Address(), pledge, cfg, scfg))
	assert.NoError(testStake.CreateStake(util.Signer1.Address(), pledge, cfg, scfg))
	assert.Equal(false, testStake.valid(ld.StakeAccount))
	testStake.Add(constants.NativeToken, big.NewInt(1000))
	assert.Equal(true, testStake.valid(ld.StakeAccount))

	assert.Equal(uint64(900), testStake.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(uint64(1000), testStake.balanceOfAll(constants.NativeToken).Uint64())
	assert.Nil(testStake.ld.MaxTotalSupply)
	assert.NotNil(testStake.ld.StakeLedger)
	assert.Equal(uint16(1), testStake.Threshold())
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
	assert.Equal(uint64(1900), testStake.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(uint64(2000), testStake.balanceOfAll(constants.NativeToken).Uint64())

	// Marshal
	data, err := testStake.Marshal()
	assert.NoError(err)
	testStake2, err := ParseAccount(testStake.id, data)
	assert.NoError(err)
	assert.Equal(testStake.ld.Bytes(), testStake2.ld.Bytes())

	// Reset
	token := ld.MustNewToken("$TEST")
	assert.ErrorContains(testStake.CheckResetStake(&ld.StakeConfig{
		Type:        1,
		WithdrawFee: 10_000,
		MinAmount:   big.NewInt(1000),
		MaxAmount:   big.NewInt(10000),
	}), "Account(0x0000000000000000000000000000002354455354).CheckResetStake error: can't change stake type")
	assert.ErrorContains(testStake.CheckResetStake(&ld.StakeConfig{
		Token:       token,
		WithdrawFee: 10_000,
		MinAmount:   big.NewInt(1000),
		MaxAmount:   big.NewInt(10000),
	}), "Account(0x0000000000000000000000000000002354455354).CheckResetStake error: can't change stake token")
	assert.ErrorContains(testStake.CheckResetStake(&ld.StakeConfig{
		WithdrawFee: 10_000,
		MinAmount:   big.NewInt(1000),
		MaxAmount:   big.NewInt(10000),
	}), "Account(0x0000000000000000000000000000002354455354).CheckResetStake error: stake in lock, please retry after lockTime")
	assert.ErrorContains(testStake.CheckDestroyStake(acc),
		"Account(0x0000000000000000000000000000002354455354).CheckDestroyStake error: stake in lock, please retry after lockTime")
	testStake.ld.Timestamp = 10
	assert.ErrorContains(testStake.CheckResetStake(&ld.StakeConfig{
		WithdrawFee: 10_000,
		MinAmount:   big.NewInt(1000),
		MaxAmount:   big.NewInt(10000),
	}), "Account(0x0000000000000000000000000000002354455354).CheckResetStake error: stake holders should not more than 1")
	assert.ErrorContains(testStake.CheckDestroyStake(acc),
		"Account(0x0000000000000000000000000000002354455354).CheckDestroyStake error: stake ledger not empty, please withdraw all except recipient")
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
	assert.NotNil(testStake.ld.Lending)
	assert.NotNil(testStake.ld.LendingLedger)

	// Destroy
	assert.NoError(testStake.Borrow(constants.NativeToken, acc.id, big.NewInt(1000), 0))
	assert.ErrorContains(testStake.CheckDestroyToken(acc),
		"Account(0x0000000000000000000000000000002354455354).CheckDestroyToken error: please repay all before close")
	assert.ErrorContains(testStake.DestroyToken(acc),
		"Account(0x0000000000000000000000000000002354455354).DestroyToken error: please repay all before close")
	actual, err := testStake.Repay(constants.NativeToken, acc.id, big.NewInt(1000))
	assert.NoError(err)
	assert.Equal(uint64(1000), actual.Uint64())

	assert.ErrorContains(testStake.CheckDestroyStake(acc2),
		"Account(0x0000000000000000000000000000002354455354).CheckDestroyStake error: recipient not exists")
	assert.NoError(testStake.CheckDestroyStake(acc))
	assert.NotNil(testStake.ld.Stake)
	assert.NoError(testStake.DestroyStake(acc))
	assert.Equal(uint64(0), testStake.balanceOf(constants.NativeToken).Uint64())
	assert.Equal(uint64(0), testStake.balanceOfAll(constants.NativeToken).Uint64())
	assert.Equal(uint16(0), testStake.Threshold())
	assert.Equal(util.EthIDs{}, testStake.Keepers())
	assert.Nil(testStake.ld.Stake)
	assert.Nil(testStake.ld.StakeLedger)
	assert.Nil(testStake.ld.Lending)
	assert.Nil(testStake.ld.LendingLedger)
	assert.Equal(0, len(testStake.ld.Tokens))
	assert.Equal(uint64(2000), acc.balanceOf(constants.NativeToken).Uint64())

	// Destroy again
	assert.ErrorContains(testStake.CheckDestroyStake(acc),
		"Account(0x0000000000000000000000000000002354455354).CheckDestroyStake error: invalid stake account")
	assert.ErrorContains(testStake.CheckResetStake(&ld.StakeConfig{
		WithdrawFee: 10_000,
		MinAmount:   big.NewInt(1000),
		MaxAmount:   big.NewInt(10000),
	}), "Account(0x0000000000000000000000000000002354455354).CheckResetStake error: invalid stake account")

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

func TestStakeFromAndTo(t *testing.T) {
	assert := assert.New(t)

	acc := NewAccount(util.Signer1.Address())
	acc.Init(big.NewInt(0), 0, 0)
	pledge := big.NewInt(1000)
	cfg := &ld.TxAccounter{
		Threshold: ld.Uint16Ptr(1),
		Keepers:   &util.EthIDs{util.Signer1.Address(), util.Signer2.Address()},
	}

	stake := ld.MustNewStake("#TEST")
	testStake := NewAccount(util.EthID(stake))
	testStake.Init(big.NewInt(100), 1, 1)
	assert.NoError(testStake.CreateStake(util.Signer1.Address(), pledge, cfg, &ld.StakeConfig{
		Type:        0,
		LockTime:    0,
		WithdrawFee: 100_000,
		MinAmount:   big.NewInt(100),
		MaxAmount:   big.NewInt(1000),
	}))
	testStake.Add(constants.NativeToken, big.NewInt(1000))

	// CheckAsFrom
	for _, ty := range ld.AllTxTypes {
		switch {
		case ld.StakeFromTxTypes0.Has(ty):
			assert.NoError(testStake.CheckAsFrom(ty))
		default:
			assert.Error(testStake.CheckAsFrom(ty))
		}
	}
	// CheckAsTo
	for _, ty := range ld.AllTxTypes {
		switch {
		case ld.StakeToTxTypes.Has(ty):
			assert.NoError(testStake.CheckAsTo(ty))
		default:
			assert.Error(testStake.CheckAsTo(ty))
		}
	}

	assert.NoError(testStake.DestroyStake(acc))
	assert.NoError(testStake.CreateStake(util.Signer1.Address(), pledge, cfg, &ld.StakeConfig{
		Type:        1,
		LockTime:    0,
		WithdrawFee: 100_000,
		MinAmount:   big.NewInt(100),
		MaxAmount:   big.NewInt(1000),
	}))
	testStake.Add(constants.NativeToken, big.NewInt(1000))

	// CheckAsFrom
	for _, ty := range ld.AllTxTypes {
		switch {
		case ld.StakeFromTxTypes1.Has(ty):
			assert.NoError(testStake.CheckAsFrom(ty))
		default:
			assert.Error(testStake.CheckAsFrom(ty))
		}
	}
	// CheckAsTo
	for _, ty := range ld.AllTxTypes {
		switch {
		case ld.StakeToTxTypes.Has(ty):
			assert.NoError(testStake.CheckAsTo(ty))
		default:
			assert.Error(testStake.CheckAsTo(ty))
		}
	}

	assert.NoError(testStake.DestroyStake(acc))
	assert.NoError(testStake.CreateStake(util.Signer1.Address(), pledge, cfg, &ld.StakeConfig{
		Type:        2,
		LockTime:    0,
		WithdrawFee: 100_000,
		MinAmount:   big.NewInt(100),
		MaxAmount:   big.NewInt(1000),
	}))
	testStake.Add(constants.NativeToken, big.NewInt(1000))

	// CheckAsFrom
	for _, ty := range ld.AllTxTypes {
		switch {
		case ld.StakeFromTxTypes2.Has(ty):
			assert.NoError(testStake.CheckAsFrom(ty))
		default:
			assert.Error(testStake.CheckAsFrom(ty))
		}
	}
	// CheckAsTo
	for _, ty := range ld.AllTxTypes {
		switch {
		case ld.StakeToTxTypes.Has(ty):
			assert.NoError(testStake.CheckAsTo(ty))
		default:
			assert.Error(testStake.CheckAsTo(ty))
		}
	}

	assert.NoError(testStake.DestroyStake(acc))
	assert.Error(testStake.CreateStake(util.Signer1.Address(), pledge, cfg, &ld.StakeConfig{
		Type:        3,
		LockTime:    0,
		WithdrawFee: 100_000,
		MinAmount:   big.NewInt(100),
		MaxAmount:   big.NewInt(1000),
	}))
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
	token := ld.MustNewToken("$LDC")
	withdrawFee := uint64(100_000)
	sa := NewAccount(util.EthID(ld.MustNewStake("#LDC"))).Init(pledge, 1, 1)
	assert.NoError(sa.CreateStake(util.Signer1.Address(), pledge, &ld.TxAccounter{
		Threshold: ld.Uint16Ptr(1),
		Keepers:   &util.EthIDs{util.Signer1.Address(), util.Signer2.Address()},
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
		"Account(0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC).CheckTakeStake error: invalid stake account")
	assert.ErrorContains(sa.CheckTakeStake(token, addr2, big.NewInt(1000), 0),
		"Account(0x00000000000000000000000000000000234c4443).CheckTakeStake error: invalid token, expected NativeLDC, got $LDC")
	assert.ErrorContains(sa.CheckTakeStake(constants.NativeToken, addr3, big.NewInt(1000), 0),
		"Account(0x00000000000000000000000000000000234c4443).CheckTakeStake error: invalid amount, expected >= 1000000000, got 1000")

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
		"Account(0x00000000000000000000000000000000234c4443).CheckTakeStake error: invalid total amount, expected <= 10000000000, got 11000000000")
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
		"Account(0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC).CheckWithdrawStake error: invalid stake account")
	assert.ErrorContains(
		sa.CheckWithdrawStake(token, addr0, util.EthIDs{}, big.NewInt(1000)),
		"Account(0x00000000000000000000000000000000234c4443).CheckWithdrawStake error: invalid token, expected NativeLDC, got $LDC")
	assert.ErrorContains(
		sa.CheckWithdrawStake(constants.NativeToken, addr0, util.EthIDs{}, big.NewInt(1000)),
		"Account(0x00000000000000000000000000000000234c4443).CheckWithdrawStake error: stake in lock, please retry after lockTime")
	sa.ld.Timestamp = 10
	assert.NoError(
		sa.CheckWithdrawStake(constants.NativeToken, addr0, util.EthIDs{}, big.NewInt(1000)))
	assert.ErrorContains(
		sa.CheckWithdrawStake(constants.NativeToken, addr3, util.EthIDs{}, big.NewInt(1000)),
		"has no stake to withdraw")
	assert.ErrorContains(
		sa.CheckWithdrawStake(constants.NativeToken, addr2, util.EthIDs{}, big.NewInt(1000)),
		"Account(0x00000000000000000000000000000000234c4443).CheckWithdrawStake error: stake in lock, please retry after lockTime")
	sa.ld.Timestamp = 11
	assert.NoError(
		sa.CheckWithdrawStake(constants.NativeToken, addr2, util.EthIDs{}, big.NewInt(1000)))

	// check UpdateStakeApprover
	assert.ErrorContains(
		sk.CheckUpdateStakeApprover(addr1, approver, util.EthIDs{}),
		"Account(0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC).CheckUpdateStakeApprover error: invalid stake account")
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
		"insufficient stake to withdraw, expected 1000000001, got 1000000000")
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
		"insufficient NativeLDC balance for withdraw, expected 13036950323, got 3036950323")

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
	assert.Equal(uint16(0), sa.Threshold())
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
		Threshold: ld.Uint16Ptr(1),
		Keepers:   &util.EthIDs{util.Signer1.Address(), util.Signer2.Address()},
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
		"Account(0x00000000000000000000000000000000234c4443).CheckTakeStake error: invalid total amount, expected <= 10000000000, got 11000000000")

	// sa take a stake in sc
	sc := NewAccount(util.EthID(ld.MustNewStake("#HODLING"))).Init(pledge, 1, 1)
	assert.NoError(sc.CreateStake(util.Signer2.Address(), pledge, &ld.TxAccounter{
		Threshold: ld.Uint16Ptr(1),
		Keepers:   &util.EthIDs{util.Signer2.Address()},
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
		"Account(0x00000000000000000000000000000000234c4443).CheckWithdrawStake error: insufficient $LDC balance for withdraw, expected 10000000000, got 1000000000")

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
	assert.Equal(uint16(0), sc.Threshold())
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
	assert.Equal(uint16(0), sa.Threshold())
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
