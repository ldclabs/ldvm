// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package ld

import (
	"encoding/json"
	"math/big"
	"testing"
	"time"

	"github.com/ldclabs/ldvm/constants"
	"github.com/ldclabs/ldvm/util"
	"github.com/stretchr/testify/assert"
)

func TestTxUpdater(t *testing.T) {
	assert := assert.New(t)

	var tx *TxUpdater
	assert.ErrorContains(tx.SyntacticVerify(), "nil pointer")

	tx = &TxUpdater{Token: &util.TokenSymbol{'a', 'b', 'c'}}
	assert.ErrorContains(tx.SyntacticVerify(), `invalid token symbol "0x6162630000000000000000000000000000000000"`)

	tx = &TxUpdater{Amount: big.NewInt(-1)}
	assert.ErrorContains(tx.SyntacticVerify(), "invalid amount")

	tx = &TxUpdater{Keepers: &util.EthIDs{}}
	assert.ErrorContains(tx.SyntacticVerify(), "nil threshold")
	tx2 := &TxUpdater{}
	assert.NoError(tx2.Unmarshal(tx.Bytes()))
	assert.ErrorContains(tx2.SyntacticVerify(), "nil threshold")

	tx = &TxUpdater{Threshold: Uint8Ptr(1)}
	assert.ErrorContains(tx.SyntacticVerify(), "nil keepers")
	tx = &TxUpdater{Threshold: Uint8Ptr(1), Keepers: &util.EthIDs{}}
	assert.ErrorContains(tx.SyntacticVerify(), "invalid threshold, expected <= 0, got 1")
	tx = &TxUpdater{Threshold: Uint8Ptr(1), Keepers: &util.EthIDs{util.EthIDEmpty}}
	assert.ErrorContains(tx.SyntacticVerify(), "invalid keeper")

	tx = &TxUpdater{ApproveList: []TxType{TypeCreateData - 1}}
	assert.ErrorContains(tx.SyntacticVerify(), "invalid TxType")

	tx = &TxUpdater{ApproveList: []TxType{TypeDeleteData + 1}}
	assert.ErrorContains(tx.SyntacticVerify(), "invalid TxType")

	tx = &TxUpdater{Token: &util.NativeToken}
	assert.NoError(tx.SyntacticVerify())

	tx = &TxUpdater{
		ID:        &util.DataID{1, 2, 3},
		Version:   1,
		Threshold: Uint8Ptr(1),
		Keepers:   &util.EthIDs{util.Signer1.Address(), util.Signer1.Address()},
		Approver:  &util.EthIDEmpty,
		Token:     &util.NativeToken,
		To:        &constants.GenesisAccount,
		Amount:    big.NewInt(1000),
		KSig:      &util.Signature{1, 2, 3},
		Expire:    uint64(time.Now().Unix()),
		Data:      []byte(`"Hello, world!"`),
	}
	assert.NoError(tx.SyntacticVerify())
	cbordata, err := tx.Marshal()
	assert.NoError(err)
	jsondata, err := json.Marshal(tx)
	assert.NoError(err)
	assert.Contains(string(jsondata), `"approver":"0x0000000000000000000000000000000000000000"`)
	assert.Contains(string(jsondata), `"token":""`)
	assert.Contains(string(jsondata), `"data":"Hello, world!"`)
	assert.NotContains(string(jsondata), `"mid":`)
	assert.NotContains(string(jsondata), `"mSig":`)

	tx2 = &TxUpdater{}
	assert.NoError(tx2.Unmarshal(cbordata))
	assert.NoError(tx2.SyntacticVerify())
	cbordata2 := tx2.Bytes()
	jsondata2, err := json.Marshal(tx2)
	assert.Equal(string(jsondata), string(jsondata2))
	assert.Equal(cbordata, cbordata2)
}
