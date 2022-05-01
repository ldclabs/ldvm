// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package service

import (
	"testing"

	"github.com/ava-labs/avalanchego/ids"

	"github.com/stretchr/testify/assert"
)

func TestName(t *testing.T) {
	assert := assert.New(t)

	address := ids.ShortID{1, 2, 3, 4}
	name := &Name{
		Name:   "xn--vuq70b.com",
		Linked: address,
	}
	data, err := name.Marshal()
	assert.Nil(err)

	name2 := &Name{}
	err = name2.Unmarshal(data)
	assert.Nil(err)
	assert.Equal(name, name2)

	name.Records = append(name.Records, "xn--vuq70b.com. IN A 10.0.0.1")
	data2, err := name.Marshal()
	assert.Nil(err)

	assert.NotEqual(data, data2)
	data, err = name.MarshalJSON()
	assert.Nil(err)

	jsonStr := `{"name":"公信.com","domain":"xn--vuq70b.com","linked":"LD6L5yRJL2iYi9PbrhRru6uKfEAzDGHwUJ","records":["xn--vuq70b.com. IN A 10.0.0.1"]}`
	assert.Equal(jsonStr, string(data))

	name3 := &Name{}
	err = name3.Unmarshal(data2)
	assert.Nil(err)
	assert.Equal(name, name3)
}
