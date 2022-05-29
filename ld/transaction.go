// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package ld

import (
	"fmt"
	"math"
	"math/big"
	"sort"
	"strconv"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ldclabs/ldvm/constants"
	"github.com/ldclabs/ldvm/util"
)

const (
	// The "test" transaction tests that a value of data at the target location
	// is equal to a specified value. test transaction will not write to the block.
	// It should be in a batch transactions.
	TypeTest TxType = iota

	// punish transaction can be issued by genesisAccount
	// we can only punish illegal data
	TypePunish

	// Transfer
	TypeEth          // send given amount of NanoLDC to a address in ETH transaction
	TypeTransfer     // send given amount of NanoLDC to a address
	TypeTransferPay  // send given amount of NanoLDC to the address who request payment
	TypeTransferCash // cash given amount of NanoLDC to sender, like cash a check.
	TypeExchange     // exchange tokens

	// Account
	TypeAddNonceTable        // add more nonce with expire time to account
	TypeUpdateAccountKeepers // update account's Keepers and Threshold
	TypeCreateToken          // create a token account
	TypeDestroyToken         // destroy a token account
	TypeCreateStake          // create a stake account
	TypeResetStake           // reset or destroy a stake account
	TypeTakeStake            // take a stake in
	TypeWithdrawStake        // withdraw stake
	TypeUpdateStakeApprover
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
)

const (
	// gasTipPerSec: A delay of 1 seconds is equivalent to 1000 gasTip
	GasTipPerSec = constants.MicroLDC
)

// gChainID will be updated by SetChainID when VM.Initialize
var gChainID = uint64(2357)

// TxType is an uint8 representing the type of the tx
type TxType uint8

func (t TxType) Gas() uint64 {
	switch t {
	case TypeTest:
		return 0
	case TypePunish:
		return 42
	case TypeEth, TypeTransfer, TypeTransferPay, TypeTransferCash,
		TypeExchange, TypeAddNonceTable:
		return 42
	case TypeUpdateAccountKeepers, TypeCreateToken,
		TypeDestroyToken, TypeCreateStake, TypeResetStake:
		return 1000
	case TypeTakeStake, TypeWithdrawStake, TypeUpdateStakeApprover:
		return 500
	case TypeOpenLending, TypeCloseLending:
		return 1000
	case TypeBorrow, TypeRepay:
		return 500
	case TypeCreateModel, TypeUpdateModelKeepers:
		return 500
	case TypeCreateData, TypeUpdateData, TypeUpdateDataKeepers:
		return 100
	case TypeUpdateDataKeepersByAuth, TypeDeleteData:
		return 200
	default:
		return 1000
	}
}

func (t TxType) String() string {
	switch t {
	case TypeTest:
		return "TestTx"
	case TypePunish:
		return "PunishTx"
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
	case TypeCreateToken:
		return "CreateTokenTx"
	case TypeDestroyToken:
		return "DestroyTokenTx"
	case TypeCreateStake:
		return "CreateStakeTx"
	case TypeResetStake:
		return "ResetStakeTx"
	case TypeTakeStake:
		return "TakeStakeTx"
	case TypeWithdrawStake:
		return "WithdrawStakeTx"
	case TypeUpdateStakeApprover:
		return "TypeUpdateStakeApprover"
	case TypeOpenLending:
		return "OpenLendingTx"
	case TypeCloseLending:
		return "CloseLendingTx"
	case TypeBorrow:
		return "BorrowTx"
	case TypeRepay:
		return "RepayTx"
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
	default:
		return "UnknownTx"
	}
}

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
	Data         RawData           `cbor:"d,omitempty" json:"data,omitempty"`
	Signatures   []util.Signature  `cbor:"ss,omitempty" json:"signatures,omitempty"`
	ExSignatures []util.Signature  `cbor:"es,omitempty" json:"exSignatures,omitempty"`

	// external assignment fields
	ID       ids.ID `cbor:"-" json:"-"`
	raw      []byte `cbor:"-" json:"-"`
	unsigned []byte `cbor:"-" json:"-"`
	eth      *TxEth `cbor:"-" json:"-"`
}

