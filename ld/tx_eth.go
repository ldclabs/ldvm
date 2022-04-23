// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package ld

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ldclabs/ldvm/util"
)

var EthSigner = types.NewLondonSigner(big.NewInt(2357))

// SetChainID will be set when VM.Initialize
func SetChainID(id uint64) {
	gChainID = id
	EthSigner = types.NewLondonSigner(big.NewInt(int64(id)))
}

type TxEth struct {
	ChainID   uint64
	Nonce     uint64
	GasTipCap uint64
	GasFeeCap uint64
	Gas       uint64
	From      ids.ShortID
	To        ids.ShortID
	Value     *big.Int
	Data      []byte
	Signature util.Signature

	// external assignment
	tx  *types.Transaction
	raw []byte
}

type jsonTxEth struct {
	ChainID   uint64          `json:"chainID,omitempty"`
	Nonce     uint64          `json:"nonce,omitempty"`
	GasTipCap uint64          `json:"gasTipCap,omitempty"`
	GasFeeCap uint64          `json:"gasPrice,omitempty"`
	Gas       uint64          `json:"gasLimit,omitempty"`
	From      string          `json:"from,omitempty"`
	To        string          `json:"to,omitempty"`
	Value     *big.Int        `json:"value,omitempty"`
	Data      json.RawMessage `json:"data,omitempty"`
}

func (t *TxEth) MarshalJSON() ([]byte, error) {
	if t == nil {
		return util.Null, nil
	}
	v := &jsonTxEth{
		ChainID:   t.ChainID,
		Nonce:     t.Nonce,
		GasTipCap: t.GasTipCap,
		GasFeeCap: t.GasFeeCap,
		Gas:       t.Gas,
		Value:     t.Value,
		Data:      util.JSONMarshalData(t.Data),
	}

	if t.From != ids.ShortEmpty {
		v.From = util.EthID(t.From).String()
	}
	if t.To != ids.ShortEmpty {
		v.To = util.EthID(t.To).String()
	}
	return json.Marshal(v)
}

// SyntacticVerify verifies that a *TxEth is well-formed.
func (t *TxEth) SyntacticVerify() error {
	if t == nil {
		return fmt.Errorf("invalid TxEth")
	}

	if t.Nonce == 0 {
		return fmt.Errorf("invalid nonce")
	}
	if t.To == ids.ShortEmpty {
		return fmt.Errorf("invalid recipient")
	}
	if t.Value == nil || t.Value.Sign() < 1 {
		return fmt.Errorf("invalid value")
	}
	t.tx.WithSignature(EthSigner, t.Signature[:])
	from, err := types.Sender(EthSigner, t.tx)
	if err != nil {
		return fmt.Errorf("invalid signature: %v", err)
	}
	t.From = ids.ShortID(from)
	if _, err := t.Marshal(); err != nil {
		return fmt.Errorf("TxEth marshal error: %v", err)
	}
	return nil
}

func (t *TxEth) Equal(o *TxEth) bool {
	if o == nil {
		return false
	}
	if len(o.raw) > 0 && len(t.raw) > 0 {
		return bytes.Equal(o.raw, t.raw)
	}
	if o.ChainID != t.ChainID {
		return false
	}
	if o.Nonce != t.Nonce {
		return false
	}
	if o.GasTipCap != t.GasTipCap {
		return false
	}
	if o.GasFeeCap != t.GasFeeCap {
		return false
	}
	if o.Gas != t.Gas {
		return false
	}
	if o.From != t.From {
		return false
	}
	if o.To != t.To {
		return false
	}
	if o.Value == nil || t.Value == nil {
		if o.Value != t.Value {
			return false
		}
	}
	if o.Value.Cmp(t.Value) != 0 {
		return false
	}
	return bytes.Equal(o.Data, t.Data)
}

func (t *TxEth) Bytes() []byte {
	if len(t.raw) == 0 {
		if _, err := t.Marshal(); err != nil {
			panic(err)
		}
	}
	return t.raw
}

func (t *TxEth) ToTransaction() *Transaction {
	return &Transaction{
		Type:       TypeEth,
		ChainID:    t.ChainID,
		Nonce:      t.Nonce,
		GasFeeCap:  t.GasFeeCap,
		GasTip:     t.GasTipCap,
		From:       t.From,
		To:         t.To,
		Amount:     t.Value,
		Data:       t.Bytes(),
		Signatures: []util.Signature{t.Signature},
	}
}

