// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package ld

import (
	"fmt"
	"math"
	"math/big"
	"sort"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ldclabs/ldvm/constants"
	"github.com/ldclabs/ldvm/util"
	"github.com/ldclabs/ldvm/util/signer"
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

// TxData represents a complete transaction issued from client
type TxData struct {
	Type      TxType            `cbor:"t" json:"type"`
	ChainID   uint64            `cbor:"c" json:"chainID"`
	Nonce     uint64            `cbor:"n" json:"nonce"`
	GasTip    uint64            `cbor:"gt" json:"gasTip"`
	GasFeeCap uint64            `cbor:"gf" json:"gasFeeCap"`
	From      util.Address      `cbor:"fr" json:"from"`                   // Address of the sender
	To        *util.Address     `cbor:"to,omitempty" json:"to,omitempty"` // Address of the recipient
	Token     *util.TokenSymbol `cbor:"tk,omitempty" json:"token,omitempty"`
	Amount    *big.Int          `cbor:"a,omitempty" json:"amount,omitempty"`
	Data      util.RawData      `cbor:"d,omitempty" json:"data,omitempty"`

	// external assignment fields
	raw []byte `cbor:"-" json:"-"`
}

// SyntacticVerify verifies that a *TxData is well-formed.
func (t *TxData) SyntacticVerify() error {
	errp := util.ErrPrefix("ld.TxData.SyntacticVerify: ")

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
	return util.ErrPrefix("ld.TxData.Unmarshal: ").
		ErrorIf(util.UnmarshalCBOR(data, t))
}

func (t *TxData) Marshal() ([]byte, error) {
	return util.ErrPrefix("ld.TxData.Marshal: ").
		ErrorMap(util.MarshalCBOR(t))
}

func (t *TxData) ToTransaction() *Transaction {
	return &Transaction{Tx: *t}
}

type Transaction struct {
	Tx           TxData      `cbor:"tx" json:"tx"`
	Signatures   signer.Sigs `cbor:"ss,omitempty" json:"sigs,omitempty"`
	ExSignatures signer.Sigs `cbor:"es,omitempty" json:"exSigs,omitempty"`

	// external assignment fields
	ID        util.Hash `cbor:"-" json:"id"`
	Err       error     `cbor:"-" json:"error,omitempty"`
	Height    uint64    `cbor:"-" json:"-"` // block's timestamp
	Timestamp uint64    `cbor:"-" json:"-"` // block's timestamp
	priority  uint64    `cbor:"-" json:"-"`
	dp        uint64    `cbor:"-" json:"-"` // dynamic priority for sorting
	raw       []byte    `cbor:"-" json:"-"` // the transaction's raw bytes, included id and sigs.
	gas       uint64    `cbor:"-" json:"-"`
	eth       *TxEth    `cbor:"-" json:"-"`
	txHash    []byte    `cbor:"-" json:"-"`
	exHash    []byte    `cbor:"-" json:"-"`
	// support for batch transactions
	// they are processed in the same block, one fail all fail
	batch Txs `cbor:"-" json:"-"`
}

// SyntacticVerify verifies that a *Transaction is well-formed.
func (t *Transaction) SyntacticVerify() error {
	errp := util.ErrPrefix("ld.Transaction.SyntacticVerify: ")
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
	if err = t.Signatures.Valid(); err != nil {
		return errp.ErrorIf(err)
	}
	if err = t.ExSignatures.Valid(); err != nil {
		return errp.ErrorIf(err)
	}

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

	t.txHash = util.Sum256(t.Tx.Bytes())
	t.ID = util.HashFromData(t.raw)
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

func (t *Transaction) TxHash() []byte {
	if len(t.txHash) == 0 {
		t.txHash = util.Sum256(t.Tx.Bytes())
	}
	return t.txHash
}

func (t *Transaction) ExHash() []byte {
	if len(t.exHash) == 0 {
		t.exHash = util.Sum256(t.Tx.Data)
	}
	return t.exHash
}

func (t *Transaction) Unmarshal(data []byte) error {
	return util.ErrPrefix("ld.Transaction.Unmarshal: ").
		ErrorIf(util.UnmarshalCBOR(data, t))
}

func (t *Transaction) Marshal() ([]byte, error) {
	return util.ErrPrefix("ld.Transaction.Marshal: ").
		ErrorMap(util.MarshalCBOR(t))
}

func (t *Transaction) SignWith(signers ...signer.Signer) error {
	datahash := t.TxHash()
	t.Signatures = make([]signer.Sig, 0, len(signers))
	for _, s := range signers {
		sig, err := s.SignHash(datahash)
		if err != nil {
			return util.ErrPrefix("ld.Transaction.SignWith: ").ErrorIf(err)
		}
		t.Signatures = append(t.Signatures, sig)
	}
	return nil
}

func (t *Transaction) ExSignWith(signers ...signer.Signer) error {
	datahash := t.ExHash()
	t.ExSignatures = make([]signer.Sig, 0, len(signers))
	for _, s := range signers {
		sig, err := s.SignHash(datahash)
		if err != nil {
			return util.ErrPrefix("ld.Transaction.ExSignWith: ").ErrorIf(err)
		}
		t.ExSignatures = append(t.ExSignatures, sig)
	}
	return nil
}

type TxIsApprovedFn func(signer.Key, TxTypes, bool) bool

func (t *Transaction) IsApproved(
	approver signer.Key, approveList TxTypes, byEx bool) bool {
	if t.needApprove(approver, approveList) {
		if byEx {
			return approver.Verify(t.exHash, t.ExSignatures)
		}

		return approver.Verify(t.txHash, t.Signatures)
	}

	return true
}

func (t *Transaction) needApprove(approver signer.Key, approveList TxTypes) bool {
	switch {
	case len(approver) == 0:
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

// Copy is not a deep copy, used for json.Marshaling
func (t *Transaction) Copy() *Transaction {
	x := new(Transaction)
	*x = *t
	x.Tx.raw = nil
	x.raw = nil
	return x
}

func (t *Transaction) Eth() *TxEth {
	return t.eth
}

func NewBatchTx(txs ...*Transaction) (*Transaction, error) {
	errp := util.ErrPrefix("ld.NewBatchTx: ")

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
	return util.ErrPrefix("ld.Txs.Unmarshal: ").
		ErrorIf(util.UnmarshalCBOR(data, txs))
}

func (txs Txs) Marshal() ([]byte, error) {
	return util.ErrPrefix("ld.Txs.Marshal: ").
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

func (t *Transaction) Txs() Txs {
	switch {
	case t.IsBatched():
		return t.batch
	default:
		return Txs{t}
	}
}

func (t *Transaction) Size() int {
	switch {
	case t.IsBatched():
		return len(t.batch)
	default:
		return 1
	}
}

func (txs Txs) Size() int {
	s := 0
	for _, tx := range txs {
		s += tx.Size()
	}
	return s
}

func (t *Transaction) IDs() util.IDList[util.Hash] {
	switch {
	case t.IsBatched():
		return t.batch.IDs()
	default:
		return util.IDList[util.Hash]{t.ID}
	}
}

func (txs Txs) IDs() util.IDList[util.Hash] {
	txIDs := util.NewIDList[util.Hash](txs.Size())
	for _, tx := range txs {
		switch {
		case tx.IsBatched():
			txIDs = append(txIDs, tx.batch.IDs()...)
		default:
			txIDs = append(txIDs, tx.ID)
		}
	}

	return txIDs
}

type group struct {
	ts Txs
	dp uint64 // dynamic priority for sorting
}

type groupSet map[util.Address]*group

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
