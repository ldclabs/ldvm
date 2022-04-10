// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package chain

import (
	"fmt"
	"math/big"
	"time"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ldclabs/ldvm/ld"
)

type TxTransferCash struct {
	ld        *ld.Transaction
	from      *Account
	issuer    *Account
	signers   []ids.ShortID
	exSigners []ids.ShortID
	data      *ld.TxTransfer
}

func (tx *TxTransferCash) MarshalJSON() ([]byte, error) {
	if tx == nil {
		return ld.Null, nil
	}
	v := tx.ld.Copy()
	if tx.data == nil {
		tx.data = &ld.TxTransfer{}
		if err := tx.data.Unmarshal(tx.ld.Data); err != nil {
			return nil, fmt.Errorf("TxTransferCash unmarshal data failed: %v", err)
		}
	}
	d, err := tx.data.MarshalJSON()
	if err != nil {
		return nil, err
	}
	v.Data = d
	return v.MarshalJSON()
}

func (tx *TxTransferCash) ID() ids.ID {
	return tx.ld.ID()
}

func (tx *TxTransferCash) Type() ld.TxType {
	return tx.ld.Type
}

func (tx *TxTransferCash) Bytes() []byte {
	return tx.ld.Bytes()
}

func (tx *TxTransferCash) SyntacticVerify() error {
	if tx == nil ||
		len(tx.ld.Data) == 0 {
		return fmt.Errorf("invalid TxTransferCash")
	}

	var err error
	tx.data = &ld.TxTransfer{}
	if err = tx.data.Unmarshal(tx.ld.Data); err != nil {
		return fmt.Errorf("TxTransferCash unmarshal data failed: %v", err)
	}
	if err = tx.data.SyntacticVerify(); err != nil {
		return fmt.Errorf("TxTransferCash SyntacticVerify failed: %v", err)
	}
	if tx.data.Nonce == 0 {
		return fmt.Errorf("TxTransferCash invalid nonce")
	}
	if tx.data.To != tx.ld.From {
		return fmt.Errorf("TxTransferCash invalid recipient")
	}
	if tx.data.To != tx.ld.To {
		return fmt.Errorf("TxTransferCash invalid recipient")
	}
	if tx.data.Expire > 0 && tx.data.Expire < uint64(time.Now().Unix()) {
		return fmt.Errorf("TxTransferCash expired")
	}
	if tx.data.Amount != nil && tx.data.Amount.Cmp(tx.ld.Amount) != 0 {
		return fmt.Errorf("TxTransferCash invalid amount")
	}
	return nil
}

func (tx *TxTransferCash) Verify(blk *Block) error {
	var err error
	tx.signers, err = ld.DeriveSigners(tx.ld.UnsignedBytes(), tx.ld.Signatures)
	if err != nil {
		return fmt.Errorf("invalid signatures: %v", err)
	}
	tx.exSigners, err = ld.DeriveSigners(tx.ld.Data, tx.ld.ExSignatures)
	if err != nil {
		return fmt.Errorf("invalid exSignatures: %v", err)
	}

	tx.from, err = verifyBase(blk, tx.ld, tx.signers)
	if err != nil {
		return err
	}
	tx.issuer, err = blk.State().LoadAccount(tx.data.From)
	if err != nil {
		return err
	}
	if tx.issuer.Nonce() != tx.data.Nonce {
		return fmt.Errorf("TxTransferCash invalid issuer nonce")
	}
	// verify issuer's signatures
	if !ld.SatisfySigning(tx.issuer.Threshold(), tx.issuer.Keepers(), tx.exSigners, false) {
		return fmt.Errorf("TxTransferCash need more exSignatures")
	}
	if tx.ld.Amount.Cmp(tx.issuer.Balance()) > 0 {
		return fmt.Errorf("insufficient balance %d of issuer %s, required %d",
			tx.issuer.Balance(), tx.data.From, tx.ld.Amount)
	}
	return err
}

func (tx *TxTransferCash) Accept(blk *Block) error {
	var err error
	cost := new(big.Int).Mul(tx.ld.BigIntGas(), blk.GasPrice())
	if err = tx.from.SubByNonce(tx.ld.Nonce, cost); err != nil {
		return err
	}
	if err = tx.issuer.SubByNonce(tx.data.Nonce, tx.ld.Amount); err != nil {
		return err
	}
	if err = tx.from.Add(tx.ld.Amount); err != nil {
		return err
	}
	return nil
}

func (tx *TxTransferCash) Event(ts int64) *Event {
	return nil
}
