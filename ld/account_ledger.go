// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package ld

import (
	"math/big"

	"github.com/fxamacker/cbor/v2"
	"github.com/ldclabs/ldvm/signer"
	"github.com/ldclabs/ldvm/util/encoding"
	"github.com/ldclabs/ldvm/util/erring"
)

type AccountLedger struct {
	Lending map[cbor.ByteString]*LendingEntry `cbor:"l"`
	Stake   map[cbor.ByteString]*StakeEntry   `cbor:"s"`

	// external assignment fields
	raw []byte `cbor:"-"`
}

// SyntacticVerify verifies that a *AccountLedger is well-formed.
func (a *AccountLedger) SyntacticVerify() error {
	var err error
	errp := erring.ErrPrefix("ld.AccountLedger.SyntacticVerify: ")

	if a == nil {
		return errp.Errorf("nil pointer")
	}

	if a.Lending == nil {
		a.Lending = make(map[cbor.ByteString]*LendingEntry)
	}

	for _, entry := range a.Lending {
		if entry == nil || entry.Amount == nil || entry.Amount.Sign() <= 0 {
			return errp.Errorf("invalid amount on LendingEntry")
		}
	}

	if a.Stake == nil {
		a.Stake = make(map[cbor.ByteString]*StakeEntry)
	}

	for _, entry := range a.Stake {
		if entry == nil || entry.Amount == nil || entry.Amount.Sign() < 0 ||
			(entry.Amount.Sign() == 0 && entry.Approver == nil) {
			return errp.Errorf("invalid amount on StakeEntry")
		}

		if entry.Approver != nil {
			if err := entry.Approver.Valid(); err != nil {
				return errp.Errorf("invalid approver on StakeEntry, %v", err)
			}
		}
	}

	if a.raw, err = a.Marshal(); err != nil {
		return errp.ErrorIf(err)
	}
	return nil
}

func (a *AccountLedger) Bytes() []byte {
	if len(a.raw) == 0 {
		a.raw = MustMarshal(a)
	}
	return a.raw
}

func (a *AccountLedger) Unmarshal(data []byte) error {
	if len(data) == 0 {
		return nil
	}

	return erring.ErrPrefix("ld.AccountLedger.Unmarshal: ").
		ErrorIf(encoding.UnmarshalCBOR(data, a))
}

func (a *AccountLedger) Marshal() ([]byte, error) {
	return erring.ErrPrefix("ld.AccountLedger.Marshal: ").
		ErrorMap(encoding.MarshalCBOR(a))
}

type LendingEntry struct {
	_ struct{} `cbor:",toarray"`

	Amount   *big.Int `json:"amount"`
	UpdateAt uint64   `json:"updateAt"`
	DueTime  uint64   `json:"dueTime"`
}

type StakeEntry struct {
	_ struct{} `cbor:",toarray"`

	Amount   *big.Int    `json:"amount"`
	LockTime uint64      `json:"lockTime"`
	Approver *signer.Key `json:"approver"`
}
