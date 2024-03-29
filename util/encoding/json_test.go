// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package encoding

import (
	"encoding/hex"
	"encoding/json"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJSON(t *testing.T) {
	assert := assert.New(t)

	type testCase struct {
		input, output []byte
	}

	addr1, err := hex.DecodeString("8db97c7cECe249C2b98bdc0226cc4C2A57bF52fc")
	require.NoError(t, err)

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
			output: []byte(`"SGVsbG-Mpm7m"`),
		},
		{
			input:  addr1[:],
			output: []byte(`"jbl8fOziScK5i9wCJsxMKle_UvwKxwPH"`),
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
		assert.Equal(c.output, []byte(o), strconv.Quote(string(c.input)))
		assert.Equal(c.input, UnmarshalJSONData(o), strconv.Quote(string(c.input)))
	}

	type testRawData struct {
		Data RawData `json:"data"`
	}

	data := &testRawData{Data: addr1[:]}
	b, err := json.Marshal(data)
	require.NoError(t, err)
	assert.Equal(`{"data":"jbl8fOziScK5i9wCJsxMKle_UvwKxwPH"}`, string(b))

	v := &testRawData{}
	assert.NoError(json.Unmarshal(b, v))
	assert.Equal(addr1[:], []byte(v.Data))

	data = &testRawData{Data: []byte(`"Hello 👋"`)}
	b, err = json.Marshal(data)
	require.NoError(t, err)
	assert.Equal(`{"data":"Hello 👋"}`, string(b))

	v = &testRawData{}
	assert.NoError(json.Unmarshal(b, &v))
	assert.Equal([]byte(`"Hello 👋"`), []byte(v.Data))
}
