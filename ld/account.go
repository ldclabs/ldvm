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
	ApproveList TxTypes                       `cbor:"apl,omitempty" json:"approveList,omitempty"`
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
	var err error
	errPrefix := "Account.SyntacticVerify failed:"

	switch {
	case a == nil:
		return fmt.Errorf("%s nil pointer", errPrefix)

	case a.Balance == nil || a.Balance.Sign() < 0:
		return fmt.Errorf("%s invalid balance", errPrefix)

	case a.Keepers == nil:
		return fmt.Errorf("%s invalid keepers", errPrefix)

	case len(a.Keepers) > math.MaxUint8:
		return fmt.Errorf("%s invalid keepers, too many", errPrefix)

	case int(a.Threshold) > len(a.Keepers):
		return fmt.Errorf("%s invalid threshold", errPrefix)

	case a.Tokens == nil:
		return fmt.Errorf("%s invalid tokens", errPrefix)

	case a.NonceTable == nil:
		return fmt.Errorf("%s invalid nonceTable", errPrefix)

	case a.Approver != nil && *a.Approver == util.EthIDEmpty:
		return fmt.Errorf("%s invalid approver", errPrefix)
	}

	if err = a.Keepers.CheckDuplicate(); err != nil {
		return fmt.Errorf("%s invalid keepers, %v", errPrefix, err)
	}

	if err = a.Keepers.CheckEmptyID(); err != nil {
		return fmt.Errorf("%s invalid keepers, %v", errPrefix, err)
	}

	if a.ApproveList != nil {
		if err = a.ApproveList.CheckDuplicate(); err != nil {
			return fmt.Errorf("%s invalid approveList, %v", errPrefix, err)
		}

		for _, ty := range a.ApproveList {
			if !AllTxTypes.Has(ty) {
				return fmt.Errorf("%s invalid TxType %s in approveList", errPrefix, ty)
			}
		}
	}

	switch a.Type {
	case NativeAccount:
		if a.MaxTotalSupply != nil {
			return fmt.Errorf("%s maxTotalSupply should be nil", errPrefix)
		}
		if a.Stake != nil || a.StakeLedger != nil {
			return fmt.Errorf("%s invalid stake on NativeAccount", errPrefix)
		}

	case TokenAccount:
		if a.Stake != nil || a.StakeLedger != nil {
			return fmt.Errorf("%s invalid stake on TokenAccount", errPrefix)
		}
		if a.MaxTotalSupply == nil || a.MaxTotalSupply.Sign() < 0 {
			return fmt.Errorf("%s invalid maxTotalSupply", errPrefix)
		}

	case StakeAccount:
		if a.MaxTotalSupply != nil {
			return fmt.Errorf("%s maxTotalSupply should be nil", errPrefix)
		}
		if a.Stake == nil {
			return fmt.Errorf("%s invalid stake on StakeAccount", errPrefix)
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
				return fmt.Errorf("%s invalid amount on StakeEntry", errPrefix)
			}
			if entry.Approver != nil && *entry.Approver == util.EthIDEmpty {
				return fmt.Errorf("%s invalid approver on StakeEntry", errPrefix)
			}
		}

	default:
		return fmt.Errorf("%s invalid type", errPrefix)
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
				return fmt.Errorf("%s invalid amount on StakeEntry", errPrefix)
			}
		}
	}

	if a.raw, err = a.Marshal(); err != nil {
		return fmt.Errorf("%s %v", errPrefix, err)
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
	errPrefix := "StakeConfig.SyntacticVerify failed:"

	switch {
	case c == nil:
		return fmt.Errorf("%s nil pointer", errPrefix)

	case !c.Token.Valid():
		return fmt.Errorf("%s invalid token %s", errPrefix, c.Token.GoString())

	case c.Type > 2:
		return fmt.Errorf("%s invalid type", errPrefix)

	case c.WithdrawFee < 1 || c.WithdrawFee > 200_000:
		return fmt.Errorf("%s invalid withdrawFee, should be in [1, 200_000]", errPrefix)

	case c.MinAmount == nil || c.MinAmount.Sign() < 1:
		return fmt.Errorf("%s invalid minAmount", errPrefix)

	case c.MaxAmount == nil || c.MaxAmount.Cmp(c.MinAmount) < 0:
		return fmt.Errorf("%s invalid maxAmount", errPrefix)
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
	errPrefix := "LendingConfig.SyntacticVerify failed:"
	switch {
	case c == nil:
		return fmt.Errorf("%s nil pointer", errPrefix)

	case !c.Token.Valid():
		return fmt.Errorf("%s invalid token %s", errPrefix, c.Token.GoString())

	case c.DailyInterest < 1 || c.DailyInterest > 10_000:
		return fmt.Errorf("%s invalid dailyInterest, should be in [1, 10_000]", errPrefix)

	case c.OverdueInterest < 1 || c.OverdueInterest > 10_000:
		return fmt.Errorf("%s invalid overdueInterest, should be in [1, 10_000]", errPrefix)

	case c.MinAmount == nil || c.MinAmount.Sign() < 1:
		return fmt.Errorf("%s invalid minAmount", errPrefix)

	case c.MaxAmount == nil || c.MaxAmount.Cmp(c.MinAmount) < 0:
		return fmt.Errorf("%s invalid maxAmount", errPrefix)
	}
	return nil
}

func (c *LendingConfig) Unmarshal(data []byte) error {
	return DecMode.Unmarshal(data, c)
}

func (c *LendingConfig) Marshal() ([]byte, error) {
	return EncMode.Marshal(c)
}
