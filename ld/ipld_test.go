// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package ld

import (
	"testing"

	cborpatch "github.com/ldclabs/cbor-patch"
	"github.com/ldclabs/ldvm/ids"
	"github.com/ldclabs/ldvm/util/encoding"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIPLDModel(t *testing.T) {
	assert := assert.New(t)

	sc := `
	type SomeModel string
	`
	_, err := NewIPLDModel("SomeModel", sc)
	assert.ErrorContains(err, `NewIPLDModel("SomeModel"): should be a map, list or struct`)

	sc = `
	type SomeModel {String:Any}
	`
	_, err = NewIPLDModel("SomeModel2", sc)
	assert.ErrorContains(err, `NewIPLDModel("SomeModel2"): type not found`)

	sc = `
	abc
	`
	_, err = NewIPLDModel("SomeModel", sc)
	assert.ErrorContains(err, `unexpected token: "abc"`)

	sc = `
	type SomeModel {String:Any}
	`
	im, err := NewIPLDModel("SomeModel", sc)
	require.NoError(t, err)
	assert.Equal("SomeModel", im.Name())
	assert.Equal(sc, string(im.Schema()))
	assert.Equal("map", im.Type().TypeKind().String())

	data, err := encoding.MarshalCBOR(map[string]interface{}{"a": 1, "b": "a"})
	require.NoError(t, err)
	assert.NoError(im.Valid(data))

	data, err = encoding.MarshalCBOR([]interface{}{"a", "b", 1})
	require.NoError(t, err)
	assert.ErrorContains(im.Valid(data), `IPLDModel("SomeModel").Valid: decode:`)

	sc = `
	type NameService struct {
		name    String        (rename "n")
		records [String]      (rename "rs")
	}
`
	im, err = NewIPLDModel("NameService", sc)
	require.NoError(t, err)
	assert.Equal("struct", im.Type().TypeKind().String())

	data, err = encoding.MarshalCBOR(map[string]interface{}{"n": "test", "rs": []string{"AAA"}})
	require.NoError(t, err)
	assert.NoError(im.Valid(data))

	data, err = encoding.MarshalCBOR(map[string]interface{}{"n": "test", "rs": []string{"AAA"}, "x": 1})
	require.NoError(t, err)
	assert.ErrorContains(im.Valid(data), `invalid key: "x"`)

	data, err = encoding.MarshalCBOR(map[string]interface{}{"n": "test"})
	require.NoError(t, err)
	assert.ErrorContains(im.Valid(data), `missing required fields`)
}

func TestIPLDModelApplyPatch(t *testing.T) {
	assert := assert.New(t)

	sc := `
	type ID20 bytes
	type ProfileService struct {
		type       Int             (rename "t")
		name       String          (rename "n")
		image      String          (rename "i")
		url        String          (rename "u")
		follows    [ID20]          (rename "fs")
		members    optional [ID20] (rename "ms")
	}
`

	type profile struct {
		Type    uint16                 `cbor:"t"`
		Name    string                 `cbor:"n"`
		Image   string                 `cbor:"i"`
		URL     string                 `cbor:"u"`
		Follows ids.IDList[ids.DataID] `cbor:"fs"`
		Members ids.IDList[ids.DataID] `cbor:"ms,omitempty"`
	}

	mo, err := NewIPLDModel("ProfileService", sc)
	require.NoError(t, err)

	v1 := &profile{
		Type:    0,
		Name:    "Test",
		Follows: ids.IDList[ids.DataID]{},
	}

	od := encoding.MustMarshalCBOR(v1)
	ipldops := cborpatch.Patch{
		{Op: "replace", Path: "/n", Value: encoding.MustMarshalCBOR("John")},
		{Op: "replace", Path: "/t", Value: encoding.MustMarshalCBOR(uint16(1))},
	}

	data, err := mo.ApplyPatch(od, encoding.MustMarshalCBOR(ipldops))
	require.NoError(t, err)

	v2 := &profile{}
	assert.NoError(encoding.UnmarshalCBOR(data, v2))
	assert.Equal("John", v2.Name)
	assert.Equal(uint16(1), v2.Type)

	ipldops = cborpatch.Patch{
		{Op: "test", Path: "/n", Value: encoding.MustMarshalCBOR("Test")},
	}

	_, err = mo.ApplyPatch(od, encoding.MustMarshalCBOR(ipldops))
	require.NoError(t, err)

	_, err = mo.ApplyPatch(data, encoding.MustMarshalCBOR(ipldops))
	assert.ErrorContains(err,
		`IPLDModel("ProfileService").ApplyPatch: test operation for path "/n" failed, expected "Test", got "John"`)

	ipldops = cborpatch.Patch{
		{Op: "add", Path: "/x", Value: encoding.MustMarshalCBOR("Test")},
	}

	_, err = mo.ApplyPatch(data, encoding.MustMarshalCBOR(ipldops))
	assert.ErrorContains(err,
		`invalid key: "x" is not a field in type ProfileService`)
}
