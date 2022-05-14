// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package ld

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIPLDModel(t *testing.T) {
	assert := assert.New(t)

	sch := `
	type SomeModel string
	`
	_, err := NewIPLDModel("SomeModel", []byte(sch))
	assert.ErrorContains(err, `NewIPLDModel "SomeModel" error: should be a map, list or struct`)

	sch = `
	type SomeModel {String:Any}
	`
	_, err = NewIPLDModel("SomeModel2", []byte(sch))
	assert.ErrorContains(err, `NewIPLDModel "SomeModel2" error: type not found`)

	sch = `
	abc
	`
	_, err = NewIPLDModel("SomeModel", []byte(sch))
	assert.ErrorContains(err, `unexpected token: "abc"`)

	sch = `
	type SomeModel {String:Any}
	`
	im, err := NewIPLDModel("SomeModel", []byte(sch))
	assert.NoError(err)
	assert.Equal("SomeModel", im.Name())
	assert.Equal(sch, string(im.Schema()))
	assert.Equal("map", im.Type().TypeKind().String())

	data, err := EncMode.Marshal(map[string]interface{}{"a": 1, "b": "a"})
	assert.NoError(err)
	assert.NoError(im.Valid(data))

	data, err = EncMode.Marshal([]interface{}{"a", "b", 1})
	assert.NoError(err)
	assert.ErrorContains(im.Valid(data), `IPLDModel "SomeModel" error`)

	sch = `
	type NameService struct {
		name    String        (rename "n")
		records [String]      (rename "rs")
	}
`
	im, err = NewIPLDModel("NameService", []byte(sch))
	assert.NoError(err)
	assert.Equal("struct", im.Type().TypeKind().String())

	data, err = EncMode.Marshal(map[string]interface{}{"n": "test", "rs": []string{"AAA"}})
	assert.NoError(err)
	assert.NoError(im.Valid(data))

	data, err = EncMode.Marshal(map[string]interface{}{"n": "test", "rs": []string{"AAA"}, "x": 1})
	assert.NoError(err)
	assert.ErrorContains(im.Valid(data), `invalid key: "x"`)

	data, err = EncMode.Marshal(map[string]interface{}{"n": "test"})
	assert.NoError(err)
	assert.ErrorContains(im.Valid(data), `missing required fields`)
}
