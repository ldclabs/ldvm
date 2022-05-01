// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package chain

import (
	"encoding/json"
	"fmt"
	"math/big"
	"strconv"
	"time"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ldclabs/ldvm/constants"
	"github.com/ldclabs/ldvm/ld"
	"github.com/ldclabs/ldvm/ld/service"
	"github.com/ldclabs/ldvm/util"
)

type TxCreateData struct {
	*TxBase
	exSigners []ids.ShortID
	data      *ld.TxUpdater
	dm        *ld.DataMeta
	name      *service.Name
	kSigner   ids.ShortID
	sSigner   ids.ShortID
}

func (tx *TxCreateData) MarshalJSON() ([]byte, error) {
	if tx == nil || tx.ld == nil {
		return util.Null, nil
	}
	v := tx.ld.Copy()
	if tx.data == nil {
		tx.data = &ld.TxUpdater{}
		if err := tx.data.Unmarshal(tx.ld.Data); err != nil {
			return nil, fmt.Errorf("TxCreateData unmarshal failed: %v", err)
		}
	}

	d, err := tx.data.MarshalJSON()
	if err != nil {
		return nil, err
	}
	v.Data = d
	return v.MarshalJSON()
}

func (tx *TxCreateData) SyntacticVerify() error {
	var err error
	if err = tx.TxBase.SyntacticVerify(); err != nil {
		return err
	}

	if len(tx.ld.Data) == 0 {
		return fmt.Errorf("TxCreateData invalid")
	}

	if len(tx.ld.ExSignatures) > 0 {
		tx.exSigners, err = util.DeriveSigners(tx.ld.Data, tx.ld.ExSignatures)
		if err != nil {
			return fmt.Errorf("TxCreateData invalid exSignatures: %v", err)
		}
	}

	tx.data = &ld.TxUpdater{}
	if err = tx.data.Unmarshal(tx.ld.Data); err != nil {
		return fmt.Errorf("TxCreateData unmarshal data failed: %v", err)
	}
	if err = tx.data.SyntacticVerify(); err != nil {
		return fmt.Errorf("TxCreateData SyntacticVerify failed: %v", err)
	}
	if tx.data.Version != 1 {
		return fmt.Errorf("TxCreateData version must be 1")
	}
	if len(tx.data.Keepers) == 0 {
		return fmt.Errorf("TxCreateData should have at least one keeper")
	}
	if tx.data.KSig == nil {
		return fmt.Errorf("TxCreateData TxUpdater kSig invalid")
	}
	tx.kSigner, err = util.DeriveSigner(tx.data.Data, (*tx.data.KSig)[:])
	if err != nil {
		return fmt.Errorf("TxCreateData invalid kSig: %v", err)
	}

	// with model keepers
	if tx.data.To != ids.ShortEmpty {
		if tx.data.To != tx.ld.To {
			return fmt.Errorf("TxCreateData invalid recipient")
		}
		if tx.data.Expire < uint64(time.Now().Unix()) {
			return fmt.Errorf("TxCreateData expired")
		}
		if tx.data.Amount == nil || tx.ld.Amount == nil || tx.data.Amount.Cmp(tx.ld.Amount) != 0 {
			return fmt.Errorf("TxCreateData invalid amount")
		}
		if tx.data.SSig == nil {
			return fmt.Errorf("TxCreateData TxUpdater sSig invalid")
		}
		tx.sSigner, err = util.DeriveSigner(tx.data.Data, (*tx.data.SSig)[:])
		if err != nil {
			return fmt.Errorf("TxCreateData invalid sSig: %v", err)
		}
	}

	tx.dm = &ld.DataMeta{
		ModelID:   tx.data.ID,
		Version:   tx.data.Version,
		Threshold: tx.data.Threshold,
		Keepers:   tx.data.Keepers,
		Data:      tx.data.Data,
		KSig:      *tx.data.KSig,
		SSig:      tx.data.SSig,
	}
	return nil
}

// TxCreateData{ID, Version, Threshold, Keepers, Data, KSig} no model keepers
// TxCreateData{ID, Version, To, Amount, Threshold, Keepers, Data, KSig, SSig, Expire} with model keepers

