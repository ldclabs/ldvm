// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package ld

import (
	"fmt"
	"math"
	"math/big"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ldclabs/ldvm/constants"
	"github.com/ldclabs/ldvm/util"
)

// TxUpdater is a hybrid data model for:
//
// TxCreateData{ID, Version, Threshold, Keepers, Data, KSig} no model keepers
// TxCreateData{ID, Version, To, Amount, Threshold, Keepers, Data, KSig, SSig, Expire} with model keepers
// TxUpdateData{ID, Version, Data, KSig} no model keepers
// TxUpdateData{ID, Version, To, Amount, Data, KSig, SSig, Expire} with model keepers
// TxDeleteData{ID, Version[, Data]}
// TxUpdateDataKeepers{ID, Version, Threshold, Keepers, KSig[, Data]}
// TxUpdateDataKeepersByAuth{ID, Version, To, Amount, Threshold, Keepers, KSig, Expire[, Token, Data]}
// TxUpdateModelKeepers{ID, Threshold, Keepers[, Data]}
type TxUpdater struct {
	ID        ids.ShortID      `cbor:"id,omitempty" json:"id,omitempty"`     // data id or model id
	Version   uint64           `cbor:"v,omitempty" json:"version,omitempty"` // data version
	Threshold uint8            `cbor:"th,omitempty" json:"threshold,omitempty"`
	Keepers   []util.EthID     `cbor:"kp,omitempty" json:"keepers,omitempty"`
	Token     util.TokenSymbol `cbor:"tk,omitempty" json:"token,omitempty"` // token symbol, default is NativeToken
	To        util.EthID       `cbor:"to,omitempty" json:"to,omitempty"`    // optional recipient
	Amount    *big.Int         `cbor:"a,omitempty" json:"amount,omitempty"` // transfer amount
	KSig      *util.Signature  `cbor:"ks,omitempty" json:"kSig,omitempty"`  // full data signature signing by Data Keeper
	SSig      *util.Signature  `cbor:"ss,omitempty" json:"sSig,omitempty"`  // full data signature signing by Service Authority
	Expire    uint64           `cbor:"e,omitempty" json:"expire,omitempty"`
	Data      RawData          `cbor:"d,omitempty" json:"data,omitempty"`
}

// SyntacticVerify verifies that a *TxUpdater is well-formed.
func (t *TxUpdater) SyntacticVerify() error {
	if t == nil {
		return fmt.Errorf("invalid TxUpdater")
	}

	if t.Token != constants.NativeToken && t.Token.String() == "" {
		return fmt.Errorf("invalid token symbol")
	}
	if t.Amount != nil && t.Amount.Sign() < 0 {
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
			return fmt.Errorf("invalid data keeper")
		}
	}
	if _, err := t.Marshal(); err != nil {
		return fmt.Errorf("TxUpdater marshal error: %v", err)
	}
	return nil
}

func (t *TxUpdater) Unmarshal(data []byte) error {
	return DecMode.Unmarshal(data, t)
}

func (t *TxUpdater) Marshal() ([]byte, error) {
	data, err := EncMode.Marshal(t)
	if err != nil {
		return nil, err
	}
	return data, nil
}
