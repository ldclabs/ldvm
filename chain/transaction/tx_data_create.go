// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package transaction

import (
	"encoding/json"
	"fmt"
	"math/big"
	"strconv"

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
	name      *service.Name
}

func (tx *TxCreateData) MarshalJSON() ([]byte, error) {
	if tx == nil || tx.ld == nil {
		return []byte("null"), nil
	}
	v := tx.ld.Copy()
	if tx.input == nil {
		return nil, fmt.Errorf("TxCreateModel.MarshalJSON failed: invalid tx.input")
	}
	d, err := json.Marshal(tx.input)
	if err != nil {
		return nil, err
	}
	v.Data = d
	return json.Marshal(v)
}

// TxCreateData{ID, Version, Threshold, Keepers, Data, KSig} no model keepers
// TxCreateData{ID, Version, To, Amount, Threshold, Keepers, Data, KSig, MSig, Expire} with model keepers
func (tx *TxCreateData) SyntacticVerify() error {
	var err error
	errPrefix := "TxCreateData.SyntacticVerify failed:"
	if err = tx.TxBase.SyntacticVerify(); err != nil {
		return fmt.Errorf("%s %v", errPrefix, err)
	}

	switch {
	case tx.ld.Token != nil:
		return fmt.Errorf("%s invalid token, should be nil", errPrefix)

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
	case tx.input.ModelID == nil:
		return fmt.Errorf("%s nil mid", errPrefix)

	case tx.input.Version != 1:
		return fmt.Errorf("%s invalid version, expected 1, got %d", errPrefix, tx.input.Version)

	case tx.input.Threshold == nil:
		return fmt.Errorf("%s nil threshold", errPrefix)

	case len(*tx.input.Keepers) == 0:
		return fmt.Errorf("%s empty keepers", errPrefix)

	case len(tx.input.Data) == 0:
		return fmt.Errorf("%s invalid data", errPrefix)

	case tx.input.KSig == nil:
		return fmt.Errorf("%s nil kSig", errPrefix)
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
		return fmt.Errorf("%s invalid kSig, %v", errPrefix, err)
	}

	if tx.input.To == nil {
		switch {
		case tx.ld.To != nil:
			return fmt.Errorf("%s invalid to, should be nil", errPrefix)

		case tx.ld.Amount != nil:
			return fmt.Errorf("%s invalid amount, should be nil", errPrefix)

		case tx.ld.ExSignatures != nil:
			return fmt.Errorf("%s invalid exSignatures, should be nil", errPrefix)

		case tx.input.MSig != nil:
			return fmt.Errorf("%s invalid mSig, should be nil", errPrefix)
		}
	} else {
		// with model keepers
		switch {
		case tx.ld.To == nil || *tx.input.To != *tx.ld.To:
			return fmt.Errorf("%s invalid to, expected %s, got %s",
				errPrefix, tx.input.To, tx.ld.To)

		case tx.input.Expire < tx.ld.Timestamp:
			return fmt.Errorf("%s data expired", errPrefix)

		case tx.input.MSig == nil:
			return fmt.Errorf("%s nil mSig", errPrefix)

		case tx.input.Amount == nil || tx.ld.Amount == nil:
			return fmt.Errorf("%s nil amount", errPrefix)

		case tx.input.Amount.Cmp(tx.ld.Amount) != 0:
			return fmt.Errorf("%s invalid amount, expected %s, got %s",
				errPrefix, tx.input.Amount, tx.ld.Amount)
		}

		tx.exSigners, err = tx.ld.ExSigners()
		if err != nil {
			return fmt.Errorf("%s invalid exSignatures: %v", errPrefix, err)
		}
	}
	return nil
}

