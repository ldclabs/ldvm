// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package ld

import (
	"encoding/json"
	"math/big"
	"testing"
	"time"

	"github.com/ldclabs/ldvm/util"
	"github.com/ldclabs/ldvm/util/signer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBlock(t *testing.T) {
	assert := assert.New(t)

	var blk *Block
	assert.ErrorContains(blk.SyntacticVerify(), "nil pointer")

	blk = &Block{}
	assert.ErrorContains(blk.SyntacticVerify(), "invalid state AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAACeYpGX")

	blk = &Block{State: util.Hash{1, 2, 3}, Timestamp: uint64(time.Now().Unix() + 20)}
	assert.ErrorContains(blk.SyntacticVerify(), "invalid timestamp")

	blk = &Block{State: util.Hash{1, 2, 3}, GasPrice: 1}
	assert.ErrorContains(blk.SyntacticVerify(), "invalid gasPrice")

	blk = &Block{State: util.Hash{1, 2, 3}, GasPrice: 100, GasRebateRate: 1001}
	assert.ErrorContains(blk.SyntacticVerify(), "invalid gasRebateRate")

	blk = &Block{State: util.Hash{1, 2, 3}, GasPrice: 100, Builder: util.StakeSymbol{1, 2, 3}}
	assert.ErrorContains(blk.SyntacticVerify(), "invalid builder address")

	blk = &Block{State: util.Hash{1, 2, 3}, GasPrice: 100}
	assert.ErrorContains(blk.SyntacticVerify(), "nil validators")

	blk = &Block{State: util.Hash{1, 2, 3}, GasPrice: 100,
		Validators: util.IDList[util.StakeSymbol]{}}
	assert.ErrorContains(blk.SyntacticVerify(), "no txs")

	blk = &Block{State: util.Hash{1, 2, 3}, GasPrice: 100,
		Validators: util.IDList[util.StakeSymbol]{{1, 2, 3}},
		Txs:        util.IDList[util.Hash]{{1, 2, 3}}}
	assert.ErrorContains(blk.SyntacticVerify(), "invalid validator address")

	blk = &Block{State: util.Hash{1, 2, 3}, GasPrice: 100,
		Validators: util.IDList[util.StakeSymbol]{},
		Txs:        util.IDList[util.Hash]{util.Hash{}}}
	assert.ErrorContains(blk.SyntacticVerify(), "empty id exists")

	tx := &Transaction{
		Tx: TxData{
			Type:      TypeTransfer,
			ChainID:   gChainID,
			Nonce:     1,
			GasTip:    0,
			GasFeeCap: 1000,
			From:      signer.Signer1.Key().Address(),
			To:        signer.Signer2.Key().Address().Ptr(),
			Amount:    big.NewInt(1_000_000),
		},
	}
	assert.NoError(tx.SignWith(signer.Signer1))
	assert.NoError(tx.SyntacticVerify())

	blk = &Block{
		State:         util.Hash{1, 2, 3},
		Gas:           tx.Gas(),
		GasPrice:      1000,
		GasRebateRate: 200,
		Validators:    util.IDList[util.StakeSymbol]{},
		Txs:           util.IDList[util.Hash]{tx.ID},
	}

	assert.NoError(blk.SyntacticVerify())
	cbordata, err := blk.Marshal()
	require.NoError(t, err)

	jsondata, err := json.Marshal(blk)
	require.NoError(t, err)
	// fmt.Println(string(jsondata))
	assert.Equal(`{"parent":"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAACeYpGX","height":0,"timestamp":0,"state":"AQIDAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAoWLSv","gas":638,"gasPrice":1000,"gasRebateRate":200,"builder":"","validators":[],"pChainHeight":0,"txs":["aLokjgaVT95weTdJmhe2T1VjnvqfqaDNx7JHtRuo8TAsHAps"],"id":"YazT1E6_dY1V6m3OAPQYdxmm86crUGm7VVPddkAKwUX-bsbE"}`, string(jsondata))

	blk2 := &Block{}
	assert.NoError(blk2.Unmarshal(cbordata))
	assert.NoError(blk2.SyntacticVerify())

	jsondata2, err := json.Marshal(blk2)
	require.NoError(t, err)
	assert.Equal(string(jsondata), string(jsondata2))
	assert.Equal(blk.ID, blk2.ID)
	assert.Equal(cbordata, blk2.Bytes())

	blk2.Gas++
	assert.NoError(blk2.SyntacticVerify())
	assert.NotEqual(blk.ID, blk2.ID)
	assert.NotEqual(cbordata, blk2.Bytes())
}
