// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package transaction

import (
	"encoding/json"
	"fmt"
	"math/big"
	"strconv"

	"github.com/ldclabs/ldvm/constants"
	"github.com/ldclabs/ldvm/ld"
	"github.com/ldclabs/ldvm/util"
)

type TxCreateModel struct {
	TxBase
	mm *ld.ModelMeta
}

func (tx *TxCreateModel) MarshalJSON() ([]byte, error) {
	if tx == nil || tx.ld == nil {
		return []byte("null"), nil
	}
	v := tx.ld.Copy()
	if tx.mm == nil {
		return nil, fmt.Errorf("MarshalJSON failed: data not exists")
	}
	d, err := json.Marshal(tx.mm)
	if err != nil {
		return nil, err
	}
	v.Data = d
	return json.Marshal(v)
}

func (tx *TxCreateModel) SyntacticVerify() error {
	var err error
	if err = tx.TxBase.SyntacticVerify(); err != nil {
		return err
	}

	if tx.ld.Token != nil {
		return fmt.Errorf("invalid token, expected NativeToken, got %s",
			strconv.Quote(tx.ld.Token.GoString()))
	}
	if len(tx.ld.Data) == 0 {
		return fmt.Errorf("TxCreateModel invalid")
	}
	tx.mm = &ld.ModelMeta{}
	if err = tx.mm.Unmarshal(tx.ld.Data); err != nil {
		return fmt.Errorf("TxCreateModel unmarshal data failed: %v", err)
	}
	if err = tx.mm.SyntacticVerify(); err != nil {
		return fmt.Errorf("TxCreateModel SyntacticVerify failed: %v", err)
	}
	return nil
}

// VerifyGenesis skipping signature verification
func (tx *TxCreateModel) VerifyGenesis(bctx BlockContext, bs BlockState) error {
	var err error
	tx.mm = &ld.ModelMeta{}
	if err = tx.mm.Unmarshal(tx.ld.Data); err != nil {
		return fmt.Errorf("TxCreateModel unmarshal data failed: %v", err)
	}
	if err = tx.mm.SyntacticVerify(); err != nil {
		return fmt.Errorf("TxCreateModel SyntacticVerify failed: %v", err)
	}

	tx.tip = new(big.Int)
	tx.fee = new(big.Int)
	tx.cost = new(big.Int)

	if tx.ldc, err = bs.LoadAccount(constants.LDCAccount); err != nil {
		return err
	}
	if tx.miner, err = bs.LoadMiner(bctx.Miner()); err != nil {
		return err
	}
	tx.from, err = bs.LoadAccount(tx.ld.From)
	return err
}

func (tx *TxCreateModel) Accept(bctx BlockContext, bs BlockState) error {
	var err error
	if err = bs.SaveModel(util.ModelID(tx.ld.ShortID()), tx.mm); err != nil {
		return err
	}
	return tx.TxBase.Accept(bctx, bs)
}
