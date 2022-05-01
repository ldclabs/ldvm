// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package ld

import (
	"testing"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/stretchr/testify/assert"

	"github.com/ldclabs/ldvm/util"
)

func TestDataMeta(t *testing.T) {
	assert := assert.New(t)

	address := util.Signer1.Address()
	data := []byte("testdata")
	kSig, err := util.Signer1.Sign(data)
	dm := &DataMeta{
		ModelID:   ids.ShortID{4, 5, 6, 7, 8},
		Version:   10,
		Threshold: 1,
		Keepers:   []ids.ShortID{address},
		KSig:      kSig,
		Data:      data,
	}
	data, err = dm.Marshal()
	assert.Nil(err)

	dm2 := &DataMeta{}
	err = dm2.Unmarshal(data)
	assert.Nil(err)
	assert.True(dm2.Equal(dm))

	dm.Version++
	data2, err := dm.Marshal()
	assert.Nil(err)
	assert.NotEqual(data, data2)
}
