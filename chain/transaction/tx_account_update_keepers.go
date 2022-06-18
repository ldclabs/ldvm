// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package transaction

import (
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/ldclabs/ldvm/constants"
	"github.com/ldclabs/ldvm/ld"
)

type TxUpdateAccountKeepers struct {
	TxBase
	input *ld.TxAccounter
}

func (tx *TxUpdateAccountKeepers) MarshalJSON() ([]byte, error) {
	if tx == nil || tx.ld == nil {
		return []byte("null"), nil
	}
	v := tx.ld.Copy()
	if tx.input == nil {
		return nil, fmt.Errorf("TxUpdateAccountKeepers.MarshalJSON failed: invalid tx.input")
	}
	d, err := json.Marshal(tx.input)
	if err != nil {
		return nil, err
	}
	v.Data = d
	return json.Marshal(v)
}

func (tx *TxUpdateAccountKeepers) SyntacticVerify() error {
	var err error
	errPrefix := "TxUpdateAccountKeepers.SyntacticVerify failed:"
	if err = tx.TxBase.SyntacticVerify(); err != nil {
		return fmt.Errorf("%s %v", errPrefix, err)
	}

	switch {
	case tx.ld.To != nil:
		return fmt.Errorf("%s invalid to, should be nil", errPrefix)

	case tx.ld.Token != nil:
		return fmt.Errorf("%s invalid token, should be nil", errPrefix)

	case tx.ld.Amount != nil:
		return fmt.Errorf("%s invalid amount, should be nil", errPrefix)

	case len(tx.ld.Data) == 0:
		return fmt.Errorf("%s invalid data", errPrefix)
	}

	tx.input = &ld.TxAccounter{}
	if err = tx.input.Unmarshal(tx.ld.Data); err != nil {
		return fmt.Errorf("%s %v", errPrefix, err)
	}
	if err = tx.input.SyntacticVerify(); err != nil {
		return fmt.Errorf("%s %v", errPrefix, err)
	}

	if tx.input.Threshold == nil && tx.input.Approver == nil && tx.input.ApproveList == nil {
		return fmt.Errorf("%s no keepers nor approver", errPrefix)
	}
	if tx.input.Threshold != nil && *tx.input.Threshold == 0 {
		return fmt.Errorf("%s invalid threshold, expected >= 1", errPrefix)
	}
	return nil
}

// VerifyGenesis skipping signature verification
func (tx *TxUpdateAccountKeepers) VerifyGenesis(bctx BlockContext, bs BlockState) error {
	var err error
	tx.input = &ld.TxAccounter{}
	if err = tx.input.Unmarshal(tx.ld.Data); err != nil {
		return fmt.Errorf("TxUpdateAccountKeepers.VerifyGenesis failed: %v", err)
	}
	if err = tx.input.SyntacticVerify(); err != nil {
		return fmt.Errorf("TxUpdateAccountKeepers.VerifyGenesis failed: %v", err)
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

func (tx *TxUpdateAccountKeepers) Verify(bctx BlockContext, bs BlockState) error {
	var err error
	if err = tx.TxBase.Verify(bctx, bs); err != nil {
		return fmt.Errorf("TxUpdateAccountKeepers.Verify failed: %v", err)
	}
	if !tx.from.SatisfySigningPlus(tx.signers) {
		return fmt.Errorf("TxUpdateAccountKeepers.Verify failed: invalid signatures for keepers")
	}
	return nil
}

func (tx *TxUpdateAccountKeepers) Accept(bctx BlockContext, bs BlockState) error {
	var err error
	if err = tx.from.UpdateKeepers(
		tx.input.Threshold, tx.input.Keepers, tx.input.Approver, tx.input.ApproveList); err != nil {
		return err
	}

	return tx.TxBase.Accept(bctx, bs)
}
