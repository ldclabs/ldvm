// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package chain

import (
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/ava-labs/avalanchego/ids"

	"github.com/ldclabs/ldvm/constants"
	"github.com/ldclabs/ldvm/db"
	"github.com/ldclabs/ldvm/ld"
	"github.com/ldclabs/ldvm/util"
)

type Account struct {
	ld  *ld.Account
	mu  sync.RWMutex
	id  ids.ShortID  // account address
	vdb *db.PrefixDB // account version database
}

func NewAccount(id ids.ShortID) *Account {
	return &Account{
		id: id,
		ld: &ld.Account{
			ID:      id,
			Balance: big.NewInt(0),
			Ledger:  make(map[ids.ShortID]*big.Int),
		},
	}
}

func ParseAccount(id ids.ShortID, data []byte) (*Account, error) {
	a := &Account{id: id, ld: &ld.Account{Balance: new(big.Int)}}
	if err := a.ld.Unmarshal(data); err != nil {
		return nil, err
	}
	if err := a.ld.SyntacticVerify(); err != nil {
		return nil, err
	}
	a.ld.ID = id
	return a, nil
}

func (a *Account) Init(vdb *db.PrefixDB) {
	a.vdb = vdb
}

func (a *Account) Type() ld.AccountType {
	return a.ld.Type
}

func (a *Account) IsEmpty() bool {
	a.mu.RLock()
	defer a.mu.RUnlock()

	return a.isEmpty()
}

func (a *Account) isEmpty() bool {
	return len(a.ld.Keepers) == 0
}

func (a *Account) ValidStake(minValidatorStake *big.Int) bool {
	a.mu.RLock()
	defer a.mu.RUnlock()

	return !a.isEmpty() && a.ld.Type == ld.StakeAccount && a.ld.Balance.Cmp(minValidatorStake) >= 0
}

func (a *Account) Nonce() uint64 {
	a.mu.RLock()
	defer a.mu.RUnlock()

	return a.ld.Nonce
}

func (a *Account) BalanceOf(token ids.ShortID) *big.Int {
	a.mu.RLock()
	defer a.mu.RUnlock()

	switch token {
	case constants.LDCAccount:
		return new(big.Int).Set(a.ld.Balance)
	default:
		if v := a.ld.Ledger[token]; v != nil {
			return new(big.Int).Set(v)
		}
		return new(big.Int)
	}
}

func (a *Account) Threshold() uint8 {
	a.mu.RLock()
	defer a.mu.RUnlock()

	return a.ld.Threshold
}

func (a *Account) Keepers() []ids.ShortID {
	a.mu.RLock()
	defer a.mu.RUnlock()

	return a.ld.Keepers
}

func (a *Account) SatisfySigning(signers []ids.ShortID) bool {
	a.mu.RLock()
	defer a.mu.RUnlock()

	return util.SatisfySigning(a.ld.Threshold, a.ld.Keepers, signers, false)
}

func (a *Account) Add(token ids.ShortID, amount *big.Int) error {
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
	case constants.LDCAccount:
		a.ld.Balance.Add(a.ld.Balance, amount)
	default:
		if v := a.ld.Ledger[token]; v != nil {
			v.Add(v, amount)
		}
		a.ld.Ledger[token] = new(big.Int).Set(amount)
	}

	if a.isEmpty() && a.ld.Type == ld.NativeAccount {
		a.ld.Threshold = 1
		a.ld.Keepers = []ids.ShortID{a.id}
	}
	return nil
}

func (a *Account) Sub(token ids.ShortID, amount *big.Int) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	return a.sub(token, amount)
}

func (a *Account) sub(token ids.ShortID, amount *big.Int) error {
	if amount == nil || amount.Sign() < 0 {
		return fmt.Errorf(
			"Account.Sub %s invalid amount %v",
			util.EthID(a.id), amount)
	}
	if amount.Sign() == 0 {
		return nil
	}

	switch token {
	case constants.LDCAccount:
		if amount.Cmp(a.ld.Balance) > 0 {
			return fmt.Errorf(
				"Account.Sub %s insufficient balance %v",
				util.EthID(a.id), a.ld.Balance)
		}
		a.ld.Balance.Sub(a.ld.Balance, amount)
	default:
		v := a.ld.Ledger[token]
		if v == nil || amount.Cmp(v) > 0 {
			return fmt.Errorf(
				"Account.Sub %s, %s insufficient balance %v",
				util.EthID(a.id), token.String(), v)
		}
		v.Sub(v, amount)
	}
	return nil
}

