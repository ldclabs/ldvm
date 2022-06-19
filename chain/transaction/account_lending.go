// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package transaction

import (
	"fmt"
	"math"
	"math/big"

	"github.com/ldclabs/ldvm/ld"
	"github.com/ldclabs/ldvm/util"
)

func (a *Account) CheckOpenLending(cfg *ld.LendingConfig) error {
	a.mu.RLock()
	defer a.mu.RUnlock()

	errp := util.ErrPrefix(fmt.Sprintf("Account(%s).CheckOpenLending error: ", a.id))
	return errp.ErrorIf(a.openLending(cfg, false))
}

func (a *Account) OpenLending(cfg *ld.LendingConfig) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	errp := util.ErrPrefix(fmt.Sprintf("Account(%s).OpenLending error: ", a.id))
	return errp.ErrorIf(a.openLending(cfg, true))
}

func (a *Account) openLending(cfg *ld.LendingConfig, write bool) error {
	if a.ld.Lending != nil || a.ld.LendingLedger != nil {
		return fmt.Errorf("lending exists")
	}

	if err := cfg.SyntacticVerify(); err != nil {
		return err
	}

	if write {
		a.ld.Lending = cfg
		a.ld.LendingLedger = make(map[util.EthID]*ld.LendingEntry)
	}
	return nil
}

func (a *Account) CheckCloseLending() error {
	a.mu.RLock()
	defer a.mu.RUnlock()

	errp := util.ErrPrefix(fmt.Sprintf("Account(%s).CheckCloseLending error: ", a.id))
	return errp.ErrorIf(a.closeLending(false, false))
}

func (a *Account) CloseLending() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	errp := util.ErrPrefix(fmt.Sprintf("Account(%s).CloseLending error: ", a.id))
	return errp.ErrorIf(a.closeLending(true, false))
}

func (a *Account) closeLending(write, ignoreNone bool) error {
	if ignoreNone && a.ld.Lending == nil {
		return nil
	}

	if a.ld.Lending == nil || a.ld.LendingLedger == nil {
		return fmt.Errorf("invalid lending")
	}

	if len(a.ld.LendingLedger) != 0 {
		return fmt.Errorf("please repay all before close")
	}

	if write {
		a.ld.Lending = nil
		a.ld.LendingLedger = nil
	}
	return nil
}

func (a *Account) CheckBorrow(
	token util.TokenSymbol,
	from util.EthID,
	amount *big.Int,
	dueTime uint64,
) error {
	a.mu.RLock()
	defer a.mu.RUnlock()

	errp := util.ErrPrefix(fmt.Sprintf("Account(%s).CheckBorrow error: ", a.id))
	return errp.ErrorIf(a.borrow(token, from, amount, dueTime, false))
}

func (a *Account) Borrow(
	token util.TokenSymbol,
	from util.EthID,
	amount *big.Int,
	dueTime uint64,
) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	errp := util.ErrPrefix(fmt.Sprintf("Account(%s).Borrow error: ", a.id))
	return errp.ErrorIf(a.borrow(token, from, amount, dueTime, true))
}

func (a *Account) borrow(
	token util.TokenSymbol,
	from util.EthID,
	amount *big.Int,
	dueTime uint64,
	write bool,
) error {
	switch {
	case a.ld.Lending == nil || a.ld.LendingLedger == nil:
		return fmt.Errorf("invalid lending")

	case a.ld.Lending.Token != token:
		return fmt.Errorf("invalid token, expected %s, got %s",
			a.ld.Lending.Token.GoString(), token.GoString())

	case dueTime > 0 && dueTime <= a.ld.Timestamp:
		return fmt.Errorf("invalid dueTime, expected > %d, got %d", a.ld.Timestamp, dueTime)

	case amount.Cmp(a.ld.Lending.MinAmount) < 0:
		return fmt.Errorf("invalid amount, expected >= %v, got %v", a.ld.Lending.MinAmount, amount)
	}

	e := a.ld.LendingLedger[from]
	total := new(big.Int).Set(amount)
	switch {
	case e == nil:
		e = &ld.LendingEntry{Amount: new(big.Int).Set(amount)}

	default:
		total.Add(total, a.calcBorrowTotal(from))
	}

	if total.Cmp(a.ld.Lending.MaxAmount) > 0 {
		return fmt.Errorf("invalid amount, expected <= %v, got %v", a.ld.Lending.MaxAmount, total)
	}

	ba := a.balanceOf(token)
	if ba.Cmp(amount) < 0 {
		return fmt.Errorf("insufficient %s balance, expected %v, got %v", token.GoString(), amount, ba)
	}

	if write {
		e.Amount.Set(total)
		e.UpdateAt = a.ld.Timestamp
		e.DueTime = dueTime
		a.ld.LendingLedger[from] = e
	}
	return nil
}