// SyntacticVerify verifies that a *Tx is well-formed.
func (t *TxData) SyntacticVerify() error {
	if t == nil {
		return fmt.Errorf("TxData.SyntacticVerify failed: nil pointer")
	}
	if t.Type > TypeDeleteData {
		return fmt.Errorf("TxData.SyntacticVerify failed: invalid type")
	}
	if t.ChainID != gChainID {
		return fmt.Errorf("TxData.SyntacticVerify failed: invalid ChainID, expected %d, got %d",
			gChainID, t.ChainID)
	}
	if t.Token != nil && !t.Token.Valid() {
		return fmt.Errorf("TxData.SyntacticVerify failed: invalid token symbol %s",
			strconv.Quote(t.Token.GoString()))
	}
	if t.Amount != nil {
		if t.Amount.Sign() <= 0 {
			return fmt.Errorf("TxData.SyntacticVerify failed: invalid amount")
		}
		if t.To == nil {
			return fmt.Errorf("TxData.SyntacticVerify failed: invalid to")
		}
	}
	if t.Data != nil && len(t.Data) == 0 {
		return fmt.Errorf("TxData.SyntacticVerify failed: empty data")
	}
	if t.Signatures != nil && len(t.Signatures) == 0 {
		return fmt.Errorf("TxData.SyntacticVerify failed: empty signatures")
	}
	if t.ExSignatures != nil && len(t.ExSignatures) == 0 {
		return fmt.Errorf("TxData.SyntacticVerify failed: empty exSignatures")
	}
	if len(t.Signatures) > math.MaxUint8 {
		return fmt.Errorf("TxData.SyntacticVerify failed: too many signatures")
	}
	if len(t.ExSignatures) > math.MaxUint8 {
		return fmt.Errorf("TxData.SyntacticVerify failed: too many exSignatures")
	}
	var err error
	if t.raw, err = t.Marshal(); err != nil {
		return fmt.Errorf("TxData.SyntacticVerify marshal error: %v", err)
	}
	t.ID = util.IDFromData(t.Bytes())
	return nil
}

func (t *TxData) Bytes() []byte {
	if len(t.raw) == 0 {
		t.raw = MustMarshal(t)
	}
	return t.raw
}

func (t *TxData) Unmarshal(data []byte) error {
	return DecMode.Unmarshal(data, t)
}

func (t *TxData) Marshal() ([]byte, error) {
	return EncMode.Marshal(t)
}

func (t *TxData) UnsignedBytes() []byte {
	if len(t.unsigned) == 0 {
		sigs := t.Signatures
		exSigs := t.ExSignatures
		t.Signatures = nil
		t.ExSignatures = nil
		t.unsigned = MustMarshal(t)
		t.Signatures = sigs
		t.ExSignatures = exSigs
	}
	return t.unsigned
}

func (t *TxData) RequiredGas(threshold uint64) uint64 {
	gas := uint64(len(t.UnsignedBytes())) + t.Type.Gas()
	if gas <= threshold {
		return gas
	}
	return threshold + uint64(math.Pow(float64(gas-threshold), math.SqrtPhi))
}

func (t *TxData) SignWith(signers ...Signer) error {
	data := t.UnsignedBytes()
	for _, signer := range signers {
		sig, err := signer.Sign(data)
		if err != nil {
			return err
		}
		t.Signatures = append(t.Signatures, sig)
	}
	return nil
}

