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
	TokenAccount              // The first 10 bytes of account address must be 0
	StakeAccount              // The first byte of account address must be $
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
	Keepers        []util.EthID                  `cbor:"kp" json:"keepers"`
	NonceTable     map[uint64][]uint64           `cbor:"nt" json:"nonceTable"` // map[expire][]nonce
	Tokens         map[util.TokenSymbol]*big.Int `cbor:"tk" json:"tokens"`
	MaxTotalSupply *big.Int                      `cbor:"mts,omitempty" json:"maxTotalSupply,omitempty"` // only used with TokenAccount
	Stake          *StakeConfig                  `cbor:"st,omitempty" json:"stake,omitempty"`
	StakeLedger    Ledger                        `cbor:"stl,omitempty" json:"stakeLedger,omitempty"`
	Lending        *LendingConfig                `cbor:"le,omitempty" json:"lending,omitempty"`
	LendingLedger  Ledger                        `cbor:"lel,omitempty" json:"lendingLedger,omitempty"`

	// external assignment
	Height    uint64     `cbor:"-" json:"height"`    // block's timestamp
	Timestamp uint64     `cbor:"-" json:"timestamp"` // block's timestamp
	ID        util.EthID `cbor:"-" json:"address"`
	raw       []byte     `cbor:"-" json:"-"`
}

type Ledger map[util.EthID]*LedgerEntry

type LedgerEntry struct {
	_ struct{} `cbor:",toarray"`

	Amount   *big.Int `json:"amount"`
	UpdateAt uint64   `json:"updateAt,omitempty"`
	DueTime  uint64   `json:"dueTime,omitempty"`
}

type StakeConfig struct {
	_ struct{} `cbor:",toarray"`
	// 0: account keepers can not use stake token
	// 1: account keepers can take a stake in other stake account
	// 2: in addition to 1, account keepers can transfer stake token to other account
	// 3: in addition to 2, account keepers can liquidate shareholder (use with lending)
	Type        uint8    `json:"type"`
	Token       string   `json:"token"`
	LockTime    uint64   `json:"lockTime"`
	WithdrawFee uint64   `json:"withdrawFee"` // 1_000_000 == 100%, should be in [1, 200_000]
	MinAmount   *big.Int `json:"minAmount"`
	MaxAmount   *big.Int `json:"maxAmount"`

	// external assignment
	TokenID util.TokenSymbol `cbor:"-" json:"-"`
}

type LendingConfig struct {
	_ struct{} `cbor:",toarray"`

	Token           string   `json:"token"`
	DailyInterest   uint64   `json:"dailyInterest"`   // 1_000_000 == 100%, should be in [1, 10_000]
	OverdueInterest uint64   `json:"overdueInterest"` // 1_000_000 == 100%, should be in [1, 10_000]
	MinAmount       *big.Int `json:"minAmount"`
	MaxAmount       *big.Int `json:"maxAmount"`

	// external assignment
	TokenID util.TokenSymbol `cbor:"-" json:"-"`
}

// SyntacticVerify verifies that a *StakeConfig is well-formed.
func (c *StakeConfig) SyntacticVerify() error {
	if c == nil {
		return fmt.Errorf("invalid StakeConfig")
	}
	if c.Type > 3 {
		return fmt.Errorf("invalid StakeConfig type")
	}
	token, err := util.NewSymbol(c.Token)
	if err != nil {
		return err
	}
	c.TokenID = token

	if c.WithdrawFee < 1 || c.WithdrawFee > 200_000 {
		return fmt.Errorf("invalid WithdrawFee, should be in [1, 200_000]")
	}

	if c.MinAmount == nil || c.MinAmount.Sign() < 1 {
		return fmt.Errorf("invalid MinAmount")
	}
	if c.MaxAmount == nil || c.MaxAmount.Cmp(c.MinAmount) < 0 {
		return fmt.Errorf("invalid MaxAmount")
	}
	return nil
}

func (c *StakeConfig) Unmarshal(data []byte) error {
	return DecMode.Unmarshal(data, c)
}

func (c *StakeConfig) Marshal() ([]byte, error) {
	return EncMode.Marshal(c)
}

