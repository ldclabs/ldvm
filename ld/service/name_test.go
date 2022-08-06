// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package service

import (
	"encoding/json"
	"testing"

	cborpatch "github.com/ldclabs/cbor-patch"
	"github.com/ldclabs/ldvm/util"
	"github.com/stretchr/testify/assert"
)

func TestName(t *testing.T) {
	assert := assert.New(t)

	var name *Name
	assert.ErrorContains(name.SyntacticVerify(), "nil pointer")

	name = &Name{Name: "ab=c"}
	assert.ErrorContains(name.SyntacticVerify(), `ToASCII error, idna: disallowed rune`)

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
		DID:     util.DataID{5, 6, 7, 8},
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

	// fmt.Println(string(data))
	assert.Equal(`{"name":"公信.com.","linked":"LD6L5yRJL2iYi9PbrhRru6uKfEAzDGHwUJ","records":["xn--vuq70b.com. IN A 10.0.0.1"],"did":"LDTZbknDJJaphGQrsiR4qU5LxyFX98cfQX"}`, string(data))

	nm, err := NameModel()
	assert.NoError(err)
	assert.NoError(nm.Valid(name.Bytes()))

	ipldops := cborpatch.Patch{
		{Op: "add", Path: "/rs/-", Value: util.MustMarshalCBOR("xn--vuq70b.com. IN AAAA ::1")},
	}
	data, err = nm.ApplyPatch(name.Bytes(), util.MustMarshalCBOR(ipldops))
	assert.NoError(err)

	name2 = &Name{}
	assert.NoError(name2.Unmarshal(data))
	assert.NoError(name2.SyntacticVerify())
	name2.DID = name.DID

	data, err = json.Marshal(name2)
	assert.NoError(err)

	// fmt.Println(string(data))
	assert.Equal(`{"name":"公信.com.","linked":"LD6L5yRJL2iYi9PbrhRru6uKfEAzDGHwUJ","records":["xn--vuq70b.com. IN A 10.0.0.1","xn--vuq70b.com. IN AAAA ::1"],"did":"LDTZbknDJJaphGQrsiR4qU5LxyFX98cfQX"}`, string(data))
}
