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

type AccountCache map[util.EthID]*Account

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
	case t == ld.NativeAccount && a.ld.Balance.Sign() >= 0:
		return true
	case (a.isEmpty() && a.id != util.EthIDEmpty) || a.ld.Balance.Cmp(a.pledge) < 0:
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
			"Account(%s).CheckBalance failed: invalid amount %v", a.id, amount)
	case amount.Sign() > 0:
		if ba := a.balanceOf(token); amount.Cmp(ba) > 0 {
			return fmt.Errorf(
				"Account(%s).CheckBalance failed: insufficient %s balance, expected %v, got %v",
				a.id, token.GoString(), amount, ba)
		}
	}
	return nil
}

func (a *Account) CheckAsFrom(txType ld.TxType) error {
	switch a.ld.Type {
	case ld.TokenAccount:
		switch {
		case ld.TokenFromTxTypes.Has(txType):
			// just go ahead
		default:
			return fmt.Errorf(
				"Account(%s).CheckAsFrom failed: can't use TokenAccount as sender for %s",
				a.id, txType.String())
		}
	case ld.StakeAccount:
		if a.ld.Stake == nil {
			return fmt.Errorf(
				"Account(%s).CheckAsFrom failed: invalid StakeAccount as sender for %s",
				a.id, txType.String())
		}
		ty := a.ld.Stake.Type
		if ty > 2 {
			return fmt.Errorf(
				"Account(%s).CheckAsFrom failed: can't use unknown type %d StakeAccount as sender for %s",
				a.id, ty, txType.String())
		}

		// 0: account keepers can not use stake token
		// 1: account keepers can take a stake in other stake account
		// 2: in addition to 1, account keepers can transfer stake token to other account
		switch {
		case ld.StakeFromTxTypes0.Has(txType):
			// just go ahead
		case ld.StakeFromTxTypes1.Has(txType):
			if ty < 1 {
				return fmt.Errorf(
					"Account(%s).CheckAsFrom failed: can't use type %d StakeAccount as sender for %s",
					a.id, ty, txType.String())
			}
		case ld.StakeFromTxTypes2.Has(txType):
			if ty < 2 {
				return fmt.Errorf(
					"Account(%s).CheckAsFrom failed: can't use type %d StakeAccount as sender for %s",
					a.id, ty, txType.String())
			}
		default:
			return fmt.Errorf(
				"Account(%s).CheckAsFrom failed: can't use type %d StakeAccount as sender for %s",
				a.id, ty, txType.String())
		}
	}
	return nil
}

func (a *Account) CheckAsTo(txType ld.TxType) error {
	switch a.ld.Type {
	case ld.TokenAccount:
		switch {
		case ld.TokenToTxTypes.Has(txType):
			// just go ahead
		default:
			return fmt.Errorf(
				"Account(%s).CheckAsTo failed: can't use TokenAccount as recipient for %s",
				a.id, txType.String())
		}
	case ld.StakeAccount:
		switch {
		case ld.StakeToTxTypes.Has(txType):
			// just go ahead
		default:
			return fmt.Errorf(
				"Account(%s).CheckAsTo failed: can't use StakeAccount as recipient for %s",
				a.id, txType.String())
		}
	}
	return nil
}

func (a *Account) Threshold() uint16 {
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
			"Account(%s).Add failed: invalid amount %v", a.id, amount)
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

func (a *Account) SubByNonce(
	token util.TokenSymbol,
	nonce uint64,
	amount *big.Int,
) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.ld.Nonce != nonce {
		return fmt.Errorf(
			"Account(%s).SubByNonce failed: invalid nonce, expected %d, got %d",
			a.id, a.ld.Nonce, nonce)
	}

	if err := a.checkBalance(token, amount); err != nil {
		return err
	}

	a.ld.Nonce++
	a.subNoCheck(token, amount)
	return nil
}

func (a *Account) CheckSubByNonceTable(
	token util.TokenSymbol,
	expire, nonce uint64,
	amount *big.Int) error {
	a.mu.RLock()
	defer a.mu.RUnlock()

	return a.subByNonceTable(token, expire, nonce, amount, false)
}

func (a *Account) SubByNonceTable(
	token util.TokenSymbol,
	expire, nonce uint64,
	amount *big.Int) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	return a.subByNonceTable(token, expire, nonce, amount, true)
}