// SyntacticVerify verifies that a *LendingConfig is well-formed.
func (c *LendingConfig) SyntacticVerify() error {
	if c == nil {
		return fmt.Errorf("invalid LendingConfig")
	}
	token, err := util.NewSymbol(c.Token)
	if err != nil {
		return err
	}
	c.TokenID = token

	if c.DailyInterest < 1 || c.DailyInterest > 10_000 {
		return fmt.Errorf("invalid DailyInterest, should be in [1, 10_000]")
	}
	if c.OverdueInterest < 1 || c.OverdueInterest > 10_000 {
		return fmt.Errorf("invalid OverdueInterest, should be in [1, 10_000]")
	}

	if c.MinAmount == nil || c.MinAmount.Sign() < 1 {
		return fmt.Errorf("invalid MinAmount")
	}
	if c.MaxAmount == nil || c.MaxAmount.Cmp(c.MinAmount) < 0 {
		return fmt.Errorf("invalid MaxAmount")
	}
	return nil
}

func (c *LendingConfig) Unmarshal(data []byte) error {
	return DecMode.Unmarshal(data, c)
}

func (c *LendingConfig) Marshal() ([]byte, error) {
	return EncMode.Marshal(c)
}

// func (a *Account) Copy() *Account {
// 	x := new(Account)
// 	*x = *a
// 	x.Balance = new(big.Int).Set(a.Balance)
// 	x.Keepers = make([]ids.ShortID, len(a.Keepers))
// 	copy(x.Keepers, a.Keepers)
// 	x.NonceTable = make(map[uint64][]uint64, len(a.NonceTable))
// 	for k, v := range a.NonceTable {
// 		x.NonceTable[k] = make([]uint64, len(v))
// 		for i := range v {
// 			x.NonceTable[k][i] = v[i]
// 		}
// 	}
// 	x.Tokens = make(map[ids.ShortID]*big.Int, len(a.Tokens))
// 	for k, v := range a.Tokens {
// 		x.Tokens[k] = new(big.Int).Set(v)
// 	}
// 	if a.MaxTotalSupply != nil {
// 		x.MaxTotalSupply = new(big.Int).Set(a.MaxTotalSupply)
// 	}
// 	x.raw = nil
// 	return x
// }

// SyntacticVerify verifies that a *Account is well-formed.
func (a *Account) SyntacticVerify() error {
	if a == nil {
		return fmt.Errorf("invalid Account")
	}

	if a.Balance == nil || a.Balance.Sign() < 0 {
		return fmt.Errorf("invalid balance")
	}
	if len(a.Keepers) > math.MaxUint8 {
		return fmt.Errorf("invalid keepers, too many")
	}
	if int(a.Threshold) > len(a.Keepers) {
		return fmt.Errorf("invalid threshold")
	}
	if a.NonceTable == nil {
		return fmt.Errorf("invalid nonceTable")
	}
	if a.Tokens == nil {
		return fmt.Errorf("invalid tokens")
	}

	switch a.Type {
	case NativeAccount:
		if a.MaxTotalSupply != nil {
			return fmt.Errorf("invalid maxTotalSupply, should be nil")
		}
		if a.Stake != nil || len(a.StakeLedger) > 0 {
			return fmt.Errorf("invalid stake on NativeAccount")
		}
	case TokenAccount:
		if a.Stake != nil || len(a.StakeLedger) > 0 {
			return fmt.Errorf("invalid stake on TokenAccount")
		}
		if a.MaxTotalSupply == nil || a.MaxTotalSupply.Sign() < 0 {
			return fmt.Errorf("invalid maxTotalSupply")
		}
	case StakeAccount:
		if a.MaxTotalSupply != nil {
			return fmt.Errorf("invalid maxTotalSupply, should be nil")
		}
		if a.Stake == nil || a.StakeLedger == nil {
			return fmt.Errorf("invalid stake on StakeAccount")
		}
	default:
		return fmt.Errorf("invalid type")
	}

	if _, err := a.Marshal(); err != nil {
		return fmt.Errorf("Account marshal error: %v", err)
	}
	return nil
}

func (a *Account) Bytes() []byte {
	if len(a.raw) == 0 {
		MustMarshal(a)
	}
	return a.raw
}

func (a *Account) Unmarshal(data []byte) error {
	return DecMode.Unmarshal(data, a)
}

func (a *Account) Marshal() ([]byte, error) {
	data, err := EncMode.Marshal(a)
	if err != nil {
		return nil, err
	}
	a.raw = data
	return data, nil
}
