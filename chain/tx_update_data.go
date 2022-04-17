// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package chain

import (
	"fmt"
	"math/big"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/snow/choices"
	jsonpatch "github.com/evanphx/json-patch/v5"

	"github.com/ldclabs/ldvm/constants"
	"github.com/ldclabs/ldvm/ld"
)

type TxUpdateData struct {
	ld          *ld.Transaction
	from        *Account
	genesisAddr *Account
	signers     []ids.ShortID
	exSigners   []ids.ShortID
	data        *ld.TxUpdater
	dm          *ld.DataMeta
	prevDM      *ld.DataMeta
	jsonPatch   jsonpatch.Patch
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

func (tx *TxUpdateData) Status() string {
	return tx.ld.Status.String()
}

func (tx *TxUpdateData) SetStatus(s choices.Status) {
	tx.ld.Status = s
}

func (tx *TxUpdateData) SyntacticVerify() error {
	if tx == nil ||
		len(tx.ld.Data) == 0 {
		return fmt.Errorf("invalid TxUpdateData")
	}

	var err error
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
	if tx.from, err = verifyBase(blk, tx.ld, tx.signers); err != nil {
		return err
	}

	bs := blk.State()
	if tx.genesisAddr, err = bs.LoadAccount(constants.GenesisAddr); err != nil {
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

	tx.prevDM = tx.dm.Copy()
	switch ld.ModelID(tx.dm.ModelID) {
	case constants.RawModelID:
		tx.dm.Version++
		tx.dm.Data = tx.data.Data
		return nil
	case constants.JsonModelID:
		tx.jsonPatch, err = jsonpatch.DecodePatch(tx.data.Data)
		if err != nil {
			return fmt.Errorf("invalid JSON patch: %v", err)
		}

		tx.dm.Data, err = tx.jsonPatch.Apply(tx.dm.Data)
		if err != nil {
			return fmt.Errorf("apply patch failed: %v", err)
		}
		tx.dm.Version++
		return tx.dm.SyntacticVerify()
	}

	mm, err := bs.LoadModel(tx.dm.ModelID)
	if err != nil {
		return fmt.Errorf("TxUpdateData load data model failed: %v", err)
	}

	if !ld.SatisfySigning(mm.Threshold, mm.Keepers, tx.exSigners, true) {
		return fmt.Errorf("need more exSignatures")
	}
	// TODO: apply patch operations
	// tx.data.Validate(mm.SchemaType)
	if blk.ctx.Chain().IsNameApp(tx.dm.ModelID) {
		// TODO: should not update name
	}
	return nil
}

func (tx *TxUpdateData) Accept(blk *Block) error {
	var err error
	bs := blk.State()
	fee := new(big.Int).Mul(tx.ld.BigIntGas(), blk.GasPrice())
	if err = tx.from.SubByNonce(tx.ld.Nonce, fee); err != nil {
		return err
	}
	if err = tx.genesisAddr.Add(fee); err != nil {
		return err
	}
	if err = bs.SavePrevData(tx.data.ID, tx.prevDM); err != nil {
		return err
	}
	return bs.SaveData(tx.data.ID, tx.dm)
}

func (tx *TxUpdateData) Event(ts int64) *Event {
	e := NewEvent(tx.data.ID, SrcData, ActionUpdate)
	e.Time = ts
	return nil
}
