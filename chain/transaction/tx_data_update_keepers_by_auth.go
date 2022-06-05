// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package transaction

import (
	"encoding/json"
	"fmt"

	"github.com/ldclabs/ldvm/constants"
	"github.com/ldclabs/ldvm/ld"
	"github.com/ldclabs/ldvm/util"
)

type TxUpdateDataKeepersByAuth struct {
	TxBase
	exSigners util.EthIDs
	input     *ld.TxUpdater
	dm        *ld.DataMeta
}

func (tx *TxUpdateDataKeepersByAuth) MarshalJSON() ([]byte, error) {
	if tx == nil || tx.ld == nil {
		return []byte("null"), nil
	}
	v := tx.ld.Copy()
	if tx.input == nil {
		return nil, fmt.Errorf("TxUpdateDataKeepersByAuth.MarshalJSON failed: invalid tx.input")
	}
	d, err := json.Marshal(tx.input)
	if err != nil {
		return nil, err
	}
	v.Data = d
	return json.Marshal(v)
}

func (tx *TxUpdateDataKeepersByAuth) SyntacticVerify() error {
	var err error
	if err = tx.TxBase.SyntacticVerify(); err != nil {
		return err
	}

	switch {
	case tx.ld.To == nil:
		return fmt.Errorf("TxUpdateDataKeepersByAuth.SyntacticVerify failed: nil to")
	case len(tx.ld.Data) == 0:
		return fmt.Errorf("TxUpdateDataKeepersByAuth.SyntacticVerify failed: invalid data")
	}

	tx.input = &ld.TxUpdater{}
	if err = tx.input.Unmarshal(tx.ld.Data); err != nil {
		return fmt.Errorf("TxUpdateDataKeepersByAuth.SyntacticVerify failed: %v", err)
	}
	if err = tx.input.SyntacticVerify(); err != nil {
		return fmt.Errorf("TxUpdateDataKeepersByAuth.SyntacticVerify failed: %v", err)
	}

	switch {
	case tx.input.ID == nil || *tx.input.ID == util.DataIDEmpty:
		return fmt.Errorf(
			"TxUpdateDataKeepersByAuth.SyntacticVerify failed: invalid data id")
	case tx.input.Version == 0:
		return fmt.Errorf(
			"TxUpdateDataKeepersByAuth.SyntacticVerify failed: invalid data version")
	case tx.input.Keepers != nil:
		return fmt.Errorf(
			"TxUpdateModelKeepers.SyntacticVerify failed: invalid keepers, should be nil")
	case tx.input.KSig != nil:
		return fmt.Errorf(
			"TxUpdateDataKeepersByAuth.SyntacticVerify failed: invalid kSig, should be nil")
	case tx.input.Approver != nil:
		return fmt.Errorf(
			"TxUpdateDataKeepersByAuth.SyntacticVerify failed: invalid approver, should be nil")
	case tx.input.ApproveList != nil:
		return fmt.Errorf(
			"TxUpdateDataKeepersByAuth.SyntacticVerify failed: invalid approveList, should be nil")
	case tx.input.To == nil:
		return fmt.Errorf(
			"TxUpdateDataKeepersByAuth.SyntacticVerify failed: nil to")
	case *tx.input.To != *tx.ld.To:
		return fmt.Errorf(
			"TxUpdateDataKeepersByAuth.SyntacticVerify failed: invalid to, expected %s, got %s",
			tx.input.To, tx.ld.To)
	case tx.input.Amount == nil:
		return fmt.Errorf(
			"TxUpdateDataKeepersByAuth.SyntacticVerify failed: nil amount")
	case tx.input.Amount.Cmp(tx.amount) != 0:
		return fmt.Errorf(
			"TxUpdateDataKeepersByAuth.SyntacticVerify failed: invalid amount, expected %v, got %v",
			tx.input.Amount, tx.amount)
	case tx.input.Token == nil && tx.token != constants.NativeToken:
		return fmt.Errorf(
			"TxUpdateDataKeepersByAuth.SyntacticVerify failed: invalid token, expected NativeToken, got %s",
			tx.token.GoString())
	case tx.input.Token != nil && tx.token != *tx.input.Token:
		return fmt.Errorf(
			"TxUpdateDataKeepersByAuth.SyntacticVerify failed: invalid token, expected %s, got %s",
			tx.input.Token.GoString(), tx.token.GoString())
	}

	tx.exSigners, err = tx.ld.ExSigners()
	if err != nil {
		return fmt.Errorf("TxUpdateDataKeepersByAuth.SyntacticVerify failed: invalid exSignatures: %v", err)
	}
	return nil
}

func (tx *TxUpdateDataKeepersByAuth) Verify(bctx BlockContext, bs BlockState) error {
	var err error
	if err = tx.TxBase.Verify(bctx, bs); err != nil {
		return err
	}

	tx.dm, err = bs.LoadData(*tx.input.ID)
	switch {
	case err != nil:
		return fmt.Errorf("TxUpdateDataKeepersByAuth.Verify failed: %v", err)
	case tx.dm.Version != tx.input.Version:
		return fmt.Errorf("TxUpdateDataKeepersByAuth.Verify failed: invalid version, expected %d, got %d",
			tx.dm.Version, tx.input.Version)
	case !util.SatisfySigningPlus(tx.dm.Threshold, tx.dm.Keepers, tx.exSigners):
		return fmt.Errorf("TxUpdateDataKeepersByAuth.Verify failed: invalid exSignatures for data keepers")
	case tx.ld.NeedApprove(tx.dm.Approver, tx.dm.ApproveList) && !tx.exSigners.Has(*tx.dm.Approver):
		return fmt.Errorf("TxUpdateDataKeepersByAuth.Verify failed: invalid signature for data approver")
	}
	return nil
}

func (tx *TxUpdateDataKeepersByAuth) Accept(bctx BlockContext, bs BlockState) error {
	var err error

	tx.dm.Version++
	tx.dm.KSig = util.Signature{}
	tx.dm.Threshold = tx.from.Threshold()
	tx.dm.Keepers = tx.from.Keepers()
	if len(tx.dm.Keepers) == 0 {
		tx.dm.Threshold = 1
		tx.dm.Keepers = util.EthIDs{tx.from.id}
	}
	tx.dm.Approver = nil
	tx.dm.ApproveList = nil

	if err = bs.SaveData(*tx.input.ID, tx.dm); err != nil {
		return err
	}
	return tx.TxBase.Accept(bctx, bs)
}
