// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package encoding

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCheckSumHex(t *testing.T) {
	assert := assert.New(t)

	addr, err := hex.DecodeString("8db97c7cECe249C2b98bdc0226cc4C2A57bF52fc")
	require.NoError(t, err)

	data := []byte{}
	assert.Equal("0x", CheckSumHex(data))

	data = addr[:8]
	// fmt.Printf("%b\n", Sum256(data))
	// 0b11010011100110011001100 ...
	// 0x8Db97c7CEce249c2
	assert.Equal("0x8Db97c7CEce249c2", CheckSumHex(data))

	data = addr[:]
	assert.Equal("0x8db97c7cECe249C2b98bdc0226cc4C2A57bF52fc", CheckSumHex(data))

	data = make([]byte, 40)
	copy(data, addr[:])
	copy(data[20:], addr[:])
	assert.Equal("0x8DB97c7cECE249C2B98bdc0226CC4c2a57BF52FC8DB97C7Cece249c2b98bdC0226Cc4c2A57BF52fC", CheckSumHex(data))

	data = make([]byte, 8*20)
	for i := 0; i < 8; i++ {
		copy(data[i*20:(i+1)*20], addr[:])
	}

	hexStr := "0x8db97c7cECe249C2B98bDC0226cC4c2A57bf52fC8dB97c7ceCE249c2b98bdc0226CC4C2A57BF52fC8dB97C7CecE249c2B98bdC0226cc4c2A57Bf52FC8db97c7CeCE249c2b98bdc0226CC4C2A57bF52Fc8dB97c7CeCE249c2B98bDc0226cC4C2a57bF52Fc8Db97C7cECE249c2b98BdC0226cC4C2a57Bf52Fc8Db97c7CEcE249c2b98bdc0226cc4c2a57bf52fc8db97c7cece249c2b98bdc0226cc4c2a57bf52fc"
	assert.Equal(hexStr, CheckSumHex(data))

	data[0] = 0
	hexStr = "0x00b97c7cECe249C2B98bDC0226cC4c2A57bf52fC8dB97c7ceCE249c2b98bdc0226CC4C2A57BF52fC8dB97C7CecE249c2B98bdC0226cc4c2A57Bf52FC8db97c7CeCE249c2b98bdc0226CC4C2A57bF52Fc8dB97c7CeCE249c2B98bDc0226cC4C2a57bF52Fc8Db97C7cECE249c2b98BdC0226cC4C2a57Bf52Fc8Db97c7CEcE249c2b98bdc0226cc4c2a57bf52fc8db97c7cece249c2b98bdc0226cc4c2a57bf52fc"
	assert.NotEqual(hexStr, CheckSumHex(data))

	b, err := hex.DecodeString(hexStr[2:])
	require.NoError(t, err)
	assert.Equal(data, b)
}

func TestEncodeToStringAndDecodeString(t *testing.T) {
	assert := assert.New(t)

	addr, err := hex.DecodeString("8db97c7cECe249C2b98bdc0226cc4C2A57bF52fc")
	require.NoError(t, err)

	str := EncodeToString(addr[:])
	assert.Equal("jbl8fOziScK5i9wCJsxMKle_UvwKxwPH", str)

	data, err := DecodeString(str)
	require.NoError(t, err)
	assert.Equal(addr, data)

	data, err = DecodeStringWithLen(str, 20)
	require.NoError(t, err)
	assert.Equal(addr, data)

	_, err = DecodeStringWithLen(str, 22)
	assert.ErrorContains(err, "invalid bytes length, expected 22, got 20")

	_, err = DecodeString("abc")
	assert.ErrorContains(err, "no checksum bytes")

	_, err = DecodeString(str[:len(str)-1] + "h")
	assert.ErrorContains(err, "invalid input checksum")
}

func TestEncodeToQuoteStringAndDecodeQuoteString(t *testing.T) {
	assert := assert.New(t)

	addr, err := hex.DecodeString("8db97c7cECe249C2b98bdc0226cc4C2A57bF52fc")
	require.NoError(t, err)

	str := EncodeToQuoteString(addr[:])
	assert.Equal(`"jbl8fOziScK5i9wCJsxMKle_UvwKxwPH"`, str)

	data, err := DecodeQuoteString(str)
	require.NoError(t, err)
	assert.Equal(addr, data)

	data, err = DecodeQuoteStringWithLen(str, 20)
	require.NoError(t, err)
	assert.Equal(addr, data)

	_, err = DecodeQuoteStringWithLen(str, 32)
	assert.ErrorContains(err, "invalid bytes length, expected 32, got 20")

	_, err = DecodeQuoteString(`"abc`)
	assert.ErrorContains(err, "invalid quote string")

	_, err = DecodeQuoteString(str[:len(str)-2] + `h"`)
	assert.ErrorContains(err, "invalid input checksum")
}
