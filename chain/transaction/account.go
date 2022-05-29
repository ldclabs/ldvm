// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package transaction

import (
	"fmt"
	"math"
	"math/big"
	"sync"

	"github.com/ldclabs/ldvm/constants"
	"github.com/ldclabs/ldvm/ld"
	"github.com/ldclabs/ldvm/util"
)

type Account struct {
	ld     *ld.Account
	mu     sync.RWMutex
	id     util.EthID // account address
	pledge *big.Int   // token account and stake account should have pledge
}

func NewAccount(id util.EthID) *Account {
	return &Account{
		id:     id,
		pledge: new(big.Int),
		ld: &ld.Account{
			ID:         util.EthID(id),
			Balance:    big.NewInt(0),
			Keepers:    util.EthIDs{},
			Tokens:     make(map[util.TokenSymbol]*big.Int),
			NonceTable: make(map[uint64][]uint64),
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

func (a *Account) Init(pledge *big.Int, height, timestamp uint64) *Account {
	a.pledge.Set(pledge)
	a.ld.Height = height
	a.ld.Timestamp = timestamp
	return a
}

func (a *Account) ID() util.EthID {
	return a.id
}

func (a *Account) IDBytes() []byte {
	return a.id[:]
}

func (a *Account) Type() ld.AccountType {
	return a.ld.Type
}

func (a *Account) isEmpty() bool {
	return len(a.ld.Keepers) == 0 && a.ld.Balance.Sign() == 0
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
	case t == ld.NativeAccount && a.ld.Balance.Sign() >= 0:
		return true
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
	switch {
	case amount == nil || amount.Sign() < 0:
		return fmt.Errorf(
			"Account.CheckBalance failed: invalid amount %v", amount)
	case amount.Sign() > 0:
		if ba := a.balanceOf(token); amount.Cmp(ba) > 0 {
			return fmt.Errorf(
				"Account.CheckBalance failed: %s has an insufficient %s balance, expected %v, got %v",
				a.id, token.GoString(), amount, a.ld.Balance)
		}
	}
	return nil
}

func (a *Account) CheckAsFrom(txType ld.TxType) error {
	switch a.ld.Type {
	case ld.TokenAccount:
		switch txType {
		case ld.TypeEth, ld.TypeTransfer, ld.TypeUpdateAccountKeepers, ld.TypeDestroyToken:
			// just go ahead
		default:
			return fmt.Errorf(
				"Account.CheckAsFrom failed: can't use TokenAccount as sender for %s",
				txType.String())
		}
	case ld.StakeAccount:
		if a.ld.Stake == nil {
			return fmt.Errorf(
				"Account.CheckAsFrom failed: invalid StakeAccount as sender for %s",
				txType.String())
		}
		ty := a.ld.Stake.Type
		if ty > 2 {
			return fmt.Errorf(
				"Account.CheckAsFrom failed: can't use unknown type %d StakeAccount as sender for %s",
				ty, txType.String())
		}

		// 0: account keepers can not use stake token
		// 1: account keepers can take a stake in other stake account
		// 2: in addition to 1, account keepers can transfer stake token to other account
		switch txType {
		case ld.TypeUpdateAccountKeepers, ld.TypeResetStake:
			// just go ahead
		case ld.TypeTakeStake, ld.TypeWithdrawStake:
			if ty < 1 {
				return fmt.Errorf(
					"Account.CheckAsFrom failed: can't use type %d StakeAccount as sender for %s",
					ty, txType.String())
			}
		case ld.TypeEth, ld.TypeTransfer:
			if ty < 2 {
				return fmt.Errorf(
					"Account.CheckAsFrom failed: can't use type %d StakeAccount as sender for %s",
					ty, txType.String())
			}
		default:
			return fmt.Errorf(
				"Account.CheckAsFrom failed: can't use type %d StakeAccount as sender for %s",
				ty, txType.String())
		}
	}
	return nil
}

func (a *Account) CheckAsTo(txType ld.TxType) error {
	switch a.ld.Type {
	case ld.TokenAccount:
		switch txType {
		case ld.TypeTest, ld.TypeEth, ld.TypeTransfer, ld.TypeExchange, ld.TypeCreateToken:
			// just go ahead
		default:
			return fmt.Errorf(
				"Account.CheckAsTo failed: can't use TokenAccount as recipient for %s",
				txType.String())
		}
	case ld.StakeAccount:
		switch txType {
		case ld.TypeTest, ld.TypeEth, ld.TypeTransfer, ld.TypeCreateStake, ld.TypeTakeStake, ld.TypeWithdrawStake:
			// just go ahead
		default:
			return fmt.Errorf(
				"Account.CheckAsTo failed: can't use StakeAccount as recipient for %s",
				txType.String())
		}
	}
	return nil
}

func (a *Account) Threshold() uint8 {
	a.mu.RLock()
	defer a.mu.RUnlock()

	return a.ld.Threshold
}

func (a *Account) Keepers() util.EthIDs {
	a.mu.RLock()
	defer a.mu.RUnlock()

	return a.ld.Keepers
}

func (a *Account) SatisfySigning(signers util.EthIDs) bool {
	a.mu.RLock()
	defer a.mu.RUnlock()

	switch {
	case a.id == constants.LDCAccount:
		return false
	case a.isEmpty() && signers.Has(a.id):
		return true
	default:
		return util.SatisfySigning(a.ld.Threshold, a.ld.Keepers, signers, false)
	}
}

func (a *Account) SatisfySigningPlus(signers util.EthIDs) bool {
	a.mu.RLock()
	defer a.mu.RUnlock()

	switch {
	case a.id == constants.LDCAccount:
		return false
	case a.isEmpty() && signers.Has(a.id):
		return true
	default:
		return util.SatisfySigningPlus(a.ld.Threshold, a.ld.Keepers, signers)
	}
}

func (a *Account) Add(token util.TokenSymbol, amount *big.Int) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	switch {
	case amount == nil || amount.Sign() < 0:
		return fmt.Errorf(
			"Account.Add failed: invalid amount %v", amount)
	case amount.Sign() > 0:
		switch token {
		case constants.NativeToken:
			a.ld.Balance.Add(a.ld.Balance, amount)
		default:
			v := a.ld.Tokens[token]
			if v == nil {
				v = new(big.Int)
				a.ld.Tokens[token] = v
			}
			v.Add(v, amount)
		}
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
			"Account.SubByNonce failed: invalid nonce for %s, expected %d, got %d",
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
		return fmt.Errorf(
			"Account.SubByNonceTable failed: nonce %d not exists at %d on %s",
			nonce, expire, a.id)
	}

	if err := a.checkBalance(token, amount); err != nil {
		return err
	}

	if write {
		copy(uu[i:], uu[i+1:])
		uu = uu[:len(uu)-1]
		if len(uu) == 0 {
			delete(a.ld.NonceTable, expire)
		} else {
			a.ld.NonceTable[expire] = uu
		}
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
		return fmt.Errorf(
			"Account.CheckNonceTable failed: %s has too many NonceTable groups, expected <= 64",
			a.id)
	}
	us := util.Uint64Set(make(map[uint64]struct{}, len(a.ld.NonceTable[expire])+len(ns)))
	if uu, ok := a.ld.NonceTable[expire]; ok {
		us.Add(uu...)
	}
	for _, u := range ns {
		if us.Has(u) {
			return fmt.Errorf(
				"Account.CheckNonceTable failed: nonce %d exists at %d on %s",
				u, expire, a.id)
		}
		us.Add(u)
	}
	if write {
		a.ld.NonceTable[expire] = us.List()
	}
	return nil
}

func (a *Account) UpdateKeepers(
	threshold uint8,
	keepers []util.EthID,
	approver *util.EthID,
	approveList []ld.TxType,
) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if approver != nil {
		if *approver == util.EthIDEmpty {
			a.ld.Approver = nil
		} else {
			a.ld.Approver = approver
		}
	}
	if approveList != nil {
		a.ld.ApproveList = approveList
	}
	if len(keepers) > 0 {
		a.ld.Threshold = threshold
		a.ld.Keepers = keepers
	}
	return nil
}

func (a *Account) CheckCreateToken(data *ld.TxAccounter) error {
	a.mu.RLock()
	defer a.mu.RUnlock()

	return a.createToken(data, false)
}

func (a *Account) CreateToken(data *ld.TxAccounter) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	return a.createToken(data, true)
}