func (tx *TxCreateData) Verify(blk *Block, bs BlockState) error {
	var err error
	if err = tx.TxBase.Verify(blk, bs); err != nil {
		return err
	}

	if tx.ld.Token != constants.LDCAccount {
		return fmt.Errorf("invalid token %s, required LDC", util.EthID(tx.ld.Token))
	}
	if tx.kSigner != ids.ShortEmpty && !util.ShortIDs(tx.from.Keepers()).Has(tx.kSigner) {
		return fmt.Errorf("TxCreateData invalid kSig, no signer in account keepers")
	}
	switch util.ModelID(tx.dm.ModelID) {
	case constants.RawModelID:
		return nil
	case constants.JsonModelID:
		if !json.Valid(tx.data.Data) {
			return fmt.Errorf("TxCreateData invalid JSON encoding data")
		}
		return nil
	}

	mm, err := bs.LoadModel(tx.dm.ModelID)
	if err != nil {
		return fmt.Errorf("TxCreateData load data model failed: %v", err)
	}
	if err = tx.dm.Validate(mm.SchemaType()); err != nil {
		return fmt.Errorf("TxCreateData validate error: %v", err)
	}
	if !util.SatisfySigning(mm.Threshold, mm.Keepers, tx.exSigners, true) {
		return fmt.Errorf("TxCreateData need more model keepers signatures")
	}
	if tx.sSigner != ids.ShortEmpty && !util.ShortIDs(mm.Keepers).Has(tx.sSigner) {
		return fmt.Errorf("TxCreateData invalid sSig, no signer in model keepers")
	}

	if blk.ctx.Chain().IsNameService(tx.dm.ModelID) {
		tx.name = &service.Name{}
		if err = tx.name.Unmarshal(tx.dm.Data); err != nil {
			return err
		}
		_, err = bs.ResolveNameID(tx.name.Name)
		if err == nil {
			return fmt.Errorf("TxCreateData validate error: name %s conflict",
				strconv.Quote(tx.name.Name))
		}
	}
	return nil
}

// VerifyGenesis skipping signature verification
func (tx *TxCreateData) VerifyGenesis(blk *Block, bs BlockState) error {
	var err error
	tx.data = &ld.TxUpdater{}
	if err = tx.data.Unmarshal(tx.ld.Data); err != nil {
		return fmt.Errorf("TxCreateData unmarshal data failed: %v", err)
	}
	if err = tx.data.SyntacticVerify(); err != nil {
		return fmt.Errorf("TxCreateData SyntacticVerify failed: %v", err)
	}
	tx.dm = &ld.DataMeta{
		ModelID:   tx.data.ID,
		Version:   tx.data.Version,
		Threshold: tx.data.Threshold,
		Keepers:   tx.data.Keepers,
		Data:      tx.data.Data,
	}

	tx.tip = new(big.Int)
	tx.fee = new(big.Int)
	tx.cost = new(big.Int)

	if tx.ldc, err = bs.LoadAccount(constants.LDCAccount); err != nil {
		return err
	}
	if tx.miner, err = blk.Miner(); err != nil {
		return err
	}
	tx.from, err = bs.LoadAccount(tx.ld.From)
	return err
}

func (tx *TxCreateData) Accept(blk *Block, bs BlockState) error {
	var err error

	id := tx.ld.ShortID()
	if tx.name != nil {
		if err = bs.SetName(tx.name.Name, id); err != nil {
			return err
		}
	}
	if err = bs.SaveData(id, tx.dm); err != nil {
		return err
	}
	return tx.TxBase.Accept(blk, bs)
}

func (tx *TxCreateData) Event(ts int64) *Event {
	var e *Event
	if tx.name != nil {
		e = NewEvent(tx.ld.ShortID(), SrcName, ActionAdd)
		e.Subject = tx.name.Name
	} else {
		e = NewEvent(tx.ld.ShortID(), SrcData, ActionAdd)
	}
	e.Time = ts
	return e
}
