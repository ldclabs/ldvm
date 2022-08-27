// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package util

import (
	"crypto/rand"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJSON(t *testing.T) {
	assert := assert.New(t)

	type testCase struct {
		input, output []byte
	}
	addr1 := Signer1.Address()
	tcs := []testCase{
		{
			input:  []byte(``),
			output: []byte(``),
		},
		{
			input:  []byte(`null`),
			output: []byte(`null`),
		},
		{
			input:  []byte(`0`),
			output: []byte(`0`),
		},
		{
			input:  []byte(`Hello`),
			output: []byte(`"0x48656c6c6f26381969"`),
		},
		{
			input:  addr1[:],
			output: []byte(`"0x8db97c7cece249c2b98bdc0226cc4c2a57bf52fcb2822649"`),
		},
		{
			input:  []byte(`{}`),
			output: []byte(`{}`),
		},
		{
			input:  []byte(`[1,2,3]`),
			output: []byte(`[1,2,3]`),
		},
		{
			input:  []byte(`"Hello 👋"`),
			output: []byte(`"Hello 👋"`),
		},
	}
	for _, c := range tcs {
		o := MarshalJSONData(c.input)
		assert.Equal(c.output, []byte(o))
		assert.Equal(c.input, UnmarshalJSONData(o))
	}
}

func TestHashFromData(t *testing.T) {
	assert := assert.New(t)

	assert.Equal(HashFromData([]byte{'a', 'b', 'c'}), HashFromData([]byte{'a', 'b', 'c'}))
	assert.NotEqual(HashFromData([]byte{'a', 'b', 'c'}), HashFromData([]byte{'a', 'b', 'c', 'd'}))
}

func TestEthIDs(t *testing.T) {
	assert := assert.New(t)

	var ids EthIDs
	assert.NoError(ids.CheckDuplicate())
	assert.NoError(ids.CheckEmptyID())

	ids = EthIDs{
		EthIDEmpty,
		{1, 2, 3},
		Signer1.Address(),
		Signer2.Address(),
	}
	assert.True(ids.Has(EthIDEmpty))
	assert.True(ids.Has(EthID{1, 2, 3}))
	assert.True(ids.Has(Signer2.Address()))

	assert.False(ids.Has(EthID{1, 2, 4}))
	assert.Nil(ids.CheckDuplicate())
	ids = append(ids, EthID{1, 2, 3})

	assert.ErrorContains(ids.CheckDuplicate(), EthID{1, 2, 3}.String())
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
	assert.ErrorContains(ids.CheckEmptyID(), "empty data id exists")
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
