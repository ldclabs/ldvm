// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package ld

import (
	"fmt"
	"math"
	"math/big"
	"sort"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ldclabs/ldvm/constants"
	"github.com/ldclabs/ldvm/util"
)

const (
	// gasTipPerSec: A delay of 1 seconds is equivalent to 1000 gasTip
	gasTipPerSec  = constants.MicroLDC
	maxTxDataSize = 1024 * 256
)

// gChainID will be updated by SetChainID when VM.Initialize
var gChainID = uint64(2357)

type Signer interface {
	Sign(data []byte) (util.Signature, error)
}

// TxData represents a complete transaction issued from client
type TxData struct {
	Type         TxType            `cbor:"t" json:"type"`
	ChainID      uint64            `cbor:"c" json:"chainID"`
	Nonce        uint64            `cbor:"n" json:"nonce"`
	GasTip       uint64            `cbor:"gt" json:"gasTip"`
	GasFeeCap    uint64            `cbor:"gf" json:"gasFeeCap"`
	From         util.EthID        `cbor:"fr" json:"from"`
	To           *util.EthID       `cbor:"to,omitempty" json:"to,omitempty"`
	Token        *util.TokenSymbol `cbor:"tk,omitempty" json:"token,omitempty"`
	Amount       *big.Int          `cbor:"a,omitempty" json:"amount,omitempty"`
	Data         util.RawData      `cbor:"d,omitempty" json:"data,omitempty"`
	Signatures   []util.Signature  `cbor:"ss,omitempty" json:"signatures,omitempty"`
	ExSignatures []util.Signature  `cbor:"es,omitempty" json:"exSignatures,omitempty"`

	// external assignment fields
	id       ids.ID `cbor:"-" json:"-"`
	gas      uint64 `cbor:"-" json:"-"`
	raw      []byte `cbor:"-" json:"-"`
	unsigned []byte `cbor:"-" json:"-"`
	eth      *TxEth `cbor:"-" json:"-"`
}

// SyntacticVerify verifies that a *TxData is well-formed.
func (t *TxData) SyntacticVerify() error {
	errp := util.ErrPrefix("TxData.SyntacticVerify error: ")

	switch {
	case t == nil:
		return errp.Errorf("nil pointer")

	case !AllTxTypes.Has(t.Type):
		return errp.Errorf("invalid type %d", t.Type)

	case t.ChainID != gChainID:
		return errp.Errorf("invalid ChainID, expected %d, got %d", gChainID, t.ChainID)

	case t.Token != nil && !t.Token.Valid():
		return errp.Errorf("invalid token symbol %q", t.Token.GoString())

	case t.Data != nil && len(t.Data) == 0:
		return errp.Errorf("empty data")

	case t.Signatures != nil && len(t.Signatures) == 0:
		return errp.Errorf("empty signatures")

	case t.ExSignatures != nil && len(t.ExSignatures) == 0:
		return errp.Errorf("empty exSignatures")

	case len(t.Signatures) > MaxKeepers:
		return errp.Errorf("too many signatures")

	case len(t.ExSignatures) > MaxKeepers:
		return errp.Errorf("too many exSignatures")
	}

	if t.Amount != nil {
		if t.Amount.Sign() < 0 {
			return errp.Errorf("invalid amount")
		}
		if t.To == nil {
			return errp.Errorf("nil to together with amount")
		}
	}

	var err error
	if t.raw, err = t.Marshal(); err != nil {
		return errp.ErrorIf(err)
	}
	size := len(t.raw)
	if size > maxTxDataSize {
		return errp.Errorf("size too large, expected <= %d, got %d", maxTxDataSize, size)
	}

	t.calcUnsignedBytes()
	t.gas = t.Type.Gas() + uint64(math.Pow(float64(size), math.SqrtPhi))
	t.id = ids.ID(util.HashFromData(t.raw))
	return nil
}

// ID returns the ID of the transaction that generated in TxData.SyntacticVerify().
func (t *TxData) ID() ids.ID {
	return t.id
}

// Gas returns the gas of the transaction that generated in TxData.SyntacticVerify().
func (t *TxData) Gas() uint64 {
	return t.gas
}

