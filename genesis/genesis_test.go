// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package genesis

import (
	"encoding/json"
	"math/big"
	"os"
	"testing"

	jsonpatch "github.com/ldclabs/json-patch"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ldclabs/ldvm/ids"
	"github.com/ldclabs/ldvm/ld"
	"github.com/ldclabs/ldvm/signer"
	"github.com/ldclabs/ldvm/util/encoding"
)

func TestGenesis(t *testing.T) {
	assert := assert.New(t)

	address1 := signer.Signer1.Key().Address()
	address2 := signer.Signer2.Key().Address()

	file, err := os.ReadFile("./genesis_sample.json")
	require.NoError(t, err)

	gs, err := FromJSON(file)
	require.NoError(t, err)

	assert.Equal(uint64(2357), gs.Chain.ChainID)
	assert.Equal(0, gs.Chain.MaxTotalSupply.Cmp(big.NewInt(1000000000000000000)))
	assert.Equal("Hello, LDVM!", gs.Chain.Message)

	assert.Equal(0, len(gs.Chain.FeeConfigs))
	assert.Equal(uint64(0), gs.Chain.FeeConfig.StartHeight)
	assert.Equal(uint64(10000), gs.Chain.FeeConfig.MinGasPrice)
	assert.Equal(uint64(100000), gs.Chain.FeeConfig.MaxGasPrice)
	assert.Equal(uint64(4200000), gs.Chain.FeeConfig.MaxTxGas)
	assert.Equal(uint64(1000), gs.Chain.FeeConfig.GasRebateRate)
	assert.Equal(0, gs.Chain.FeeConfig.MinTokenPledge.Cmp(big.NewInt(10000000000000)))
	assert.Equal(0, gs.Chain.FeeConfig.MinStakePledge.Cmp(big.NewInt(1000000000000)))

	alloc1 := gs.Alloc[ids.GenesisAccount]
	assert.Equal(0, alloc1.Balance.Cmp(big.NewInt(400000000000000000)))
	assert.Equal(uint16(2), alloc1.Threshold)
	assert.True(alloc1.Keepers.HasAddress(address1))
	assert.True(alloc1.Keepers.HasAddress(address2))

	alloc2 := gs.Alloc[address1]
	assert.Equal(0, alloc2.Balance.Cmp(big.NewInt(100000000000000000)))
	assert.Equal(uint16(1), alloc2.Threshold)
	assert.True(alloc2.Keepers.HasAddress(address1))
	assert.True(alloc2.Keepers.HasAddress(address2))

	_, err = gs.Chain.AppendFeeConfig([]byte{})
	assert.ErrorContains(err, "ChainConfig.AppendFeeConfig: EOF")

	_, err = gs.Chain.AppendFeeConfig(encoding.MustMarshalCBOR(map[string]any{
		"sh":  0,
		"min": 10000,
		"max": 100000,
		"mtg": 42000000,
		"grr": 1000,
		"mtp": 1000000,
		"msp": 1000000,
	}))
	assert.ErrorContains(err, "invalid minTokenPledge")

	_, err = gs.Chain.AppendFeeConfig(encoding.MustMarshalCBOR(map[string]any{
		"sh":  1000,
		"min": 10000,
		"max": 100000,
		"mtg": 42000000,
		"grr": 1000,
		"mtp": 10000000000000,
		"msp": 1000000000000,
		"ntb": 1000000000,
	}))
	assert.ErrorContains(err, "nil builders")

	_, err = gs.Chain.AppendFeeConfig(encoding.MustMarshalCBOR(map[string]any{
		"sh":  1000,
		"min": 10000,
		"max": 100000,
		"mtg": 42000000,
		"grr": 1000,
		"mtp": 10000000000000,
		"msp": 1000000000000,
		"ntb": 1000000000,
		"bs":  ids.IDList[ids.StakeSymbol]{},
	}))
	require.NoError(t, err)

	_, err = gs.Chain.AppendFeeConfig(encoding.MustMarshalCBOR(map[string]any{
		"sh":  100,
		"min": 10000,
		"max": 100000,
		"mtg": 42000000,
		"grr": 1000,
		"mtp": 10000000000000,
		"msp": 1000000000000,
		"ntb": 1000000000,
		"bs":  ids.IDList[ids.StakeSymbol]{},
	}))
	require.NoError(t, err)

	txs, err := gs.ToTxs()
	require.NoError(t, err)
	assert.Equal("SInVTCakmkru9ymxdaIrR4R0S0i_9El1p39szAQtvqP414se", gs.Chain.FeeConfigID.String())
	assert.Equal("b8onI5zOwqPZO9jxMBBgZWnnCUzd-187", gs.Chain.NameServiceID.String())
	assert.True(gs.Chain.IsNameService(gs.Chain.NameServiceID))

	jsondata, err := json.Marshal(txs)
	require.NoError(t, err)

	file, err = os.ReadFile("./genesis_sample_txs.json")
	require.NoError(t, err)
	// fmt.Println(string(jsondata))
	assert.True(jsonpatch.Equal(jsondata, file))

	cbordata, err := txs.Marshal()
	require.NoError(t, err)
	txs2 := ld.Txs{}
	assert.NoError(txs2.Unmarshal(cbordata))
	cbordata2, err := txs2.Marshal()
	require.NoError(t, err)
	assert.Equal(cbordata, cbordata2)
}
