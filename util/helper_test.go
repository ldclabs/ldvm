// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHashFromData(t *testing.T) {
	assert := assert.New(t)

	assert.Equal(HashFromData([]byte{'a', 'b', 'c'}), HashFromData([]byte{'a', 'b', 'c'}))
	assert.NotEqual(HashFromData([]byte{'a', 'b', 'c'}), HashFromData([]byte{'a', 'b', 'c', 'd'}))
}

func TestAddressToStakeSymbol(t *testing.T) {
	assert := assert.New(t)

	ldc, err := StakeFrom("#LDC")
	addr1, _ := AddressFrom(address1)
	addr2, _ := AddressFrom(address2)
	assert.Nil(err)
	ids := IDList[Address]{
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
