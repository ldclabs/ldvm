// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package ld

import (
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/core/types"

	"github.com/ldclabs/ldvm/util"
)

type TxEth struct {
	tx   *types.Transaction
	from util.EthID
	to   util.EthID
	sigs []util.Signature
	raw  []byte
}

func (t *TxEth) MarshalJSON() ([]byte, error) {
	if t == nil {
		return []byte("null"), nil
	}
	return json.Marshal(t.ToTransaction())
}

// SyntacticVerify verifies that a *TxEth is well-formed.
func (t *TxEth) SyntacticVerify() error {
	errp := util.ErrPrefix("TxEth.SyntacticVerify error: ")

	if t == nil || t.tx == nil {
		return errp.Errorf("nil pointer")
	}

	if chainID := t.tx.ChainId().Uint64(); chainID > 0 && chainID != gChainID {
		return errp.Errorf("invalid chainId, expected %d, got %d", gChainID, chainID)
	}

	if t.tx.Value().Sign() < 0 {
		return errp.Errorf("invalid value")
	}

	from, err := types.Sender(EthSigner, t.tx)
	if err != nil {
		return errp.ErrorIf(err)
	}
	t.from = util.EthID(from)

	to := t.tx.To()
	if to == nil {
		return errp.Errorf("invalid to")
	}
	t.to = util.EthID(*to)
	if t.to == util.EthIDEmpty {
		return errp.Errorf("invalid recipient")
	}

	t.sigs = []util.Signature{encodeSignature(t.tx.RawSignatureValues())}
	if t.raw, err = t.Marshal(); err != nil {
		return errp.ErrorIf(err)
	}
	return nil
}

func (t *TxEth) Bytes() []byte {
	if len(t.raw) == 0 {
		t.raw = MustMarshal(t)
	}
	return t.raw
}

func (t *TxEth) RawSignatureValues() (v, r, s *big.Int) {
	return t.tx.RawSignatureValues()
}

func (t *TxEth) Unmarshal(data []byte) error {
	t.tx = new(types.Transaction)
	return util.ErrPrefix("TxEth.Unmarshal error: ").
		ErrorIf(t.tx.UnmarshalBinary(data))
}

func (t *TxEth) Marshal() ([]byte, error) {
	return util.ErrPrefix("TxEth.Marshal error: ").
		ErrorMap(t.tx.MarshalBinary())
}

func (t *TxEth) TxData(tx *TxData) *TxData {
	if tx == nil {
		tx = new(TxData)
	}
	tx.Type = TypeEth
	tx.ChainID = gChainID
	tx.Nonce = t.tx.Nonce()
	tx.GasTip = 0 // legacy transaction and EIP2718 typed transaction don't have GasTipCap
	tx.GasFeeCap = FromEthBalance(t.tx.GasFeeCap()).Uint64()
	tx.From = t.from
	tx.To = &t.to
	tx.Token = nil
	tx.Amount = FromEthBalance(t.tx.Value())
	tx.Data = t.Bytes()
	tx.Signatures = t.sigs
	tx.ExSignatures = nil
	tx.eth = t
	return tx
}

func (t *TxEth) ToTransaction() *Transaction {
	return t.TxData(nil).ToTransaction()
}

func (t *TxEth) Signers() (util.EthIDs, error) {
	if t.from == util.EthIDEmpty {
		return nil, fmt.Errorf("TxEth.Signers error: invalid signature")
	}
	return util.EthIDs{t.from}, nil
}

func (t *TxEth) Data() []byte {
	return t.tx.Data()
}

func encodeSignature(v, r, s *big.Int) util.Signature {
	sig := util.Signature{}
	if v != nil && r != nil && s != nil {
		copy(sig[:32], r.Bytes())
		copy(sig[32:64], s.Bytes())
		vv := byte(v.Uint64())
		if vv >= 27 {
			vv -= 27
		}
		sig[64] = vv
	}
	return sig
}

func FromEthBalance(amount *big.Int) *big.Int {
	wei := new(big.Int).SetUint64(1e9)
	res := new(big.Int)
	if amount == nil || amount.Cmp(wei) < 0 {
		return res
	}
	return res.Quo(amount, wei)
}

func ToEthBalance(amount *big.Int) *big.Int {
	res := new(big.Int)
	if amount == nil {
		return res
	}
	return res.Mul(amount, new(big.Int).SetUint64(1e9))
}
