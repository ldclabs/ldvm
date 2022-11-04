// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package util

import (
	"crypto/rand"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIDList(t *testing.T) {
	t.Run("Address", func(t *testing.T) {
		assert := assert.New(t)

		var ids IDList[Address]
		assert.NoError(ids.CheckDuplicate())
		assert.NoError(ids.CheckValid())
		assert.NoError(ids.Valid())

		addr1, _ := AddressFrom(address1)
		addr2, _ := AddressFrom(address2)

		ids = IDList[Address]{
			AddressEmpty,
			{1, 2, 3},
			addr1,
			addr2,
		}
		assert.True(ids.Has(AddressEmpty))
		assert.True(ids.Has(Address{1, 2, 3}))
		assert.True(ids.Has(addr2))

		assert.False(ids.Has(Address{1, 2, 4}))
		assert.Nil(ids.CheckDuplicate())
		ids = append(ids, Address{1, 2, 3})

		assert.ErrorContains(ids.CheckDuplicate(), Address{1, 2, 3}.String())
		assert.ErrorContains(ids.CheckValid(), "empty id exists")
		assert.Error(ids.Valid())
	})

	t.Run("DataID", func(t *testing.T) {
		assert := assert.New(t)

		var ids IDList[DataID]
		assert.NoError(ids.CheckDuplicate())
		assert.NoError(ids.CheckValid())
		assert.NoError(ids.Valid())

		id1 := DataID{}
		rand.Read(id1[:])
		id2 := DataID{}
		rand.Read(id2[:])

		ids = IDList[DataID]{
			DataIDEmpty,
			{1, 2, 3},
			id1,
			id2,
		}
		assert.True(ids.Has(DataIDEmpty))
		assert.True(ids.Has(DataID{1, 2, 3}))
		assert.True(ids.Has(id2))

		assert.False(ids.Has(DataID{1, 2, 4}))
		assert.Nil(ids.CheckDuplicate())
		ids = append(ids, DataID{1, 2, 3})

		assert.ErrorContains(ids.CheckDuplicate(), DataID{1, 2, 3}.String())
		assert.ErrorContains(ids.CheckValid(), "empty id exists")
		assert.Error(ids.Valid())
	})
}

func TestSet(t *testing.T) {
	t.Run("unt64", func(t *testing.T) {
		assert := assert.New(t)

		set := NewSet[uint64](0)
		assert.False(set.Has(0))

		set.Add(0, 9, 888, 1, 5, 2, 0, 1)
		assert.True(set.Has(0))
		assert.True(set.Has(888))
		assert.Equal([]uint64{0, 1, 2, 5, 9, 888}, set.List())
	})
}
