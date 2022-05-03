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

type TxOpenLending struct {
	*TxBase
	data *ld.LendingConfig
}

func (tx *TxOpenLending) MarshalJSON() ([]byte, error) {
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

func (tx *TxOpenLending) SyntacticVerify() error {
	var err error
	if err = tx.TxBase.SyntacticVerify(); err != nil {
		return err
	}

	if tx.ld.Token != constants.LDCAccount {
		return fmt.Errorf("invalid token %s, required native LDC", util.EthID(tx.ld.Token))
	}

	if tx.ld.To != ids.ShortEmpty {
		return fmt.Errorf("TxOpenLending invalid to: %s", util.EthID(tx.ld.To).String())
	}
	if tx.ld.Amount.Sign() != 0 {
		return fmt.Errorf("TxOpenLending invalid amount, expected 0, got %v", tx.ld.Amount)
	}

	if len(tx.ld.Data) == 0 {
		return fmt.Errorf("TxOpenLending invalid")
	}
	tx.data = &ld.LendingConfig{}
	if err = tx.data.Unmarshal(tx.ld.Data); err != nil {
		return fmt.Errorf("TxOpenLending unmarshal data failed: %v", err)
	}
	if err = tx.data.SyntacticVerify(); err != nil {
		return fmt.Errorf("TxOpenLending SyntacticVerify failed: %v", err)
	}
	return nil
}

func (tx *TxOpenLending) Verify(blk *Block, bs BlockState) error {
	var err error
	if err = tx.TxBase.Verify(blk, bs); err != nil {
		return err
	}
	return tx.from.CheckOpenLending(tx.data)
}

func (tx *TxOpenLending) Accept(blk *Block, bs BlockState) error {
	var err error
	if err = tx.from.OpenLending(tx.data); err != nil {
		return err
	}
	return tx.TxBase.Accept(blk, bs)
}
