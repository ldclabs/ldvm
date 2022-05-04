// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package chain

import (
	"fmt"
	"math"
	"math/big"
	"sync"
	"time"

	"github.com/ldclabs/ldvm/constants"
	"github.com/ldclabs/ldvm/db"
	"github.com/ldclabs/ldvm/ld"
	"github.com/ldclabs/ldvm/util"
)

type Account struct {
	ld     *ld.Account
	mu     sync.RWMutex
	id     util.EthID   // account address
	vdb    *db.PrefixDB // account version database
	pledge *big.Int     // token account and stake account should have pledge
}

func NewAccount(id util.EthID) *Account {
	return &Account{
		id:     id,
		pledge: new(big.Int),
		ld: &ld.Account{
			ID:      util.EthID(id),
			Balance: big.NewInt(0),
			Tokens:  make(map[util.TokenSymbol]*big.Int),
		},
	}
}

func ParseAccount(id util.EthID, data []byte) (*Account, error) {
	a := &Account{id: id, pledge: new(big.Int), ld: &ld.Account{Balance: new(big.Int)}}
	if err := a.ld.Unmarshal(data); err != nil {
		return nil, err
	}
	if err := a.ld.SyntacticVerify(); err != nil {
		return nil, err
	}
	a.ld.ID = id
	return a, nil
}

func (a *Account) Init(vdb *db.PrefixDB, pledge *big.Int, height, timestamp uint64) {
	a.vdb = vdb
	a.pledge.Set(pledge)
	a.ld.Height = height
	a.ld.Timestamp = timestamp
}

func (a *Account) Type() ld.AccountType {
	return a.ld.Type
}

func (a *Account) isEmpty() bool {
	return len(a.ld.Keepers) == 0
}

func (a *Account) Valid(t ld.AccountType) bool {
	a.mu.RLock()
	defer a.mu.RUnlock()

	return a.valid(t)
}

func (a *Account) valid(t ld.AccountType) bool {
	switch {
	case a.ld.Type != t:
		return false
	case a.isEmpty() || a.ld.Balance.Cmp(a.pledge) < 0:
		return false
	case t == ld.TokenAccount && (a.ld.MaxTotalSupply == nil || a.ld.MaxTotalSupply.Sign() <= 0):
		return false
	case t == ld.StakeAccount && (a.ld.Stake == nil || a.ld.StakeLedger == nil):
		return false
	default:
		return true
	}
}

func (a *Account) Nonce() uint64 {
	a.mu.RLock()
	defer a.mu.RUnlock()

	return a.ld.Nonce
}

func (a *Account) balanceOf(token util.TokenSymbol) *big.Int {
	switch token {
	case constants.NativeToken:
		if b := new(big.Int).Sub(a.ld.Balance, a.pledge); b.Sign() >= 0 {
			return b
		}
		return new(big.Int)
	default:
		if v := a.ld.Tokens[token]; v != nil {
			return new(big.Int).Set(v)
		}
		return new(big.Int)
	}
}

func (a *Account) balanceOfAll(token util.TokenSymbol) *big.Int {
	switch token {
	case constants.NativeToken:
		return new(big.Int).Set(a.ld.Balance)
	default:
		if v := a.ld.Tokens[token]; v != nil {
			return new(big.Int).Set(v)
		}
		return new(big.Int)
	}
}

func (a *Account) CheckBalance(token util.TokenSymbol, amount *big.Int) error {
	a.mu.RLock()
	defer a.mu.RUnlock()

	return a.checkBalance(token, amount)
}

func (a *Account) checkBalance(token util.TokenSymbol, amount *big.Int) error {
	if amount.Sign() > 0 {
		if ba := a.balanceOf(token); amount.Cmp(ba) > 0 {
			return fmt.Errorf(
				"Account(%s) token(%s) insufficient balance, expected %v, got %v",
				a.id, token, amount, a.ld.Balance)
		}
	}
	return nil
}

