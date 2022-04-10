// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package ld

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"math/big"

	"github.com/ava-labs/avalanchego/ids"
)

const (
	// 0: mint tx fee, issued by validators
	TypeMintFee TxType = iota
	// 1: send given amount of NanoLDC to a address
	TypeTransfer
	// 2: send given amount of NanoLDC to the address who request payment
	TypeTransferReply
	// 3: cash given amount of NanoLDC to sender, like cash a check.
	TypeTransferCash
	// 4. update account's Keepers and Threshold
	TypeUpdateAccountKeepers
	// 5. create a data model
	TypeCreateModel
	// 6. update data model's Keepers and Threshold
	TypeUpdateModelKeepers
	// 7. create a data from the model
	TypeCreateData
	// 8. update the data's Data
	TypeUpdateData
	// 9. update data's Keepers and Threshold
	TypeUpdateDataKeepers
	// 10. update data's Keepers and Threshold by authorization
	TypeUpdateDataKeepersByAuth
	// 11. delete the data
	TypeDeleteData
)

const (
	// the meaning of life, the universe, and everything
	TxTransferReply              = uint64(42)
	TxTransferGas                = uint64(100)
	TxTransferCash               = uint64(100)
	TxUpdateDataGas              = uint64(200)
	TxDeleteDataGas              = uint64(200)
	TxCreateDataGas              = uint64(500)
	TxCreateModelGas             = uint64(500)
	TxUpdateAccountKeepersGas    = uint64(1000)
	TxUpdateDataKeepersGas       = uint64(1000)
	TxUpdateModelKeepersGas      = uint64(1000)
	TxUpdateDataKeepersByAuthGas = uint64(1000)
	MinThresholdGas              = uint64(1000)
)

// TxType is an uint8 representing the type of the tx
type TxType uint8

func TxTypeString(t TxType) string {
	switch t {
	case TypeMintFee:
		return "TypeMintFee"
	case TypeTransfer:
		return "TypeTransfer"
	case TypeTransferReply:
		return "TypeTransferReply"
	case TypeTransferCash:
		return "TypeTransferCash"
	case TypeUpdateAccountKeepers:
		return "TypeUpdateAccountKeepers"
	case TypeCreateModel:
		return "TypeCreateModel"
	case TypeUpdateModelKeepers:
		return "TypeUpdateModelKeepers"
	case TypeCreateData:
		return "TypeCreateData"
	case TypeUpdateData:
		return "TypeUpdateData"
	case TypeUpdateDataKeepers:
		return "TypeUpdateDataKeepers"
	case TypeUpdateDataKeepersByAuth:
		return "TypeUpdateDataKeepersByAuth"
	case TypeDeleteData:
		return "TypeDeleteData"
	default:
		return "TypeUnknown"
	}
}

type Transaction struct {
	Type         TxType
	ChainID      uint64
	Nonce        uint64
	Gas          uint64 // calculate when build block.
	GasTip       uint64
	GasFeeCap    uint64
	From         ids.ShortID
	To           ids.ShortID
	Amount       *big.Int
	Data         []byte
	Signatures   []Signature
	ExSignatures []Signature
	gas          uint64
	id           ids.ID
	unsignedRaw  []byte // raw bytes for sign
	raw          []byte // the transaction's raw bytes, included id and sigs.
}

type jsonTransaction struct {
	ID           ids.ID          `json:"id"`
	Type         string          `json:"type"`
	ChainID      uint64          `json:"chainID"`
	Nonce        uint64          `json:"Nonce"`
	Gas          uint64          `json:"gas"` // calculate when build block.
	GasTip       uint64          `json:"gasTip"`
	GasFeeCap    uint64          `json:"gasFeeCap"`
	From         string          `json:"from"`
	To           string          `json:"to"`
	Amount       *big.Int        `json:"amount"`
	Data         json.RawMessage `json:"data"`
	Signatures   []Signature     `json:"signatures"`
	ExSignatures []Signature     `json:"exSignatures"`
}

func (t *Transaction) MarshalJSON() ([]byte, error) {
	if t == nil {
		return Null, nil
	}
	v := &jsonTransaction{
		ID:           t.ID(),
		Type:         TxTypeString(t.Type),
		ChainID:      t.ChainID,
		Nonce:        t.Nonce,
		Gas:          t.Gas,
		GasTip:       t.GasTip,
		GasFeeCap:    t.GasFeeCap,
		From:         EthID(t.From).String(),
		To:           EthID(t.To).String(),
		Data:         JsonMarshalData(t.Data),
		Amount:       t.Amount,
		Signatures:   t.Signatures,
		ExSignatures: t.ExSignatures,
	}
	return json.Marshal(v)
}

func (t *Transaction) ID() ids.ID {
	if t.id == ids.Empty {
		if _, err := t.calcID(); err != nil {
			panic(err)
		}
	}
	return t.id
}

func (t *Transaction) ShortID() ids.ShortID {
	id := t.ID()
	sid := ids.ShortID{}
	copy(sid[:], id[:])
	return sid
}

