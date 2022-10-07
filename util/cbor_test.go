// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCBOR(t *testing.T) {
	assert := assert.New(t)

	addr, err := AddressFrom(address1)
	assert.Nil(err)

	data, err := MarshalCBOR(addr)
	assert.NoError(err)

	assert.NoError(ValidCBOR(data))
	assert.ErrorContains(ValidCBOR(addr[:]), "unexpected EOF")

	var addr1 Address
	assert.NoError(UnmarshalCBOR(data, &addr1))
	assert.Equal(addr.Bytes(), addr1.Bytes())

	data2, err := UnmarshalCBORWithLen(data, 20)
	assert.NoError(err)
	assert.Equal(addr.Bytes(), data2)

	_, err = UnmarshalCBORWithLen(data, 32)
	assert.ErrorContains(err, "invalid bytes length, expected 32, got 20")
}
