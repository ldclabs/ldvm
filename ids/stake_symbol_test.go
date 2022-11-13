// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package ids

import (
	"encoding/json"
	"testing"

	"github.com/fxamacker/cbor/v2"
	"github.com/stretchr/testify/assert"
)

func TestStakeSymbol(t *testing.T) {
	assert := assert.New(t)

	token := "#LDC"
	id, err := StakeFromStr(token)
	assert.Nil(err)

	assert.Equal(
		StakeSymbol{0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
			0, 0, 0, 0, 0, 0, '#', 'L', 'D', 'C'}, id)

	cbordata, err := cbor.Marshal([32]byte{'#', 'L', 'D', 'C'})
	assert.Nil(err)
	var id2 StakeSymbol
	assert.ErrorContains(cbor.Unmarshal(cbordata, &id2), "invalid bytes length")

	cbordata, err = cbor.Marshal(id)
	assert.Nil(err)
	assert.Nil(cbor.Unmarshal(cbordata, &id2))
	assert.Equal(id, id2)

	data, err := json.Marshal(id)
	assert.Nil(err)
	assert.Equal(`"#LDC"`, string(data))

	id, err = StakeFromStr("")
	assert.Nil(err)
	assert.Equal(EmptyStake, id)

	type testCase struct {
		shouldErr bool
		symbol    string
		token     StakeSymbol
	}
	tcs := []testCase{
		{shouldErr: true, symbol: "",
			token: StakeSymbol{
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
			}},
		{shouldErr: false, symbol: "#D",
			token: StakeSymbol{
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, '#', 'D',
			}},
		{shouldErr: false, symbol: "#USD",
			token: StakeSymbol{
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, '#', 'U', 'S', 'D',
			}},
		{shouldErr: false, symbol: "#1D",
			token: StakeSymbol{
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, '#', '1', 'D',
			}},
		{shouldErr: false, symbol: "#USD1",
			token: StakeSymbol{
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, '#', 'U', 'S', 'D', '1',
			}},
		{shouldErr: false, symbol: "#012345678",
			token: StakeSymbol{
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				'#', '0', '1', '2', '3', '4', '5', '6', '7', '8',
			}},
		{shouldErr: true, symbol: "",
			token: StakeSymbol{
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 'L', 'D', 0,
			}},
		{shouldErr: true, symbol: "",
			token: StakeSymbol{
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, '#',
			}},
		{shouldErr: true, symbol: "",
			token: StakeSymbol{
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, '0', 'L', 'D', 'C',
			}},
		{shouldErr: true, symbol: "",
			token: StakeSymbol{
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, '#', 'L', 'D', 'c',
			}},
		{shouldErr: true, symbol: "",
			token: StakeSymbol{
				'#', 'L', 'L', 'L', 'L', 'L', 'L', 'L', 'L', 'L',
				'L', 'L', 'L', 'L', 'L', 'L', 'L', 'L', 'L', 'l',
			}},
		{shouldErr: true, symbol: "#LDc",
			token: StakeSymbol{
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
			}},
		{shouldErr: true, symbol: "#L_C",
			token: StakeSymbol{
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
			}},
		{shouldErr: true, symbol: "#L C",
			token: StakeSymbol{
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
			}},
		{shouldErr: true, symbol: "1LDC",
			token: StakeSymbol{
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
			}},
		{shouldErr: true, symbol: "1234567890",
			token: StakeSymbol{
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
			}},
		{shouldErr: true, symbol: "#LD\u200dC", // with Zero Width Joiner
			token: StakeSymbol{
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
			}},
	}
	for _, c := range tcs {
		switch {
		case c.shouldErr:
			assert.Equal("", c.token.String())
			if c.symbol != "" {
				_, err := StakeFromStr(c.symbol)
				assert.Error(err)
			}
		default:
			assert.Equal(c.symbol, c.token.String())
			assert.True(c.token.Valid())
			id, err := StakeFromStr(c.symbol)
			assert.Nil(err)
			assert.Equal(c.token, id)
		}
	}
}
