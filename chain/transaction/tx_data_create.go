// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package transaction

import (
	"encoding/json"
	"fmt"
	"math/big"
	"strconv"

	"github.com/fxamacker/cbor/v2"
	"github.com/ldclabs/ldvm/constants"
	"github.com/ldclabs/ldvm/ld"
	"github.com/ldclabs/ldvm/ld/service"
	"github.com/ldclabs/ldvm/util"
)

type TxCreateData struct {
	TxBase
	exSigners util.EthIDs
	input     *ld.TxUpdater
	dm        *ld.DataMeta
	name      *service.Name
	kSigner   *util.EthID
	mSigner   *util.EthID
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
	if err = tx.TxBase.SyntacticVerify(); err != nil {
		return err
	}

	switch {
	case tx.ld.Token != nil:
		return fmt.Errorf("TxCreateData.SyntacticVerify failed: invalid token, should be nil")
	case len(tx.ld.Data) == 0:
		return fmt.Errorf("TxCreateData.SyntacticVerify failed: invalid data")
	}

	tx.input = &ld.TxUpdater{}
	if err = tx.input.Unmarshal(tx.ld.Data); err != nil {
		return fmt.Errorf("TxCreateData.SyntacticVerify failed: %v", err)
	}
	if err = tx.input.SyntacticVerify(); err != nil {
		return fmt.Errorf("TxCreateData.SyntacticVerify failed: %v", err)
	}

	switch {
	case tx.input.ModelID == nil:
		return fmt.Errorf("TxCreateData.SyntacticVerify failed: nil mid")
	case tx.input.Version != 1:
		return fmt.Errorf("TxCreateData.SyntacticVerify failed: invalid version, expected 1, got %d",
			tx.input.Version)
	case tx.input.Threshold == nil:
		return fmt.Errorf("TxCreateData.SyntacticVerify failed: nil threshold")
	case len(*tx.input.Keepers) == 0:
		return fmt.Errorf("TxCreateData.SyntacticVerify failed: empty keepers")
	case len(tx.input.Data) == 0:
		return fmt.Errorf("TxCreateData.SyntacticVerify failed: invalid data")
	case tx.input.KSig == nil:
		return fmt.Errorf("TxCreateData.SyntacticVerify failed: nil kSig")
	}

	kSigner, err := util.DeriveSigner(tx.input.Data, (*tx.input.KSig)[:])
	if err != nil {
		return fmt.Errorf("TxCreateData.SyntacticVerify failed: invalid kSig: %v", err)
	}
	if !tx.input.Keepers.Has(kSigner) {
		return fmt.Errorf("TxCreateData.SyntacticVerify failed: invalid kSig for keepers")
	}
	tx.kSigner = &kSigner

	// with model keepers
	if tx.input.To != nil {
		switch {
		case tx.ld.To == nil || *tx.input.To != *tx.ld.To:
			return fmt.Errorf("TxCreateData.SyntacticVerify failed: invalid to, expected %s, got %s",
				tx.input.To, tx.ld.To)
		case tx.input.Expire < tx.ld.Timestamp:
			return fmt.Errorf("TxCreateData.SyntacticVerify failed: data expired")
		case tx.input.MSig == nil:
			return fmt.Errorf("TxCreateData.SyntacticVerify failed: nil mSig")
		case tx.input.Amount == nil || tx.ld.Amount == nil:
			return fmt.Errorf("TxCreateData.SyntacticVerify failed: nil amount")
		case tx.input.Amount.Cmp(tx.ld.Amount) != 0:
			return fmt.Errorf("TxCreateData.SyntacticVerify failed: invalid amount, expected %s, got %s",
				tx.input.Amount, tx.ld.Amount)
		}

		mSigner, err := util.DeriveSigner(tx.input.Data, (*tx.input.MSig)[:])
		if err != nil {
			return fmt.Errorf("TxCreateData.SyntacticVerify failed: invalid mSig: %v", err)
		}
		tx.mSigner = &mSigner

		tx.exSigners, err = tx.ld.ExSigners()
		if err != nil {
			return fmt.Errorf("TxCreateData.SyntacticVerify failed: invalid exSignatures: %v", err)
		}
	}

	tx.dm = &ld.DataMeta{
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
		return fmt.Errorf("TxCreateData.SyntacticVerify failed: nil mid")
	case tx.input.Version != 1:
		return fmt.Errorf("TxCreateData.SyntacticVerify failed: invalid version, expected 1")
	case tx.input.Threshold == nil:
		return fmt.Errorf("TxCreateData.SyntacticVerify failed: nil threshold")
	case len(*tx.input.Keepers) == 0:
		return fmt.Errorf("TxCreateData.SyntacticVerify failed: tx.input.Keepers keepers")
	case len(tx.input.Data) == 0:
		return fmt.Errorf("TxCreateData.SyntacticVerify failed: invalid data")
	}

	tx.dm = &ld.DataMeta{
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
	if err = tx.TxBase.Verify(bctx, bs); err != nil {
		return err
	}

	switch tx.dm.ModelID {
	case constants.RawModelID:
		return nil
	case constants.CBORModelID:
		if err = cbor.Valid(tx.input.Data); err != nil {
			return fmt.Errorf("TxCreateData.Verify failed: invalid CBOR encoding data: %v", err)
		}
		return nil
	case constants.JSONModelID:
		if !json.Valid(tx.input.Data) {
			return fmt.Errorf("TxCreateData.Verify failed: invalid JSON encoding data")
		}
		return nil
	}

	mm, err := bs.LoadModel(tx.dm.ModelID)
	if err != nil {
		return fmt.Errorf("TxCreateData.Verify failed: %v", err)
	}
	if mm.Threshold > 0 {
		if tx.mSigner == nil || !mm.Keepers.Has(*tx.mSigner) {
			return fmt.Errorf("TxCreateData.Verify failed: invalid mSig for model keepers")
		}
		if !util.SatisfySigning(mm.Threshold, mm.Keepers, tx.exSigners, true) {
			return fmt.Errorf("TxCreateData.Verify failed: invalid exSignatures for model keepers")
		}
	}

	if err = mm.Model().Valid(tx.dm.Data); err != nil {
		return fmt.Errorf("TxCreateData.Verify failed: %v", err)
	}

	if bctx.Chain().IsNameService(tx.dm.ModelID) {
		tx.name = &service.Name{}
		if err = tx.name.Unmarshal(tx.dm.Data); err != nil {
			return err
		}
		if err = tx.name.SyntacticVerify(); err != nil {
			return err
		}
		_, err = bs.ResolveNameID(tx.name.Name)
		if err == nil {
			return fmt.Errorf("TxCreateData.Verify failed: name %s conflict",
				strconv.Quote(tx.name.Name))
		}
	}
	return nil
}

func (tx *TxCreateData) Accept(bctx BlockContext, bs BlockState) error {
	var err error

	if tx.name != nil {
		if err = bs.SetName(tx.name.Name, tx.dm.ID); err != nil {
			return err
		}
	}
	if err = bs.SaveData(tx.dm.ID, tx.dm); err != nil {
		return err
	}
	return tx.TxBase.Accept(bctx, bs)
}
