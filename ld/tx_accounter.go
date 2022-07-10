// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package ld

import (
	"math/big"

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
	Data        util.RawData `cbor:"d,omitempty" json:"data,omitempty"`

	// external assignment fields
	raw []byte `cbor:"-" json:"-"`
}

// SyntacticVerify verifies that a *TxAccounter is well-formed.
func (t *TxAccounter) SyntacticVerify() error {
	var err error
	errp := util.ErrPrefix("TxAccounter.SyntacticVerify error: ")

	switch {
	case t == nil:
		return errp.Errorf("nil pointer")

	case t.Name != "" && !util.ValidName(t.Name):
		return errp.Errorf("invalid name %q", t.Name)

	case t.Amount != nil && t.Amount.Sign() < 0:
		return errp.Errorf("invalid amount")
	}

	if t.Keepers != nil || t.Threshold != nil {
		switch {
		case t.Threshold == nil:
			return errp.Errorf("nil threshold together with keepers")

		case t.Keepers == nil:
			return errp.Errorf("nil keepers together with threshold")

		case int(*t.Threshold) > len(*t.Keepers):
			return errp.Errorf("invalid threshold, expected <= %d, got %d",
				len(*t.Keepers), *t.Threshold)

		case len(*t.Keepers) > MaxKeepers:
			return errp.Errorf("invalid keepers, expected <= %d, got %d",
				MaxKeepers, len(*t.Keepers))
		}

		if err = t.Keepers.CheckDuplicate(); err != nil {
			return errp.Errorf("invalid keepers, %v", err)
		}

		if err = t.Keepers.CheckEmptyID(); err != nil {
			return errp.Errorf("invalid keepers, %v", err)
		}
	}

	if t.ApproveList != nil {
		if err = t.ApproveList.CheckDuplicate(); err != nil {
			return errp.Errorf("invalid approveList, %v", err)
		}
		for _, ty := range t.ApproveList {
			if !AllTxTypes.Has(ty) {
				return errp.Errorf("invalid TxType %s in approveList", ty)
			}
		}
	}

	if t.raw, err = t.Marshal(); err != nil {
		return errp.ErrorIf(err)
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
	return util.ErrPrefix("TxAccounter.Unmarshal error: ").
		ErrorIf(util.UnmarshalCBOR(data, t))
}

func (t *TxAccounter) Marshal() ([]byte, error) {
	return util.ErrPrefix("TxAccounter.Marshal error: ").
		ErrorMap(util.MarshalCBOR(t))
}
