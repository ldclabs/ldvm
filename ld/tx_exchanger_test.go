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

func TestTxExchanger(t *testing.T) {
	assert := assert.New(t)

	token, _ := util.NewToken("$USD")

	var tx *TxExchanger
	assert.ErrorContains(tx.SyntacticVerify(), "nil pointer")

	tx = &TxExchanger{}
	assert.ErrorContains(tx.SyntacticVerify(), "invalid nonce")

	tx = &TxExchanger{Nonce: 1, Sell: util.TokenSymbol{'a', 'b', 'c'}}
	assert.ErrorContains(tx.SyntacticVerify(), `invalid sell token symbol "0x6162630000000000000000000000000000000000"`)

	tx = &TxExchanger{Nonce: 1, Sell: util.NativeToken}
	assert.ErrorContains(tx.SyntacticVerify(), "sell and receive token should not equal")

	tx = &TxExchanger{Nonce: 1, Sell: util.NativeToken, Receive: token}
	assert.ErrorContains(tx.SyntacticVerify(), "invalid minimum")

	tx = &TxExchanger{Nonce: 1, Sell: util.NativeToken, Receive: token,
		Minimum: big.NewInt(1000), Quota: big.NewInt(999)}
	assert.ErrorContains(tx.SyntacticVerify(), "invalid quota")

	tx = &TxExchanger{
		Nonce:   1,
		Sell:    token,
		Price:   big.NewInt(0),
		Quota:   big.NewInt(1000000),
		Minimum: big.NewInt(1000),
		Expire:  uint64(time.Now().Unix()),
		Payee:   constants.GenesisAccount,
	}

	assert.Error(tx.SyntacticVerify())
	tx.Price = big.NewInt(1000)
	assert.NoError(tx.SyntacticVerify())
	cbordata, err := tx.Marshal()
	assert.NoError(err)
	jsondata, err := json.Marshal(tx)
	assert.NoError(err)
	assert.Contains(string(jsondata), `"payee":"0xFFfFfFffFFfffFFfFFfFFFFFffFFFffffFfFFFfF"`)
	assert.Contains(string(jsondata), `"receive":""`)
	assert.Contains(string(jsondata), `"sell":"$USD"`)
	assert.NotContains(string(jsondata), `"from":`)

	tx2 := &TxExchanger{}
	assert.NoError(tx2.Unmarshal(cbordata))
	assert.NoError(tx2.SyntacticVerify())
	cbordata2 := tx2.Bytes()
	jsondata2, _ := json.Marshal(tx2)
	assert.Equal(string(jsondata), string(jsondata2))
	assert.Equal(cbordata, cbordata2)
}
