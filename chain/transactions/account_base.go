// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package transactions

import (
	"fmt"
	"math/big"
	"sync"

	"github.com/ldclabs/ldvm/constants"
	"github.com/ldclabs/ldvm/ld"
	"github.com/ldclabs/ldvm/util"
	"github.com/ldclabs/ldvm/util/signer"
)

type AccountCache map[util.Address]*Account

type Account struct {
	ld         *ld.Account
	ledger     *ld.AccountLedger
	mu         sync.RWMutex
	id         util.Address // account address
	pledge     *big.Int     // token account and stake account should have pledge
	ldHash     util.Hash
	ledgerHash util.Hash
}

func NewAccount(id util.Address) *Account {
	ld := &ld.Account{
		ID:         id,
		Balance:    big.NewInt(0),
		Keepers:    signer.Keys{},
		Tokens:     make(map[string]*big.Int),
		NonceTable: make(map[uint64][]uint64),
	}

	if err := ld.SyntacticVerify(); err != nil {
		panic(err)
	}
	return &Account{
		id:     id,
		pledge: new(big.Int),
		ld:     ld,
		ldHash: util.HashFromData(ld.Bytes()),
	}
}

func ParseAccount(id util.Address, data []byte) (*Account, error) {
	errp := util.ErrPrefix(fmt.Sprintf("ParseAccount(%s): ", id))

	a := NewAccount(id)
	if err := a.ld.Unmarshal(data); err != nil {
		return nil, errp.ErrorIf(err)
	}
	if err := a.ld.SyntacticVerify(); err != nil {
		return nil, errp.ErrorIf(err)
	}
	a.ld.ID = id
	a.ldHash = util.HashFromData(a.ld.Bytes())
	return a, nil
}

func (a *Account) Init(pledge *big.Int, height, timestamp uint64) *Account {
	a.pledge.Set(pledge)
	a.ld.Height = height
	a.ld.Timestamp = timestamp
	return a
}

func (a *Account) Ledger() *ld.AccountLedger {
	a.mu.RLock()
	defer a.mu.RUnlock()

	return a.ledger
}

func (a *Account) InitLedger(data []byte) error {
	errp := util.ErrPrefix(fmt.Sprintf("Account(%s).InitLedger: ", a.id))

	a.mu.Lock()
	defer a.mu.Unlock()

	a.ledger = &ld.AccountLedger{}
	if err := a.ledger.Unmarshal(data); err != nil {
		errp.ErrorIf(err)
	}
	if err := a.ledger.SyntacticVerify(); err != nil {
		errp.ErrorIf(err)
	}
	a.ledgerHash = util.HashFromData(a.ledger.Bytes())
	return nil
}

func (a *Account) ID() util.Address {
	return a.id
}

func (a *Account) IDKey() signer.Key {
	return signer.Key(a.id.Bytes())
}

func (a *Account) LD() *ld.Account {
	return a.ld
}

func (a *Account) Type() ld.AccountType {
	return a.ld.Type
}

func (a *Account) IsEmpty() bool {
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

	case (a.IsEmpty() && a.id != util.AddressEmpty) || a.ld.Balance.Cmp(a.pledge) < 0:
		return false

	case t == ld.TokenAccount && (a.ld.MaxTotalSupply == nil || a.ld.MaxTotalSupply.Sign() <= 0):
		return false

	case t == ld.StakeAccount && a.ld.Stake == nil:
		return false

	default:
		return true
	}
}

func (a *Account) Verify(digestHash []byte, sigs signer.Sigs, accountKey signer.Key) bool {
	return a.ld.Verify(digestHash, sigs, accountKey)
}

func (a *Account) VerifyPlus(digestHash []byte, sigs signer.Sigs, accountKey signer.Key) bool {
	return a.ld.VerifyPlus(digestHash, sigs, accountKey)
}

func (a *Account) Nonce() uint64 {
	a.mu.RLock()
	defer a.mu.RUnlock()

	return a.ld.Nonce
}

func (a *Account) Balance() *big.Int {
	a.mu.RLock()
	defer a.mu.RUnlock()

	return a.balanceOf(constants.NativeToken)
}

func (a *Account) BalanceOf(token util.TokenSymbol) *big.Int {
	a.mu.RLock()
	defer a.mu.RUnlock()

	return a.balanceOf(token)
}

func (a *Account) TotalSupply() *big.Int {
	a.mu.RLock()
	defer a.mu.RUnlock()

	total := new(big.Int)
	if a.ld.Type != ld.TokenAccount {
		return total
	}
	total.Set(a.ld.MaxTotalSupply)
	return total.Sub(total, a.balanceOf(util.TokenSymbol(a.id)))
}

