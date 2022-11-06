// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package ld

import (
	"encoding/json"
	"testing"

	"github.com/ldclabs/ldvm/util"
	"github.com/ldclabs/ldvm/util/signer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestModelInfo(t *testing.T) {
	assert := assert.New(t)

	sc := `
	type ID20 bytes
	type NameService struct {
		name    String        (rename "n")
		linked  nullable ID20 (rename "l")
		records [String]      (rename "rs")
	}
`

	var tx *ModelInfo
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

	tx = &ModelInfo{Name: "test"}
	assert.ErrorContains(tx.SyntacticVerify(), "invalid name")

	tx = &ModelInfo{Name: "Name", Threshold: 1}
	assert.ErrorContains(tx.SyntacticVerify(), "invalid threshold")

	tx = &ModelInfo{Name: "Name", Schema: "abc"}
	assert.ErrorContains(tx.SyntacticVerify(), "invalid schema string")

	tx = &ModelInfo{Name: "Name", Schema: sc, Approver: signer.Key(util.AddressEmpty[:])}
	assert.ErrorContains(tx.SyntacticVerify(), "invalid approver")

	tx = &ModelInfo{Name: "Name", Schema: sc, Threshold: 1, Keepers: signer.Keys{signer.Key(util.AddressEmpty[:])}}
	assert.ErrorContains(tx.SyntacticVerify(), "empty secp256k1 key")

	tx = &ModelInfo{Name: "Name", Schema: sc, Approver: signer.Key{}}
	assert.ErrorContains(tx.SyntacticVerify(), `empty key`)

	tx = &ModelInfo{Name: "Name", Schema: sc}
	assert.ErrorContains(tx.SyntacticVerify(), `NewIPLDModel("Name")`)

	tx = &ModelInfo{
		Name:      "NameService",
		Threshold: 1,
		Keepers:   signer.Keys{signer.Signer1.Key(), signer.Signer1.Key()},
		Schema:    sc,
	}
	assert.ErrorContains(tx.SyntacticVerify(),
		"duplicate key jbl8fOziScK5i9wCJsxMKle_UvwKxwPH")

	tx = &ModelInfo{
		Name:      "NameService",
		Threshold: 1,
		Keepers:   signer.Keys{signer.Signer1.Key()},
		Schema:    sc,
	}
	assert.NoError(tx.SyntacticVerify())
	cbordata, err := tx.Marshal()
	require.NoError(t, err)
	jsondata, err := json.Marshal(tx)
	require.NoError(t, err)
	// fmt.Println(string(jsondata))
	assert.Equal(`{"name":"NameService","threshold":1,"keepers":["jbl8fOziScK5i9wCJsxMKle_UvwKxwPH"],"schema":"\n\ttype ID20 bytes\n\ttype NameService struct {\n\t\tname    String        (rename \"n\")\n\t\tlinked  nullable ID20 (rename \"l\")\n\t\trecords [String]      (rename \"rs\")\n\t}\n","id":"AAAAAAAAAAAAAAAAAAAAAAAAAADzaDye"}`, string(jsondata))

	tx2 := &ModelInfo{}
	assert.NoError(tx2.Unmarshal(cbordata))
	assert.NoError(tx2.SyntacticVerify())

	cbordata2 := tx2.Bytes()
	jsondata2, _ := json.Marshal(tx2)
	assert.Equal(string(jsondata), string(jsondata2))
	assert.Equal(cbordata, cbordata2)

	tx2.Approver = signer.Signer3.Key()
	assert.NoError(tx2.SyntacticVerify())
	jsondata, err = json.Marshal(tx2)
	require.NoError(t, err)
	// fmt.Println(string(jsondata))
	assert.Equal(`{"name":"NameService","threshold":1,"keepers":["jbl8fOziScK5i9wCJsxMKle_UvwKxwPH"],"approver":"OVlX-75gy0DuaRuz2k5QnlFVSuKOJezRd4CQdkIjkn5pYt0F","schema":"\n\ttype ID20 bytes\n\ttype NameService struct {\n\t\tname    String        (rename \"n\")\n\t\tlinked  nullable ID20 (rename \"l\")\n\t\trecords [String]      (rename \"rs\")\n\t}\n","id":"AAAAAAAAAAAAAAAAAAAAAAAAAADzaDye"}`, string(jsondata))

	assert.NotEqual(cbordata, tx2.Bytes())
	tx3 := &ModelInfo{}
	assert.NoError(tx3.Unmarshal(tx2.Bytes()))
	assert.NoError(tx3.SyntacticVerify())
	assert.Equal(tx2.Bytes(), tx3.Bytes())
}
