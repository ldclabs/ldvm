// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package ld

import (
	"encoding/json"
	"math/big"
	"testing"

	"github.com/ldclabs/ldvm/constants"
	"github.com/ldclabs/ldvm/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
		Expire: uint64(1000),
		Data:   []byte(`"👋"`),
	}
	assert.NoError(tx.SyntacticVerify())
	cbordata, err := tx.Marshal()
	require.NoError(t, err)
	jsondata, err := json.Marshal(tx)
	require.NoError(t, err)
	// fmt.Println(string(jsondata))
	assert.Equal(`{"nonce":1,"to":"0xFFfFFFfFfffFFfFFffFFFfFfFffFFFfffFfFFFff","token":"","amount":1000,"expire":1000,"data":"👋"}`, string(jsondata))

	tx2 := &TxTransfer{}
	assert.NoError(tx2.Unmarshal(cbordata))
	assert.NoError(tx2.SyntacticVerify())
	cbordata2 := tx2.Bytes()
	jsondata2, _ := json.Marshal(tx2)
	assert.Equal(string(jsondata), string(jsondata2))
	assert.Equal(cbordata, cbordata2)
}