func (t *TxData) ExSignWith(signers ...Signer) error {
	for _, signer := range signers {
		sig, err := signer.Sign(t.Data)
		if err != nil {
			return err
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
	Data         RawData           `cbor:"d,omitempty" json:"data,omitempty"`
	Signatures   []util.Signature  `cbor:"ss,omitempty" json:"signatures,omitempty"`
	ExSignatures []util.Signature  `cbor:"es,omitempty" json:"exSignatures,omitempty"`

	// external assignment fields
	Gas       uint64  `cbor:"g" json:"gas"` // calculate when build block.
	Name      string  `cbor:"-" json:"name"`
	ID        ids.ID  `cbor:"id" json:"id"`
	Err       error   `cbor:"-" json:"error,omitempty"`
	AddedTime uint64  `cbor:"-" json:"-"`
	Height    uint64  `cbor:"-" json:"-"` // block's timestamp
	Timestamp uint64  `cbor:"-" json:"-"` // block's timestamp
	priority  uint64  `cbor:"-" json:"-"`
	tx        *TxData `cbor:"-" json:"-"`
	gas       uint64  `cbor:"-" json:"-"`
	raw       []byte  `cbor:"-" json:"-"` // the transaction's raw bytes, included id and sigs.
	// support for batch transactions
	// they are processed in the same block, one fail all fail
	batch Txs `cbor:"-" json:"-"`
}

// SyntacticVerify verifies that a *Transaction is well-formed.
func (t *Transaction) SyntacticVerify() error {
	if t == nil {
		return fmt.Errorf("Transaction.SyntacticVerify failed: nil pointer")
	}

	t.tx = t.TxData(t.tx)
	var err error
	if err = t.tx.SyntacticVerify(); err != nil {
		return err
	}

	t.ID = t.tx.ID
	t.Name = t.Type.String()
	t.gas = t.Gas
	if t.raw, err = t.Marshal(); err != nil {
		return fmt.Errorf("Transaction.SyntacticVerify marshal error: %v", err)
	}
	return nil
}

func (t *Transaction) Bytes() []byte {
	if len(t.raw) == 0 || t.gas != t.Gas {
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
	case t.Type != TypeTest:
		total = len(t.Bytes())
	}
	return total
}

func (t *Transaction) UnmarshalTx(data []byte) error {
	tx := new(TxData)
	if err := tx.Unmarshal(data); err != nil {
		return err
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
	return DecMode.Unmarshal(data, t)
}

func (t *Transaction) Marshal() ([]byte, error) {
	return EncMode.Marshal(t)
}

func (t *Transaction) SetPriority(threshold, delay uint64) {
	switch {
	case t.Type == TypeTest:
		t.priority = 0
	case t.IsBatched():
		mx := uint64(0)
		for _, tx := range t.batch {
			tx.SetPriority(threshold, delay)
			if mx < tx.priority {
				mx = tx.priority
			}
		}
		t.priority = mx
	default:
		// tip fee as priority
		priority := t.GasTip * threshold
		gas := t.RequiredGas(threshold)
		if v := t.GasTip * gas; v > priority {
			priority = v
		}
		// Consider gossip network overhead, ignoring small time differences
		// not promote priority if not processed more than 120 seconds(tx maybe invalid...)
		if delay > 3 && delay <= 120 {
			priority += delay * GasTipPerSec * threshold
		}
		t.priority = priority
	}
}

func (t *Transaction) RequiredGas(threshold uint64) uint64 {
	return t.tx.RequiredGas(threshold)
}

func (t *Transaction) GasUnits() *big.Int {
	return new(big.Int).SetUint64(t.Gas)
}

func (t *Transaction) Signers() (util.EthIDs, error) {
	switch t.Type {
	case TypeEth:
		if t.tx.eth == nil {
			return nil, fmt.Errorf("Transaction.Signers invalid TypeEth tx")
		}
		return t.tx.eth.Signers()
	}
	return util.DeriveSigners(t.tx.UnsignedBytes(), t.Signatures)
}

func (t *Transaction) ExSigners() (util.EthIDs, error) {
	return util.DeriveSigners(t.Data, t.ExSignatures)
}

func (t *Transaction) NeedApprove(approver *util.EthID, approveList []TxType) bool {
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

func NewBatchTx(txs ...*Transaction) (*Transaction, error) {
	if len(txs) <= 1 {
		return nil, fmt.Errorf("NewBatchTx: not batch transactions")
	}

	maxSize := -1
	var err error
	var tx *Transaction
	for i, t := range txs {
		if t.Type == TypeTest {
			continue
		}
		if err = t.SyntacticVerify(); err != nil {
			return nil, err
		}
		if size := len(t.tx.UnsignedBytes()); size > maxSize {
			tx = txs[i] // find the max UnsignedBytes tx as batch transactions' container
			maxSize = size
		}
	}
	if maxSize == -1 {
		return nil, fmt.Errorf("NewBatchTx: no invalid transaction")
	}
	tx = tx.Copy()
	if err = tx.SyntacticVerify(); err != nil {
		return nil, err
	}

	tx.batch = txs
	return tx, nil
}

type Txs []*Transaction

func (txs *Txs) Unmarshal(data []byte) error {
	return DecMode.Unmarshal(data, txs)
}

func (txs Txs) Marshal() ([]byte, error) {
	return EncMode.Marshal(txs)
}

type group struct {
	ts Txs
	ps uint64
}

type groupSet map[util.EthID]*group

func (set groupSet) Add(txs ...*Transaction) {
	for i := range txs {
		tx := txs[i]
		g := set[tx.From]
		switch {
		case tx.Type == TypeTest:
			// nothing
		case g == nil:
			g = &group{ts: Txs{tx}, ps: tx.priority}
			set[tx.From] = g
		default:
			// txs from the same sender share the priority
			g.ps += tx.priority
			g.ts = append(g.ts, tx)
		}
	}
}

func (txs Txs) SortWith(threshold, nowSeconds uint64) {
	// first: group by sender - tx.From
	set := make(groupSet, len(txs))
	for i := range txs {
		tx := txs[i]
		delay := uint64(0)
		if nowSeconds > tx.AddedTime {
			delay = nowSeconds - tx.AddedTime
		}
		tx.SetPriority(threshold, delay)
		if tx.IsBatched() {
			set.Add(tx.Txs()...)
		} else {
			set.Add(tx)
		}
	}
	// then: rebalance the priority for the same sender, small nonce tx get larger priority.
	for _, g := range set {
		n := uint64(len(g.ts))
		g.ps = n + g.ps/n
		sort.SliceStable(g.ts, func(i, j int) bool {
			return g.ts[i].Nonce < g.ts[j].Nonce
		})
		for i, tx := range g.ts {
			tx.priority = g.ps - uint64(i)
		}
	}
	// then: pick the max priority for batch txs again.
	for _, tx := range txs {
		if tx.IsBatched() {
			tx.priority = 0
			for _, ti := range tx.Txs() {
				if tx.priority < ti.priority {
					tx.priority = ti.priority
				}
			}
		}
	}
	// last: all txs sort by priority
	sort.SliceStable(txs, func(i, j int) bool {
		return txs[i].priority > txs[j].priority
	})
}
