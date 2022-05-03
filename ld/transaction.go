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
	"github.com/ipld/go-ipld-prime/codec/dagcbor"
	"github.com/ipld/go-ipld-prime/datamodel"
	"github.com/ipld/go-ipld-prime/node/basicnode"
	"github.com/ldclabs/ldvm/constants"
	"github.com/ldclabs/ldvm/util"
)

const (
	// The "test" transaction tests that a value of data at the target location
	// is equal to a specified value. test transaction will not write to the block.
	// It should be in a batch transactions.
	TypeTest TxType = iota

	// Transfer
	TypeEth          // send given amount of NanoLDC to a address in ETH transaction
	TypeTransfer     // send given amount of NanoLDC to a address
	TypeTransferPay  // send given amount of NanoLDC to the address who request payment
	TypeTransferCash // cash given amount of NanoLDC to sender, like cash a check.
	TypeExchange     // exchange tokens

	// Account
	TypeAddNonceTable        // add more nonce with expire time to account
	TypeUpdateAccountKeepers // update account's Keepers and Threshold
	TypeCreateTokenAccount   // create a token account
	TypeDestroyTokenAccount  // destroy a token account
	TypeCreateStakeAccount   // create a stake account
	TypeResetStakeAccount    // reset or destroy a stake account
	TypeTakeStake            // take a stake in
	TypeWithdrawStake        // withdraw stake
	TypeOpenLending
	TypeCloseLending
	TypeBorrow
	TypeRepay

	// Model
	TypeCreateModel        // create a data model
	TypeUpdateModelKeepers // update data model's Keepers and Threshold

	// Data
	TypeCreateData              // create a data from the model
	TypeUpdateData              // update the data's Data
	TypeUpdateDataKeepers       // update data's Keepers and Threshold
	TypeUpdateDataKeepersByAuth // update data's Keepers and Threshold by authorization
	TypeDeleteData              // delete the data

	// punish transaction can be issued by genesisAccount
	// we can only punish illegal data
	TypePunish TxType = 255
)

const (
	TxEthGas          = uint64(42)
	TxTransferGas     = uint64(42)
	TxTransferPayGas  = uint64(42)
	TxTransferCashGas = uint64(42)
	TxExchangeGas     = uint64(42)

	TxAddNonceTableGas        = uint64(100)
	TxUpdateAccountKeepersGas = uint64(1000)
	TxCreateTokenAccountGas   = uint64(1000)
	TxDestroyTokenAccountGas  = uint64(1000)
	TxCreateStakeAccountGas   = uint64(1000)
	TxTakeStakeGas            = uint64(1000)
	TxWithdrawStakeGas        = uint64(1000)
	TxResetStakeAccountGas    = uint64(1000)

	TxCreateModelGas        = uint64(500)
	TxUpdateModelKeepersGas = uint64(500)

	TxCreateDataGas              = uint64(100)
	TxUpdateDataGas              = uint64(100)
	TxUpdateDataKeepersGas       = uint64(100)
	TxUpdateDataKeepersByAuthGas = uint64(200)
	TxDeleteDataGas              = uint64(200)

	TxPunishGas = uint64(1000)

	MinThresholdGas = uint64(1000)
)

// gChainID will be updated by SetChainID when VM.Initialize
var gChainID = uint64(2357)

// TxType is an uint8 representing the type of the tx
type TxType uint8

func TxTypeString(t TxType) string {
	switch t {
	case TypeTest:
		return "TestTx"
	case TypeEth:
		return "EthTx"
	case TypeTransfer:
		return "TransferTx"
	case TypeTransferPay:
		return "TransferPayTx"
	case TypeTransferCash:
		return "TransferCashTx"
	case TypeExchange:
		return "ExchangeTx"
	case TypeAddNonceTable:
		return "TypeAddNonceTable"
	case TypeUpdateAccountKeepers:
		return "UpdateAccountKeepersTx"
	case TypeCreateTokenAccount:
		return "CreateTokenAccountTx"
	case TypeDestroyTokenAccount:
		return "DestroyTokenAccountTx"
	case TypeCreateStakeAccount:
		return "CreateStakeAccountTx"
	case TypeTakeStake:
		return "TakeStakeTx"
	case TypeWithdrawStake:
		return "WithdrawStakeTx"
	case TypeResetStakeAccount:
		return "ResetStakeAccountTx"
	case TypeCreateModel:
		return "CreateModelTx"
	case TypeUpdateModelKeepers:
		return "UpdateModelKeepersTx"
	case TypeCreateData:
		return "CreateDataTx"
	case TypeUpdateData:
		return "UpdateDataTx"
	case TypeUpdateDataKeepers:
		return "UpdateDataKeepersTx"
	case TypeUpdateDataKeepersByAuth:
		return "UpdateDataKeepersByAuthTx"
	case TypeDeleteData:
		return "DeleteDataTx"
	case TypePunish:
		return "PunishTx"
	default:
		return "UnknownTx"
	}
}

