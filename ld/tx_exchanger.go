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
	"github.com/ldclabs/ldvm/constants"
	"github.com/ldclabs/ldvm/util"
)

// TxExchanger
type TxExchanger struct {
	Nonce   uint64      // saler' account nonce
	Sell    ids.ShortID // token symbol to sell
	Receive ids.ShortID // token symbol to receive
	Quota   *big.Int    // token sales quota per a tx
	Minimum *big.Int    // minimum amount to buy
	Price   *big.Int    // receive token amount = Quota * Price
	Expire  uint64
	Seller  ids.ShortID
	To      ids.ShortID // optional designated purchaser

	// external assignment
	raw []byte
}

type jsonTxExchanger struct {
	Nonce   uint64   `json:"nonce"`
	Sell    string   `json:"sell"`
	Receive string   `json:"receive"`
	Quota   *big.Int `json:"quota"`
	Minimum *big.Int `json:"Minimum"`
	Price   *big.Int `json:"Price"`
	Expire  uint64   `json:"expire"`
	Seller  string   `json:"payee"`
	To      string   `json:"to,omitempty"`
}

func (t *TxExchanger) MarshalJSON() ([]byte, error) {
	if t == nil {
		return util.Null, nil
	}
	v := &jsonTxExchanger{
		Nonce:   t.Nonce,
		Sell:    util.TokenSymbol(t.Sell).String(),
		Receive: util.TokenSymbol(t.Receive).String(),
		Quota:   t.Quota,
		Minimum: t.Minimum,
		Price:   t.Price,
		Expire:  t.Expire,
		Seller:  util.EthID(t.Seller).String(),
	}
	if t.To != ids.ShortEmpty {
		v.To = util.EthID(t.To).String()
	}
	return json.Marshal(v)
}

func (t *TxExchanger) Copy() *TxExchanger {
	x := new(TxExchanger)
	*x = *t
	x.Quota = new(big.Int).Set(t.Quota)
	x.Minimum = new(big.Int).Set(t.Minimum)
	x.Price = new(big.Int).Set(t.Price)
	x.raw = nil
	return x
}

// SyntacticVerify verifies that a *TxExchanger is well-formed.
func (t *TxExchanger) SyntacticVerify() error {
	if t == nil {
		return fmt.Errorf("invalid TxExchanger")
	}

	if t.Nonce == 0 {
		return fmt.Errorf("invalid nonce")
	}
	if t.Sell != constants.LDCAccount && util.TokenSymbol(t.Sell).String() == "" {
		return fmt.Errorf("invalid token symbol to sell")
	}
	if t.Receive != constants.LDCAccount && util.TokenSymbol(t.Receive).String() == "" {
		return fmt.Errorf("invalid token symbol to receive")
	}
	if t.Sell == t.Receive {
		return fmt.Errorf("invalid token symbol to receive")
	}
	if t.Quota == nil || t.Quota.Sign() < 1 {
		return fmt.Errorf("invalid quota")
	}
	if t.Minimum == nil || t.Minimum.Sign() < 1 {
		return fmt.Errorf("invalid minimum")
	}
	if t.Price == nil || t.Price.Sign() < 1 {
		return fmt.Errorf("invalid price")
	}
	if t.Expire < uint64(time.Now().Unix()) {
		return fmt.Errorf("invalid expire")
	}
	if t.Seller == ids.ShortEmpty {
		return fmt.Errorf("invalid payee")
	}
	if _, err := t.Marshal(); err != nil {
		return fmt.Errorf("TxExchanger marshal error: %v", err)
	}
	return nil
}

