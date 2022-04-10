// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package ld

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/big"
	"time"

	"github.com/ava-labs/avalanchego/ids"
)

// TxTransfer is a hybrid data model for:
//
// TxTransferReply{To[, Amount, Expire, Data]}
// TxTransferCash{Nonce, From, Amount, Expire[, To, Data]}
type TxTransfer struct {
	Nonce  uint64      // sender's nonce
	From   ids.ShortID // amount sender
	To     ids.ShortID // amount recipient
	Amount *big.Int    // transfer amount
	Expire uint64
	Data   []byte
	raw    []byte
}

type jsonTxTransfer struct {
	Nonce  uint64          `json:"nonce,omitempty"`
	From   string          `json:"from,omitempty"`
	To     string          `json:"to,omitempty"`
	Amount *big.Int        `json:"amount,omitempty"`
	Expire uint64          `json:"expire,omitempty"`
	Data   json.RawMessage `json:"data,omitempty"`
}

func (d *TxTransfer) MarshalJSON() ([]byte, error) {
	if d == nil {
		return Null, nil
	}
	v := &jsonTxTransfer{
		Nonce:  d.Nonce,
		Amount: d.Amount,
		Expire: d.Expire,
		Data:   JsonMarshalData(d.Data),
	}

	if d.From != ids.ShortEmpty {
		v.From = EthID(d.From).String()
	}
	if d.To != ids.ShortEmpty {
		v.To = EthID(d.To).String()
	}
	return json.Marshal(v)
}

func (d *TxTransfer) Copy() *TxTransfer {
	x := new(TxTransfer)
	*x = *d
	if d.Amount != nil {
		x.Amount = new(big.Int).Set(d.Amount)
	}
	x.Data = make([]byte, len(d.Data))
	copy(x.Data, d.Data)
	x.raw = make([]byte, len(d.raw))
	copy(x.raw, d.raw)
	return x
}

// SyntacticVerify verifies that a *DataMeta is well-formed.
func (d *TxTransfer) SyntacticVerify() error {
	if d.Nonce == 0 {
		return fmt.Errorf("invalid transaction nonce")
	}
	if d.From == ids.ShortEmpty {
		return fmt.Errorf("invalid transaction from")
	}
	if d.Amount != nil && d.Amount.Sign() < 1 {
		return fmt.Errorf("invalid transaction amount")
	}
	if d.Expire < uint64(time.Now().Unix()) {
		return fmt.Errorf("invalid expire")
	}
	if _, err := d.Marshal(); err != nil {
		return fmt.Errorf("TxTransfer marshal error: %v", err)
	}
	return nil
}

func (d *TxTransfer) Equal(o *TxTransfer) bool {
	if o == nil {
		return false
	}
	if len(o.raw) > 0 && len(d.raw) > 0 {
		return bytes.Equal(o.raw, d.raw)
	}
	if o.Nonce != d.Nonce {
		return false
	}
	if o.From != d.From {
		return false
	}
	if o.To != d.To {
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
	if o.Expire != d.Expire {
		return false
	}
	return bytes.Equal(o.Data, d.Data)
}

func (d *TxTransfer) Bytes() []byte {
	if len(d.raw) == 0 {
		if _, err := d.Marshal(); err != nil {
			panic(err)
		}
	}

	return d.raw
}

func (d *TxTransfer) Unmarshal(data []byte) error {
	p, err := txTransferLDBuilder.Unmarshal(data)
	if err != nil {
		return err
	}
	if v, ok := p.(*bindTxTransfer); ok {
		d.Nonce = v.Nonce.Value()
		d.Expire = v.Expire.Value()
		d.Amount = PtrToBigInt(v.Amount)
		d.Data = PtrToBytes(v.Data)
		if d.From, err = PtrToShortID(v.From); err != nil {
			return fmt.Errorf("unmarshal error: %v", err)
		}
		if d.To, err = PtrToShortID(v.To); err != nil {
			return fmt.Errorf("unmarshal error: %v", err)
		}
		d.raw = data
		return nil
	}
	return fmt.Errorf("unmarshal error: expected *bindTxTransfer")
}

func (d *TxTransfer) Marshal() ([]byte, error) {
	v := &bindTxTransfer{
		Nonce:  PtrFromUint64(d.Nonce),
		To:     PtrFromShortID(d.To),
		From:   PtrFromShortID(d.From),
		Amount: PtrFromBigInt(d.Amount),
		Expire: PtrFromUint64(d.Expire),
		Data:   PtrFromBytes(d.Data),
	}
	data, err := txTransferLDBuilder.Marshal(v)
	if err != nil {
		return nil, err
	}
	d.raw = data
	return data, nil
}

type bindTxTransfer struct {
	Nonce  *Uint64
	From   *[]byte
	To     *[]byte
	Amount *[]byte
	Expire *Uint64
	Data   *[]byte
}

var txTransferLDBuilder *LDBuilder

func init() {
	sch := `
	type Uint8 bytes
	type Uint64 bytes
	type ID20 bytes
	type BigInt bytes
	type TxTransfer struct {
		Nonce  nullable Uint64 (rename "n")
		From   nullable ID20   (rename "fr")
		To     nullable ID20   (rename "to")
		Amount nullable BigInt (rename "a")
		Expire nullable Uint64 (rename "e")
		Data   nullable Bytes  (rename "d")
	}
`

	builder, err := NewLDBuilder("TxTransfer", []byte(sch), (*bindTxTransfer)(nil))
	if err != nil {
		panic(err)
	}
	txTransferLDBuilder = builder
}
