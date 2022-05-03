// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package chain

import (
	"fmt"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ldclabs/ldvm/constants"
	"github.com/ldclabs/ldvm/ld"
	"github.com/ldclabs/ldvm/util"
)

type TxUpdateModelKeepers struct {
	*TxBase
	data *ld.TxUpdater
	mm   *ld.ModelMeta
}

func (tx *TxUpdateModelKeepers) MarshalJSON() ([]byte, error) {
	if tx == nil || tx.ld == nil {
		return util.Null, nil
	}
	v := tx.ld.Copy()
	if tx.data == nil {
		return nil, fmt.Errorf("MarshalJSON failed: data not exists")
	}
	d, err := tx.data.MarshalJSON()
	if err != nil {
		return nil, err
	}
	v.Data = d
	return v.MarshalJSON()
}

func (tx *TxUpdateModelKeepers) SyntacticVerify() error {
	var err error
	if err = tx.TxBase.SyntacticVerify(); err != nil {
		return err
	}

	if tx.ld.Token != constants.LDCAccount {
		return fmt.Errorf("invalid token %s, required LDC", util.EthID(tx.ld.Token))
	}
	if len(tx.ld.Data) == 0 {
		return fmt.Errorf("TxUpdateModelKeepers invalid")
	}

	tx.data = &ld.TxUpdater{}
	if err = tx.data.Unmarshal(tx.ld.Data); err != nil {
		return fmt.Errorf("TxUpdateModelKeepers unmarshal data failed: %v", err)
	}
	if err = tx.data.SyntacticVerify(); err != nil {
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

func (tx *TxUpdateModelKeepers) Verify(blk *Block, bs BlockState) error {
	var err error
	if err = tx.TxBase.Verify(blk, bs); err != nil {
		return err
	}

	tx.mm, err = bs.LoadModel(tx.data.ID)
	if err != nil {
		return fmt.Errorf("TxUpdateModelKeepers load model failed: %v", err)
	}

	if !util.SatisfySigningPlus(tx.mm.Threshold, tx.mm.Keepers, tx.signers) {
		return fmt.Errorf("TxUpdateModelKeepers need more signatures")
	}
	return nil
}

func (tx *TxUpdateModelKeepers) Accept(blk *Block, bs BlockState) error {
	var err error

	tx.mm.Threshold = tx.data.Threshold
	tx.mm.Keepers = tx.data.Keepers
	if err = tx.mm.SyntacticVerify(); err != nil {
		return err
	}
	if err = bs.SaveModel(tx.data.ID, tx.mm); err != nil {
		return err
	}
	return tx.TxBase.Accept(blk, bs)
}
