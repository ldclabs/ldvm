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
	From      util.EthID
	To        util.EthID
	Value     *big.Int
	Data      []byte
	Signature util.Signature

	// external assignment fields
	tx      *types.Transaction
	signers util.EthIDs
	raw     []byte
}

func (t *TxEth) MarshalJSON() ([]byte, error) {
	if t == nil {
		return []byte("null"), nil
	}
	return json.Marshal(t.ToTransaction())
}

// SyntacticVerify verifies that a *TxEth is well-formed.
func (t *TxEth) SyntacticVerify() error {
	if t == nil {
		return fmt.Errorf("TxEth.SyntacticVerify failed: nil pointer")
	}

	if t.Nonce == 0 {
		return fmt.Errorf("TxEth.SyntacticVerify failed: invalid nonce")
	}
	if t.To == util.EthIDEmpty {
		return fmt.Errorf("TxEth.SyntacticVerify failed: invalid recipient")
	}
	if t.Value == nil || t.Value.Sign() < 1 {
		return fmt.Errorf("TxEth.SyntacticVerify failed: invalid value")
	}
	var err error
	if t.raw, err = t.Marshal(); err != nil {
		return fmt.Errorf("TxEth.SyntacticVerify marshal error: %v", err)
	}
	return nil
}

func (t *TxEth) Bytes() []byte {
	if len(t.raw) == 0 {
		t.raw = MustMarshal(t)
	}
	return t.raw
}

func (t *TxEth) Unmarshal(data []byte) error {
	t.tx = new(types.Transaction)
	if err := t.tx.UnmarshalBinary(data); err != nil {
		return nil
	}

	if chainID := t.tx.ChainId().Uint64(); chainID > 0 && chainID != gChainID {
		return fmt.Errorf("TxEth.Unmarshal failed: invalid chainId, expected %d, got %d", gChainID, chainID)
	}
	t.ChainID = gChainID
	t.Nonce = t.tx.Nonce()
	t.Gas = t.tx.Gas()
	t.GasTipCap = t.tx.GasTipCap().Uint64()
	t.GasFeeCap = t.tx.GasFeeCap().Uint64()
	t.Value = t.tx.Value()
	t.Data = t.tx.Data()
	t.Signature = encodeSignature(t.tx.RawSignatureValues())
	to := t.tx.To()
	if to == nil {
		return fmt.Errorf("TxEth.Unmarshal failed: invalid to")
	}
	t.To = util.EthID(*to)
	signers, err := t.Signers()
	if err != nil {
		return err
	}
	t.From = signers[0]
	return nil
}

func (t *TxEth) Marshal() ([]byte, error) {
	return t.tx.MarshalBinary()
}

func (t *TxEth) ToTransaction() *Transaction {
	return (&TxData{
		Type:       TypeEth,
		ChainID:    t.ChainID,
		Nonce:      t.Nonce,
		GasFeeCap:  t.Gas,
		GasTip:     t.GasTipCap,
		From:       t.From,
		To:         &t.To,
		Amount:     t.Value,
		Data:       t.Bytes(),
		Signatures: []util.Signature{t.Signature},
		eth:        t,
	}).ToTransaction()
}

func (t *TxEth) Signers() (util.EthIDs, error) {
	if len(t.signers) == 0 {
		from, err := types.Sender(EthSigner, t.tx)
		if err != nil {
			return nil, fmt.Errorf("TxEth.Signers failed: %v", err)
		}
		t.signers = util.EthIDs{util.EthID(from)}
	}
	return t.signers, nil
}

func encodeSignature(v, r, s *big.Int) util.Signature {
	sig := util.Signature{}
	if v != nil && r != nil && s != nil {
		copy(sig[:32], r.Bytes())
		copy(sig[32:64], s.Bytes())
		vv := uint8(v.Uint64())
		if vv >= 27 {
			vv -= 27
		}
		sig[64] = vv
	}
	return sig
}
