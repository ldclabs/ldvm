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
	from    *Account
	to      *Account
	signers []ids.ShortID
}

func (tx *TxTransfer) MarshalJSON() ([]byte, error) {
	if tx == nil {
		return ld.Null, nil
	}
	return tx.ld.MarshalJSON()
}

func (tx *TxTransfer) ID() ids.ID {
	return tx.ld.ID()
}

func (tx *TxTransfer) Type() ld.TxType {
	return tx.ld.Type
}

func (tx *TxTransfer) Bytes() []byte {
	return tx.ld.Bytes()
}

func (tx *TxTransfer) SyntacticVerify() error {
	if tx == nil {
		return fmt.Errorf("invalid TxTransfer")
	}

	if tx.ld.Amount.Sign() <= 0 {
		return fmt.Errorf("invalid amount")
	}
	return nil
}

func (tx *TxTransfer) Verify(blk *Block) error {
	var err error
	tx.signers, err = ld.DeriveSigners(tx.ld.UnsignedBytes(), tx.ld.Signatures)
	if err != nil {
		return fmt.Errorf("invalid signatures: %v", err)
	}
	tx.from, err = verifyBase(blk, tx.ld, tx.signers)
	if err != nil {
		return err
	}
	tx.to, err = blk.State().LoadAccount(tx.ld.To)
	return err
}

// VerifyGenesis skipping signature verification
func (tx *TxTransfer) VerifyGenesis(blk *Block) error {
	var err error
	bs := blk.State()
	tx.from, err = bs.LoadAccount(tx.ld.From)
	if err != nil {
		return err
	}
	tx.to, err = bs.LoadAccount(tx.ld.To)
	if err != nil {
		return err
	}
	return nil
}

func (tx *TxTransfer) Accept(blk *Block) error {
	var err error
	cost := new(big.Int).Mul(tx.ld.BigIntGas(), blk.GasPrice())
	cost = new(big.Int).Add(tx.ld.Amount, cost)
	if err = tx.from.SubByNonce(tx.ld.Nonce, cost); err != nil {
		return err
	}
	if err = tx.to.Add(tx.ld.Amount); err != nil {
		return err
	}
	return nil
}

func (tx *TxTransfer) Event(ts int64) *Event {
	return nil
}
