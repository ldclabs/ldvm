// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package transaction

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ldclabs/ldvm/constants"
	"github.com/ldclabs/ldvm/ld"
)

type TxPunish struct {
	TxBase
	data *ld.TxUpdater
	dm   *ld.DataMeta
}

func (tx *TxPunish) MarshalJSON() ([]byte, error) {
	if tx == nil || tx.ld == nil {
		return []byte("null"), nil
	}
	v := tx.ld.Copy()
	if tx.data == nil {
		return nil, fmt.Errorf("MarshalJSON failed: data not exists")
	}
	d, err := json.Marshal(tx.data)
	if err != nil {
		return nil, err
	}
	v.Data = d
	return json.Marshal(v)
}

func (tx *TxPunish) SyntacticVerify() error {
	var err error
	if err = tx.TxBase.SyntacticVerify(); err != nil {
		return err
	}

	if tx.ld.Token != nil {
		return fmt.Errorf("invalid token, expected NativeToken, got %s",
			strconv.Quote(tx.ld.Token.GoString()))
	}
	if tx.ld.From != constants.GenesisAccount {
		return fmt.Errorf("TxPunish invalid from, expected GenesisAccount")
	}
	if len(tx.ld.Data) == 0 {
		return fmt.Errorf("TxPunish invalid")
	}
	tx.data = &ld.TxUpdater{}
	if err = tx.data.Unmarshal(tx.ld.Data); err != nil {
		return fmt.Errorf("TxPunish unmarshal data failed: %v", err)
	}
	if err = tx.data.SyntacticVerify(); err != nil {
		return fmt.Errorf("TxPunish SyntacticVerify failed: %v", err)
	}
	if tx.data.ID == nil {
		return fmt.Errorf("TxPunish invalid TxUpdater")
	}
	return nil
}

func (tx *TxPunish) Verify(bctx BlockContext, bs BlockState) error {
	var err error
	if err = tx.TxBase.Verify(bctx, bs); err != nil {
		return err
	}

	tx.dm, err = bs.LoadData(*tx.data.ID)
	if err != nil {
		return fmt.Errorf("TxPunish load data failed: %v", err)
	}
	return nil
}

func (tx *TxPunish) Accept(bctx BlockContext, bs BlockState) error {
	var err error

	tx.dm.Data = tx.data.Data
	if err = bs.DeleteData(*tx.data.ID, tx.dm); err != nil {
		return err
	}
	return tx.TxBase.Accept(bctx, bs)
}

func (tx *TxPunish) Event(ts int64) *Event {
	e := NewEvent(ids.ShortID(*tx.data.ID), SrcData, ActionDelete)
	e.Time = ts
	return e
}
