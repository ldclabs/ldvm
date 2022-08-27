// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package service

import (
	"encoding/json"
	"testing"

	cborpatch "github.com/ldclabs/cbor-patch"
	"github.com/ldclabs/ldvm/ld"
	"github.com/ldclabs/ldvm/util"
	"github.com/stretchr/testify/assert"
)

func TestProfile(t *testing.T) {
	assert := assert.New(t)

	var p *Profile
	assert.ErrorContains(p.SyntacticVerify(), "nil pointer")

	p = &Profile{Type: 0, Name: ""}
	assert.ErrorContains(p.SyntacticVerify(), `invalid name ""`)

	p = &Profile{Type: 0, Name: "a\na"}
	assert.ErrorContains(p.SyntacticVerify(), `invalid name "a\na"`)

	p = &Profile{Type: 0, Name: "LDC", Desc: "\nLinked Data chain"}
	assert.ErrorContains(p.SyntacticVerify(), `invalid description "\nLinked Data chain"`)

	p = &Profile{Type: 0, Name: "LDC", Image: "a\na"}
	assert.ErrorContains(p.SyntacticVerify(), `invalid image "a\na"`)

	p = &Profile{Type: 0, Name: "LDC", URL: "a\na"}
	assert.ErrorContains(p.SyntacticVerify(), `invalid url "a\na"`)

	p = &Profile{Type: 0, Name: "LDC"}
	assert.ErrorContains(p.SyntacticVerify(), "nil follows")
	p = &Profile{Type: 0, Name: "LDC", Follows: util.DataIDs{util.DataIDEmpty}}
	assert.ErrorContains(p.SyntacticVerify(), "invalid follows")

	p = &Profile{Type: 0, Name: "LDC", Follows: util.DataIDs{}}
	assert.ErrorContains(p.SyntacticVerify(), "nil extensions")

	p = &Profile{
		Type:    1,
		Name:    "LDC",
		Follows: util.DataIDs{},
		Extensions: []*Extension{{
			ModelID: ld.JSONModelID,
			Title:   "test",
			Properties: map[string]interface{}{
				"age": 23,
			},
		}},
	}
	assert.NoError(p.SyntacticVerify())
	data, err := json.Marshal(p)
	assert.NoError(err)

	// fmt.Println(string(data))
	assert.Equal(`{"type":"Person","name":"LDC","description":"","image":"","url":"","follows":[],"extensions":[{"mid":"1111111111111111111L17Xp3","title":"test","properties":{"age":23}}],"did":"11111111111111111111111111111111LpoYY"}`, string(data))

	p2 := &Profile{}
	assert.NoError(p2.Unmarshal(p.Bytes()))
	assert.NoError(p2.SyntacticVerify())
	assert.Equal(p.Bytes(), p2.Bytes())

	p.Extensions[0].Properties["email"] = "ldc@example.com"
	assert.NoError(p.SyntacticVerify())
	assert.NotEqual(p.Bytes(), p2.Bytes())

	pm, err := ProfileModel()
	assert.NoError(err)
	assert.NoError(pm.Valid(p.Bytes()))

	p.Members = util.DataIDs{{1, 2, 3}}
	assert.NoError(p.SyntacticVerify())
	assert.NoError(pm.Valid(p.Bytes()))

	p2 = &Profile{}
	assert.NoError(p2.Unmarshal(p.Bytes()))
	assert.NoError(p2.SyntacticVerify())
	assert.Equal(p.Bytes(), p2.Bytes())

	data, err = json.Marshal(p2)
	assert.NoError(err)

	// fmt.Println(string(data))
	assert.Equal(`{"type":"Person","name":"LDC","description":"","image":"","url":"","follows":[],"members":["SkB7qHwfMsyF2PgrjhMvtFxJKhuR5ZfVoW9VATWRV4P9jV7J"],"extensions":[{"mid":"1111111111111111111L17Xp3","title":"test","properties":{"age":23,"email":"ldc@example.com"}}],"did":"11111111111111111111111111111111LpoYY"}`, string(data))

	ipldops := cborpatch.Patch{
		{Op: "replace", Path: "/u", Value: util.MustMarshalCBOR("https://ldclabs.org")},
		{Op: "add", Path: "/fs/-", Value: util.MustMarshalCBOR(util.DataID{1, 2, 3})},
		{Op: "remove", Path: "/ms/0"},
	}
	data, err = pm.ApplyPatch(p.Bytes(), util.MustMarshalCBOR(ipldops))
	assert.NoError(err)

	p2 = &Profile{}
	assert.NoError(p2.Unmarshal(data))
	assert.NoError(p2.SyntacticVerify())
	assert.NotEqual(p.Bytes(), p2.Bytes())

	data, err = json.Marshal(p2)
	assert.NoError(err)
	// fmt.Println(string(data))
	assert.Equal(`{"type":"Person","name":"LDC","description":"","image":"","url":"https://ldclabs.org","follows":["SkB7qHwfMsyF2PgrjhMvtFxJKhuR5ZfVoW9VATWRV4P9jV7J"],"extensions":[{"mid":"1111111111111111111L17Xp3","title":"test","properties":{"age":23,"email":"ldc@example.com"}}],"did":"11111111111111111111111111111111LpoYY"}`, string(data))
}