func (t *Transaction) Copy() *Transaction {
	x := new(Transaction)
	*x = *t
	if t.Amount != nil {
		x.Amount = new(big.Int).Set(t.Amount)
	}
	x.Data = make([]byte, len(t.Data))
	copy(x.Data, t.Data)
	x.raw = make([]byte, len(t.raw))
	copy(x.raw, t.raw)
	x.Signatures = make([]Signature, len(t.Signatures))
	copy(x.Signatures, t.Signatures)
	x.ExSignatures = make([]Signature, len(t.ExSignatures))
	copy(x.ExSignatures, t.ExSignatures)
	return x
}

// SyntacticVerify verifies that a *Transaction is well-formed.
func (t *Transaction) SyntacticVerify() error {
	if t.Type > TypeDeleteData {
		return fmt.Errorf("invalid transaction type")
	}
	if t.Amount != nil && t.Amount.Sign() < 0 {
		return fmt.Errorf("invalid transaction amount")
	}
	if len(t.Signatures) > math.MaxUint8 {
		return fmt.Errorf("too many transaction signatures")
	}
	if len(t.ExSignatures) > math.MaxUint8 {
		return fmt.Errorf("too many transaction exSignatures")
	}

	if _, err := t.Marshal(); err != nil {
		return fmt.Errorf("transaction marshal error: %v", err)
	}
	if _, err := t.calcID(); err != nil {
		return err
	}
	if _, err := t.calcUnsignedBytes(); err != nil {
		return err
	}
	return nil
}

func (t *Transaction) RequireGas(threshold uint64) uint64 {
	if threshold < MinThresholdGas {
		threshold = MinThresholdGas
	}
	gas := uint64(len(t.Bytes()))
	switch t.Type {
	case TypeMintFee:
		return uint64(0)
	case TypeTransfer:
		return requireGas(threshold, gas+TxTransferGas)
	case TypeTransferReply:
		return requireGas(threshold, gas+TxTransferReply)
	case TypeTransferCash:
		return requireGas(threshold, gas+TxTransferCash)
	case TypeUpdateAccountKeepers:
		return requireGas(threshold, gas+TxUpdateAccountKeepersGas)
	case TypeCreateModel:
		return requireGas(threshold, gas+TxCreateModelGas)
	case TypeUpdateModelKeepers:
		return requireGas(threshold, gas+TxUpdateModelKeepersGas)
	case TypeCreateData:
		return requireGas(threshold, gas+TxCreateDataGas)
	case TypeUpdateData:
		return requireGas(threshold, gas+TxUpdateDataGas)
	case TypeUpdateDataKeepers:
		return requireGas(threshold, gas+TxUpdateDataKeepersGas)
	case TypeUpdateDataKeepersByAuth:
		return requireGas(threshold, gas+TxUpdateDataKeepersByAuthGas)
	case TypeDeleteData:
		return requireGas(threshold, gas+TxDeleteDataGas)
	default:
		return requireGas(threshold, gas)
	}
}

func requireGas(threshold, gas uint64) uint64 {
	if gas <= threshold {
		return gas
	}
	return threshold + uint64(math.Pow(float64(gas-threshold), math.SqrtPhi))
}

func (t *Transaction) BigIntGas() *big.Int {
	return new(big.Int).SetUint64(t.Gas)
}

