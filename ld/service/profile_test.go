// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProfile(t *testing.T) {
	_ = assert.New(t)

	// address := util.DataID{1, 2, 3, 4}
	// p1 := &Profile{
	// 	Name:    "lvdm",
	// 	Follows: []util.DataID{address},
	// 	Extra:   new(ld.MapStringAny),
	// }

	// p1.Extra.Set("name", basicnode.NewString("张三"))
	// p1.Extra.Set("age", basicnode.NewInt(21))
	// // i.Extra.Set("field", ipld.Null)

	// assert.Equal(p1.Extra.Keys, []string{"age", "name"})

	// data, err := p1.Marshal()
	// assert.Nil(err)

	// p2 := &Profile{}
	// err = p2.Unmarshal(data)
	// assert.Nil(err)
	// assert.True(p2.Equal(p1))

	// p1.Extra.Set("eth", basicnode.NewString(util.EthID(address).String()))
	// p1.URL = "http://127.0.0.1"
	// data2, err := p1.Marshal()
	// assert.Nil(err)
	// assert.NotEqual(data, data2)

	// data, err = p1.MarshalJSON()
	// assert.Nil(err)

	// jsonStr := `{"@type":"","name":"lvdm","image":"","url":"http://127.0.0.1","kyc":"","follows":["LD6L5yRJL2iYi9PbrhRru6uKfEAzDGHwUJ"],"exmID":"","extra":{"age":21,"eth":"0x0102030400000000000000000000000000000000","name":"张三"}}`
	// assert.Equal(jsonStr, string(data))

	// p3 := &Profile{}
	// err = p3.Unmarshal(data2)
	// assert.Nil(err)
	// assert.True(p3.Equal(p1))
}
