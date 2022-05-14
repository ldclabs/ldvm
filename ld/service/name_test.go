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
	assert.ErrorContains(name.SyntacticVerify(), `converts "ab=c" error: idna: disallowed rune`)

	name = &Name{Name: "你好"}
	assert.ErrorContains(name.SyntacticVerify(), `"你好" is not ASCII form (IDNA2008)`)

	address := util.DataID{1, 2, 3, 4}
	name = &Name{
		Name:    "xn--vuq70b.com",
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

	jsonStr := `{"name":"xn--vuq70b.com","linked":"LD6L5yRJL2iYi9PbrhRru6uKfEAzDGHwUJ","records":["xn--vuq70b.com. IN A 10.0.0.1"],"displayName":"公信.com"}`
	assert.Equal(jsonStr, string(data))

	nm, err := NameModel()
	assert.NoError(err)
	assert.NoError(nm.Valid(name.Bytes()))
}
