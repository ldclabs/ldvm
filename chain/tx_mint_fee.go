// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package chain

import (
	"fmt"
	"math/big"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ldclabs/ldvm/constants"
	"github.com/ldclabs/ldvm/ld"
)

// TxMintFee can't be issued from external environment
// There is only one TxMintFee in a block
type TxMintFee struct {
	ld      *ld.Transaction
	from    *Account
	to      *Account
	mintFee *big.Int
}

func (tx *TxMintFee) MarshalJSON() ([]byte, error) {
	if tx == nil {
		return ld.Null, nil
	}

	return tx.ld.MarshalJSON()
}

func (tx *TxMintFee) ID() ids.ID {
	return tx.ld.ID()
}

func (tx *TxMintFee) Type() ld.TxType {
	return tx.ld.Type
}

func (tx *TxMintFee) Bytes() []byte {
	return tx.ld.Bytes()
}

func (tx *TxMintFee) SyntacticVerify() error {
	if tx.ld.Nonce != 0 ||
		tx.ld.Gas != 0 ||
		tx.ld.GasTip != 0 ||
		tx.ld.GasFeeCap != 0 ||
		tx.ld.Amount == nil ||
		tx.ld.From != constants.GenesisAddr ||
		tx.ld.To == constants.BlackholeAddr ||
		tx.ld.To == constants.GenesisAddr ||
		len(tx.ld.Data) == 0 ||
		len(tx.ld.Signatures) != 0 ||
		len(tx.ld.ExSignatures) != 0 {
		return fmt.Errorf("invalid TxMintFee")
	}
	return nil
}

func (tx *TxMintFee) NodeID() (id ids.ShortID, err error) {
	switch {
	case len(tx.ld.Data) != 52:
		err = fmt.Errorf("TxMintFee invalid data")
	default:
		copy(id[:], tx.ld.Data)
	}
	return
}

func (tx *TxMintFee) BlockID() (id ids.ID, err error) {
	switch {
	case len(tx.ld.Data) != 52:
		err = fmt.Errorf("TxMintFee invalid data")
	default:
		copy(id[:], tx.ld.Data[20:])
	}
	return
}

func (tx *TxMintFee) Verify(blk *Block) error {
	bs := blk.State()
	var err error
	if err = bs.ChainConfig().CheckChainID(tx.ld.ChainID); err != nil {
		return err
	}

	if tx.ID() != blk.txs[0].ID() {
		return fmt.Errorf("TxMintFee not matching, expected %s, got %s",
			blk.txs[0].ID(), tx.ID())
	}

	parent := blk.State().PreferredBlock()
	parentID, err := tx.BlockID()
	if err != nil || parentID != parent.ID() {
		return fmt.Errorf("TxMintFee invalid data, expected %s, got %s",
			parent.ID(), parentID)
	}
	if tx.ld.Nonce != parent.Height() {
		return fmt.Errorf("TxMintFee invalid nonce, expected %d, got %d",
			parent.Height(), tx.ld.Nonce)
	}
	if tx.ld.Amount.Cmp(parent.FeeCost()) != 0 {
		return fmt.Errorf("TxMintFee invalid amount, expected %v, got %v",
			parent.FeeCost(), tx.ld.Amount)
	}

	tx.to, err = bs.LoadAccount(tx.ld.To)
	if err != nil {
		return err
	}
	feeCfg := bs.FeeConfig()
	if tx.to.Balance().Cmp(feeCfg.MinMinerStake) < 0 {
		return fmt.Errorf("miner should stake more than %d", feeCfg.MinMinerStake)
	}

	tx.from, err = bs.LoadAccount(tx.ld.From)
	if err != nil {
		return err
	}
	tx.mintFee = blk.GasRebate20()
	if tx.from.Balance().Cmp(tx.mintFee) < 0 {
		return fmt.Errorf("TxMintFee insufficient genesis account balance, expected %v, got %v",
			tx.mintFee, tx.from.Balance())
	}
	return nil
}

func (tx *TxMintFee) VerifyGenesis(blk *Block) error {
	var err error
	bs := blk.State()
	tx.mintFee = tx.ld.Amount
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

func (tx *TxMintFee) Accept(blk *Block) error {
	if tx.mintFee == nil || tx.mintFee.Sign() <= 0 {
		return nil
	}

	blk.State().Log().Info("before from: %v\nto: %v", tx.from.Balance(), tx.to.Balance())

	var err error
	if err = tx.from.Sub(tx.mintFee); err != nil {
		return err
	}
	if err = tx.to.Add(tx.mintFee); err != nil {
		return err
	}
	blk.State().Log().Info("after from: %v\nto: %v", tx.from.Balance(), tx.to.Balance())
	return nil
}

func (tx *TxMintFee) Event(ts int64) *Event {
	return nil
}