func (a *Account) Threshold() uint8 {
	a.mu.RLock()
	defer a.mu.RUnlock()

	return a.ld.Threshold
}

func (a *Account) Keepers() []util.EthID {
	a.mu.RLock()
	defer a.mu.RUnlock()

	return a.ld.Keepers
}

func (a *Account) SatisfySigning(signers []util.EthID) bool {
	a.mu.RLock()
	defer a.mu.RUnlock()

	return util.SatisfySigning(a.ld.Threshold, a.ld.Keepers, signers, false)
}

func (a *Account) SatisfySigningPlus(signers []util.EthID) bool {
	a.mu.RLock()
	defer a.mu.RUnlock()

	return util.SatisfySigningPlus(a.ld.Threshold, a.ld.Keepers, signers)
}

func (a *Account) Add(token util.TokenSymbol, amount *big.Int) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if amount == nil || amount.Sign() < 0 {
		return fmt.Errorf(
			"Account.Add %s invalid amount %v",
			util.EthID(a.id), amount)
	}
	if amount.Sign() == 0 {
		return nil
	}

	switch token {
	case constants.NativeToken:
		a.ld.Balance.Add(a.ld.Balance, amount)
	default:
		if v := a.ld.Tokens[token]; v != nil {
			v.Add(v, amount)
		}
		a.ld.Tokens[token] = new(big.Int).Set(amount)
	}

	if a.isEmpty() && a.ld.Type == ld.NativeAccount {
		a.ld.Threshold = 1
		a.ld.Keepers = []util.EthID{a.id}
	}
	return nil
}

func (a *Account) Sub(token util.TokenSymbol, amount *big.Int) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if err := a.checkBalance(token, amount); err != nil {
		return err
	}

	a.subNoCheck(token, amount)
	return nil
}

func (a *Account) subNoCheck(token util.TokenSymbol, amount *big.Int) {
	if amount.Sign() > 0 {
		switch token {
		case constants.NativeToken:
			a.ld.Balance.Sub(a.ld.Balance, amount)
		default:
			v := a.ld.Tokens[token]
			v.Sub(v, amount)
		}
	}
}

func (a *Account) SubByNonce(token util.TokenSymbol, nonce uint64, amount *big.Int) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.ld.Nonce != nonce {
		return fmt.Errorf(
			"Account.SubByNonce %s invalid nonce, expected %v, got %v",
			a.id, a.ld.Nonce, nonce)
	}

	if err := a.checkBalance(token, amount); err != nil {
		return err
	}

	a.ld.Nonce++
	a.subNoCheck(token, amount)
	return nil
}

func (a *Account) CheckSubByNonceTable(token util.TokenSymbol, expire, nonce uint64, amount *big.Int) error {
	a.mu.RLock()
	defer a.mu.RUnlock()

	return a.subByNonceTable(token, expire, nonce, amount, false)
}

func (a *Account) SubByNonceTable(token util.TokenSymbol, expire, nonce uint64, amount *big.Int) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	return a.subByNonceTable(token, expire, nonce, amount, true)
}

func (a *Account) subByNonceTable(token util.TokenSymbol, expire, nonce uint64, amount *big.Int, write bool) error {
	uu, ok := a.ld.NonceTable[expire]
	i := -1
	if ok {
		for j, u := range uu {
			if u == nonce {
				i = j
				break
			}
		}
	}
	if i == -1 {
		return fmt.Errorf("Account(%s) NonceTable %d not exists at %d",
			a.id, nonce, expire)
	}

	if err := a.checkBalance(token, amount); err != nil {
		return err
	}

	if write {
		copy(uu[i:], uu[i+1:])
		a.ld.NonceTable[expire] = uu[:len(uu)-1]
		a.subNoCheck(token, amount)
	}
	return nil
}

func (a *Account) CheckNonceTable(expire uint64, ns []uint64) error {
	a.mu.RLock()
	defer a.mu.RUnlock()

	return a.updateNonceTable(expire, ns, false)
}

