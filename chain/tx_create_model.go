// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package chain

import (
	"fmt"
	"math/big"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ldclabs/ldvm/ld"
)

type TxCreateModel struct {
	ld      *ld.Transaction
	from    *Account
	signers []ids.ShortID
	data    *ld.ModelMeta
}

func (tx *TxCreateModel) MarshalJSON() ([]byte, error) {
	if tx == nil {
		return ld.Null, nil
	}
	v := tx.ld.Copy()
	if tx.data == nil {
		tx.data = &ld.ModelMeta{}
		if err := tx.data.Unmarshal(tx.ld.Data); err != nil {
			return nil, fmt.Errorf("TxCreateModel unmarshal data failed: %v", err)
		}
	}
	d, err := tx.data.MarshalJSON()
	if err != nil {
		return nil, err
	}
	v.Data = d
	return v.MarshalJSON()
}

func (tx *TxCreateModel) ID() ids.ID {
	return tx.ld.ID()
}

func (tx *TxCreateModel) Type() ld.TxType {
	return tx.ld.Type
}

func (tx *TxCreateModel) Bytes() []byte {
	return tx.ld.Bytes()
}

func (tx *TxCreateModel) SyntacticVerify() error {
	if tx.ld.Gas == 0 ||
		tx.ld.GasFeeCap == 0 ||
		tx.ld.Amount != nil ||
		tx.ld.From == ids.ShortEmpty ||
		tx.ld.To != ids.ShortEmpty ||
		len(tx.ld.Signatures) == 0 ||
		len(tx.ld.ExSignatures) != 0 {
		return fmt.Errorf("invalid TxCreateModel")
	}

	var err error
	tx.signers, err = ld.DeriveSigners(tx.ld.UnsignedBytes(), tx.ld.Signatures)
	if err != nil {
		return fmt.Errorf("invalid signatures")
	}

	tx.data = &ld.ModelMeta{}
	if err = tx.data.Unmarshal(tx.ld.Data); err != nil {
		return fmt.Errorf("TxCreateModel unmarshal data failed: %v", err)
	}
	if err = tx.data.SyntacticVerify(); err != nil {
		return fmt.Errorf("TxCreateModel SyntacticVerify failed: %v", err)
	}
	return nil
}

func (tx *TxCreateModel) Verify(blk *Block) error {
	var err error
	tx.from, err = verifyBase(blk, tx.ld, tx.signers)
	return err
}

func (tx *TxCreateModel) VerifyGenesis(blk *Block) error {
	var err error
	tx.data = &ld.ModelMeta{}
	if err = tx.data.Unmarshal(tx.ld.Data); err != nil {
		return fmt.Errorf("TxCreateModel unmarshal data failed: %v", err)
	}
	if err = tx.data.SyntacticVerify(); err != nil {
		return fmt.Errorf("TxCreateModel SyntacticVerify failed: %v", err)
	}
	tx.from, err = blk.State().LoadAccount(tx.ld.From)
	return err
}

func (tx *TxCreateModel) Accept(blk *Block) error {
	blk.State().Log().Info("before from: %v\nto: %v", tx.from.Balance(), nil)
	cost := new(big.Int).Mul(tx.ld.BigIntGas(), blk.GasPrice())
	if err := tx.from.SubByNonce(tx.ld.Nonce, cost); err != nil {
		return err
	}
	blk.State().Log().Info("after from: %v\nto: %v", tx.from.Balance(), nil)
	return blk.State().SaveModel(tx.ld.ShortID(), tx.data)
}

func (tx *TxCreateModel) Event(ts int64) *Event {
	e := NewEvent(tx.ld.ShortID(), SrcModel, ActionAdd)
	e.Time = ts
	e.Data = tx.data.Data
	return e
}
