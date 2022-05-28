// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package ld

import (
	"encoding/json"
	"testing"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/stretchr/testify/assert"

	"github.com/ldclabs/ldvm/constants"
	"github.com/ldclabs/ldvm/util"
)

func TestState(t *testing.T) {
	assert := assert.New(t)

	var s *State
	assert.ErrorContains(s.SyntacticVerify(), "nil pointer")

	s = &State{}
	assert.ErrorContains(s.SyntacticVerify(), "nil accounts")

	s = &State{
		Accounts: make(map[util.EthID]ids.ID),
	}
	assert.ErrorContains(s.SyntacticVerify(), "nil datas")

	s = &State{
		Accounts: make(map[util.EthID]ids.ID),
		Datas:    make(map[util.DataID]ids.ID),
	}
	assert.ErrorContains(s.SyntacticVerify(), "nil models")

	s = NewState(ids.Empty)
	assert.NoError(s.SyntacticVerify())
	cbordata, err := s.Marshal()
	assert.NoError(err)
	jsondata, err := json.Marshal(s)
	assert.NoError(err)

	assert.Equal(`{"parent":"11111111111111111111111111111111LpoYY","accounts":{},"datas":{},"models":{},"id":"24LaJiEty1cprbCodNPvHxch3KyirdvxyW1Ksyz1VsFYXEst2E"}`,
		string(jsondata))

	s2 := &State{}
	assert.NoError(s2.Unmarshal(cbordata))
	assert.NoError(s2.SyntacticVerify())

	cbordata2 := s2.Bytes()
	jsondata2, err := json.Marshal(s2)
	assert.Equal(string(jsondata), string(jsondata2))
	assert.Equal(s.ID, s2.ID)
	assert.Equal(cbordata, cbordata2)

	s.Accounts[constants.GenesisAccount] = ids.ID{1, 2, 3}
	assert.NoError(s.SyntacticVerify())
	cbordata, err = s.Marshal()
	assert.NoError(err)
	assert.NotEqual(s.ID, s2.ID)
	assert.NotEqual(s.Bytes(), s2.Bytes())
}
