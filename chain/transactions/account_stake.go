// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package transactions

import (
	"fmt"
	"math/big"

	"github.com/ldclabs/ldvm/constants"
	"github.com/ldclabs/ldvm/ld"
	"github.com/ldclabs/ldvm/util"
)

func (a *Account) CreateStake(
	from util.EthID,
	pledge *big.Int,
	acc *ld.TxAccounter,
	cfg *ld.StakeConfig,
) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	errp := util.ErrPrefix(fmt.Sprintf("Account(%s).CreateStake error: ", a.id))
	stake := util.StakeSymbol(a.id)
	if !stake.Valid() {
		return errp.Errorf("invalid stake account")
	}
	if !a.IsEmpty() {
		return errp.Errorf("stake account %s exists", stake)
	}
	if err := cfg.SyntacticVerify(); err != nil {
		return errp.ErrorIf(err)
	}

	if a.ledger == nil {
		return errp.Errorf("invalid ledger")
	}

	a.ld.Type = ld.StakeAccount
	a.ld.Threshold = *acc.Threshold
	a.ld.Keepers = *acc.Keepers
	a.ld.Approver = acc.Approver
	a.ld.ApproveList = acc.ApproveList
	a.ld.Stake = cfg
	a.ld.MaxTotalSupply = nil
	switch cfg.Token {
	case constants.NativeToken:
		a.ledger.Stake[from.AsKey()] = &ld.StakeEntry{Amount: new(big.Int).Set(pledge)}
	default:
		if b := a.ld.Tokens[cfg.Token.AsKey()]; b == nil {
			a.ld.Tokens[cfg.Token.AsKey()] = new(big.Int)
		}
	}

	return nil
}

func (a *Account) ResetStake(cfg *ld.StakeConfig) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	errp := util.ErrPrefix(fmt.Sprintf("Account(%s).ResetStake error: ", a.id))
	if !a.valid(ld.StakeAccount) {
		return errp.Errorf("invalid stake account")
	}
	if a.ledger == nil {
		return errp.Errorf("invalid ledger")
	}

	if err := cfg.SyntacticVerify(); err != nil {
		return errp.ErrorIf(err)
	}
	if cfg.Type != a.ld.Stake.Type {
		return errp.Errorf("can't change stake type, expected %d, got %d",
			a.ld.Stake.Type, cfg.Type)
	}
	if cfg.Token != a.ld.Stake.Token {
		return errp.Errorf("can't change stake token, expected %s, got %s",
			a.ld.Stake.Token.GoString(), cfg.Token.GoString())
	}
	if a.ld.Stake.LockTime >= a.ld.Timestamp {
		return errp.Errorf("stake in lock, please retry after lockTime, Unix(%d)",
			a.ld.Stake.LockTime)
	}

	holders := 0
	for _, v := range a.ledger.Stake {
		if v.Amount.Sign() > 0 {
			holders++
		}
	}
	if holders > 1 {
		return errp.Errorf("stake holders should not more than 1")
	}

	a.ld.Stake.LockTime = cfg.LockTime
	a.ld.Stake.WithdrawFee = cfg.WithdrawFee
	if cfg.MinAmount.Sign() > 0 {
		a.ld.Stake.MinAmount.Set(cfg.MinAmount)
	}
	if cfg.MaxAmount.Sign() > 0 {
		a.ld.Stake.MaxAmount.Set(cfg.MaxAmount)
	}
	return nil
}

