// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package chain

import (
	"fmt"
	"math/big"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ldclabs/ldvm/ld"
)

type TxUpdateDataKeepersByAuth struct {
	ld        *ld.Transaction
	from      *Account
	signers   []ids.ShortID
	exSigners []ids.ShortID
	data      *ld.TxUpdater
	dm        *ld.DataMeta
}

func (tx *TxUpdateDataKeepersByAuth) MarshalJSON() ([]byte, error) {
	if tx == nil {
		return ld.Null, nil
	}
	v := tx.ld.Copy()
	if tx.data == nil {
		tx.data = &ld.TxUpdater{}
		if err := tx.data.Unmarshal(tx.ld.Data); err != nil {
			return nil, fmt.Errorf("TxUpdateDataKeepersByAuth unmarshal data failed: %v", err)
		}
	}
	d, err := tx.data.MarshalJSON()
	if err != nil {
		return nil, err
	}
	v.Data = d
	return v.MarshalJSON()
}

func (tx *TxUpdateDataKeepersByAuth) ID() ids.ID {
	return tx.ld.ID()
}

func (tx *TxUpdateDataKeepersByAuth) Type() ld.TxType {
	return tx.ld.Type
}

func (tx *TxUpdateDataKeepersByAuth) Bytes() []byte {
	return tx.ld.Bytes()
}

func (tx *TxUpdateDataKeepersByAuth) SyntacticVerify() error {
	if tx == nil ||
		len(tx.ld.Data) == 0 {
		return fmt.Errorf("invalid TxUpdateDataKeepersByAuth")
	}

	var err error
	tx.data = &ld.TxUpdater{}
	if err = tx.data.Unmarshal(tx.ld.Data); err != nil {
		return fmt.Errorf("TxUpdateDataKeepersByAuth unmarshal data failed: %v", err)
	}
	if err = tx.data.SyntacticVerify(); err != nil {
		return fmt.Errorf("TxUpdateDataKeepersByAuth syntacticVerify failed: %v", err)
	}
	if tx.data.ID == ids.ShortEmpty ||
		tx.data.Version == 0 ||
		tx.data.Amount == nil ||
		tx.data.Amount.Cmp(tx.ld.Amount) != 0 ||
		tx.data.To != tx.ld.To {
		return fmt.Errorf("TxUpdateDataKeepersByAuth invalid TxUpdater")
	}
	return nil
}

func (tx *TxUpdateDataKeepersByAuth) Verify(blk *Block) error {
	var err error
	tx.signers, err = ld.DeriveSigners(tx.ld.UnsignedBytes(), tx.ld.Signatures)
	if err != nil {
		return fmt.Errorf("invalid signatures: %v", err)
	}

	tx.exSigners, err = ld.DeriveSigners(tx.ld.Data, tx.ld.ExSignatures)
	if err != nil {
		return fmt.Errorf("invalid exSignatures: %v", err)
	}

	if tx.from, err = verifyBase(blk, tx.ld, tx.signers); err != nil {
		return err
	}

	tx.dm, err = blk.State().LoadData(tx.data.ID)
	if err != nil {
		return fmt.Errorf("TxUpdateDataKeepersByAuth load data failed: %v", err)
	}
	if tx.dm.Version != tx.data.Version {
		return fmt.Errorf("TxUpdateDataKeepersByAuth version mismatch, expected %v, got %v",
			tx.dm.Version, tx.data.Version)
	}
	// verify seller's signatures
	if !ld.SatisfySigning(tx.dm.Threshold, tx.dm.Keepers, tx.exSigners, false) {
		return fmt.Errorf("TxUpdateDataKeepersByAuth need more exSignatures")
	}
	return nil
}

func (tx *TxUpdateDataKeepersByAuth) Accept(blk *Block) error {
	bs := blk.State()

	var err error
	tx.dm.Version++
	tx.dm.Threshold = tx.data.Threshold
	tx.dm.Keepers = tx.data.Keepers
	if len(tx.dm.Keepers) == 0 {
		tx.dm.Threshold = tx.from.Threshold()
		tx.dm.Keepers = tx.from.Keepers()
	}
	if err = tx.dm.SyntacticVerify(); err != nil {
		return err
	}

	var to *Account
	if to, err = bs.LoadAccount(tx.ld.To); err != nil {
		return err
	}

	cost := new(big.Int).Mul(tx.ld.BigIntGas(), blk.GasPrice())
	cost = new(big.Int).Add(tx.ld.Amount, cost)
	if err = tx.from.SubByNonce(tx.ld.Nonce, cost); err != nil {
		return err
	}
	if err = to.Add(tx.ld.Amount); err != nil {
		return err
	}
	return bs.SaveData(tx.data.ID, tx.dm)
}

func (tx *TxUpdateDataKeepersByAuth) Event(ts int64) *Event {
	return nil
}
