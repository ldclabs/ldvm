// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package chain

import (
	"fmt"
	"math/big"
	"time"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ldclabs/ldvm/ld"
)

type TxTransferReply struct {
	ld        *ld.Transaction
	from      *Account
	to        *Account
	signers   []ids.ShortID
	exSigners []ids.ShortID
	data      *ld.TxTransfer
}

func (tx *TxTransferReply) MarshalJSON() ([]byte, error) {
	if tx == nil {
		return ld.Null, nil
	}
	v := tx.ld.Copy()
	if tx.data == nil {
		tx.data = &ld.TxTransfer{}
		if err := tx.data.Unmarshal(tx.ld.Data); err != nil {
			return nil, fmt.Errorf("TxTransferReply unmarshal data failed: %v", err)
		}
	}
	d, err := tx.data.MarshalJSON()
	if err != nil {
		return nil, err
	}
	v.Data = d
	return v.MarshalJSON()
}

func (tx *TxTransferReply) ID() ids.ID {
	return tx.ld.ID()
}

func (tx *TxTransferReply) Type() ld.TxType {
	return tx.ld.Type
}

func (tx *TxTransferReply) Bytes() []byte {
	return tx.ld.Bytes()
}

func (tx *TxTransferReply) SyntacticVerify() error {
	if tx.ld.Gas == 0 ||
		tx.ld.GasFeeCap == 0 ||
		tx.ld.Amount == nil ||
		tx.ld.From == ids.ShortEmpty ||
		tx.ld.To == ids.ShortEmpty ||
		len(tx.ld.Signatures) == 0 ||
		len(tx.ld.ExSignatures) == 0 {
		return fmt.Errorf("invalid TxTransferReply")
	}

	var err error
	tx.signers, err = ld.DeriveSigners(tx.ld.UnsignedBytes(), tx.ld.Signatures)
	if err != nil {
		return fmt.Errorf("invalid signatures")
	}
	tx.exSigners, err = ld.DeriveSigners(tx.ld.Data, tx.ld.ExSignatures)
	if err != nil {
		return fmt.Errorf("invalid exSignatures")
	}
	set := ids.NewShortSet(len(tx.exSigners))
	set.Add(tx.exSigners...)
	if !set.Contains(tx.ld.To) {
		return fmt.Errorf("invalid recipient")
	}

	tx.data = &ld.TxTransfer{}
	if err = tx.data.Unmarshal(tx.ld.Data); err != nil {
		return fmt.Errorf("TxTransferReply unmarshal data failed: %v", err)
	}
	if err = tx.data.SyntacticVerify(); err != nil {
		return fmt.Errorf("TxTransferReply SyntacticVerify failed: %v", err)
	}
	if tx.data.To != tx.ld.To {
		return fmt.Errorf("TxTransferReply invalid recipient")
	}
	if tx.data.Expire > 0 && tx.data.Expire < uint64(time.Now().Unix()) {
		return fmt.Errorf("TxTransferReply expired")
	}
	if tx.data.Amount != nil && tx.data.Amount.Cmp(tx.ld.Amount) != 0 {
		return fmt.Errorf("TxTransferReply invalid amount")
	}
	return nil
}

func (tx *TxTransferReply) Verify(blk *Block) error {
	var err error
	tx.from, err = verifyBase(blk, tx.ld, tx.signers)
	if err != nil {
		return err
	}
	tx.to, err = blk.State().LoadAccount(tx.ld.To)
	return err
}

func (tx *TxTransferReply) Accept(blk *Block) error {
	var err error
	cost := new(big.Int).Mul(tx.ld.BigIntGas(), blk.GasPrice())
	cost = new(big.Int).Add(tx.ld.Amount, cost)
	if err = tx.from.SubByNonce(tx.ld.Nonce, cost); err != nil {
		return err
	}
	if err = tx.to.Add(tx.ld.Amount); err != nil {
		return err
	}
	return nil
}

func (tx *TxTransferReply) Event(ts int64) *Event {
	return nil
}