func (t *Transaction) Equal(o *Transaction) bool {
	if o == nil {
		return false
	}
	if len(o.raw) > 0 && len(t.raw) > 0 {
		return bytes.Equal(o.raw, t.raw)
	}
	if o.Type != t.Type {
		return false
	}
	if o.ChainID != t.ChainID {
		return false
	}
	if o.Nonce != t.Nonce {
		return false
	}
	if o.Gas != t.Gas {
		return false
	}
	if o.GasTip != t.GasTip {
		return false
	}
	if o.GasFeeCap != t.GasFeeCap {
		return false
	}
	if o.From != t.From {
		return false
	}
	if o.To != t.To {
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
	if len(o.Signatures) != len(t.Signatures) {
		return false
	}
	for i := range t.Signatures {
		if o.Signatures[i] != t.Signatures[i] {
			return false
		}
	}
	if len(o.ExSignatures) != len(t.ExSignatures) {
		return false
	}
	for i := range t.ExSignatures {
		if o.ExSignatures[i] != t.ExSignatures[i] {
			return false
		}
	}
	return bytes.Equal(o.Data, t.Data)
}

func (t *Transaction) UnsignedBytes() []byte {
	if len(t.unsignedRaw) == 0 {
		if _, err := t.calcUnsignedBytes(); err != nil {
			panic(err)
		}
	}
	return t.unsignedRaw
}

func (t *Transaction) Bytes() []byte {
	if len(t.raw) == 0 || t.gas != t.Gas {
		if _, err := t.Marshal(); err != nil {
			panic(err)
		}
	}
	return t.raw
}

func (t *Transaction) calcID() (ids.ID, error) {
	if t.id == ids.Empty {
		gas := t.Gas
		raw := t.raw
		// clear gas
		t.Gas = 0
		b, err := t.Marshal()
		t.raw = raw
		t.gas = gas
		t.Gas = gas

		if err != nil {
			return ids.Empty, err
		}
		t.id = IDFromBytes(b)
	}
	return t.id, nil
}

func (t *Transaction) calcUnsignedBytes() ([]byte, error) {
	if len(t.unsignedRaw) == 0 {
		id := t.id
		gas := t.Gas
		raw := t.raw
		sigs := t.Signatures
		exSigs := t.ExSignatures
		// clear gas, id, Signatures, ExSignatures
		t.Gas = 0
		t.id = ids.Empty
		t.Signatures = nil
		t.ExSignatures = nil
		b, err := t.Marshal()
		t.id = id
		t.raw = raw
		t.gas = gas
		t.Gas = gas
		t.Signatures = sigs
		t.ExSignatures = exSigs

		if err != nil {
			return nil, err
		}
		t.unsignedRaw = b
	}
	return t.unsignedRaw, nil
}

func (t *Transaction) Unmarshal(data []byte) error {
	p, err := transactionLDBuilder.Unmarshal(data)
	if err != nil {
		return err
	}
	if v, ok := p.(*bindTransaction); ok {
		t.Type = TxType(v.Type.Value())
		t.ChainID = v.ChainID.Value()
		t.Nonce = v.Nonce.Value()
		t.GasTip = v.GasTip.Value()
		t.GasFeeCap = v.GasFeeCap.Value()
		t.Gas = v.Gas.Value()
		t.Amount = PtrToBigInt(v.Amount)
		t.Data = PtrToBytes(v.Data)
		if t.From, err = PtrToShortID(v.From); err != nil {
			return fmt.Errorf("unmarshal error: %v", err)
		}
		if t.To, err = PtrToShortID(v.To); err != nil {
			return fmt.Errorf("unmarshal error: %v", err)
		}
		if t.Signatures, err = PtrToSignatures(v.Signatures); err != nil {
			return fmt.Errorf("unmarshal error: %v", err)
		}
		if t.ExSignatures, err = PtrToSignatures(v.ExSignatures); err != nil {
			return fmt.Errorf("unmarshal error: %v", err)
		}
		if t.id, err = PtrToID(v.ID); err != nil {
			return fmt.Errorf("unmarshal error: %v", err)
		}
		t.gas = t.Gas
		t.raw = data
		return nil
	}
	return fmt.Errorf("unmarshal error: expected *bindTransaction")
}

func (t *Transaction) Marshal() ([]byte, error) {
	v := &bindTransaction{
		Type:         FromUint8(uint8(t.Type)),
		ChainID:      FromUint64(t.ChainID),
		Nonce:        PtrFromUint64(t.Nonce),
		Gas:          PtrFromUint64(t.Gas),
		GasTip:       PtrFromUint64(t.GasTip),
		GasFeeCap:    PtrFromUint64(t.GasFeeCap),
		From:         PtrFromShortID(t.From),
		To:           PtrFromShortID(t.To),
		Amount:       PtrFromBigInt(t.Amount),
		Data:         PtrFromBytes(t.Data),
		Signatures:   PtrFromSignatures(t.Signatures),
		ExSignatures: PtrFromSignatures(t.ExSignatures),
		ID:           PtrFromID(t.id),
	}
	data, err := transactionLDBuilder.Marshal(v)
	if err != nil {
		return nil, err
	}
	t.gas = t.Gas
	t.raw = data
	return data, nil
}

type bindTransaction struct {
	Type         Uint8
	ChainID      Uint64
	Nonce        *Uint64
	Gas          *Uint64
	GasTip       *Uint64
	GasFeeCap    *Uint64
	From         *[]byte
	To           *[]byte
	Amount       *[]byte
	Data         *[]byte
	Signatures   *[][]byte
	ExSignatures *[][]byte
	ID           *[]byte
}

var transactionLDBuilder *LDBuilder

func init() {
	sch := `
	type Uint8 bytes
	type Uint64 bytes
	type ID20 bytes
	type ID32 bytes
	type Sig65 bytes
	type BigInt bytes
	type Transaction struct {
		Type         Uint8            (rename "t")
		ChainID      Uint64           (rename "c")
		Nonce        nullable Uint64  (rename "n")
		Gas          nullable Uint64  (rename "g")
		GasTip       nullable Uint64  (rename "gt")
		GasFeeCap    nullable Uint64  (rename "gf")
		From         nullable ID20    (rename "fr")
		To           nullable ID20    (rename "to")
		Amount       nullable BigInt  (rename "a")
		Data         nullable Bytes   (rename "d")
		Signatures   nullable [Sig65] (rename "ss")
		ExSignatures nullable [Sig65] (rename "es")
		ID           nullable ID32    (rename "id")
	}
`
	builder, err := NewLDBuilder("Transaction", []byte(sch), (*bindTransaction)(nil))
	if err != nil {
		panic(err)
	}
	transactionLDBuilder = builder
}