func (a *Account) subByNonceTable(
	token util.TokenSymbol,
	expire, nonce uint64,
	amount *big.Int,
	write bool) error {
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
			"Account(%s).SubByNonceTable failed: nonce %d not exists at %d",
			a.id, nonce, expire)
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
			"Account(%s).CheckNonceTable failed: too many NonceTable groups, expected <= 64",
			a.id)
	}
	us := util.Uint64Set(make(map[uint64]struct{}, len(a.ld.NonceTable[expire])+len(ns)))
	if uu, ok := a.ld.NonceTable[expire]; ok {
		us.Add(uu...)
	}
	for _, u := range ns {
		if us.Has(u) {
			return fmt.Errorf(
				"Account(%s).CheckNonceTable failed: nonce %d exists at %d",
				a.id, u, expire)
		}
		us.Add(u)
	}
	if write {
		a.ld.NonceTable[expire] = us.List()
		// clear expired nonces
		for e := range a.ld.NonceTable {
			if e < a.ld.Timestamp {
				delete(a.ld.NonceTable, e)
			}
		}
	}
	return nil
}

func (a *Account) UpdateKeepers(
	threshold *uint16,
	keepers *util.EthIDs,
	approver *util.EthID,
	approveList ld.TxTypes,
) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if approver != nil {
		if *approver == util.EthIDEmpty {
			a.ld.Approver = nil
			a.ld.ApproveList = nil
		} else {
			a.ld.Approver = approver
		}
	}
	if approveList != nil {
		a.ld.ApproveList = approveList
	}
	if threshold != nil && keepers != nil {
		a.ld.Threshold = *threshold
		a.ld.Keepers = *keepers
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
			"Account(%s).CheckCreateToken failed: invalid token %s",
			a.id, token.GoString())
	}

	if !a.isEmpty() {
		return fmt.Errorf(
			"Account(%s).CheckCreateToken failed: token account %s exists",
			a.id, token)
	}

	if write {
		a.ld.Type = ld.TokenAccount
		a.ld.MaxTotalSupply = new(big.Int).Set(data.Amount)
		switch token {
		case constants.NativeToken: // NativeToken created by genesis tx
			a.ld.Balance.Set(data.Amount)
		default:
			a.ld.Threshold = *data.Threshold
			a.ld.Keepers = *data.Keepers
			a.ld.Approver = data.Approver
			a.ld.ApproveList = data.ApproveList
			a.ld.Tokens[token] = new(big.Int).Set(data.Amount)
		}
	}
	return nil
}

func (a *Account) CheckDestroyToken(recipient *Account) error {
	a.mu.RLock()
	defer a.mu.RUnlock()

	if err := a.closeLending(false, true); err != nil {
		return err
	}
	return a.destroyToken(recipient, false)
}

func (a *Account) DestroyToken(recipient *Account) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if err := a.closeLending(true, true); err != nil {
		return err
	}
	return a.destroyToken(recipient, true)
}

func (a *Account) destroyToken(recipient *Account, write bool) error {
	token := util.TokenSymbol(a.id)
	if !a.valid(ld.TokenAccount) {
		return fmt.Errorf(
			"Account(%s).CheckDestroyToken failed: invalid token account %s",
			a.id, token.GoString())
	}

	tk := a.ld.Tokens[token]
	if tk == nil {
		return fmt.Errorf("Account(%s).CheckDestroyToken failed: invalid token %s",
			a.id, token.GoString())
	} else if tk.Cmp(a.ld.MaxTotalSupply) != 0 {
		return fmt.Errorf("Account(%s).CheckDestroyToken failed: some token in the use, expected %v, got %v",
			a.id, a.ld.MaxTotalSupply, tk)
	}

	if write {
		recipient.Add(constants.NativeToken, a.ld.Balance)
		a.ld.Type = 0
		a.ld.Balance.SetUint64(0)
		a.ld.Threshold = 0
		a.ld.Keepers = a.ld.Keepers[:0]
		a.ld.NonceTable = make(map[uint64][]uint64)
		a.ld.Approver = nil
		a.ld.ApproveList = nil
		a.ld.MaxTotalSupply = nil
		delete(a.ld.Tokens, token)
	}
	return nil
}

func (a *Account) CheckCreateStake(
	from util.EthID,
	pledge *big.Int,
	acc *ld.TxAccounter,
	cfg *ld.StakeConfig,
) error {
	a.mu.RLock()
	defer a.mu.RUnlock()

	return a.createStake(from, pledge, acc, cfg, false)
}

