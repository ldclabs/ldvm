// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package chain

import (
	"fmt"
	"math/big"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/snow/choices"
	"github.com/ldclabs/ldvm/ld"
)

// TxMintFee can't be issued from external environment
// There is only one TxMintFee in a block
type TxMintFee struct {
	ld          *ld.Transaction
	addTime     int64
	from        *Account
	to          *Account
	genesisAddr *Account
	mintFee     *big.Int
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

func (tx *TxMintFee) Status() string {
	return tx.ld.Status.String()
}

func (tx *TxMintFee) SetStatus(s choices.Status) {
	tx.ld.Status = s
}

func (tx *TxMintFee) SyntacticVerify() error {
	if tx == nil {
		return fmt.Errorf("invalid TxMintFee")
	}
	return nil
}

func (tx *TxMintFee) Verify(blk *Block) error {

	var err error
	if err = blk.ctx.Chain().CheckChainID(tx.ld.ChainID); err != nil {
		return err
	}

	if tx.ID() != blk.txs[0].ID() {
		return fmt.Errorf("TxMintFee not matching, expected %s, got %s",
			blk.txs[0].ID(), tx.ID())
	}

	bs := blk.State()
	parent, err := blk.ParentBlock()
	if err != nil {
		return err
	}
	_, parentID, err := unwrapMintTxData(tx.ld.Data)
	if err != nil || parentID != parent.ID() {
		return fmt.Errorf("TxMintFee invalid data, expected %s, got %s",
			parent.ID(), parentID)
	}
	if tx.ld.Nonce != parent.Height {
		return fmt.Errorf("TxMintFee invalid nonce, expected %d, got %d",
			parent.Height, tx.ld.Nonce)
	}
	if tx.ld.Amount.Cmp(parent.FeeCost()) != 0 {
		return fmt.Errorf("TxMintFee invalid amount, expected %v, got %v",
			parent.FeeCost(), tx.ld.Amount)
	}

	tx.to, err = bs.LoadAccount(tx.ld.To)
	if err != nil {
		return err
	}
	feeCfg := blk.FeeConfig()
	if tx.to.Balance().Cmp(new(big.Int).SetUint64(feeCfg.MinMinerStake)) < 0 {
		return fmt.Errorf("miner should stake more than %d", feeCfg.MinMinerStake)
	}

	tx.from, err = bs.LoadAccount(tx.ld.From)
	if err != nil {
		return err
	}
	tx.mintFee = blk.GasRebate20()
	if tx.from.Balance().Cmp(tx.mintFee) < 0 {
		return fmt.Errorf("TxMintFee insufficient balance, expected %v, got %v",
			tx.mintFee, tx.from.Balance())
	}
	return nil
}

// VerifyGenesis skipping signature verification
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
	var err error
	if err = tx.from.Sub(tx.mintFee); err != nil {
		return err
	}
	if err = tx.to.Add(tx.mintFee); err != nil {
		return err
	}
	return nil
}

func (tx *TxMintFee) Event(ts int64) *Event {
	return nil
}
