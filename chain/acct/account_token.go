// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package acct

import (
	"fmt"
	"math/big"

	"github.com/ldclabs/ldvm/ids"
	"github.com/ldclabs/ldvm/ld"
	"github.com/ldclabs/ldvm/util/erring"
)

func (a *Account) TotalSupply() *big.Int {
	a.mu.RLock()
	defer a.mu.RUnlock()

	total := new(big.Int)
	if a.ld.Type != ld.TokenAccount {
		return total
	}

	total.Set(a.ld.MaxTotalSupply)
	return total.Sub(total, a.balanceOf(ids.TokenSymbol(a.ld.ID), false))
}

func (a *Account) CreateToken(data *ld.TxAccounter) error {
	errp := erring.ErrPrefix(fmt.Sprintf("acct.Account(%s).CreateToken: ", a.ld.ID.String()))

	a.mu.Lock()
	defer a.mu.Unlock()

	token := ids.TokenSymbol(a.ld.ID)
	if !token.Valid() {
		return errp.Errorf("invalid token %s", token.GoString())
	}

	if !a.IsEmpty() {
		return errp.Errorf("token account %s exists", token)
	}

	a.ld.Type = ld.TokenAccount
	a.ld.MaxTotalSupply = new(big.Int).Set(data.Amount)
	switch token {
	case ids.NativeToken: // NativeToken created by genesis tx
		a.ld.Balance.Set(data.Amount)
	default:
		a.ld.Threshold = *data.Threshold
		a.ld.Keepers = *data.Keepers

		if data.Approver != nil && data.Approver.Valid() == nil {
			a.ld.Approver = *data.Approver
		}
		if data.ApproveList != nil {
			a.ld.ApproveList = *data.ApproveList
		}
		a.ld.Tokens[token.AsKey()] = new(big.Int).Set(data.Amount)
	}
	return nil
}

func (a *Account) DestroyToken(recipient *Account) error {
	errp := erring.ErrPrefix(fmt.Sprintf("acct.Account(%s).DestroyToken: ", a.ld.ID.String()))

	a.mu.Lock()
	defer a.mu.Unlock()

	token := ids.TokenSymbol(a.ld.ID)
	if !a.valid(ld.TokenAccount) {
		return errp.Errorf("invalid token account %s", token.GoString())
	}

	tk := a.ld.Tokens[token.AsKey()]
	if tk == nil {
		return errp.Errorf("invalid token %s", token.GoString())
	} else if tk.Cmp(a.ld.MaxTotalSupply) != 0 {
		return errp.Errorf("some token in the use, maxTotalSupply expected %v, got %v",
			a.ld.MaxTotalSupply, tk)
	}

	if err := a.closeLending(true); err != nil {
		return errp.ErrorIf(err)
	}

	recipient.Add(ids.NativeToken, a.ld.Balance)
	a.ld.Type = 0
	a.ld.Balance.SetUint64(0)
	a.ld.Threshold = 0
	a.ld.Keepers = a.ld.Keepers[:0]
	a.ld.NonceTable = make(map[uint64][]uint64)
	a.ld.Approver = nil
	a.ld.ApproveList = nil
	a.ld.MaxTotalSupply = nil
	delete(a.ld.Tokens, token.AsKey())
	return nil
}
