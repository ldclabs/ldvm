// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package transaction

import (
	"math/big"

	"github.com/ldclabs/ldvm/genesis"
	"github.com/ldclabs/ldvm/ld"
	"github.com/ldclabs/ldvm/util"
)

type BlockContext interface {
	Chain() *genesis.ChainConfig
	FeeConfig() *genesis.FeeConfig
	GasPrice() *big.Int
	Miner() util.StakeSymbol
}

type BlockState interface {
	LoadAccount(util.EthID) (*Account, error)
	LoadMiner(util.StakeSymbol) (*Account, error)
	ResolveNameID(name string) (util.DataID, error)
	ResolveName(name string) (*ld.DataMeta, error)
	SetName(name string, id util.DataID) error
	LoadModel(util.ModelID) (*ld.ModelMeta, error)
	SaveModel(util.ModelID, *ld.ModelMeta) error
	LoadData(util.DataID) (*ld.DataMeta, error)
	SaveData(util.DataID, *ld.DataMeta) error
	SavePrevData(util.DataID, *ld.DataMeta) error
	DeleteData(util.DataID, *ld.DataMeta) error
}
