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

func TestTxTransfer(t *testing.T) {
	assert := assert.New(t)

	var tx *TxTransfer
	assert.ErrorContains(tx.SyntacticVerify(), "nil pointer")

	tx = &TxTransfer{Token: &util.TokenSymbol{'a', 'b', 'c'}}
	assert.ErrorContains(tx.SyntacticVerify(), `invalid token symbol "0x6162630000000000000000000000000000000000"`)

	tx = &TxTransfer{Amount: big.NewInt(-1)}
	assert.ErrorContains(tx.SyntacticVerify(), "invalid amount")

	tx = &TxTransfer{Amount: big.NewInt(0)}
	assert.NoError(tx.SyntacticVerify())

	tx = &TxTransfer{Token: &util.NativeToken}
	assert.NoError(tx.SyntacticVerify())

	tx = &TxTransfer{
		Nonce:  1,
		Token:  &util.NativeToken,
		To:     &constants.GenesisAccount,
		Amount: big.NewInt(1000),
		Expire: uint64(time.Now().Unix()),
		Data:   []byte(`"👋"`),
	}
	assert.NoError(tx.SyntacticVerify())
	cbordata, err := tx.Marshal()
	assert.NoError(err)
	jsondata, err := json.Marshal(tx)
	assert.NoError(err)

	assert.Contains(string(jsondata), `"0xFFfFfFffFFfffFFfFFfFFFFFffFFFffffFfFFFfF"`)
	assert.Contains(string(jsondata), `"token":""`)
	assert.Contains(string(jsondata), `"data":"👋"`)
	assert.NotContains(string(jsondata), `"from":`)

	tx2 := &TxTransfer{}
	assert.NoError(tx2.Unmarshal(cbordata))
	assert.NoError(tx2.SyntacticVerify())
	cbordata2 := tx2.Bytes()
	jsondata2, err := json.Marshal(tx2)
	assert.Equal(string(jsondata), string(jsondata2))
	assert.Equal(cbordata, cbordata2)
}