func (a *Account) DestroyStake(recipient *Account) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	errp := util.ErrPrefix(fmt.Sprintf("Account(%s).DestroyStake error: ", a.id))
	if !a.valid(ld.StakeAccount) {
		return errp.Errorf("invalid stake account")
	}
	if a.ledger == nil {
		return errp.Errorf("invalid ledger")
	}

	if a.ld.Stake.LockTime >= a.ld.Timestamp {
		return errp.Errorf("stake in lock, please retry after lockTime, Unix(%d)",
			a.ld.Stake.LockTime)
	}

	holders := 0
	for _, v := range a.ledger.Stake {
		if v.Amount.Sign() > 0 {
			holders++
		}
	}

	switch holders {
	case 0:
		// just go ahead
	case 1:
		if v, ok := a.ledger.Stake[recipient.id.AsKey()]; !ok || v.Amount.Sign() <= 0 {
			return errp.Errorf("recipient not exists")
		}

	default:
		return errp.Errorf("stake ledger not empty, please withdraw all except recipient")
	}

	if err := a.closeLending(true); err != nil {
		return errp.ErrorIf(err)
	}

	recipient.Add(constants.NativeToken, a.ld.Balance)
	a.ld.Balance.SetUint64(0)
	if a.ld.Stake.Token != constants.NativeToken {
		if b, ok := a.ld.Tokens[a.ld.Stake.Token.AsKey()]; ok && b.Sign() > 0 {
			recipient.Add(a.ld.Stake.Token, b)
			b.SetUint64(0)
		}
	}
	a.ld.Type = 0
	a.ld.Threshold = 0
	a.ld.Keepers = a.ld.Keepers[:0]
	a.ld.NonceTable = make(map[uint64][]uint64)
	a.ld.Approver = nil
	a.ld.ApproveList = nil
	a.ld.Stake = nil
	a.ledger.Stake = make(map[string]*ld.StakeEntry)
	return nil
}

func (a *Account) TakeStake(
	token util.TokenSymbol,
	from util.EthID,
	amount *big.Int,
	lockTime uint64) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	errp := util.ErrPrefix(fmt.Sprintf("Account(%s).TakeStake error: ", a.id))
	if !a.valid(ld.StakeAccount) {
		return errp.Errorf("invalid stake account")
	}
	if a.ledger == nil {
		return errp.Errorf("invalid ledger")
	}

	stake := a.ld.Stake
	if token != a.ld.Stake.Token {
		return errp.Errorf("invalid token, expected %s, got %s",
			stake.Token.GoString(), token.GoString())
	}

	if amount.Cmp(stake.MinAmount) < 0 {
		return errp.Errorf("invalid amount, expected >= %v, got %v",
			stake.MinAmount, amount)
	}

	total := new(big.Int).Set(amount)
	v := a.ledger.Stake[from.AsKey()]
	rate := a.calcStakeBonusRate()
	if v != nil {
		bonus, _ := new(big.Float).Mul(new(big.Float).SetInt(v.Amount), rate).Int(nil)
		total.Add(total, v.Amount)
		total.Add(total, bonus)
	}
	if total.Cmp(stake.MaxAmount) > 0 {
		return errp.Errorf("invalid total amount for %s, expected <= %v, got %v",
			from, stake.MaxAmount, total)
	}
	if lockTime > 0 && lockTime <= stake.LockTime {
		return errp.Errorf("invalid lockTime, expected > %v, got %v",
			stake.LockTime, lockTime)
	}

	a.allocStakeBonus(rate)
	if v == nil {
		v = &ld.StakeEntry{Amount: new(big.Int)}
		a.ledger.Stake[from.AsKey()] = v
	}
	v.Amount.Add(v.Amount, amount)
	if lockTime > 0 {
		v.LockTime = lockTime
	}
	return nil
}

func (a *Account) UpdateStakeApprover(
	from, approver util.EthID,
	signers util.EthIDs,
) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	errp := util.ErrPrefix(fmt.Sprintf("Account(%s).UpdateStakeApprover error: ", a.id))
	if !a.valid(ld.StakeAccount) {
		return errp.Errorf("invalid stake account")
	}
	if a.ledger == nil {
		return errp.Errorf("invalid ledger")
	}

	v := a.ledger.Stake[from.AsKey()]
	if v == nil {
		return errp.Errorf("%s has no stake ledger to update", util.EthID(from))
	}
	if v.Approver != nil && !signers.Has(*v.Approver) {
		return errp.Errorf("%s need approver signing", util.EthID(from))
	}

	if approver == util.EthIDEmpty {
		v.Approver = nil
	} else {
		v.Approver = &approver
	}
	return nil
}

