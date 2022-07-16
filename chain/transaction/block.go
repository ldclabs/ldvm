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
	ChainConfig() *genesis.ChainConfig
	FeeConfig() *genesis.FeeConfig
	GasPrice() *big.Int
	Miner() util.StakeSymbol
}

type BlockState interface {
	Height() uint64
	Timestamp() uint64
	LoadAccount(util.EthID) (*Account, error)
	LoadLedger(*Account) error
	LoadMiner(util.StakeSymbol) (*Account, error)
	ResolveNameID(name string) (util.DataID, error)
	ResolveName(name string) (*ld.DataInfo, error)
	SetName(name string, id util.DataID) error
	LoadModel(util.ModelID) (*ld.ModelInfo, error)
	SaveModel(util.ModelID, *ld.ModelInfo) error
	LoadData(util.DataID) (*ld.DataInfo, error)
	SaveData(util.DataID, *ld.DataInfo) error
	SavePrevData(util.DataID, *ld.DataInfo) error
	DeleteData(util.DataID, *ld.DataInfo, []byte) error
}
