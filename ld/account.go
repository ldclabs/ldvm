// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package ld

import (
	"fmt"
	"math/big"

	"github.com/ldclabs/ldvm/ids"
	"github.com/ldclabs/ldvm/signer"
	"github.com/ldclabs/ldvm/util/encoding"
	"github.com/ldclabs/ldvm/util/erring"
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

const MaxKeepers = 64

type Account struct {
	Type AccountType `cbor:"t" json:"type"`
	// Nonce should increase 1 when sender issuing tx, but not increase when receiving
	Nonce uint64 `cbor:"n" json:"nonce"`
	// the decimals is 9, the smallest unit "NanoLDC" equal to gwei.
	Balance *big.Int `cbor:"b" json:"balance"`
	// M of N threshold signatures, aka MultiSig: threshold is m, keepers length is n.
	// The minimum value is 1, the maximum value is len(keepers)
	Threshold uint16 `cbor:"th" json:"threshold"`
	// keepers who can use this account, no more than 64
	// the account id must be one of them.
	Keepers     signer.Keys         `cbor:"kp" json:"keepers"`
	Tokens      map[string]*big.Int `cbor:"tk" json:"tokens"`
	NonceTable  map[uint64][]uint64 `cbor:"nt" json:"nonceTable"` // map[expire][]nonce
	Approver    signer.Key          `cbor:"ap,omitempty" json:"approver,omitempty"`
	ApproveList TxTypes             `cbor:"apl,omitempty" json:"approveList,omitempty"`
	// MaxTotalSupply only used with TokenAccount
	MaxTotalSupply *big.Int       `cbor:"mts,omitempty" json:"maxTotalSupply,omitempty"`
	Stake          *StakeConfig   `cbor:"st,omitempty" json:"stake,omitempty"`
	Lending        *LendingConfig `cbor:"le,omitempty" json:"lending,omitempty"`

	// external assignment fields
	Height    uint64      `cbor:"-" json:"height"`    // block's timestamp
	Timestamp uint64      `cbor:"-" json:"timestamp"` // block's timestamp
	ID        ids.Address `cbor:"-" json:"address"`
	raw       []byte      `cbor:"-" json:"-"`
}

// SyntacticVerify verifies that a *Account is well-formed.
func (a *Account) SyntacticVerify() error {
	var err error
	errp := erring.ErrPrefix("ld.Account.SyntacticVerify: ")

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
	}

	if err = a.Keepers.Valid(); err != nil {
		return errp.Errorf("invalid keepers, %v", err)
	}

	if a.Approver != nil {
		if err = a.Approver.Valid(); err != nil {
			return errp.Errorf("invalid approver, %v", err)
		}
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

func (a *Account) CheckAsFrom(txType TxType) error {
	errp := erring.ErrPrefix("ld.Account.CheckAsFrom: ")

	switch a.Type {
	case TokenAccount:
		switch {
		case TokenFromTxTypes.Has(txType):
			// just go ahead
		default:
			return errp.Errorf("can't use TokenAccount as sender for %s", txType.String())
		}

	case StakeAccount:
		if a.Stake == nil {
			return errp.Errorf("invalid StakeAccount as sender for %s", txType.String())
		}

		ty := a.Stake.Type
		if ty > 2 {
			return errp.Errorf("can't use unknown type %d StakeAccount as sender for %s",
				ty, txType.String())
		}

		// 0: account keepers can not use stake token
		// 1: account keepers can take a stake in other stake account
		// 2: in addition to 1, account keepers can transfer stake token to other account
		switch {
		case StakeFromTxTypes0.Has(txType):
			// just go ahead
		case StakeFromTxTypes1.Has(txType):
			if ty < 1 {
				return errp.Errorf("can't use type %d StakeAccount as sender for %s",
					ty, txType.String())
			}

		case StakeFromTxTypes2.Has(txType):
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

func (a *Account) CheckAsTo(txType TxType) error {
	errp := erring.ErrPrefix("ld.Account.CheckAsTo: ")

	switch a.Type {
	case TokenAccount:
		switch {
		case TokenToTxTypes.Has(txType):
			// just go ahead
		default:
			return errp.Errorf("can't use TokenAccount as recipient for %s", txType.String())
		}

	case StakeAccount:
		switch {
		case StakeToTxTypes.Has(txType):
			// just go ahead
		default:
			return errp.Errorf("can't use StakeAccount as recipient for %s", txType.String())
		}
	}

	return nil
}

// func (a *Account) VerifyOne(digestHash []byte, sigs signer.Sigs) bool {
// 	switch {
// 	case a.ID == ids.LDCAccount:
// 		return false

// 	case len(a.Keepers) == 0:
// 		return false

// 	default:
// 		return a.Keepers.Verify(digestHash, sigs, 1)
// 	}
// }

func (a *Account) Verify(digestHash []byte, sigs signer.Sigs, accountKey signer.Key) bool {
	switch {
	case a.ID == ids.LDCAccount:
		return false

	case len(a.Keepers) == 0 && accountKey.IsAddress(a.ID):
		return accountKey.Verify(digestHash, sigs)

	default:
		return a.Keepers.Verify(digestHash, sigs, a.Threshold)
	}
}

func (a *Account) VerifyPlus(digestHash []byte, sigs signer.Sigs, accountKey signer.Key) bool {
	switch {
	case a.ID == ids.LDCAccount:
		return false

	case len(a.Keepers) == 0 && accountKey.IsAddress(a.ID):
		return accountKey.Verify(digestHash, sigs)

	default:
		return a.Keepers.VerifyPlus(digestHash, sigs, a.Threshold)
	}
}

func (a *Account) Bytes() []byte {
	if len(a.raw) == 0 {
		a.raw = MustMarshal(a)
	}
	return a.raw
}

func (a *Account) Unmarshal(data []byte) error {
	return erring.ErrPrefix("ld.Account.Unmarshal: ").
		ErrorIf(encoding.UnmarshalCBOR(data, a))
}

func (a *Account) Marshal() ([]byte, error) {
	return erring.ErrPrefix("ld.Account.Marshal: ").
		ErrorMap(encoding.MarshalCBOR(a))
}

type StakeConfig struct {
	_     struct{}        `cbor:",toarray"`
	Token ids.TokenSymbol `json:"token"`
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
	errp := erring.ErrPrefix("ld.StakeConfig.SyntacticVerify: ")

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
	return erring.ErrPrefix("ld.StakeConfig.Unmarshal: ").
		ErrorIf(encoding.UnmarshalCBOR(data, c))
}

func (c *StakeConfig) Marshal() ([]byte, error) {
	return erring.ErrPrefix("ld.StakeConfig.Marshal: ").
		ErrorMap(encoding.MarshalCBOR(c))
}

type LendingConfig struct {
	_ struct{} `cbor:",toarray"`

	Token           ids.TokenSymbol `json:"token"`
	DailyInterest   uint64          `json:"dailyInterest"`   // 1_000_000 == 100%, should be in [1, 10_000]
	OverdueInterest uint64          `json:"overdueInterest"` // 1_000_000 == 100%, should be in [1, 10_000]
	MinAmount       *big.Int        `json:"minAmount"`
	MaxAmount       *big.Int        `json:"maxAmount"`
}

// SyntacticVerify verifies that a *LendingConfig is well-formed.
func (c *LendingConfig) SyntacticVerify() error {
	errp := erring.ErrPrefix("ld.LendingConfig.SyntacticVerify: ")

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
	return erring.ErrPrefix("ld.LendingConfig.Unmarshal: ").
		ErrorIf(encoding.UnmarshalCBOR(data, c))
}

func (c *LendingConfig) Marshal() ([]byte, error) {
	return erring.ErrPrefix("ld.LendingConfig.Marshal: ").
		ErrorMap(encoding.MarshalCBOR(c))
}
