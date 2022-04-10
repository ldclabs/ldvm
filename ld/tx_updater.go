// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package ld

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"math/big"
	"time"

	"github.com/ava-labs/avalanchego/ids"
)

// TxUpdater is a hybrid data model for:
//
// TxDeleteData{ID, Version[, Data]}
// TxUpdateDataKeepers{ID, Version, Threshold, Keepers[, Data]}
// TxUpdateData{ID, Version, Data[, Expire]}
// TxUpdateAccountKeepers{Threshold, Keepers[, Data]}
// TxUpdateModelKeepers{ID, Threshold, Keepers[, Data]}
// TxUpdateDataKeepersByAuth{ID, Version, To, Amount[, Expire, Threshold, Keepers, Data]}
type TxUpdater struct {
	ID        ids.ShortID // data id
	Version   uint64      // data version
	Threshold uint8
	Keepers   []ids.ShortID
	To        ids.ShortID // amount recipient
	Amount    *big.Int    // transfer amount
	Expire    uint64
	Data      []byte

	// external assignment
	raw []byte
}

type jsonTxUpdater struct {
	ID        string          `json:"id,omitempty"`
	Version   uint64          `json:"version,omitempty"`
	Threshold uint8           `json:"threshold,omitempty"`
	Keepers   []string        `json:"keepers,omitempty"`
	To        string          `json:"to,omitempty"`
	Amount    *big.Int        `json:"amount,omitempty"`
	Expire    uint64          `json:"expire,omitempty"`
	Data      json.RawMessage `json:"data,omitempty"`
}

func (d *TxUpdater) MarshalJSON() ([]byte, error) {
	if d == nil {
		return Null, nil
	}
	v := &jsonTxUpdater{
		Version:   d.Version,
		Threshold: d.Threshold,
		Amount:    d.Amount,
		Expire:    d.Expire,
		Data:      JsonMarshalData(d.Data),
	}

	if d.ID != ids.ShortEmpty {
		v.ID = EthID(d.ID).String()
	}
	if len(d.Keepers) > 0 {
		v.Keepers = make([]string, len(d.Keepers))
		for i := range d.Keepers {
			v.Keepers[i] = EthID(d.Keepers[i]).String()
		}
	}
	if d.To != ids.ShortEmpty {
		v.To = EthID(d.To).String()
	}
	return json.Marshal(v)
}

func (d *TxUpdater) Copy() *TxUpdater {
	x := new(TxUpdater)
	*x = *d
	if d.Amount != nil {
		x.Amount = new(big.Int).Set(d.Amount)
	}
	x.Keepers = make([]ids.ShortID, len(d.Keepers))
	copy(x.Keepers, d.Keepers)
	x.Data = make([]byte, len(d.Data))
	copy(x.Data, d.Data)
	x.raw = make([]byte, len(d.raw))
	copy(x.raw, d.raw)
	return x
}

// SyntacticVerify verifies that a *TxUpdater is well-formed.
func (d *TxUpdater) SyntacticVerify() error {
	if d == nil {
		return fmt.Errorf("invalid TxUpdater")
	}

	if d.Amount != nil && d.Amount.Sign() < 0 {
		return fmt.Errorf("invalid amount")
	}
	if len(d.Keepers) > math.MaxUint8 {
		return fmt.Errorf("invalid keepers, too many")
	}
	if int(d.Threshold) > len(d.Keepers) {
		return fmt.Errorf("invalid threshold")
	}
	if d.Expire > 0 && d.Expire < uint64(time.Now().Unix()) {
		return fmt.Errorf("invalid expire")
	}
	for _, id := range d.Keepers {
		if id == ids.ShortEmpty {
			return fmt.Errorf("invalid data keeper")
		}
	}
	if _, err := d.Marshal(); err != nil {
		return fmt.Errorf("TxUpdater marshal error: %v", err)
	}
	return nil
}

