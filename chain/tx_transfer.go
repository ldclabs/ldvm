// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package chain

import (
	"fmt"
	"math/big"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/snow/choices"
	"github.com/ldclabs/ldvm/constants"
	"github.com/ldclabs/ldvm/ld"
)

type TxTransfer struct {
	ld          *ld.Transaction
	from        *Account
	to          *Account
	genesisAddr *Account
	signers     []ids.ShortID
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

func (tx *TxTransfer) Status() string {
	return tx.ld.Status.String()
}

func (tx *TxTransfer) SetStatus(s choices.Status) {
	tx.ld.Status = s
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
	bs := blk.State()
	tx.from, err = verifyBase(blk, tx.ld, tx.signers)
	if err != nil {
		return err
	}
	if tx.genesisAddr, err = bs.LoadAccount(constants.GenesisAddr); err != nil {
		return err
	}
	tx.to, err = bs.LoadAccount(tx.ld.To)
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
	if tx.genesisAddr, err = bs.LoadAccount(constants.GenesisAddr); err != nil {
		return err
	}
	tx.to, err = bs.LoadAccount(tx.ld.To)
	return err
}

func (tx *TxTransfer) Accept(blk *Block) error {
	var err error
	fee := new(big.Int).Mul(tx.ld.BigIntGas(), blk.GasPrice())
	cost := new(big.Int).Add(tx.ld.Amount, fee)
	if err = tx.from.SubByNonce(tx.ld.Nonce, cost); err != nil {
		return err
	}
	if err = tx.genesisAddr.Add(fee); err != nil {
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