func (a *Account) CheckRepay(
	token util.TokenSymbol,
	from util.EthID,
	amount *big.Int,
) error {
	a.mu.RLock()
	defer a.mu.RUnlock()

	errp := util.ErrPrefix(fmt.Sprintf("Account(%s).CheckRepay error: ", a.id))
	_, err := a.repay(token, from, amount, false)
	return errp.ErrorIf(err)
}

func (a *Account) Repay(
	token util.TokenSymbol,
	from util.EthID,
	amount *big.Int,
) (*big.Int, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	errp := util.ErrPrefix(fmt.Sprintf("Account(%s).Repay error: ", a.id))
	res, err := a.repay(token, from, amount, true)
	return res, errp.ErrorIf(err)
}

func (a *Account) repay(
	token util.TokenSymbol,
	from util.EthID,
	amount *big.Int,
	write bool,
) (*big.Int, error) {
	switch {
	case a.ld.Lending == nil || a.ld.LendingLedger == nil:
		return nil, fmt.Errorf("invalid lending")

	case a.ld.Lending.Token != token:
		return nil, fmt.Errorf("invalid token, expected %s, got %s",
			a.ld.Lending.Token.GoString(), token.GoString())
	}

	e := a.ld.LendingLedger[from]
	if e == nil {
		return nil, fmt.Errorf("don't need to repay")
	}

	if !write {
		return nil, nil
	}

	total := a.calcBorrowTotal(from)
	actual := new(big.Int).Set(amount)
	if actual.Cmp(total) >= 0 {
		actual.Set(total)
		delete(a.ld.LendingLedger, from)
	} else {
		e.Amount.Sub(total, actual)
		e.UpdateAt = a.ld.Timestamp
		a.ld.LendingLedger[from] = e
	}
	return actual, nil
}

const daysecs = 3600 * 24

func (a *Account) calcBorrowTotal(from util.EthID) *big.Int {
	cfg := a.ld.Lending
	amount := new(big.Int)

	if e := a.ld.LendingLedger[from]; e != nil {
		amount.Set(e.Amount)

		if amount.Sign() > 0 && a.ld.Timestamp > e.UpdateAt {
			var rate float64
			sec := a.ld.Timestamp - e.UpdateAt
			fa := new(big.Float).SetInt(amount)

			switch {
			case e.DueTime == 0 || a.ld.Timestamp <= e.DueTime:
				rate = math.Pow(1+float64(cfg.DailyInterest)/1_000_000, float64(sec)/daysecs)
				fa.Mul(fa, big.NewFloat(rate))

			case e.UpdateAt >= e.DueTime:
				rate = math.Pow(1+float64(cfg.DailyInterest+cfg.OverdueInterest)/1_000_000, float64(sec)/daysecs)
				fa.Mul(fa, big.NewFloat(rate))

			default:
				rate = math.Pow(1+float64(cfg.DailyInterest)/1_000_000, float64(e.DueTime-e.UpdateAt)/daysecs)
				fa.Mul(fa, big.NewFloat(rate))
				rate = math.Pow(1+float64(cfg.DailyInterest+cfg.OverdueInterest)/1_000_000,
					float64(a.ld.Timestamp-e.DueTime)/daysecs)
				fa.Mul(fa, big.NewFloat(rate))
			}

			fa.Int(amount)
		}
	}
	return amount
}
