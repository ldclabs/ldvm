// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package chain

import (
	"fmt"
	"math/big"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ldclabs/ldvm/ld"
)

type TxUpdateDataKeepers struct {
	ld      *ld.Transaction
	from    *Account
	signers []ids.ShortID
	data    *ld.TxUpdater
	dm      *ld.DataMeta
}

func (tx *TxUpdateDataKeepers) MarshalJSON() ([]byte, error) {
	if tx == nil {
		return ld.Null, nil
	}
	v := tx.ld.Copy()
	if tx.data == nil {
		tx.data = &ld.TxUpdater{}
		if err := tx.data.Unmarshal(tx.ld.Data); err != nil {
			return nil, fmt.Errorf("TxUpdateDataKeepers unmarshal data failed: %v", err)
		}
	}
	d, err := tx.data.MarshalJSON()
	if err != nil {
		return nil, err
	}
	v.Data = d
	return v.MarshalJSON()
}

func (tx *TxUpdateDataKeepers) ID() ids.ID {
	return tx.ld.ID()
}

func (tx *TxUpdateDataKeepers) Type() ld.TxType {
	return tx.ld.Type
}

func (tx *TxUpdateDataKeepers) Bytes() []byte {
	return tx.ld.Bytes()
}

func (tx *TxUpdateDataKeepers) SyntacticVerify() error {
	if tx.ld.Gas == 0 ||
		tx.ld.GasFeeCap == 0 ||
		tx.ld.Amount != nil ||
		tx.ld.From == ids.ShortEmpty ||
		tx.ld.To != ids.ShortEmpty ||
		len(tx.ld.Signatures) == 0 ||
		len(tx.ld.ExSignatures) != 0 {
		return fmt.Errorf("invalid TxUpdateDataKeepers")
	}

	var err error
	tx.signers, err = ld.DeriveSigners(tx.ld.UnsignedBytes(), tx.ld.Signatures)
	if err != nil {
		return fmt.Errorf("invalid signatures")
	}

	tx.data = &ld.TxUpdater{}
	if err := tx.data.Unmarshal(tx.ld.Data); err != nil {
		return fmt.Errorf("TxUpdateDataKeepers unmarshal data failed: %v", err)
	}
	if err := tx.data.SyntacticVerify(); err != nil {
		return fmt.Errorf("TxUpdateDataKeepers SyntacticVerify failed: %v", err)
	}
	if tx.data.ID == ids.ShortEmpty ||
		tx.data.Version == 0 {
		return fmt.Errorf("TxUpdateDataKeepers invalid txUpdater")
	}
	if len(tx.data.Keepers) == 0 {
		return fmt.Errorf("TxUpdateDataKeepers no keepers")
	}
	return nil
}

func (tx *TxUpdateDataKeepers) Verify(blk *Block) error {
	var err error
	if tx.from, err = verifyBase(blk, tx.ld, tx.signers); err != nil {
		return err
	}

	tx.dm, err = blk.State().LoadData(tx.data.ID)
	if err != nil {
		return fmt.Errorf("TxUpdateDataKeepers load data failed: %v", err)
	}
	if tx.dm.Version != tx.data.Version {
		return fmt.Errorf("TxUpdateDataKeepers version mismatch, expected %v, got %v",
			tx.dm.Version, tx.data.Version)
	}
	if !ld.SatisfySigning(tx.dm.Threshold, tx.dm.Keepers, tx.signers, false) {
		return fmt.Errorf("need more signatures")
	}
	return nil
}

func (tx *TxUpdateDataKeepers) Accept(blk *Block) error {
	var err error
	tx.dm.Version++
	tx.dm.Threshold = tx.data.Threshold
	tx.dm.Keepers = tx.data.Keepers
	if err = tx.dm.SyntacticVerify(); err != nil {
		return err
	}

	cost := new(big.Int).Mul(tx.ld.BigIntGas(), blk.GasPrice())
	if err = tx.from.SubByNonce(tx.ld.Nonce, cost); err != nil {
		return err
	}
	return blk.State().SaveData(tx.data.ID, tx.dm)
}

func (tx *TxUpdateDataKeepers) Event(ts int64) *Event {
	return nil
}
