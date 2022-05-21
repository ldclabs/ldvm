// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package chain

import (
	"encoding/json"
	"fmt"
	"math/big"
	"strconv"

	"github.com/ldclabs/ldvm/constants"
	"github.com/ldclabs/ldvm/ld"
)

type TxUpdateAccountKeepers struct {
	TxBase
	data *ld.TxAccounter
}

func (tx *TxUpdateAccountKeepers) MarshalJSON() ([]byte, error) {
	if tx == nil || tx.ld == nil {
		return []byte("null"), nil
	}
	v := tx.ld.Copy()
	if tx.data == nil {
		return nil, fmt.Errorf("MarshalJSON failed: data not exists")
	}
	d, err := json.Marshal(tx.data)
	if err != nil {
		return nil, err
	}
	v.Data = d
	return json.Marshal(v)
}

// VerifyGenesis skipping signature verification
func (tx *TxUpdateAccountKeepers) SyntacticVerify() error {
	var err error
	if err = tx.TxBase.SyntacticVerify(); err != nil {
		return err
	}

	if tx.ld.Token != nil {
		return fmt.Errorf("invalid token, expected NativeToken, got %s",
			strconv.Quote(tx.ld.Token.GoString()))
	}
	if len(tx.ld.Data) == 0 {
		return fmt.Errorf("TxUpdateAccountKeepers invalid")
	}
	tx.data = &ld.TxAccounter{}
	if err = tx.data.Unmarshal(tx.ld.Data); err != nil {
		return fmt.Errorf("TxUpdateAccountKeepers unmarshal data failed: %v", err)
	}
	if err = tx.data.SyntacticVerify(); err != nil {
		return fmt.Errorf("TxUpdateAccountKeepers SyntacticVerify failed: %v", err)
	}
	if len(tx.data.Keepers) == 0 ||
		tx.data.Threshold == 0 {
		return fmt.Errorf("TxUpdateAccountKeepers invalid keepers")
	}
	if len(tx.data.Keepers) == 0 && tx.data.Approver == nil && tx.data.ApproveList == nil {
		return fmt.Errorf("TxUpdateAccountKeepers no keepers nor approver")
	}
	return nil
}

func (tx *TxUpdateAccountKeepers) VerifyGenesis(blk *Block, bs BlockState) error {
	var err error
	tx.data = &ld.TxAccounter{}
	if err = tx.data.Unmarshal(tx.ld.Data); err != nil {
		return fmt.Errorf("TxUpdateAccountKeepers unmarshal data failed: %v", err)
	}
	if err = tx.data.SyntacticVerify(); err != nil {
		return fmt.Errorf("TxUpdateAccountKeepers SyntacticVerify failed: %v", err)
	}

	tx.tip = new(big.Int)
	tx.fee = new(big.Int)
	tx.cost = new(big.Int)

	if tx.genesisAcc, err = bs.LoadAccount(constants.GenesisAccount); err != nil {
		return err
	}
	if tx.miner, err = blk.Miner(); err != nil {
		return err
	}
	tx.from, err = bs.LoadAccount(tx.ld.From)
	return err
}

func (tx *TxUpdateAccountKeepers) Verify(blk *Block, bs BlockState) error {
	var err error
	if err = tx.TxBase.Verify(blk, bs); err != nil {
		return err
	}
	if !tx.from.SatisfySigningPlus(tx.signers) {
		return fmt.Errorf("sender account need more signers")
	}
	return nil
}

func (tx *TxUpdateAccountKeepers) Accept(blk *Block, bs BlockState) error {
	var err error
	if err = tx.from.UpdateKeepers(tx.data.Threshold, tx.data.Keepers, tx.data.Approver, tx.data.ApproveList); err != nil {
		return err
	}

	return tx.TxBase.Accept(blk, bs)
}
