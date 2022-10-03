// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package ld

import (
	"fmt"
	"math"
	"math/big"
	"sort"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ldclabs/ldvm/constants"
	"github.com/ldclabs/ldvm/util"
	"golang.org/x/crypto/sha3"
)

const (
	// gasTipPerSec: A delay of 1 seconds is equivalent to 1000 gasTip
	gasTipPerSec  = constants.MicroLDC
	maxTxDataSize = 1024 * 256
)

// gChainID will be updated by SetChainID when VM.Initialize
var gChainID = uint64(2357)

var EthSigner = types.NewLondonSigner(big.NewInt(2357))

// SetChainID will be set when VM.Initialize
func SetChainID(id uint64) {
	gChainID = id
	EthSigner = types.NewLondonSigner(big.NewInt(int64(id)))
}

type Signer interface {
	SignHash(digestHash []byte) (util.Signature, error)
}

// TxData represents a complete transaction issued from client
type TxData struct {
	Type      TxType            `cbor:"t" json:"type"`
	ChainID   uint64            `cbor:"c" json:"chainID"`
	Nonce     uint64            `cbor:"n" json:"nonce"`
	GasTip    uint64            `cbor:"gt" json:"gasTip"`
	GasFeeCap uint64            `cbor:"gf" json:"gasFeeCap"`
	From      util.EthID        `cbor:"fr" json:"from"`
	To        *util.EthID       `cbor:"to,omitempty" json:"to,omitempty"`
	Token     *util.TokenSymbol `cbor:"tk,omitempty" json:"token,omitempty"`
	Amount    *big.Int          `cbor:"a,omitempty" json:"amount,omitempty"`
	Data      util.RawData      `cbor:"d,omitempty" json:"data,omitempty"`

	// external assignment fields
	raw []byte `cbor:"-" json:"-"`
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
	}

	if t.Amount != nil {
		if t.Amount.Sign() < 0 {
			return errp.Errorf("invalid amount")
		}
		if t.To == nil {
			return errp.Errorf("nil \"to\" together with amount")
		}
	}

	var err error
	if t.raw, err = t.Marshal(); err != nil {
		return errp.ErrorIf(err)
	}
	return nil
}

func (t *TxData) Bytes() []byte {
	if len(t.raw) == 0 {
		t.raw = MustMarshal(t)
	}
	return t.raw
}

func (t *TxData) Unmarshal(data []byte) error {
	return util.ErrPrefix("TxData.Unmarshal error: ").
		ErrorIf(util.UnmarshalCBOR(data, t))
}

func (t *TxData) Marshal() ([]byte, error) {
	return util.ErrPrefix("TxData.Marshal error: ").
		ErrorMap(util.MarshalCBOR(t))
}

func (t *TxData) ToTransaction() *Transaction {
	return &Transaction{Tx: *t}
}

type Transaction struct {
	Tx           TxData           `cbor:"tx" json:"tx"`
	Signatures   []util.Signature `cbor:"ss,omitempty" json:"sigs,omitempty"`
	ExSignatures []util.Signature `cbor:"es,omitempty" json:"exSigs,omitempty"`

	// external assignment fields
	ID        ids.ID `cbor:"-" json:"id"`
	Err       error  `cbor:"-" json:"error,omitempty"`
	Height    uint64 `cbor:"-" json:"-"` // block's timestamp
	Timestamp uint64 `cbor:"-" json:"-"` // block's timestamp
	priority  uint64 `cbor:"-" json:"-"`
	dp        uint64 `cbor:"-" json:"-"` // dynamic priority for sorting
	raw       []byte `cbor:"-" json:"-"` // the transaction's raw bytes, included id and sigs.
	gas       uint64 `cbor:"-" json:"-"`
	eth       *TxEth `cbor:"-" json:"-"`
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

	switch {
	case t.Signatures != nil && len(t.Signatures) == 0:
		return errp.Errorf("empty signatures")

	case t.ExSignatures != nil && len(t.ExSignatures) == 0:
		return errp.Errorf("empty exSignatures")

	case len(t.Signatures) > MaxKeepers:
		return errp.Errorf("too many signatures")

	case len(t.ExSignatures) > MaxKeepers:
		return errp.Errorf("too many exSignatures")
	}

	var err error
	if t.Tx.Type == TypeEth && t.eth == nil {
		eth := new(TxEth)

		if err = eth.Unmarshal(t.Tx.Data); err != nil {
			return errp.ErrorIf(err)
		}
		if err = eth.SyntacticVerify(); err != nil {
			return errp.ErrorIf(err)
		}
		t.eth = eth
	}

	if err = t.Tx.SyntacticVerify(); err != nil {
		return errp.ErrorIf(err)
	}

	if t.raw, err = t.Marshal(); err != nil {
		return errp.ErrorIf(err)
	}

	size := len(t.raw)
	if size > maxTxDataSize {
		return errp.Errorf("size too large, expected <= %d, got %d", maxTxDataSize, size)
	}

	t.ID = ids.ID(util.HashFromData(t.raw))
	t.gas = t.Tx.Type.Gas() + uint64(math.Pow(float64(size), math.SqrtPhi))
	t.priority = (t.Tx.GasTip + 1) * t.gas
	return nil
}

