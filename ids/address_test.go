// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package ids

import (
	"encoding/json"
	"testing"

	"github.com/fxamacker/cbor/v2"
	"github.com/stretchr/testify/assert"
)

const address1 = "0x8db97c7cECe249C2b98bdc0226cc4C2A57bF52fc"

const address2 = "0x44171C37Ff5D7B7bb8Dcad5C81f16284A229E641"

func TestAddressID(t *testing.T) {
	assert := assert.New(t)

	addr1, err := AddressFromStr(address1)
	assert.Nil(err)
	assert.Equal(address1, addr1.String())

	addr1b, err := AddressFromStr("8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC")
	assert.Nil(err)
	assert.Equal(addr1, addr1b)

	addr2, err := AddressFromStr(address2)
	assert.Nil(err)
	assert.Equal(address2, addr2.String())

	cbordata, err := cbor.Marshal([32]byte{1, 2, 3})
	assert.Nil(err)
	var addr2b Address
	assert.ErrorContains(cbor.Unmarshal(cbordata, &addr2b), "invalid bytes length")

	cbordata, err = cbor.Marshal(addr2)
	assert.Nil(err)
	assert.Nil(cbor.Unmarshal(cbordata, &addr2b))
	assert.Equal(addr2, addr2b)

	data, err := json.Marshal(addr2)
	assert.Nil(err)
	assert.Equal(`"0x44171C37Ff5D7B7bb8Dcad5C81f16284A229E641"`, string(data))

	eids := make(IDList[Address], 0)
	err = json.Unmarshal([]byte(`[
		"0x44171C37Ff5D7B7bb8Dcad5C81f16284A229E641",
	  "44171C37Ff5D7B7bb8dcad5C81f16284A229e641",
		"",
		null
	]`), &eids)
	assert.Nil(err)

	assert.Equal(4, len(eids))
	assert.Equal(addr2, eids[0])
	assert.Equal(addr2, eids[1])
	assert.Equal(EmptyAddress, eids[2])
	assert.Equal(EmptyAddress, eids[3])

	ptrIDs := make([]*Address, 0)
	err = json.Unmarshal([]byte(`[
		"0x44171C37Ff5D7B7bb8Dcad5C81f16284A229E641",
	  "44171C37Ff5D7B7bb8dcad5C81f16284A229e641",
		"",
		null
	]`), &ptrIDs)
	assert.Nil(err)

	assert.Equal(4, len(eids))
	assert.Equal(addr2, *ptrIDs[0])
	assert.Equal(addr2, *ptrIDs[1])
	assert.Equal(EmptyAddress, *ptrIDs[2])
	assert.Nil(ptrIDs[3])

	addr2, err = AddressFromStr("")
	assert.Nil(err)
	assert.Equal(EmptyAddress, addr2)
}