func (a *Account) AddNonceTable(expire uint64, ns []uint64) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	return a.updateNonceTable(expire, ns, true)
}

func (a *Account) updateNonceTable(expire uint64, ns []uint64, write bool) error {
	if len(a.ld.NonceTable) >= 64 {
		return fmt.Errorf("Account %s NonceTable too many groups, should not more than %d",
			util.EthID(a.id), 64)
	}
	us := util.Uint64Set(make(map[uint64]struct{}, len(a.ld.NonceTable[expire])+len(ns)))
	if uu, ok := a.ld.NonceTable[expire]; ok {
		us.Add(uu...)
	}
	for _, u := range ns {
		if us.Has(u) {
			return fmt.Errorf("Account %s NonceTable %d exists at %d",
				util.EthID(a.id), u, expire)
		}
		us.Add(u)
	}
	if write {
		a.ld.NonceTable[expire] = us.List()
	}
	return nil
}

func (a *Account) UpdateKeepers(threshold uint8, keepers []util.EthID) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.ld.Threshold = threshold
	a.ld.Keepers = keepers
	return nil
}

func (a *Account) CheckCreateToken(token util.TokenSymbol, data *ld.TxAccounter) error {
	a.mu.RLock()
	defer a.mu.RUnlock()

	return a.createToken(token, data, false)
}

func (a *Account) CreateToken(token util.TokenSymbol, data *ld.TxAccounter) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	return a.createToken(token, data, true)
}

func (a *Account) createToken(token util.TokenSymbol, data *ld.TxAccounter, write bool) error {
	if !a.isEmpty() {
		return fmt.Errorf("CreateToken token account %s exists", util.TokenSymbol(a.id))
	}

	if write {
		a.ld.Type = ld.TokenAccount
		a.ld.Threshold = data.Threshold
		a.ld.Keepers = data.Keepers
		a.ld.MaxTotalSupply = data.Amount
		switch token {
		case constants.NativeToken:
			a.ld.Balance.Set(data.Amount)
		default:
			a.ld.Tokens[token] = data.Amount
		}
	}
	return nil
}

func (a *Account) CheckDestroyToken(token util.TokenSymbol, recipient *Account) error {
	a.mu.RLock()
	defer a.mu.RUnlock()

	return a.destroyToken(token, recipient, false)
}

func (a *Account) DestroyToken(token util.TokenSymbol, recipient *Account) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	return a.destroyToken(token, recipient, true)
}

func (a *Account) destroyToken(token util.TokenSymbol, recipient *Account, write bool) error {
	if a.valid(ld.TokenAccount) {
		return fmt.Errorf("Account.DestroyToken invalid token account")
	}

	if a.ld.Tokens[token] == nil || a.ld.Tokens[token].Cmp(a.ld.MaxTotalSupply) != 0 {
		return fmt.Errorf("Account.DestroyToken some token out of account")
	}

	if write {
		recipient.Add(constants.NativeToken, a.ld.Balance)
		a.ld.Balance.SetUint64(0)
		a.ld.Threshold = 0
		a.ld.Keepers = a.ld.Keepers[:0]
		a.ld.MaxTotalSupply.SetInt64(0)
		delete(a.ld.Tokens, token)
	}
	return nil
}

func (a *Account) CheckCreateStake(
	from util.EthID,
	pledge *big.Int,
	acc *ld.TxAccounter,
	stake *ld.StakeConfig,
) error {
	a.mu.RLock()
	defer a.mu.RUnlock()

	return a.createStake(from, pledge, acc, stake, false)
}

func (a *Account) CreateStake(
	from util.EthID,
	pledge *big.Int,
	acc *ld.TxAccounter,
	stake *ld.StakeConfig,
) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	return a.createStake(from, pledge, acc, stake, true)
}

