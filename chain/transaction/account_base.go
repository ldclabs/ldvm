// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package transaction

import (
	"fmt"
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
	errp := util.ErrPrefix(fmt.Sprintf("ParseAccount(%s) error: ", id))

	a := &Account{id: id, pledge: new(big.Int), ld: &ld.Account{Balance: new(big.Int)}}
	if err := a.ld.Unmarshal(data); err != nil {
		return nil, errp.ErrorIf(err)
	}
	if err := a.ld.SyntacticVerify(); err != nil {
		return nil, errp.ErrorIf(err)
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

	errp := util.ErrPrefix(fmt.Sprintf("Account(%s).CheckBalance error: ", a.id))
	return errp.ErrorIf(a.checkBalance(token, amount))
}

func (a *Account) checkBalance(token util.TokenSymbol, amount *big.Int) error {
	switch {
	case amount == nil || amount.Sign() < 0:
		return fmt.Errorf("invalid amount %v", amount)

	case amount.Sign() > 0:
		if ba := a.balanceOf(token); amount.Cmp(ba) > 0 {
			return fmt.Errorf("insufficient %s balance, expected %v, got %v", token.GoString(), amount, ba)
		}
	}
	return nil
}

func (a *Account) CheckAsFrom(txType ld.TxType) error {
	errp := util.ErrPrefix(fmt.Sprintf("Account(%s).CheckAsFrom error: ", a.id))

	switch a.ld.Type {
	case ld.TokenAccount:
		switch {
		case ld.TokenFromTxTypes.Has(txType):
			// just go ahead
		default:
			return errp.Errorf("can't use TokenAccount as sender for %s", txType.String())
		}

	case ld.StakeAccount:
		if a.ld.Stake == nil {
			return errp.Errorf("invalid StakeAccount as sender for %s", txType.String())
		}
		ty := a.ld.Stake.Type
		if ty > 2 {
			return errp.Errorf("can't use unknown type %d StakeAccount as sender for %s",
				ty, txType.String())
		}

		// 0: account keepers can not use stake token
		// 1: account keepers can take a stake in other stake account
		// 2: in addition to 1, account keepers can transfer stake token to other account
		switch {
		case ld.StakeFromTxTypes0.Has(txType):
			// just go ahead
		case ld.StakeFromTxTypes1.Has(txType):
			if ty < 1 {
				return errp.Errorf("can't use type %d StakeAccount as sender for %s",
					ty, txType.String())
			}

		case ld.StakeFromTxTypes2.Has(txType):
			if ty < 2 {
				return errp.Errorf("can't use type %d StakeAccount as sender for %s",
					ty, txType.String())
			}

		default:
			return errp.Errorf("can't use type %d StakeAccount as sender for %s",
				ty, txType.String())
		}
	}
	return nil
}

func (a *Account) CheckAsTo(txType ld.TxType) error {
	errp := util.ErrPrefix(fmt.Sprintf("Account(%s).CheckAsTo error: ", a.id))

	switch a.ld.Type {
	case ld.TokenAccount:
		switch {
		case ld.TokenToTxTypes.Has(txType):
			// just go ahead
		default:
			return errp.Errorf("can't use TokenAccount as recipient for %s", txType.String())
		}

	case ld.StakeAccount:
		switch {
		case ld.StakeToTxTypes.Has(txType):
			// just go ahead
		default:
			return errp.Errorf("can't use StakeAccount as recipient for %s", txType.String())
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

	errp := util.ErrPrefix(fmt.Sprintf("Account(%s).Add error: ", a.id))

	switch {
	case amount == nil || amount.Sign() < 0:
		return errp.Errorf("invalid amount %v", amount)

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

	errp := util.ErrPrefix(fmt.Sprintf("Account(%s).Sub error: ", a.id))
	if err := a.checkBalance(token, amount); err != nil {
		return errp.ErrorIf(err)
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

	errp := util.ErrPrefix(fmt.Sprintf("Account(%s).SubByNonce error: ", a.id))
	if a.ld.Nonce != nonce {
		return errp.Errorf("invalid nonce, expected %d, got %d", a.ld.Nonce, nonce)
	}

	if err := a.checkBalance(token, amount); err != nil {
		return errp.ErrorIf(err)
	}

	a.ld.Nonce++
	a.subNoCheck(token, amount)
	return nil
}

func (a *Account) SubByNonceTable(
	token util.TokenSymbol,
	expire, nonce uint64,
	amount *big.Int) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	errp := util.ErrPrefix(fmt.Sprintf("Account(%s).SubByNonceTable error: ", a.id))
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
		return errp.Errorf("nonce %d not exists at %d", nonce, expire)
	}

	if err := a.checkBalance(token, amount); err != nil {
		return errp.ErrorIf(err)
	}

	copy(uu[i:], uu[i+1:])
	uu = uu[:len(uu)-1]
	if len(uu) == 0 {
		delete(a.ld.NonceTable, expire)
	} else {
		a.ld.NonceTable[expire] = uu
	}
	a.subNoCheck(token, amount)
	return nil
}

func (a *Account) AddNonceTable(expire uint64, ns []uint64) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	errp := util.ErrPrefix(fmt.Sprintf("Account(%s).AddNonceTable error: ", a.id))
	if len(a.ld.NonceTable) >= 1024 {
		return errp.Errorf("too many NonceTable groups, expected <= 1024")
	}

	us := util.Uint64Set(make(map[uint64]struct{}, len(a.ld.NonceTable[expire])+len(ns)))
	if uu, ok := a.ld.NonceTable[expire]; ok {
		us.Add(uu...)
	}
	for _, u := range ns {
		if us.Has(u) {
			return errp.Errorf("nonce %d exists at %d", u, expire)
		}
		us.Add(u)
	}

	a.ld.NonceTable[expire] = us.List()
	// clear expired nonces
	for e := range a.ld.NonceTable {
		if e < a.ld.Timestamp {
			delete(a.ld.NonceTable, e)
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

func (a *Account) Marshal() ([]byte, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	if err := a.ld.SyntacticVerify(); err != nil {
		return nil, util.ErrPrefix(fmt.Sprintf("Account(%s).Marshal error: ", a.id)).ErrorIf(err)
	}
	return a.ld.Bytes(), nil
}
