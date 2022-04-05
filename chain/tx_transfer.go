// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package chain

import (
	"fmt"
	"math/big"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ldclabs/ldvm/ld"
)

type TxTransfer struct {
	ld      *ld.Transaction
	id      ids.ID
	signers []ids.ShortID
}

func (tx *TxTransfer) ID() ids.ID {
	return tx.ld.ID()
}

func (tx *TxTransfer) Type() ld.TxType {
	return tx.ld.Type
}

func (tx *TxTransfer) Gas() *big.Int {
	return new(big.Int).SetUint64(tx.ld.Gas)
}

func (tx *TxTransfer) Bytes() []byte {
	return tx.ld.Bytes()
}

func (tx *TxTransfer) SyntacticVerify() error {
	if tx.ld.Gas == 0 ||
		tx.ld.GasFeeCap == 0 ||
		tx.ld.Amount == nil ||
		tx.ld.From == ids.ShortEmpty ||
		tx.ld.To == ids.ShortEmpty ||
		len(tx.ld.Signatures) == 0 ||
		len(tx.ld.ExSignatures) != 0 {
		return fmt.Errorf("invalid TxTransfer")
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
	return err
}

func (tx *TxTransfer) Verify(blk *Block) error {
	chainCfg := blk.State().ChainConfig()
	if tx.ld.ChainID != chainCfg.ChainID {
		return fmt.Errorf("invalid ChainID %d, expected %d", tx.ld.ChainID, chainCfg.ChainID)
	}

	acc, err := blk.State().LoadAccount(tx.ld.From)
	if err != nil {
		return err
	}

	if tx.ld.AccountNonce != acc.Nonce() { // TODO: nonce changed everywhere!
		return fmt.Errorf("account nonce not matching")
	}
	if !acc.SatisfySigning(tx.signers) {
		return fmt.Errorf("need more signatures")
	}
	cost := new(big.Int).Mul(tx.Gas(), blk.GasPrice())
	cost = new(big.Int).Add(tx.ld.Amount, cost)
	if acc.Balance().Cmp(cost) < 0 {
		return fmt.Errorf("insufficient balance %d of account %s, required %d",
			acc.Balance(), tx.ld.From, cost)
	}
	return nil
}

func (tx *TxTransfer) Accept(blk *Block) error {
	acc, err := blk.State().LoadAccount(tx.ld.From)
	if err != nil {
		return err
	}
	to, err := blk.State().LoadAccount(tx.ld.To)
	if err != nil {
		return err
	}

	cost := new(big.Int).Mul(new(big.Int).SetUint64(tx.ld.Gas), blk.GasPrice())
	cost = new(big.Int).Add(tx.ld.Amount, cost)
	if err := acc.Sub(tx.ld.AccountNonce, cost); err != nil {
		return err
	}
	if err := to.Add(to.Nonce(), tx.ld.Amount); err != nil {
		return err
	}
	return nil
}
