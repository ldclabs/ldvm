// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package acct

import (
	"fmt"
	"math/big"

	"github.com/fxamacker/cbor/v2"
	"github.com/ldclabs/ldvm/ids"
	"github.com/ldclabs/ldvm/ld"
	"github.com/ldclabs/ldvm/signer"
	"github.com/ldclabs/ldvm/util/erring"
)

func (a *Account) CreateStake(
	from ids.Address,
	pledge *big.Int,
	acc *ld.TxAccounter,
	cfg *ld.StakeConfig,
) error {
	errp := erring.ErrPrefix(fmt.Sprintf("acct.Account(%s).CreateStake: ", a.ld.ID.String()))

	a.mu.Lock()
	defer a.mu.Unlock()

	stake := ids.StakeSymbol(a.ld.ID)
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

	if acc.Approver != nil && acc.Approver.Valid() == nil {
		a.ld.Approver = *acc.Approver
	}
	if acc.ApproveList != nil {
		a.ld.ApproveList = *acc.ApproveList
	}
	a.ld.Stake = cfg
	a.ld.MaxTotalSupply = nil
	switch cfg.Token {
	case ids.NativeToken:
		a.ledger.Stake[from.AsKey()] = &ld.StakeEntry{Amount: new(big.Int).Set(pledge)}
	default:
		if b := a.ld.Tokens[cfg.Token.AsKey()]; b == nil {
			a.ld.Tokens[cfg.Token.AsKey()] = new(big.Int)
		}
	}

	return nil
}

func (a *Account) ResetStake(cfg *ld.StakeConfig) error {
	errp := erring.ErrPrefix(fmt.Sprintf("acct.Account(%s).ResetStake: ", a.ld.ID.String()))

	a.mu.Lock()
	defer a.mu.Unlock()

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
	errp := erring.ErrPrefix(fmt.Sprintf("acct.Account(%s).DestroyStake: ", a.ld.ID.String()))

	a.mu.Lock()
	defer a.mu.Unlock()

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
		if v, ok := a.ledger.Stake[recipient.ID().AsKey()]; !ok || v.Amount.Sign() <= 0 {
			return errp.Errorf("recipient not exists")
		}

	default:
		return errp.Errorf("stake ledger not empty, please withdraw all except recipient")
	}

	if err := a.closeLending(true); err != nil {
		return errp.ErrorIf(err)
	}

	recipient.Add(ids.NativeToken, a.ld.Balance)
	a.ld.Balance.SetUint64(0)
	if a.ld.Stake.Token != ids.NativeToken {
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
	a.ledger.Stake = make(map[cbor.ByteString]*ld.StakeEntry)
	return nil
}

func (a *Account) TakeStake(
	token ids.TokenSymbol,
	from ids.Address,
	amount *big.Int,
	lockTime uint64) error {
	errp := erring.ErrPrefix(fmt.Sprintf("acct.Account(%s).TakeStake: ", a.ld.ID.String()))

	a.mu.Lock()
	defer a.mu.Unlock()

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
	from ids.Address,
	approver signer.Key,
	txIsApprovedFn ld.TxIsApprovedFn,
) error {
	errp := erring.ErrPrefix(fmt.Sprintf("acct.Account(%s).UpdateStakeApprover: ", a.ld.ID.String()))

	a.mu.Lock()
	defer a.mu.Unlock()

	if !a.valid(ld.StakeAccount) {
		return errp.Errorf("invalid stake account")
	}
	if a.ledger == nil {
		return errp.Errorf("invalid ledger")
	}

	v := a.ledger.Stake[from.AsKey()]
	if v == nil {
		return errp.Errorf("%s has no stake ledger to update", ids.Address(from))
	}

	if v.Approver != nil && !txIsApprovedFn(*v.Approver, nil, false) {
		return errp.Errorf("%s need approver signing", ids.Address(from))
	}

	if len(approver) == 0 {
		v.Approver = nil
	} else {
		v.Approver = &approver
	}
	return nil
}

func (a *Account) WithdrawStake(
	token ids.TokenSymbol,
	from ids.Address,
	amount *big.Int,
	txIsApprovedFn ld.TxIsApprovedFn,
) (*big.Int, error) {
	errp := erring.ErrPrefix(fmt.Sprintf("acct.Account(%s).WithdrawStake: ", a.ld.ID.String()))

	a.mu.Lock()
	defer a.mu.Unlock()

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
	if v.Approver != nil && !txIsApprovedFn(*v.Approver, nil, false) {
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

	if err := a.checkBalance(token, amount, true); err != nil {
		return nil, err
	}

	a.allocStakeBonus(rate)
	v.Amount.Sub(v.Amount, amount)
	if v.Amount.Sign() <= 0 && v.Approver == nil {
		delete(a.ledger.Stake, from.AsKey())
	}
	withdraw := new(big.Int).Mul(amount, new(big.Int).SetUint64(stake.WithdrawFee))
	return withdraw.Sub(amount, withdraw.Quo(withdraw, big.NewInt(1_000_000))), nil
}

func (a *Account) GetStakeAmount(token ids.TokenSymbol, from ids.Address) *big.Int {
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
