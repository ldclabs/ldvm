//go:build test && e2e

// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package rpcapi

import (
	"context"
	"testing"

	"github.com/ldclabs/ldvm/ids"
	"github.com/ldclabs/ldvm/rpc/httprpc"
	"github.com/ldclabs/ldvm/util/encoding"
	"github.com/ldclabs/ldvm/vm"

	"github.com/stretchr/testify/assert"
)

// should run after local VM server started.
// should update chainEndpoint
const chainEndpoint = "http://127.0.0.1:65506/ext/bc/2LtxspSnut8ytae5duaMKjttmuqSam5jhqk99aTjeDXYAaei3v/rpc"

func TestChainAPI(t *testing.T) {
	assert := assert.New(t)
	cli := httprpc.NewJSONClient(chainEndpoint, nil)

	genesisBlockID, err := ids.ID32FromStr("zX9qnJZAP9zTEmTooYx1gQsiqOsxFiACh3S1clRYKuvgSiVy")
	assert.NoError(err)

	version := map[string]string{}
	res := cli.Request(context.Background(), "version", nil, &version)
	if res.Error != nil {
		t.Fatal(res.Error.Error())
	}
	assert.Equal(vm.Name, version["name"])
	assert.Equal(vm.Version.String(), version["version"])

	{
		var result map[string]any
		res := cli.Request(context.Background(), "info", nil, &result)
		if assert.Nil(res.Error) {
			// fmt.Println(string(encoding.MustMarshalJSON(result)))
			assert.Equal(`{"builderID":"#WN5NAOFSNZKD2LDHKUB","networkID":1337,"nodeID":"NodeID-7Xhw2mDxuDS44j42TCB6U5579esbSt3Lg","state":"Normal operations state","subnetID":"p433wpuXyJiDhyazPYyZMJeaoPSW76CBZ2x7wrVPLgvokotXz"}`,
				string(encoding.MustMarshalJSON(result)))
		}
	}

	{
		var result uint64
		res := cli.Request(context.Background(), "chainID", nil, &result)
		if assert.Nil(res.Error) {
			if !assert.Equal(uint64(2357), result) {
				t.Logf("Actual: %v", result)
			}
		}
	}

	{
		var result uint8
		res := cli.Request(context.Background(), "chainState", nil, &result)
		if assert.Nil(res.Error) {
			if !assert.Equal(uint8(3), result) {
				t.Logf("Actual: %v", result)
			}
		}
	}

	{
		var result ids.ID32
		res := cli.Request(context.Background(), "lastAccepted", nil, &result)
		if assert.Nil(res.Error) {
			if !assert.Equal(genesisBlockID.String(), result.String()) {
				t.Logf("Actual: %v", result.String())
			}
		}
	}

	{
		var result uint64
		res := cli.Request(context.Background(), "lastAcceptedHeight", nil, &result)
		if assert.Nil(res.Error) {
			if !assert.Equal(uint64(0), result) {
				t.Logf("Actual: %v", result)
			}
		}
	}
}