func (a *Account) WithdrawStake(
	token util.TokenSymbol,
	from util.EthID,
	signers util.EthIDs,
	amount *big.Int,
) (*big.Int, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	errp := util.ErrPrefix(fmt.Sprintf("Account(%s).WithdrawStake error: ", a.id))
	if !a.valid(ld.StakeAccount) {
		return nil, errp.Errorf("invalid stake account")
	}
	if a.ledger == nil {
		return nil, errp.Errorf("invalid ledger")
	}

	stake := a.ld.Stake
	if token != stake.Token {
		return nil, errp.Errorf("invalid token, expected %s, got %s",
			stake.Token.GoString(), token.GoString())
	}
	if stake.LockTime >= a.ld.Timestamp {
		return nil, errp.Errorf("stake in lock, please retry after lockTime, Unix(%d)", stake.LockTime)
	}

	v := a.ledger.Stake[from.AsKey()]
	if v == nil {
		return nil, errp.Errorf("%s has no stake to withdraw", from)
	}
	if v.LockTime >= a.ld.Timestamp {
		return nil, errp.Errorf("stake in lock, please retry after lockTime, Unix(%d)", v.LockTime)
	}
	if v.Approver != nil && !signers.Has(*v.Approver) {
		return nil, errp.Errorf("%s need approver signing", from)
	}

	total := new(big.Int).Set(v.Amount)
	rate := a.calcStakeBonusRate()
	bonus, _ := new(big.Float).Mul(new(big.Float).SetInt(v.Amount), rate).Int(nil)
	total = total.Add(total, bonus)
	if total.Cmp(amount) < 0 {
		return nil, errp.Errorf("%s has an insufficient stake to withdraw, expected %v, got %v",
			from, total, amount)
	}

	if ba := a.balanceOf(token); ba.Cmp(amount) < 0 {
		return nil, errp.Errorf("insufficient %s balance for withdraw, expected %v, got %v",
			token.GoString(), amount, ba)
	}

	a.allocStakeBonus(rate)
	v.Amount.Sub(v.Amount, amount)
	if v.Amount.Sign() <= 0 && v.Approver == nil {
		delete(a.ledger.Stake, from.AsKey())
	}
	withdraw := new(big.Int).Mul(amount, new(big.Int).SetUint64(stake.WithdrawFee))
	return withdraw.Sub(amount, withdraw.Quo(withdraw, big.NewInt(1_000_000))), nil
}

func (a *Account) GetStakeAmount(token util.TokenSymbol, from util.EthID) *big.Int {
	total := new(big.Int)
	stake := a.ld.Stake
	if a.valid(ld.StakeAccount) && a.ledger != nil && token == stake.Token {
		if v := a.ledger.Stake[from.AsKey()]; v != nil && v.Amount.Sign() > 0 {
			total.Set(v.Amount)
			rate := a.calcStakeBonusRate()
			bonus, _ := new(big.Float).Mul(new(big.Float).SetInt(total), rate).Int(nil)
			total.Add(total, bonus)
		}
	}
	return total
}

func (a *Account) calcStakeBonusRate() *big.Float {
	total := new(big.Int)
	rate := new(big.Float)
	for _, v := range a.ledger.Stake {
		total = total.Add(total, v.Amount)
	}
	if total.Sign() > 0 {
		ba := a.balanceOfAll(a.ld.Stake.Token)
		if alloc := new(big.Int).Sub(ba, total); alloc.Sign() > 0 {
			return rate.Quo(new(big.Float).SetInt(alloc), new(big.Float).SetInt(total))
		}
	}
	return rate
}

func (a *Account) allocStakeBonus(rate *big.Float) {
	if rate.Sign() > 0 {
		for _, v := range a.ledger.Stake {
			award, _ := new(big.Float).Mul(new(big.Float).SetInt(v.Amount), rate).Int(nil)
			v.Amount.Add(v.Amount, award)
		}
	}
}
