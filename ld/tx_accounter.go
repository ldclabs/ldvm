// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package ld

import (
	"fmt"
	"math"
	"math/big"
	"strconv"

	"github.com/ldclabs/ldvm/util"
)

// TxAccounter
type TxAccounter struct {
	Threshold   uint8        `cbor:"th,omitempty" json:"threshold,omitempty"`
	Keepers     []util.EthID `cbor:"kp,omitempty" json:"keepers,omitempty"`
	Approver    *util.EthID  `cbor:"ap,omitempty" json:"approver,omitempty"`
	ApproveList []TxType     `cbor:"apl,omitempty" json:"approveList,omitempty"`
	Amount      *big.Int     `cbor:"a,omitempty" json:"amount,omitempty"`
	Name        string       `cbor:"n,omitempty" json:"name,omitempty"`
	Data        RawData      `cbor:"d,omitempty" json:"data,omitempty"`

	// external assignment fields
	raw []byte `cbor:"-" json:"-"`
}

// SyntacticVerify verifies that a *TxAccounter is well-formed.
func (t *TxAccounter) SyntacticVerify() error {
	if t == nil {
		return fmt.Errorf("TxAccounter.SyntacticVerify failed: nil pointer")
	}
	if t.Name != "" && !util.ValidName(t.Name) {
		return fmt.Errorf("TxAccounter.SyntacticVerify failed: invalid name %s", strconv.Quote(t.Name))
	}

	if t.Amount != nil && t.Amount.Sign() < 1 {
		return fmt.Errorf("TxAccounter.SyntacticVerify failed: invalid amount")
	}
	if len(t.Keepers) > math.MaxUint8 {
		return fmt.Errorf("TxAccounter.SyntacticVerify failed: too many keepers")
	}
	if int(t.Threshold) > len(t.Keepers) {
		return fmt.Errorf("TxAccounter.SyntacticVerify failed: invalid threshold")
	}
	for _, id := range t.Keepers {
		if id == util.EthIDEmpty {
			return fmt.Errorf("TxAccounter.SyntacticVerify failed: invalid keeper")
		}
	}
	if t.ApproveList != nil {
		for _, ty := range t.ApproveList {
			if ty > TypeDeleteData {
				return fmt.Errorf("TxAccounter.SyntacticVerify failed: invalid TxType %d in approveList", ty)
			}
		}
	}
	var err error
	if t.raw, err = t.Marshal(); err != nil {
		return fmt.Errorf("TxAccounter.SyntacticVerify marshal error: %v", err)
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
	return DecMode.Unmarshal(data, t)
}

func (t *TxAccounter) Marshal() ([]byte, error) {
	return EncMode.Marshal(t)
}