func (t *TxData) Bytes() []byte {
	if len(t.raw) == 0 {
		t.raw = MustMarshal(t)
	}
	return t.raw
}

func (t *TxData) UnsignedBytes() []byte {
	if len(t.unsigned) == 0 {
		t.calcUnsignedBytes()
	}
	return t.unsigned
}

func (t *TxData) calcUnsignedBytes() {
	sigs := t.Signatures
	exSigs := t.ExSignatures
	t.Signatures = nil
	t.ExSignatures = nil
	t.unsigned = MustMarshal(t)
	t.Signatures = sigs
	t.ExSignatures = exSigs
}

func (t *TxData) Unmarshal(data []byte) error {
	return util.ErrPrefix("TxData.Unmarshal error: ").
		ErrorIf(util.UnmarshalCBOR(data, t))
}

func (t *TxData) Marshal() ([]byte, error) {
	return util.ErrPrefix("TxData.Marshal error: ").
		ErrorMap(util.MarshalCBOR(t))
}

func (t *TxData) SignWith(signers ...Signer) error {
	data := t.UnsignedBytes()
	for _, signer := range signers {
		sig, err := signer.Sign(data)
		if err != nil {
			return util.ErrPrefix("TxData.SignWith error: ").ErrorIf(err)
		}
		t.Signatures = append(t.Signatures, sig)
	}
	return nil
}

func (t *TxData) ExSignWith(signers ...Signer) error {
	for _, signer := range signers {
		sig, err := signer.Sign(t.Data)
		if err != nil {
			return util.ErrPrefix("TxData.ExSignWith error: ").ErrorIf(err)
		}
		t.ExSignatures = append(t.ExSignatures, sig)
	}
	return nil
}

func (t *TxData) ToTransaction() *Transaction {
	tx := new(Transaction)
	tx.setTxData(t)
	return tx
}

type Transaction struct {
	// same as TxData
	Type         TxType            `cbor:"t" json:"type"`
	ChainID      uint64            `cbor:"c" json:"chainID"`
	Nonce        uint64            `cbor:"n" json:"nonce"`
	GasTip       uint64            `cbor:"gt" json:"gasTip"`
	GasFeeCap    uint64            `cbor:"gf" json:"gasFeeCap"`
	From         util.EthID        `cbor:"fr" json:"from"`
	To           *util.EthID       `cbor:"to,omitempty" json:"to,omitempty"`
	Token        *util.TokenSymbol `cbor:"tk,omitempty" json:"token,omitempty"`
	Amount       *big.Int          `cbor:"a,omitempty" json:"amount,omitempty"`
	Data         util.RawData      `cbor:"d,omitempty" json:"data,omitempty"`
	Signatures   []util.Signature  `cbor:"ss,omitempty" json:"signatures,omitempty"`
	ExSignatures []util.Signature  `cbor:"es,omitempty" json:"exSignatures,omitempty"`

	// external assignment fields
	ID        ids.ID  `cbor:"id" json:"id"`
	Err       error   `cbor:"-" json:"error,omitempty"`
	Height    uint64  `cbor:"-" json:"-"` // block's timestamp
	Timestamp uint64  `cbor:"-" json:"-"` // block's timestamp
	priority  uint64  `cbor:"-" json:"-"`
	dp        uint64  `cbor:"-" json:"-"` // dynamic priority for sorting
	tx        *TxData `cbor:"-" json:"-"`
	raw       []byte  `cbor:"-" json:"-"` // the transaction's raw bytes, included id and sigs.
	// support for batch transactions
	// they are processed in the same block, one fail all fail
	batch Txs `cbor:"-" json:"-"`
}

