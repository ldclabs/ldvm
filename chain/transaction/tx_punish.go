// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package transaction

import (
	"encoding/json"
	"fmt"

	"github.com/ldclabs/ldvm/constants"
	"github.com/ldclabs/ldvm/ld"
	"github.com/ldclabs/ldvm/util"
)

type TxPunish struct {
	TxBase
	input *ld.TxUpdater
	di    *ld.DataInfo
}

func (tx *TxPunish) MarshalJSON() ([]byte, error) {
	if tx == nil || tx.ld == nil {
		return []byte("null"), nil
	}
	v := tx.ld.Copy()
	if tx.input == nil {
		return nil, fmt.Errorf("TxPunish.MarshalJSON failed: invalid tx.input")
	}
	d, err := json.Marshal(tx.input)
	if err != nil {
		return nil, err
	}
	v.Data = d
	return json.Marshal(v)
}

func (tx *TxPunish) SyntacticVerify() error {
	var err error
	errPrefix := "TxPunish.SyntacticVerify failed:"
	if err = tx.TxBase.SyntacticVerify(); err != nil {
		return fmt.Errorf("%s %v", errPrefix, err)
	}

	switch {
	case tx.ld.From != constants.GenesisAccount:
		return fmt.Errorf("%s invalid from, expected GenesisAccount, got %s", errPrefix, tx.ld.From)

	case tx.ld.To != nil:
		return fmt.Errorf("%s invalid to, should be nil", errPrefix)

	case tx.ld.Token != nil:
		return fmt.Errorf("%s invalid token, should be nil", errPrefix)

	case tx.ld.Amount != nil:
		return fmt.Errorf("%s invalid amount, should be nil", errPrefix)

	case len(tx.ld.Data) == 0:
		return fmt.Errorf("%s invalid data", errPrefix)
	}

	tx.input = &ld.TxUpdater{}
	if err = tx.input.Unmarshal(tx.ld.Data); err != nil {
		return fmt.Errorf("%s %v", errPrefix, err)
	}
	if err = tx.input.SyntacticVerify(); err != nil {
		return fmt.Errorf("%s %v", errPrefix, err)
	}

	switch {
	case tx.input.ID == nil:
		return fmt.Errorf("%s nil data id", errPrefix)

	case *tx.input.ID == util.DataIDEmpty:
		return fmt.Errorf("%s invalid data id", errPrefix)
	}
	return nil
}

func (tx *TxPunish) Verify(bctx BlockContext, bs BlockState) error {
	var err error
	if err = tx.TxBase.Verify(bctx, bs); err != nil {
		return fmt.Errorf("TxPunish.Verify failed: %v", err)
	}

	if tx.di, err = bs.LoadData(*tx.input.ID); err != nil {
		return fmt.Errorf("TxPunish.Verify failed: %v", err)
	}
	return nil
}

func (tx *TxPunish) Accept(bctx BlockContext, bs BlockState) error {
	var err error

	if err = bs.DeleteData(*tx.input.ID, tx.di, tx.input.Data); err != nil {
		return err
	}
	return tx.TxBase.Accept(bctx, bs)
}
