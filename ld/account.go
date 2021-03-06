// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package ld

import (
	"fmt"
	"math/big"

	"github.com/ldclabs/ldvm/util"
)

const (
	NativeAccount AccountType = iota
	TokenAccount              // The first byte of account address must be $
	StakeAccount              // The first byte of account address must be #
)

// AccountType is an uint16 representing the type of account
type AccountType uint16

func (t AccountType) String() string {
	switch t {
	case NativeAccount:
		return "Native"
	case TokenAccount:
		return "Token"
	case StakeAccount:
		return "Stake"
	default:
		return fmt.Sprintf("UnknownAccountType(%d)", t)
	}
}

func (t AccountType) MarshalJSON() ([]byte, error) {
	return []byte("\"" + t.String() + "\""), nil
}

const MaxKeepers = 1024

type Account struct {
	Type AccountType `cbor:"t" json:"type"`
	// Nonce should increase 1 when sender issuing tx, but not increase when receiving
	Nonce uint64 `cbor:"n" json:"nonce"`
	// the decimals is 9, the smallest unit "NanoLDC" equal to gwei.
	Balance *big.Int `cbor:"b" json:"balance"`
	// M of N threshold signatures, aka MultiSig: threshold is m, keepers length is n.
	// The minimum value is 1, the maximum value is len(keepers)
	Threshold uint16 `cbor:"th" json:"threshold"`
	// keepers who can use this account, no more than 1024
	// the account id must be one of them.
	Keepers     util.EthIDs         `cbor:"kp" json:"keepers"`
	Tokens      map[string]*big.Int `cbor:"tk" json:"tokens"`
	NonceTable  map[uint64][]uint64 `cbor:"nt" json:"nonceTable"` // map[expire][]nonce
	Approver    *util.EthID         `cbor:"ap,omitempty" json:"approver,omitempty"`
	ApproveList TxTypes             `cbor:"apl,omitempty" json:"approveList,omitempty"`
	// MaxTotalSupply only used with TokenAccount
	MaxTotalSupply *big.Int       `cbor:"mts,omitempty" json:"maxTotalSupply,omitempty"`
	Stake          *StakeConfig   `cbor:"st,omitempty" json:"stake,omitempty"`
	Lending        *LendingConfig `cbor:"le,omitempty" json:"lending,omitempty"`

	// external assignment fields
	Height    uint64     `cbor:"-" json:"height"`    // block's timestamp
	Timestamp uint64     `cbor:"-" json:"timestamp"` // block's timestamp
	ID        util.EthID `cbor:"-" json:"address"`
	raw       []byte     `cbor:"-" json:"-"`
}

// SyntacticVerify verifies that a *Account is well-formed.
func (a *Account) SyntacticVerify() error {
	var err error
	errp := util.ErrPrefix("Account.SyntacticVerify error: ")

	switch {
	case a == nil:
		return errp.Errorf("nil pointer")

	case a.Balance == nil || a.Balance.Sign() < 0:
		return errp.Errorf("invalid balance")

	case a.Keepers == nil:
		return errp.Errorf("invalid keepers")

	case len(a.Keepers) > MaxKeepers:
		return errp.Errorf("invalid keepers, too many")

	case int(a.Threshold) > len(a.Keepers):
		return errp.Errorf("invalid threshold")

	case a.Tokens == nil:
		return errp.Errorf("invalid tokens")

	case a.NonceTable == nil:
		return errp.Errorf("invalid nonceTable")

	case a.Approver != nil && *a.Approver == util.EthIDEmpty:
		return errp.Errorf("invalid approver")
	}

	if err = a.Keepers.CheckDuplicate(); err != nil {
		return errp.Errorf("invalid keepers, %v", err)
	}

	if err = a.Keepers.CheckEmptyID(); err != nil {
		return errp.Errorf("invalid keepers, %v", err)
	}

	if a.ApproveList != nil {
		if err = a.ApproveList.CheckDuplicate(); err != nil {
			return errp.Errorf("invalid approveList, %v", err)
		}

		for _, ty := range a.ApproveList {
			if !AllTxTypes.Has(ty) {
				return errp.Errorf("invalid TxType %s in approveList", ty)
			}
		}
	}

	switch a.Type {
	case NativeAccount:
		if a.MaxTotalSupply != nil {
			return errp.Errorf("invalid maxTotalSupply, should be nil")
		}
		if a.Stake != nil {
			return errp.Errorf("invalid stake on NativeAccount")
		}

	case TokenAccount:
		if a.Stake != nil {
			return errp.Errorf("invalid stake on TokenAccount")
		}
		if a.MaxTotalSupply == nil || a.MaxTotalSupply.Sign() < 0 {
			return errp.Errorf("invalid maxTotalSupply")
		}

	case StakeAccount:
		if a.MaxTotalSupply != nil {
			return errp.Errorf("invalid maxTotalSupply, should be nil")
		}
		if a.Stake == nil {
			return errp.Errorf("invalid stake on StakeAccount")
		}

		if err := a.Stake.SyntacticVerify(); err != nil {
			return err
		}

	default:
		return errp.Errorf("invalid type")
	}

	if a.Lending != nil {
		if err := a.Lending.SyntacticVerify(); err != nil {
			return err
		}
	}

	if a.raw, err = a.Marshal(); err != nil {
		return errp.ErrorIf(err)
	}
	return nil
}