func (a *Account) SubByNonce(token ids.ShortID, nonce uint64, amount *big.Int) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.ld.Nonce != nonce {
		return fmt.Errorf(
			"Account.SubByNonce %s invalid nonce, expected %v, got %v",
			util.EthID(a.id), a.ld.Nonce, nonce)
	}

	if err := a.sub(token, amount); err != nil {
		return err
	}
	a.ld.Nonce++
	return nil
}

func (a *Account) SubByNonceTable(token ids.ShortID, expire, nonce uint64, amount *big.Int) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if err := a.nonceTableSpent(expire, nonce, true); err != nil {
		return fmt.Errorf(
			"Account.SubByNonceTable %s: %v", util.EthID(a.id), err)
	}
	if err := a.sub(token, amount); err != nil {
		return fmt.Errorf(
			"Account.SubByNonceTable %s: %v", util.EthID(a.id), err)
	}
	return nil
}

func (a *Account) NonceTableHas(expire uint64, nonce uint64) error {
	a.mu.RLock()
	defer a.mu.RUnlock()

	return a.nonceTableSpent(expire, nonce, false)
}

func (a *Account) nonceTableSpent(expire uint64, nonce uint64, update bool) error {
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
		return fmt.Errorf("Account %s NonceTable %d not exists at %d",
			util.EthID(a.id), nonce, expire)
	}

	if update {
		copy(uu[i:], uu[i+1:])
		a.ld.NonceTable[expire] = uu[:len(uu)-1]
	}
	return nil
}

func (a *Account) NonceTableValid(expire uint64, ns []uint64) error {
	a.mu.RLock()
	defer a.mu.RUnlock()

	return a.nonceTableValidAndUpdate(expire, ns, false)
}

func (a *Account) NonceTableAdd(expire uint64, ns []uint64) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	return a.nonceTableValidAndUpdate(expire, ns, true)
}

func (a *Account) nonceTableValidAndUpdate(expire uint64, ns []uint64, update bool) error {
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
	if update {
		a.ld.NonceTable[expire] = us.List()
	}
	return nil
}

func (a *Account) UpdateKeepers(threshold uint8, keepers []ids.ShortID) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.ld.Threshold = threshold
	a.ld.Keepers = keepers
	return nil
}

func (a *Account) CreateToken(token ids.ShortID, data *ld.TxMinter) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if !a.isEmpty() {
		return fmt.Errorf("CreateToken token account %s exists", util.EthID(a.id).String())
	}

	a.ld.Type = ld.TokenAccount
	a.ld.Threshold = data.Threshold
	a.ld.Keepers = data.Keepers
	a.ld.LockTime = 0
	a.ld.DelegationFee = 0
	a.ld.MaxTotalSupply = data.Amount
	if token != constants.LDCAccount {
		a.ld.Ledger[token] = data.Amount
	}
	return nil
}

func (a *Account) DestroyToken(token ids.ShortID, recipient ids.ShortID) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.isEmpty() || a.ld.Type != ld.TokenAccount {
		return fmt.Errorf("Account.DestroyToken invalid token account")
	}

	if a.ld.Ledger[token] == nil || a.ld.Ledger[token].Cmp(a.ld.MaxTotalSupply) != 0 {
		return fmt.Errorf("Account.DestroyToken some token out of account")
	}

	delete(a.ld.Ledger, token)
	a.ld.Threshold = 0
	a.ld.MaxTotalSupply.SetInt64(0)
	a.ld.Keepers = a.ld.Keepers[:0]
	return nil
}

func (a *Account) CreateStake(from ids.ShortID, amount *big.Int, data *ld.TxMinter) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if !a.isEmpty() {
		return fmt.Errorf("CreateStake stake account %s exists", util.EthID(a.id))
	}

	a.ld.Type = ld.StakeAccount
	a.ld.Threshold = data.Threshold
	a.ld.Keepers = data.Keepers
	a.ld.LockTime = data.LockTime
	a.ld.DelegationFee = data.DelegationFee
	a.ld.MaxTotalSupply = nil
	a.ld.Ledger[from] = amount
	return nil
}

