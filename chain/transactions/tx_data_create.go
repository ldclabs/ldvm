// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package transactions

import (
	"encoding/json"
	"math/big"

	"github.com/ldclabs/ldvm/constants"
	"github.com/ldclabs/ldvm/ld"
	"github.com/ldclabs/ldvm/ld/service"
	"github.com/ldclabs/ldvm/util"
)

type TxCreateData struct {
	TxBase
	exSigners util.EthIDs
	input     *ld.TxUpdater
	di        *ld.DataInfo
	ns        *service.Name
}

func (tx *TxCreateData) MarshalJSON() ([]byte, error) {
	if tx == nil || tx.ld == nil {
		return []byte("null"), nil
	}

	v := tx.ld.Copy()
	errp := util.ErrPrefix("TxCreateData.MarshalJSON error: ")
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

// TxCreateData{ID, Version, Threshold, Keepers, Data} no model keepers
// TxCreateData{ID, Version, To, Amount, Threshold, Keepers, Data, Expire} with model keepers
func (tx *TxCreateData) SyntacticVerify() error {
	var err error
	errp := util.ErrPrefix("TxCreateData.SyntacticVerify error: ")

	if err = tx.TxBase.SyntacticVerify(); err != nil {
		return errp.ErrorIf(err)
	}

	switch {
	case tx.ld.Token != nil:
		return errp.Errorf("invalid token, should be nil")

	case len(tx.ld.Data) == 0:
		return errp.Errorf("invalid data")
	}

	tx.input = &ld.TxUpdater{}
	if err = tx.input.Unmarshal(tx.ld.Data); err != nil {
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
		ModelID:     *tx.input.ModelID,
		Version:     1,
		Threshold:   *tx.input.Threshold,
		Keepers:     *tx.input.Keepers,
		Approver:    tx.input.Approver,
		ApproveList: tx.input.ApproveList,
		Data:        tx.input.Data,
		ID:          util.DataID(tx.ld.ShortID()),
	}

	if tx.input.To == nil {
		switch {
		case tx.ld.To != nil:
			return errp.Errorf("invalid to, should be nil")

		case tx.ld.Amount != nil:
			return errp.Errorf("invalid amount, should be nil")

		case tx.ld.ExSignatures != nil:
			return errp.Errorf("invalid exSignatures, should be nil")
		}
	} else {
		// with model keepers
		switch {
		case tx.ld.To == nil || *tx.input.To != *tx.ld.To:
			return errp.Errorf("invalid to, expected %s, got %s",
				tx.input.To, tx.ld.To)

		case tx.input.Expire < tx.ld.Timestamp:
			return errp.Errorf("data expired")

		case tx.input.Amount == nil || tx.ld.Amount == nil:
			return errp.Errorf("nil amount")

		case tx.input.Amount.Cmp(tx.ld.Amount) != 0:
			return errp.Errorf("invalid amount, expected %s, got %s",
				tx.input.Amount, tx.ld.Amount)
		}

		tx.exSigners, err = tx.ld.ExSigners()
		if err != nil {
			return errp.Errorf("invalid exSignatures, %v", err)
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
	errp := util.ErrPrefix("TxCreateData.ApplyGenesis error: ")

	tx.input = &ld.TxUpdater{}
	if err = tx.input.Unmarshal(tx.ld.Data); err != nil {
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
		Data:      tx.input.Data,
		ID:        util.DataID(tx.ld.ShortID()),
	}
	if err = tx.di.SyntacticVerify(); err != nil {
		return errp.ErrorIf(err)
	}

	tx.amount = new(big.Int)
	tx.tip = new(big.Int)
	tx.fee = new(big.Int)
	tx.cost = new(big.Int)

	if tx.ldc, err = cs.LoadAccount(constants.LDCAccount); err != nil {
		return errp.ErrorIf(err)
	}
	if tx.miner, err = cs.LoadMiner(ctx.Miner()); err != nil {
		return errp.ErrorIf(err)
	}

	if tx.from, err = cs.LoadAccount(tx.ld.From); err != nil {
		return errp.ErrorIf(err)
	}

	if err = cs.SaveData(tx.di); err != nil {
		return errp.ErrorIf(err)
	}
	return errp.ErrorIf(tx.TxBase.accept(ctx, cs))
}

func (tx *TxCreateData) Apply(ctx ChainContext, cs ChainState) error {
	var err error
	errp := util.ErrPrefix("TxCreateData.Apply error: ")

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
		if err = util.ValidCBOR(tx.input.Data); err != nil {
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

			if !util.SatisfySigning(mi.Threshold, mi.Keepers, tx.exSigners, true) {
				return errp.Errorf("invalid exSignatures for model keepers")
			}
		}

		if err = mi.Model().Valid(tx.di.Data); err != nil {
			return errp.ErrorIf(err)
		}

		if ctx.ChainConfig().IsNameService(tx.di.ModelID) {
			tx.ns = &service.Name{}
			if err = tx.ns.Unmarshal(tx.di.Data); err != nil {
				return errp.ErrorIf(err)
			}
			if err = tx.ns.SyntacticVerify(); err != nil {
				return errp.ErrorIf(err)
			}

			tx.ns.DID = tx.di.ID
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
