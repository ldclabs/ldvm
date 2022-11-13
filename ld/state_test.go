// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package ld

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ldclabs/ldvm/ids"
)

func TestState(t *testing.T) {
	assert := assert.New(t)

	var s *State
	assert.ErrorContains(s.SyntacticVerify(), "nil pointer")

	s = &State{}
	assert.ErrorContains(s.SyntacticVerify(), "nil accounts")

	s = &State{
		Accounts: make(map[string]ids.ID32),
	}
	assert.ErrorContains(s.SyntacticVerify(), "nil ledgers")

	s = &State{
		Accounts: make(map[string]ids.ID32),
		Ledgers:  make(map[string]ids.ID32),
	}
	assert.ErrorContains(s.SyntacticVerify(), "nil datas")

	s = &State{
		Accounts: make(map[string]ids.ID32),
		Ledgers:  make(map[string]ids.ID32),
		Datas:    make(map[string]ids.ID32),
	}
	assert.ErrorContains(s.SyntacticVerify(), "nil models")

	s = NewState(ids.EmptyID32)
	assert.NoError(s.SyntacticVerify())
	cbordata, err := s.Marshal()
	require.NoError(t, err)
	jsondata, err := json.Marshal(s)
	require.NoError(t, err)

	// fmt.Println(string(jsondata))
	assert.Equal(`{"parent":"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAACeYpGX","accounts":{},"ledgers":{},"datas":{},"models":{},"id":"K_pzrWr2Xqa8eiIkQyUCwO6b6Lh9LMvDRyFxGJVIYdvFo_Or"}`,
		string(jsondata))

	s2 := &State{}
	assert.NoError(s2.Unmarshal(cbordata))
	assert.NoError(s2.SyntacticVerify())

	cbordata2 := s2.Bytes()
	jsondata2, _ := json.Marshal(s2)
	assert.Equal(string(jsondata), string(jsondata2))
	assert.Equal(s.ID, s2.ID)
	assert.Equal(cbordata, cbordata2)

	s.Accounts[string(ids.GenesisAccount[:])] = ids.ID32{1, 2, 3}
	assert.NoError(s.SyntacticVerify())
	assert.NotEqual(s.ID, s2.ID)
	assert.NotEqual(s.Bytes(), s2.Bytes())

	s3 := s.Clone()
	assert.NoError(s3.SyntacticVerify())
	assert.Equal(s.ID, s3.ID)
	assert.Equal(s.Bytes(), s3.Bytes())

	s3.Ledgers[string(ids.GenesisAccount[:])] = ids.ID32{1, 2, 3}
	assert.NoError(s3.SyntacticVerify())
	assert.NotEqual(s.ID, s3.ID)
	assert.NotEqual(s.Bytes(), s3.Bytes())
}
