// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package txn

import (
	"encoding/json"
	"math/big"

	"github.com/ldclabs/ldvm/ids"
	"github.com/ldclabs/ldvm/ld"
	"github.com/ldclabs/ldvm/ld/service"
	"github.com/ldclabs/ldvm/util/encoding"
	"github.com/ldclabs/ldvm/util/erring"
)

type TxCreateData struct {
	TxBase
	input *ld.TxUpdater
	di    *ld.DataInfo
	ns    *service.Name
}

func (tx *TxCreateData) MarshalJSON() ([]byte, error) {
	if tx == nil || tx.ld == nil {
		return []byte("null"), nil
	}

	v := tx.ld.Copy()
	errp := erring.ErrPrefix("txn.TxCreateData.MarshalJSON: ")
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

// TxCreateData{ID, Version, Threshold, Keepers, Data} no model keepers
// TxCreateData{ID, Version, To, Amount, Threshold, Keepers, Data, Expire} with model keepers
func (tx *TxCreateData) SyntacticVerify() error {
	var err error
	errp := erring.ErrPrefix("txn.TxCreateData.SyntacticVerify: ")

	if err = tx.TxBase.SyntacticVerify(); err != nil {
		return errp.ErrorIf(err)
	}

	switch {
	case tx.ld.Tx.Token != nil:
		return errp.Errorf("invalid token, should be nil")

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

	switch {
	case tx.input.ModelID == nil:
		return errp.Errorf("nil mid")

	case tx.input.Version != 1:
		return errp.Errorf("invalid version, expected 1, got %d", tx.input.Version)

	case tx.input.Threshold == nil:
		return errp.Errorf("nil threshold")

	case len(*tx.input.Keepers) == 0:
		return errp.Errorf("empty keepers")

	case len(tx.input.Data) == 0:
		return errp.Errorf("empty data")

	case tx.input.SigClaims != nil:
		return errp.Errorf("invalid sigClaims, should be nil")
	}

	tx.di = &ld.DataInfo{
		ModelID:   *tx.input.ModelID,
		Version:   1,
		Threshold: *tx.input.Threshold,
		Keepers:   *tx.input.Keepers,
		Payload:   tx.input.Data,
		ID:        ids.DataID(tx.ld.ID),
	}

	if tx.input.Approver != nil {
		tx.di.Approver = *tx.input.Approver
	}
	if tx.input.ApproveList != nil {
		tx.di.ApproveList = *tx.input.ApproveList
	}

	if tx.input.To == nil {
		switch {
		case tx.ld.Tx.To != nil:
			return errp.Errorf("invalid to, should be nil")

		case tx.ld.Tx.Amount != nil:
			return errp.Errorf("invalid amount, should be nil")

		case tx.ld.ExSignatures != nil:
			return errp.Errorf("invalid exSignatures, should be nil")
		}
	} else {
		// with model keepers
		switch {
		case tx.ld.Tx.To == nil || *tx.input.To != *tx.ld.Tx.To:
			return errp.Errorf("invalid to, expected %s, got %s",
				tx.input.To, tx.ld.Tx.To)

		case tx.input.Expire < tx.ld.Timestamp:
			return errp.Errorf("data expired")

		case tx.input.Amount == nil || tx.ld.Tx.Amount == nil:
			return errp.Errorf("nil amount")

		case tx.input.Amount.Cmp(tx.ld.Tx.Amount) != 0:
			return errp.Errorf("invalid amount, expected %s, got %s",
				tx.input.Amount, tx.ld.Tx.Amount)

		case len(tx.ld.ExSignatures) == 0:
			return errp.Errorf("no exSignatures")
		}
	}

	if err = tx.di.SyntacticVerify(); err != nil {
		return errp.ErrorIf(err)
	}
	return nil
}

// ApplyGenesis skipping signature verification
func (tx *TxCreateData) ApplyGenesis(ctx ChainContext, cs ChainState) error {
	var err error
	errp := erring.ErrPrefix("txn.TxCreateData.ApplyGenesis: ")

	tx.input = &ld.TxUpdater{}
	if err = tx.input.Unmarshal(tx.ld.Tx.Data); err != nil {
		return errp.ErrorIf(err)
	}
	if err = tx.input.SyntacticVerify(); err != nil {
		return errp.ErrorIf(err)
	}

	switch {
	case tx.input.ModelID == nil:
		return errp.Errorf("nil mid")

	case tx.input.Version != 1:
		return errp.Errorf("invalid version, expected 1")

	case tx.input.Threshold == nil:
		return errp.Errorf("nil threshold")

	case len(*tx.input.Keepers) == 0:
		return errp.Errorf("tx.input.Keepers keepers")

	case len(tx.input.Data) == 0:
		return errp.Errorf("invalid data")
	}

	tx.di = &ld.DataInfo{
		ModelID:   *tx.input.ModelID,
		Version:   1,
		Threshold: *tx.input.Threshold,
		Keepers:   *tx.input.Keepers,
		Payload:   tx.input.Data,
		ID:        ids.DataID(tx.ld.ID),
	}
	if err = tx.di.SyntacticVerify(); err != nil {
		return errp.ErrorIf(err)
	}

	tx.amount = new(big.Int)
	tx.tip = new(big.Int)
	tx.fee = new(big.Int)
	tx.cost = new(big.Int)

	if tx.ldc, err = cs.LoadAccount(ids.LDCAccount); err != nil {
		return errp.ErrorIf(err)
	}
	if tx.miner, err = cs.LoadAccount(ctx.Builder()); err != nil {
		return errp.ErrorIf(err)
	}

	if tx.from, err = cs.LoadAccount(tx.ld.Tx.From); err != nil {
		return errp.ErrorIf(err)
	}

	if err = cs.SaveData(tx.di); err != nil {
		return errp.ErrorIf(err)
	}
	return errp.ErrorIf(tx.TxBase.accept(ctx, cs))
}

func (tx *TxCreateData) Apply(ctx ChainContext, cs ChainState) error {
	var err error
	errp := erring.ErrPrefix("txn.TxCreateData.Apply: ")

	if err = tx.TxBase.verify(ctx, cs); err != nil {
		return errp.ErrorIf(err)
	}

	switch tx.di.ModelID {
	case ld.RawModelID:
		if tx.input.To != nil {
			return errp.Errorf("invalid to, should be nil")
		}

	case ld.CBORModelID:
		if tx.input.To != nil {
			return errp.Errorf("invalid to, should be nil")
		}
		if err = encoding.ValidCBOR(tx.input.Data); err != nil {
			return errp.Errorf("invalid CBOR encoding data: %v", err)
		}

	case ld.JSONModelID:
		if tx.input.To != nil {
			return errp.Errorf("invalid to, should be nil")
		}
		if !json.Valid(tx.input.Data) {
			return errp.Errorf("invalid JSON encoding data")
		}

	default:
		mi, err := cs.LoadModel(tx.di.ModelID)
		if err != nil {
			return errp.ErrorIf(err)
		}

		switch {
		case mi.Threshold == 0:
			if tx.input.To != nil {
				return errp.Errorf("invalid to, should be nil")
			}

		case mi.Threshold > 0:
			if tx.input.To == nil {
				return errp.Errorf("nil to")
			}

			if !mi.Verify(tx.ld.ExHash(), tx.ld.ExSignatures) {
				return errp.Errorf("invalid exSignatures for model keepers")
			}
		}

		if err = mi.Model().Valid(tx.di.Payload); err != nil {
			return errp.ErrorIf(err)
		}

		if ctx.ChainConfig().IsNameService(tx.di.ModelID) {
			tx.ns = &service.Name{}
			if err = tx.ns.Unmarshal(tx.di.Payload); err != nil {
				return errp.ErrorIf(err)
			}
			if err = tx.ns.SyntacticVerify(); err != nil {
				return errp.ErrorIf(err)
			}

			tx.ns.DataID = tx.di.ID
			if err = cs.SaveName(tx.ns); err != nil {
				return errp.ErrorIf(err)
			}
		}
	}

	if err = cs.SaveData(tx.di); err != nil {
		return errp.ErrorIf(err)
	}
	return errp.ErrorIf(tx.TxBase.accept(ctx, cs))
}
