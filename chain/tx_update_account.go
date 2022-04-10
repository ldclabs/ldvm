// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package chain

import (
	"fmt"
	"math/big"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ldclabs/ldvm/ld"
)

type TxUpdateAccountKeepers struct {
	ld      *ld.Transaction
	from    *Account
	signers []ids.ShortID
	data    *ld.TxUpdater
}

func (tx *TxUpdateAccountKeepers) MarshalJSON() ([]byte, error) {
	if tx == nil {
		return ld.Null, nil
	}
	v := tx.ld.Copy()
	if tx.data == nil {
		tx.data = &ld.TxUpdater{}
		if err := tx.data.Unmarshal(tx.ld.Data); err != nil {
			return nil, fmt.Errorf("TxUpdateAccountKeepers unmarshal data failed: %v", err)
		}
	}
	d, err := tx.data.MarshalJSON()
	if err != nil {
		return nil, err
	}
	v.Data = d
	return v.MarshalJSON()
}

func (tx *TxUpdateAccountKeepers) ID() ids.ID {
	return tx.ld.ID()
}

func (tx *TxUpdateAccountKeepers) Type() ld.TxType {
	return tx.ld.Type
}

func (tx *TxUpdateAccountKeepers) Bytes() []byte {
	return tx.ld.Bytes()
}

func (tx *TxUpdateAccountKeepers) SyntacticVerify() error {
	if tx.ld.Nonce == 0 ||
		tx.ld.Gas == 0 ||
		tx.ld.GasFeeCap == 0 ||
		tx.ld.Amount != nil ||
		tx.ld.From == ids.ShortEmpty ||
		tx.ld.To != ids.ShortEmpty ||
		len(tx.ld.Data) == 0 ||
		len(tx.ld.Signatures) == 0 ||
		len(tx.ld.ExSignatures) != 0 {
		return fmt.Errorf("invalid TxUpdateAccountKeepers")
	}

	var err error
	tx.signers, err = ld.DeriveSigners(tx.ld.UnsignedBytes(), tx.ld.Signatures)
	if err != nil {
		return fmt.Errorf("invalid signatures")
	}

	tx.data = &ld.TxUpdater{}
	if err = tx.data.Unmarshal(tx.ld.Data); err != nil {
		return fmt.Errorf("TxUpdateAccountKeepers unmarshal data failed: %v", err)
	}
	if err = tx.data.SyntacticVerify(); err != nil {
		return fmt.Errorf("TxUpdateAccountKeepers SyntacticVerify failed: %v", err)
	}
	if len(tx.data.Keepers) == 0 ||
		tx.data.Threshold == 0 {
		return fmt.Errorf("TxUpdateAccountKeepers invalid TxUpdater")
	}
	return nil
}

func (tx *TxUpdateAccountKeepers) Verify(blk *Block) error {
	var err error
	if tx.from, err = verifyBase(blk, tx.ld, tx.signers); err != nil {
		return err
	}
	return nil
}

func (tx *TxUpdateAccountKeepers) VerifyGenesis(blk *Block) error {
	var err error
	tx.data = &ld.TxUpdater{}
	if err = tx.data.Unmarshal(tx.ld.Data); err != nil {
		return fmt.Errorf("TxUpdateAccountKeepers unmarshal data failed: %v", err)
	}
	if err = tx.data.SyntacticVerify(); err != nil {
		return fmt.Errorf("TxUpdateAccountKeepers SyntacticVerify failed: %v", err)
	}
	if len(tx.data.Keepers) == 0 ||
		tx.data.Threshold == 0 {
		return fmt.Errorf("TxUpdateAccountKeepers invalid TxUpdater")
	}
	tx.from, err = blk.State().LoadAccount(tx.ld.From)
	return err
}

func (tx *TxUpdateAccountKeepers) Accept(blk *Block) error {
	cost := new(big.Int).Mul(tx.ld.BigIntGas(), blk.GasPrice())
	blk.State().Log().Info("before from: %v\nto: %v", tx.from.Balance(), nil)
	if err := tx.from.UpdateKeepers(tx.ld.Nonce, cost, tx.data.Threshold,
		tx.data.Keepers); err != nil {
		return err
	}
	blk.State().Log().Info("after from: %v\nto: %v", tx.from.Balance(), nil)
	return nil
}

func (tx *TxUpdateAccountKeepers) Event(ts int64) *Event {
	return nil
}
