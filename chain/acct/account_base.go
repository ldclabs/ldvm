// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package acct

import (
	"fmt"
	"math/big"
	"sync"

	"github.com/ldclabs/ldvm/ids"
	"github.com/ldclabs/ldvm/ld"
	"github.com/ldclabs/ldvm/signer"
	"github.com/ldclabs/ldvm/util/erring"
)

type ActiveAccounts map[ids.Address]*Account

type Account struct {
	mu         sync.RWMutex
	ld         *ld.Account
	ledger     *ld.AccountLedger
	ntb        *big.Int // non-transferable balance
	pledge     *big.Int // token account and stake account should have pledge
	ldHash     *ids.ID32
	ledgerHash *ids.ID32
}

func NewAccount(id ids.Address) *Account {
	ld := &ld.Account{
		ID:         id,
		Balance:    big.NewInt(0),
		Keepers:    make(signer.Keys, 0),
		Tokens:     make(map[string]*big.Int),
		NonceTable: make(map[uint64][]uint64),
	}

	if err := ld.SyntacticVerify(); err != nil {
		panic(err)
	}

	return &Account{
		ld:     ld,
		ntb:    big.NewInt(0),
		pledge: big.NewInt(0),
		ldHash: ids.ID32FromData(ld.Bytes()).Ptr(),
	}
}

func ParseAccount(id ids.Address, data []byte) (a *Account, err error) {
	errp := erring.ErrPrefix(fmt.Sprintf("acct.ParseAccount(%s): ", id.String()))

	a = NewAccount(id)
	if err = a.ld.Unmarshal(data); err != nil {
		return nil, errp.ErrorIf(err)
	}
	if err = a.ld.SyntacticVerify(); err != nil {
		return nil, errp.ErrorIf(err)
	}

	a.ld.ID = id
	a.ldHash = ids.ID32FromData(a.ld.Bytes()).Ptr()
	return a, nil
}

func (a *Account) Init(nonTransferableBalance, pledge *big.Int, height, timestamp uint64) *Account {
	a.ntb.Set(nonTransferableBalance)
	a.pledge.Set(pledge)
	a.ld.Height = height
	a.ld.Timestamp = timestamp
	return a
}

func (a *Account) LoadLedger(force bool, fn func() ([]byte, error)) error {
	errp := erring.ErrPrefix(fmt.Sprintf("acct.Account(%s).LoadLedger: ", a.ld.ID.String()))

	a.mu.Lock()
	defer a.mu.Unlock()

	if !force && a.ledger != nil {
		return nil
	}

	data, err := fn()
	if err != nil {
		return errp.ErrorIf(err)
	}

	l := &ld.AccountLedger{}
	if err = l.Unmarshal(data); err != nil {
		return errp.ErrorIf(err)
	}
	if err = l.SyntacticVerify(); err != nil {
		return errp.ErrorIf(err)
	}

	a.ledger = l
	a.ledgerHash = ids.ID32FromData(a.ledger.Bytes()).Ptr()
	return nil
}

func (a *Account) Type() ld.AccountType {
	return a.ld.Type
}

func (a *Account) ID() ids.Address {
	return a.ld.ID
}

func (a *Account) IDKey() signer.Key {
	return signer.Key(a.ld.ID.Bytes())
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

func (a *Account) LD() *ld.Account {
	return a.ld
}

func (a *Account) Ledger() *ld.AccountLedger {
	a.mu.RLock()
	defer a.mu.RUnlock()

	return a.ledger
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

	case a.ld.ID == ids.LDCAccount:
		return true

	case t == ld.NativeAccount && a.ld.Balance.Sign() >= 0:
		return true

	case a.IsEmpty() || a.ld.Balance.Cmp(a.pledge) < 0:
		return false

	case t == ld.TokenAccount && (a.ld.MaxTotalSupply == nil || a.ld.MaxTotalSupply.Sign() <= 0):
		return false

	case t == ld.StakeAccount && a.ld.Stake == nil:
		return false

	default:
		return true
	}
}

func (a *Account) ValidValidator() bool {
	a.mu.RLock()
	defer a.mu.RUnlock()

	return a.valid(ld.StakeAccount) && a.ld.Stake.Type == 0
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

	return a.balanceOf(ids.NativeToken, true)
}

func (a *Account) BalanceOf(token ids.TokenSymbol) *big.Int {
	a.mu.RLock()
	defer a.mu.RUnlock()

	return a.balanceOf(token, true)
}

func (a *Account) balanceOf(token ids.TokenSymbol, checkNTB bool) *big.Int {
	b := new(big.Int)

	switch token {
	case ids.NativeToken:
		b.Sub(a.ld.Balance, a.pledge)
		if checkNTB {
			b.Sub(b, a.ntb)
		}
		if b.Sign() >= 0 {
			return b
		}
		return new(big.Int)

	default:
		if v := a.ld.Tokens[token.AsKey()]; v != nil {
			b.Set(v)
		}
		return b
	}
}

func (a *Account) BalanceOfAll(token ids.TokenSymbol) *big.Int {
	a.mu.RLock()
	defer a.mu.RUnlock()

	return a.balanceOfAll(token)
}

