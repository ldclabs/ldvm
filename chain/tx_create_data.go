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
)

type TxCreateData struct {
	ld        *ld.Transaction
	from      *Account
	to        *Account
	signers   []ids.ShortID
	exSigners []ids.ShortID
	data      *ld.DataMeta
	name      *app.Name
}

func (tx *TxCreateData) MarshalJSON() ([]byte, error) {
	if tx == nil {
		return ld.Null, nil
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

func (tx *TxCreateData) ID() ids.ID {
	return tx.ld.ID()
}

func (tx *TxCreateData) Type() ld.TxType {
	return tx.ld.Type
}

func (tx *TxCreateData) Bytes() []byte {
	return tx.ld.Bytes()
}

func (tx *TxCreateData) SyntacticVerify() error {
	if tx == nil ||
		len(tx.ld.Data) == 0 {
		return fmt.Errorf("invalid TxCreateData")
	}

	var err error
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
	tx.signers, err = ld.DeriveSigners(tx.ld.UnsignedBytes(), tx.ld.Signatures)
	if err != nil {
		return fmt.Errorf("invalid signatures: %v", err)
	}
	if len(tx.ld.ExSignatures) > 0 {
		tx.exSigners, err = ld.DeriveSigners(tx.ld.UnsignedBytes(), tx.ld.ExSignatures)
		if err != nil {
			return fmt.Errorf("invalid exSignatures: %v", err)
		}
	}

	bs := blk.State()
	if tx.from, err = verifyBase(blk, tx.ld, tx.signers); err != nil {
		return err
	}
	if tx.ld.To != ids.ShortEmpty {
		if tx.to, err = bs.LoadAccount(tx.ld.To); err != nil {
			return err
		}
	}
	switch ld.ModelID(tx.data.ModelID) {
	case constants.RawModelID:
		return nil
	case constants.JsonModelID:
		if !json.Valid(tx.data.Data) {
			return fmt.Errorf("invalid JSON encoding data")
		}
		return nil
	}

	mm, err := bs.LoadModel(tx.data.ModelID)
	if err != nil {
		return fmt.Errorf("TxCreateData load data model failed: %v", err)
	}
	if err := tx.data.Validate(mm.SchemaType()); err != nil {
		return fmt.Errorf("TxCreateData validate error: %v", err)
	}
	if !ld.SatisfySigning(mm.Threshold, mm.Keepers, tx.exSigners, true) {
		return fmt.Errorf("need more model keepers signatures")
	}

	if bs.ChainConfig().IsNameService(tx.data.ModelID) {
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
	tx.from, err = blk.State().LoadAccount(tx.ld.From)
	return err
}

func (tx *TxCreateData) Accept(blk *Block) error {
	var err error
	bs := blk.State()
	amount := new(big.Int)
	if tx.ld.Amount != nil {
		amount.Set(tx.ld.Amount)
	}
	cost := new(big.Int).Mul(tx.ld.BigIntGas(), blk.GasPrice())
	cost = new(big.Int).Add(amount, cost)
	if err = tx.from.SubByNonce(tx.ld.Nonce, cost); err != nil {
		return err
	}

	if tx.to != nil {
		if err = tx.to.Add(amount); err != nil {
			return err
		}
	}

	id := tx.ld.ShortID()
	if tx.name != nil {
		if err = bs.SetName(tx.name.Name, id); err != nil {
			return err
		}
	}
	return bs.SaveData(id, tx.data)
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
	e.Data = tx.data.Data
	return e
}