func (a *Account) createToken(data *ld.TxAccounter, write bool) error {
	token := util.TokenSymbol(a.id)
	if !token.Valid() {
		return fmt.Errorf(
			"Account.CheckCreateToken failed: invalid token %s",
			token.GoString())
	}

	if !a.isEmpty() {
		return fmt.Errorf(
			"Account.CheckCreateToken failed: token account %s exists", token)
	}

	if write {
		a.ld.Type = ld.TokenAccount
		a.ld.MaxTotalSupply = new(big.Int).Set(data.Amount)
		switch token {
		case constants.NativeToken: // NativeToken created by genesis tx
			a.ld.Balance.Set(data.Amount)
		default:
			a.ld.Threshold = data.Threshold
			a.ld.Keepers = data.Keepers
			a.ld.Tokens[token] = new(big.Int).Set(data.Amount)
		}
	}
	return nil
}

func (a *Account) CheckDestroyToken(recipient *Account) error {
	a.mu.RLock()
	defer a.mu.RUnlock()

	return a.destroyToken(recipient, false)
}

func (a *Account) DestroyToken(recipient *Account) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	return a.destroyToken(recipient, true)
}

func (a *Account) destroyToken(recipient *Account, write bool) error {
	token := util.TokenSymbol(a.id)
	if !a.valid(ld.TokenAccount) {
		return fmt.Errorf(
			"Account.CheckDestroyToken failed: invalid token account %s",
			token.GoString())
	}

	tk := a.ld.Tokens[token]
	if tk == nil {
		return fmt.Errorf("Account.CheckDestroyToken failed: invalid token %s",
			token.GoString())
	} else if tk.Cmp(a.ld.MaxTotalSupply) != 0 {
		return fmt.Errorf("Account.CheckDestroyToken failed: some token in the use %v", tk)
	}

	if write {
		recipient.Add(constants.NativeToken, a.ld.Balance)
		a.ld.Type = 0
		a.ld.Balance.SetUint64(0)
		a.ld.Threshold = 0
		a.ld.Keepers = a.ld.Keepers[:0]
		a.ld.MaxTotalSupply = nil
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
		return fmt.Errorf(
			"Account.CheckCreateStake failed: stake account %s exists", a.id)
	}
	if token := util.StakeSymbol(a.id); !token.Valid() {
		return fmt.Errorf(
			"Account.CheckCreateStake failed: invalid stake account %s", token.GoString())
	}
	if write {
		a.ld.Type = ld.StakeAccount
		a.ld.Threshold = acc.Threshold
		a.ld.Keepers = acc.Keepers
		a.ld.Stake = stake
		a.ld.StakeLedger = make(map[util.EthID]*ld.StakeEntry)
		a.ld.MaxTotalSupply = nil
		switch stake.Token {
		case constants.NativeToken:
			a.ld.StakeLedger[from] = &ld.StakeEntry{Amount: new(big.Int).Set(pledge)}
		default:
			if b := a.ld.Tokens[stake.Token]; b == nil {
				a.ld.Tokens[stake.Token] = new(big.Int)
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
	if !a.valid(ld.StakeAccount) {
		return fmt.Errorf(
			"Account.CheckResetStake failed: invalid stake account %s", a.id)
	}
	if stake.Type != a.ld.Stake.Type {
		return fmt.Errorf(
			"Account.CheckResetStake failed: can't change stake type")
	}
	if stake.Token != a.ld.Stake.Token {
		return fmt.Errorf(
			"Account.CheckResetStake failed: can't change stake token")
	}
	if a.ld.Stake.LockTime > a.ld.Timestamp {
		return fmt.Errorf(
			"Account.CheckResetStake failed: stake in lock, please retry after lockTime")
	}
	holders := 0
	for _, v := range a.ld.StakeLedger {
		if v.Amount.Sign() > 0 {
			holders++
		}
	}
	if holders > 1 {
		return fmt.Errorf(
			"Account.CheckResetStake failed: stake holders should not more than 1")
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
	if !a.valid(ld.StakeAccount) {
		return fmt.Errorf(
			"Account.CheckDestroyStake failed: invalid stake account %s", a.id)
	}
	if a.ld.Stake.LockTime > a.ld.Timestamp {
		return fmt.Errorf(
			"Account.CheckDestroyStake failed: stake in lock, please retry after lockTime")
	}

	holders := 0
	for _, v := range a.ld.StakeLedger {
		if v.Amount.Sign() > 0 {
			holders++
		}
	}

	switch holders {
	case 0:
		// just go ahead
	case 1:
		if v, ok := a.ld.StakeLedger[recipient.id]; !ok || v.Amount.Sign() <= 0 {
			return fmt.Errorf(
				"Account.CheckDestroyStake failed: recipient not exists")
		}
	default:
		return fmt.Errorf(
			"Account.CheckDestroyStake failed: stake ledger not empty, please withdraw all except recipient")
	}

	if write {
		recipient.Add(constants.NativeToken, a.ld.Balance)
		a.ld.Balance.SetUint64(0)
		if a.ld.Stake.Token != constants.NativeToken {
			if b, ok := a.ld.Tokens[a.ld.Stake.Token]; ok && b.Sign() > 0 {
				recipient.Add(a.ld.Stake.Token, b)
				b.SetUint64(0)
			}
		}
		a.ld.Type = 0
		a.ld.Threshold = 0
		a.ld.Keepers = a.ld.Keepers[:0]
		a.ld.Stake = nil
		a.ld.StakeLedger = nil
	}
	return nil
}

func (a *Account) CheckTakeStake(token util.TokenSymbol, from util.EthID, amount *big.Int, lockTime uint64) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	return a.takeStake(token, from, amount, lockTime, false)
}

func (a *Account) TakeStake(token util.TokenSymbol, from util.EthID, amount *big.Int, lockTime uint64) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	return a.takeStake(token, from, amount, lockTime, true)
}

func (a *Account) takeStake(token util.TokenSymbol, from util.EthID, amount *big.Int, lockTime uint64, write bool) error {
	if !a.valid(ld.StakeAccount) {
		return fmt.Errorf(
			"Account.CheckTakeStake failed: invalid stake account %s", a.id)
	}

	stake := a.ld.Stake
	if token != stake.Token {
		return fmt.Errorf(
			"Account.CheckTakeStake failed: invalid token, expected %s, got %s",
			stake.Token.GoString(), token.GoString())
	}

	if amount.Cmp(stake.MinAmount) < 0 {
		return fmt.Errorf(
			"Account.CheckTakeStake failed: invalid amount, expected >= %v, got %v",
			stake.MinAmount, amount)
	}

	total := new(big.Int).Set(amount)
	v := a.ld.StakeLedger[from]
	rate := a.calcStakeBonusRate()
	if v != nil {
		bonus, _ := new(big.Float).Mul(new(big.Float).SetInt(v.Amount), rate).Int(nil)
		total.Add(total, v.Amount)
		total.Add(total, bonus)
	}
	if total.Cmp(stake.MaxAmount) > 0 {
		return fmt.Errorf(
			"Account.CheckTakeStake failed: invalid total amount, expected <= %v, got %v",
			stake.MaxAmount, total)
	}

	if write {
		a.allocStakeBonus(rate)
		if v == nil {
			v = &ld.StakeEntry{Amount: new(big.Int)}
			a.ld.StakeLedger[from] = v
		}
		v.Amount.Add(v.Amount, amount)
		if lockTime > 0 {
			v.LockTime = lockTime
		}
	}
	return nil
}

func (a *Account) CheckUpdateStakeApprover(
	from, approver util.EthID,
	signers util.EthIDs,
) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	return a.updateStakeApprover(from, approver, signers, false)
}

func (a *Account) UpdateStakeApprover(
	from, approver util.EthID,
	signers util.EthIDs,
) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	return a.updateStakeApprover(from, approver, signers, true)
}

