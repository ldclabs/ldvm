// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package chain

import (
	"fmt"
	"math/big"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ldclabs/ldvm/ld"
)

type TxUpdateModelKeepers struct {
	ld      *ld.Transaction
	from    *Account
	signers []ids.ShortID
	data    *ld.TxUpdater
	mm      *ld.ModelMeta
}

func (tx *TxUpdateModelKeepers) MarshalJSON() ([]byte, error) {
	if tx == nil {
		return ld.Null, nil
	}
	v := tx.ld.Copy()
	if tx.data == nil {
		tx.data = &ld.TxUpdater{}
		if err := tx.data.Unmarshal(tx.ld.Data); err != nil {
			return nil, fmt.Errorf("TxUpdateModelKeepers unmarshal data failed: %v", err)
		}
	}
	d, err := tx.data.MarshalJSON()
	if err != nil {
		return nil, err
	}
	v.Data = d
	return v.MarshalJSON()
}

func (tx *TxUpdateModelKeepers) ID() ids.ID {
	return tx.ld.ID()
}

func (tx *TxUpdateModelKeepers) Type() ld.TxType {
	return tx.ld.Type
}

func (tx *TxUpdateModelKeepers) Bytes() []byte {
	return tx.ld.Bytes()
}

func (tx *TxUpdateModelKeepers) SyntacticVerify() error {
	if tx.ld.Gas == 0 ||
		tx.ld.GasFeeCap == 0 ||
		tx.ld.Amount != nil ||
		tx.ld.From == ids.ShortEmpty ||
		tx.ld.To != ids.ShortEmpty ||
		len(tx.ld.Signatures) == 0 ||
		len(tx.ld.ExSignatures) != 0 {
		return fmt.Errorf("invalid TxUpdateModelKeepers")
	}

	var err error
	tx.signers, err = ld.DeriveSigners(tx.ld.UnsignedBytes(), tx.ld.Signatures)
	if err != nil {
		return fmt.Errorf("invalid signatures")
	}

	tx.data = &ld.TxUpdater{}
	if err := tx.data.Unmarshal(tx.ld.Data); err != nil {
		return fmt.Errorf("TxUpdateModelKeepers unmarshal data failed: %v", err)
	}
	if err := tx.data.SyntacticVerify(); err != nil {
		return fmt.Errorf("TxUpdateModelKeepers SyntacticVerify failed: %v", err)
	}
	if tx.data.ID == ids.ShortEmpty {
		return fmt.Errorf("TxUpdateModelKeepers invalid TxUpdater")
	}
	if len(tx.data.Keepers) == 0 {
		return fmt.Errorf("TxUpdateModelKeepers no keepers")
	}
	return nil
}

func (tx *TxUpdateModelKeepers) Verify(blk *Block) error {
	var err error
	if tx.from, err = verifyBase(blk, tx.ld, tx.signers); err != nil {
		return err
	}

	tx.mm, err = blk.State().LoadModel(tx.data.ID)
	if err != nil {
		return fmt.Errorf("TxUpdateModelKeepers load model failed: %v", err)
	}

	if !ld.SatisfySigning(tx.mm.Threshold, tx.mm.Keepers, tx.signers, false) {
		return fmt.Errorf("TxUpdateModelKeepers need more signatures")
	}
	return nil
}

func (tx *TxUpdateModelKeepers) Accept(blk *Block) error {
	var err error
	tx.mm.Threshold = tx.data.Threshold
	tx.mm.Keepers = tx.data.Keepers
	if err = tx.mm.SyntacticVerify(); err != nil {
		return err
	}

	cost := new(big.Int).Mul(tx.ld.BigIntGas(), blk.GasPrice())
	if err = tx.from.SubByNonce(tx.ld.Nonce, cost); err != nil {
		return err
	}
	return blk.State().SaveModel(tx.data.ID, tx.mm)
}

func (tx *TxUpdateModelKeepers) Event(ts int64) *Event {
	return nil
}
