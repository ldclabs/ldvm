//go:build test && e2e

// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package rpcapi

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ldclabs/ldvm/genesis"
	"github.com/ldclabs/ldvm/rpc/httprpc"
	"github.com/ldclabs/ldvm/util/encoding"
	"github.com/ldclabs/ldvm/vm"
)

// should run after local VM server started.
// should update vmEndpoint
const vmEndpoint = "http://127.0.0.1:65506/ext/vm/pjjsfTNAgQnP7zdpKfRcmicXGbk87xXznJmJZtqDAyRaNEhEL/rpc"

func TestVMAPI(t *testing.T) {
	assert := assert.New(t)
	cli := httprpc.NewJSONClient(vmEndpoint, nil)

	genesisJSON := `{"chain":{"chainID":2357,"maxTotalSupply":1000000000000000000,"message":"Hello, LDVM!","feeConfig":{"startHeight":0,"minGasPrice":10000,"maxGasPrice":100000,"maxTxGas":42000000,"gasRebateRate":1000,"minTokenPledge":10000000000000,"minStakePledge":1000000000000,"nonTransferableBalance":1000000000,"builders":["#WN5NAOFSNZKD2LDHKUB"]},"feeConfigs":[],"feeConfigID":"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAACeYpGX","nameServiceID":"AAAAAAAAAAAAAAAAAAAAAAAAAADzaDye"},"alloc":{"0x8db97c7cECe249C2b98bdc0226cc4C2A57bF52fc":{"balance":100000000000000000,"threshold":1,"keepers":["jbl8fOziScK5i9wCJsxMKle_UvwKxwPH","RBccN_9de3u43K1cgfFihKIp5kE1lmGG","hJEADz4AlkZ_NSt41-9x5eTaahzNzgMzd0wOBF-B2kJGSpWTCQutstgl0tXrZKQVIsBdNQ"]},"0xFFfFFFfFfffFFfFFffFFFfFfFffFFFfffFfFFFff":{"balance":400000000000000000,"threshold":2,"keepers":["jbl8fOziScK5i9wCJsxMKle_UvwKxwPH","RBccN_9de3u43K1cgfFihKIp5kE1lmGG","OVlX-75gy0DuaRuz2k5QnlFVSuKOJezRd4CQdkIjkn5pYt0F"]}}}`

	version := map[string]string{}
	res := cli.Request(context.Background(), "ldvm.version", nil, &version)
	if res.Error != nil {
		t.Fatal(res.Error.Error())
	}
	assert.Equal(vm.Name, version["name"])
	assert.Equal(vm.Version.String(), version["version"])

	{
		result := genesis.Genesis{}
		res := cli.Request(context.Background(), "ldvm.genesis", nil, &result)
		if assert.Nil(res.Error) {
			// fmt.Println(string(encoding.MustMarshalJSON(result)))
			assert.Equal(genesisJSON, string(encoding.MustMarshalJSON(result)))
		}
	}
}
