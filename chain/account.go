// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package chain

import (
	"fmt"
	"math/big"
	"sync"

	"github.com/ava-labs/avalanchego/database"
	"github.com/ava-labs/avalanchego/ids"

	"github.com/ldclabs/ldvm/ld"
)

type Account struct {
	ld  *ld.Account
	mu  sync.RWMutex
	id  ids.ShortID       // account address
	vdb database.Database // account version database
}

func NewAccount(id ids.ShortID) *Account {
	return &Account{
		id: id,
		ld: &ld.Account{
			ID:        id,
			Balance:   big.NewInt(0),
			Threshold: 1,
			Keepers:   []ids.ShortID{id},
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

func (a *Account) Init(vdb database.Database) {
	a.vdb = vdb
}

func (a *Account) Account() *ld.Account {
	a.mu.RLock()
	defer a.mu.RUnlock()

	return a.ld
}

func (a *Account) Nonce() uint64 {
	a.mu.RLock()
	defer a.mu.RUnlock()

	return a.ld.Nonce
}

func (a *Account) Balance() *big.Int {
	a.mu.RLock()
	defer a.mu.RUnlock()

	return a.ld.Balance
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

	return ld.SatisfySigning(a.ld.Threshold, a.ld.Keepers, signers, false)
}

func (a *Account) Add(amount *big.Int) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if amount == nil || amount.Sign() < 0 {
		return fmt.Errorf(
			"Account.Add %s invalid amount %v",
			ld.EthID(a.id), amount)
	}
	a.ld.Balance.Add(a.ld.Balance, amount)
	return nil
}

func (a *Account) Sub(amount *big.Int) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if amount == nil || amount.Sign() < 0 {
		return fmt.Errorf(
			"Account.Sub %s invalid amount %v",
			ld.EthID(a.id), amount)
	}
	if amount.Cmp(a.ld.Balance) > 0 {
		return fmt.Errorf(
			"Account.Sub %s insufficient balance %v",
			ld.EthID(a.id), amount)
	}
	a.ld.Balance.Sub(a.ld.Balance, amount)
	return nil
}

func (a *Account) SubByNonce(nonce uint64, amount *big.Int) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.ld.Nonce != nonce {
		return fmt.Errorf(
			"Account.SubByNonce %s invalid nonce %d",
			ld.EthID(a.id), nonce)
	}
	if amount == nil || amount.Sign() < 0 {
		return fmt.Errorf(
			"Account.SubByNonce %s invalid amount %v",
			ld.EthID(a.id), amount)
	}
	if amount.Cmp(a.ld.Balance) > 0 {
		return fmt.Errorf(
			"Account.SubByNonce %s insufficient balance to spent %v",
			ld.EthID(a.id), amount)
	}
	a.ld.Nonce++
	a.ld.Balance.Sub(a.ld.Balance, amount)
	return nil
}

func (a *Account) UpdateKeepers(nonce uint64, fee *big.Int, threshold uint8, keepers []ids.ShortID) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.ld.Nonce != nonce {
		return fmt.Errorf(
			"Account.UpdateKeepers %s invalid nonce %d",
			ld.EthID(a.id), nonce)
	}

	if a.ld.Balance.Cmp(fee) < 0 {
		return fmt.Errorf(
			"Account.UpdateKeepers %s insufficient balance to spent %v, required %v",
			ld.EthID(a.id), a.ld.Balance, fee)
	}

	a.ld.Nonce++
	a.ld.Balance.Sub(a.ld.Balance, fee)
	a.ld.Threshold = threshold
	a.ld.Keepers = keepers
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
