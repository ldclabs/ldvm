// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package util

import (
	"crypto/rand"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHashFromData(t *testing.T) {
	assert := assert.New(t)

	assert.Equal(HashFromData([]byte{'a', 'b', 'c'}), HashFromData([]byte{'a', 'b', 'c'}))
	assert.NotEqual(HashFromData([]byte{'a', 'b', 'c'}), HashFromData([]byte{'a', 'b', 'c', 'd'}))
}

func TestAddresses(t *testing.T) {
	assert := assert.New(t)

	var ids Addresses
	assert.NoError(ids.CheckDuplicate())
	assert.NoError(ids.CheckEmptyID())

	addr1, _ := AddressFrom(address1)
	addr2, _ := AddressFrom(address2)

	ids = Addresses{
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
	assert.ErrorContains(ids.CheckEmptyID(), "empty address exists")
}

func TestDataIDs(t *testing.T) {
	assert := assert.New(t)

	var ids DataIDs
	assert.NoError(ids.CheckDuplicate())
	assert.NoError(ids.CheckEmptyID())

	id1 := DataID{}
	rand.Read(id1[:])
	id2 := DataID{}
	rand.Read(id2[:])

	ids = DataIDs{
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
	assert.ErrorContains(ids.CheckEmptyID(), "empty dataID exists")
}

func TestUint64Set(t *testing.T) {
	assert := assert.New(t)

	set := make(Uint64Set)
	assert.False(set.Has(0))

	set.Add(0, 9, 888, 1, 5, 2, 0, 1)
	assert.True(set.Has(0))
	assert.True(set.Has(888))
	assert.Equal([]uint64{0, 1, 2, 5, 9, 888}, set.List())
}

func TestAddressToStakeSymbol(t *testing.T) {
	assert := assert.New(t)

	ldc, err := StakeFrom("#LDC")
	addr1, _ := AddressFrom(address1)
	addr2, _ := AddressFrom(address2)
	assert.Nil(err)
	ids := Addresses{
		Address(ldc),
		AddressEmpty,
		addr1,
		addr2,
	}
	ss := AddressToStakeSymbol(ids...)
	assert.Equal(ldc, ss[0])
	assert.Equal("#LDC", ss[0].String())
	assert.Equal("", ss[1].String())
	assert.Equal(string(AddressEmpty[:]), string(ss[1][:]))
	assert.Equal("#BLDQHR4QOJZMNIC5Q5U", ss[2].String())
	assert.Equal("#BLDQHR4QOJZMNIC5Q5U", string(ss[2][:]))
	assert.Equal("#GWLGDBWNPCOAN55PCUX", ss[3].String())
	assert.Equal("#GWLGDBWNPCOAN55PCUX", string(ss[3][:]))
}
