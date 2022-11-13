// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package txn

import (
	"encoding/json"

	"github.com/ldclabs/ldvm/ids"
	"github.com/ldclabs/ldvm/ld"
	"github.com/ldclabs/ldvm/util/erring"
)

type TxTest struct {
	TxBase
	input *ld.TxTester
}

func (tx *TxTest) MarshalJSON() ([]byte, error) {
	if tx == nil || tx.ld == nil {
		return []byte("null"), nil
	}

	v := tx.ld.Copy()
	errp := erring.ErrPrefix("txn.TxTest.MarshalJSON: ")
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

func (tx *TxTest) SyntacticVerify() error {
	var err error
	errp := erring.ErrPrefix("txn.TxTest.SyntacticVerify: ")

	if err = tx.TxBase.SyntacticVerify(); err != nil {
		return errp.ErrorIf(err)
	}

	switch {
	case tx.ld.Tx.To != nil:
		return errp.Errorf("invalid to, should be nil")

	case tx.ld.Tx.Token != nil:
		return errp.Errorf("invalid token, should be nil")

	case tx.ld.Tx.Amount != nil:
		return errp.Errorf("invalid amount, should be nil")

	case len(tx.ld.Tx.Data) == 0:
		return errp.Errorf("invalid data")
	}

	tx.input = &ld.TxTester{}
	if err = tx.input.Unmarshal(tx.ld.Tx.Data); err != nil {
		return errp.ErrorIf(err)
	}
	if err = tx.input.SyntacticVerify(); err != nil {
		return errp.ErrorIf(err)
	}

	return nil
}

// call after SyntacticVerify
func (tx *TxTest) Apply(ctx ChainContext, cs ChainState) error {
	var err error
	errp := erring.ErrPrefix("txn.TxTest.Apply: ")

	if err = tx.TxBase.verify(ctx, cs); err != nil {
		return errp.ErrorIf(err)
	}

	var data []byte
	switch tx.input.ObjectType {
	case ld.AddressObject:
		acc, err := cs.LoadAccount(ids.Address(tx.input.ID20))
		if err == nil {
			data, _, err = acc.Marshal()
		}
		if err != nil {
			return errp.ErrorIf(err)
		}

	case ld.LedgerObject:
		acc, err := cs.LoadAccount(ids.Address(tx.input.ID20))
		if err == nil {
			_, data, err = acc.Marshal()
		}
		if err != nil {
			return errp.ErrorIf(err)
		}

	case ld.ModelObject:
		mi, err := cs.LoadModel(ids.ModelID(tx.input.ID20))
		if err != nil {
			return errp.ErrorIf(err)
		}
		data = mi.Bytes()

	case ld.DataObject:
		di, err := cs.LoadData(ids.DataID(tx.input.ID32))
		if err != nil {
			return errp.ErrorIf(err)
		}
		data = di.Bytes()

	default:
		return errp.Errorf("invalid type %s", tx.input.ObjectType)
	}

	if err = tx.input.Test(data); err != nil {
		return errp.ErrorIf(err)
	}
	return errp.ErrorIf(tx.TxBase.accept(ctx, cs))
}
