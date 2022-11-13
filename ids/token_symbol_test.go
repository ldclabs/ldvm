// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package ids

import (
	"encoding/json"
	"testing"

	"github.com/fxamacker/cbor/v2"
	"github.com/stretchr/testify/assert"
)

func TestTokenSymbol(t *testing.T) {
	assert := assert.New(t)

	token := "$LDC"
	id, err := TokenFromStr(token)
	assert.Nil(err)

	assert.Equal(
		TokenSymbol{0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
			0, 0, 0, 0, 0, 0, '$', 'L', 'D', 'C'}, id)

	cbordata, err := cbor.Marshal([32]byte{'$', 'L', 'D', 'C'})
	assert.Nil(err)
	var id2 TokenSymbol
	assert.ErrorContains(cbor.Unmarshal(cbordata, &id2), "invalid bytes length")

	cbordata, err = cbor.Marshal(id)
	assert.Nil(err)
	assert.Nil(cbor.Unmarshal(cbordata, &id2))
	assert.Equal(id, id2)

	data, err := json.Marshal(id)
	assert.Nil(err)
	assert.Equal(`"$LDC"`, string(data))

	id, err = TokenFromStr("")
	assert.Nil(err)
	assert.Equal(NativeToken, id)

	type testCase struct {
		shouldErr bool
		symbol    string
		token     TokenSymbol
	}
	tcs := []testCase{
		{shouldErr: false, symbol: "",
			token: TokenSymbol{
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
			}},
		{shouldErr: false, symbol: "$D",
			token: TokenSymbol{
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, '$', 'D',
			}},
		{shouldErr: false, symbol: "$USD",
			token: TokenSymbol{
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, '$', 'U', 'S', 'D',
			}},
		{shouldErr: false, symbol: "$1D",
			token: TokenSymbol{
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, '$', '1', 'D',
			}},
		{shouldErr: false, symbol: "$USD1",
			token: TokenSymbol{
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, '$', 'U', 'S', 'D', '1',
			}},
		{shouldErr: false, symbol: "$012345678",
			token: TokenSymbol{
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				'$', '0', '1', '2', '3', '4', '5', '6', '7', '8',
			}},
		{shouldErr: false, symbol: "$ABCDEFGHIJ012345678",
			token: TokenSymbol{
				'$', 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I',
				'J', '0', '1', '2', '3', '4', '5', '6', '7', '8',
			}},
		{shouldErr: true, symbol: "",
			token: TokenSymbol{
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 'L', 'D', 0,
			}},
		{shouldErr: true, symbol: "",
			token: TokenSymbol{
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 'C',
			}},
		{shouldErr: true, symbol: "",
			token: TokenSymbol{
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, '$',
			}},
		{shouldErr: true, symbol: "",
			token: TokenSymbol{
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, '0', 'L', 'D', 'C',
			}},
		{shouldErr: true, symbol: "",
			token: TokenSymbol{
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, '$', 0, 'c',
			}},
		{shouldErr: true, symbol: "",
			token: TokenSymbol{
				0, 0, 0, 'L', 'L', 'L', 'L', 'L', 'L', 'L',
				'L', 'L', 'L', 'L', 'L', 'L', 'L', 'L', 'L', 'L',
			}},
		{shouldErr: true, symbol: "$LDc",
			token: TokenSymbol{
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
			}},
		{shouldErr: true, symbol: "$L_C",
			token: TokenSymbol{
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
			}},
		{shouldErr: true, symbol: "$L C",
			token: TokenSymbol{
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
			}},
		{shouldErr: true, symbol: "1LDC",
			token: TokenSymbol{
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
			}},
		{shouldErr: true, symbol: "1234567890",
			token: TokenSymbol{
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
			}},
		{shouldErr: true, symbol: "$LD\u200dC", // with Zero Width Joiner
			token: TokenSymbol{
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
			}},
	}
	for _, c := range tcs {
		switch {
		case c.shouldErr:
			assert.Equal("", c.token.String())
			if c.token != NativeToken {
				assert.False(c.token.Valid())
			}
			if c.symbol != "" {
				_, err := TokenFromStr(c.symbol)
				assert.Error(err)
			}
		default:
			assert.Equal(c.symbol, c.token.String())
			assert.True(c.token.Valid())
			id, err := TokenFromStr(c.symbol)
			assert.Nil(err)
			assert.Equal(c.token, id)
		}
	}
}

func FuzzTokenSymbol(f *testing.F) {
	for _, seed := range []string{
		"",
		"$AVAX",
		"abc",
		"$A100",
		"$ABCDEFGHIJKLMNOPQRST",
	} {
		f.Add(seed)
	}
	counter := 0
	f.Fuzz(func(t *testing.T, in string) {
		id, err := TokenFromStr(in)
		switch {
		case err == nil:
			counter++
			assert.Equal(t, in, id.String())
		default:
		}
	})
	assert.True(f, counter > 0)
}
