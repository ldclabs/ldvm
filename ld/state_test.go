// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package ld

import (
	"encoding/json"
	"testing"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/stretchr/testify/assert"

	"github.com/ldclabs/ldvm/constants"
)

func TestState(t *testing.T) {
	assert := assert.New(t)

	var s *State
	assert.ErrorContains(s.SyntacticVerify(), "nil pointer")

	s = &State{}
	assert.ErrorContains(s.SyntacticVerify(), "nil accounts")

	s = &State{
		Accounts: make(map[string]ids.ID),
	}
	assert.ErrorContains(s.SyntacticVerify(), "nil ledgers")

	s = &State{
		Accounts: make(map[string]ids.ID),
		Ledgers:  make(map[string]ids.ID),
	}
	assert.ErrorContains(s.SyntacticVerify(), "nil datas")

	s = &State{
		Accounts: make(map[string]ids.ID),
		Ledgers:  make(map[string]ids.ID),
		Datas:    make(map[string]ids.ID),
	}
	assert.ErrorContains(s.SyntacticVerify(), "nil models")

	s = NewState(ids.Empty)
	assert.NoError(s.SyntacticVerify())
	cbordata, err := s.Marshal()
	assert.NoError(err)
	jsondata, err := json.Marshal(s)
	assert.NoError(err)

	// fmt.Println(string(jsondata))
	assert.Equal(`{"parent":"11111111111111111111111111111111LpoYY","accounts":{},"ledgers":{},"datas":{},"models":{},"id":"LNNJn3JUTcSQMCY5HsFnZ3A75ZXDY5EtQyqc7XbMLaW8N4xLj"}`,
		string(jsondata))

	s2 := &State{}
	assert.NoError(s2.Unmarshal(cbordata))
	assert.NoError(s2.SyntacticVerify())

	cbordata2 := s2.Bytes()
	jsondata2, err := json.Marshal(s2)
	assert.Equal(string(jsondata), string(jsondata2))
	assert.Equal(s.ID, s2.ID)
	assert.Equal(cbordata, cbordata2)

	s.Accounts[string(constants.GenesisAccount[:])] = ids.ID{1, 2, 3}
	assert.NoError(s.SyntacticVerify())
	assert.NotEqual(s.ID, s2.ID)
	assert.NotEqual(s.Bytes(), s2.Bytes())

	s3 := s.Clone()
	assert.NoError(s3.SyntacticVerify())
	assert.Equal(s.ID, s3.ID)
	assert.Equal(s.Bytes(), s3.Bytes())

	s3.Ledgers[string(constants.GenesisAccount[:])] = ids.ID{1, 2, 3}
	assert.NoError(s3.SyntacticVerify())
	assert.NotEqual(s.ID, s3.ID)
	assert.NotEqual(s.Bytes(), s3.Bytes())
}
