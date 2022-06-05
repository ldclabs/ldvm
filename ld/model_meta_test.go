// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package ld

import (
	"encoding/json"
	"testing"

	"github.com/ldclabs/ldvm/util"
	"github.com/stretchr/testify/assert"
)

func TestModelMeta(t *testing.T) {
	assert := assert.New(t)

	sch := `
	type ID20 bytes
	type NameService struct {
		name    String        (rename "n")
		linked  nullable ID20 (rename "l")
		records [String]      (rename "rs")
	}
`

	var tx *ModelMeta
	assert.ErrorContains(tx.SyntacticVerify(), "nil pointer")

	assert.False(ModelNameReg.MatchString("test"))
	assert.False(ModelNameReg.MatchString("T"))
	assert.False(ModelNameReg.MatchString("T_t"))
	assert.False(ModelNameReg.MatchString("123"))
	assert.False(ModelNameReg.MatchString("_123"))
	assert.True(ModelNameReg.MatchString("Tt"))
	assert.True(ModelNameReg.MatchString("T1"))
	assert.True(ModelNameReg.MatchString("Name"))
	assert.True(ModelNameReg.MatchString("TestService"))

	tx = &ModelMeta{Name: "test"}
	assert.ErrorContains(tx.SyntacticVerify(), "invalid name")

	tx = &ModelMeta{Name: "Name", Threshold: 1}
	assert.ErrorContains(tx.SyntacticVerify(), "invalid threshold")

	tx = &ModelMeta{Name: "Name", Approver: &util.EthIDEmpty}
	assert.ErrorContains(tx.SyntacticVerify(), "invalid approver")

	tx = &ModelMeta{Name: "Name", Data: []byte("abc")}
	assert.ErrorContains(tx.SyntacticVerify(), "invalid data")

	tx = &ModelMeta{Name: "Name", Data: []byte(sch), Threshold: 1, Keepers: []util.EthID{util.EthIDEmpty}}
	assert.ErrorContains(tx.SyntacticVerify(), "invalid keeper")

	tx = &ModelMeta{Name: "Name", Data: []byte(sch)}
	assert.ErrorContains(tx.SyntacticVerify(), `NewIPLDModel "Name" error`)

	tx = &ModelMeta{
		Name:      "NameService",
		Threshold: 1,
		Keepers:   util.EthIDs{util.Signer1.Address(), util.Signer1.Address()},
		Data:      []byte(sch),
	}
	assert.NoError(tx.SyntacticVerify())
	cbordata, err := tx.Marshal()
	assert.NoError(err)
	jsondata, err := json.Marshal(tx)
	assert.NoError(err)

	assert.NotContains(string(jsondata), `"approver":`)
	assert.Contains(string(jsondata), `"name":"NameService"`)

	tx2 := &ModelMeta{}
	assert.NoError(tx2.Unmarshal(cbordata))
	assert.NoError(tx2.SyntacticVerify())

	cbordata2 := tx2.Bytes()
	jsondata2, err := json.Marshal(tx2)
	assert.Equal(string(jsondata), string(jsondata2))
	assert.Equal(cbordata, cbordata2)
}