func (t *TxEth) Unmarshal(data []byte) error {
	t.tx = new(types.Transaction)
	t.raw = data
	if err := t.tx.UnmarshalBinary(data); err != nil {
		return nil
	}

	if chainID := t.tx.ChainId().Uint64(); chainID > 0 && chainID != gChainID {
		return fmt.Errorf("invalid EthTx chainId, expected %d, got %d", gChainID, chainID)
	}
	t.ChainID = gChainID

	to := t.tx.To()
	if to == nil {
		return fmt.Errorf("invalid EthTx to")
	}
	t.To = ids.ShortID(*to)
	t.Nonce = t.tx.Nonce()
	t.Gas = t.tx.Gas()
	t.GasTipCap = t.tx.GasTipCap().Uint64()
	t.GasFeeCap = t.tx.GasFeeCap().Uint64()
	t.Value = t.tx.Value()
	t.Data = t.tx.Data()
	return nil
}

func (t *TxEth) Marshal() ([]byte, error) {
	var td types.TxData
	var buf bytes.Buffer

	if t.tx == nil {
		return nil, fmt.Errorf("invalid TxEth, inner tx missing")
	}
	switch t.tx.Type() {
	case types.LegacyTxType:
		td = &types.LegacyTx{
			Nonce:    t.tx.Nonce(),
			GasPrice: t.tx.GasPrice(),
			Gas:      t.tx.Gas(),
			To:       t.tx.To(),
			Value:    t.tx.Value(),
			Data:     t.tx.Data(),
		}
	case types.AccessListTxType:
		td = &types.AccessListTx{
			ChainID:    t.tx.ChainId(),
			Nonce:      t.tx.Nonce(),
			GasPrice:   t.tx.GasPrice(),
			Gas:        t.tx.Gas(),
			To:         t.tx.To(),
			Value:      t.tx.Value(),
			Data:       t.tx.Data(),
			AccessList: t.tx.AccessList(),
		}
		buf.WriteByte(t.tx.Type())
	case types.DynamicFeeTxType:
		td = &types.DynamicFeeTx{
			ChainID:    t.tx.ChainId(),
			Nonce:      t.tx.Nonce(),
			GasTipCap:  t.tx.GasTipCap(),
			GasFeeCap:  t.tx.GasFeeCap(),
			Gas:        t.tx.Gas(),
			To:         t.tx.To(),
			Value:      t.tx.Value(),
			Data:       t.tx.Data(),
			AccessList: t.tx.AccessList(),
		}
		buf.WriteByte(t.tx.Type())
	default:
		return nil, fmt.Errorf("TxEth Marshal error: invalid txType %d", t.tx.Type())
	}

	if err := rlp.Encode(&buf, td); err != nil {
		return nil, err
	}
	t.raw = buf.Bytes()
	return t.raw, nil
}

func TxEthFromSigned(data []byte) (*TxEth, error) {
	t := &TxEth{tx: new(types.Transaction)}
	if err := t.tx.UnmarshalBinary(data); err != nil {
		return nil, err
	}

	if chainID := t.tx.ChainId().Uint64(); chainID > 0 && chainID != gChainID {
		return nil, fmt.Errorf("invalid EthTx chainId, expected %d, got %d", gChainID, chainID)
	}
	t.ChainID = gChainID

	to := t.tx.To()
	if to == nil {
		return nil, fmt.Errorf("invalid EthTx to")
	}
	t.To = ids.ShortID(*to)
	from, err := types.Sender(EthSigner, t.tx)
	if err != nil {
		return nil, fmt.Errorf("invalid EthTx signature: %v", err)
	}
	t.From = ids.ShortID(from)
	t.Signature = encodeSignature(t.tx.RawSignatureValues())
	t.Nonce = t.tx.Nonce()
	t.Gas = t.tx.Gas()
	t.GasTipCap = t.tx.GasTipCap().Uint64()
	t.GasFeeCap = t.tx.GasFeeCap().Uint64()
	t.Value = t.tx.Value()
	t.Data = t.tx.Data()

	if _, err = t.Marshal(); err != nil {
		return nil, err
	}
	return t, nil
}

// func decodeSignature(sig []byte) (r, s, v *big.Int) {
// 	if len(sig) != crypto.SignatureLength {
// 		panic(fmt.Sprintf("wrong size for signature: got %d, want %d", len(sig), crypto.SignatureLength))
// 	}
// 	r = new(big.Int).SetBytes(sig[:32])
// 	s = new(big.Int).SetBytes(sig[32:64])
// 	v = new(big.Int).SetBytes([]byte{sig[64] + 27})
// 	return r, s, v
// }

func encodeSignature(r, s, v *big.Int) util.Signature {
	sig := util.Signature{}
	copy(sig[:32], r.Bytes())
	copy(sig[32:64], s.Bytes())
	vv := uint8(v.Uint64())
	if vv >= 27 {
		vv -= 27
	}
	sig[64] = vv
	return sig
}