func (a *Account) CreateStake(
	from util.EthID,
	pledge *big.Int,
	acc *ld.TxAccounter,
	cfg *ld.StakeConfig,
) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	return a.createStake(from, pledge, acc, cfg, true)
}

func (a *Account) createStake(
	from util.EthID,
	pledge *big.Int,
	acc *ld.TxAccounter,
	cfg *ld.StakeConfig,
	write bool,
) error {
	if !a.isEmpty() {
		return fmt.Errorf(
			"Account(%s).CheckCreateStake failed: stake account exists", a.id)
	}
	if stake := util.StakeSymbol(a.id); !stake.Valid() {
		return fmt.Errorf(
			"Account(%s).CheckCreateStake failed: invalid stake account", a.id)
	}
	if err := cfg.SyntacticVerify(); err != nil {
		return err
	}
	if write {
		a.ld.Type = ld.StakeAccount
		a.ld.Threshold = *acc.Threshold
		a.ld.Keepers = *acc.Keepers
		a.ld.Approver = acc.Approver
		a.ld.ApproveList = acc.ApproveList
		a.ld.Stake = cfg
		a.ld.StakeLedger = make(map[util.EthID]*ld.StakeEntry)
		a.ld.MaxTotalSupply = nil
		switch cfg.Token {
		case constants.NativeToken:
			a.ld.StakeLedger[from] = &ld.StakeEntry{Amount: new(big.Int).Set(pledge)}
		default:
			if b := a.ld.Tokens[cfg.Token]; b == nil {
				a.ld.Tokens[cfg.Token] = new(big.Int)
			}
		}
	}
	return nil
}

func (a *Account) CheckResetStake(cfg *ld.StakeConfig) error {
	a.mu.RLock()
	defer a.mu.RUnlock()

	return a.resetStake(cfg, false)
}

func (a *Account) ResetStake(cfg *ld.StakeConfig) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	return a.resetStake(cfg, true)
}

func (a *Account) resetStake(cfg *ld.StakeConfig, write bool) error {
	if !a.valid(ld.StakeAccount) {
		return fmt.Errorf(
			"Account(%s).CheckResetStake failed: invalid stake account", a.id)
	}
	if err := cfg.SyntacticVerify(); err != nil {
		return err
	}
	if cfg.Type != a.ld.Stake.Type {
		return fmt.Errorf(
			"Account(%s).CheckResetStake failed: can't change stake type", a.id)
	}
	if cfg.Token != a.ld.Stake.Token {
		return fmt.Errorf(
			"Account(%s).CheckResetStake failed: can't change stake token", a.id)
	}
	if a.ld.Stake.LockTime > a.ld.Timestamp {
		return fmt.Errorf(
			"Account(%s).CheckResetStake failed: stake in lock, please retry after lockTime", a.id)
	}
	holders := 0
	for _, v := range a.ld.StakeLedger {
		if v.Amount.Sign() > 0 {
			holders++
		}
	}
	if holders > 1 {
		return fmt.Errorf(
			"Account(%s).CheckResetStake failed: stake holders should not more than 1", a.id)
	}

	if write {
		a.ld.Stake.LockTime = cfg.LockTime
		a.ld.Stake.WithdrawFee = cfg.WithdrawFee
		if cfg.MinAmount.Sign() > 0 {
			a.ld.Stake.MinAmount.Set(cfg.MinAmount)
		}
		if cfg.MaxAmount.Sign() > 0 {
			a.ld.Stake.MaxAmount.Set(cfg.MaxAmount)
		}
	}
	return nil
}

func (a *Account) CheckDestroyStake(recipient *Account) error {
	a.mu.RLock()
	defer a.mu.RUnlock()

	if err := a.closeLending(false, true); err != nil {
		return err
	}
	return a.destroyStake(recipient, false)
}

func (a *Account) DestroyStake(recipient *Account) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if err := a.closeLending(true, true); err != nil {
		return err
	}
	return a.destroyStake(recipient, true)
}

