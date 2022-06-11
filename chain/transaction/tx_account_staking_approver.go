// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package transaction

import (
	"encoding/json"
	"fmt"

	"github.com/ldclabs/ldvm/ld"
	"github.com/ldclabs/ldvm/util"
)

type TxUpdateStakeApprover struct {
	TxBase
	input *ld.TxAccounter
}

func (tx *TxUpdateStakeApprover) MarshalJSON() ([]byte, error) {
	if tx == nil || tx.ld == nil {
		return []byte("null"), nil
	}
	v := tx.ld.Copy()
	if tx.input == nil {
		return nil, fmt.Errorf("TxUpdateStakeApprover.MarshalJSON failed: invalid tx.input")
	}
	d, err := json.Marshal(tx.input)
	if err != nil {
		return nil, err
	}
	v.Data = d
	return json.Marshal(v)
}

func (tx *TxUpdateStakeApprover) SyntacticVerify() error {
	var err error
	errPrefix := "TxUpdateStakeApprover.SyntacticVerify failed:"
	if err = tx.TxBase.SyntacticVerify(); err != nil {
		return fmt.Errorf("%s %v", errPrefix, err)
	}

	switch {
	case tx.ld.To == nil:
		return fmt.Errorf("%s nil to as stake account", errPrefix)

	case tx.ld.Amount != nil:
		return fmt.Errorf("%s invalid amount, should be nil", errPrefix)

	case tx.ld.Token != nil:
		return fmt.Errorf("%s invalid token, should be nil", errPrefix)

	case len(tx.ld.Data) == 0:
		return fmt.Errorf("%s invalid data", errPrefix)
	}

	if stake := util.StakeSymbol(*tx.ld.To); !stake.Valid() {
		return fmt.Errorf("%s invalid stake account %s", errPrefix, stake.GoString())
	}

	tx.input = &ld.TxAccounter{}
	if err = tx.input.Unmarshal(tx.ld.Data); err != nil {
		return fmt.Errorf("%s %v", errPrefix, err)
	}
	if err = tx.input.SyntacticVerify(); err != nil {
		return fmt.Errorf("%s %v", errPrefix, err)
	}

	switch {
	case tx.input.Approver == nil:
		return fmt.Errorf("%s nil approver", errPrefix)

	case tx.input.ApproveList != nil:
		return fmt.Errorf("%s invalid approveList, should be nil", errPrefix)
	}

	return nil
}

func (tx *TxUpdateStakeApprover) Verify(bctx BlockContext, bs BlockState) error {
	var err error
	if err = tx.TxBase.Verify(bctx, bs); err != nil {
		return fmt.Errorf("TxUpdateStakeApprover.Verify failed: %v", err)
	}
	if err = tx.to.CheckUpdateStakeApprover(tx.ld.From, *tx.input.Approver, tx.signers); err != nil {
		return fmt.Errorf("TxUpdateStakeApprover.Verify failed: %v", err)
	}
	return nil
}

func (tx *TxUpdateStakeApprover) Accept(bctx BlockContext, bs BlockState) error {
	err := tx.to.UpdateStakeApprover(tx.ld.From, *tx.input.Approver, tx.signers)
	if err != nil {
		return err
	}
	return tx.TxBase.Accept(bctx, bs)
}
