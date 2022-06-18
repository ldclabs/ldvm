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
	dm        *ld.DataInfo
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
	errPrefix := "TxUpdateDataKeepersByAuth.SyntacticVerify failed:"
	if err = tx.TxBase.SyntacticVerify(); err != nil {
		return fmt.Errorf("%s %v", errPrefix, err)
	}

	switch {
	case tx.ld.To == nil:
		return fmt.Errorf("%s nil to", errPrefix)
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
	case tx.input.ID == nil || *tx.input.ID == util.DataIDEmpty:
		return fmt.Errorf("%s invalid data id", errPrefix)

	case tx.input.Version == 0:
		return fmt.Errorf("%s invalid data version", errPrefix)

	case tx.input.Keepers != nil:
		return fmt.Errorf("%s invalid keepers, should be nil", errPrefix)

	case tx.input.KSig != nil:
		return fmt.Errorf("%s invalid kSig, should be nil", errPrefix)

	case tx.input.Approver != nil:
		return fmt.Errorf("%s invalid approver, should be nil", errPrefix)

	case tx.input.ApproveList != nil:
		return fmt.Errorf("%s invalid approveList, should be nil", errPrefix)

	case tx.input.To == nil:
		return fmt.Errorf("%s nil to", errPrefix)

	case *tx.input.To != *tx.ld.To:
		return fmt.Errorf("%s invalid to, expected %s, got %s",
			errPrefix, tx.input.To, tx.ld.To)

	case tx.input.Amount == nil:
		return fmt.Errorf("%s nil amount", errPrefix)

	case tx.input.Amount.Cmp(tx.amount) != 0:
		return fmt.Errorf("%s invalid amount, expected %v, got %v",
			errPrefix, tx.input.Amount, tx.amount)

	case tx.input.Token == nil && tx.token != constants.NativeToken:
		return fmt.Errorf("%s invalid token, expected NativeToken, got %s",
			errPrefix, tx.token.GoString())

	case tx.input.Token != nil && tx.token != *tx.input.Token:
		return fmt.Errorf("%s invalid token, expected %s, got %s",
			errPrefix, tx.input.Token.GoString(), tx.token.GoString())
	}

	if tx.exSigners, err = tx.ld.ExSigners(); err != nil {
		return fmt.Errorf("%s invalid exSignatures: %v", errPrefix, err)
	}
	return nil
}

func (tx *TxUpdateDataKeepersByAuth) Verify(bctx BlockContext, bs BlockState) error {
	var err error
	errPrefix := "TxUpdateDataKeepersByAuth.Verify failed:"
	if err = tx.TxBase.Verify(bctx, bs); err != nil {
		return fmt.Errorf("%s %v", errPrefix, err)
	}

	tx.dm, err = bs.LoadData(*tx.input.ID)
	switch {
	case err != nil:
		return fmt.Errorf("%s %v", errPrefix, err)

	case tx.dm.Version != tx.input.Version:
		return fmt.Errorf("%s invalid version, expected %d, got %d",
			errPrefix, tx.dm.Version, tx.input.Version)

	case !util.SatisfySigningPlus(tx.dm.Threshold, tx.dm.Keepers, tx.exSigners):
		return fmt.Errorf("%s invalid exSignatures for data keepers", errPrefix)

	case tx.ld.NeedApprove(tx.dm.Approver, tx.dm.ApproveList) && !tx.exSigners.Has(*tx.dm.Approver):
		return fmt.Errorf("%s invalid signature for data approver", errPrefix)
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
