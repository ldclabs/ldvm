// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package ld

import (
	"encoding/json"
	"math/big"
	"testing"
	"time"

	"github.com/ldclabs/ldvm/util"
	"github.com/stretchr/testify/assert"
)

func TestBlock(t *testing.T) {
	assert := assert.New(t)

	var blk *Block
	assert.ErrorContains(blk.SyntacticVerify(), "nil pointer")

	blk = &Block{Timestamp: uint64(time.Now().Unix() + 20)}
	assert.ErrorContains(blk.SyntacticVerify(), "invalid timestamp")

	blk = &Block{GasRebateRate: 1001}
	assert.ErrorContains(blk.SyntacticVerify(), "invalid gasRebateRate")

	blk = &Block{Miner: util.StakeSymbol{1, 2, 3}}
	assert.ErrorContains(blk.SyntacticVerify(), "invalid miner address")

	blk = &Block{Shares: []util.StakeSymbol{{1, 2, 3}}}
	assert.ErrorContains(blk.SyntacticVerify(), "invalid share address")

	blk = &Block{}
	assert.ErrorContains(blk.SyntacticVerify(), "no txs")

	blk = &Block{Txs: make([]*Transaction, 1)}
	assert.ErrorContains(blk.SyntacticVerify(), "Block.SyntacticVerify failed: Transaction.SyntacticVerify failed: nil pointer")

	to := util.Signer2.Address()
	txData := &TxData{
		Type:      TypeTransfer,
		ChainID:   gChainID,
		Nonce:     1,
		GasTip:    0,
		GasFeeCap: 1000,
		From:      util.Signer1.Address(),
		To:        &to,
		Amount:    big.NewInt(1_000_000),
	}
	sig1, err := util.Signer1.Sign(txData.UnsignedBytes())
	assert.NoError(err)
	txData.Signatures = append(txData.Signatures, sig1)
	tx := txData.ToTransaction()
	blk = &Block{
		Gas:           tx.RequiredGas(1000),
		GasPrice:      1000,
		GasRebateRate: 200,
		Txs:           Txs{tx},
	}

	assert.NoError(blk.SyntacticVerify())
	assert.Equal(uint64(119000), blk.FeeCost().Uint64())
	cbordata, err := blk.Marshal()
	assert.NoError(err)
	assert.NoError(blk.MarshalTxsJSON())
	jsondata, err := json.Marshal(blk)
	assert.NoError(err)

	assert.Contains(string(jsondata), `"parent":"11111111111111111111111111111111LpoYY"`)
	assert.Contains(string(jsondata), `"height":0,"timestamp":0`)
	assert.Contains(string(jsondata), `"gas":119,"gasPrice":1000,"gasRebateRate":200`)
	assert.Contains(string(jsondata), `"miner":"","shares":null`)
	assert.Contains(string(jsondata), `"name":"TransferTx"`)

	blk2 := &Block{}
	assert.NoError(blk2.Unmarshal(cbordata))
	assert.NoError(blk2.SyntacticVerify())
	assert.NoError(blk2.MarshalTxsJSON())

	jsondata2, err := json.Marshal(blk2)
	assert.NoError(err)
	assert.Equal(string(jsondata), string(jsondata2))
	assert.Equal(cbordata, blk2.Bytes())
}