func (a *Account) updateStakeApprover(
	from, approver util.EthID,
	signers util.EthIDs,
	write bool,
) error {
	if !a.valid(ld.StakeAccount) {
		return fmt.Errorf(
			"Account.CheckUpdateStakeApprover failed: invalid stake account %s", a.id)
	}

	v := a.ld.StakeLedger[from]
	if v == nil {
		return fmt.Errorf(
			"Account.CheckUpdateStakeApprover failed: %s has no stake ledger to update",
			util.EthID(from))
	}
	if v.Approver != nil && !signers.Has(*v.Approver) {
		return fmt.Errorf(
			"Account.CheckUpdateStakeApprover failed: %s need approver signing",
			util.EthID(from))
	}
	if write {
		if approver == util.EthIDEmpty {
			v.Approver = nil
		} else {
			v.Approver = &approver
		}
	}
	return nil
}

func (a *Account) CheckWithdrawStake(
	token util.TokenSymbol,
	from util.EthID,
	signers util.EthIDs,
	amount *big.Int,
) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	_, err := a.withdrawStake(token, from, signers, amount, false)
	return err
}

func (a *Account) WithdrawStake(
	token util.TokenSymbol,
	from util.EthID,
	signers util.EthIDs,
	amount *big.Int,
) (*big.Int, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	return a.withdrawStake(token, from, signers, amount, true)
}

