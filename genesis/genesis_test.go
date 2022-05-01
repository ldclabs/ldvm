// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package genesis

import (
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ldclabs/ldvm/util"
)

func TestGenesis(t *testing.T) {
	assert := assert.New(t)
	address1 := util.EthID(util.Signer1.Address())
	address2 := util.EthID(util.Signer2.Address())

	file := "./genesis_sample.json"
	data, err := os.ReadFile(file)
	assert.Nil(err)

	gs, err := FromJSON(data)
	assert.Nil(err)

	assert.Equal(uint64(2357), gs.Chain.ChainID)
	assert.Equal(0, gs.Chain.MaxTotalSupply.Cmp(big.NewInt(1000000000000000000)))
	assert.Equal(address2, gs.Alloc[address1].Keepers[1])

	blk, err := gs.ToBlock()
	assert.Nil(err)

	data, err = json.Marshal(blk)
	assert.Nil(err)
	fmt.Printf("\n%s\n", string(data))
	// t.Fatalf("finish")
}