func (a *Account) createStake(
	from util.EthID,
	pledge *big.Int,
	acc *ld.TxAccounter,
	stake *ld.StakeConfig,
	write bool,
) error {
	if !a.isEmpty() {
		return fmt.Errorf("CreateStake stake account %s exists", util.EthID(a.id))
	}
	if a.ld.Stake != nil || a.ld.StakeLedger != nil {
		return fmt.Errorf("Account stake exists: %v, %v", a.ld.Stake, a.ld.StakeLedger)
	}
	if write {
		a.ld.Type = ld.StakeAccount
		a.ld.Threshold = acc.Threshold
		a.ld.Keepers = acc.Keepers
		a.ld.Stake = stake
		a.ld.StakeLedger = make(ld.Ledger)
		a.ld.MaxTotalSupply = nil
		switch stake.TokenID {
		case constants.NativeToken:
			a.ld.StakeLedger[from] = &ld.LedgerEntry{Amount: pledge}
		default:
			if b := a.ld.Tokens[stake.TokenID]; b == nil {
				a.ld.Tokens[stake.TokenID] = new(big.Int)
			}
		}
	}
	return nil
}

func (a *Account) CheckResetStake(stake *ld.StakeConfig) error {
	a.mu.RLock()
	defer a.mu.RUnlock()

	return a.resetStake(stake, false)
}

func (a *Account) ResetStake(stake *ld.StakeConfig) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	return a.resetStake(stake, true)
}

func (a *Account) resetStake(stake *ld.StakeConfig, write bool) error {
	if a.valid(ld.StakeAccount) {
		return fmt.Errorf("Account.ResetStake invalid stake account")
	}
	if stake.Type != a.ld.Stake.Type {
		return fmt.Errorf("Account.ResetStake can not change stake type")
	}
	if stake.Token != a.ld.Stake.Token {
		return fmt.Errorf("Account.ResetStake can not change stake token")
	}
	if a.ld.Stake.LockTime > a.ld.Timestamp {
		return fmt.Errorf("Account.ResetStake stake in lock, please try again after lockTime")
	}
	if len(a.ld.StakeLedger) > 1 {
		return fmt.Errorf("Account.ResetStake stake ledger not empty, please withdraw all except holder")
	}

	if write {
		a.ld.Stake.LockTime = stake.LockTime
		a.ld.Stake.WithdrawFee = stake.WithdrawFee
		if stake.MinAmount.Sign() > 0 {
			a.ld.Stake.MinAmount.Set(stake.MinAmount)
		}
		if stake.MaxAmount.Sign() > 0 {
			a.ld.Stake.MaxAmount.Set(stake.MaxAmount)
		}
	}
	return nil
}

func (a *Account) CheckDestroyStake(recipient *Account) error {
	a.mu.RLock()
	defer a.mu.RUnlock()

	return a.destroyStake(recipient, false)
}

func (a *Account) DestroyStake(recipient *Account) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	return a.destroyStake(recipient, true)
}

func (a *Account) destroyStake(recipient *Account, write bool) error {
	if a.valid(ld.StakeAccount) {
		return fmt.Errorf("Account.DestroyStake invalid stake account")
	}
	if a.ld.Stake.LockTime > a.ld.Timestamp {
		return fmt.Errorf("Account.DestroyStake stake in lock, please retry after %s",
			time.Second*time.Duration(a.ld.Stake.LockTime))
	}

	switch len(a.ld.StakeLedger) {
	case 0:
		// just go ahead
	case 1:
		if _, ok := a.ld.StakeLedger[recipient.id]; !ok {
			return fmt.Errorf("Account.DestroyStake recipient not exists")
		}
	default:
		return fmt.Errorf("Account.DestroyStake stake ledger not empty, please withdraw all except recipient")
	}

	if write {
		recipient.Add(constants.NativeToken, a.ld.Balance)
		a.ld.Balance.SetUint64(0)
		if a.ld.Stake.TokenID != constants.NativeToken {
			if b := a.ld.Tokens[a.ld.Stake.TokenID]; b.Sign() > 0 {
				recipient.Add(a.ld.Stake.TokenID, b)
				b.SetUint64(0)
			}
		}
		a.ld.Threshold = 0
		a.ld.Keepers = a.ld.Keepers[:0]
		a.ld.Stake = nil
		a.ld.StakeLedger = nil
	}
	return nil
}

