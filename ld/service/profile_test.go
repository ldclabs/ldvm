// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package service

import (
	"encoding/json"
	"testing"

	"github.com/ldclabs/ldvm/util"
	"github.com/stretchr/testify/assert"
)

func TestProfile(t *testing.T) {
	assert := assert.New(t)

	var p *Profile
	assert.ErrorContains(p.SyntacticVerify(), "nil pointer")

	p = &Profile{Type: "LDC"}
	assert.ErrorContains(p.SyntacticVerify(), `invalid type "LDC"`)

	p = &Profile{Type: "Thing", Name: ""}
	assert.ErrorContains(p.SyntacticVerify(), `invalid name ""`)

	p = &Profile{Type: "Thing", Name: "a\na"}
	assert.ErrorContains(p.SyntacticVerify(), `invalid name "a\na"`)

	p = &Profile{Type: "Thing", Name: "LDC", Image: "a\na"}
	assert.ErrorContains(p.SyntacticVerify(), `invalid image "a\na"`)

	p = &Profile{Type: "Thing", Name: "LDC", URL: "a\na"}
	assert.ErrorContains(p.SyntacticVerify(), `invalid url "a\na"`)

	p = &Profile{Type: "Thing", Name: "LDC"}
	assert.ErrorContains(p.SyntacticVerify(), "nil follows")
	p = &Profile{Type: "Thing", Name: "LDC", Follows: []util.DataID{util.DataIDEmpty}}
	assert.ErrorContains(p.SyntacticVerify(), "invalid follow address")

	p = &Profile{Type: "Thing", Name: "LDC", Follows: []util.DataID{}}
	assert.ErrorContains(p.SyntacticVerify(), "nil extra")

	p = &Profile{
		Type:    "Person",
		Name:    "LDC",
		Follows: []util.DataID{},
		Extra: map[string]interface{}{
			"age": 23,
		},
	}
	assert.NoError(p.SyntacticVerify())
	data, err := json.Marshal(p)
	assert.NoError(err)

	jsonStr := `{"@type":"Person","name":"LDC","image":"","url":"","follows":[],"extra":{"age":23}}`
	assert.Equal(jsonStr, string(data))

	p2 := &Profile{}
	assert.NoError(p2.Unmarshal(p.Bytes()))
	assert.NoError(p2.SyntacticVerify())
	assert.Equal(p.Bytes(), p2.Bytes())
}
