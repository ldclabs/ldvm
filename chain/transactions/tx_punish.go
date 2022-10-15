// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package transactions

import (
	"encoding/json"

	"github.com/ldclabs/ldvm/constants"
	"github.com/ldclabs/ldvm/ld"
	"github.com/ldclabs/ldvm/ld/service"
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
	errp := util.ErrPrefix("transactions.TxPunish.MarshalJSON: ")
	if tx.input == nil {
		return nil, errp.Errorf("nil tx.input")
	}
	d, err := json.Marshal(tx.input)
	if err != nil {
		return nil, errp.ErrorIf(err)
	}
	v.Tx.Data = d
	return errp.ErrorMap(json.Marshal(v))
}

func (tx *TxPunish) SyntacticVerify() error {
	var err error
	errp := util.ErrPrefix("transactions.TxPunish.SyntacticVerify: ")

	if err = tx.TxBase.SyntacticVerify(); err != nil {
		return errp.ErrorIf(err)
	}

	switch {
	case tx.ld.Tx.From != constants.GenesisAccount:
		return errp.Errorf("invalid from, expected GenesisAccount, got %s", tx.ld.Tx.From)

	case tx.ld.Tx.To != nil:
		return errp.Errorf("invalid to, should be nil")

	case tx.ld.Tx.Token != nil:
		return errp.Errorf("invalid token, should be nil")

	case tx.ld.Tx.Amount != nil:
		return errp.Errorf("invalid amount, should be nil")

	case len(tx.ld.Tx.Data) == 0:
		return errp.Errorf("invalid data")
	}

	tx.input = &ld.TxUpdater{}
	if err = tx.input.Unmarshal(tx.ld.Tx.Data); err != nil {
		return errp.ErrorIf(err)
	}
	if err = tx.input.SyntacticVerify(); err != nil {
		return errp.ErrorIf(err)
	}

	// we can pubish deleted data.
	switch {
	case tx.input.ID == nil || *tx.input.ID == util.DataIDEmpty:
		return errp.Errorf("invalid data id")

	case !util.ValidMessage(string(tx.input.Data)):
		return errp.Errorf("invalid deleting message")
	}
	return nil
}

func (tx *TxPunish) Apply(ctx ChainContext, cs ChainState) error {
	var err error
	errp := util.ErrPrefix("transactions.TxPunish.Apply: ")

	if err = tx.TxBase.verify(ctx, cs); err != nil {
		return errp.ErrorIf(err)
	}

	if tx.di, err = cs.LoadData(*tx.input.ID); err != nil {
		return errp.ErrorIf(err)
	}

	if ctx.ChainConfig().IsNameService(tx.di.ModelID) {
		ns := &service.Name{}
		if err = ns.Unmarshal(tx.di.Payload); err != nil {
			return errp.ErrorIf(err)
		}
		if err = ns.SyntacticVerify(); err != nil {
			return errp.ErrorIf(err)
		}

		ns.DataID = tx.di.ID
		if err = cs.DeleteName(ns); err != nil {
			return errp.ErrorIf(err)
		}
	}

	if err = cs.DeleteData(tx.di, tx.input.Data); err != nil {
		return errp.ErrorIf(err)
	}
	return errp.ErrorIf(tx.TxBase.accept(ctx, cs))
}
