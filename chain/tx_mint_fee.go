// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package chain

import (
	"fmt"
	"math/big"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ldclabs/ldvm/ld"
)

// TxMintFee can't be issued from external environment
// There is only one TxMintFee in a block
type TxMintFee struct {
	ld *ld.Transaction
}

func (tx *TxMintFee) ID() ids.ID {
	return tx.ld.ID()
}

func (tx *TxMintFee) Type() ld.TxType {
	return tx.ld.Type
}

func (tx *TxMintFee) Gas() *big.Int {
	return new(big.Int).SetUint64(tx.ld.Gas)
}

func (tx *TxMintFee) Bytes() []byte {
	return tx.ld.Bytes()
}

func (tx *TxMintFee) SyntacticVerify() error {
	if tx.ld.AccountNonce != 0 ||
		tx.ld.Gas != 0 ||
		tx.ld.GasTipCap != 0 ||
		tx.ld.GasFeeCap != 0 ||
		tx.ld.Amount != nil ||
		tx.ld.From == ids.ShortEmpty ||
		tx.ld.To != ids.ShortEmpty ||
		len(tx.ld.Data) != 0 ||
		len(tx.ld.Signatures) != 0 ||
		len(tx.ld.ExSignatures) != 0 {
		return fmt.Errorf("invalid TxMintFee")
	}
	return nil
}

func (tx *TxMintFee) Verify(blk *Block) error {
	chainCfg := blk.State().ChainConfig()
	if tx.ld.ChainID != chainCfg.ChainID {
		return fmt.Errorf("invalid ChainID %d, expected %d", tx.ld.ChainID, chainCfg.ChainID)
	}

	for _, o := range blk.txs {
		if o.Type() == ld.TypeMintFee && o.ID() != tx.ID() {
			return fmt.Errorf("one more TxMintFee %s exists", o.ID())
		}
	}

	account, err := blk.State().LoadAccount(tx.ld.From)
	if err != nil {
		return err
	}
	feeCfg := blk.State().FeeConfig()
	if account.Balance().Cmp(feeCfg.MinMinerStake) < 1 {
		return fmt.Errorf("miner state should > %d", feeCfg.MinMinerStake)
	}
	return nil
}

func (tx *TxMintFee) Accept(blk *Block) error {
	return nil
}