type Transaction struct {
	Type         TxType
	ChainID      uint64
	Nonce        uint64
	Gas          uint64 // calculate when build block.
	GasTip       uint64
	GasFeeCap    uint64
	Token        ids.ShortID
	From         ids.ShortID
	To           ids.ShortID
	Amount       *big.Int
	Data         []byte
	Signatures   []util.Signature
	ExSignatures []util.Signature

	// external assignment
	gas         uint64
	id          ids.ID
	unsignedRaw []byte // raw bytes for sign
	raw         []byte // the transaction's raw bytes, included id and sigs.
	AddTime     uint64
	Priority    uint64
	Height      uint64 // block's timestamp
	Timestamp   uint64 // block's timestamp
	Err         error
	// support for batch transactions
	// they are processed in the same block, one fail all fail
	batch Txs
}

func NewBatchTx(txs []*Transaction) (*Transaction, error) {
	if len(txs) <= 1 {
		return nil, fmt.Errorf("not batch transactions")
	}

	tx := &Transaction{}
	maxSize := 0
	var err error
	for i, t := range txs {
		if t.Type == TypeTest {
			continue
		}

		if err = t.SyntacticVerify(); err != nil {
			return nil, err
		}
		if size := len(t.UnsignedBytes()); size > maxSize {
			tx = txs[i] // find the max UnsignedBytes tx as batch transactions' container
			maxSize = size
		}
	}
	tx = tx.Copy()
	if err = tx.SyntacticVerify(); err != nil {
		return nil, err
	}

	tx.batch = txs
	return tx, nil
}

type jsonTransaction struct {
	ID           ids.ID           `json:"id"`
	Type         TxType           `json:"type"`
	TypeStr      string           `json:"typeString"`
	ChainID      uint64           `json:"chainID"`
	Nonce        uint64           `json:"nonce"`
	Gas          uint64           `json:"gas"` // calculate when build block.
	GasTip       uint64           `json:"gasTip"`
	GasFeeCap    uint64           `json:"gasFeeCap"`
	Token        string           `json:"token,omitempty"`
	From         string           `json:"from"`
	To           string           `json:"to"`
	Amount       *big.Int         `json:"amount,omitempty"`
	Data         json.RawMessage  `json:"data"`
	Signatures   []util.Signature `json:"signatures"`
	ExSignatures []util.Signature `json:"exSignatures"`
	Err          string           `json:"error"`
}

type Txs []*Transaction

