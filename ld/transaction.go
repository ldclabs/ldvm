// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package ld

import (
	"bytes"
	"fmt"
	"math"
	"math/big"

	"github.com/ava-labs/avalanchego/ids"
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
	TypeEthGas          = uint64(42)
	TypeTransferGas     = uint64(42)
	TypeTransferPayGas  = uint64(42)
	TypeTransferCashGas = uint64(42)
	TypeExchangeGas     = uint64(42)

	TypeAddNonceTableGas        = uint64(100)
	TypeUpdateAccountKeepersGas = uint64(1000)
	TypeCreateTokenAccountGas   = uint64(1000)
	TypeDestroyTokenAccountGas  = uint64(1000)
	TypeCreateStakeAccountGas   = uint64(1000)
	TypeResetStakeAccountGas    = uint64(1000)
	TypeTakeStakeGas            = uint64(500)
	TypeWithdrawStakeGas        = uint64(500)
	TypeOpenLendingGas          = uint64(1000)
	TypeCloseLendingGas         = uint64(1000)
	TypeBorrowGas               = uint64(100)
	TypeRepayGas                = uint64(100)

	TypeCreateModelGas        = uint64(500)
	TypeUpdateModelKeepersGas = uint64(500)

	TypeCreateDataGas              = uint64(100)
	TypeUpdateDataGas              = uint64(100)
	TypeUpdateDataKeepersGas       = uint64(100)
	TypeUpdateDataKeepersByAuthGas = uint64(200)
	TypeDeleteDataGas              = uint64(200)

	TypePunishGas = uint64(200)

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
	case TypeResetStakeAccount:
		return "ResetStakeAccountTx"
	case TypeTakeStake:
		return "TakeStakeTx"
	case TypeWithdrawStake:
		return "WithdrawStakeTx"
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
	case TypePunish:
		return "PunishTx"
	default:
		return "UnknownTx"
	}
}

