// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package encoding

import (
	"encoding/hex"
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCBOR(t *testing.T) {
	assert := assert.New(t)

	addr, err := hex.DecodeString("8db97c7cECe249C2b98bdc0226cc4C2A57bF52fc")
	require.NoError(t, err)

	data, err := MarshalCBOR(addr)
	require.NoError(t, err)

	assert.NoError(ValidCBOR(data))
	assert.ErrorContains(ValidCBOR(addr[:]), "unexpected EOF")

	var addr1 [20]byte
	assert.NoError(UnmarshalCBOR(data, &addr1))
	assert.Equal(addr, addr1[:])

	data2, err := UnmarshalCBORWithLen(data, 20)
	require.NoError(t, err)
	assert.Equal(addr, data2)

	_, err = UnmarshalCBORWithLen(data, 32)
	assert.ErrorContains(err, "invalid bytes length, expected 32, got 20")

	data, err = MarshalCBOR(map[string]interface{}{
		"a":     "a",
		"aa":    "aa",
		"hello": "world",
		"ab":    "ab",
	})
	require.NoError(t, err)

	data2, err = MarshalCBOR(map[string]interface{}{
		"hello": "world",
		"ab":    "ab",
		"a":     "a",
		"aa":    "aa",
	})
	require.NoError(t, err)
	assert.Equal(data, data2)

	assert.NotEqual(MustMarshalCBOR(big.NewInt(1)), MustMarshalCBOR(1))
}
