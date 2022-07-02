// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package transaction

import (
	"encoding/json"

	"github.com/ldclabs/ldvm/ld"
	"github.com/ldclabs/ldvm/util"
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
	errp := util.ErrPrefix("TxTest.MarshalJSON error: ")
	if tx.input == nil {
		return nil, errp.Errorf("nil tx.input")
	}
	d, err := json.Marshal(tx.input)
	if err != nil {
		return nil, errp.ErrorIf(err)
	}
	v.Data = d
	return errp.ErrorMap(json.Marshal(v))
}

func (tx *TxTest) SyntacticVerify() error {
	var err error
	errp := util.ErrPrefix("TxTest.SyntacticVerify error: ")

	if err = tx.TxBase.SyntacticVerify(); err != nil {
		return errp.ErrorIf(err)
	}

	switch {
	case tx.ld.To != nil:
		return errp.Errorf("invalid to, should be nil")

	case tx.ld.Token != nil:
		return errp.Errorf("invalid token, should be nil")

	case tx.ld.Amount != nil:
		return errp.Errorf("invalid amount, should be nil")

	case len(tx.ld.Data) == 0:
		return errp.Errorf("invalid data")
	}

	tx.input = &ld.TxTester{}
	if err = tx.input.Unmarshal(tx.ld.Data); err != nil {
		return errp.ErrorIf(err)
	}
	if err = tx.input.SyntacticVerify(); err != nil {
		return errp.ErrorIf(err)
	}

	return nil
}

// call after SyntacticVerify
func (tx *TxTest) Apply(bctx BlockContext, bs BlockState) error {
	var err error
	errp := util.ErrPrefix("TxTest.Apply error: ")

	if err = tx.TxBase.verify(bctx, bs); err != nil {
		return errp.ErrorIf(err)
	}

	var data []byte
	switch tx.input.ObjectType {
	case ld.AddressObject:
		acc, err := bs.LoadAccount(util.EthID(tx.input.ObjectID))
		if err == nil {
			data, err = acc.Marshal()
		}
		if err != nil {
			return errp.ErrorIf(err)
		}

	case ld.ModelObject:
		mi, err := bs.LoadModel(util.ModelID(tx.input.ObjectID))
		if err != nil {
			return errp.ErrorIf(err)
		}
		data = mi.Bytes()

	case ld.DataObject:
		di, err := bs.LoadData(util.DataID(tx.input.ObjectID))
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
	return errp.ErrorIf(tx.TxBase.accept(bctx, bs))
}