type Transaction struct {
	Type         TxType           `cbor:"t" json:"type"`
	ChainID      uint64           `cbor:"c" json:"chainID"`
	Nonce        uint64           `cbor:"n" json:"nonce"`
	Gas          uint64           `cbor:"g" json:"gas"` // calculate when build block.
	GasTip       uint64           `cbor:"gt" json:"gasTip"`
	GasFeeCap    uint64           `cbor:"gf" json:"gasFeeCap"`
	Token        util.TokenSymbol `cbor:"tk" json:"token,omitempty"`
	From         util.EthID       `cbor:"fr" json:"from"`
	To           util.EthID       `cbor:"to" json:"to"`
	Amount       *big.Int         `cbor:"a" json:"amount"`
	Data         RawData          `cbor:"d" json:"data"`
	Signatures   []util.Signature `cbor:"ss" json:"signatures"`
	ExSignatures []util.Signature `cbor:"es" json:"exSignatures"`

	// external assignment
	ID          ids.ID `cbor:"id" json:"id"`
	Name        string `cbor:"-" json:"name"`
	Err         error  `cbor:"-" json:"error,omitempty"`
	gas         uint64 `cbor:"-" json:"-"`
	unsignedRaw []byte `cbor:"-" json:"-"` // raw bytes for sign
	raw         []byte `cbor:"-" json:"-"` // the transaction's raw bytes, included id and sigs.
	AddTime     uint64 `cbor:"-" json:"-"`
	Priority    uint64 `cbor:"-" json:"-"`
	Height      uint64 `cbor:"-" json:"-"` // block's timestamp
	Timestamp   uint64 `cbor:"-" json:"-"` // block's timestamp
	// support for batch transactions
	// they are processed in the same block, one fail all fail
	batch Txs `cbor:"-" json:"-"`
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

type Txs []*Transaction

func MarshalTxs(txs []*Transaction) ([]byte, error) {
	if len(txs) == 0 {
		return nil, nil
	}
	return EncMode.Marshal(txs)
}

func UnmarshalTxs(data []byte) ([]*Transaction, error) {
	if len(data) == 0 {
		return nil, nil
	}
	txs := make([]*Transaction, 0)
	if err := DecMode.Unmarshal(data, &txs); err != nil {
		return nil, err
	}
	return txs, nil
}

func (t *Transaction) IsBatched() bool {
	return len(t.batch) > 0
}

func (t *Transaction) Txs() []*Transaction {
	return t.batch
}

func (t *Transaction) ShortID() ids.ShortID {
	sid := ids.ShortID{}
	copy(sid[:], t.ID[:])
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
	if t.Token != constants.NativeToken && t.Token.String() == "" {
		return fmt.Errorf("invalid token symbol: %s", t.Token)
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
	t.Name = TxTypeString(t.Type)
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
		return requireGas(threshold, gas+TypeEthGas)
	case TypeTransfer:
		return requireGas(threshold, gas+TypeTransferGas)
	case TypeTransferPay:
		return requireGas(threshold, gas+TypeTransferPayGas)
	case TypeTransferCash:
		return requireGas(threshold, gas+TypeTransferCashGas)
	case TypeExchange:
		return requireGas(threshold, gas+TypeExchangeGas)

	case TypeAddNonceTable:
		return requireGas(threshold, gas+TypeAddNonceTableGas)
	case TypeUpdateAccountKeepers:
		return requireGas(threshold, gas+TypeUpdateAccountKeepersGas)
	case TypeCreateTokenAccount:
		return requireGas(threshold, gas+TypeCreateTokenAccountGas)
	case TypeDestroyTokenAccount:
		return requireGas(threshold, gas+TypeDestroyTokenAccountGas)
	case TypeCreateStakeAccount:
		return requireGas(threshold, gas+TypeCreateStakeAccountGas)
	case TypeResetStakeAccount:
		return requireGas(threshold, gas+TypeResetStakeAccountGas)
	case TypeTakeStake:
		return requireGas(threshold, gas+TypeTakeStakeGas)
	case TypeWithdrawStake:
		return requireGas(threshold, gas+TypeWithdrawStakeGas)
	case TypeOpenLending:
		return requireGas(threshold, gas+TypeOpenLendingGas)
	case TypeCloseLending:
		return requireGas(threshold, gas+TypeCloseLendingGas)
	case TypeBorrow:
		return requireGas(threshold, gas+TypeBorrowGas)
	case TypeRepay:
		return requireGas(threshold, gas+TypeRepayGas)
	case TypeCreateModel:
		return requireGas(threshold, gas+TypeCreateModelGas)
	case TypeUpdateModelKeepers:
		return requireGas(threshold, gas+TypeUpdateModelKeepersGas)

	case TypeCreateData:
		return requireGas(threshold, gas+TypeCreateDataGas)
	case TypeUpdateData:
		return requireGas(threshold, gas+TypeUpdateDataGas)
	case TypeUpdateDataKeepers:
		return requireGas(threshold, gas+TypeUpdateDataKeepersGas)
	case TypeUpdateDataKeepersByAuth:
		return requireGas(threshold, gas+TypeUpdateDataKeepersByAuthGas)
	case TypeDeleteData:
		return requireGas(threshold, gas+TypeDeleteDataGas)

	case TypePunish:
		return requireGas(threshold, gas+TypePunishGas)

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
		MustMarshal(t)
	}
	return t.raw
}

func (t *Transaction) calcID() (ids.ID, error) {
	if t.ID == ids.Empty {
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
		t.ID = util.IDFromBytes(b)
	}
	return t.ID, nil
}

func (t *Transaction) calcUnsignedBytes() ([]byte, error) {
	if len(t.unsignedRaw) == 0 {
		id := t.ID
		gas := t.Gas
		raw := t.raw
		sigs := t.Signatures
		exSigs := t.ExSignatures
		// clear gas, id, Signatures, ExSignatures
		t.Gas = 0
		t.ID = ids.Empty
		t.Signatures = nil
		t.ExSignatures = nil
		b, err := t.Marshal()
		t.ID = id
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
	if err := DecMode.Unmarshal(data, t); err != nil {
		return err
	}
	t.gas = t.Gas
	t.raw = data
	return nil
}

func (t *Transaction) Marshal() ([]byte, error) {
	if t.IsBatched() {
		return nil, fmt.Errorf("can not marshal batch transactions")
	}

	data, err := EncMode.Marshal(t)
	if err != nil {
		return nil, err
	}
	t.gas = t.Gas
	t.raw = data
	return data, nil
}