// SyntacticVerify verifies that a *Transaction is well-formed.
func (t *Transaction) SyntacticVerify() error {
	errp := util.ErrPrefix("Transaction.SyntacticVerify error: ")
	if t == nil {
		return errp.Errorf("nil pointer")
	}

	var err error
	t.tx = t.TxData(t.tx)
	if t.Type == TypeEth && t.tx.eth == nil {
		eth := new(TxEth)

		if err = eth.Unmarshal(t.Data); err != nil {
			return errp.ErrorIf(err)
		}
		if err = eth.SyntacticVerify(); err != nil {
			return errp.ErrorIf(err)
		}
		t.tx.eth = eth
	}

	if err = t.tx.SyntacticVerify(); err != nil {
		return errp.ErrorIf(err)
	}

	t.ID = t.tx.id
	t.priority = (t.tx.GasTip + 1) * t.tx.gas
	if t.raw, err = t.Marshal(); err != nil {
		return errp.ErrorIf(err)
	}
	return nil
}

func (t *Transaction) Gas() uint64 {
	return t.tx.gas
}

func (t *Transaction) Bytes() []byte {
	if len(t.raw) == 0 {
		t.raw = MustMarshal(t)
	}
	return t.raw
}

func (t *Transaction) BytesSize() int {
	total := 0
	switch {
	case t.IsBatched():
		for _, tx := range t.batch {
			total += tx.BytesSize()
		}
	default:
		total = len(t.Bytes())
	}
	return total
}

func (t *Transaction) UnmarshalTx(data []byte) error {
	tx := new(TxData)
	if err := tx.Unmarshal(data); err != nil {
		return util.ErrPrefix("Transaction.UnmarshalTx error: ").ErrorIf(err)
	}
	t.setTxData(tx)
	return nil
}

func (t *Transaction) setTxData(tx *TxData) {
	t.Type = tx.Type
	t.ChainID = tx.ChainID
	t.Nonce = tx.Nonce
	t.GasTip = tx.GasTip
	t.GasFeeCap = tx.GasFeeCap
	t.Token = tx.Token
	t.From = tx.From
	t.To = tx.To
	t.Amount = tx.Amount
	t.Data = tx.Data
	t.Signatures = tx.Signatures
	t.ExSignatures = tx.ExSignatures
	t.tx = tx
}

func (t *Transaction) TxData(tx *TxData) *TxData {
	if tx == nil {
		tx = new(TxData)
	}
	tx.Type = t.Type
	tx.ChainID = t.ChainID
	tx.Nonce = t.Nonce
	tx.GasTip = t.GasTip
	tx.GasFeeCap = t.GasFeeCap
	tx.Token = t.Token
	tx.From = t.From
	tx.To = t.To
	tx.Amount = t.Amount
	tx.Data = t.Data
	tx.Signatures = t.Signatures
	tx.ExSignatures = t.ExSignatures
	return tx
}

func (t *Transaction) Unmarshal(data []byte) error {
	return util.ErrPrefix("Transaction.Unmarshal error: ").
		ErrorIf(util.UnmarshalCBOR(data, t))
}

func (t *Transaction) Marshal() ([]byte, error) {
	return util.ErrPrefix("Transaction.Marshal error: ").
		ErrorMap(util.MarshalCBOR(t))
}

func (t *Transaction) Signers() (signers util.EthIDs, err error) {
	errp := util.ErrPrefix("Transaction.Signers error: ")

	switch t.Type {
	case TypeEth:
		if t.tx.eth == nil {
			return nil, errp.Errorf("invalid TypeEth tx")
		}
		signers, err = t.tx.eth.Signers()
	default:
		signers, err = util.DeriveSigners(t.tx.UnsignedBytes(), t.Signatures)
	}

	if err != nil {
		return nil, errp.ErrorIf(err)
	}
	return signers, nil
}

func (t *Transaction) ExSigners() (util.EthIDs, error) {
	errp := util.ErrPrefix("Transaction.ExSigners error: ")

	signers, err := util.DeriveSigners(t.Data, t.ExSignatures)
	if err != nil {
		return nil, errp.ErrorIf(err)
	}
	return signers, nil
}

func (t *Transaction) NeedApprove(approver *util.EthID, approveList TxTypes) bool {
	switch {
	case approver == nil:
		return false
	case len(approveList) == 0:
		return true
	default:
		for _, ty := range approveList {
			if t.Type == ty {
				return true
			}
		}
		return false
	}
}

func (t *Transaction) IsBatched() bool {
	return len(t.batch) > 0
}