func (d *TxUpdater) Equal(o *TxUpdater) bool {
	if o == nil {
		return false
	}
	if len(o.raw) > 0 && len(d.raw) > 0 {
		return bytes.Equal(o.raw, d.raw)
	}
	if o.ID != d.ID {
		return false
	}
	if o.Version != d.Version {
		return false
	}
	if o.Threshold != d.Threshold {
		return false
	}
	if o.Amount == nil || d.Amount == nil {
		if o.Amount != d.Amount {
			return false
		}
	}
	if o.Amount.Cmp(d.Amount) != 0 {
		return false
	}
	if o.To != d.To {
		return false
	}
	if len(o.Keepers) != len(d.Keepers) {
		return false
	}
	for i := range d.Keepers {
		if o.Keepers[i] != d.Keepers[i] {
			return false
		}
	}
	if o.Expire != d.Expire {
		return false
	}
	return bytes.Equal(o.Data, d.Data)
}

func (d *TxUpdater) Bytes() []byte {
	if len(d.raw) == 0 {
		if _, err := d.Marshal(); err != nil {
			panic(err)
		}
	}

	return d.raw
}

func (d *TxUpdater) Unmarshal(data []byte) error {
	p, err := txUpdaterLDBuilder.Unmarshal(data)
	if err != nil {
		return err
	}
	if v, ok := p.(*bindTxUpdater); ok {
		d.Version = v.Version.Value()
		d.Threshold = v.Threshold.Value()
		d.Expire = v.Expire.Value()
		d.Amount = PtrToBigInt(v.Amount)
		d.Data = PtrToBytes(v.Data)
		if d.ID, err = PtrToShortID(v.ID); err != nil {
			return fmt.Errorf("unmarshal error: %v", err)
		}
		if d.Keepers, err = PtrToShortIDs(v.Keepers); err != nil {
			return fmt.Errorf("unmarshal error: %v", err)
		}
		if d.To, err = PtrToShortID(v.To); err != nil {
			return fmt.Errorf("unmarshal error: %v", err)
		}
		d.raw = data
		return nil
	}
	return fmt.Errorf("unmarshal error: expected *bindTxUpdater")
}

func (d *TxUpdater) Marshal() ([]byte, error) {
	v := &bindTxUpdater{
		ID:        PtrFromShortID(d.ID),
		Version:   PtrFromUint64(d.Version),
		Threshold: PtrFromUint8(d.Threshold),
		Keepers:   PtrFromShortIDs(d.Keepers),
		Amount:    PtrFromBigInt(d.Amount),
		To:        PtrFromShortID(d.To),
		Expire:    PtrFromUint64(d.Expire),
		Data:      PtrFromBytes(d.Data),
	}
	data, err := txUpdaterLDBuilder.Marshal(v)
	if err != nil {
		return nil, err
	}
	d.raw = data
	return data, nil
}

type bindTxUpdater struct {
	ID        *[]byte
	Version   *Uint64
	Threshold *Uint8
	Keepers   *[][]byte
	To        *[]byte
	Amount    *[]byte
	Expire    *Uint64
	Data      *[]byte
}

var txUpdaterLDBuilder *LDBuilder

func init() {
	sch := `
	type Uint8 bytes
	type Uint64 bytes
	type ID20 bytes
	type BigInt bytes
	type TxUpdater struct {
		ID        nullable ID20   (rename "id")
		Version   nullable Uint64 (rename "v")
		Threshold nullable Uint8  (rename "th")
		Keepers   nullable [ID20] (rename "ks")
		To        nullable ID20   (rename "to")
		Amount    nullable BigInt (rename "a")
		Expire    nullable Uint64 (rename "e")
		Data      nullable Bytes  (rename "d")
	}
`

	builder, err := NewLDBuilder("TxUpdater", []byte(sch), (*bindTxUpdater)(nil))
	if err != nil {
		panic(err)
	}
	txUpdaterLDBuilder = builder
}
