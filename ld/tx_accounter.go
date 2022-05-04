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
	Threshold uint8        `cbor:"th,omitempty" json:"threshold,omitempty"`
	Keepers   []util.EthID `cbor:"kp,omitempty" json:"keepers,omitempty"`
	Amount    *big.Int     `cbor:"a,omitempty" json:"amount,omitempty"`
	Name      string       `cbor:"n,omitempty" json:"name,omitempty"`
	Message   string       `cbor:"m,omitempty" json:"message,omitempty"`
	Data      RawData      `cbor:"d,omitempty" json:"data,omitempty"`
}

// SyntacticVerify verifies that a *TxAccounter is well-formed.
func (t *TxAccounter) SyntacticVerify() error {
	if t == nil {
		return fmt.Errorf("invalid TxAccounter")
	}
	if t.Name != "" && !util.ValidName(t.Name) {
		return fmt.Errorf("invalid name string %s", strconv.Quote(t.Name))
	}
	if t.Message != "" && !util.ValidMessage(t.Message) {
		return fmt.Errorf("invalid message string %s", strconv.Quote(t.Message))
	}

	if t.Amount == nil || t.Amount.Sign() < 0 {
		return fmt.Errorf("invalid amount")
	}
	if len(t.Keepers) > math.MaxUint8 {
		return fmt.Errorf("invalid keepers, too many")
	}
	if int(t.Threshold) > len(t.Keepers) {
		return fmt.Errorf("invalid threshold")
	}
	for _, id := range t.Keepers {
		if id == util.EthIDEmpty {
			return fmt.Errorf("invalid keeper")
		}
	}
	if _, err := t.Marshal(); err != nil {
		return fmt.Errorf("TxAccounter marshal error: %v", err)
	}
	return nil
}

func (t *TxAccounter) Unmarshal(data []byte) error {
	return DecMode.Unmarshal(data, t)
}

func (t *TxAccounter) Marshal() ([]byte, error) {
	data, err := EncMode.Marshal(t)
	if err != nil {
		return nil, err
	}
	return data, nil
}