func MarshalTxs(txs []*Transaction) ([]byte, error) {
	if len(txs) == 0 {
		return nil, nil
	}
	var buf bytes.Buffer
	err := Recover("MarshalTxs", func() error {
		nb := basicnode.Prototype.List.NewBuilder()
		la, er := nb.BeginList(int64(len(txs)))
		if er != nil {
			return er
		}
		for i := range txs {
			la.AssembleValue().AssignBytes(txs[i].Bytes())
		}
		if er = la.Finish(); er != nil {
			return er
		}
		return dagcbor.Encode(nb.Build(), &buf)
	})
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func UnmarshalTxs(data []byte) ([]*Transaction, error) {
	if len(data) == 0 {
		return nil, nil
	}

	var rt *[]*Transaction
	err := Recover("UnmarshalTxs", func() error {
		var er error
		nb := basicnode.Prototype.List.NewBuilder()
		if er = dagcbor.Decode(nb, bytes.NewReader(data)); er != nil {
			return er
		}
		list := nb.Build()
		le := int(list.Length())
		txs := make([]*Transaction, le)

		var node datamodel.Node
		var d []byte
		for i := 0; i < le; i++ {
			if node, er = list.LookupByIndex(int64(i)); er != nil {
				return er
			}
			if d, er = node.AsBytes(); er != nil {
				return er
			}

			txs[i] = &Transaction{}
			if er = txs[i].Unmarshal(d); er != nil {
				return er
			}
			if er = txs[i].SyntacticVerify(); er != nil {
				return er
			}
		}
		rt = &txs
		return nil
	})

	if err != nil {
		return nil, err
	}
	return *rt, nil
}

func (t *Transaction) MarshalJSON() ([]byte, error) {
	if t == nil {
		return util.Null, nil
	}

	if len(t.batch) > 0 {
		arr := make([]json.RawMessage, len(t.batch))
		for i, tx := range t.batch {
			v, err := tx.MarshalJSON()
			if err != nil {
				return nil, err
			}
			arr[i] = v
		}
		return json.Marshal(arr)
	}

	v := &jsonTransaction{
		ID:           t.ID(),
		Type:         t.Type,
		TypeStr:      TxTypeString(t.Type),
		ChainID:      t.ChainID,
		Nonce:        t.Nonce,
		Gas:          t.Gas,
		GasTip:       t.GasTip,
		GasFeeCap:    t.GasFeeCap,
		Token:        util.TokenSymbol(t.Token).String(),
		From:         util.EthID(t.From).String(),
		To:           util.EthID(t.To).String(),
		Data:         util.JSONMarshalData(t.Data),
		Amount:       t.Amount,
		Signatures:   t.Signatures,
		ExSignatures: t.ExSignatures,
	}
	if t.Err != nil {
		v.Err = t.Err.Error()
	}
	return json.Marshal(v)
}

func (t *Transaction) IsBatched() bool {
	return len(t.batch) > 0
}

func (t *Transaction) Txs() []*Transaction {
	return t.batch
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

	x.Amount = new(big.Int).Set(t.Amount)
	x.Data = make([]byte, len(t.Data))
	copy(x.Data, t.Data)
	x.Signatures = make([]util.Signature, len(t.Signatures))
	copy(x.Signatures, t.Signatures)
	x.ExSignatures = make([]util.Signature, len(t.ExSignatures))
	copy(x.ExSignatures, t.ExSignatures)
	if len(t.batch) > 0 {
		x.batch = make([]*Transaction, len(t.batch))
		for i := range t.batch {
			x.batch[i] = t.batch[i].Copy()
		}
		return x
	}
	x.unsignedRaw = nil
	x.raw = nil
	return x
}

// SyntacticVerify verifies that a *Transaction is well-formed.
func (t *Transaction) SyntacticVerify() error {
	if t == nil {
		return fmt.Errorf("invalid Transaction")
	}

	if t.ChainID != gChainID {
		return fmt.Errorf("invalid ChainID, expected %d, got %d", gChainID, t.ChainID)
	}

	if t.Type > TypeDeleteData {
		return fmt.Errorf("invalid type")
	}
	if t.Token != constants.LDCAccount && util.TokenSymbol(t.Token).String() == "" {
		return fmt.Errorf("invalid token symbol")
	}
	if t.Amount == nil || t.Amount.Sign() < 0 {
		return fmt.Errorf("invalid amount")
	}
	if len(t.Signatures) > math.MaxUint8 {
		return fmt.Errorf("invalid signatures, too many")
	}
	if len(t.ExSignatures) > math.MaxUint8 {
		return fmt.Errorf("invalid exSignatures, too many")
	}

	if _, err := t.Marshal(); err != nil {
		return fmt.Errorf("Transaction marshal error: %v", err)
	}
	if _, err := t.calcID(); err != nil {
		return err
	}
	if _, err := t.calcUnsignedBytes(); err != nil {
		return err
	}
	return nil
}

func (t *Transaction) SetPriority(threshold, nowSeconds uint64) {
	priority := t.GasTip * threshold
	gas := t.RequireGas(threshold)
	if v := t.GasTip * gas; v > priority {
		priority = v
	}
	// Consider gossip network overhead, ignoring small time differences
	// not promote priority if not processed more than 120 seconds(tx maybe invalid...)
	if du := nowSeconds - t.AddTime; du > 3 && du <= 120 {
		// A delay of 1 seconds is equivalent to 100 gasTip
		priority += du * 100 * threshold
	}
	t.Priority = priority
}

func (t *Transaction) RequireGas(threshold uint64) uint64 {
	return RequireGas(t.Type, uint64(len(t.UnsignedBytes())), threshold)
}

func RequireGas(ty TxType, unsignedBytes uint64, threshold uint64) uint64 {
	if threshold < MinThresholdGas {
		threshold = MinThresholdGas
	}
	gas := unsignedBytes
	switch ty {
	case TypeEth:
		return requireGas(threshold, gas+TxEthGas)
	case TypeTransfer:
		return requireGas(threshold, gas+TxTransferGas)
	case TypeTransferPay:
		return requireGas(threshold, gas+TxTransferPayGas)
	case TypeTransferCash:
		return requireGas(threshold, gas+TxTransferCashGas)
	case TypeExchange:
		return requireGas(threshold, gas+TxExchangeGas)

	case TypeAddNonceTable:
		return requireGas(threshold, gas+TxAddNonceTableGas)
	case TypeUpdateAccountKeepers:
		return requireGas(threshold, gas+TxUpdateAccountKeepersGas)
	case TypeCreateTokenAccount:
		return requireGas(threshold, gas+TxCreateTokenAccountGas)
	case TypeDestroyTokenAccount:
		return requireGas(threshold, gas+TxDestroyTokenAccountGas)
	case TypeCreateStakeAccount:
		return requireGas(threshold, gas+TxCreateStakeAccountGas)
	case TypeTakeStake:
		return requireGas(threshold, gas+TxTakeStakeGas)
	case TypeWithdrawStake:
		return requireGas(threshold, gas+TxWithdrawStakeGas)
	case TypeResetStakeAccount:
		return requireGas(threshold, gas+TxResetStakeAccountGas)

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

	case TypePunish:
		return requireGas(threshold, gas+TxPunishGas)

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

func (t *Transaction) GasUnits() *big.Int {
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
	if o.Token != t.Token {
		return false
	}
	if o.From != t.From {
		return false
	}
	if o.To != t.To {
		return false
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
		t.id = util.IDFromBytes(b)
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
		if !v.Type.Valid() {
			return fmt.Errorf("unmarshal error: invalid uint8")
		}
		if !v.ChainID.Valid() ||
			!v.Nonce.Valid() ||
			!v.GasTip.Valid() ||
			!v.GasFeeCap.Valid() ||
			!v.Gas.Valid() {
			return fmt.Errorf("unmarshal error: invalid uint64")
		}
		t.Type = TxType(v.Type.Value())
		t.ChainID = v.ChainID.Value()
		t.Nonce = v.Nonce.Value()
		t.GasTip = v.GasTip.Value()
		t.GasFeeCap = v.GasFeeCap.Value()
		t.Gas = v.Gas.Value()
		t.Amount = v.Amount.Value()
		t.Data = PtrToBytes(v.Data)
		if t.Token, err = PtrToShortID(v.Token); err != nil {
			return fmt.Errorf("unmarshal error: %v", err)
		}
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
	if t.IsBatched() {
		return nil, fmt.Errorf("can not marshal batch transactions")
	}
	v := &bindTransaction{
		Type:         FromUint8(uint8(t.Type)),
		ChainID:      FromUint64(t.ChainID),
		Nonce:        PtrFromUint64(t.Nonce),
		Gas:          PtrFromUint64(t.Gas),
		GasTip:       PtrFromUint64(t.GasTip),
		GasFeeCap:    PtrFromUint64(t.GasFeeCap),
		Token:        PtrFromShortID(ids.ShortID(t.Token)),
		From:         PtrFromShortID(t.From),
		To:           PtrFromShortID(t.To),
		Amount:       PtrFromUint(t.Amount),
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
	Token        *[]byte
	From         *[]byte
	To           *[]byte
	Amount       *BigUint
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
	type BigUint bytes
	type Transaction struct {
		Type         Uint8            (rename "t")
		ChainID      Uint64           (rename "c")
		Nonce        nullable Uint64  (rename "n")
		Gas          nullable Uint64  (rename "g")
		GasTip       nullable Uint64  (rename "gt")
		GasFeeCap    nullable Uint64  (rename "gf")
		Token        nullable ID20    (rename "tk")
		From         nullable ID20    (rename "fr")
		To           nullable ID20    (rename "to")
		Amount       nullable BigUint (rename "a")
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