func (a *Account) withdrawStake(
	token util.TokenSymbol,
	from util.EthID,
	signers util.EthIDs,
	amount *big.Int,
	write bool,
) (*big.Int, error) {
	if !a.valid(ld.StakeAccount) {
		return nil, fmt.Errorf(
			"Account.CheckWithdrawStake failed: invalid stake account %s", a.id)
	}

	stake := a.ld.Stake
	if token != stake.Token {
		return nil, fmt.Errorf(
			"Account.CheckWithdrawStake failed: invalid token, expected %s, got %s",
			stake.Token.GoString(), token.GoString())
	}
	if stake.LockTime > a.ld.Timestamp {
		return nil, fmt.Errorf(
			"Account.CheckWithdrawStake failed: stake in lock, please retry after lockTime")
	}

	v := a.ld.StakeLedger[from]
	if v == nil {
		return nil, fmt.Errorf(
			"Account.CheckWithdrawStake failed: %s has no stake to withdraw",
			from)
	}
	if v.LockTime > a.ld.Timestamp {
		return nil, fmt.Errorf(
			"Account.CheckWithdrawStake failed: stake in lock, please retry after lockTime")
	}
	if v.Approver != nil && !signers.Has(*v.Approver) {
		return nil, fmt.Errorf(
			"Account.CheckWithdrawStake failed: %s need approver signing",
			from)
	}
	total := new(big.Int).Set(v.Amount)
	rate := a.calcStakeBonusRate()
	bonus, _ := new(big.Float).Mul(new(big.Float).SetInt(v.Amount), rate).Int(nil)
	total = total.Add(total, bonus)
	if total.Cmp(amount) < 0 {
		return nil, fmt.Errorf(
			"Account.CheckWithdrawStake failed: %s has an insufficient stake to withdraw, expected %v, got %v",
			from, amount, total)
	}

	if ba := a.balanceOf(token); ba.Cmp(amount) < 0 {
		return nil, fmt.Errorf(
			"Account.CheckWithdrawStake failed: %s has an insufficient balance for withdraw, expected %v, got %v",
			util.StakeSymbol(a.id).GoString(), amount, ba)
	}
	if !write {
		return nil, nil
	}

	a.allocStakeBonus(rate)
	v.Amount.Sub(v.Amount, amount)
	if v.Amount.Sign() <= 0 && v.Approver == nil {
		delete(a.ld.StakeLedger, from)
	}
	withdraw := new(big.Int).Mul(amount, new(big.Int).SetUint64(stake.WithdrawFee))
	return withdraw.Sub(amount, withdraw.Quo(withdraw, big.NewInt(1_000_000))), nil
}

