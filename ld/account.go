// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package ld

import (
	"fmt"
	"math"
	"math/big"

	"github.com/ldclabs/ldvm/util"
)

// AccountType is an uint8 representing the type of account
type AccountType uint8

const (
	NativeAccount AccountType = iota
	TokenAccount              // The first byte of account address must be $
	StakeAccount              // The first byte of account address must be #
)

type Account struct {
	Type AccountType `cbor:"t" json:"type"`
	// Nonce should increase 1 when sender issuing tx, but not increase when receiving
	Nonce uint64 `cbor:"n" json:"nonce"`
	// the decimals is 9, the smallest unit "NanoLDC" equal to gwei.
	Balance *big.Int `cbor:"b" json:"balance"`
	// M of N threshold signatures, aka MultiSig: threshold is m, keepers length is n.
	// The minimum value is 1, the maximum value is len(keepers)
	Threshold uint8 `cbor:"th" json:"threshold"`
	// keepers who can use this account, no more than 255
	// the account id must be one of them.
	Keepers     util.EthIDs                   `cbor:"kp" json:"keepers"`
	Tokens      map[util.TokenSymbol]*big.Int `cbor:"tk" json:"tokens"`
	NonceTable  map[uint64][]uint64           `cbor:"nt" json:"nonceTable"` // map[expire][]nonce
	Approver    *util.EthID                   `cbor:"ap,omitempty" json:"approver,omitempty"`
	ApproveList []TxType                      `cbor:"apl,omitempty" json:"approveList,omitempty"`
	// MaxTotalSupply only used with TokenAccount
	MaxTotalSupply *big.Int                     `cbor:"mts,omitempty" json:"maxTotalSupply,omitempty"`
	Stake          *StakeConfig                 `cbor:"st,omitempty" json:"stake,omitempty"`
	StakeLedger    map[util.EthID]*StakeEntry   `cbor:"stl,omitempty" json:"stakeLedger,omitempty"`
	Lending        *LendingConfig               `cbor:"le,omitempty" json:"lending,omitempty"`
	LendingLedger  map[util.EthID]*LendingEntry `cbor:"lel,omitempty" json:"lendingLedger,omitempty"`

	// external assignment fields
	Height    uint64     `cbor:"-" json:"height"`    // block's timestamp
	Timestamp uint64     `cbor:"-" json:"timestamp"` // block's timestamp
	ID        util.EthID `cbor:"-" json:"address"`
	raw       []byte     `cbor:"-" json:"-"`
}

type LendingEntry struct {
	_ struct{} `cbor:",toarray"`

	Amount   *big.Int `json:"amount"`
	UpdateAt uint64   `json:"updateAt,omitempty"`
	DueTime  uint64   `json:"dueTime,omitempty"`
}

type StakeEntry struct {
	_ struct{} `cbor:",toarray"`

	Amount   *big.Int    `json:"amount"`
	LockTime uint64      `json:"lockTime,omitempty"`
	Approver *util.EthID `json:"approver,omitempty"`
}