func (t *TxExchanger) Equal(o *TxExchanger) bool {
	if o == nil {
		return false
	}
	if len(o.raw) > 0 && len(t.raw) > 0 {
		return bytes.Equal(o.raw, t.raw)
	}
	if o.Nonce != t.Nonce {
		return false
	}
	if o.Sell != t.Sell {
		return false
	}
	if o.Receive != t.Receive {
		return false
	}
	if o.Quota == nil || t.Quota == nil || o.Quota.Cmp(t.Quota) != 0 {
		return false
	}
	if o.Minimum == nil || t.Minimum == nil || o.Minimum.Cmp(t.Quota) != 0 {
		return false
	}
	if o.Price == nil || t.Price == nil || o.Price.Cmp(t.Quota) != 0 {
		return false
	}
	if o.Expire != t.Expire {
		return false
	}
	if o.Seller != t.Seller {
		return false
	}
	if o.To != t.To {
		return false
	}
	return true
}

func (t *TxExchanger) Bytes() []byte {
	if len(t.raw) == 0 {
		if _, err := t.Marshal(); err != nil {
			panic(err)
		}
	}

	return t.raw
}

func (t *TxExchanger) Unmarshal(data []byte) error {
	p, err := txExchangerLDBuilder.Unmarshal(data)
	if err != nil {
		return err
	}
	if v, ok := p.(*bindTxExchanger); ok {
		if !v.Nonce.Valid() ||
			!v.Expire.Valid() {
			return fmt.Errorf("unmarshal error: invalid uint64")
		}

		t.Nonce = v.Nonce.Value()
		t.Expire = v.Expire.Value()
		t.Quota = v.Quota.Value()
		t.Minimum = v.Minimum.Value()
		t.Price = v.Price.Value()
		if t.Sell, err = ToShortID(v.Sell); err != nil {
			return fmt.Errorf("unmarshal error: %v", err)
		}
		if t.Receive, err = ToShortID(v.Receive); err != nil {
			return fmt.Errorf("unmarshal error: %v", err)
		}
		if t.Seller, err = ToShortID(v.Seller); err != nil {
			return fmt.Errorf("unmarshal error: %v", err)
		}
		if t.To, err = PtrToShortID(v.To); err != nil {
			return fmt.Errorf("unmarshal error: %v", err)
		}
		t.raw = data
		return nil
	}
	return fmt.Errorf("unmarshal error: expected *bindTxExchanger")
}

func (t *TxExchanger) Marshal() ([]byte, error) {
	v := &bindTxExchanger{
		Nonce:   FromUint64(t.Nonce),
		Sell:    FromShortID(t.Sell),
		Receive: FromShortID(t.Receive),
		Quota:   FromUint(t.Quota),
		Minimum: FromUint(t.Minimum),
		Price:   FromUint(t.Price),
		Expire:  FromUint64(t.Expire),
		Seller:  FromShortID(t.Seller),
		To:      PtrFromShortID(t.To),
	}
	data, err := txExchangerLDBuilder.Marshal(v)
	if err != nil {
		return nil, err
	}
	t.raw = data
	return data, nil
}

type bindTxExchanger struct {
	Nonce   Uint64
	Sell    []byte
	Receive []byte
	Quota   BigUint
	Minimum BigUint
	Price   BigUint
	Expire  Uint64
	Seller  []byte
	To      *[]byte
}

var txExchangerLDBuilder *LDBuilder

func init() {
	sch := `
	type Uint8 bytes
	type Uint64 bytes
	type ID20 bytes
	type BigUint bytes
	type TxExchanger struct {
		Nonce   Uint64        (rename "n")
		Sell    ID20          (rename "st")
		Receive ID20          (rename "rt")
		Quota   BigUint       (rename "q")
		Minimum BigUint       (rename "m")
		Price   BigUint       (rename "p")
		Expire  Uint64        (rename "e")
		Seller  ID20          (rename "sl")
		To      nullable ID20 (rename "to")
	}
`

	builder, err := NewLDBuilder("TxExchanger", []byte(sch), (*bindTxExchanger)(nil))
	if err != nil {
		panic(err)
	}
	txExchangerLDBuilder = builder
}