// Txs in batch should keep origin order.
func (t *Transaction) Txs() Txs {
	return t.batch
}

func (t *Transaction) ShortID() ids.ShortID {
	sid := ids.ShortID{}
	copy(sid[:], t.ID[:])
	return sid
}

// Copy is not a deep copy, used for json.Marshaling
func (t *Transaction) Copy() *Transaction {
	x := new(Transaction)
	*x = *t
	x.tx = new(TxData)
	*(x.tx) = *(t.tx)
	return x
}

func (t *Transaction) Eth() *TxEth {
	return t.tx.eth
}

func NewBatchTx(txs ...*Transaction) (*Transaction, error) {
	errp := util.ErrPrefix("NewBatchTx error: ")

	if len(txs) <= 1 {
		return nil, errp.Errorf("not batch transactions")
	}

	maxPriority := uint64(0)
	var err error
	var tx *Transaction
	for i, t := range txs {
		if err = t.SyntacticVerify(); err != nil {
			return nil, errp.ErrorIf(err)
		}
		if t.IsBatched() {
			return nil, errp.Errorf("tx %s is already batched", t.ID)
		}
		if t.priority > maxPriority {
			tx = txs[i] // find the max priority tx as batch transactions' container
			maxPriority = t.priority
		}
	}
	if maxPriority == 0 {
		return nil, errp.Errorf("no invalid transaction")
	}
	tx = tx.Copy()
	if err = tx.SyntacticVerify(); err != nil {
		return nil, errp.ErrorIf(err)
	}
	tx.priority = maxPriority
	tx.batch = txs
	return tx, nil
}

type Txs []*Transaction

func (txs *Txs) Unmarshal(data []byte) error {
	return util.ErrPrefix("Txs.Unmarshal error: ").
		ErrorIf(util.UnmarshalCBOR(data, txs))
}

func (txs Txs) Marshal() ([]byte, error) {
	return util.ErrPrefix("Txs.Marshal error: ").
		ErrorMap(util.MarshalCBOR(txs))
}

func (txs Txs) To() (*Transaction, error) {
	switch len(txs) {
	case 0:
		return nil, fmt.Errorf("empty txs")

	case 1:
		tx := txs[0]
		if err := tx.SyntacticVerify(); err != nil {
			return nil, err
		}
		return tx, nil

	default:
		return NewBatchTx(txs...)
	}
}

func (txs Txs) BytesSize() int {
	s := 0
	for _, tx := range txs {
		s += tx.BytesSize()
	}
	return s
}

type group struct {
	ts Txs
	dp uint64 // dynamic priority for sorting
}

type groupSet map[util.EthID]*group

func (set groupSet) Add(txs ...*Transaction) {
	for i := range txs {
		tx := txs[i]
		g := set[tx.From]
		switch {
		case g == nil:
			g = &group{ts: Txs{tx}, dp: tx.priority}
			set[tx.From] = g
		default:
			// txs from the same sender share the priority
			g.dp += tx.priority
			g.ts = append(g.ts, tx)
		}
	}
}

func (txs Txs) Sort() {
	// first: group by sender - tx.From
	set := make(groupSet, len(txs))
	for i := range txs {
		tx := txs[i]
		if tx.IsBatched() {
			set.Add(tx.Txs()...)
		} else {
			set.Add(tx)
		}
	}
	// then: rebalance the dynamic priority for the same sender, small nonce tx get larger priority.
	for _, g := range set {
		n := uint64(len(g.ts))
		g.dp = n + g.dp/n
		sort.SliceStable(g.ts, func(i, j int) bool {
			return g.ts[i].Nonce < g.ts[j].Nonce
		})
		for i, tx := range g.ts {
			tx.dp = g.dp - uint64(i)
		}
	}
	// then: pick the max priority for batch txs again.
	for _, tx := range txs {
		if tx.IsBatched() {
			tx.dp = 0
			for _, ti := range tx.Txs() {
				if tx.dp < ti.dp {
					tx.dp = ti.dp
				}
			}
		}
	}
	// last: all txs sort by priority
	sort.SliceStable(txs, func(i, j int) bool {
		return txs[i].dp > txs[j].dp
	})
}