// SyntacticVerify verifies that a *Account is well-formed.
func (a *Account) SyntacticVerify() error {
	switch {
	case a == nil:
		return fmt.Errorf("Account.SyntacticVerify failed: nil pointer")
	case a.Balance == nil || a.Balance.Sign() < 0:
		return fmt.Errorf("Account.SyntacticVerify failed: invalid balance")
	case a.Keepers == nil:
		return fmt.Errorf("Account.SyntacticVerify failed: invalid keepers")
	case len(a.Keepers) > math.MaxUint8:
		return fmt.Errorf("Account.SyntacticVerify failed: invalid keepers, too many")
	case int(a.Threshold) > len(a.Keepers):
		return fmt.Errorf("Account.SyntacticVerify failed: invalid threshold")
	case a.Tokens == nil:
		return fmt.Errorf("Account.SyntacticVerify failed: invalid tokens")
	case a.NonceTable == nil:
		return fmt.Errorf("Account.SyntacticVerify failed: invalid nonceTable")
	}

	switch a.Type {
	case NativeAccount:
		if a.MaxTotalSupply != nil {
			return fmt.Errorf("Account.SyntacticVerify failed: maxTotalSupply should be nil")
		}
		if a.Stake != nil || a.StakeLedger != nil {
			return fmt.Errorf("Account.SyntacticVerify failed: invalid stake on NativeAccount")
		}
	case TokenAccount:
		if a.Stake != nil || a.StakeLedger != nil {
			return fmt.Errorf("Account.SyntacticVerify failed: invalid stake on TokenAccount")
		}
		if a.MaxTotalSupply == nil || a.MaxTotalSupply.Sign() < 0 {
			return fmt.Errorf("Account.SyntacticVerify failed: invalid maxTotalSupply")
		}
	case StakeAccount:
		if a.MaxTotalSupply != nil {
			return fmt.Errorf("Account.SyntacticVerify failed: maxTotalSupply should be nil")
		}
		if a.Stake == nil {
			return fmt.Errorf("Account.SyntacticVerify failed: invalid stake on StakeAccount")
		}
		if a.StakeLedger == nil {
			a.StakeLedger = make(map[util.EthID]*StakeEntry)
		}
		if err := a.Stake.SyntacticVerify(); err != nil {
			return err
		}
		for _, entry := range a.StakeLedger {
			if entry.Amount == nil || entry.Amount.Sign() < 0 ||
				(entry.Amount.Sign() == 0 && entry.Approver == nil) {
				return fmt.Errorf("Account.SyntacticVerify failed: invalid amount on StakeEntry")
			}
			if entry.Approver != nil && *entry.Approver == util.EthIDEmpty {
				return fmt.Errorf("Account.SyntacticVerify failed: invalid approver on StakeEntry")
			}
		}
	default:
		return fmt.Errorf("Account.SyntacticVerify failed: invalid type")
	}

	if a.Lending != nil {
		if a.LendingLedger == nil {
			a.LendingLedger = make(map[util.EthID]*LendingEntry)
		}
		if err := a.Lending.SyntacticVerify(); err != nil {
			return err
		}
		for _, entry := range a.LendingLedger {
			if entry.Amount == nil || entry.Amount.Sign() <= 0 {
				return fmt.Errorf("Account.SyntacticVerify failed: invalid amount on StakeEntry")
			}
		}
	}

	var err error
	if a.raw, err = a.Marshal(); err != nil {
		return fmt.Errorf("Account.SyntacticVerify marshal error: %v", err)
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
	return DecMode.Unmarshal(data, a)
}

func (a *Account) Marshal() ([]byte, error) {
	return EncMode.Marshal(a)
}

type StakeConfig struct {
	_     struct{}         `cbor:",toarray"`
	Token util.TokenSymbol `json:"token"`
	// 0: account keepers can not use stake token
	// 1: account keepers can take a stake in other stake account
	// 2: in addition to 1, account keepers can transfer stake token to other account
	Type        uint8    `json:"type"`
	LockTime    uint64   `json:"lockTime"`
	WithdrawFee uint64   `json:"withdrawFee"` // 1_000_000 == 100%, should be in [1, 200_000]
	MinAmount   *big.Int `json:"minAmount"`
	MaxAmount   *big.Int `json:"maxAmount"`
}

// SyntacticVerify verifies that a *StakeConfig is well-formed.
func (c *StakeConfig) SyntacticVerify() error {
	switch {
	case c == nil:
		return fmt.Errorf("StakeConfig.SyntacticVerify failed: nil pointer")
	case !c.Token.Valid():
		return fmt.Errorf("StakeConfig.SyntacticVerify failed: invalid token %s", c.Token.GoString())
	case c.Type > 2:
		return fmt.Errorf("StakeConfig.SyntacticVerify failed: invalid type")
	case c.WithdrawFee < 1 || c.WithdrawFee > 200_000:
		return fmt.Errorf(
			"StakeConfig.SyntacticVerify failed: invalid withdrawFee, should be in [1, 200_000]")
	case c.MinAmount == nil || c.MinAmount.Sign() < 1:
		return fmt.Errorf("StakeConfig.SyntacticVerify failed: invalid minAmount")
	case c.MaxAmount == nil || c.MaxAmount.Cmp(c.MinAmount) < 0:
		return fmt.Errorf("StakeConfig.SyntacticVerify failed: invalid maxAmount")
	}
	return nil
}

func (c *StakeConfig) Unmarshal(data []byte) error {
	return DecMode.Unmarshal(data, c)
}

func (c *StakeConfig) Marshal() ([]byte, error) {
	return EncMode.Marshal(c)
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
	switch {
	case c == nil:
		return fmt.Errorf("LendingConfig.SyntacticVerify failed: nil pointer")
	case !c.Token.Valid():
		return fmt.Errorf("LendingConfig.SyntacticVerify failed: invalid token %s", c.Token.GoString())
	case c.DailyInterest < 1 || c.DailyInterest > 10_000:
		return fmt.Errorf(
			"LendingConfig.SyntacticVerify failed: invalid dailyInterest, should be in [1, 10_000]")
	case c.OverdueInterest < 1 || c.OverdueInterest > 10_000:
		return fmt.Errorf(
			"LendingConfig.SyntacticVerify failed: invalid overdueInterest, should be in [1, 10_000]")
	case c.MinAmount == nil || c.MinAmount.Sign() < 1:
		return fmt.Errorf("LendingConfig.SyntacticVerify failed: invalid minAmount")
	case c.MaxAmount == nil || c.MaxAmount.Cmp(c.MinAmount) < 0:
		return fmt.Errorf("LendingConfig.SyntacticVerify failed: invalid maxAmount")
	}

	return nil
}

func (c *LendingConfig) Unmarshal(data []byte) error {
	return DecMode.Unmarshal(data, c)
}

func (c *LendingConfig) Marshal() ([]byte, error) {
	return EncMode.Marshal(c)
}