func (a *Account) destroyStake(recipient *Account, write bool) error {
	if !a.valid(ld.StakeAccount) {
		return fmt.Errorf(
			"Account(%s).CheckDestroyStake failed: invalid stake account", a.id)
	}
	if a.ld.Stake.LockTime > a.ld.Timestamp {
		return fmt.Errorf(
			"Account(%s).CheckDestroyStake failed: stake in lock, please retry after lockTime", a.id)
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
				"Account(%s).CheckDestroyStake failed: recipient not exists", a.id)
		}
	default:
		return fmt.Errorf(
			"Account(%s).CheckDestroyStake failed: stake ledger not empty, please withdraw all except recipient", a.id)
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
		a.ld.NonceTable = make(map[uint64][]uint64)
		a.ld.Approver = nil
		a.ld.ApproveList = nil
		a.ld.Stake = nil
		a.ld.StakeLedger = nil
	}
	return nil
}

func (a *Account) CheckTakeStake(
	token util.TokenSymbol,
	from util.EthID,
	amount *big.Int,
	lockTime uint64) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	return a.takeStake(token, from, amount, lockTime, false)
}

func (a *Account) TakeStake(
	token util.TokenSymbol,
	from util.EthID,
	amount *big.Int,
	lockTime uint64) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	return a.takeStake(token, from, amount, lockTime, true)
}

func (a *Account) takeStake(
	token util.TokenSymbol,
	from util.EthID,
	amount *big.Int,
	lockTime uint64,
	write bool) error {
	if !a.valid(ld.StakeAccount) {
		return fmt.Errorf(
			"Account(%s).CheckTakeStake failed: invalid stake account", a.id)
	}

	stake := a.ld.Stake
	if token != stake.Token {
		return fmt.Errorf(
			"Account(%s).CheckTakeStake failed: invalid token, expected %s, got %s",
			a.id, stake.Token.GoString(), token.GoString())
	}

	if amount.Cmp(stake.MinAmount) < 0 {
		return fmt.Errorf(
			"Account(%s).CheckTakeStake failed: invalid amount, expected >= %v, got %v",
			a.id, stake.MinAmount, amount)
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
			"Account(%s).CheckTakeStake failed: invalid total amount, expected <= %v, got %v",
			a.id, stake.MaxAmount, total)
	}
	if lockTime > 0 && lockTime <= stake.LockTime {
		return fmt.Errorf(
			"Account(%s).CheckTakeStake failed: invalid lockTime, expected > %v, got %v",
			a.id, stake.LockTime, lockTime)
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
			"Account(%s).CheckUpdateStakeApprover failed: invalid stake account", a.id)
	}

	v := a.ld.StakeLedger[from]
	if v == nil {
		return fmt.Errorf(
			"Account(%s).CheckUpdateStakeApprover failed: %s has no stake ledger to update",
			a.id, util.EthID(from))
	}
	if v.Approver != nil && !signers.Has(*v.Approver) {
		return fmt.Errorf(
			"Account(%s).CheckUpdateStakeApprover failed: %s need approver signing",
			a.id, util.EthID(from))
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
			"Account(%s).CheckWithdrawStake failed: invalid stake account", a.id)
	}

	stake := a.ld.Stake
	if token != stake.Token {
		return nil, fmt.Errorf(
			"Account(%s).CheckWithdrawStake failed: invalid token, expected %s, got %s",
			a.id, stake.Token.GoString(), token.GoString())
	}
	if stake.LockTime > a.ld.Timestamp {
		return nil, fmt.Errorf(
			"Account(%s).CheckWithdrawStake failed: stake in lock, please retry after lockTime", a.id)
	}

	v := a.ld.StakeLedger[from]
	if v == nil {
		return nil, fmt.Errorf(
			"Account(%s).CheckWithdrawStake failed: %s has no stake to withdraw",
			a.id, from)
	}
	if v.LockTime > a.ld.Timestamp {
		return nil, fmt.Errorf(
			"Account(%s).CheckWithdrawStake failed: stake in lock, please retry after lockTime", a.id)
	}
	if v.Approver != nil && !signers.Has(*v.Approver) {
		return nil, fmt.Errorf(
			"Account(%s).CheckWithdrawStake failed: %s need approver signing",
			a.id, from)
	}
	total := new(big.Int).Set(v.Amount)
	rate := a.calcStakeBonusRate()
	bonus, _ := new(big.Float).Mul(new(big.Float).SetInt(v.Amount), rate).Int(nil)
	total = total.Add(total, bonus)
	if total.Cmp(amount) < 0 {
		return nil, fmt.Errorf(
			"Account(%s).CheckWithdrawStake failed: %s has an insufficient stake to withdraw, expected %v, got %v",
			a.id, from, amount, total)
	}

	if ba := a.balanceOf(token); ba.Cmp(amount) < 0 {
		return nil, fmt.Errorf(
			"Account(%s).CheckWithdrawStake failed: insufficient %s balance for withdraw, expected %v, got %v",
			a.id, token.GoString(), amount, ba)
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

func (a *Account) CheckOpenLending(cfg *ld.LendingConfig) error {
	a.mu.RLock()
	defer a.mu.RUnlock()

	return a.openLending(cfg, false)
}

func (a *Account) OpenLending(cfg *ld.LendingConfig) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	return a.openLending(cfg, true)
}

