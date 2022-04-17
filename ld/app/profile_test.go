// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package app

import (
	"testing"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ipld/go-ipld-prime/node/basicnode"
	"github.com/stretchr/testify/assert"

	"github.com/ldclabs/ldvm/ld"
)

func TestProfile(t *testing.T) {
	assert := assert.New(t)

	address := ids.ShortID{1, 2, 3, 4}
	i := &Profile{
		Name:    "lvdm",
		Follows: []string{ld.EthID(address).String()},
		Addrs:   new(MapStringString),
		Extra:   new(MapStringAny),
	}

	i.Addrs.Set("eth", "0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC")
	i.Addrs.Set("email", "hello@test.com")
	i.Extra.Set("name", basicnode.NewString("张三"))
	i.Extra.Set("age", basicnode.NewInt(21))
	// i.Extra.Set("field", ipld.Null)

	assert.Equal(i.Addrs.Keys, []string{"email", "eth"})
	assert.Equal(i.Extra.Keys, []string{"age", "name"})

	profile := NewProfile(i)
	data, err := profile.Marshal()
	assert.Nil(err)

	profile2 := NewProfile(nil)
	err = profile2.Unmarshal(data)
	assert.Nil(err)
	assert.True(profile2.Equal(profile))

	profile.Entity.Addrs.Set("eth", ld.EthID(address).String())
	profile.Entity.Url = "http://127.0.0.1"
	data2, err := profile.Marshal()
	assert.Nil(err)
	assert.NotEqual(data, data2)

	data, err = profile.ToJSON()
	assert.Nil(err)

	jsonStr := `{"addrs":{"email":"hello@test.com","eth":"0x0102030400000000000000000000000000000000"},"extra":{"age":21,"name":"张三"},"extraMID":"","follows":["0x0102030400000000000000000000000000000000"],"image":"","kyc":"","name":"lvdm","url":"http://127.0.0.1"}`
	assert.Equal(jsonStr, string(data))
}
