// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package chain

import (
	"encoding/json"
	"fmt"
	"math/big"
	"strconv"

	"github.com/fxamacker/cbor/v2"
	"github.com/ldclabs/ldvm/constants"
	"github.com/ldclabs/ldvm/ld"
	"github.com/ldclabs/ldvm/ld/service"
	"github.com/ldclabs/ldvm/util"
)

type TxCreateData struct {
	TxBase
	exSigners util.EthIDs
	data      *ld.TxUpdater
	dm        *ld.DataMeta
	name      *service.Name
	kSigner   util.EthID
	mSigner   util.EthID
}

func (tx *TxCreateData) MarshalJSON() ([]byte, error) {
	if tx == nil || tx.ld == nil {
		return []byte("null"), nil
	}
	v := tx.ld.Copy()
	if tx.data == nil {
		return nil, fmt.Errorf("MarshalJSON failed: data not exists")
	}
	d, err := json.Marshal(tx.data)
	if err != nil {
		return nil, err
	}
	v.Data = d
	return json.Marshal(v)
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
	if tx.data.ModelID == nil {
		return fmt.Errorf("TxCreateData TxUpdater invalid modelID")
	}
	if tx.data.KSig == nil {
		return fmt.Errorf("TxCreateData TxUpdater invalid kSig")
	}
	tx.kSigner, err = util.DeriveSigner(tx.data.Data, (*tx.data.KSig)[:])
	if err != nil {
		return fmt.Errorf("TxCreateData TxUpdater invalid kSig: %v", err)
	}

	// with model keepers
	if tx.data.To != nil {
		if *tx.data.To != tx.ld.To {
			return fmt.Errorf("TxCreateData invalid recipient")
		}
		if tx.data.Expire < tx.ld.Timestamp {
			return fmt.Errorf("TxCreateData expired")
		}
		if tx.data.Amount == nil || tx.ld.Amount == nil || tx.data.Amount.Cmp(tx.ld.Amount) != 0 {
			return fmt.Errorf("TxCreateData invalid amount")
		}
		if tx.data.MSig == nil {
			return fmt.Errorf("TxCreateData TxUpdater mSig invalid")
		}
		tx.mSigner, err = util.DeriveSigner(tx.data.Data, (*tx.data.MSig)[:])
		if err != nil {
			return fmt.Errorf("TxCreateData invalid mSig: %v", err)
		}
	}

	tx.dm = &ld.DataMeta{
		ModelID:   *tx.data.ModelID,
		Version:   tx.data.Version,
		Threshold: tx.data.Threshold,
		Keepers:   tx.data.Keepers,
		Data:      tx.data.Data,
		KSig:      *tx.data.KSig,
		MSig:      tx.data.MSig,
	}
	return nil
}

// TxCreateData{ID, Version, Threshold, Keepers, Data, KSig} no model keepers
// TxCreateData{ID, Version, To, Amount, Threshold, Keepers, Data, KSig, MSig, Expire} with model keepers
func (tx *TxCreateData) Verify(blk *Block, bs BlockState) error {
	var err error
	if err = tx.TxBase.Verify(blk, bs); err != nil {
		return err
	}

	if tx.ld.Token != constants.NativeToken {
		return fmt.Errorf("invalid token %s, required LDC", tx.ld.Token)
	}
	if tx.kSigner != util.EthIDEmpty && !tx.from.Keepers().Has(tx.kSigner) {
		return fmt.Errorf("TxCreateData invalid kSig, no signer in account keepers")
	}
	switch tx.dm.ModelID {
	case constants.RawModelID:
		return nil
	case constants.CBORModelID:
		if err = cbor.Valid(tx.data.Data); err != nil {
			return fmt.Errorf("TxCreateData invalid CBOR encoding data: %v", err)
		}
		return nil
	case constants.JSONModelID:
		if !json.Valid(tx.data.Data) {
			return fmt.Errorf("TxCreateData invalid JSON encoding data")
		}
		return nil
	}

	mm, err := bs.LoadModel(tx.dm.ModelID)
	if err != nil {
		return fmt.Errorf("TxCreateData load data model failed: %v", err)
	}
	if err = mm.Model().Valid(tx.dm.Data); err != nil {
		return fmt.Errorf("TxCreateData validate error: %v", err)
	}
	if !util.SatisfySigning(mm.Threshold, mm.Keepers, tx.exSigners, true) {
		return fmt.Errorf("TxCreateData need more model keepers signatures")
	}
	if tx.mSigner != util.EthIDEmpty && !mm.Keepers.Has(tx.mSigner) {
		return fmt.Errorf("TxCreateData invalid mSig, no signer in model keepers")
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
		ModelID:   *tx.data.ModelID,
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

	id := util.DataID(tx.ld.ShortID())
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
