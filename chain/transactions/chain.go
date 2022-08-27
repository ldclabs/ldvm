// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package transactions

import (
	"math/big"

	"github.com/ldclabs/ldvm/genesis"
	"github.com/ldclabs/ldvm/ld"
	"github.com/ldclabs/ldvm/ld/service"
	"github.com/ldclabs/ldvm/util"
)

type ChainContext interface {
	ChainConfig() *genesis.ChainConfig
	FeeConfig() *genesis.FeeConfig
	GasPrice() *big.Int
	Miner() util.StakeSymbol
}

type ChainState interface {
	Height() uint64
	Timestamp() uint64
	LoadAccount(util.EthID) (*Account, error)
	LoadLedger(*Account) error
	LoadMiner(util.StakeSymbol) (*Account, error)
	LoadModel(util.ModelID) (*ld.ModelInfo, error)
	SaveModel(*ld.ModelInfo) error
	LoadData(util.DataID) (*ld.DataInfo, error)
	SaveData(*ld.DataInfo) error
	SavePrevData(*ld.DataInfo) error
	DeleteData(*ld.DataInfo, []byte) error
	SaveName(*service.Name) error
}