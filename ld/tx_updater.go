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
	"github.com/ldclabs/ldvm/util"
)

// TxUpdater is a hybrid data model for:
//
// TxDeleteData{ID, Version[, Data]}
// TxUpdateDataKeepers{ID, Version, Threshold, Keepers[, Data]}
// TxUpdateData{ID, Version, Data[, Expire]}
// TxAccountUpdateKeepers{Threshold, Keepers[, Data]}
// TxUpdateModelKeepers{ID, Threshold, Keepers[, Data]}
// TxUpdateDataKeepersByAuth{ID, Version, To, Amount[, Token, Expire, Threshold, Keepers, Data]}
type TxUpdater struct {
	ID        ids.ShortID // data id
	Version   uint64      // data version
	Threshold uint8
	Keepers   []ids.ShortID
	Token     util.TokenSymbol // token symbol, default is NativeToken
	To        ids.ShortID      // optional recipient
	Amount    *big.Int         // transfer amount
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
	Token     string          `json:"token,omitempty"`
	To        string          `json:"to,omitempty"`
	Amount    *big.Int        `json:"amount,omitempty"`
	Expire    uint64          `json:"expire,omitempty"`
	Data      json.RawMessage `json:"data,omitempty"`
}

func (t *TxUpdater) MarshalJSON() ([]byte, error) {
	if t == nil {
		return util.Null, nil
	}
	v := &jsonTxUpdater{
		Version:   t.Version,
		Threshold: t.Threshold,
		Amount:    t.Amount,
		Expire:    t.Expire,
		Token:     t.Token.String(),
		Data:      util.JSONMarshalData(t.Data),
	}

	if t.ID != ids.ShortEmpty {
		v.ID = util.EthID(t.ID).String()
	}
	if len(t.Keepers) > 0 {
		v.Keepers = make([]string, len(t.Keepers))
		for i := range t.Keepers {
			v.Keepers[i] = util.EthID(t.Keepers[i]).String()
		}
	}
	if t.To != ids.ShortEmpty {
		v.To = util.EthID(t.To).String()
	}
	return json.Marshal(v)
}

func (t *TxUpdater) Copy() *TxUpdater {
	x := new(TxUpdater)
	*x = *t
	if t.Amount != nil {
		x.Amount = new(big.Int).Set(t.Amount)
	}
	x.Keepers = make([]ids.ShortID, len(t.Keepers))
	copy(x.Keepers, t.Keepers)
	x.Data = make([]byte, len(t.Data))
	copy(x.Data, t.Data)
	x.raw = nil
	return x
}

// SyntacticVerify verifies that a *TxUpdater is well-formed.
func (t *TxUpdater) SyntacticVerify() error {
	if t == nil {
		return fmt.Errorf("invalid TxUpdater")
	}

	if t.Token != util.NativeToken && t.Token.String() == "" {
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
	if t.Expire > 0 && t.Expire < uint64(time.Now().Unix()) {
		return fmt.Errorf("invalid expire")
	}
	for _, id := range t.Keepers {
		if id == ids.ShortEmpty {
			return fmt.Errorf("invalid data keeper")
		}
	}
	if _, err := t.Marshal(); err != nil {
		return fmt.Errorf("TxUpdater marshal error: %v", err)
	}
	return nil
}

func (t *TxUpdater) Equal(o *TxUpdater) bool {
	if o == nil {
		return false
	}
	if len(o.raw) > 0 && len(t.raw) > 0 {
		return bytes.Equal(o.raw, t.raw)
	}
	if o.ID != t.ID {
		return false
	}
	if o.Version != t.Version {
		return false
	}
	if o.Threshold != t.Threshold {
		return false
	}
	if o.Amount == nil || t.Amount == nil {
		if o.Amount != t.Amount {
			return false
		}
	}
	if o.Amount.Cmp(t.Amount) != 0 {
		return false
	}
	if o.To != t.To {
		return false
	}
	if o.Token != t.Token {
		return false
	}
	if len(o.Keepers) != len(t.Keepers) {
		return false
	}
	for i := range t.Keepers {
		if o.Keepers[i] != t.Keepers[i] {
			return false
		}
	}
	if o.Expire != t.Expire {
		return false
	}
	return bytes.Equal(o.Data, t.Data)
}

func (t *TxUpdater) Bytes() []byte {
	if len(t.raw) == 0 {
		if _, err := t.Marshal(); err != nil {
			panic(err)
		}
	}

	return t.raw
}

func (t *TxUpdater) Unmarshal(data []byte) error {
	p, err := txUpdaterLDBuilder.Unmarshal(data)
	if err != nil {
		return err
	}
	if v, ok := p.(*bindTxUpdater); ok {
		t.Version = v.Version.Value()
		t.Threshold = v.Threshold.Value()
		t.Expire = v.Expire.Value()
		t.Amount = PtrToBigInt(v.Amount)
		t.Data = PtrToBytes(v.Data)
		if t.ID, err = PtrToShortID(v.ID); err != nil {
			return fmt.Errorf("unmarshal error: %v", err)
		}
		var token ids.ShortID
		if token, err = PtrToShortID(v.Token); err != nil {
			return fmt.Errorf("unmarshal error: %v", err)
		}
		t.Token = util.TokenSymbol(token)
		if t.Keepers, err = PtrToShortIDs(v.Keepers); err != nil {
			return fmt.Errorf("unmarshal error: %v", err)
		}
		if t.To, err = PtrToShortID(v.To); err != nil {
			return fmt.Errorf("unmarshal error: %v", err)
		}
		t.raw = data
		return nil
	}
	return fmt.Errorf("unmarshal error: expected *bindTxUpdater")
}

func (t *TxUpdater) Marshal() ([]byte, error) {
	v := &bindTxUpdater{
		ID:        PtrFromShortID(t.ID),
		Version:   PtrFromUint64(t.Version),
		Threshold: PtrFromUint8(t.Threshold),
		Keepers:   PtrFromShortIDs(t.Keepers),
		Token:     PtrFromShortID(ids.ShortID(t.Token)),
		Amount:    PtrFromBigInt(t.Amount),
		To:        PtrFromShortID(t.To),
		Expire:    PtrFromUint64(t.Expire),
		Data:      PtrFromBytes(t.Data),
	}
	data, err := txUpdaterLDBuilder.Marshal(v)
	if err != nil {
		return nil, err
	}
	t.raw = data
	return data, nil
}

type bindTxUpdater struct {
	ID        *[]byte
	Version   *Uint64
	Threshold *Uint8
	Keepers   *[][]byte
	Token     *[]byte
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
		Token     nullable ID20   (rename "tk")
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
