// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package transaction

import (
	"fmt"
	"math/big"

	"github.com/ldclabs/ldvm/constants"
	"github.com/ldclabs/ldvm/ld"
	"github.com/ldclabs/ldvm/util"
)

func (a *Account) CheckCreateToken(data *ld.TxAccounter) error {
	a.mu.RLock()
	defer a.mu.RUnlock()

	errp := util.ErrPrefix(fmt.Sprintf("Account(%s).CheckCreateToken error: ", a.id))
	return errp.ErrorIf(a.createToken(data, false))
}

func (a *Account) CreateToken(data *ld.TxAccounter) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	errp := util.ErrPrefix(fmt.Sprintf("Account(%s).CreateToken error: ", a.id))
	return errp.ErrorIf(a.createToken(data, true))
}

func (a *Account) createToken(data *ld.TxAccounter, write bool) error {
	token := util.TokenSymbol(a.id)
	if !token.Valid() {
		return fmt.Errorf("invalid token %s", token.GoString())
	}

	if !a.isEmpty() {
		return fmt.Errorf("token account %s exists", token)
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

	errp := util.ErrPrefix(fmt.Sprintf("Account(%s).CheckDestroyToken error: ", a.id))
	if err := a.closeLending(false, true); err != nil {
		return errp.ErrorIf(err)
	}
	return errp.ErrorIf(a.destroyToken(recipient, false))
}

func (a *Account) DestroyToken(recipient *Account) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	errp := util.ErrPrefix(fmt.Sprintf("Account(%s).DestroyToken error: ", a.id))
	if err := a.closeLending(true, true); err != nil {
		return errp.ErrorIf(err)
	}
	return errp.ErrorIf(a.destroyToken(recipient, true))
}

func (a *Account) destroyToken(recipient *Account, write bool) error {
	token := util.TokenSymbol(a.id)
	if !a.valid(ld.TokenAccount) {
		return fmt.Errorf("invalid token account %s", token.GoString())
	}

	tk := a.ld.Tokens[token]
	if tk == nil {
		return fmt.Errorf("invalid token %s", token.GoString())
	} else if tk.Cmp(a.ld.MaxTotalSupply) != 0 {
		return fmt.Errorf("some token in the use, expected %v, got %v",
			a.ld.MaxTotalSupply, tk)
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
