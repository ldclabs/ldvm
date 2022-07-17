// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package transaction

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

// TxCreateData{ID, Version, Threshold, Keepers, Data, KSig} no model keepers
// TxCreateData{ID, Version, To, Amount, Threshold, Keepers, Data, KSig, MSig, Expire} with model keepers
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
		return errp.Errorf("invalid data")

	case tx.input.KSig == nil:
		return errp.Errorf("nil kSig")
	}

	tx.di = &ld.DataInfo{
		ModelID:     *tx.input.ModelID,
		Version:     1,
		Threshold:   *tx.input.Threshold,
		Keepers:     *tx.input.Keepers,
		Approver:    tx.input.Approver,
		ApproveList: tx.input.ApproveList,
		Data:        tx.input.Data,
		KSig:        *tx.input.KSig,
		MSig:        tx.input.MSig,
		ID:          util.DataID(tx.ld.ShortID()),
	}
	if err := tx.di.VerifySig(tx.di.Keepers, tx.di.KSig); err != nil {
		return errp.Errorf("invalid kSig, %v", err)
	}

	if tx.input.To == nil {
		switch {
		case tx.ld.To != nil:
			return errp.Errorf("invalid to, should be nil")

		case tx.ld.Amount != nil:
			return errp.Errorf("invalid amount, should be nil")

		case tx.ld.ExSignatures != nil:
			return errp.Errorf("invalid exSignatures, should be nil")

		case tx.input.MSig != nil:
			return errp.Errorf("invalid mSig, should be nil")
		}
	} else {
		// with model keepers
		switch {
		case tx.ld.To == nil || *tx.input.To != *tx.ld.To:
			return errp.Errorf("invalid to, expected %s, got %s",
				tx.input.To, tx.ld.To)

		case tx.input.Expire < tx.ld.Timestamp:
			return errp.Errorf("data expired")

		case tx.input.MSig == nil:
			return errp.Errorf("nil mSig")

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
	return nil
}

// ApplyGenesis skipping signature verification
func (tx *TxCreateData) ApplyGenesis(bctx BlockContext, bs BlockState) error {
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

	tx.amount = new(big.Int)
	tx.tip = new(big.Int)
	tx.fee = new(big.Int)
	tx.cost = new(big.Int)

	if tx.ldc, err = bs.LoadAccount(constants.LDCAccount); err != nil {
		return errp.ErrorIf(err)
	}
	if tx.miner, err = bs.LoadMiner(bctx.Miner()); err != nil {
		return errp.ErrorIf(err)
	}

	if tx.from, err = bs.LoadAccount(tx.ld.From); err != nil {
		return errp.ErrorIf(err)
	}

	if tx.ns != nil {
		if err = bs.SetName(tx.ns.Name, tx.di.ID); err != nil {
			return errp.ErrorIf(err)
		}
	}
	if err = bs.SaveData(tx.di.ID, tx.di); err != nil {
		return errp.ErrorIf(err)
	}
	return errp.ErrorIf(tx.TxBase.accept(bctx, bs))
}

func (tx *TxCreateData) Apply(bctx BlockContext, bs BlockState) error {
	var err error
	errp := util.ErrPrefix("TxCreateData.Apply error: ")

	if err = tx.TxBase.verify(bctx, bs); err != nil {
		return errp.ErrorIf(err)
	}

	switch tx.di.ModelID {
	case constants.RawModelID:
		if tx.input.To != nil {
			return errp.Errorf("invalid to, should be nil")
		}

	case constants.CBORModelID:
		if tx.input.To != nil {
			return errp.Errorf("invalid to, should be nil")
		}
		if err = util.ValidCBOR(tx.input.Data); err != nil {
			return errp.Errorf("invalid CBOR encoding data: %v", err)
		}

	case constants.JSONModelID:
		if tx.input.To != nil {
			return errp.Errorf("invalid to, should be nil")
		}
		if !json.Valid(tx.input.Data) {
			return errp.Errorf("invalid JSON encoding data")
		}

	default:
		mi, err := bs.LoadModel(tx.di.ModelID)
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
			if err = tx.di.VerifySig(mi.Keepers, *tx.input.MSig); err != nil {
				return errp.Errorf("invalid mSig for model keepers, %v", err)
			}
			if !util.SatisfySigning(mi.Threshold, mi.Keepers, tx.exSigners, true) {
				return errp.Errorf("invalid exSignatures for model keepers")
			}
			tx.di.MSig = tx.input.MSig
		}

		if err = mi.Model().Valid(tx.di.Data); err != nil {
			return errp.ErrorIf(err)
		}

		if bctx.ChainConfig().IsNameService(tx.di.ModelID) {
			tx.ns = &service.Name{}
			if err = tx.ns.Unmarshal(tx.di.Data); err != nil {
				return errp.ErrorIf(err)
			}
			if err = tx.ns.SyntacticVerify(); err != nil {
				return errp.ErrorIf(err)
			}
			if _, err = bs.ResolveNameID(tx.ns.Name); err == nil {
				return errp.Errorf("name %q conflict", tx.ns.Name)
			}

			if err = bs.SetName(tx.ns.Name, tx.di.ID); err != nil {
				return errp.ErrorIf(err)
			}
		}
	}

	if err = bs.SaveData(tx.di.ID, tx.di); err != nil {
		return errp.ErrorIf(err)
	}
	return errp.ErrorIf(tx.TxBase.accept(bctx, bs))
}
