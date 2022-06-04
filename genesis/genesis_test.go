// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package genesis

import (
	"encoding/json"
	"math/big"
	"os"
	"testing"

	jsonpatch "github.com/evanphx/json-patch/v5"
	"github.com/stretchr/testify/assert"

	"github.com/ldclabs/ldvm/constants"
	"github.com/ldclabs/ldvm/ld"
	"github.com/ldclabs/ldvm/util"
)

func TestGenesis(t *testing.T) {
	assert := assert.New(t)

	address1 := util.Signer1.Address()
	address2 := util.Signer2.Address()

	file, err := os.ReadFile("./genesis_sample.json")
	assert.NoError(err)

	gs, err := FromJSON(file)
	assert.NoError(err)

	assert.Equal(uint64(2357), gs.Chain.ChainID)
	assert.Equal(0, gs.Chain.MaxTotalSupply.Cmp(big.NewInt(1000000000000000000)))
	assert.Equal("Hello, LDVM!", gs.Chain.Message)

	assert.Equal(0, len(gs.Chain.FeeConfigs))
	assert.Equal(uint64(0), gs.Chain.FeeConfig.StartHeight)
	assert.Equal(uint64(1000), gs.Chain.FeeConfig.ThresholdGas)
	assert.Equal(uint64(10000), gs.Chain.FeeConfig.MinGasPrice)
	assert.Equal(uint64(100000), gs.Chain.FeeConfig.MaxGasPrice)
	assert.Equal(uint64(42000000), gs.Chain.FeeConfig.MaxTxGas)
	assert.Equal(uint64(4200000), gs.Chain.FeeConfig.MaxBlockTxsSize)
	assert.Equal(uint64(1000), gs.Chain.FeeConfig.GasRebateRate)
	assert.Equal(0, gs.Chain.FeeConfig.MinTokenPledge.Cmp(big.NewInt(10000000000000)))
	assert.Equal(0, gs.Chain.FeeConfig.MinStakePledge.Cmp(big.NewInt(1000000000000)))

	alloc1 := gs.Alloc[constants.GenesisAccount]
	assert.Equal(0, alloc1.Balance.Cmp(big.NewInt(400000000000000000)))
	assert.Equal(uint8(2), alloc1.Threshold)
	assert.True(alloc1.Keepers.Has(address1))
	assert.True(alloc1.Keepers.Has(address2))

	alloc2 := gs.Alloc[address1]
	assert.Equal(0, alloc2.Balance.Cmp(big.NewInt(100000000000000000)))
	assert.Equal(uint8(1), alloc2.Threshold)
	assert.True(alloc2.Keepers.Has(address1))
	assert.True(alloc2.Keepers.Has(address2))

	_, err = gs.Chain.AppendFeeConfig([]byte{})
	assert.ErrorContains(err, "JSON input")

	_, err = gs.Chain.AppendFeeConfig([]byte(`{}`))
	assert.ErrorContains(err, "invalid thresholdGas")

	_, err = gs.Chain.AppendFeeConfig([]byte(`{
		"startHeight": 0,
		"thresholdGas": 1000,
		"minGasPrice": 10000,
		"maxGasPrice": 100000,
		"maxTxGas": 42000000,
		"maxBlockTxsSize": 4200000,
		"gasRebateRate": 1000,
		"minTokenPledge": 1000000,
		"minStakePledge": 1000000
	}`))
	assert.ErrorContains(err, "invalid minTokenPledge")

	_, err = gs.Chain.AppendFeeConfig([]byte(`{
		"startHeight": 1000,
		"thresholdGas": 9999,
		"minGasPrice": 10000,
		"maxGasPrice": 100000,
		"maxTxGas": 42000000,
		"maxBlockTxsSize": 4200000,
		"gasRebateRate": 1000,
		"minTokenPledge": 10000000000000,
		"minStakePledge": 1000000000000
	}`))
	assert.NoError(err)
	_, err = gs.Chain.AppendFeeConfig([]byte(`{
		"startHeight": 100,
		"thresholdGas": 88888,
		"minGasPrice": 10000,
		"maxGasPrice": 100000,
		"maxTxGas": 42000000,
		"maxBlockTxsSize": 4200000,
		"gasRebateRate": 1000,
		"minTokenPledge": 10000000000000,
		"minStakePledge": 1000000000000
	}`))
	assert.NoError(err)
	assert.Equal(uint64(1000), gs.Chain.Fee(10).ThresholdGas)
	assert.Equal(uint64(88888), gs.Chain.Fee(100).ThresholdGas)
	assert.Equal(uint64(88888), gs.Chain.Fee(999).ThresholdGas)
	assert.Equal(uint64(9999), gs.Chain.Fee(1000).ThresholdGas)
	assert.Equal(uint64(9999), gs.Chain.Fee(10000).ThresholdGas)

	txs, err := gs.ToTxs()
	assert.NoError(err)
	assert.Equal("LDBu29pdwT1xK2CZr8DCVx5NqvLiSUzwzEG", gs.Chain.FeeConfigID.String())
	assert.Equal("LM4rB4RoU8Xa2FAJRVAER8bcprHcpAYFRBs", gs.Chain.NameServiceID.String())
	assert.Equal("LMDWuG2ggqziTRsZRvVwCf5W9Vr6j1QqWNt", gs.Chain.ProfileServiceID.String())
	assert.True(gs.Chain.IsNameService(gs.Chain.NameServiceID))

	jsondata, err := json.Marshal(txs)
	assert.NoError(err)

	file, err = os.ReadFile("./genesis_sample_txs.json")
	assert.NoError(err)
	// fmt.Println(string(jsondata))
	assert.True(jsonpatch.Equal(jsondata, file))

	cbordata, err := txs.Marshal()
	assert.NoError(err)
	txs2 := ld.Txs{}
	assert.NoError(txs2.Unmarshal(cbordata))
	cbordata2, err := txs2.Marshal()
	assert.NoError(err)
	assert.Equal(cbordata, cbordata2)
}