// VerifyGenesis skipping signature verification
func (tx *TxCreateData) VerifyGenesis(bctx BlockContext, bs BlockState) error {
	var err error
	tx.input = &ld.TxUpdater{}
	if err = tx.input.Unmarshal(tx.ld.Data); err != nil {
		return fmt.Errorf("TxCreateData.VerifyGenesis failed: %v", err)
	}
	if err = tx.input.SyntacticVerify(); err != nil {
		return fmt.Errorf("TxCreateData.VerifyGenesis failed: %v", err)
	}

	switch {
	case tx.input.ModelID == nil:
		return fmt.Errorf("TxCreateData.VerifyGenesis failed: nil mid")

	case tx.input.Version != 1:
		return fmt.Errorf("TxCreateData.VerifyGenesis failed: invalid version, expected 1")

	case tx.input.Threshold == nil:
		return fmt.Errorf("TxCreateData.VerifyGenesis failed: nil threshold")

	case len(*tx.input.Keepers) == 0:
		return fmt.Errorf("TxCreateData.VerifyGenesis failed: tx.input.Keepers keepers")

	case len(tx.input.Data) == 0:
		return fmt.Errorf("TxCreateData.VerifyGenesis failed: invalid data")
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
		return err
	}
	if tx.miner, err = bs.LoadMiner(bctx.Miner()); err != nil {
		return err
	}
	tx.from, err = bs.LoadAccount(tx.ld.From)
	return err
}

func (tx *TxCreateData) Verify(bctx BlockContext, bs BlockState) error {
	var err error
	errPrefix := "TxCreateData.Verify failed:"
	if err = tx.TxBase.Verify(bctx, bs); err != nil {
		return fmt.Errorf("%s %v", errPrefix, err)
	}

	switch tx.di.ModelID {
	case constants.RawModelID:
		if tx.input.To != nil {
			return fmt.Errorf("%s invalid to, should be nil", errPrefix)
		}
		return nil

	case constants.CBORModelID:
		if tx.input.To != nil {
			return fmt.Errorf("%s invalid to, should be nil", errPrefix)
		}
		if err = ld.ValidCBOR(tx.input.Data); err != nil {
			return fmt.Errorf("%s invalid CBOR encoding data: %v", errPrefix, err)
		}
		return nil

	case constants.JSONModelID:
		if tx.input.To != nil {
			return fmt.Errorf("%s invalid to, should be nil", errPrefix)
		}
		if !json.Valid(tx.input.Data) {
			return fmt.Errorf("%s invalid JSON encoding data", errPrefix)
		}
		return nil
	}

	mi, err := bs.LoadModel(tx.di.ModelID)
	if err != nil {
		return fmt.Errorf("%s %v", errPrefix, err)
	}

	switch {
	case mi.Threshold == 0:
		if tx.input.To != nil {
			return fmt.Errorf("%s invalid to, should be nil", errPrefix)
		}

	case mi.Threshold > 0:
		if tx.input.To == nil {
			return fmt.Errorf("%s nil to", errPrefix)
		}
		if err = tx.di.VerifySig(mi.Keepers, *tx.input.MSig); err != nil {
			return fmt.Errorf("%s invalid mSig for model keepers, %v", errPrefix, err)
		}
		if !util.SatisfySigning(mi.Threshold, mi.Keepers, tx.exSigners, true) {
			return fmt.Errorf("%s invalid exSignatures for model keepers", errPrefix)
		}
		tx.di.MSig = tx.input.MSig
	}

	if err = mi.Model().Valid(tx.di.Data); err != nil {
		return fmt.Errorf("%s %v", errPrefix, err)
	}

	if bctx.Chain().IsNameService(tx.di.ModelID) {
		tx.name = &service.Name{}
		if err = tx.name.Unmarshal(tx.di.Data); err != nil {
			return err
		}
		if err = tx.name.SyntacticVerify(); err != nil {
			return err
		}
		_, err = bs.ResolveNameID(tx.name.Name)
		if err == nil {
			return fmt.Errorf("%s name %s conflict", errPrefix, strconv.Quote(tx.name.Name))
		}
	}
	return nil
}

func (tx *TxCreateData) Accept(bctx BlockContext, bs BlockState) error {
	var err error

	if tx.name != nil {
		if err = bs.SetName(tx.name.Name, tx.di.ID); err != nil {
			return err
		}
	}
	if err = bs.SaveData(tx.di.ID, tx.di); err != nil {
		return err
	}
	return tx.TxBase.Accept(bctx, bs)
}
