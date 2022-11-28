//go:build test && e2e

// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package rpcapi

import (
	"context"
	"testing"

	"github.com/ldclabs/ldvm/ids"
	"github.com/ldclabs/ldvm/ld"
	"github.com/ldclabs/ldvm/rpc/httprpc"
	"github.com/ldclabs/ldvm/util/encoding"
	"github.com/ldclabs/ldvm/vm"

	"github.com/stretchr/testify/assert"
)

// should run after local VM server started.
const cborrpcEndpoint = "h2c://localhost:2357/cborrpc/v1"

func TestAPI(t *testing.T) {
	assert := assert.New(t)
	cli := httprpc.NewCBORClient(cborrpcEndpoint, nil)

	genesisBlockID, err := ids.ID32FromStr("zX9qnJZAP9zTEmTooYx1gQsiqOsxFiACh3S1clRYKuvgSiVy")
	assert.NoError(err)

	genesisBlockJSON := `{"parent":"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAACeYpGX","state":"J0ixQfFg5fP19svV_UE3aVLL_K8GD8xASBd92CFSWF4BxqbG","pChainHeight":0,"height":0,"timestamp":0,"gas":10597,"gasPrice":10000,"gasRebateRate":1000,"builder":"","validators":[],"txs":["7YHUWKN9LHEZ_j6jl2WoX_duMMgIQJOu_oPcyQ6HgrK-OluF","IqxPxepqwH8OnSLIHiPGhKCfd7YQ9v8cWRDm32D7_kGYuqOe","crF6I_WYHUZahZi76mf6hnszJTx66ii3a3_u3GMslGWAPjSw","P4XE-FDt5MbB8BTjiC0G13790-jcIhxKmtdfqcKCbyPw-YrE","5_3AMSxPqhb0A_-ntTzJd27bpi7IlXaFiVv_6_5R1OR_OJvu","QouAozxQAthLBGZ7_NCo9ycpN-5Q0eBKVsMxe15w8ROOVDt4","b8onI5zOwqPZO9jxMBBgZWnnCUxZH8Cy-SD0YLXpLJRDy5Wq","-xzGonDgQ_-M5FXiFi3MQGDvEDiks_Tqu1jFa27Z2cu0Zgmx"],"id":"zX9qnJZAP9zTEmTooYx1gQsiqOsxFiACh3S1clRYKuvgSiVy"}`

	genesisStateID, err := ids.ID32FromStr("J0ixQfFg5fP19svV_UE3aVLL_K8GD8xASBd92CFSWF4BxqbG")
	assert.NoError(err)

	genesisStateJSON := `{"parent":"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAACeYpGX","accounts":{"0x0000000000000000000000000000000000000000":"s3eCn6EBZn41eFzXl9u1V26jDL5Os7zMWdFsfCDKDybrYHX9","0x6FcA27239CCEC2A3d93BD8F13010606569E7094C":"idwHWfkOI5eIoQAW2Ua2xaDhojExK1t1xX1azM0Zw_8agDJ8","0x8db97c7cECe249C2b98bdc0226cc4C2A57bF52fc":"8qoXYTpCRDGAOyhNIAB9iAAg5tf7MXz_CI_RU5gJKOXs5sHW","0xFFfFFFfFfffFFfFFffFFFfFfFffFFFfffFfFFFff":"rnnBsv7dzXD68nEQgLB6B5TKn31LM6eMn2UH9ngIlrGlGLLC","0xfb1Cc6A270e043FF8ce455E2162dCc4060EF1038":"idwHWfkOI5eIoQAW2Ua2xaDhojExK1t1xX1azM0Zw_8agDJ8"},"ledgers":{},"datas":{"QouAozxQAthLBGZ7_NCo9ycpN-5Q0eBKVsMxe15w8ROOVDt4":"kneQuMw3IURDNNxTIWKbxVBwYPPC_mghi_4qaHJakfALt6Yr"},"models":{"-xzGonDgQ_-M5FXiFi3MQGDvEDgTw4dJ":"sgQAbRkdomrP9Lwps6bgmPjqXn8ak_oAY4p-hSbp5-y88mHd","b8onI5zOwqPZO9jxMBBgZWnnCUzd-187":"j_buypCMiYayF1xeX9syCB9L1Owa6w4DH3YD0h5isnZrR0rx"},"id":"J0ixQfFg5fP19svV_UE3aVLL_K8GD8xASBd92CFSWF4BxqbG"}`

	genesisTxsJSON := `[{"tx":{"type":"TypeCreateToken","chainID":2357,"nonce":0,"gasTip":0,"gasFeeCap":0,"from":"0xFFfFFFfFfffFFfFFffFFFfFfFffFFFfffFfFFFff","to":"0x0000000000000000000000000000000000000000","data":"o2FhwkgN4Lazp2QAAGFkTiJIZWxsbywgTERWTSEiYW5rTmF0aXZlVG9rZW63HrPF"},"id":"7YHUWKN9LHEZ_j6jl2WoX_duMMgIQJOu_oPcyQ6HgrK-OluF"},{"tx":{"type":"TypeTransfer","chainID":2357,"nonce":0,"gasTip":0,"gasFeeCap":0,"from":"0x0000000000000000000000000000000000000000","to":"0x8db97c7cECe249C2b98bdc0226cc4C2A57bF52fc","amount":100000000000000000},"id":"IqxPxepqwH8OnSLIHiPGhKCfd7YQ9v8cWRDm32D7_kGYuqOe"},{"tx":{"type":"TypeUpdateAccountInfo","chainID":2357,"nonce":0,"gasTip":0,"gasFeeCap":0,"from":"0x8db97c7cECe249C2b98bdc0226cc4C2A57bF52fc","data":"omJrcINUjbl8fOziScK5i9wCJsxMKle_UvxURBccN_9de3u43K1cgfFihKIp5kFYIL5VNr0Ot3044q0VrSSb1XJ3FZ8gWhaL7Hl2AoAkJEvmYnRoAa9tMeU"},"id":"crF6I_WYHUZahZi76mf6hnszJTx66ii3a3_u3GMslGWAPjSw"},{"tx":{"type":"TypeTransfer","chainID":2357,"nonce":1,"gasTip":0,"gasFeeCap":0,"from":"0x0000000000000000000000000000000000000000","to":"0xFFfFFFfFfffFFfFFffFFFfFfFffFFFfffFfFFFff","amount":400000000000000000},"id":"P4XE-FDt5MbB8BTjiC0G13790-jcIhxKmtdfqcKCbyPw-YrE"},{"tx":{"type":"TypeUpdateAccountInfo","chainID":2357,"nonce":1,"gasTip":0,"gasFeeCap":0,"from":"0xFFfFFFfFfffFFfFFffFFFfFfFffFFFfffFfFFFff","data":"omJrcINUjbl8fOziScK5i9wCJsxMKle_UvxURBccN_9de3u43K1cgfFihKIp5kFYIDlZV_u-YMtA7mkbs9pOUJ5RVUrijiXs0XeAkHZCI5J-YnRoAu9OoMU"},"id":"5_3AMSxPqhb0A_-ntTzJd27bpi7IlXaFiVv_6_5R1OR_OJvu"},{"tx":{"type":"TypeCreateData","chainID":2357,"nonce":2,"gasTip":0,"gasFeeCap":0,"from":"0xFFfFFFfFfffFFfFFffFFFfFfFffFFFfffFfFFFff","data":"pWFkWF-pYmJzgVQjV041TkFPRlNOWktEMkxESEtVQmJzaABjZ3JyGQPoY21heBoAAYagY21pbhknEGNtc3DCRejUpRAAY210ZxoAQBZAY210cMJGCRhOcqAAY250YsJEO5rKAGF2AWJrcINUjbl8fOziScK5i9wCJsxMKle_UvxURBccN_9de3u43K1cgfFihKIp5kFYIDlZV_u-YMtA7mkbs9pOUJ5RVUrijiXs0XeAkHZCI5J-YnRoAmNtaWRUAAAAAAAAAAAAAAAAAAAAAAAAAAEkiRPC"},"id":"QouAozxQAthLBGZ7_NCo9ycpN-5Q0eBKVsMxe15w8ROOVDt4"},{"tx":{"type":"TypeCreateModel","chainID":2357,"nonce":3,"gasTip":0,"gasFeeCap":0,"from":"0xFFfFFFfFfffFFfFFffFFFfFfFffFFFfffFfFFFff","data":"pWFua05hbWVTZXJ2aWNlYmFw9mJrcINUjbl8fOziScK5i9wCJsxMKle_UvxURBccN_9de3u43K1cgfFihKIp5kFYIDlZV_u-YMtA7mkbs9pOUJ5RVUrijiXs0XeAkHZCI5J-YnNjeM90eXBlIElEMjAgYnl0ZXMKCXR5cGUgTmFtZVNlcnZpY2Ugc3RydWN0IHsKCQluYW1lICAgICAgIFN0cmluZyAgICAgICAgKHJlbmFtZSAibiIpCgkJbGlua2VkICAgICBvcHRpb25hbCBJRDIwIChyZW5hbWUgImwiKQoJCXJlY29yZHMgICAgW1N0cmluZ10gICAgICAocmVuYW1lICJycyIpCgkJZXh0ZW5zaW9ucyBbQW55XSAgICAgICAgIChyZW5hbWUgImVzIikKCX1idGgCj_buyg"},"id":"b8onI5zOwqPZO9jxMBBgZWnnCUxZH8Cy-SD0YLXpLJRDy5Wq"},{"tx":{"type":"TypeCreateModel","chainID":2357,"nonce":4,"gasTip":0,"gasFeeCap":0,"from":"0xFFfFFFfFfffFFfFFffFFFfFfFffFFFfffFfFFFff","data":"pWFublByb2ZpbGVTZXJ2aWNlYmFw9mJrcINUjbl8fOziScK5i9wCJsxMKle_UvxURBccN_9de3u43K1cgfFihKIp5kFYIDlZV_u-YMtA7mkbs9pOUJ5RVUrijiXs0XeAkHZCI5J-YnNjeQGLdHlwZSBJRDIwIGJ5dGVzCgl0eXBlIFByb2ZpbGVTZXJ2aWNlIHN0cnVjdCB7CgkJdHlwZSAgICAgICAgSW50ICAgICAgICAgICAgIChyZW5hbWUgInQiKQoJCW5hbWUgICAgICAgIFN0cmluZyAgICAgICAgICAocmVuYW1lICJuIikKCQlkZXNjcmlwdGlvbiBTdHJpbmcgICAgICAgICAgKHJlbmFtZSAiZCIpCgkJaW1hZ2UgICAgICAgU3RyaW5nICAgICAgICAgIChyZW5hbWUgImkiKQoJCXVybCAgICAgICAgIFN0cmluZyAgICAgICAgICAocmVuYW1lICJ1IikKCQlmb2xsb3dzICAgICBbSUQyMF0gICAgICAgICAgKHJlbmFtZSAiZnMiKQoJCW1lbWJlcnMgICAgIG9wdGlvbmFsIFtJRDIwXSAocmVuYW1lICJtcyIpCgkJZXh0ZW5zaW9ucyAgW0FueV0gICAgICAgICAgIChyZW5hbWUgImVzIikKCX1idGgCsgQAbQ"},"id":"-xzGonDgQ_-M5FXiFi3MQGDvEDiks_Tqu1jFa27Z2cu0Zgmx"}]`

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
			assert.Equal(`{"builderID":"#WN5NAOFSNZKD2LDHKUB","networkID":1337,"nodeID":"NodeID-7Xhw2mDxuDS44j42TCB6U5579esbSt3Lg","state":"Normal operations state","subnetID":"p433wpuXyJiDhyazPYyZMJeaoPSW76CBZ2x7wrVPLgvokotXz"}`, string(encoding.MustMarshalJSON(result)))
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

	{
		var result uint64
		res := cli.Request(context.Background(), "nextGasPrice", nil, &result)
		if assert.Nil(res.Error) {
			if !assert.Equal(uint64(10000), result) {
				t.Logf("Actual: %v", result)
			}
		}
	}

	{
		// TODO: test preVerifyTxs
	}

	{
		result := ld.Txs{}
		res := cli.Request(context.Background(), "getGenesisTxs", nil, &result)
		if assert.Nil(res.Error) {
			for i := range result {
				assert.NoError(result[i].SyntacticVerify())
				// fmt.Println(ids.ModelIDFromHash(result[i].ID))
			}
			// fmt.Println(string(encoding.MustMarshalJSON(result)))
			assert.Equal(genesisTxsJSON, string(encoding.MustMarshalJSON(result)))
		}
	}

	{
		result := &ld.Block{}
		res := cli.Request(context.Background(), "getBlock", genesisBlockID, result)
		if assert.Nil(res.Error) {
			assert.NoError(result.SyntacticVerify())
			// fmt.Println(string(encoding.MustMarshalJSON(result)))
			assert.Equal(genesisBlockJSON, string(encoding.MustMarshalJSON(result)))
		}
	}

	{
		result := &ld.Block{}
		res := cli.Request(context.Background(), "getBlockAtHeight", uint64(0), result)
		if assert.Nil(res.Error) {
			assert.NoError(result.SyntacticVerify())
			// fmt.Println(string(encoding.MustMarshalJSON(result)))
			assert.Equal(genesisBlockJSON, string(encoding.MustMarshalJSON(result)))
		}
	}

	{
		result := &ld.State{}
		res := cli.Request(context.Background(), "getState", genesisStateID, result)
		if assert.Nil(res.Error) {
			assert.NoError(result.SyntacticVerify())
			// fmt.Println(string(encoding.MustMarshalJSON(result)))
			assert.Equal(genesisStateJSON, string(encoding.MustMarshalJSON(result)))
		}
	}

	{
		result := &ld.Account{}
		res := cli.Request(context.Background(), "getAccount", ids.LDCAccount, result)
		if assert.Nil(res.Error) {
			assert.NoError(result.SyntacticVerify())
			result.ID = ids.LDCAccount
			// fmt.Println(string(encoding.MustMarshalJSON(result)))
			assert.Equal(`{"type":"Token","nonce":2,"balance":500000000000000000,"threshold":0,"keepers":[],"tokens":{},"nonceTable":{},"maxTotalSupply":1000000000000000000,"height":0,"timestamp":0,"address":"0x0000000000000000000000000000000000000000"}`, string(encoding.MustMarshalJSON(result)))
		}

		result = &ld.Account{}
		res = cli.Request(context.Background(), "getAccount", ids.GenesisAccount, result)
		if assert.Nil(res.Error) {
			assert.NoError(result.SyntacticVerify())
			result.ID = ids.GenesisAccount
			// fmt.Println(string(encoding.MustMarshalJSON(result)))
			assert.Equal(`{"type":"Native","nonce":5,"balance":400000000000000000,"threshold":2,"keepers":["jbl8fOziScK5i9wCJsxMKle_UvwKxwPH","RBccN_9de3u43K1cgfFihKIp5kE1lmGG","OVlX-75gy0DuaRuz2k5QnlFVSuKOJezRd4CQdkIjkn5pYt0F"],"tokens":{},"nonceTable":{},"height":0,"timestamp":0,"address":"0xFFfFFFfFfffFFfFFffFFFfFfFffFFFfffFfFFFff"}`, string(encoding.MustMarshalJSON(result)))
		}
	}

	{
		result := &ld.AccountLedger{}
		res := cli.Request(context.Background(), "getLedger", ids.LDCAccount, result)
		assert.NotNil(res.Error)
		assert.Nil(res.Result)

		result = &ld.AccountLedger{}
		res = cli.Request(context.Background(), "getLedger", ids.GenesisAccount, result)
		assert.NotNil(res.Error)
		assert.Nil(res.Result)
	}

	{
		result := &ld.ModelInfo{}
		mid, err := ids.ModelIDFromStr("b8onI5zOwqPZO9jxMBBgZWnnCUzd-187")
		assert.NoError(err)
		res := cli.Request(context.Background(), "getModel", mid, result)
		if assert.Nil(res.Error) {
			result.ID = mid
			assert.NoError(result.SyntacticVerify())
			// fmt.Println(string(encoding.MustMarshalJSON(result)))
			assert.Equal(`{"name":"NameService","threshold":2,"keepers":["jbl8fOziScK5i9wCJsxMKle_UvwKxwPH","RBccN_9de3u43K1cgfFihKIp5kE1lmGG","OVlX-75gy0DuaRuz2k5QnlFVSuKOJezRd4CQdkIjkn5pYt0F"],"schema":"type ID20 bytes\n\ttype NameService struct {\n\t\tname       String        (rename \"n\")\n\t\tlinked     optional ID20 (rename \"l\")\n\t\trecords    [String]      (rename \"rs\")\n\t\textensions [Any]         (rename \"es\")\n\t}","id":"b8onI5zOwqPZO9jxMBBgZWnnCUzd-187"}`, string(encoding.MustMarshalJSON(result)))
		}

		result = &ld.ModelInfo{}
		mid, err = ids.ModelIDFromStr("-xzGonDgQ_-M5FXiFi3MQGDvEDgTw4dJ")
		assert.NoError(err)
		res = cli.Request(context.Background(), "getModel", mid, result)
		if assert.Nil(res.Error) {
			result.ID = mid
			assert.NoError(result.SyntacticVerify())
			// fmt.Println(string(encoding.MustMarshalJSON(result)))
			assert.Equal(`{"name":"ProfileService","threshold":2,"keepers":["jbl8fOziScK5i9wCJsxMKle_UvwKxwPH","RBccN_9de3u43K1cgfFihKIp5kE1lmGG","OVlX-75gy0DuaRuz2k5QnlFVSuKOJezRd4CQdkIjkn5pYt0F"],"schema":"type ID20 bytes\n\ttype ProfileService struct {\n\t\ttype        Int             (rename \"t\")\n\t\tname        String          (rename \"n\")\n\t\tdescription String          (rename \"d\")\n\t\timage       String          (rename \"i\")\n\t\turl         String          (rename \"u\")\n\t\tfollows     [ID20]          (rename \"fs\")\n\t\tmembers     optional [ID20] (rename \"ms\")\n\t\textensions  [Any]           (rename \"es\")\n\t}","id":"-xzGonDgQ_-M5FXiFi3MQGDvEDgTw4dJ"}`, string(encoding.MustMarshalJSON(result)))
		}
	}

	{
		result := &ld.DataInfo{}
		id, err := ids.DataIDFromStr("QouAozxQAthLBGZ7_NCo9ycpN-5Q0eBKVsMxe15w8ROOVDt4")
		assert.NoError(err)
		res := cli.Request(context.Background(), "getData", id, result)
		if assert.Nil(res.Error) {
			result.ID = id
			assert.NoError(result.SyntacticVerify())
			// fmt.Println(string(encoding.MustMarshalJSON(result)))
			assert.Equal(`{"mid":"AAAAAAAAAAAAAAAAAAAAAAAAAAGIYKah","version":1,"threshold":2,"keepers":["jbl8fOziScK5i9wCJsxMKle_UvwKxwPH","RBccN_9de3u43K1cgfFihKIp5kE1lmGG","OVlX-75gy0DuaRuz2k5QnlFVSuKOJezRd4CQdkIjkn5pYt0F"],"payload":"qWJic4FUI1dONU5BT0ZTTlpLRDJMREhLVUJic2gAY2dychkD6GNtYXgaAAGGoGNtaW4ZJxBjbXNwwkXo1KUQAGNtdGcaAEAWQGNtdHDCRgkYTnKgAGNudGLCRDuaygCBPflS","id":"QouAozxQAthLBGZ7_NCo9ycpN-5Q0eBKVsMxe15w8ROOVDt4"}`, string(encoding.MustMarshalJSON(result)))
		}

		// TODO getPrevData
	}

	{
		// TODO getNameID getNameData
	}
}
