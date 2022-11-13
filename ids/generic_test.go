// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package ids

import (
	"crypto/rand"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIDList(t *testing.T) {
	t.Run("Address", func(t *testing.T) {
		assert := assert.New(t)

		var list IDList[Address]
		assert.NoError(list.CheckDuplicate())
		assert.NoError(list.CheckValid())
		assert.NoError(list.Valid())

		addr1, _ := AddressFromStr(address1)
		addr2, _ := AddressFromStr(address2)

		list = IDList[Address]{
			EmptyAddress,
			{1, 2, 3},
			addr1,
			addr2,
		}
		assert.True(list.Has(EmptyAddress))
		assert.True(list.Has(Address{1, 2, 3}))
		assert.True(list.Has(addr2))
		assert.True(list.Equal(IDList[Address]{
			EmptyAddress,
			{1, 2, 3},
			addr1,
			addr2,
		}))
		assert.False(list.Equal(IDList[Address]{
			{1, 2, 3},
			addr1,
			addr2,
		}))

		assert.False(list.Has(Address{1, 2, 4}))
		assert.Nil(list.CheckDuplicate())
		list = append(list, Address{1, 2, 3})

		assert.ErrorContains(list.CheckDuplicate(), Address{1, 2, 3}.String())
		assert.ErrorContains(list.CheckValid(), "empty id exists")
		assert.Error(list.Valid())
	})

	t.Run("DataID", func(t *testing.T) {
		assert := assert.New(t)

		var list IDList[DataID]
		assert.NoError(list.CheckDuplicate())
		assert.NoError(list.CheckValid())
		assert.NoError(list.Valid())

		id1 := DataID{}
		rand.Read(id1[:])
		id2 := DataID{}
		rand.Read(id2[:])

		list = IDList[DataID]{
			EmptyDataID,
			{1, 2, 3},
			id1,
			id2,
		}
		assert.True(list.Has(EmptyDataID))
		assert.True(list.Has(DataID{1, 2, 3}))
		assert.True(list.Has(id2))
		assert.True(list.Equal(IDList[DataID]{
			EmptyDataID,
			{1, 2, 3},
			id1,
			id2,
		}))
		assert.False(list.Equal(IDList[DataID]{
			{1, 2, 3},
			id1,
			id2,
		}))

		assert.False(list.Has(DataID{1, 2, 4}))
		assert.Nil(list.CheckDuplicate())
		list = append(list, DataID{1, 2, 3})

		assert.ErrorContains(list.CheckDuplicate(), DataID{1, 2, 3}.String())
		assert.ErrorContains(list.CheckValid(), "empty id exists")
		assert.Error(list.Valid())
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
		assert.Equal(List[uint64]{0, 1, 2, 5, 9, 888}, set.List())
	})
}
