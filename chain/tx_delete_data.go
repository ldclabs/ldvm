// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package chain

import (
	"fmt"
	"math/big"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ldclabs/ldvm/ld"
)

type TxDeleteData struct {
	ld      *ld.Transaction
	from    *Account
	signers []ids.ShortID
	data    *ld.TxUpdater
	dm      *ld.DataMeta
}

func (tx *TxDeleteData) MarshalJSON() ([]byte, error) {
	if tx == nil {
		return ld.Null, nil
	}
	v := tx.ld.Copy()
	if tx.data == nil {
		tx.data = &ld.TxUpdater{}
		if err := tx.data.Unmarshal(tx.ld.Data); err != nil {
			return nil, fmt.Errorf("TxDeleteData unmarshal data failed: %v", err)
		}
	}
	d, err := tx.data.MarshalJSON()
	if err != nil {
		return nil, err
	}
	v.Data = d
	return v.MarshalJSON()
}

func (tx *TxDeleteData) ID() ids.ID {
	return tx.ld.ID()
}

func (tx *TxDeleteData) Type() ld.TxType {
	return tx.ld.Type
}

func (tx *TxDeleteData) Bytes() []byte {
	return tx.ld.Bytes()
}

func (tx *TxDeleteData) SyntacticVerify() error {
	if tx == nil ||
		len(tx.ld.Data) == 0 {
		return fmt.Errorf("invalid TxDeleteData")
	}

	var err error
	tx.data = &ld.TxUpdater{}
	if err = tx.data.Unmarshal(tx.ld.Data); err != nil {
		return fmt.Errorf("TxDeleteData unmarshal data failed: %v", err)
	}
	if err = tx.data.SyntacticVerify(); err != nil {
		return fmt.Errorf("TxDeleteData SyntacticVerify failed: %v", err)
	}
	if tx.data.ID == ids.ShortEmpty ||
		tx.data.Version == 0 {
		return fmt.Errorf("invalid TxUpdater for TxDeleteData")
	}
	return nil
}

func (tx *TxDeleteData) Verify(blk *Block) error {
	var err error
	tx.signers, err = ld.DeriveSigners(tx.ld.UnsignedBytes(), tx.ld.Signatures)
	if err != nil {
		return fmt.Errorf("invalid signatures")
	}

	if tx.from, err = verifyBase(blk, tx.ld, tx.signers); err != nil {
		return err
	}

	tx.dm, err = blk.State().LoadData(tx.data.ID)
	if err != nil {
		return fmt.Errorf("TxDeleteData load data failed: %v", err)
	}
	if tx.dm.Version != tx.data.Version {
		return fmt.Errorf("TxDeleteData version mismatch, expected %v, got %v",
			tx.dm.Version, tx.data.Version)
	}
	if !ld.SatisfySigning(tx.dm.Threshold, tx.dm.Keepers, tx.signers, false) {
		return fmt.Errorf("need more signatures")
	}
	return nil
}

func (tx *TxDeleteData) Accept(blk *Block) error {
	tx.dm.Version = 0 // mark dropped
	var err error
	if err = tx.dm.SyntacticVerify(); err != nil {
		return err
	}
	cost := new(big.Int).Mul(tx.ld.BigIntGas(), blk.GasPrice())
	if err = tx.from.SubByNonce(tx.ld.Nonce, cost); err != nil {
		return err
	}
	return blk.State().SaveData(tx.data.ID, tx.dm)
}

func (tx *TxDeleteData) Event(ts int64) *Event {
	e := NewEvent(tx.data.ID, SrcData, ActionDelete)
	e.Time = ts
	return e
}
