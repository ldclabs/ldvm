// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package txn

import (
	"math/big"

	"github.com/ldclabs/ldvm/chain/acct"
	"github.com/ldclabs/ldvm/genesis"
	"github.com/ldclabs/ldvm/ids"
	"github.com/ldclabs/ldvm/ld"
	"github.com/ldclabs/ldvm/ld/service"
)

type ChainContext interface {
	ChainConfig() *genesis.ChainConfig
	FeeConfig() *genesis.FeeConfig
	GasPrice() *big.Int
	Builder() ids.Address
}

type ChainState interface {
	Height() uint64
	Timestamp() uint64
	LoadAccount(ids.Address) (*acct.Account, error)
	LoadLedger(*acct.Account) error
	LoadModel(ids.ModelID) (*ld.ModelInfo, error)
	SaveModel(*ld.ModelInfo) error
	LoadData(ids.DataID) (*ld.DataInfo, error)
	SaveData(*ld.DataInfo) error
	SavePrevData(*ld.DataInfo) error
	DeleteData(*ld.DataInfo, []byte) error
	SaveName(*service.Name) error
	DeleteName(*service.Name) error
}
