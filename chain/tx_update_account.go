// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package chain

import (
	"fmt"
	"math/big"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ldclabs/ldvm/ld"
)

type TxUpdateAccountGuardians struct {
	ld      *ld.Transaction
	id      ids.ID
	signers []ids.ShortID
	update  *ld.TxUpdateAccountGuardians
}

func (tx *TxUpdateAccountGuardians) ID() ids.ID {
	return tx.ld.ID()
}

func (tx *TxUpdateAccountGuardians) Type() ld.TxType {
	return tx.ld.Type
}

func (tx *TxUpdateAccountGuardians) Gas() *big.Int {
	return new(big.Int).SetUint64(tx.ld.Gas)
}

func (tx *TxUpdateAccountGuardians) Bytes() []byte {
	return tx.ld.Bytes()
}

func (tx *TxUpdateAccountGuardians) SyntacticVerify() error {
	if tx.ld.AccountNonce == 0 ||
		tx.ld.Gas == 0 ||
		tx.ld.GasFeeCap == 0 ||
		tx.ld.Amount != nil ||
		tx.ld.From == ids.ShortEmpty ||
		tx.ld.To != ids.ShortEmpty ||
		len(tx.ld.Data) == 0 ||
		len(tx.ld.Signatures) == 0 ||
		len(tx.ld.ExSignatures) != 0 {
		return fmt.Errorf("invalid TxMintFee")
	}

	x := tx.ld.Copy()
	x.Gas = 0
	x.Signatures = nil
	data, err := x.Marshal()
	if err != nil {
		return err
	}
	tx.signers, err = ld.DeriveSigners(data, tx.ld.Signatures)
	if err == nil {
		if tx.ld.From != tx.signers[0] {
			return fmt.Errorf("invalid sender %s, expected %s", tx.ld.From, tx.signers[0])
		}
	}

	tx.update = &ld.TxUpdateAccountGuardians{}
	if err := tx.update.Unmarshal(tx.ld.Data); err != nil {
		return fmt.Errorf("TxUpdateAccountGuardians Unmarshal failed: %v", err)
	}
	if err := tx.update.SyntacticVerify(); err != nil {
		return fmt.Errorf("TxUpdateAccountGuardians SyntacticVerify failed: %v", err)
	}
	return nil
}

func (tx *TxUpdateAccountGuardians) Verify(blk *Block) error {
	chainCfg := blk.State().ChainConfig()
	if tx.ld.ChainID != chainCfg.ChainID {
		return fmt.Errorf("invalid ChainID %d, expected %d", tx.ld.ChainID, chainCfg.ChainID)
	}

	acc, err := blk.State().LoadAccount(tx.ld.From)
	if err != nil {
		return err
	}

	if tx.ld.AccountNonce != acc.Nonce() {
		return fmt.Errorf("account nonce not matching")
	}
	if !acc.SatisfySigning(tx.signers) {
		return fmt.Errorf("need more signatures")
	}
	cost := new(big.Int).Mul(tx.Gas(), blk.GasPrice())
	if acc.Balance().Cmp(cost) < 0 {
		return fmt.Errorf("insufficient balance %d of account %s, required %d",
			acc.Balance(), tx.ld.From, cost)
	}
	return nil
}

func (tx *TxUpdateAccountGuardians) Accept(blk *Block) error {
	acc, err := blk.State().LoadAccount(tx.ld.From)
	if err != nil {
		return err
	}

	cost := new(big.Int).Mul(new(big.Int).SetUint64(tx.ld.Gas), blk.GasPrice())
	if err := acc.UpdateGuardians(tx.ld.AccountNonce, cost, tx.update.Threshold,
		tx.update.Guardians); err != nil {
		return err
	}
	return nil
}
