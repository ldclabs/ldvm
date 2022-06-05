// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package ld

import (
	"encoding/json"
	"math/big"
	"testing"

	"github.com/ldclabs/ldvm/util"
	"github.com/stretchr/testify/assert"
)

func TestTxAccounter(t *testing.T) {
	assert := assert.New(t)

	var tx *TxAccounter
	assert.ErrorContains(tx.SyntacticVerify(), "nil pointer")

	tx = &TxAccounter{Amount: big.NewInt(-1)}
	assert.ErrorContains(tx.SyntacticVerify(), "invalid amount")

	tx = &TxAccounter{Keepers: util.EthIDs{}}
	assert.ErrorContains(tx.SyntacticVerify(), "nil threshold")
	tx = &TxAccounter{Threshold: Uint8Ptr(1)}
	assert.ErrorContains(tx.SyntacticVerify(), "nil keepers")
	tx = &TxAccounter{Threshold: Uint8Ptr(1), Keepers: []util.EthID{}}
	assert.ErrorContains(tx.SyntacticVerify(), "invalid threshold, expected <= 0, got 1")
	tx = &TxAccounter{Threshold: Uint8Ptr(1), Keepers: []util.EthID{util.EthIDEmpty}}
	assert.ErrorContains(tx.SyntacticVerify(), "invalid keeper")

	tx = &TxAccounter{ApproveList: []TxType{TypeDeleteData + 1}}
	assert.ErrorContains(tx.SyntacticVerify(), "invalid TxType")

	tx = &TxAccounter{
		Threshold: Uint8Ptr(1),
		Keepers:   util.EthIDs{util.Signer1.Address(), util.Signer1.Address()},
		Amount:    big.NewInt(1000),
		Data:      []byte(`42`),
	}
	assert.NoError(tx.SyntacticVerify())
	cbordata, err := tx.Marshal()
	assert.NoError(err)
	jsondata, err := json.Marshal(tx)
	assert.NoError(err)

	assert.NotContains(string(jsondata), `"approver":`)
	assert.Contains(string(jsondata), `"data":42`)

	tx2 := &TxAccounter{}
	assert.NoError(tx2.Unmarshal(cbordata))
	assert.NoError(tx2.SyntacticVerify())

	cbordata2 := tx2.Bytes()
	jsondata2, err := json.Marshal(tx2)
	assert.Equal(string(jsondata), string(jsondata2))
	assert.Equal(cbordata, cbordata2)
}
