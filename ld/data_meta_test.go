// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package ld

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ldclabs/ldvm/util"
)

func TestDataMeta(t *testing.T) {
	assert := assert.New(t)

	var tx *DataMeta
	assert.ErrorContains(tx.SyntacticVerify(), "nil pointer")

	tx = &DataMeta{Threshold: 1}
	assert.ErrorContains(tx.SyntacticVerify(), "invalid threshold")

	tx = &DataMeta{Keepers: []util.EthID{util.EthIDEmpty}}
	assert.ErrorContains(tx.SyntacticVerify(), "invalid keeper")

	tx = &DataMeta{Version: 1, Approver: &util.EthIDEmpty}
	assert.ErrorContains(tx.SyntacticVerify(), "invalid approver")

	tx = &DataMeta{ApproveList: []TxType{TypeDeleteData + 1}}
	assert.ErrorContains(tx.SyntacticVerify(), "invalid TxType")
	tx = &DataMeta{
		Data: []byte(`42`),
		KSig: util.Signature{1, 2, 3},
	}
	assert.ErrorContains(tx.SyntacticVerify(), "DeriveSigner: recovery failed")

	kSig, err := util.Signer1.Sign([]byte(`42`))
	assert.NoError(err)
	tx = &DataMeta{
		Version:   1,
		Threshold: 1,
		Keepers:   util.EthIDs{util.Signer1.Address(), util.Signer2.Address()},
		Data:      []byte(`42`),
		KSig:      kSig,
	}
	assert.NoError(tx.SyntacticVerify())
	cbordata, err := tx.Marshal()
	assert.NoError(err)
	jsondata, err := json.Marshal(tx)
	assert.NoError(err)

	assert.Contains(string(jsondata), `"mid":"LM111111111111111111116DBWJs"`)
	assert.Contains(string(jsondata), `"keepers":["0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC","0x44171C37Ff5D7B7bb8dcad5C81f16284A229e641"]`)
	assert.Contains(string(jsondata), `"kSig":"`)
	assert.Contains(string(jsondata), `"data":42`)
	assert.NotContains(string(jsondata), `"approver":`)

	tx2 := &DataMeta{}
	assert.NoError(tx2.Unmarshal(cbordata))
	assert.NoError(tx2.SyntacticVerify())

	cbordata2 := tx2.Bytes()
	jsondata2, err := json.Marshal(tx2)
	assert.Equal(string(jsondata), string(jsondata2))
	assert.Equal(cbordata, cbordata2)
}