func (a *Account) openLending(cfg *ld.LendingConfig, write bool) error {
	if a.ld.Lending != nil || a.ld.LendingLedger != nil {
		return fmt.Errorf(
			"Account(%s).CheckOpenLending failed: lending exists", a.id)
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

	return a.closeLending(false, false)
}

func (a *Account) CloseLending() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	return a.closeLending(true, false)
}

func (a *Account) closeLending(write, ignoreNone bool) error {
	if ignoreNone && a.ld.Lending == nil {
		return nil
	}

	if a.ld.Lending == nil || a.ld.LendingLedger == nil {
		return fmt.Errorf(
			"Account(%s).CheckCloseLending failed: invalid lending", a.id)
	}

	if len(a.ld.LendingLedger) != 0 {
		return fmt.Errorf(
			"Account(%s).CheckCloseLending failed: please repay all before close", a.id)
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

	return a.borrow(token, from, amount, dueTime, false)
}

func (a *Account) Borrow(
	token util.TokenSymbol,
	from util.EthID,
	amount *big.Int,
	dueTime uint64,
) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	return a.borrow(token, from, amount, dueTime, true)
}

func (a *Account) borrow(
	token util.TokenSymbol,
	from util.EthID,
	amount *big.Int,
	dueTime uint64,
	write bool,
) error {
	if a.ld.Lending == nil || a.ld.LendingLedger == nil {
		return fmt.Errorf(
			"Account(%s).CheckBorrow failed: invalid lending", a.id)
	}

	if a.ld.Lending.Token != token {
		return fmt.Errorf(
			"Account(%s).CheckBorrow failed: invalid token, expected %s, got %s",
			a.id, a.ld.Lending.Token.GoString(), token.GoString())
	}
	if dueTime > 0 && dueTime <= a.ld.Timestamp {
		return fmt.Errorf(
			"Account(%s).CheckBorrow failed: invalid dueTime, expected > %d, got %d",
			a.id, a.ld.Timestamp, dueTime)
	}
	if amount.Cmp(a.ld.Lending.MinAmount) < 0 {
		return fmt.Errorf(
			"Account(%s).CheckBorrow failed: invalid amount, expected >= %v, got %v",
			a.id, a.ld.Lending.MinAmount, amount)
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
			"Account(%s).CheckBorrow failed: invalid amount, expected <= %v, got %v",
			a.id, a.ld.Lending.MaxAmount, total)
	}
	ba := a.balanceOf(token)
	if ba.Cmp(amount) < 0 {
		return fmt.Errorf(
			"Account(%s).CheckBorrow failed: insufficient %s balance, expected %v, got %v",
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

func (a *Account) CheckRepay(
	token util.TokenSymbol,
	from util.EthID,
	amount *big.Int,
) error {
	a.mu.RLock()
	defer a.mu.RUnlock()

	_, err := a.repay(token, from, amount, false)
	return err
}

func (a *Account) Repay(
	token util.TokenSymbol,
	from util.EthID,
	amount *big.Int,
) (*big.Int, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	return a.repay(token, from, amount, true)
}

func (a *Account) repay(
	token util.TokenSymbol,
	from util.EthID,
	amount *big.Int,
	write bool,
) (*big.Int, error) {
	if a.ld.Lending == nil || a.ld.LendingLedger == nil {
		return nil, fmt.Errorf(
			"Account(%s).CheckRepay failed: invalid lending", a.id)
	}

	if a.ld.Lending.Token != token {
		return nil, fmt.Errorf(
			"Account(%s).CheckRepay failed: invalid token, expected %s, got %s",
			a.id, a.ld.Lending.Token.GoString(), token.GoString())
	}

	e := a.ld.LendingLedger[from]
	if e == nil {
		return nil, fmt.Errorf("Account(%s).CheckRepay failed: don't need to repay", a.id)
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