func (a *Account) balanceOf(token util.TokenSymbol) *big.Int {
	switch token {
	case constants.NativeToken:
		if b := new(big.Int).Sub(a.ld.Balance, a.pledge); b.Sign() >= 0 {
			return b
		}
		return new(big.Int)

	default:
		if v := a.ld.Tokens[token.AsKey()]; v != nil {
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
		if v := a.ld.Tokens[token.AsKey()]; v != nil {
			return new(big.Int).Set(v)
		}
		return new(big.Int)
	}
}

func (a *Account) CheckBalance(token util.TokenSymbol, amount *big.Int) error {
	a.mu.RLock()
	defer a.mu.RUnlock()

	errp := util.ErrPrefix(fmt.Sprintf("Account(%s).CheckBalance: ", a.id))
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
	errp := util.ErrPrefix(fmt.Sprintf("Account(%s).CheckAsFrom: ", a.id))

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
	errp := util.ErrPrefix(fmt.Sprintf("Account(%s).CheckAsTo: ", a.id))

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

func (a *Account) Keepers() signer.Keys {
	a.mu.RLock()
	defer a.mu.RUnlock()

	return a.ld.Keepers
}

func (a *Account) Add(token util.TokenSymbol, amount *big.Int) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	errp := util.ErrPrefix(fmt.Sprintf("Account(%s).Add: ", a.id))

	switch {
	case amount == nil || amount.Sign() < 0:
		return errp.Errorf("invalid amount %v", amount)

	case amount.Sign() > 0:
		switch token {
		case constants.NativeToken:
			a.ld.Balance.Add(a.ld.Balance, amount)

		default:
			v := a.ld.Tokens[token.AsKey()]
			if v == nil {
				v = new(big.Int)
				a.ld.Tokens[token.AsKey()] = v
			}
			v.Add(v, amount)
		}
	}
	return nil
}

func (a *Account) Sub(token util.TokenSymbol, amount *big.Int) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	errp := util.ErrPrefix(fmt.Sprintf("Account(%s).Sub: ", a.id))
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
			v := a.ld.Tokens[token.AsKey()]
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

	errp := util.ErrPrefix(fmt.Sprintf("Account(%s).SubByNonce: ", a.id))
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

	errp := util.ErrPrefix(fmt.Sprintf("Account(%s).SubByNonceTable: ", a.id))
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

func (a *Account) UpdateNonceTable(expire uint64, ns []uint64) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	errp := util.ErrPrefix(fmt.Sprintf("Account(%s).UpdateNonceTable: ", a.id))
	if expire <= a.ld.Timestamp {
		return errp.Errorf("invalid expire, expected >= %d, got %d", a.ld.Timestamp, expire)
	}

	us := util.Uint64Set(make(map[uint64]struct{}, len(ns)))
	for _, u := range ns {
		if us.Has(u) {
			return errp.Errorf("nonce %d exists at %d", u, expire)
		}
		us.Add(u)
	}

	// clear expired nonces
	for e := range a.ld.NonceTable {
		if e < a.ld.Timestamp {
			delete(a.ld.NonceTable, e)
		}
	}

	a.ld.NonceTable[expire] = us.List()
	if len(a.ld.NonceTable) > 1024 {
		return errp.Errorf("too many NonceTable groups, expected <= 1024")
	}
	return nil
}

func (a *Account) UpdateKeepers(
	threshold *uint16,
	keepers *signer.Keys,
	approver *signer.Key,
	approveList *ld.TxTypes,
) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if approver != nil {
		if len(*approver) == 0 {
			a.ld.Approver = nil
			a.ld.ApproveList = nil
		} else {
			a.ld.Approver = *approver
		}
	}
	if approveList != nil {
		a.ld.ApproveList = *approveList
	}
	if threshold != nil && keepers != nil {
		a.ld.Threshold = *threshold
		a.ld.Keepers = *keepers
	}
	return nil
}

func (a *Account) Marshal() ([]byte, []byte, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	if err := a.ld.SyntacticVerify(); err != nil {
		return nil, nil, util.ErrPrefix(fmt.Sprintf("Account(%s).Marshal: ", a.id)).ErrorIf(err)
	}

	var ledger []byte
	if a.ledger != nil {
		if err := a.ledger.SyntacticVerify(); err != nil {
			return nil, nil, util.ErrPrefix(fmt.Sprintf("Account(%s).Marshal: ", a.id)).ErrorIf(err)
		}
		ledger = a.ledger.Bytes()
	}
	return a.ld.Bytes(), ledger, nil
}

func (a *Account) AccountChanged(data []byte) bool {
	return a.ldHash != util.HashFromData(data)
}

func (a *Account) LedgerChanged(data []byte) bool {
	return a.ledgerHash != util.HashFromData(data)
}