func (a *Account) GetStakeAmount(token util.TokenSymbol, from util.EthID) *big.Int {
	total := new(big.Int)
	stake := a.ld.Stake
	if a.valid(ld.StakeAccount) && token == stake.Token {
		if v := a.ld.StakeLedger[from]; v != nil && v.Amount.Sign() > 0 {
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
	for _, v := range a.ld.StakeLedger {
		total = total.Add(total, v.Amount)
	}
	rate := new(big.Float)
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
		return fmt.Errorf(
			"Account.CheckOpenLending failed: lending exists on %s", a.id)
	}

	if write {
		a.ld.Lending = data
		a.ld.LendingLedger = make(map[util.EthID]*ld.LendingEntry)
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
		return fmt.Errorf(
			"Account.CheckCloseLending failed: invalid lending on %s", a.id)
	}

	if len(a.ld.LendingLedger) != 0 {
		return fmt.Errorf(
			"Account.CheckCloseLending failed: please repay all before close")
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
		return fmt.Errorf(
			"Account.CheckBorrow failed: invalid lending on %s", a.id)
	}

	if a.ld.Lending.Token != token {
		return fmt.Errorf(
			"Account.CheckBorrow failed: invalid token, expected %s, got %s",
			a.ld.Lending.Token.GoString(), token.GoString())
	}
	if dueTime > 0 && dueTime <= a.ld.Timestamp {
		return fmt.Errorf(
			"Account.CheckBorrow failed: invalid dueTime, expected > %d, got %d",
			a.ld.Timestamp, dueTime)
	}
	if amount.Cmp(a.ld.Lending.MinAmount) < 0 {
		return fmt.Errorf(
			"Account.CheckBorrow failed: invalid amount, expected >= %v, got %v",
			a.ld.Lending.MinAmount, amount)
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
		return fmt.Errorf(
			"Account.CheckBorrow failed: invalid amount, expected <= %v, got %v",
			a.ld.Lending.MaxAmount, total)
	}
	ba := a.balanceOf(token)
	if ba.Cmp(amount) < 0 {
		return fmt.Errorf(
			"Account.CheckBorrow failed: %s has an insufficient %s balance, expected %v, got %v",
			a.id, token.GoString(), amount, ba)
	}

	if write {
		e.Amount.Set(total)
		e.UpdateAt = a.ld.Timestamp
		e.DueTime = dueTime
		a.ld.LendingLedger[from] = e
	}
	return nil
}

func (a *Account) CheckRepay(token util.TokenSymbol, from util.EthID, amount *big.Int) error {
	a.mu.RLock()
	defer a.mu.RUnlock()

	_, err := a.repay(token, from, amount, false)
	return err
}

func (a *Account) Repay(token util.TokenSymbol, from util.EthID, amount *big.Int) (*big.Int, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	return a.repay(token, from, amount, true)
}

func (a *Account) repay(token util.TokenSymbol, from util.EthID, amount *big.Int, write bool) (*big.Int, error) {
	if a.ld.Lending == nil || a.ld.LendingLedger == nil {
		return nil, fmt.Errorf(
			"Account.CheckRepay failed: invalid lending on %s",
			a.id)
	}

	if a.ld.Lending.Token != token {
		return nil, fmt.Errorf(
			"Account.CheckRepay failed: invalid token, expected %s, got %s",
			a.ld.Lending.Token.GoString(), token.GoString())
	}

	e := a.ld.LendingLedger[from]
	if e == nil {
		return nil, fmt.Errorf("Account.CheckRepay failed: don't need to repay")
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

func (a *Account) Marshal() ([]byte, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	if err := a.ld.SyntacticVerify(); err != nil {
		return nil, err
	}
	return a.ld.Bytes(), nil
}
