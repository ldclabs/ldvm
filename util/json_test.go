// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package util

import (
	"encoding/json"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJSON(t *testing.T) {
	assert := assert.New(t)

	type testCase struct {
		input, output []byte
	}

	addr1, _ := AddressFrom(address1)
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
	assert.NoError(err)
	assert.Equal(`{"data":"jbl8fOziScK5i9wCJsxMKle_UvwKxwPH"}`, string(b))

	v := &testRawData{}
	assert.NoError(json.Unmarshal(b, v))
	assert.Equal(addr1[:], []byte(v.Data))

	data = &testRawData{Data: []byte(`"Hello 👋"`)}
	b, err = json.Marshal(data)
	assert.NoError(err)
	assert.Equal(`{"data":"Hello 👋"}`, string(b))

	v = &testRawData{}
	assert.NoError(json.Unmarshal(b, &v))
	assert.Equal([]byte(`"Hello 👋"`), []byte(v.Data))
}
