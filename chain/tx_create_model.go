// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package chain

import (
	"fmt"
	"math/big"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/snow/choices"
	"github.com/ldclabs/ldvm/constants"
	"github.com/ldclabs/ldvm/ld"
)

type TxCreateModel struct {
	ld          *ld.Transaction
	from        *Account
	genesisAddr *Account
	signers     []ids.ShortID
	data        *ld.ModelMeta
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

func (tx *TxCreateModel) Status() string {
	return tx.ld.Status.String()
}

func (tx *TxCreateModel) SetStatus(s choices.Status) {
	tx.ld.Status = s
}

func (tx *TxCreateModel) SyntacticVerify() error {
	if tx == nil ||
		len(tx.ld.Data) == 0 {
		return fmt.Errorf("invalid TxCreateModel")
	}

	var err error
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
	tx.signers, err = ld.DeriveSigners(tx.ld.UnsignedBytes(), tx.ld.Signatures)
	if err != nil {
		return fmt.Errorf("invalid signatures: %v", err)
	}
	if tx.genesisAddr, err = blk.State().LoadAccount(constants.GenesisAddr); err != nil {
		return err
	}
	tx.from, err = verifyBase(blk, tx.ld, tx.signers)
	return err
}

// VerifyGenesis skipping signature verification
func (tx *TxCreateModel) VerifyGenesis(blk *Block) error {
	var err error
	bs := blk.State()
	if tx.genesisAddr, err = bs.LoadAccount(constants.GenesisAddr); err != nil {
		return err
	}
	tx.from, err = bs.LoadAccount(tx.ld.From)
	return err
}

func (tx *TxCreateModel) Accept(blk *Block) error {
	var err error
	fee := new(big.Int).Mul(tx.ld.BigIntGas(), blk.GasPrice())
	if err = tx.from.SubByNonce(tx.ld.Nonce, fee); err != nil {
		return err
	}
	if err = tx.genesisAddr.Add(fee); err != nil {
		return err
	}
	return blk.State().SaveModel(tx.ld.ShortID(), tx.data)
}

func (tx *TxCreateModel) Event(ts int64) *Event {
	e := NewEvent(tx.ld.ShortID(), SrcModel, ActionAdd)
	e.Time = ts
	return e
}