func (a *Account) Bytes() []byte {
	if len(a.raw) == 0 {
		a.raw = MustMarshal(a)
	}
	return a.raw
}

func (a *Account) Unmarshal(data []byte) error {
	return util.ErrPrefix("Account.Unmarshal error: ").
		ErrorIf(util.UnmarshalCBOR(data, a))
}

func (a *Account) Marshal() ([]byte, error) {
	return util.ErrPrefix("Account.Marshal error: ").
		ErrorMap(util.MarshalCBOR(a))
}

type StakeConfig struct {
	_     struct{}         `cbor:",toarray"`
	Token util.TokenSymbol `json:"token"`
	// 0: account keepers can not use stake token
	// 1: account keepers can take a stake in other stake account
	// 2: in addition to 1, account keepers can transfer stake token to other account
	Type        uint16   `json:"type"`
	LockTime    uint64   `json:"lockTime"`
	WithdrawFee uint64   `json:"withdrawFee"` // 1_000_000 == 100%, should be in [1, 200_000]
	MinAmount   *big.Int `json:"minAmount"`
	MaxAmount   *big.Int `json:"maxAmount"`
}

// SyntacticVerify verifies that a *StakeConfig is well-formed.
func (c *StakeConfig) SyntacticVerify() error {
	errp := util.ErrPrefix("StakeConfig.SyntacticVerify error: ")

	switch {
	case c == nil:
		return errp.Errorf("nil pointer")

	case !c.Token.Valid():
		return errp.Errorf("invalid token %s", c.Token.GoString())

	case c.Type > 2:
		return errp.Errorf("invalid type")

	case c.WithdrawFee < 1 || c.WithdrawFee > 200_000:
		return errp.Errorf("invalid withdrawFee, should be in [1, 200_000]")

	case c.MinAmount == nil || c.MinAmount.Sign() < 1:
		return errp.Errorf("invalid minAmount")

	case c.MaxAmount == nil || c.MaxAmount.Cmp(c.MinAmount) < 0:
		return errp.Errorf("invalid maxAmount")
	}
	return nil
}

func (c *StakeConfig) Unmarshal(data []byte) error {
	return util.ErrPrefix("StakeConfig.Unmarshal error: ").
		ErrorIf(util.UnmarshalCBOR(data, c))
}

func (c *StakeConfig) Marshal() ([]byte, error) {
	return util.ErrPrefix("StakeConfig.Marshal error: ").
		ErrorMap(util.MarshalCBOR(c))
}

type LendingConfig struct {
	_ struct{} `cbor:",toarray"`

	Token           util.TokenSymbol `json:"token"`
	DailyInterest   uint64           `json:"dailyInterest"`   // 1_000_000 == 100%, should be in [1, 10_000]
	OverdueInterest uint64           `json:"overdueInterest"` // 1_000_000 == 100%, should be in [1, 10_000]
	MinAmount       *big.Int         `json:"minAmount"`
	MaxAmount       *big.Int         `json:"maxAmount"`
}

// SyntacticVerify verifies that a *LendingConfig is well-formed.
func (c *LendingConfig) SyntacticVerify() error {
	errp := util.ErrPrefix("LendingConfig.SyntacticVerify error: ")

	switch {
	case c == nil:
		return errp.Errorf("nil pointer")

	case !c.Token.Valid():
		return errp.Errorf("invalid token %s", c.Token.GoString())

	case c.DailyInterest < 1 || c.DailyInterest > 10_000:
		return errp.Errorf("invalid dailyInterest, should be in [1, 10_000]")

	case c.OverdueInterest < 1 || c.OverdueInterest > 10_000:
		return errp.Errorf("invalid overdueInterest, should be in [1, 10_000]")

	case c.MinAmount == nil || c.MinAmount.Sign() < 1:
		return errp.Errorf("invalid minAmount")

	case c.MaxAmount == nil || c.MaxAmount.Cmp(c.MinAmount) < 0:
		return errp.Errorf("invalid maxAmount")
	}
	return nil
}

func (c *LendingConfig) Unmarshal(data []byte) error {
	return util.ErrPrefix("LendingConfig.Unmarshal error: ").
		ErrorIf(util.UnmarshalCBOR(data, c))
}

func (c *LendingConfig) Marshal() ([]byte, error) {
	return util.ErrPrefix("LendingConfig.Marshal error: ").
		ErrorMap(util.MarshalCBOR(c))
}