func (a *Account) CheckTakeStake(token util.TokenSymbol, from util.EthID, amount *big.Int) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	return a.takeStake(token, from, amount, false)
}

func (a *Account) TakeStake(token util.TokenSymbol, from util.EthID, amount *big.Int) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	return a.takeStake(token, from, amount, true)
}

func (a *Account) takeStake(token util.TokenSymbol, from util.EthID, amount *big.Int, write bool) error {
	if a.valid(ld.StakeAccount) {
		return fmt.Errorf("Account.TakeStake invalid stake account")
	}

	stake := a.ld.Stake
	if token != stake.TokenID {
		return fmt.Errorf("Account.TakeStake invalid stake token, expected %s, got %s",
			stake.TokenID, token)
	}

	if amount.Cmp(stake.MinAmount) < 0 {
		return fmt.Errorf("Account.TakeStake invalid stake amount, expected >= %v, got %v",
			stake.MinAmount, amount)
	}

	total := new(big.Int).Set(amount)
	v := a.ld.StakeLedger[from]
	rate := a.calcStakeBonusRate()
	if v != nil {
		bonus, _ := new(big.Float).Mul(new(big.Float).SetInt(v.Amount), rate).Int(nil)
		total = total.Add(total, v.Amount)
		total = total.Add(total, bonus)
	}
	if total.Cmp(stake.MaxAmount) > 0 {
		return fmt.Errorf("Account.TakeStake invalid stake amount when taken, expected <= %v, got %v",
			stake.MaxAmount, total)
	}

	if write {
		a.allocStakeBonus(rate)
		if v == nil {
			v = &ld.LedgerEntry{Amount: new(big.Int)}
			a.ld.StakeLedger[from] = v
		}
		v.Amount.Add(v.Amount, amount)
	}
	return nil
}

func (a *Account) CheckWithdrawStake(token util.TokenSymbol, from util.EthID, amount *big.Int) (*big.Int, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	return a.withdrawStake(token, from, amount, false)
}

func (a *Account) WithdrawStake(token util.TokenSymbol, from util.EthID, amount *big.Int) (*big.Int, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	return a.withdrawStake(token, from, amount, true)
}

func (a *Account) withdrawStake(token util.TokenSymbol, from util.EthID, amount *big.Int, write bool) (*big.Int, error) {
	if a.valid(ld.StakeAccount) {
		return nil, fmt.Errorf("Account.withdrawStake invalid stake account")
	}

	stake := a.ld.Stake
	if token != stake.TokenID {
		return nil, fmt.Errorf("Account.withdrawStake invalid stake token, expected %s, got %s",
			util.TokenSymbol(stake.TokenID), util.TokenSymbol(token))
	}
	if stake.LockTime > a.ld.Timestamp {
		return nil, fmt.Errorf("Account.WithdrawStake stake in lock, please retry after lockTime")
	}

	v := a.ld.StakeLedger[from]
	if v == nil {
		return nil, fmt.Errorf("Account.WithdrawStake %s no stake to withdraw",
			util.EthID(from))
	}
	total := new(big.Int).Set(v.Amount)
	rate := a.calcStakeBonusRate()
	bonus, _ := new(big.Float).Mul(new(big.Float).SetInt(v.Amount), rate).Int(nil)
	total = total.Add(total, bonus)
	if total.Cmp(amount) < 0 {
		return nil, fmt.Errorf("Account.WithdrawStake %s insufficient stake to withdraw, expected %v, got %v",
			util.EthID(from), amount, total)
	}

	ba := a.balanceOf(token)
	if ba.Cmp(amount) < 0 {
		return nil, fmt.Errorf("Account.WithdrawStake %s insufficient balance to withdraw, expected %v, got %v",
			util.EthID(a.id), amount, ba)
	}

	if write {
		a.allocStakeBonus(rate)
		v.Amount.Sub(v.Amount, amount)
	}
	withdraw := new(big.Int).Mul(amount, big.NewInt(1_000_000-int64(stake.WithdrawFee)))
	return withdraw.Quo(withdraw, big.NewInt(1_000_000)), nil
}