func (a *Account) balanceOfAll(token ids.TokenSymbol) *big.Int {
	b := new(big.Int)

	switch token {
	case ids.NativeToken:
		return b.Set(a.ld.Balance)

	default:
		if v := a.ld.Tokens[token.AsKey()]; v != nil {
			b.Set(v)
		}
		return b
	}
}

func (a *Account) CheckBalance(token ids.TokenSymbol, amount *big.Int, checkNTB bool) error {
	a.mu.RLock()
	defer a.mu.RUnlock()

	errp := erring.ErrPrefix(fmt.Sprintf("acct.Account(%s).CheckBalance: ", a.ld.ID.String()))
	return errp.ErrorIf(a.checkBalance(token, amount, false))
}

func (a *Account) checkBalance(token ids.TokenSymbol, amount *big.Int, checkNTB bool) error {
	ba := a.balanceOf(token, checkNTB)

	switch {
	case amount == nil || amount.Sign() < 0:
		return fmt.Errorf("invalid amount %v", amount)

	case amount.Cmp(ba) > 0:
		switch checkNTB {
		case true:
			return fmt.Errorf("insufficient transferable %s balance, expected %v, got %v",
				token.GoString(), amount, ba)
		default:
			return fmt.Errorf("insufficient %s balance, expected %v, got %v",
				token.GoString(), amount, ba)
		}

	default:
		return nil
	}
}

func (a *Account) Add(token ids.TokenSymbol, amount *big.Int) error {
	errp := erring.ErrPrefix(fmt.Sprintf("acct.Account(%s).Add: ", a.ld.ID.String()))

	a.mu.Lock()
	defer a.mu.Unlock()

	switch {
	case amount == nil || amount.Sign() < 0:
		return errp.Errorf("invalid amount %v", amount)

	case amount.Sign() > 0:
		switch token {
		case ids.NativeToken:
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

func (a *Account) Sub(token ids.TokenSymbol, amount *big.Int) error {
	errp := erring.ErrPrefix(fmt.Sprintf("acct.Account(%s).Sub: ", a.ld.ID.String()))

	a.mu.Lock()
	defer a.mu.Unlock()

	if err := a.checkBalance(token, amount, true); err != nil {
		return errp.ErrorIf(err)
	}

	a.subNoCheck(token, amount)
	return nil
}

func (a *Account) subNoCheck(token ids.TokenSymbol, amount *big.Int) {
	if amount.Sign() > 0 {
		switch token {
		case ids.NativeToken:
			a.ld.Balance.Sub(a.ld.Balance, amount)
		default:
			v := a.ld.Tokens[token.AsKey()]
			v.Sub(v, amount)
		}
	}
}

func (a *Account) SubGasByNonce(token ids.TokenSymbol, nonce uint64, amount *big.Int) error {
	errp := erring.ErrPrefix(fmt.Sprintf("acct.Account(%s).SubGasByNonce: ", a.ld.ID.String()))

	a.mu.Lock()
	defer a.mu.Unlock()

	if a.ld.Nonce != nonce {
		return errp.Errorf("invalid nonce, expected %d, got %d", a.ld.Nonce, nonce)
	}

	if err := a.checkBalance(token, amount, false); err != nil {
		return errp.ErrorIf(err)
	}

	a.ld.Nonce++
	a.subNoCheck(token, amount)
	return nil
}

func (a *Account) SubByNonceTable(token ids.TokenSymbol, expire, nonce uint64, amount *big.Int) error {
	errp := erring.ErrPrefix(fmt.Sprintf("acct.Account(%s).SubByNonceTable: ", a.ld.ID.String()))

	a.mu.Lock()
	defer a.mu.Unlock()

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

	if err := a.checkBalance(token, amount, true); err != nil {
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
	errp := erring.ErrPrefix(fmt.Sprintf("acct.Account(%s).UpdateNonceTable: ", a.ld.ID.String()))

	a.mu.Lock()
	defer a.mu.Unlock()

	if expire <= a.ld.Timestamp {
		return errp.Errorf("invalid expire, expected >= %d, got %d", a.ld.Timestamp, expire)
	}

	us := ids.NewSet[uint64](len(ns))
	if err := us.CheckAdd(ns...); err != nil {
		return errp.ErrorIf(err)
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
		if ap := *approver; len(ap) == 0 {
			a.ld.Approver = nil
			a.ld.ApproveList = nil
		} else {
			a.ld.Approver = ap
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
	errp := erring.ErrPrefix(fmt.Sprintf("acct.Account(%s).Marshal: ", a.ld.ID.String()))

	if err := a.ld.SyntacticVerify(); err != nil {
		return nil, nil, errp.ErrorIf(err)
	}

	var ledger []byte
	if a.ledger != nil {
		if err := a.ledger.SyntacticVerify(); err != nil {
			return nil, nil, errp.ErrorIf(err)
		}
		ledger = a.ledger.Bytes()
	}
	return a.ld.Bytes(), ledger, nil
}

func (a *Account) ResetPledge() {
	a.pledge.SetUint64(0)
}

func (a *Account) AccountChanged(data []byte) bool {
	return a.ldHash == nil || *a.ldHash != ids.ID32FromData(data)
}

func (a *Account) LedgerChanged(data []byte) bool {
	return a.ledgerHash == nil || *a.ledgerHash != ids.ID32FromData(data)
}
