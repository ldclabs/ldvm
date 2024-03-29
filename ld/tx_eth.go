// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package ld

import (
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/core/types"

	"github.com/ldclabs/ldvm/ids"
	"github.com/ldclabs/ldvm/signer"
	"github.com/ldclabs/ldvm/util/erring"
)

type TxEth struct {
	tx   *types.Transaction
	from ids.Address
	to   ids.Address
	sig  signer.Sig
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
	errp := erring.ErrPrefix("ld.TxEth.SyntacticVerify: ")

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
	t.from = ids.Address(from)

	to := t.tx.To()
	if to == nil {
		return errp.Errorf("invalid to")
	}
	t.to = ids.Address(*to)
	if t.to == ids.EmptyAddress {
		return errp.Errorf("invalid recipient")
	}

	t.sig = encodeSignature(t.tx.RawSignatureValues())
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
	return erring.ErrPrefix("ld.TxEth.Unmarshal: ").
		ErrorIf(t.tx.UnmarshalBinary(data))
}

func (t *TxEth) Marshal() ([]byte, error) {
	return erring.ErrPrefix("ld.TxEth.Marshal: ").
		ErrorMap(t.tx.MarshalBinary())
}

func (t *TxEth) ToTransaction() *Transaction {
	tx := &Transaction{
		Tx: TxData{
			Type:      TypeEth,
			ChainID:   gChainID,
			Nonce:     t.tx.Nonce(),
			GasTip:    0, // legacy transaction and EIP2718 typed transaction don't have GasTipCap
			GasFeeCap: FromEthBalance(t.tx.GasFeeCap()).Uint64(),
			From:      t.from,
			To:        t.to.Ptr(),
			Amount:    FromEthBalance(t.tx.Value()),
			Data:      t.Bytes(),
		},
		Signatures: signer.Sigs{t.sig},
	}

	tx.eth = t
	return tx
}

func (t *TxEth) Signers() (signer.Keys, error) {
	if t.from == ids.EmptyAddress {
		return nil, fmt.Errorf("ld.TxEth.Signers: invalid signature")
	}
	return signer.Keys{signer.Key(t.from[:])}, nil
}

func (t *TxEth) Data() []byte {
	return t.tx.Data()
}

func encodeSignature(v, r, s *big.Int) signer.Sig {
	sig := make([]byte, 65)
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
