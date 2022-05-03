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

type TxAddAccountNonceTable struct {
	*TxBase
	data *ld.TxUpdater
}

func (tx *TxAddAccountNonceTable) MarshalJSON() ([]byte, error) {
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

// VerifyGenesis skipping signature verification
func (tx *TxAddAccountNonceTable) SyntacticVerify() error {
	var err error
	if err = tx.TxBase.SyntacticVerify(); err != nil {
		return err
	}

	if tx.ld.Token != constants.LDCAccount {
		return fmt.Errorf("invalid token %s, required LDC", util.EthID(tx.ld.Token))
	}
	if tx.ld.To != ids.ShortEmpty {
		return fmt.Errorf("TxAddAccountNonceTable invalid to")
	}
	if tx.ld.Amount != nil {
		return fmt.Errorf("TxAddAccountNonceTable invalid amount")
	}
	if len(tx.ld.Data) == 0 {
		return fmt.Errorf("TxAddAccountNonceTable invalid")
	}
	tx.data = &ld.TxUpdater{}
	if err = tx.data.Unmarshal(tx.ld.Data); err != nil {
		return fmt.Errorf("TxAddAccountNonceTable unmarshal data failed: %v", err)
	}
	if err = tx.data.SyntacticVerify(); err != nil {
		return fmt.Errorf("TxAddAccountNonceTable SyntacticVerify failed: %v", err)
	}
	if len(tx.data.Numbers) == 0 {
		return fmt.Errorf("TxAddAccountNonceTable numbers empty")
	}
	if len(tx.data.Numbers) > 1024 {
		return fmt.Errorf("TxAddAccountNonceTable too many numbers")
	}
	now := tx.ld.Timestamp
	if tx.data.Expire < now || tx.data.Expire > (now+3600*24*7) {
		return fmt.Errorf("TxAddAccountNonceTable invalid expire")
	}
	return nil
}

func (tx *TxAddAccountNonceTable) Verify(blk *Block, bs BlockState) error {
	var err error
	if err = tx.TxBase.Verify(blk, bs); err != nil {
		return err
	}
	if err = tx.from.CheckNonceTable(tx.data.Expire, tx.data.Numbers); err != nil {
		return err
	}
	return nil
}

func (tx *TxAddAccountNonceTable) Accept(blk *Block, bs BlockState) error {
	var err error
	if err = tx.from.AddNonceTable(tx.data.Expire, tx.data.Numbers); err != nil {
		return err
	}

	return tx.TxBase.Accept(blk, bs)
}
