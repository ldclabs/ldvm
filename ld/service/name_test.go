// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package service

import (
	"encoding/json"
	"testing"

	"github.com/ldclabs/ldvm/util"
	"github.com/stretchr/testify/assert"
)

func TestName(t *testing.T) {
	assert := assert.New(t)

	var name *Name
	assert.ErrorContains(name.SyntacticVerify(), "nil pointer")

	name = &Name{Name: "ab=c"}
	assert.ErrorContains(name.SyntacticVerify(), `NewDN("ab=c"): ToASCII error, idna: disallowed rune`)

	name = &Name{Name: "xn--vuq70b.com."}
	assert.ErrorContains(name.SyntacticVerify(), `"xn--vuq70b.com." is not unicode form`)

	name = &Name{Name: "xn--vuq70b"}
	assert.ErrorContains(name.SyntacticVerify(), `"xn--vuq70b" is not unicode form`)

	name = &Name{Name: "公信"}
	assert.ErrorContains(name.SyntacticVerify(), `nil records`)

	address := util.DataID{1, 2, 3, 4}
	name = &Name{
		Name:    "公信.com.",
		Linked:  &address,
		Records: []string{},
	}
	assert.NoError(name.SyntacticVerify())

	name2 := &Name{}
	assert.NoError(name2.Unmarshal(name.Bytes()))
	assert.Equal(name.Bytes(), name2.Bytes())

	name.Records = append(name.Records, "xn--vuq70b.com. IN A 10.0.0.1")
	assert.NoError(name.SyntacticVerify())
	assert.NotEqual(name.Bytes(), name2.Bytes())

	data, err := json.Marshal(name)
	assert.NoError(err)

	jsonStr := `{"name":"公信.com.","linked":"LD6L5yRJL2iYi9PbrhRru6uKfEAzDGHwUJ","records":["xn--vuq70b.com. IN A 10.0.0.1"]}`
	assert.Equal(jsonStr, string(data))

	nm, err := NameModel()
	assert.NoError(err)
	assert.NoError(nm.Valid(name.Bytes()))
}