func (t *Transaction) Gas() uint64 {
	return t.gas
}

func (t *Transaction) Bytes() []byte {
	if len(t.raw) == 0 {
		t.raw = MustMarshal(t)
	}
	return t.raw
}

func (t *Transaction) UnsignedBytes() []byte {
	return t.Tx.Bytes()
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

// func (t *Transaction) UnmarshalTx(data []byte) error {
// 	return util.ErrPrefix("Transaction.UnmarshalTx error: ").
// 		ErrorIf(t.Tx.Unmarshal(data))
// }

func (t *Transaction) Unmarshal(data []byte) error {
	return util.ErrPrefix("Transaction.Unmarshal error: ").
		ErrorIf(util.UnmarshalCBOR(data, t))
}

func (t *Transaction) Marshal() ([]byte, error) {
	return util.ErrPrefix("Transaction.Marshal error: ").
		ErrorMap(util.MarshalCBOR(t))
}

func (t *Transaction) SignWith(signers ...Signer) error {
	datahash := sha3.Sum256(t.Tx.Bytes())
	t.Signatures = make([]util.Signature, 0, len(signers))
	for _, signer := range signers {
		sig, err := signer.SignHash(datahash[:])
		if err != nil {
			return util.ErrPrefix("TxData.SignWith error: ").ErrorIf(err)
		}
		t.Signatures = append(t.Signatures, sig)
	}
	return nil
}

func (t *Transaction) ExSignWith(signers ...Signer) error {
	datahash := sha3.Sum256(t.Tx.Data)
	t.ExSignatures = make([]util.Signature, 0, len(signers))
	for _, signer := range signers {
		sig, err := signer.SignHash(datahash[:])
		if err != nil {
			return util.ErrPrefix("TxData.ExSignWith error: ").ErrorIf(err)
		}
		t.ExSignatures = append(t.ExSignatures, sig)
	}
	return nil
}

func (t *Transaction) Signers() (signers util.EthIDs, err error) {
	errp := util.ErrPrefix("Transaction.Signers error: ")

	switch t.Tx.Type {
	case TypeEth:
		if t.eth == nil {
			return nil, errp.Errorf("invalid TypeEth tx")
		}
		signers, err = t.eth.Signers()
	default:
		signers, err = util.DeriveSigners(t.Tx.Bytes(), t.Signatures)
	}

	if err != nil {
		return nil, errp.ErrorIf(err)
	}
	return signers, nil
}

func (t *Transaction) ExSigners() (util.EthIDs, error) {
	errp := util.ErrPrefix("Transaction.ExSigners error: ")

	signers, err := util.DeriveSigners(t.Tx.Data, t.ExSignatures)
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
			if t.Tx.Type == ty {
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
	return x
}

func (t *Transaction) Eth() *TxEth {
	return t.eth
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
		g := set[tx.Tx.From]
		switch {
		case g == nil:
			g = &group{ts: Txs{tx}, dp: tx.priority}
			set[tx.Tx.From] = g
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
			return g.ts[i].Tx.Nonce < g.ts[j].Tx.Nonce
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
