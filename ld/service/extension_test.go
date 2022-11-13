// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package service

import (
	"testing"

	"github.com/ldclabs/ldvm/ids"
	"github.com/stretchr/testify/assert"
)

func TestExtensions(t *testing.T) {
	assert := assert.New(t)

	var es Extensions
	assert.ErrorContains(es.SyntacticVerify(), "nil pointer")

	es = make(Extensions, 0)
	assert.NoError(es.SyntacticVerify())

	es = Extensions{nil}
	assert.ErrorContains(es.SyntacticVerify(), "nil pointer at 0")

	es = Extensions{{Title: ""}}
	assert.ErrorContains(es.SyntacticVerify(), `invalid title "" at 0`)

	es = Extensions{{Title: "Test"}}
	assert.ErrorContains(es.SyntacticVerify(), `nil properties at 0`)

	es = Extensions{{
		Title:      "Test",
		Properties: map[string]interface{}{},
		DataID:     &ids.DataID{1, 2, 3},
	}}
	assert.ErrorContains(es.SyntacticVerify(), `nil model id at 0`)

	es = Extensions{{
		Title:      "Test",
		Properties: map[string]interface{}{},
		ModelID:    &ids.ModelID{1, 2, 3},
	}}
	assert.ErrorContains(es.SyntacticVerify(), `no data id at 0, model id be nil`)

	es = Extensions{{
		Title:      "Test",
		Properties: map[string]interface{}{},
		DataID:     &ids.EmptyDataID,
		ModelID:    &ids.EmptyModelID,
	}}
	assert.ErrorContains(es.SyntacticVerify(), `invalid data id at 0`)

	es = Extensions{{
		Title:      "Test",
		Properties: map[string]interface{}{},
		DataID:     &ids.DataID{1, 2, 3},
		ModelID:    &ids.EmptyModelID,
	}}
	assert.ErrorContains(es.SyntacticVerify(), `invalid model id at 0`)

	es = Extensions{{
		Title:      "Test",
		Properties: map[string]interface{}{},
	}}
	assert.NoError(es.SyntacticVerify())

	es = Extensions{{
		Title:      "Test",
		Properties: map[string]interface{}{},
		DataID:     &ids.DataID{1, 2, 3},
		ModelID:    &ids.ModelID{1, 2, 3},
	}}
	assert.NoError(es.SyntacticVerify())

	es = Extensions{{
		Title:      "Test",
		Properties: map[string]interface{}{},
	}, {
		Title:      "Test",
		Properties: map[string]interface{}{},
	}}
	assert.ErrorContains(es.SyntacticVerify(), `"Test" exists in extensions at 1`)

	es = Extensions{{
		Title:      "Test",
		Properties: map[string]interface{}{},
		DataID:     &ids.DataID{1, 2, 3},
		ModelID:    &ids.ModelID{1, 2, 3},
	}, {
		Title:      "Test",
		Properties: map[string]interface{}{},
		DataID:     &ids.DataID{4, 5, 6},
		ModelID:    &ids.ModelID{1, 2, 3},
	}}
	assert.ErrorContains(es.SyntacticVerify(), `"Test" exists in extensions at 1`)

	es = Extensions{{
		Title:      "Test",
		Properties: map[string]interface{}{},
	}, {
		Title:      "Test",
		Properties: map[string]interface{}{},
		DataID:     &ids.DataID{4, 5, 6},
		ModelID:    &ids.ModelID{1, 2, 3},
	}}
	assert.NoError(es.SyntacticVerify())
}
