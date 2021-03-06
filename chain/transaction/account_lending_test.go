// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package transaction

import (
	"math"
	"math/big"
	"testing"

	"github.com/ldclabs/ldvm/constants"
	"github.com/ldclabs/ldvm/ld"
	"github.com/ldclabs/ldvm/util"
	"github.com/stretchr/testify/assert"
)

func TestLending(t *testing.T) {
	assert := assert.New(t)

	addr0 := util.NewSigner().Address()
	na := NewAccount(util.Signer1.Address()).Init(big.NewInt(0), 10, 100)

	// Lending
	ldc := new(big.Int).SetUint64(constants.LDC)
	token := ld.MustNewToken("$LDC")
	lcfg := &ld.LendingConfig{
		DailyInterest:   10_000,
		OverdueInterest: 10_000,
		MinAmount:       new(big.Int).SetUint64(constants.LDC),
		MaxAmount:       new(big.Int).SetUint64(constants.LDC * 10),
	}
	assert.ErrorContains(na.CloseLending(),
		"Account(0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC).CloseLending error: invalid lending")
	assert.ErrorContains(na.Borrow(constants.NativeToken, addr0, ldc, 0),
		"Account(0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC).Borrow error: invalid lending")

	_, err := na.Repay(constants.NativeToken, addr0, ldc)
	assert.ErrorContains(err,
		"Account(0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC).Repay error: invalid lending")
	assert.ErrorContains(na.OpenLending(lcfg),
		"Account(0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC).OpenLending error: invalid ledger")

	assert.NoError(na.InitLedger(nil))
	assert.NoError(na.OpenLending(lcfg))
	assert.ErrorContains(na.OpenLending(lcfg),
		"Account(0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC).OpenLending error: lending exists")

	assert.ErrorContains(na.Borrow(token, addr0, ldc, 0),
		"Account(0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC).Borrow error: invalid token, expected NativeLDC, got $LDC")
	_, err = na.Repay(token, addr0, ldc)
	assert.ErrorContains(err,
		"Account(0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC).Repay error: invalid token, expected NativeLDC, got $LDC")
	_, err = na.Repay(constants.NativeToken, addr0, ldc)
	assert.ErrorContains(err,
		"Account(0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC).Repay error: don't need to repay")

	assert.ErrorContains(na.Borrow(constants.NativeToken, addr0, ldc, 100),
		"Account(0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC).Borrow error: invalid dueTime, expected > 100, got 100")
	assert.ErrorContains(na.Borrow(constants.NativeToken, addr0, new(big.Int).SetUint64(constants.LDC-1), 0),
		"Account(0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC).Borrow error: invalid amount, expected >= 1000000000, got 999999999")
	assert.ErrorContains(na.Borrow(constants.NativeToken, addr0, ldc, 0),
		"Account(0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC).Borrow error: insufficient NativeLDC balance, expected 1000000000, got 0")

	na.Add(constants.NativeToken, new(big.Int).SetUint64(constants.LDC*10))
	assert.Nil(na.ledger.Lending[addr0.AsKey()])
	assert.NoError(na.Borrow(constants.NativeToken, addr0, ldc, daysecs+100))
	assert.NotNil(na.ledger.Lending[addr0.AsKey()])
	assert.Equal(constants.LDC, na.ledger.Lending[addr0.AsKey()].Amount.Uint64())
	assert.Equal(uint64(100), na.ledger.Lending[addr0.AsKey()].UpdateAt)
	assert.Equal(uint64(daysecs+100), na.ledger.Lending[addr0.AsKey()].DueTime)

	assert.ErrorContains(na.Borrow(constants.NativeToken, addr0,
		new(big.Int).SetUint64(constants.LDC*10), 0),
		"Account(0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC).Borrow error: invalid amount, expected <= 10000000000, got 11000000000")
	na.ld.Timestamp = uint64(daysecs + 100)
	assert.NoError(na.Borrow(constants.NativeToken, addr0, ldc, daysecs*2+100))
	total := constants.LDC*2 + uint64(float64(constants.LDC*10_000/1_000_000))
	assert.Equal(total, na.ledger.Lending[addr0.AsKey()].Amount.Uint64(), "should has interest")
	assert.Equal(uint64(daysecs+100), na.ledger.Lending[addr0.AsKey()].UpdateAt)
	assert.Equal(uint64(daysecs*2+100), na.ledger.Lending[addr0.AsKey()].DueTime)

	na.ld.Timestamp = uint64(daysecs*3 + 100)
	assert.NoError(na.Borrow(constants.NativeToken, addr0, ldc, 0))
	total += uint64(float64(total * 10_000 / 1_000_000))            // DailyInterest
	total += uint64(float64(total * (10_000 + 10_000) / 1_000_000)) // DailyInterest and OverdueInterest
	total += constants.LDC                                          // new borrow
	assert.Equal(total, na.ledger.Lending[addr0.AsKey()].Amount.Uint64(), "should has interest")
	assert.Equal(uint64(daysecs*3+100), na.ledger.Lending[addr0.AsKey()].UpdateAt)
	assert.Equal(uint64(0), na.ledger.Lending[addr0.AsKey()].DueTime)

	assert.ErrorContains(na.CloseLending(),
		"Account(0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC).CloseLending error: please repay all before close")

	// Marshal
	data, ledger, err := na.Marshal()
	assert.NoError(err)
	na2, err := ParseAccount(na.id, data)
	assert.NoError(err)
	assert.Equal(na.ld.Bytes(), na2.ld.Bytes())

	lg := &ld.AccountLedger{}
	assert.NoError(lg.Unmarshal(ledger))
	assert.NoError(lg.SyntacticVerify())
	assert.Equal(ledger, lg.Bytes())

	// Repay
	am, err := na.Repay(constants.NativeToken, addr0, ldc)
	assert.NoError(err)
	assert.Equal(constants.LDC, am.Uint64())
	total -= constants.LDC
	assert.Equal(total, na.ledger.Lending[addr0.AsKey()].Amount.Uint64())
	na.ld.Timestamp = uint64(daysecs*4 + 100)
	total += uint64(float64(total * 10_000 / 1_000_000)) // DailyInterest
	am, err = na.Repay(constants.NativeToken, addr0, new(big.Int).SetUint64(total+1))
	assert.NoError(err)
	assert.Equal(total, am.Uint64())
	assert.NotNil(na.ledger.Lending)
	assert.Equal(0, len(na.ledger.Lending))

	_, err = na.Repay(constants.NativeToken, addr0, new(big.Int).SetUint64(total+1))
	assert.ErrorContains(err,
		"Account(0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC).Repay error: don't need to repay")

	// Close and Marshal again
	data, ledger, err = na.Marshal()
	assert.NoError(err)
	na2, err = ParseAccount(na.id, data)
	assert.NoError(err)
	assert.Equal(na.ld.Bytes(), na2.ld.Bytes())

	lg = &ld.AccountLedger{}
	assert.NoError(lg.Unmarshal(ledger))
	assert.NoError(lg.SyntacticVerify())
	assert.Equal(ledger, lg.Bytes())

	na.ledger = nil
	assert.ErrorContains(na.CloseLending(),
		"Account(0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC).CloseLending error: invalid ledger")

	na.ledger = lg
	assert.NoError(na.CloseLending())
	data, ledger, err = na.Marshal()
	assert.NoError(err)
	na2, err = ParseAccount(na.id, data)
	assert.NoError(err)
	assert.Equal(na.ld.Bytes(), na2.ld.Bytes())

	lg = &ld.AccountLedger{}
	assert.NoError(lg.Unmarshal(ledger))
	assert.NoError(lg.SyntacticVerify())
	assert.Equal(ledger, lg.Bytes())

	// OpenLending again
	assert.NoError(na.OpenLending(&ld.LendingConfig{
		Token:           token,
		DailyInterest:   10_000,
		OverdueInterest: 10_000,
		MinAmount:       new(big.Int).SetUint64(constants.LDC),
		MaxAmount:       new(big.Int).SetUint64(constants.LDC * 10),
	}))

	assert.ErrorContains(na.Borrow(constants.NativeToken, addr0, ldc, 0),
		"Account(0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC).Borrow error: invalid token, expected $LDC, got NativeLDC")
	assert.ErrorContains(na.Borrow(token, addr0, new(big.Int).SetUint64(constants.LDC-1), 0),
		"Account(0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC).Borrow error: invalid amount, expected >= 1000000000, got 999999999")
	assert.ErrorContains(na.Borrow(token, addr0, ldc, 0),
		"Account(0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC).Borrow error: insufficient $LDC balance, expected 1000000000, got 0")

	na.ld.Timestamp = uint64(daysecs * 5)
	na.Add(token, new(big.Int).SetUint64(constants.LDC*10))
	assert.Nil(na.ledger.Lending[addr0.AsKey()])
	assert.NoError(na.Borrow(token, addr0, ldc, 0))
	assert.NotNil(na.ledger.Lending[addr0.AsKey()])
	assert.Equal(constants.LDC, na.ledger.Lending[addr0.AsKey()].Amount.Uint64())
	assert.Equal(uint64(daysecs*5), na.ledger.Lending[addr0.AsKey()].UpdateAt)
	assert.Equal(uint64(0), na.ledger.Lending[addr0.AsKey()].DueTime)

	// Save again
	data, ledger, err = na.Marshal()
	assert.NoError(err)
	na2, err = ParseAccount(na.id, data)
	assert.NoError(err)
	assert.Equal(na.ld.Bytes(), na2.ld.Bytes())

	lg = &ld.AccountLedger{}
	assert.NoError(lg.Unmarshal(ledger))
	assert.NoError(lg.SyntacticVerify())
	assert.Equal(ledger, lg.Bytes())

	// Repay
	na.ld.Timestamp = uint64(daysecs * 6)
	_, err = na.Repay(constants.NativeToken, addr0, ldc)
	assert.Error(err)
	am, err = na.Repay(token, addr0, ldc)
	assert.NoError(err)
	assert.Equal(constants.LDC, am.Uint64())
	total = constants.LDC
	total = uint64(float64(total * 10_000 / 1_000_000)) // DailyInterest
	assert.Equal(total, na.ledger.Lending[addr0.AsKey()].Amount.Uint64())
	assert.Equal(1, len(na.ledger.Lending))

	am, err = na.Repay(token, addr0, ldc)
	assert.NoError(err)
	assert.Equal(total, am.Uint64())
	assert.Equal(0, len(na.ledger.Lending))
	assert.NotNil(na.ledger.Lending)

	data, ledger, err = na.Marshal()
	assert.NoError(err)
	na2, err = ParseAccount(na.id, data)
	assert.NoError(err)
	assert.Equal(na.ld.Bytes(), na2.ld.Bytes())

	lg = &ld.AccountLedger{}
	assert.NoError(lg.Unmarshal(ledger))
	assert.NoError(lg.SyntacticVerify())
	assert.Equal(ledger, lg.Bytes())

	// calcBorrowTotal
	na.ld.Timestamp = uint64(0)
	assert.NoError(na.Borrow(token, addr0, ldc, uint64(daysecs*10)))
	entry := na.ledger.Lending[addr0.AsKey()]
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