func (a *Account) calcStakeBonusRate() *big.Float {
	total := new(big.Int)
	for _, v := range a.ld.StakeLedger {
		total = total.Add(total, v.Amount)
	}
	rate := new(big.Float)
	if total.Sign() > 0 {
		ba := a.balanceOfAll(a.ld.Stake.TokenID)
		if alloc := new(big.Int).Sub(ba, total); alloc.Sign() > 0 {
			return rate.Quo(new(big.Float).SetInt(alloc), new(big.Float).SetInt(total))
		}
	}
	return rate
}

func (a *Account) allocStakeBonus(rate *big.Float) {
	if rate.Sign() > 0 {
		for _, v := range a.ld.StakeLedger {
			award, _ := new(big.Float).Mul(new(big.Float).SetInt(v.Amount), rate).Int(nil)
			v.Amount.Add(v.Amount, award)
		}
	}
}

func (a *Account) CheckOpenLending(data *ld.LendingConfig) error {
	a.mu.RLock()
	defer a.mu.RUnlock()

	return a.openLending(data, false)
}

func (a *Account) OpenLending(data *ld.LendingConfig) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	return a.openLending(data, true)
}

func (a *Account) openLending(data *ld.LendingConfig, write bool) error {
	if a.ld.Lending != nil || a.ld.LendingLedger != nil {
		return fmt.Errorf("Account lending exists: %v, %v", a.ld.Lending, a.ld.LendingLedger)
	}

	if write {
		a.ld.Lending = data
		a.ld.LendingLedger = make(ld.Ledger)
	}
	return nil
}

func (a *Account) CheckCloseLending() error {
	a.mu.RLock()
	defer a.mu.RUnlock()

	return a.closeLending(false)
}

func (a *Account) CloseLending() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	return a.closeLending(true)
}

func (a *Account) closeLending(write bool) error {
	if a.ld.Lending == nil || a.ld.LendingLedger == nil {
		return fmt.Errorf("Account invalid lending: %v, %v", a.ld.Lending, a.ld.LendingLedger)
	}

	if len(a.ld.LendingLedger) != 0 {
		return fmt.Errorf(" please repay all before close")
	}

	if write {
		a.ld.Lending = nil
		a.ld.LendingLedger = nil
	}
	return nil
}

func (a *Account) CheckBorrow(token util.TokenSymbol, from util.EthID, amount *big.Int, dueTime uint64) error {
	a.mu.RLock()
	defer a.mu.RUnlock()

	return a.borrow(token, from, amount, dueTime, false)
}

func (a *Account) Borrow(token util.TokenSymbol, from util.EthID, amount *big.Int, dueTime uint64) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	return a.borrow(token, from, amount, dueTime, true)
}

