// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package util

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

// DvNUrvtQgPynDZN7kFckpjZgmTvW8FX5i
const address1 = "0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC"

// 7D2dmjrr9Fzg7D6tUQAbPKVdhho4uTmo6
const address2 = "0x44171C37Ff5D7B7bb8dcad5C81f16284A229e641"

func TestEthID(t *testing.T) {
	assert := assert.New(t)

	id1, err := EthIDFromString(address1)
	assert.Nil(err)
	assert.Equal(Signer1.Address(), id1)

	id2, err := EthIDFromString("8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC")
	assert.Nil(err)
	assert.Equal(id1, id2)

	id3, err := EthIDFromString("DvNUrvtQgPynDZN7kFckpjZgmTvW8FX5i")
	assert.Nil(err)
	assert.Equal(id1, id3)

	id, err := EthIDFromString(address2)
	assert.Nil(err)
	assert.Equal(Signer2.Address(), id)

	eids := make([]EthID, 0)
	err = json.Unmarshal([]byte(`[
		"0x44171C37Ff5D7B7bb8dcad5C81f16284A229e641",
	  "44171C37Ff5D7B7bb8dcad5C81f16284A229e641",
	  "7D2dmjrr9Fzg7D6tUQAbPKVdhho4uTmo6",
		"",
		null
	]`), &eids)
	assert.Nil(err)

	assert.Equal(5, len(eids))
	assert.Equal(id, eids[0])
	assert.Equal(id, eids[1])
	assert.Equal(id, eids[2])
	assert.Equal(EthIDEmpty, eids[3])
	assert.Equal(EthIDEmpty, eids[4])

	id, err = EthIDFromString("")
	assert.Nil(err)
	assert.Equal(EthIDEmpty, id)
}

func TestModelID(t *testing.T) {
	assert := assert.New(t)

	mid := "LM7tTg8ExJDoq8cgufYnU7EbisEdSbkiEov"
	id, err := ModelIDFromString(mid)
	assert.Nil(err)

	mids := make([]ModelID, 0)
	err = json.Unmarshal([]byte(`[
		"LM7tTg8ExJDoq8cgufYnU7EbisEdSbkiEov",
		"",
		null
	]`), &mids)
	assert.Nil(err)

	assert.Equal(3, len(mids))
	assert.Equal(id, mids[0])
	assert.Equal(ModelIDEmpty, mids[1])
	assert.Equal(ModelIDEmpty, mids[2])

	id, err = ModelIDFromString("")
	assert.Nil(err)
	assert.Equal(ModelIDEmpty, id)
}

func TestDataID(t *testing.T) {
	assert := assert.New(t)

	mid := "LD7tTg8ExJDoq8cgufYnU7EbisEdSbkiEov"
	id, err := DataIDFromString(mid)
	assert.Nil(err)

	mids := make([]DataID, 0)
	err = json.Unmarshal([]byte(`[
		"LD7tTg8ExJDoq8cgufYnU7EbisEdSbkiEov",
		"",
		null
	]`), &mids)
	assert.Nil(err)

	assert.Equal(3, len(mids))
	assert.Equal(id, mids[0])
	assert.Equal(DataIDEmpty, mids[1])
	assert.Equal(DataIDEmpty, mids[2])

	id, err = DataIDFromString("")
	assert.Nil(err)
	assert.Equal(DataIDEmpty, id)
}
