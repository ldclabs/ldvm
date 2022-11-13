// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package ld

import (
	"encoding/json"
	"math/big"
	"testing"

	"github.com/ldclabs/ldvm/ids"
	"github.com/ldclabs/ldvm/signer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSendOutputs(t *testing.T) {
	assert := assert.New(t)

	var so SendOutputs
	assert.ErrorContains(so.SyntacticVerify(), "empty SendOutputs")

	so = SendOutputs{}
	assert.ErrorContains(so.SyntacticVerify(), "empty SendOutputs")

	so = SendOutputs{{}}
	assert.ErrorContains(so.SyntacticVerify(), "invalid to address at 0")

	so = SendOutputs{{To: ids.EmptyAddress}}
	assert.ErrorContains(so.SyntacticVerify(), "invalid to address at 0")

	so = SendOutputs{{To: ids.Address{1, 2, 3}}}
	assert.ErrorContains(so.SyntacticVerify(), "invalid amount at 0")

	so = SendOutputs{{To: ids.Address{1, 2, 3}, Amount: big.NewInt(0)}}
	assert.ErrorContains(so.SyntacticVerify(), "invalid amount at 0")

	so = SendOutputs{{To: ids.Address{1, 2, 3}, Amount: big.NewInt(1)}}
	assert.NoError(so.SyntacticVerify())

	so = SendOutputs{{To: ids.Address{1, 2, 3}, Amount: big.NewInt(1)}, {}}
	assert.ErrorContains(so.SyntacticVerify(), "invalid to address at 1")

	so = SendOutputs{
		{To: ids.Address{1, 2, 3}, Amount: big.NewInt(1)},
		{To: ids.Address{1, 2, 3}, Amount: big.NewInt(1)}}
	assert.ErrorContains(so.SyntacticVerify(), "duplicate to address 0x0102030000000000000000000000000000000000 at 1")

	so = SendOutputs{
		{To: signer.Signer1.Key().Address(), Amount: big.NewInt(1111)},
		{To: signer.Signer2.Key().Address(), Amount: big.NewInt(22222)},
		{To: signer.Signer3.Key().Address(), Amount: big.NewInt(333333)},
		{To: signer.Signer4.Key().Address(), Amount: big.NewInt(4444444)},
	}
	assert.NoError(so.SyntacticVerify())
	data, err := so.Marshal()
	require.NoError(t, err)

	jsondata, err := json.Marshal(so)
	require.NoError(t, err)
	// fmt.Println(string(jsondata))
	assert.Equal(`[{"to":"0x8db97c7cECe249C2b98bdc0226cc4C2A57bF52fc","amount":1111},{"to":"0x44171C37Ff5D7B7bb8Dcad5C81f16284A229E641","amount":22222},{"to":"0x6962DD0564Fb1f8459624e5b7c5dD9A38b2F990d","amount":333333},{"to":"0x22C05D35Be1305c33810086d3A4dB598c3E1Cf48","amount":4444444}]`, string(jsondata))

	var so2 SendOutputs
	assert.NoError(so2.Unmarshal(data))
	assert.Equal(data, MustMarshal(so2))
}