func (a *Account) borrow(token util.TokenSymbol, from util.EthID, amount *big.Int, dueTime uint64, write bool) error {
	if a.ld.Lending == nil || a.ld.LendingLedger == nil {
		return fmt.Errorf("Account invalid lending: %v, %v", a.ld.Lending, a.ld.LendingLedger)
	}

	if a.ld.Lending.TokenID != token {
		return fmt.Errorf("Account invalid lending token, expected %s, got %s",
			a.ld.Lending.Token, util.TokenSymbol(token).String())
	}
	if amount.Cmp(a.ld.Lending.MinAmount) < 0 {
		return fmt.Errorf("Account invalid lending amount, expected >= %v, got %v",
			a.ld.Lending.MinAmount, amount)
	}

	e := a.ld.LendingLedger[from]
	total := new(big.Int).Set(amount)
	switch {
	case e == nil:
		e = &ld.LedgerEntry{Amount: amount}
	default:
		total = total.Add(total, a.calcBorrowTotal(from))
	}

	if total.Cmp(a.ld.Lending.MaxAmount) > 0 {
		return fmt.Errorf("Account invalid lending amount, expected <= %v, got %v",
			a.ld.Lending.MaxAmount, total)
	}
	ba := a.balanceOf(token)
	if ba.Cmp(amount) < 0 {
		return fmt.Errorf("Account.Borrow %s insufficient balance, expected %v, got %v",
			util.EthID(a.id), amount, ba)
	}

	if write {
		e.Amount.Set(total)
		e.UpdateAt = a.ld.Timestamp
		e.DueTime = dueTime
		a.ld.LendingLedger[from] = e
	}
	return nil
}

func (a *Account) CheckRepay(token util.TokenSymbol, from util.EthID, amount *big.Int) (*big.Int, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	return a.repay(token, from, amount, false)
}

func (a *Account) Repay(token util.TokenSymbol, from util.EthID, amount *big.Int) (*big.Int, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	return a.repay(token, from, amount, true)
}

func (a *Account) repay(token util.TokenSymbol, from util.EthID, amount *big.Int, write bool) (*big.Int, error) {
	if a.ld.Lending == nil || a.ld.LendingLedger == nil {
		return nil, fmt.Errorf("Account invalid lending: %v, %v", a.ld.Lending, a.ld.LendingLedger)
	}

	if a.ld.Lending.TokenID != token {
		return nil, fmt.Errorf("Account invalid lending token, expected %s, got %s",
			a.ld.Lending.Token, util.TokenSymbol(token).String())
	}

	e := a.ld.LendingLedger[from]
	if e == nil {
		return nil, fmt.Errorf("Account don't need to repay")
	}

	total := a.calcBorrowTotal(from)
	actual := new(big.Int).Set(amount)
	cleanup := amount.Cmp(total) >= 0
	if cleanup {
		actual.Set(total)
	}
	if write {
		if cleanup {
			delete(a.ld.LendingLedger, from)
		} else {
			e.Amount.Sub(total, amount)
			e.UpdateAt = a.ld.Timestamp
			a.ld.LendingLedger[from] = e
		}
	}
	return actual, nil
}

const day = 3600 * 24

func (a *Account) calcBorrowTotal(from util.EthID) *big.Int {
	cfg := a.ld.Lending
	e := a.ld.LendingLedger[from]
	amount := new(big.Int).Set(e.Amount)
	if sec := a.ld.Timestamp - e.UpdateAt; sec > 0 && amount.Sign() > 0 {
		var rate float64
		switch {
		case e.DueTime == 0 || a.ld.Timestamp <= e.DueTime:
			rate = math.Pow(1+float64(cfg.DailyInterest)/1_000_000, float64(sec)/day)
		case e.UpdateAt >= e.DueTime:
			rate = math.Pow(1+float64(cfg.DailyInterest+cfg.OverdueInterest)/1_000_000, float64(sec)/day)
		default:
			rate = math.Pow(1+float64(cfg.DailyInterest)/1_000_000, float64(e.DueTime-e.UpdateAt)/day)
			rate = math.Pow(rate+float64(cfg.DailyInterest+cfg.OverdueInterest)/1_000_000, float64(a.ld.Timestamp-e.DueTime)/day)
		}

		amount, _ = new(big.Float).Mul(new(big.Float).SetInt(amount), big.NewFloat(rate)).Int(nil)
	}
	return amount
}

// Commit will be called when blockState.SaveBlock
func (a *Account) Commit() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if err := a.ld.SyntacticVerify(); err != nil {
		return err
	}
	return a.vdb.Put(a.id[:], a.ld.Bytes())
}
