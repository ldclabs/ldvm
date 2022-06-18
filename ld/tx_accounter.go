// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package ld

import (
	"fmt"
	"math/big"
	"strconv"

	"github.com/ldclabs/ldvm/util"
)

// TxAccounter
type TxAccounter struct {
	Threshold   *uint16      `cbor:"th,omitempty" json:"threshold,omitempty"`
	Keepers     *util.EthIDs `cbor:"kp,omitempty" json:"keepers,omitempty"`
	Approver    *util.EthID  `cbor:"ap,omitempty" json:"approver,omitempty"`
	ApproveList TxTypes      `cbor:"apl,omitempty" json:"approveList,omitempty"`
	Amount      *big.Int     `cbor:"a,omitempty" json:"amount,omitempty"`
	Name        string       `cbor:"n,omitempty" json:"name,omitempty"`
	Data        RawData      `cbor:"d,omitempty" json:"data,omitempty"`

	// external assignment fields
	raw []byte `cbor:"-" json:"-"`
}

// SyntacticVerify verifies that a *TxAccounter is well-formed.
func (t *TxAccounter) SyntacticVerify() error {
	var err error
	errPrefix := "TxAccounter.SyntacticVerify failed:"

	switch {
	case t == nil:
		return fmt.Errorf("%s nil pointer", errPrefix)

	case t.Name != "" && !util.ValidName(t.Name):
		return fmt.Errorf("%s invalid name %s", errPrefix, strconv.Quote(t.Name))

	case t.Amount != nil && t.Amount.Sign() < 0:
		return fmt.Errorf("%s invalid amount", errPrefix)
	}

	if t.Keepers != nil || t.Threshold != nil {
		switch {
		case t.Threshold == nil:
			return fmt.Errorf("%s nil threshold together with keepers", errPrefix)

		case t.Keepers == nil:
			return fmt.Errorf("%s nil keepers together with threshold", errPrefix)

		case int(*t.Threshold) > len(*t.Keepers):
			return fmt.Errorf("%s invalid threshold, expected <= %d, got %d",
				errPrefix, len(*t.Keepers), *t.Threshold)

		case len(*t.Keepers) > MaxKeepers:
			return fmt.Errorf("%s invalid keepers, expected <= %d, got %d",
				errPrefix, MaxKeepers, len(*t.Keepers))
		}

		if err = t.Keepers.CheckDuplicate(); err != nil {
			return fmt.Errorf("%s invalid keepers, %v", errPrefix, err)
		}

		if err = t.Keepers.CheckEmptyID(); err != nil {
			return fmt.Errorf("%s invalid keepers, %v", errPrefix, err)
		}
	}

	if t.ApproveList != nil {
		if err = t.ApproveList.CheckDuplicate(); err != nil {
			return fmt.Errorf("%s invalid approveList, %v", errPrefix, err)
		}
		for _, ty := range t.ApproveList {
			if !AllTxTypes.Has(ty) {
				return fmt.Errorf("%s invalid TxType %s in approveList", errPrefix, ty)
			}
		}
	}

	if t.raw, err = t.Marshal(); err != nil {
		return fmt.Errorf("%s %v", errPrefix, err)
	}
	return nil
}

func (t *TxAccounter) Bytes() []byte {
	if len(t.raw) == 0 {
		t.raw = MustMarshal(t)
	}
	return t.raw
}

func (t *TxAccounter) Unmarshal(data []byte) error {
	return UnmarshalCBOR(data, t)
}

func (t *TxAccounter) Marshal() ([]byte, error) {
	return MarshalCBOR(t)
}
