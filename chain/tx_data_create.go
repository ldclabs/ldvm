// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package chain

import (
	"encoding/json"
	"fmt"
	"math/big"
	"strconv"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ldclabs/ldvm/constants"
	"github.com/ldclabs/ldvm/ld"
	"github.com/ldclabs/ldvm/ld/app"
	"github.com/ldclabs/ldvm/util"
)

type TxCreateData struct {
	*TxBase
	exSigners []ids.ShortID
	data      *ld.DataMeta
	name      *app.Name
}

func (tx *TxCreateData) MarshalJSON() ([]byte, error) {
	if tx == nil || tx.ld == nil {
		return util.Null, nil
	}
	v := tx.ld.Copy()
	if tx.data == nil {
		tx.data = &ld.DataMeta{}
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
		tx.exSigners, err = util.DeriveSigners(tx.ld.UnsignedBytes(), tx.ld.ExSignatures)
		if err != nil {
			return fmt.Errorf("TxCreateData invalid exSignatures: %v", err)
		}
	}

	tx.data = &ld.DataMeta{}
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
	return nil
}

func (tx *TxCreateData) Verify(blk *Block) error {
	var err error
	if err = tx.TxBase.Verify(blk); err != nil {
		return err
	}

	if tx.ld.Token != constants.LDCAccount {
		return fmt.Errorf("invalid token %s, required LDC", util.EthID(tx.ld.Token))
	}
	switch util.ModelID(tx.data.ModelID) {
	case constants.RawModelID:
		return nil
	case constants.JsonModelID:
		if !json.Valid(tx.data.Data) {
			return fmt.Errorf("TxCreateData invalid JSON encoding data")
		}
		return nil
	}

	bs := blk.State()
	mm, err := bs.LoadModel(tx.data.ModelID)
	if err != nil {
		return fmt.Errorf("TxCreateData load data model failed: %v", err)
	}
	if err := tx.data.Validate(mm.SchemaType()); err != nil {
		return fmt.Errorf("TxCreateData validate error: %v", err)
	}
	if !util.SatisfySigning(mm.Threshold, mm.Keepers, tx.exSigners, true) {
		return fmt.Errorf("TxCreateData need more model keepers signatures")
	}

	if blk.ctx.Chain().IsNameApp(tx.data.ModelID) {
		bn, err := app.NameFrom(tx.data.Data)
		if err != nil {
			return err
		}
		_, err = bs.ResolveNameID(bn.Entity.Name)
		if err == nil {
			return fmt.Errorf("TxCreateData validate error: name %s conflict",
				strconv.Quote(bn.Entity.Name))
		}
		tx.name = bn.Entity
	}
	return nil
}

// VerifyGenesis skipping signature verification
func (tx *TxCreateData) VerifyGenesis(blk *Block) error {
	var err error
	tx.tip = new(big.Int)
	tx.fee = new(big.Int)
	tx.cost = new(big.Int)

	bs := blk.State()
	if tx.ldc, err = bs.LoadAccount(constants.LDCAccount); err != nil {
		return err
	}
	if tx.miner, err = blk.Miner(); err != nil {
		return err
	}
	tx.from, err = bs.LoadAccount(tx.ld.From)
	return err
}

func (tx *TxCreateData) Accept(blk *Block) error {
	var err error

	bs := blk.State()
	id := tx.ld.ShortID()
	if tx.name != nil {
		if err = bs.SetName(tx.name.Name, id); err != nil {
			return err
		}
	}
	if err = bs.SaveData(id, tx.data); err != nil {
		return err
	}
	return tx.TxBase.Accept(blk)
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