func (a *Account) ResetStake(holder ids.ShortID, data *ld.TxMinter) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.isEmpty() || a.ld.Type != ld.StakeAccount {
		return fmt.Errorf("Account.ResetStake invalid stake account")
	}
	if a.ld.LockTime > uint64(time.Now().Unix()) {
		return fmt.Errorf("Account.ResetStake stake in lock, please retry after %s",
			time.Second*time.Duration(a.ld.LockTime))
	}
	if _, ok := a.ld.Ledger[holder]; !ok {
		return fmt.Errorf("Account.ResetStake holder not exists")
	}
	if len(a.ld.Ledger) > 1 {
		return fmt.Errorf("Account.ResetStake stake not empty, please withdraw all except holder")
	}

	if data.Threshold > 0 {
		a.ld.Threshold = data.Threshold
	}
	if len(data.Keepers) > 0 {
		a.ld.Keepers = data.Keepers
	}

	a.ld.LockTime = data.LockTime
	a.ld.DelegationFee = data.DelegationFee
	return nil
}

func (a *Account) DestroyStake(recipient ids.ShortID) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.isEmpty() || a.ld.Type != ld.StakeAccount {
		return fmt.Errorf("Account.DestroyStake invalid stake account")
	}
	if a.ld.LockTime > uint64(time.Now().Unix()) {
		return fmt.Errorf("Account.DestroyStake stake in lock, please retry after %s",
			time.Second*time.Duration(a.ld.LockTime))
	}

	if _, ok := a.ld.Ledger[recipient]; ok {
		delete(a.ld.Ledger, recipient)
	}
	if len(a.ld.Ledger) > 0 {
		return fmt.Errorf("Account.DestroyStake stake not empty, please withdraw all")
	}

	a.ld.Threshold = 0
	a.ld.Keepers = a.ld.Keepers[:0]
	return nil
}

func (a *Account) TakeStake(from ids.ShortID, amount, limit *big.Int) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.isEmpty() || a.ld.Type != ld.StakeAccount {
		return fmt.Errorf("Account.TakeStake invalid stake account")
	}

	if amount.Sign() == 0 {
		return nil
	}

	if err := a.allocStake(); err != nil {
		return err
	}

	v := a.ld.Ledger[from]
	if v == nil {
		v = new(big.Int)
	}
	a.ld.Ledger[from] = v.Add(v, amount)
	if v.Cmp(limit) > 0 {
		return fmt.Errorf("Account.TakeStake amount exceed")
	}
	return nil
}

func (a *Account) WithdrawStake(from ids.ShortID, amount *big.Int) (*big.Int, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.isEmpty() || a.ld.Type != ld.StakeAccount {
		return nil, fmt.Errorf("Account.WithdrawStake invalid stake account")
	}
	if a.ld.LockTime > uint64(time.Now().Unix()) {
		return nil, fmt.Errorf("Account.WithdrawStake stake in lock, please retry after %s",
			time.Second*time.Duration(a.ld.LockTime))
	}

	if amount.Sign() == 0 {
		return new(big.Int), nil
	}

	if err := a.allocStake(); err != nil {
		return nil, err
	}

	v := a.ld.Ledger[from]
	if v == nil || v.Cmp(amount) < 0 {
		return nil, fmt.Errorf("Account.WithdrawStake %s insufficient balance to withdraw, expected %v, got %v",
			util.EthID(from), amount, v)
	}
	a.ld.Ledger[from] = v.Sub(v, amount)
	withdraw := new(big.Int).Mul(amount, big.NewInt(1000-int64(a.ld.DelegationFee)))
	return withdraw.Quo(withdraw, big.NewInt(1000)), nil
}

func (a *Account) allocStake() error {
	total := new(big.Int)
	for _, v := range a.ld.Ledger {
		total = total.Add(total, v)
	}
	if total.Sign() <= 0 {
		return fmt.Errorf("Account.allocStake invalid ledger")
	}
	alloc := new(big.Int).Sub(a.ld.Balance, total)
	if alloc.Sign() < 0 {
		return fmt.Errorf("Account.allocStake invalid ledger")
	}

	if alloc.Sign() > 0 {
		ratio := new(big.Float).Quo(new(big.Float).SetInt(alloc), new(big.Float).SetInt(total))
		for _, v := range a.ld.Ledger {
			award, _ := new(big.Float).Mul(new(big.Float).SetInt(v), ratio).Int(nil)
			v.Add(v, award)
		}
	}
	return nil
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
