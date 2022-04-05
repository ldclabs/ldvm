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

var (
	BigInt0 = big.NewInt(0)
)

type Account struct {
	ld        *ld.Account
	mu        sync.RWMutex
	id        ids.ShortID // account address
	guardians ids.ShortSet
	vdb       database.Database // account version database
}

func NewAccount() *Account {
	return &Account{
		ld: &ld.Account{
			Balance:   big.NewInt(0),
			Threshold: 1,
			Guardians: make([]ids.ShortID, 0),
		},
	}
}

func ParseAccount(data []byte) (*Account, error) {
	a := &Account{ld: &ld.Account{}}
	if err := a.ld.Unmarshal(data); err != nil {
		return nil, err
	}
	if err := a.ld.SyntacticVerify(); err != nil {
		return nil, err
	}
	return a, nil
}

func (a *Account) Init(id ids.ShortID, vdb database.Database) {
	a.id = id
	a.vdb = vdb
	a.guardians = ids.NewShortSet(len(a.ld.Guardians))
	a.guardians.Add(a.ld.Guardians...)
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

func (a *Account) Add(nonce uint64, amount *big.Int) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.ld.Nonce != nonce {
		return fmt.Errorf("invalid nonce %d of account %s", nonce, a.id)
	}
	if amount == nil || amount.Cmp(BigInt0) < 1 {
		return fmt.Errorf("invalid amount %v of account %s", amount, a.id)
	}
	amount.Add(a.ld.Balance, amount)
	// if amount.Cmp(MaxSuply) > 0 {
	// 	return fmt.Errorf("invalid amount %v of account %s", amount, a.id)
	// }
	a.ld.Nonce++
	a.ld.Balance = amount
	return nil
}

func (a *Account) Sub(nonce uint64, amount *big.Int) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.ld.Nonce != nonce {
		return fmt.Errorf("invalid nonce %d of account %s", nonce, a.id)
	}
	if amount == nil || amount.Cmp(BigInt0) < 1 {
		return fmt.Errorf("invalid amount %v of account %s", amount, a.id)
	}
	if amount.Cmp(a.ld.Balance) > 0 {
		return fmt.Errorf("invalid amount %v of account %s", amount, a.id)
	}
	a.ld.Nonce++
	a.ld.Balance.Sub(a.ld.Balance, amount)
	return nil
}

func (a *Account) SatisfySigning(signers []ids.ShortID) bool {
	a.mu.RLock()
	defer a.mu.RUnlock()

	if len(signers) < int(a.ld.Threshold) {
		return false
	}

	// the first signer should be the sender
	if a.id != signers[0] {
		return false
	}

	t := uint(1)
	for _, id := range signers[1:] {
		if a.guardians.Contains(id) {
			t++
		}
	}
	return t >= uint(a.ld.Threshold)
}

func (a *Account) UpdateGuardians(nonce uint64, fee *big.Int, threshold uint8, guardians []ids.ShortID) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.ld.Nonce != nonce {
		return fmt.Errorf("invalid nonce %d of account %s", nonce, a.id)
	}

	if a.ld.Balance.Cmp(fee) < 0 {
		return fmt.Errorf("insufficient balance %d of account %s, required %d",
			a.ld.Balance, a.id, fee)
	}

	a.ld.Nonce++
	a.ld.Balance.Sub(a.ld.Balance, fee)
	a.ld.Threshold = threshold
	a.ld.Guardians = guardians

	a.guardians = ids.NewShortSet(len(a.ld.Guardians))
	a.guardians.Add(a.ld.Guardians...)
	return nil
}

// Commit will be called when stateBlock.SaveBlock
func (a *Account) Commit() error {
	if err := a.ld.SyntacticVerify(); err != nil {
		return err
	}
	return a.vdb.Put(a.id[:], a.ld.Bytes())
}
