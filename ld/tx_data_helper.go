// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package ld

import (
	"math/big"

	"github.com/ldclabs/ldvm/util"
)

// MaxSendOutputs is the maximum number of SendOutput that can be included in TransferMultiple tx.
// Recommend 2 ~ 100 for gas cost performance.
const MaxSendOutputs = 1024

// SendOutput specifies that [Amount] of token be sent to [To]
type SendOutput struct {
	To     util.Address `cbor:"to" json:"to"` // Address of the recipient
	Amount *big.Int     `cbor:"a" json:"amount"`
}

type SendOutputs []SendOutput

func (so SendOutputs) SyntacticVerify() error {
	errp := util.ErrPrefix("ld.SendOutputs.SyntacticVerify: ")

	if len(so) == 0 {
		return errp.Errorf("empty SendOutputs")
	}

	if len(so) > MaxSendOutputs {
		return errp.Errorf("too many SendOutputs")
	}

	set := make(map[util.Address]struct{}, len(so))

	for i, o := range so {
		switch {
		case o.To == util.AddressEmpty:
			return errp.Errorf("invalid to address at %d", i)

		case o.Amount == nil || o.Amount.Sign() <= 0:
			return errp.Errorf("invalid amount at %d", i)
		}

		if _, ok := set[o.To]; ok {
			return errp.Errorf("duplicate to address %s at %d", o.To.String(), i)
		}
		set[o.To] = struct{}{}
	}

	return nil
}

func (so *SendOutputs) Unmarshal(data []byte) error {
	return util.ErrPrefix("ld.SendOutputs.Unmarshal: ").
		ErrorIf(util.UnmarshalCBOR(data, so))
}

func (so SendOutputs) Marshal() ([]byte, error) {
	return util.ErrPrefix("ld.SendOutputs.Marshal: ").
		ErrorMap(util.MarshalCBOR(so))
}
