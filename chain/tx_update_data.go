// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package chain

import (
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ldclabs/ldvm/constants"
	"github.com/ldclabs/ldvm/ld"
)

type TxUpdateData struct {
	ld        *ld.Transaction
	from      *Account
	signers   []ids.ShortID
	exSigners []ids.ShortID
	data      *ld.TxUpdater
	dm        *ld.DataMeta
}

func (tx *TxUpdateData) MarshalJSON() ([]byte, error) {
	if tx == nil {
		return ld.Null, nil
	}
	v := tx.ld.Copy()
	if tx.data == nil {
		tx.data = &ld.TxUpdater{}
		if err := tx.data.Unmarshal(tx.ld.Data); err != nil {
			return nil, fmt.Errorf("TxUpdateData unmarshal data failed: %v", err)
		}
	}
	d, err := tx.data.MarshalJSON()
	if err != nil {
		return nil, err
	}
	v.Data = d
	return v.MarshalJSON()
}

func (tx *TxUpdateData) ID() ids.ID {
	return tx.ld.ID()
}

func (tx *TxUpdateData) Type() ld.TxType {
	return tx.ld.Type
}

func (tx *TxUpdateData) Bytes() []byte {
	return tx.ld.Bytes()
}

func (tx *TxUpdateData) SyntacticVerify() error {
	if tx.ld.Gas == 0 ||
		tx.ld.GasFeeCap == 0 ||
		tx.ld.Amount != nil ||
		tx.ld.From == ids.ShortEmpty ||
		tx.ld.To != ids.ShortEmpty ||
		len(tx.ld.Signatures) == 0 ||
		len(tx.ld.ExSignatures) != 0 {
		return fmt.Errorf("invalid TxDeleteData")
	}

	var err error
	tx.signers, err = ld.DeriveSigners(tx.ld.UnsignedBytes(), tx.ld.Signatures)
	if err != nil {
		return fmt.Errorf("invalid signatures")
	}
	if len(tx.ld.ExSignatures) > 0 {
		tx.exSigners, err = ld.DeriveSigners(tx.ld.UnsignedBytes(), tx.ld.ExSignatures)
		if err != nil {
			return fmt.Errorf("invalid exSignatures")
		}
	}

	tx.data = &ld.TxUpdater{}
	if err = tx.data.Unmarshal(tx.ld.Data); err != nil {
		return fmt.Errorf("TxUpdateData unmarshal data failed: %v", err)
	}
	if err = tx.data.SyntacticVerify(); err != nil {
		return fmt.Errorf("TxUpdateData SyntacticVerify failed: %v", err)
	}
	if tx.data.ID == ids.ShortEmpty ||
		tx.data.Version == 0 {
		return fmt.Errorf("invalid TxUpdater for TxUpdateData")
	}
	return nil
}

func (tx *TxUpdateData) Verify(blk *Block) error {
	var err error
	bs := blk.State()
	if tx.from, err = verifyBase(blk, tx.ld, tx.signers); err != nil {
		return err
	}

	tx.dm, err = bs.LoadData(tx.data.ID)
	if err != nil {
		return fmt.Errorf("TxUpdateData load data failed: %v", err)
	}
	if tx.dm.Version != tx.data.Version {
		return fmt.Errorf("TxUpdateData version mismatch, expected %v, got %v",
			tx.dm.Version, tx.data.Version)
	}
	if !ld.SatisfySigning(tx.dm.Threshold, tx.dm.Keepers, tx.signers, false) {
		return fmt.Errorf("need more signatures")
	}

	switch ld.ModelID(tx.dm.ModelID) {
	case constants.RawModelID:
	case constants.JsonModelID:
		if !json.Valid(tx.data.Data) {
			return fmt.Errorf("invalid JSON encoding")
		}
		// apply json patch operations
	default:
		mm, err := bs.LoadModel(tx.dm.ModelID)
		if err != nil {
			return fmt.Errorf("TxUpdateData load data model failed: %v", err)
		}

		if !ld.SatisfySigning(mm.Threshold, mm.Keepers, tx.exSigners, true) {
			return fmt.Errorf("need more exSignatures")
		}
		// TODO: apply patch operations
		// tx.data.Validate(mm.SchemaType)
	}
	if bs.ChainConfig().IsNameService(tx.dm.ModelID) {
		// TODO: should not update name
	}
	return nil
}

func (tx *TxUpdateData) Accept(blk *Block) error {
	bs := blk.State()

	var err error
	tx.dm.Version++
	// TODO: apply patch operations
	if err = tx.dm.SyntacticVerify(); err != nil {
		return err
	}

	cost := new(big.Int).Mul(tx.ld.BigIntGas(), blk.GasPrice())
	if err = tx.from.SubByNonce(tx.ld.Nonce, cost); err != nil {
		return err
	}
	return bs.SaveData(tx.data.ID, tx.dm)
}

func (tx *TxUpdateData) Event(ts int64) *Event {
	e := NewEvent(tx.data.ID, SrcData, ActionUpdate)
	e.Time = ts
	return nil
}
